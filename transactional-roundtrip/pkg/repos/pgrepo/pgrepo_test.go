package pgrepo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEnsureDB(t *testing.T) {
	zlg := zap.NewNop()
	t.Run("with default config", func(t *testing.T) {
		db, _, err := EnsureDB(context.Background(), nil, zlg, "unittest_db")
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Close())
	})
}
