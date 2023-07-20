package tcinfra

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// postgresImageName specifies the Docker image name for PostgreSQL.
	postgresImageName = "postgres:13.4-alpine"
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

// RunPostgreSQLDockerContainer creates a new PostgreSQL test container and initializes the application repositories.
// It returns a cleanup function.
func RunPostgreSQLDockerContainer() (PostgreSQLConfig, func(), error) {
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
			log.Println("Failed to terminate PostgreSQL test container")
			return
		}
		log.Println("PostgreSQL test container terminated")
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
	log.Printf("PostgreSQL container started, running at: %q\n", connURL)
	return cfg, terminateFn, nil
}
