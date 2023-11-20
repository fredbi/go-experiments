package main

import (
	"fmt"
	"io/fs"

	"github.com/fredbi/go-cli/cli"
	"github.com/fredbi/go-cli/cli/cli-utils/resolve"
	"github.com/fredbi/go-cli/cli/injectable"
	"github.com/fredbi/go-cli/config"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-trace/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Injector() *cli.Command {
	zlog, closer := log.MustGetLogger(
		"injector",
		log.WithLevel("debug"),
		log.WithCallerSkip(0),
	)

	cfg, err := config.Load("",
		config.WithSearchParentDir(true),
		config.WithMute(true),
	)
	if err != nil {
		cli.Die("unable to load configuration file: %v", err)
	}
	if cfg == nil {
		zlog.Warn("no configuration file found. Defaults apply")
	}

	return cli.NewCommand(
		&cobra.Command{
			Use:          "injector [files...]",
			Short:        "an API message injector (http client)",
			Run:          func(_ *cobra.Command, _ []string) {},
			SilenceUsage: true,
			PersistentPreRun: func(c *cobra.Command, _ []string) {
				_, _, injectedConfig := resolve.InjectedZapConfig(c,
					resolve.WithZapLoggerDefaulter(zap.NewNop),
					resolve.WithConfigDefaulter(configkeys.DefaultConfig),
				)
				_, _ = resolve.RelevelInjectedZapLogger(c, injectedConfig.GetString(configkeys.LogLevel))
			},
			PersistentPostRun: func(_ *cobra.Command, _ []string) {
				closer()
			},
			RunE: inject,
			Args: cobra.MinimumNArgs(1),
		},
		cli.WithConfig(cfg),
		cli.WithFlag("log-level", "info",
			"controls logging verbosity",
			cli.FlagIsPersistent(),
			cli.BindFlagToConfig(configkeys.LogLevel),
		),
		cli.WithFlag("target-base-url", "http://localhost:9090/",
			"where sent http requests",
			cli.BindFlagToConfig("injector.url"),
		),
		cli.WithFlag("repeat-randomized", 1000,
			"repeat the provided input n times with randomized content",
			cli.BindFlagToConfig("injector.repeat"),
		),
		cli.WithFlag("with-embedded-fixtures", true,
			"use built-in fixtures",
			cli.BindFlagToConfig("injector.UseEmbedded"),
		),
		cli.WithInjectables(
			injectable.NewZapLogger(zlog),
		),
		cli.WithSubCommands(
			// additional help topics
			cli.NewCommand(
				&cobra.Command{
					Use:   "fixtures",
					Short: "prints available built-in fixtures",
				},
				cli.WithCobraOptions(func(helpConfig *cobra.Command) {
					helpConfig.SetHelpFunc(func(_ *cobra.Command, _ []string) {
						fmt.Println("# Available built-in fixtures:") //nolint:forbidigo
						_ = fs.WalkDir(embeddedFixtures, ".", func(pth string, entry fs.DirEntry, _ error) error {
							if entry.IsDir() {
								return nil
							}

							fmt.Println(pth) //nolint:forbidigo

							return nil
						})
					})
				}),
			),
		),
	)
}
