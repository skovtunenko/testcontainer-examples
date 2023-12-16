package stdapproachwithmaintest

import (
	stdlog "log"
	"os"
	"testing"

	"github.com/skovtunenko/testcontainer-examples/integrationtesting"
)

// Global variables to store configuration of running test containers.
var (
	esDockerInstance       integrationtesting.ElasticDockerInstance
	mongoDockerInstance    integrationtesting.MongoDockerInstance
	postgresDockerInstance integrationtesting.PostgresDockerInstance
)

func TestMain(m *testing.M) {
	es, terminateElasticFn, err := integrationtesting.RunElasticsearchDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise ElasticSearch test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("ElasticSearch configuration: %+v", es)
	esDockerInstance = es

	mongo, terminateMongoFn, err := integrationtesting.RunMongoDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialise MongoDB test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("MongoDB configuration: %+v", mongo)
	mongoDockerInstance = mongo

	postgres, terminatePostgresFn, err := integrationtesting.RunPostgresDockerContainer()
	if err != nil {
		stdlog.Printf("failed to initialize Postgres test container: %+v", err)
		os.Exit(1)
		return
	}
	stdlog.Printf("Postgres configuration: %+v", postgres)
	postgresDockerInstance = postgres

	var exitCode int
	func() {
		defer terminateElasticFn()
		defer terminateMongoFn()
		defer terminatePostgresFn()

		exitCode = m.Run() // execute the tests
	}()
	os.Exit(exitCode)
}

func TestSampleMongo(t *testing.T) {
	t.Logf("Executing simple Mongo test with configuration: %+v", mongoDockerInstance)
}

func TestSampleElastic(t *testing.T) {
	t.Logf("Executing simple Elastic test with configuration: %+v", esDockerInstance)
}

func TestSamplePostgres(t *testing.T) {
	t.Logf("Executing simple Postgres test with configuration: %+v", postgresDockerInstance)
	postgresDockerInstance.MustTruncateData()
}
