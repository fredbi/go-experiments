package commands

import (
	"time"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/produce"
	"github.com/goombaio/namegenerator"
	"github.com/spf13/cobra"
)

const daemonName = "producer"

func producer(c *cobra.Command, _ []string) error {
	ctx, zlg, cfg := resolveInjected(c)
	lg := log.NewFactory(zlg.Named(daemonName))

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
	// serve producer
	generator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
	id := namer.Generate()
	producer := produce.New(rt, id)

	return producer.Start()
}
