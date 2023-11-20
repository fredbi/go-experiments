package commands

import (
	"fmt"

	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/consumer"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/nats"
	"github.com/spf13/cobra"
)

func consumerCommand(c *cobra.Command, _ []string) error {
	rt, err := newRuntimeForCommand(c)
	if err != nil {
		return err
	}

	// 1. Start a NATS embedded server in the background
	server := nats.New(rt)
	if err = server.Start(); err != nil {
		return fmt.Errorf("could not start NATS embedded server: %w", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	// 2. Start a producer client
	consumer := consumer.New(rt, rt.ID())
	defer func() {
		_ = rt.Close()
	}()

	return consumer.Start()
}
