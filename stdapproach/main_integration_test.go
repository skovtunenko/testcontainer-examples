package stdapproach

import (
	stdlog "log"
	"os"
	"testing"

	"github.com/skovtunenko/testcontainer-examples/integrationtesting"
)

func TestMain(m *testing.M) {
	esConf, terminateElasticFn, err := integrationtesting.RunElasticsearchDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise ElasticSearch test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("ElasticSearch configuration: %+v", esConf)

	mongoConf, terminateMongoFn, err := integrationtesting.RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("MongoDB configuration: %+v", mongoConf)

	postgresConf, terminatePostgresFn, err := integrationtesting.RunPostgresDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialize Postgres test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("Postgres configuration: %+v", postgresConf)

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
