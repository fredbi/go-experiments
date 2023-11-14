package migrations

import (
	"embed"
	"time"

	"github.com/fredbi/gooseplus"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	// initializes pgx driver
	_ "github.com/jackc/pgx/v5"
)

const (
	dbDialect = "postgres"
	dbDriver  = "pgx"
)

//go:embed sql/*.sql
var embeddedMigrations embed.FS

type Migrator struct {
	*gooseplus.Migrator
}

// New database migrator, using github.com/fredbi/gooseplus.
//
// Suited for a postgres DB, uses embedded FS to store migrations.
func New(db *sqlx.DB, logger *zap.Logger) *Migrator {
	// TODO: inject config
	return &Migrator{
		Migrator: gooseplus.New(
			db.DB,
			gooseplus.WithDialect(dbDialect),
			gooseplus.SetEnvironments(nil), // disable env folders
			gooseplus.WithFS(embeddedMigrations),
			gooseplus.WithLogger(logger),
			gooseplus.WithGlobalLock(true),
			gooseplus.WithTimeout(5*time.Minute),
		),
	}
}
