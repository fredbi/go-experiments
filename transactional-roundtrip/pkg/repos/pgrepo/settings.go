package pgrepo

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/fredbi/go-trace/log"
	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/opencensus-integrations/ocsql"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DefaultURL      = "postgresql://postgres@localhost:5432/testdb?sslmode=disable"
	DefaultDB       = "default"
	DefaultLogLevel = "info"
)

var defaultSettings = settings{
	Config: &poolSettings{
		MaxIdleConns:    25,
		MaxOpenConns:    50,
		ConnMaxLifeTime: 5 * time.Minute,
		Log: logSettings{
			Level: "warn",
		},
		Trace: traceSettings{
			Enabled: false,
		},
		PingTimeout: 10 * time.Second,
	},
	Databases: map[string]databaseSettings{
		"default": {
			URL: DefaultURL,
		},
	},
}

type (
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
	//	                 pingTimeout: 10s
	//	                 log:
	//	                   level: info
	//	                 trace:
	//	                  enabled: false
	//					     otherDB:
	//			           url: postgres://user:password@localhost:5433/other
	//					     config:  # pool settings for a postgres databases
	//				          maxIdleConns: 55
	//					   redis:
	//					     default:
	//					          ...
	//					     otherDB:
	settings struct {
		Config    *poolSettings
		Databases map[string]databaseSettings `mapstructure:"postgres"`
	}

	poolSettings struct {
		MaxIdleConns    int
		MaxOpenConns    int
		ConnMaxLifeTime time.Duration
		Log             logSettings
		Trace           traceSettings
		PingTimeout     time.Duration
		Set             map[string]string //	plan_cache_mode: auto|force_custom_plan|force_generic_plan
	}

	logSettings struct {
		Level string
	}

	traceSettings struct {
		Enabled bool
	}

	databaseSettings struct {
		URL      string
		User     string
		Password string
		Config   *poolSettings
		// Replicas []string
	}
)

func makeSettings(cfg *viper.Viper, l *zap.Logger) (settings, error) {
	s := defaultSettings

	if cfg == nil {
		l.Warn("no config passed. Using defaults")

		return s, nil
	}

	allDBConfig := cfg.Sub("databases")
	if allDBConfig == nil {
		l.Warn("no databases section passed in config. Using defaults")

		return s, nil
	}

	if err := allDBConfig.Unmarshal(&s); err != nil {
		return s, err
	}

	return s, nil
}

func makeDBSettings(s settings, db string, l *zap.Logger) (databaseSettings, error) {
	dbConfig, ok := s.Databases[db]
	if !ok {
		return dbConfig, fmt.Errorf("non-default db %q is not configured", db)
	}

	if dbConfig.Config == nil {
		dbConfig.Config = s.Config
	}

	l.Info("database configured", zap.String("db_url", dbConfig.RedactedURL()))

	return dbConfig, nil
}

// LogLevel returns a pgx log level from config
func (r databaseSettings) LogLevel() tracelog.LogLevel {
	if r.Config == nil {
		level, _ := tracelog.LogLevelFromString(DefaultLogLevel)

		return level
	}

	level, err := tracelog.LogLevelFromString(r.Config.Log.Level)
	if err != nil {
		level, _ = tracelog.LogLevelFromString(DefaultLogLevel)
	}

	return level
}

// SetPool sets the connection pool parameters from config
func (r databaseSettings) SetPool(db *sql.DB) {
	if r.Config == nil {
		return
	}

	if r.Config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(r.Config.MaxIdleConns)
	}
	if r.Config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(r.Config.MaxOpenConns)
	}
	if r.Config.ConnMaxLifeTime > 0 {
		db.SetConnMaxLifetime(r.Config.ConnMaxLifeTime)
	}
}

// TraceOptions returns the trace options for the opencensus driver wrapper
func (r databaseSettings) TraceOptions(u string) []ocsql.TraceOption {
	if r.Config == nil {
		return nil
	}

	if !r.Config.Trace.Enabled {
		return nil
	}

	v, _ := url.Parse(u)

	return append(sqlDefaultTraceOptions(), ocsql.WithInstanceName(v.Redacted()))
}

func (r databaseSettings) ConnConfig(u string, lg log.Factory, app string) *pgx.ConnConfig {
	// driver settings with logs and tag for logs
	l := lg.Bg()

	rtParams := map[string]string{
		"application_name": app,
	}

	dcfg, e := pgx.ParseConfig(u)
	if e != nil {
		return nil
	}

	if user := os.ExpandEnv(r.User); user != "" {
		dcfg.User = user
	}

	if password := os.ExpandEnv(r.Password); password != "" {
		dcfg.Password = password
	}

	if r.Config == nil && len(r.Config.Set) > 0 {
		// execute SET key = value commands when the connection is established
		for k, v := range r.Config.Set {
			l.Info("set command configured after db connect", zap.String("db_set_cmd", fmt.Sprintf(`SET %s = %s`, k, v)))
		}

		dcfg.AfterConnect = func(ctx context.Context, conn *pgconn.PgConn) error {
			for k, v := range r.Config.Set {
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

func (r databaseSettings) DBURL() string {
	u := os.ExpandEnv(r.URL)

	return u
}
func (r databaseSettings) RedactedURL() string {
	v, _ := url.Parse(r.DBURL())

	return v.Redacted()
}

func (r databaseSettings) validateURL(value string) error {
	if value == "" {
		return fmt.Errorf(`connection string is required`)
	}

	_, err := url.Parse(value)

	return err
}

// Validate the configuration
func (r databaseSettings) Validate() error {
	if err := r.validateURL(r.DBURL()); err != nil {
		return err
	}

	_, err := pgx.ParseConfig(r.DBURL())
	if err != nil {
		return fmt.Errorf("invalid connection string: %s", err)
	}

	if r.Config != nil && r.Config.Log.Level != "" {
		lvl := r.Config.Log.Level
		if _, err := tracelog.LogLevelFromString(lvl); err != nil {
			return fmt.Errorf("invalid log level for pgx driver [%q]: %w", lvl, err)
		}
	}

	return nil
}
