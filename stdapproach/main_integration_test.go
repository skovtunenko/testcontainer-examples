package stdapproach

import (
	stdlog "log"
	"os"
	"testing"

	tcinfra "github.com/skovtunenko/testcontainer-examples"
)

func TestMain(m *testing.M) {
	esConf, terminateElasticFn, err := tcinfra.RunElasticsearchDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise ElasticSearch test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("ElasticSearch configuration: %+v", esConf)

	mongoConf, terminateMongoFn, err := tcinfra.RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("MongoDB configuration: %+v", mongoConf)

	postgresConf, terminatePostgresFn, err := tcinfra.RunPostgreSQLDockerContainer()
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

func TestSample(t *testing.T) {
	t.Log("Executing simple test")
}
