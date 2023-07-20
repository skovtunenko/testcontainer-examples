package integrationtesting

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// postgresImageName specifies the Docker image name for PostgreSQL.
	postgresImageName = "postgres:13.4-alpine"
)
const integrationRunnerEnvVar = "RUN_INTEGRATION_TESTS"

// PostgresSuite is a basic integration suite for Postgres-related integration tests.
type PostgresSuite struct {
	suite.Suite
	postgresConfig               PostgresConfig
	postgresContainerTerminateFn func()
	postgresPool                 *pgxpool.Pool
}

// GetPostgresConnectionURL returns connection URL to integration PostgreSQL in Docker.
func (suite *PostgresSuite) GetPostgresConnectionURL() string {
	return suite.postgresConfig.ConnURL
}

// SetupSuite will run before the tests in the suite are run.
func (suite *PostgresSuite) SetupSuite() {
	if os.Getenv(integrationRunnerEnvVar) == "" {
		suite.T().Skipf("Skipping Postgres integration tests. To enable them, set non-empty value in %q environment variable", integrationRunnerEnvVar)
		return
	}

	// run temp. integration Docker container:
	r := suite.Require()
	conf, cleanFn, err := runPostgreSQLDockerContainer(suite.T())
	r.NoError(err)
	suite.postgresConfig = conf
	suite.postgresContainerTerminateFn = cleanFn

	// setup connection to PostgreSQL:
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, suite.GetPostgresConnectionURL())
	r.NoError(err)
	suite.postgresPool = pool
}

// TearDownSuite will run after all the tests in the suite have been run.
func (suite *PostgresSuite) TearDownSuite() {
	if suite.postgresContainerTerminateFn != nil {
		suite.postgresContainerTerminateFn()
	}
	if suite.postgresPool != nil {
		suite.postgresPool.Close()
	}
}

// TearDownTest will run after each test in the suite.
func (suite *PostgresSuite) TearDownTest() {
	ctx := context.Background()
	r := suite.Require()

	tableNames := func() []string {
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		rows, err := suite.postgresPool.Query(ctx, query)
		r.NoError(err)
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var table string
			err := rows.Scan(&table)
			r.NoError(err)
			tables = append(tables, table)
		}
		r.NoError(rows.Err())
		return tables
	}()

	for _, tableName := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
		_, err := suite.postgresPool.Exec(ctx, query)
		r.NoError(err)
	}
}

// PostgresConfig contains the PostgreSQL connection settings.
type PostgresConfig struct {
	// ConnURL is fully-constructed connection URL with all resolved values, using this template: "postgres://%s:%s@%s:%s/%s?sslmode=disable".
	ConnURL string
	// UserName is a user name used for DB connection.
	UserName string
	// UserPass is a user password used for DB connection.
	UserPass string
	// DbName is a name of the DB.
	DbName string
}

// runPostgreSQLDockerContainer creates a new PostgreSQL test container and initializes the application repositories.
// It returns a cleanup function.
func runPostgreSQLDockerContainer(t *testing.T) (PostgresConfig, func(), error) {
	ctx := context.Background()
	const (
		postgresInternalPort = "5432"
		userName             = "testuser"
		userPass             = "testpassword"
		dbName               = "integrationdb"
		connURLTemplate      = "postgres://%s:%s@%s:%s/%s?sslmode=disable"
	)

	postgresPort := nat.Port(postgresInternalPort + "/tcp")
	containerRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        postgresImageName,
			ExposedPorts: []string{postgresPort.Port()},
			Env: map[string]string{
				"POSTGRES_USER":     userName,
				"POSTGRES_PASSWORD": userPass,
				"POSTGRES_DB":       dbName,
			},
			WaitingFor: wait.ForListeningPort(postgresPort),
		},
		Started: true, // auto-start the container
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, containerRequest)
	if err != nil {
		return PostgresConfig{}, func() {}, errors.Wrap(err, "PostgreSQL container start")
	}

	// Test container cleanup function:
	terminateFn := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Log("Failed to terminate PostgreSQL test container")
			return
		}
		t.Log("PostgreSQL test container successfully terminated")
	}

	postgresHostIP, err := postgresContainer.Host(ctx)
	if err != nil {
		return PostgresConfig{}, func() {}, errors.Wrap(err, "map PostgreSQL host")
	}

	postgresHostPort, err := postgresContainer.MappedPort(ctx, postgresPort)
	if err != nil {
		return PostgresConfig{}, func() {}, errors.Wrap(err, "map PostgreSQL port")
	}

	connURL := fmt.Sprintf(connURLTemplate, userName, userPass, postgresHostIP, postgresHostPort.Port(), dbName)
	cfg := PostgresConfig{
		ConnURL:  connURL,
		UserName: userName,
		UserPass: userPass,
		DbName:   dbName,
	}
	t.Logf("PostgreSQL container started, running at: %q", connURL)
	return cfg, terminateFn, nil
}

var (
	_ suite.SetupAllSuite     = &PostgresSuite{}
	_ suite.TearDownAllSuite  = &PostgresSuite{}
	_ suite.TearDownTestSuite = &PostgresSuite{}
)
