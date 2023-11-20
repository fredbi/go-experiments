package commands

import (
	"fmt"
	stdlog "log"
	"os"
	"strings"

	"github.com/fredbi/go-cli/cli"
	"github.com/fredbi/go-cli/cli/cli-utils/resolve"
	"github.com/fredbi/go-cli/cli/injectable"
	"github.com/fredbi/go-cli/config"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/db/migrations"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/consumer"
	natsembedded "github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/producer"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos/pgrepo"
	"github.com/fredbi/go-trace/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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
	//
	// NOTE: in this demo, the deployment context "env" is not used.
	// This demo doesn't use merged secrets. Apply LoadWithSecrets to merge extra secret config.
	cfg, err := config.Load(env,
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
			Use:          "daemon",
			Short:        "a message processing client",
			Run:          func(_ *cobra.Command, _ []string) {},
			SilenceUsage: true,
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
		cli.WithFlag("participant-id", "",
			"identifies the logical routing ID of this daemon to the messaging cluster. A random name is chosen by default",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.AppConfig, configkeys.ParticipantID)),
		),
		cli.WithFlag("nats-url", natsembedded.DefaultSettings.URL,
			"URL to the messaging cluster. NATS defaults is to listen on localhost:4222",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "url")),
		),
		cli.WithFlag("nats-cluster-id", "messaging",
			"NATS cluster ID",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "server", "clusterId")),
		),
		cli.WithFlag("nats-cluster-url", "nats://localhost:5333",
			"NATS cluster advertised discovery URL",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "server", "clusterURL")),
		),
		cli.WithFlag("nats-cluster-routes", "",
			"NATS cluster discovery routes to other nodes, as a comma separated list of URLs",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "server", "clusterRoutes")),
		),
		cli.WithFlag("nats-cluster-headless-service", "",
			"NATS cluster discovery resolved from DNS lookup on a kubernetes headless service (overriden by any value set with nats-cluster-routes",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "server", "clusterHeadlessService")),
		),
		cli.WithFlag("nats-postings-topic", "postings",
			"NATS prefix for postings subjects (producers post there, consumers listen to that)",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "topics", "postings")),
		),
		cli.WithFlag("nats-results-topic", "results",
			"NATS prefix for results subjects (consumers post there, producers listen to that",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(inSection(configkeys.NatsConfig, "topics", "results")),
		),
		cli.WithInjectables(
			// inject root logger
			injectable.NewZapLogger(zlog),
		),
		cli.WithSubCommands(
			// consumer sub-command
			cli.NewCommand(
				&cobra.Command{
					Use:     "consumer",
					Short:   "a message consumer client",
					PreRunE: prerun,
					RunE:    consumerCommand,
				},
				cli.WithFlag("no-replay", false,
					"disable background message redeliveries by consumer",
					cli.BindFlagToConfig(inSection(configkeys.AppConfig, "consumer", "noreplay")),
				),
			),
			// producer sub-command
			cli.NewCommand(
				&cobra.Command{
					Use:     "producer",
					Short:   "a message producer client",
					PreRunE: prerun,
					RunE:    producerCommand,
				},
				cli.WithFlag("port", 9090,
					"port to serve http requests",
					cli.BindFlagToConfig(inSection(configkeys.AppConfig, "producer.api.port")),
				),
				cli.WithFlag("no-replay", false,
					"disable background message redeliveries by producer",
					cli.BindFlagToConfig(inSection(configkeys.AppConfig, "producer", "noreplay")),
				),
			),
			// additional help topics
			cli.NewCommand(
				&cobra.Command{
					Use:   "config",
					Short: "prints default settings as a YAML configuration file",
				},
				cli.WithCobraOptions(func(helpConfig *cobra.Command) {
					helpConfig.SetHelpFunc(func(_ *cobra.Command, _ []string) {
						cfg := viper.New()
						repoCfg := pgrepo.DefaultSettings().AllSettings()
						cfg.Set(configkeys.DBConfig, repoCfg)

						natsCfg := natsembedded.DefaultSettingsNATS().AllSettings()
						cfg.Set(configkeys.NatsConfig, natsCfg)

						migrationsCfg := migrations.DefaultSettings().AllSettings()
						cfg.Set(configkeys.MigrationsConfig, migrationsCfg)

						p := producer.DefaultSettings()
						consumerCfg := consumer.DefaultSettings().AllSettings()
						_ = p.MergeConfigMap(consumerCfg)
						appCfg := p.AllSettings()
						cfg.Set(configkeys.AppConfig, appCfg)

						asYAML, _ := yaml.Marshal(cfg.AllSettings())

						fmt.Println("--\n# default settings\n\n" + string(asYAML)) //nolint:forbidigo
					})
				}),
			),
		),
	)
}

// prerun is run prior to any command.
//
// It initializes a logger and makes sure the database is created and up to date.
func prerun(c *cobra.Command, _ []string) error {
	_, lg, cfg := resolve.InjectedZapConfig(c,
		resolve.WithZapLoggerDefaulter(zap.NewNop),
		resolve.WithConfigDefaulter(configkeys.DefaultConfig),
	)

	// 1. Config
	// Debug by config if requested from env
	if dumpConfig() {
		stdlog.Println(cfg.AllSettings())
	}

	// 2. Logging
	// Re-level logger to account for possible CLI flag settings
	ctx, zlg := resolve.RelevelInjectedZapLogger(c, cfg.GetString(configkeys.LogLevel))

	lg.Info("starting app",
		zap.String("app_name", appName),
		zap.Stringer("log_level", zlg.Level()),
	)

	// 3. Database
	// Create DB if needed
	// NOTE: this is transactional, so only the first started deployment will get there.
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

// helper functions

func inSection(section string, keys ...string) string {
	return strings.Join(append([]string{section}, keys...), ".")
}

// dumpConfig	prints out all settings, for debugging config issues
func dumpConfig() bool {
	return os.Getenv("DUMP_CONFIG") != ""
}
