package suiteapproach

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/skovtunenko/testcontainer-examples/integrationtesting"
)

func TestDemoSuite(t *testing.T) {
	// to enable integration tests, set this environment variable:
	os.Setenv(integrationtesting.IntegrationRunnerEnvVar, "yes, please!")
	suite.Run(t, &DemoPostgresSuite{})
}

type DemoPostgresSuite struct {
	integrationtesting.PostgresSuite
}

var _ suite.SetupAllSuite = &DemoPostgresSuite{}
var _ suite.TearDownAllSuite = &DemoPostgresSuite{}

func (suite *DemoPostgresSuite) TestExample1() {
	suite.T().Logf("Running example test 1, initialized postgres connection URL: %+v", suite.GetPostgresConnectionURL())

	ctx := context.Background()
	r := suite.Require()

	db := suite.PgxPool()

	tableNames := func() []string {
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
		rows, err := db.Query(ctx, query)
		r.NoError(err)
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var table string
			err := rows.Scan(&table)
			r.NoError(err)
			tables = append(tables, table)
		}
		r.NoError(rows.Err())
		return tables
	}()

	suite.T().Log("List of available DB tables:\n")
	for _, tableName := range tableNames {
		suite.T().Logf("  %s", tableName)
	}
}

func (suite *DemoPostgresSuite) TestExample2() {
	suite.T().Logf("Running example test 2, initialized postgres connection URL: %+v", suite.GetPostgresConnectionURL())
}
