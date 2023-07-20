package integrationtesting

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunPostgresDockerContainer creates a new Postgres test container and initializes the application repositories.
// It returns a cleanup function.
func RunPostgresDockerContainer() (PostgresConfig, func(), error) {
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
		return PostgresConfig{}, func() {}, errors.Wrap(err, "Postgres container start")
	}

	// Test container cleanup function:
	terminateFn := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Println("Failed to terminate Postgres test container")
			return
		}
		log.Println("Postgres test container terminated")
	}

	postgresHostIP, err := postgresContainer.Host(ctx)
	if err != nil {
		return PostgresConfig{}, func() {}, errors.Wrap(err, "map Postgres host")
	}

	postgresHostPort, err := postgresContainer.MappedPort(ctx, postgresPort)
	if err != nil {
		return PostgresConfig{}, func() {}, errors.Wrap(err, "map Postgres port")
	}

	connURL := fmt.Sprintf(connURLTemplate, userName, userPass, postgresHostIP, postgresHostPort.Port(), dbName)
	cfg := PostgresConfig{
		ConnURL:  connURL,
		UserName: userName,
		UserPass: userPass,
		DbName:   dbName,
	}
	log.Printf("Postgres container started, running at: %q\n", connURL)
	return cfg, terminateFn, nil
}
