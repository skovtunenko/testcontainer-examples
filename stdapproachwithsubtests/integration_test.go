package stdapproachwithsubtests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skovtunenko/testcontainer-examples/postgresintegration"
)

func TestPostgresIntegrationTest(t *testing.T) {
	if postgresintegration.IsSkipIntegrationTest(t) {
		return
	}

	postgres, cleanupFn := postgresintegration.MustRunPostgresDockerContainer()
	defer cleanupFn()

	t.Run("TestExample1", func(t *testing.T) {
		defer postgres.TruncateDataInTest(t)

		t.Logf("Running example test 1, initialized postgres connection URL: %+v", postgres.ConnURL())
	})

	t.Run("TestExample2", func(t *testing.T) {
		defer postgres.TruncateDataInTest(t)

		_, err := postgres.PgxPool().Exec(context.Background(), "SELECT 1")
		require.NoError(t, err)
	})
}
