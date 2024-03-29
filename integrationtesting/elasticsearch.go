package integrationtesting

import (
	"context"
	"fmt"
	stdlog "log"
	"math/rand"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// elasticImageName specifies Docker image name for ElasticSearch.
	elasticImageName = "elasticsearch:7.17.4"
)

// ElasticDockerInstance is a config with ElasticSearch connection settings.
type ElasticDockerInstance struct {
	ConnURL string
}

// RunElasticsearchDockerContainer creates new ElasticSearch test container and initializes application repositories.
// Returns cleanup function that must be called.
func RunElasticsearchDockerContainer() (ElasticDockerInstance, func(), error) {
	ctx := context.Background()
	rand.Seed(time.Now().UnixMilli())
	const (
		elasticInternalPort = "9200"

		elasticConnectionURLTemplate = "http://%s:%s"
	)

	elasticPort := nat.Port(elasticInternalPort + "/tcp")
	containerRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: elasticImageName,
			Env: map[string]string{
				"discovery.type":         "single-node",
				"cluster.name":           fmt.Sprintf("testcontainer-%d", rand.Int()),
				"ES_JAVA_OPTS":           "-Xms512m -Xmx1024m",
				"bootstrap.memory_lock":  "true",
				"xpack.security.enabled": "false",
			},
			ExposedPorts: []string{elasticPort.Port()},
			WaitingFor:   wait.ForListeningPort(elasticPort),
		},
		Started: true, // auto-start the container
	}
	elasticContainer, err := testcontainers.GenericContainer(ctx, containerRequest)
	if err != nil {
		return ElasticDockerInstance{}, func() {}, fmt.Errorf("elasticSearch container start: %w", err)
	}

	// Test container clean-up function:
	terminateFn := func() {
		if err := elasticContainer.Terminate(ctx); err != nil {
			stdlog.Printf("failed to terminate ElasticSearch test container: %+v", err)
			return
		}
		stdlog.Println("ElasticSearch test container terminated")
	}

	elasticHostIP, err := elasticContainer.Host(ctx)
	if err != nil {
		return ElasticDockerInstance{}, terminateFn, fmt.Errorf("map ElasticSearch host: %w", err)
	}

	elasticHostPort, err := elasticContainer.MappedPort(ctx, elasticPort)
	if err != nil {
		return ElasticDockerInstance{}, terminateFn, fmt.Errorf("map ElasticSearch port: %w", err)
	}

	elasticURL := fmt.Sprintf(elasticConnectionURLTemplate, elasticHostIP, elasticHostPort.Port())
	instance := ElasticDockerInstance{
		ConnURL: elasticURL,
	}

	stdlog.Printf("ElasticSearch container started, running at: %q\n", elasticURL)
	return instance, terminateFn, nil
}
