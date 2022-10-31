package testcontainer_examples

import (
	"context"
	"fmt"
	stdlog "log"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// mongoImageName specifies Docker image name for MongoDB.
	mongoImageName = "mongo:4.2.21"
)

// MongoConfig is a config with Mongo connection settings.
type MongoConfig struct {
	ConnURL  string
	UserName string
	UserPass string
}

// RunMongoDockerContainer creates new MongoDB test container and initializes application repositories.
// Returns cleanup function.
func RunMongoDockerContainer() (MongoConfig, func(), error) {
	ctx := context.Background()
	const (
		mongoInternalPort = "27017"

		userName                   = "root"
		userPass                   = "pass"
		mongoConnectionURLTemplate = "mongodb://%s:%s@%s:%s/?connect=direct"
	)

	mongoPort := nat.Port(mongoInternalPort + "/tcp")
	containerRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        mongoImageName,
			ExposedPorts: []string{mongoPort.Port()},
			Env: map[string]string{
				"MONGO_INITDB_ROOT_USERNAME": userName,
				"MONGO_INITDB_ROOT_PASSWORD": userPass,
			},
			WaitingFor: wait.ForListeningPort(mongoPort),
		},
		Started: true, // auto-start the container
	}
	mongoContainer, err := testcontainers.GenericContainer(ctx, containerRequest)
	if err != nil {
		return MongoConfig{}, func() {}, errors.Wrap(err, "MongoDB container start")
	}

	// Test container clean up function:
	terminateFn := func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			stdlog.Println("failed to terminate MongoDB test container")
			return
		}
		stdlog.Println("MongoDB test container terminated")
	}

	mongoHostIP, err := mongoContainer.Host(ctx)
	if err != nil {
		return MongoConfig{}, func() {}, errors.Wrap(err, "map MongoDB host")
	}

	mongoHostPort, err := mongoContainer.MappedPort(ctx, mongoPort)
	if err != nil {
		return MongoConfig{}, func() {}, errors.Wrap(err, "map MongoDB port")
	}

	mongoURL := fmt.Sprintf(mongoConnectionURLTemplate, userName, userPass, mongoHostIP, mongoHostPort.Port())
	cfg := MongoConfig{
		ConnURL:  mongoURL,
		UserName: userName,
		UserPass: userPass,
	}
	stdlog.Printf("MongoDB container started, running at: %q\n", mongoURL)
	return cfg, terminateFn, nil
}
