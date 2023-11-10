package commands

import (
	"github.com/fredbi/go-trace/log"
	"github.com/spf13/cobra"
)

func consumer(c *cobra.Command, _ []string) error {
	ctx, zlg, cfg := resolveInjected(c)
	lg := log.NewFactory(zlg.Named("consumer"))

	// open DB
	db, err := ensureDB(cfg, false)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	rt := runtime{
		db:     db,
		logger: lg,
		cfg:    cfg,
	}

	// serve consumer
	return nil
}
