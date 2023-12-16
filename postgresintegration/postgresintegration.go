package postgresintegration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// postgresImageName specifies the default Docker image name for Postgres.
const postgresImageName = "postgres:16.1-alpine"

const IntegrationRunnerEnvVar = "RUN_INTEGRATION_TESTS"

// MustRunPostgresDockerContainer creates a new Postgres test container and initializes the application repositories.
// It returns a cleanup function that must be called to terminate the container.
// It panics if the container cannot be started.
func MustRunPostgresDockerContainer() (PostgresDockerInstance, func()) {
	instance, cleanupFn, err := runPostgresDockerContainer(postgresImageName)
	if err != nil {
		panic(err)
	}
	return instance, cleanupFn
}

// RunPostgresDockerContainer creates a new Postgres test container and initializes the application repositories.
// It returns a cleanup function that must be called to terminate the container.
func RunPostgresDockerContainer() (PostgresDockerInstance, func(), error) {
	return runPostgresDockerContainer(postgresImageName)
}

// runPostgresDockerContainer creates a new Postgres test container and initializes the application repositories.
// It returns a cleanup function that must be called to terminate the container.
func runPostgresDockerContainer(postgresImageName string) (PostgresDockerInstance, func(), error) {
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
		return PostgresDockerInstance{}, func() {}, fmt.Errorf("postgres container start: %w", err)
	}

	// Test container cleanup function:
	terminateFn := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate Postgres test container: %+v", err)
			return
		}
		log.Println("Postgres test container terminated")
	}

	postgresHostIP, err := postgresContainer.Host(ctx)
	if err != nil {
		return PostgresDockerInstance{}, terminateFn, fmt.Errorf("map Postgres host: %w", err)
	}

	postgresHostPort, err := postgresContainer.MappedPort(ctx, postgresPort)
	if err != nil {
		return PostgresDockerInstance{}, terminateFn, fmt.Errorf("map Postgres port: %w", err)
	}

	connURL := fmt.Sprintf(connURLTemplate, userName, userPass, postgresHostIP, postgresHostPort.Port(), dbName)

	// setup PGX connection pool:
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return PostgresDockerInstance{}, terminateFn, fmt.Errorf("failed to create PGX connection pool: %w", err)
	}

	instance := PostgresDockerInstance{
		connURL:      connURL,
		postgresPool: pool,
	}
	log.Printf("Postgres container started, running at: %q\n", connURL)
	return instance, terminateFn, nil
}

// PostgresDockerInstance represents a running Postgres test container with settings.
type PostgresDockerInstance struct {
	// connURL is a fully constructed connection URL with all resolved values, using this template: "postgres://%s:%s@%s:%s/%s?sslmode=disable".
	connURL string
	// postgresPool is a PGX connection pool that can be used to execute queries against the Postgres DB.
	postgresPool *pgxpool.Pool
}

// ConnURL returns a fully constructed connection URL with all resolved values, using this template: "postgres://%s:%s@%s:%s/%s?sslmode=disable".
func (p *PostgresDockerInstance) ConnURL() string {
	return p.connURL
}

// PgxPool returns a PGX connection pool that can be used to execute queries against the Postgres DB.
func (p *PostgresDockerInstance) PgxPool() *pgxpool.Pool {
	return p.postgresPool
}

// MustTruncateData truncates all data in the Postgres DB.
// Can be used after the tests to clean up all the user's data.
//
// It panics if the truncation fails.
func (p *PostgresDockerInstance) MustTruncateData() {
	ctx := context.Background()
	tableNames := func() []string {
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		rows, err := p.postgresPool.Query(ctx, query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var table string
			err := rows.Scan(&table)
			if err != nil {
				panic(err)
			}
			tables = append(tables, table)
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}
		return tables
	}()

	for _, tableName := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
		_, err := p.postgresPool.Exec(ctx, query)
		if err != nil {
			panic(err)
		}
	}
}

func (p *PostgresDockerInstance) TruncateDataInTest(t *testing.T) {
	ctx := context.Background()
	tableNames := func() []string {
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		rows, err := p.postgresPool.Query(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var table string
			err := rows.Scan(&table)
			require.NoError(t, err)
			tables = append(tables, table)
		}
		require.NoError(t, rows.Err())
		return tables
	}()

	for _, tableName := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
		_, err := p.postgresPool.Exec(ctx, query)
		require.NoError(t, err)
	}
}

// IsSkipIntegrationTest returns true if the integration test should be skipped.
func IsSkipIntegrationTest(t *testing.T) bool {
	if _, ok := os.LookupEnv(IntegrationRunnerEnvVar); !ok {
		t.Skipf("skipping integration test, set %q env variable to run it", IntegrationRunnerEnvVar)
		return true
	}
	return false
}
