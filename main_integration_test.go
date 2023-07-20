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
		return
	}
	stdlog.Printf("ElasticSearch configuration: %+v", esConf)

	mongoConf, terminateMongoFn, err := RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		return
	}
	stdlog.Printf("MongoDB configuration: %+v", mongoConf)

	var exitCode int
	func() {
		defer terminateElasticFn()
		defer terminateMongoFn()
		exitCode = m.Run() // execute the tests
	}()
	os.Exit(exitCode)
}
