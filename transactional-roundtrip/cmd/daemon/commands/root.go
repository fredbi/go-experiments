package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/fredbi/go-cli/cli"
	"github.com/fredbi/go-cli/cli/injectable"
	"github.com/fredbi/go-cli/config"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/db/migrations"
	"github.com/fredbi/go-trace/log"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5"
)

const (
	// the name of this app.
	appName  = "daemon"
	dbDriver = "pgx"
)

func Root() *cli.Command {
	// load config
	cfg, err := config.LoadWithSecrets("",
		config.WithSearchParentDir(true),
		config.WithMute(true),
	)
	if err != nil {
		cli.Die("unable to load configuration file: %v", err)
	}

	// root structured zap logger
	zlog, closer := log.MustGetLogger(appName,
		log.WithLevel(cfg.GetString("log.level")),
	)

	return cli.NewCommand(
		&cobra.Command{
			Use:               "consumer",
			Short:             "a message consumer client",
			RunE:              root,
			PersistentPostRun: func(_ *cobra.Command, _ []string) { closer() },
		},
		cli.WithConfig(cfg),
		cli.WithFlag("log-level", "info", "controls logging verbosity",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(configkeys.LogLevel),
		),
		cli.WithInjectables(
			injectable.NewZapLogger(zlog),
		),
		cli.WithSubCommands(
			cli.NewCommand(
				&cobra.Command{
					Use:   "consumer",
					Short: "a message consumer client",
					RunE:  consumer,
				},
			),
			cli.NewCommand(
				&cobra.Command{
					Use:   "producer",
					Short: "a message producer client",
					RunE:  producer,
				},
			),
		))
}

func root(c *cobra.Command, _ []string) error {
	// resolve dependencies and start the server
	ctx, zlg, cfg := resolveInjected(c)
	zlg.Info("starting app", zap.String("app_name", appName))

	// open DB, with create DB if needed
	db, err := ensureDB(cfg, true)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	// apply DB migrations
	m := migrations.New(db, zlg)

	return m.Migrate(ctx)
}

func resolveInjected(c *cobra.Command) (context.Context, *zap.Logger, *viper.Viper) {
	ctx := c.Context()
	zlg := injectable.ZapLoggerFromContext(ctx, zap.NewNop)
	cfg := injectable.ConfigFromContext(ctx, func() *viper.Viper {
		cfg := viper.New()
		configkeys.SetDefaults(cfg)

		return cfg
	})

	return ctx, zlg, cfg
}

func ensureDB(cfg *viper.Viper, withCreate bool) (*sqlx.DB, error) {
	// TODO(fredbi): move to migrations pkg

	dbConfig := cfg.Sub(configkeys.DBConfig)
	if dbConfig == nil {
		return nil, fmt.Errorf("no database configuration found in config file. Expect a %s section", configkeys.DBConfig)
	}

	dsn := dbConfig.GetString(configkeys.DSN)

	if withCreate {
		// connect without DB
		if err := migrations.CreateDB(dsn); err != nil {
			return nil, err
		}
	}

	// TODO: move this to service using it
	return sqlx.Open(dbDriver, dsn)
}
