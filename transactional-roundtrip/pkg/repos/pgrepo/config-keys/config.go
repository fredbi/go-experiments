package keys

import (
	"time"

	"github.com/spf13/viper"
)

// Database configuration keys.
//
// The repo system is designed to be multi-databases.
// The configuration is hierarchized like so:
//
//					 databases:
//					   postgres:
//					     default:
//			           url: postgres://localhost:5432/test
//		             user: $PG_USER
//		             password: $PG_PASSWORD
//					       config: # pool settings for this database
//				           maxIdleConns: 25
//				           maxOpenConns: 50
//				           connMaxLifetime: 5m
//	                pingTimeout: 10s
//	                log:
//	                  level: info
//	                trace:
//	                  enabled: false
//					     otherDB:
//			           url: postgres://user:password@localhost:5433/other
//					     config:  # pool settings for a postgres databases
//				          maxIdleConns: 55
//					   redis:
//					     default:
//					          ...
//					     otherDB:
const (
	// Global DB config section
	DBConfig = "databases"

	// PostgreSQL config section
	PGConfig = "postgres"

	// default DB section
	DefaultDB = "default"

	// database config (for default or any other configured DB)
	PGURL      = "url"
	PGReplicas = "replicas"
	PGUser     = "user"
	PGPassword = "password"

	// Default or db specific config section
	Config = "config"

	PGMaxIdleConns    = "maxIdleConns"
	PGMaxOpenConns    = "maxOpenConns"
	PGConnMaxLifetime = "connMaxLifetime"
	PGLogLevel        = "log.level"
	PGTraceEnabled    = "trace.enabled"
	PGSet             = "set" //	plan_cache_mode: auto|force_custom_plan|force_generic_plan
	PGPingTimeout     = "pingTimeout"

	DefaultPGLogLevel = "info"
	DefaultURL        = "postgresql://postgres@localhost:5432/testdb?sslmode=disable"
)

// DefaultAllDBConfig yields a default database config, suitable for local testing with Postgres.
//
// It provides defaults for the top-level database configuration:
//
//	databases:
//	  ...
func DefaultAllDBConfig() *viper.Viper {
	v := viper.New()
	v.SetDefault(DBConfig, DefaultPGConfig())

	return v
}

// DefaultPGConfig yields a default postgres database conf, suitable for local testing.
//
// It provides defaults for the postgres database configuration:
//
//	  databases:
//	    postgres:
//	      default:
//		       ...
func DefaultPGConfig() *viper.Viper {
	v := viper.New()
	v.SetDefault(PGConfig, map[string]interface{}{
		DefaultDB: DefaultDBConfig(),
	})

	return v
}

// DefaultDBConfig yields defaults for a postgres database conf, suitable for local testing.
//
// It provides defaults for the postgres database configuration:
//
//		  databases:
//		    postgres:
//		      default:
//			      ...
//	        config:
//			      ...
func DefaultDBConfig() *viper.Viper {
	v := viper.New()
	v.SetDefault(PGURL, DefaultURL)

	return v
}

// DefaultPoolConfig yields defaults for a connection pool.
//
// It provides defaults for the postgres database configuration:
//
//		 databases:
//		   postgres:
//		     default:
//		        config:
//	            maxIdleConns: 25
//		          ...
func DefaultPoolConfig() *viper.Viper {
	v := viper.New()
	SetDefaults(v)

	return v
}

// SetDefaults applies defaults for pool configuration settings.
func SetDefaults(cfg *viper.Viper) {
	cfg.SetDefault(PGMaxIdleConns, 25)
	cfg.SetDefault(PGMaxOpenConns, 50)
	cfg.SetDefault(PGConnMaxLifetime, "5m")
	cfg.SetDefault(PGLogLevel, DefaultPGLogLevel)
	cfg.SetDefault(PGTraceEnabled, false)
	cfg.SetDefault(PGPingTimeout, 10*time.Second)
}
