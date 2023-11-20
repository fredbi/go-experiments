package pgrepo

import (
	"time"

	pgpool "github.com/fredbi/pgxutils/pgrepo"
	"github.com/spf13/viper"
)

const (
	DefaultURL = pgpool.DefaultURL
	DefaultDB  = pgpool.DefaultDBAlias
)

func init() {
	// override package-level defaults
	pgpool.SetDefaults(
		pgpool.WithDefaultPoolOptions(
			pgpool.WithMaxIdleConns(25),
			pgpool.WithMaxOpenConns(50),
			pgpool.WithConnMaxIdleTime(5*time.Minute),
			pgpool.WithConnMaxLifeTime(1*time.Hour),
			pgpool.WithPingTimeout(5*time.Second),
		),
	)
}

// DefaultSettings returns all defaults for this package as a viper register.
//
// This is primarily intended for documentation & help purpose.
func DefaultSettings() *viper.Viper {
	return pgpool.DefaultSettings()
}
