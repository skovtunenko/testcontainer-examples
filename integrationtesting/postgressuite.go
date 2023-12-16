package integrationtesting

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// postgresImageName specifies the Docker image name for Postgres.
	postgresImageName = "postgres:16.1-alpine"
)
const IntegrationRunnerEnvVar = "RUN_INTEGRATION_TESTS"

// PostgresSuite is a basic integration suite for Postgres-related integration tests.
type PostgresSuite struct {
	suite.Suite
	postgresConfig               PostgresConfig
	postgresContainerTerminateFn func()
	postgresPool                 *pgxpool.Pool
}

// GetPostgresConnectionURL returns connection URL to integration Postgres in Docker.
func (suite *PostgresSuite) GetPostgresConnectionURL() string {
	return suite.postgresConfig.ConnURL
}

// PgxPool returns a connection pool to integration Postgres in Docker.
func (suite *PostgresSuite) PgxPool() *pgxpool.Pool {
	return suite.postgresPool
}

// SetupSuite will run before the tests in the suite are run.
func (suite *PostgresSuite) SetupSuite() {
	if os.Getenv(IntegrationRunnerEnvVar) == "" {
		suite.T().Skipf("Skipping Postgres integration tests. To enable them, set non-empty value in %q environment variable", IntegrationRunnerEnvVar)
		return
	}

	// run temp. integration Docker container:
	r := suite.Require()
	conf, cleanFn, err := runPostgresDockerContainer(suite.T())
	r.NoError(err)
	suite.postgresConfig = conf
	suite.postgresContainerTerminateFn = cleanFn

	// setup connection to Postgres:
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

// PostgresConfig contains the Postgres connection settings.
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

// runPostgresDockerContainer creates a new Postgres test container and initializes the application repositories.
// It returns a cleanup function.
func runPostgresDockerContainer(t *testing.T) (PostgresConfig, func(), error) {
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
		return PostgresConfig{}, func() {}, fmt.Errorf("postgres container start: %w", err)
	}

	// Test container cleanup function:
	terminateFn := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Log("Failed to terminate Postgres test container")
			return
		}
		t.Log("Postgres test container successfully terminated")
	}

	postgresHostIP, err := postgresContainer.Host(ctx)
	if err != nil {
		return PostgresConfig{}, terminateFn, fmt.Errorf("map Postgres host: %w", err)
	}

	postgresHostPort, err := postgresContainer.MappedPort(ctx, postgresPort)
	if err != nil {
		return PostgresConfig{}, terminateFn, fmt.Errorf("map Postgres port: %w", err)
	}

	connURL := fmt.Sprintf(connURLTemplate, userName, userPass, postgresHostIP, postgresHostPort.Port(), dbName)
	cfg := PostgresConfig{
		ConnURL:  connURL,
		UserName: userName,
		UserPass: userPass,
		DbName:   dbName,
	}
	t.Logf("Postgres container started, running at: %q", connURL)
	return cfg, terminateFn, nil
}

var (
	_ suite.SetupAllSuite     = &PostgresSuite{}
	_ suite.TearDownAllSuite  = &PostgresSuite{}
	_ suite.TearDownTestSuite = &PostgresSuite{}
)
