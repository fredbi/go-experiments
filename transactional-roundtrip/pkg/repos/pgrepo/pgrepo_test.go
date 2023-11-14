package pgrepo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureDB(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		db, _, err := EnsureDB(context.Background(), nil, "unittest_db")
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Close())
	})
}
