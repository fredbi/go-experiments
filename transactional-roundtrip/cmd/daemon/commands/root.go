package commands

import (
	stdlog "log"
	"os"
	"strings"

	"github.com/fredbi/go-cli/cli"
	"github.com/fredbi/go-cli/cli/cli-utils/resolve"
	"github.com/fredbi/go-cli/cli/injectable"
	"github.com/fredbi/go-cli/config"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/db/migrations"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pgrepo"
	"github.com/fredbi/go-trace/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	// the name of this app.
	appName = "daemon"

	env = ""
)

func Root() *cli.Command {
	// zap structured logger with debug level: the prerun() will
	// set up the appropriate level filtering.
	zlog, closer := log.MustGetLogger(
		appName,
		log.WithLevel("debug"),
		log.WithRedirectStdLog(true), // redirects global standard log to this logger
		log.WithCallerSkip(0),
	)

	// load config from file or environment variables
	cfg, err := config.LoadWithSecrets(env,
		config.WithSearchParentDir(true),
		config.WithMute(!dumpConfig()),
		config.WithOutput(stdlog.Writer()),
	)
	if err != nil {
		cli.Die("unable to load configuration file: %v", err)
	}
	if cfg == nil {
		zlog.Warn("no configuration file found. Defaults apply")
	}

	return cli.NewCommand(
		&cobra.Command{
			Use:               "daemon",
			Short:             "a message processing client",
			Run:               func(_ *cobra.Command, _ []string) {},
			SilenceUsage:      true,
			PersistentPreRunE: prerun,
			PersistentPostRun: func(_ *cobra.Command, _ []string) {
				// sync the logger output
				closer()
			},
		},
		cli.WithConfig(cfg), // inject config file

		// root flags are inherited by sub-commands. All CLI flags are bound to config keys.
		//
		// Settings are resolved with the precedence: {flags} > {env variable} > {config file}
		cli.WithFlag("log-level", "info",
			"controls logging verbosity",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(configkeys.LogLevel),
		),
		cli.WithFlag("participant_id", "",
			"identifies the logical routing ID of this daemon to the messaging cluster. A random name is chosen by default",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.AppConfig, configkeys.ParticipantID)),
		),
		cli.WithFlag("nats-url", "",
			"URL to the messaging cluster",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "url")),
		),
		cli.WithFlag("nats-cluster-id", "",
			"NATS cluster ID",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "clusterId")),
		),
		cli.WithFlag("nats-postings-topic", "postings",
			"NATS prefix for postings subjects",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "postings")),
		),
		cli.WithFlag("nats-results-topic", "results",
			"NATS prefix for results subjects",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "results")),
		),
		cli.WithInjectables(
			// inject root logger
			injectable.NewZapLogger(zlog),
		),
		cli.WithSubCommands(
			// consumer sub-command
			cli.NewCommand(
				&cobra.Command{
					Use:   "consumer",
					Short: "a message consumer client",
					RunE:  consumerCommand,
				},
			),
			// producer sub-command
			cli.NewCommand(
				&cobra.Command{
					Use:   "producer",
					Short: "a message producer client",
					RunE:  producerCommand,
				},
				cli.WithFlag("port", "9090",
					"port to serve http requests",
					cli.BindFlagToConfig(inSection(configkeys.AppConfig, "producer.api.port")),
				),
			),
		))
}

// prerun is run prior to any command.
//
// It initializes a logger and makes sure the database is created and up to date.
func prerun(c *cobra.Command, _ []string) error {
	_, _, cfg := resolve.InjectedZapConfig(c,
		resolve.WithZapLoggerDefaulter(zap.NewNop),
		resolve.WithConfigDefaulter(configkeys.DefaultConfig),
	)

	// 1. Config
	// Debug config if requested
	if dumpConfig() {
		stdlog.Println(cfg.AllSettings())
	}

	// 2. Logging
	// Re-level logger to account for possible CLI flag settings
	ctx, zlg := resolve.RelevelInjectedZapLogger(c, cfg.GetString(configkeys.LogLevel))

	zlg.Info("starting app",
		zap.String("app_name", appName),
		zap.Stringer("log_level", zlg.Level()),
	)

	// 3. Database
	// Create DB if needed
	db, _, err := pgrepo.EnsureDB(ctx, cfg, zlg, pgrepo.DefaultDB)
	if err != nil {
		return err
	}
	defer func() {
		// scratch this pool after the migrations have completed
		_ = db.Close()
	}()

	// 4. Apply DB migrations
	rt := runtime{db: db, logger: log.NewFactory(zlg), cfg: cfg}
	m := migrations.New(rt)

	return m.Migrate(ctx)
}

func inSection(section string, keys ...string) string {
	return strings.Join(append([]string{section}, keys...), ".")
}

// dumpConfig	prints out all settings, for debugging config issues
func dumpConfig() bool {
	return os.Getenv("DUMP_CONFIG") != ""
}
