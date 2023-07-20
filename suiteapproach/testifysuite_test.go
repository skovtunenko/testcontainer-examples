package suiteapproach

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tcinfra "github.com/skovtunenko/testcontainer-examples"
)

func TestPostgreSQLSuite(t *testing.T) {
	suite.Run(t, &PostgreSQLSuite{})
}

type PostgreSQLSuite struct {
	suite.Suite
	postgresqlConfig  tcinfra.PostgreSQLConfig
	postgresqlCleanFn func()
}

func (suite *PostgreSQLSuite) SetupSuite() {
	r := suite.Require()
	conf, cleanFn, err := tcinfra.RunPostgreSQLDockerContainer()
	r.NoError(err)
	suite.postgresqlConfig = conf
	suite.postgresqlCleanFn = cleanFn
}

func (suite *PostgreSQLSuite) TearDownSuite() {
	suite.postgresqlCleanFn()
}

var _ suite.SetupAllSuite = &PostgreSQLSuite{}
var _ suite.TearDownAllSuite = &PostgreSQLSuite{}

func (suite *PostgreSQLSuite) TestExample() {
	suite.T().Logf("Running example test, initialized postgreSQL config: %+v", suite.postgresqlConfig)
}
