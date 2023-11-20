package pgrepo

import (
	"context"

	pgpool "github.com/fredbi/pgxutils/pgrepo"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func EnsureDB(ctx context.Context, cfg *viper.Viper, l *zap.Logger, dbName string) (db *sqlx.DB, created bool, err error) {
	withViper, err := pgpool.SettingsFromViper(cfg, pgpool.WithLogger(l))
	if err != nil {
		return nil, false, err
	}

	return pgpool.EnsureDB(ctx, dbName,
		pgpool.WithLogger(l),
		withViper,
	)
}
