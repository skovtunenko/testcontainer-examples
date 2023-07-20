package integrationtesting

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
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

// Suite is a basic integration suite for PostgreSQL-related integration tests.
type Suite struct {
	suite.Suite
	postgresqlConfig  PostgreSQLConfig
	postgresqlCleanFn func()
}

// GetPostgresConnectionURL returns connection URL to integration PostgreSQL in Docker.
func (suite *Suite) GetPostgresConnectionURL() string {
	return suite.postgresqlConfig.ConnURL
}

// SetupSuite will run before the tests in the suite are run.
func (suite *Suite) SetupSuite() {
	if os.Getenv(integrationRunnerEnvVar) == "" {
		suite.T().Skipf("Skipping Postgres integration tests. To enable them, set non-empty value in %q environment variable", integrationRunnerEnvVar)
		return
	}

	// run temp. integration Docker container:
	r := suite.Require()
	conf, cleanFn, err := runPostgreSQLDockerContainer(suite.T())
	r.NoError(err)
	suite.postgresqlConfig = conf
	suite.postgresqlCleanFn = cleanFn
}

// TearDownSuite will run after all the tests in the suite have been run.
func (suite *Suite) TearDownSuite() {
	if suite.postgresqlCleanFn != nil {
		suite.postgresqlCleanFn()
	}
}

// TearDownTest will run after each test in the suite.
func (suite *Suite) TearDownTest() {
	// TODO: truncate all table values...
}

var (
	_ suite.SetupAllSuite     = &Suite{}
	_ suite.TearDownAllSuite  = &Suite{}
	_ suite.TearDownTestSuite = &Suite{}
)

// PostgreSQLConfig contains the PostgreSQL connection settings.
type PostgreSQLConfig struct {
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
func runPostgreSQLDockerContainer(t *testing.T) (PostgreSQLConfig, func(), error) {
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
		return PostgreSQLConfig{}, func() {}, errors.Wrap(err, "PostgreSQL container start")
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
		return PostgreSQLConfig{}, func() {}, errors.Wrap(err, "map PostgreSQL host")
	}

	postgresHostPort, err := postgresContainer.MappedPort(ctx, postgresPort)
	if err != nil {
		return PostgreSQLConfig{}, func() {}, errors.Wrap(err, "map PostgreSQL port")
	}

	connURL := fmt.Sprintf(connURLTemplate, userName, userPass, postgresHostIP, postgresHostPort.Port(), dbName)
	cfg := PostgreSQLConfig{
		ConnURL:  connURL,
		UserName: userName,
		UserPass: userPass,
		DbName:   dbName,
	}
	t.Logf("PostgreSQL container started, running at: %q", connURL)
	return cfg, terminateFn, nil
}
