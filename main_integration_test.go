package testcontainer_examples

import (
	stdlog "log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	esConf, terminateElasticFn, err := RunElasticsearchDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise ElasticSearch test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("ElasticSearch configuration: %+v", esConf)

	mongoConf, terminateMongoFn, err := RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("MongoDB configuration: %+v", mongoConf)

	postgresConf, terminatePostgresFn, err := RunPostgreSQLDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialize PostgreSQL test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("PostgreSQL configuration: %+v", postgresConf)

	var exitCode int
	func() {
		defer terminateElasticFn()
		defer terminateMongoFn()
		defer terminatePostgresFn()

		exitCode = m.Run() // execute the tests
	}()
	os.Exit(exitCode)
}
