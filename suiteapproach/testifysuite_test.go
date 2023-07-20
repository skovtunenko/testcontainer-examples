package suiteapproach

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/skovtunenko/testcontainer-examples/integrationtesting"
)

func TestDemoSuite(t *testing.T) {
	os.Setenv("RUN_INTEGRATION_TESTS", "yes, please!")
	suite.Run(t, &DemoPostgresSuite{})
}

type DemoPostgresSuite struct {
	integrationtesting.PostgresSuite
}

var _ suite.SetupAllSuite = &DemoPostgresSuite{}
var _ suite.TearDownAllSuite = &DemoPostgresSuite{}

func (suite *DemoPostgresSuite) TestExample1() {
	suite.T().Logf("Running example test, initialized postgres connection URL: %+v", suite.GetPostgresConnectionURL())
}

func (suite *DemoPostgresSuite) TestExample2() {
	suite.T().Logf("Running example test, initialized postgres connection URL: %+v", suite.GetPostgresConnectionURL())
}
