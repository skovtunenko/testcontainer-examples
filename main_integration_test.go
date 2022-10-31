//go:build integration

package testcontainer_examples

import (
	stdlog "log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	terminateElasticFn, err := RunElasticsearchDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise ElasticSearch test container: %+v", err)
		return
	}

	terminateMongoFn, err := RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		return
	}

	var exitCode int
	func() {
		defer terminateElasticFn()
		defer terminateMongoFn()
		exitCode = m.Run() // execute the tests
	}()
	os.Exit(exitCode)
}
