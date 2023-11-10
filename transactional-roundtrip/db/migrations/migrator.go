package migrations

import (
	"embed"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/fredbi/gooseplus"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

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

func New(db *sqlx.DB, logger *zap.Logger) *Migrator {
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

func CreateDB(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("DB URL is invalid: %w", err)
	}

	dbName := path.Base(u.Path)
	u.Path = path.Dir(u.Path)

	x, err := sqlx.Open(dbDriver, u.String())
	if err != nil {
		return fmt.Errorf("could not connect to database server %v: %w", u, err)
	}
	defer func() {
		// will restart from a fresh pool
		_ = x.Close()
	}()

	_, err = x.Exec(
		fmt.Sprintf(`CREATE DATABASE %s IF NOT EXISTS`, dbName),
	)
	if err != nil {
		return fmt.Errorf("could not ensure database %s is created", dbName)
	}

	return nil
}
