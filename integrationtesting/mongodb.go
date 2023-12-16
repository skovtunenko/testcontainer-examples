package integrationtesting

import (
	"context"
	"fmt"
	stdlog "log"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// mongoImageName specifies Docker image name for MongoDB.
	mongoImageName = "mongo:4.2.21"
)

// MongoDockerInstance is a config with MongoDB connection settings.
type MongoDockerInstance struct {
	ConnURL  string
	UserName string
	UserPass string
}

// RunMongoDockerContainer creates new MongoDB test container and initializes application repositories.
// Returns cleanup function that must be called.
func RunMongoDockerContainer() (MongoDockerInstance, func(), error) {
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
		return MongoDockerInstance{}, func() {}, fmt.Errorf("mongoDB container start: %w", err)
	}

	// Test container clean up function:
	terminateFn := func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			stdlog.Printf("failed to terminate MongoDB test container: %v", err)
			return
		}
		stdlog.Println("MongoDB test container terminated")
	}

	mongoHostIP, err := mongoContainer.Host(ctx)
	if err != nil {
		return MongoDockerInstance{}, terminateFn, fmt.Errorf("map MongoDB host: %w", err)
	}

	mongoHostPort, err := mongoContainer.MappedPort(ctx, mongoPort)
	if err != nil {
		return MongoDockerInstance{}, terminateFn, fmt.Errorf("map MongoDB port: %w", err)
	}

	mongoURL := fmt.Sprintf(mongoConnectionURLTemplate, userName, userPass, mongoHostIP, mongoHostPort.Port())
	instance := MongoDockerInstance{
		ConnURL:  mongoURL,
		UserName: userName,
		UserPass: userPass,
	}
	stdlog.Printf("MongoDB container started, running at: %q\n", mongoURL)
	return instance, terminateFn, nil
}
