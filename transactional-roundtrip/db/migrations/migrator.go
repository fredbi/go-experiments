package migrations

import (
	"embed"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/injected"
	"github.com/fredbi/gooseplus"

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
func New(rt injected.Runtime) *Migrator {
	cfg := rt.Config()
	s, _ := makeSettings(cfg)

	return &Migrator{
		Migrator: gooseplus.New(
			rt.DB().DB,
			gooseplus.WithDialect(dbDialect),
			gooseplus.SetEnvironments(nil), // disable env folders
			gooseplus.WithFS(embeddedMigrations),
			gooseplus.WithLogger(rt.Logger().Bg().Zap()),
			gooseplus.WithGlobalLock(true),
			gooseplus.WithTimeout(s.Timeout),
			gooseplus.WithMigrationTimeout(s.MigrationTimeout),
		),
	}
}
