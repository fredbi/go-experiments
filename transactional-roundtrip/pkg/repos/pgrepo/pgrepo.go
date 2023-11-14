package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/db/migrations"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	messages "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pg-messages"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pgrepo/config-keys"
	"github.com/fredbi/go-trace/log"
	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/jmoiron/sqlx"
	"github.com/opencensus-integrations/ocsql"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const driverName = "pgx"

var _ repos.Iface = &Repository{}

type (
	// Repository knows how to handle a postgres backend database.
	//
	// The database driver is instrumented for tracing.
	Repository struct {
		db  *sqlx.DB // master instance
		log log.Factory

		*config
		messageRepo repos.MessageRepo
		app         string
	}

	config struct {
		parentCfg *viper.Viper
		cfg       *viper.Viper
		poolCfg   *viper.Viper
	}
)

// NewRepository creates a new postgres repository.
//
// The new repository needs to be started wih Start() in order to create the connection pools.
func NewRepository(app string, lg log.Factory, cfg *viper.Viper) *Repository {
	return &Repository{
		app:    app,
		log:    lg,
		config: newConfig(cfg, configkeys.DefaultDB),
	}
}

func newConfig(cfg *viper.Viper, db string) *config {
	if cfg == nil {
		cfg = viper.New()
	}

	// databases.(...)
	allDBConfig := cfg.Sub(configkeys.DBConfig)
	if allDBConfig == nil {
		allDBConfig = configkeys.DefaultAllDBConfig()
	}

	// databases.postgres.(...)
	pgConfig := allDBConfig.Sub(configkeys.PGConfig)
	if pgConfig == nil {
		pgConfig = configkeys.DefaultPGConfig()
	}

	// databases.postgres.config.(...)
	defaultConfig := pgConfig.Sub(configkeys.DBConfig)
	if defaultConfig == nil {
		defaultConfig = configkeys.DefaultPoolConfig()
	}

	// databases.postgres.{db name}.(...)
	dbConfig := pgConfig.Sub(db)
	if dbConfig == nil {
		dbConfig = configkeys.DefaultDBConfig()
	}

	// databases.postgres.default.config.(...)
	specificConfig := dbConfig.Sub(configkeys.DBConfig)
	if specificConfig == nil {
		specificConfig = defaultConfig
	}

	configkeys.SetDefaults(specificConfig)

	return &config{
		parentCfg: cfg,
		cfg:       dbConfig,
		poolCfg:   specificConfig,
	}
}

func (r *Repository) Messages() repos.MessageRepo {
	if r.messageRepo == nil {
		panic("dev error: Repository not started yet")
	}
	return r.messageRepo
}

// DB master instance
func (r *Repository) DB() *sqlx.DB {
	return r.db
}

// Logger returns a logger factory
func (r Repository) Logger() log.Factory {
	return r.log
}

// Config returns a configuration registry
func (r Repository) Config() *viper.Viper {
	return r.cfg
}

func (r Repository) open(dcfg *pgx.ConnConfig) (*sqlx.DB, error) {
	addr := stdlib.RegisterConnConfig(dcfg)
	lg := r.log.Bg()
	lg.Debug("registered driver", zap.String("driver", driverName), zap.String("config", addr), zap.String("db", dcfg.Database))

	opts := r.config.TraceOptions(dcfg.ConnString())
	instrumentedDriver := driverName

	if len(opts) > 0 {
		lg.Info("trace enabled for sql driver", zap.String("db", dcfg.Database))

		// opencensus tracing registered in the sql driver
		// (this wraps the sql driver with an instrumented version)
		var err error
		instrumentedDriver, err = ocsql.RegisterWithSource(driverName, addr, opts...)
		if err != nil {
			lg.Error("failed to register trace driver", zap.Error(err))
			return nil, err
		}

		lg.Debug("registered instrumented driver", zap.String("driver", instrumentedDriver))
	}

	db, err := sql.Open(instrumentedDriver, addr)
	if err != nil {
		return nil, err
	}

	err = waitPing(db, r.poolCfg.GetDuration(configkeys.PGPingTimeout))
	if err != nil {
		return nil, err
	}

	// connection pool settings
	r.config.SetPool(db)

	lg.Info("db pool settings",
		zap.String("driver", driverName),
		zap.Int("maxIdleConns", r.poolCfg.GetInt(configkeys.PGMaxIdleConns)),
		zap.Int("maxOpenConns", r.poolCfg.GetInt(configkeys.PGMaxOpenConns)),
		zap.Duration("connMaxLifetime", r.poolCfg.GetDuration(configkeys.PGConnMaxLifetime)),
	)
	return sqlx.NewDb(db, driverName), nil
}

// Start a connection pool to a database, plus possibly another one to the read-only version of it
func (r *Repository) Start() error {
	l := r.log.Bg()

	err := r.config.Validate()
	if err != nil {
		return err
	}

	connCfg := r.config.ConnConfig(r.config.DBURL(), r.log, r.app)
	r.db, err = r.open(connCfg)
	if err != nil {
		return err
	}

	// use the postgres implementation of the messageRepo
	r.messageRepo = messages.New(r.db, r.log, r.parentCfg)

	l.Info("connection pool ok", zap.String("db", connCfg.Database))

	return nil
}

// Stop the repository and close all connection pools.
//
// Stop may be called safely even if the database connection failed to start properly.
func (r *Repository) Stop() error {
	if r.db == nil {
		return nil
	}

	return r.db.Close()
}

// HealthCheck pings the database
func (r *Repository) HealthCheck() error {
	if r.db == nil {
		return errors.New("db not initialized")
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), r.cfg.GetDuration(configkeys.PGPingTimeout))
	defer cancel()

	return r.db.PingContext(ctxTimeout)
}

// waitPing checks for the availability of the database connection for maxWait.
//
// If the database is not immediately available, it tries every second up to maxWait.
//
// This avoids a hard container restart when the database is not immediatly available
// (e.g. when a db proxy container is not ready yet).
func waitPing(db *sql.DB, maxWait time.Duration) (err error) {
	if maxWait < time.Second {
		maxWait = time.Second
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), maxWait)
	defer cancel()

	err = db.PingContext(ctxTimeout)
	if err == nil {
		return
	}

	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err = db.PingContext(ctxTimeout)
			if err == nil {
				return
			}
		case <-timer.C:
			return db.PingContext(ctxTimeout)
		}
	}
}

func sqlDefaultTraceOptions() []ocsql.TraceOption {
	// _almost_ WithAllTraceOptions: just remove the WithRowsNext and Ping which produce a lot of clutter in traces
	return []ocsql.TraceOption{
		ocsql.WithAllowRoot(true),
		ocsql.WithLastInsertID(true),
		ocsql.WithQuery(true),
		ocsql.WithQueryParams(true),
		ocsql.WithRowsAffected(true),
		ocsql.WithRowsClose(true),
	}
}

// Config returns a configuration registry
func (r config) Config() *viper.Viper {
	return r.cfg
}

// LogLevel returns a pgx log level from config
func (r config) LogLevel() tracelog.LogLevel {
	level, err := tracelog.LogLevelFromString(r.poolCfg.GetString(configkeys.PGLogLevel))
	if err != nil {
		level, _ = tracelog.LogLevelFromString(configkeys.DefaultPGLogLevel)
	}

	return level
}

// SetPool sets the connection pool parameters from config
func (r config) SetPool(db *sql.DB) {
	db.SetMaxIdleConns(r.poolCfg.GetInt(configkeys.PGMaxIdleConns))
	db.SetMaxOpenConns(r.poolCfg.GetInt(configkeys.PGMaxOpenConns))
	db.SetConnMaxLifetime(r.poolCfg.GetDuration(configkeys.PGConnMaxLifetime))
}

// TraceOptions returns the trace options for the opencensus driver wrapper
func (r config) TraceOptions(u string) []ocsql.TraceOption {
	if !r.poolCfg.GetBool(configkeys.PGTraceEnabled) {
		return nil
	}
	v, _ := url.Parse(u)

	return append(sqlDefaultTraceOptions(), ocsql.WithInstanceName(v.Redacted()))
}

func (r config) ConnConfig(u string, lg log.Factory, app string) *pgx.ConnConfig {
	// driver config with logs and tag for logs
	l := lg.Bg()

	rtParams := map[string]string{
		"application_name": app,
	}

	dcfg, e := pgx.ParseConfig(u)
	if e != nil {
		return nil
	}

	if user := os.ExpandEnv(r.cfg.GetString(configkeys.PGUser)); user != "" {
		dcfg.User = user
	}

	if password := os.ExpandEnv(r.cfg.GetString(configkeys.PGPassword)); password != "" {
		dcfg.Password = password
	}

	if setCommands := r.poolCfg.GetStringMapString(configkeys.PGSet); len(setCommands) > 0 {
		// execute SET key = value commands when the connection is established
		for k, v := range setCommands {
			l.Info("set command configured after db connect", zap.String("db_set_cmd", fmt.Sprintf(`SET %s = %s`, k, v)))
		}

		dcfg.AfterConnect = func(ctx context.Context, conn *pgconn.PgConn) error {
			for k, v := range setCommands {
				k = os.ExpandEnv(k)
				v = os.ExpandEnv(v)

				m := conn.Exec(ctx, fmt.Sprintf(`SET %s = %s`, k, v))
				_, err := m.ReadAll()
				if err != nil {
					return err
				}

				err = m.Close()
				if err != nil {
					return err
				}
			}

			return nil
		}
	}

	tr := &tracelog.TraceLog{
		Logger:   zapadapter.NewLogger(l.Zap().Named(fmt.Sprintf("pg-%s", app)).WithOptions(zap.AddCallerSkip(1))),
		LogLevel: r.LogLevel(),
	}
	dcfg.Tracer = tr
	dcfg.Config.RuntimeParams = rtParams

	tr.Logger.Log(context.Background(),
		tracelog.LogLevelInfo, "db log level",
		map[string]interface{}{
			"log-level": tr.LogLevel.String(),
		})

	return dcfg
}

func (r config) DBURL() string {
	u := os.ExpandEnv(r.cfg.GetString(configkeys.PGURL))

	return u
}

func (r config) validateURL(value string) error {
	if value == "" {
		return fmt.Errorf(`connection string is required`)
	}

	_, err := url.Parse(value)

	return err
}

// Validate the configuration
func (r config) Validate() error {
	if r.cfg == nil || r.poolCfg == nil {
		return fmt.Errorf("missing config")
	}

	if err := r.validateURL(r.DBURL()); err != nil {
		return err
	}

	_, err := pgx.ParseConfig(r.DBURL())
	if err != nil {
		return fmt.Errorf("invalid connection string: %s", err)
	}

	lvl := r.poolCfg.GetString(configkeys.PGLogLevel)
	if _, err := tracelog.LogLevelFromString(lvl); err != nil {
		return fmt.Errorf("invalid log level for pgx driver [%q]: %w", lvl, err)
	}

	return nil
}

func EnsureDB(ctx context.Context, cfg *viper.Viper, dbName string) (db *sqlx.DB, created bool, err error) {
	dbConfig := newConfig(cfg, dbName)
	dsn := dbConfig.cfg.GetString(configkeys.PGURL)
	if dsn == "" {
		return nil, false, fmt.Errorf("no database URL found in config file. Expected %s in %s section", configkeys.PGURL, dbName)
	}

	// connect without DB (TODO: retry, like for pgrepo)
	// TODO: this should not be part of migrations, but of some pg utils package
	created, err = migrations.CreateDB(ctx, dsn)
	if err != nil {
		return nil, true, err
	}

	db, err = sqlx.Open(driverName, dsn)

	return db, created, err
}
