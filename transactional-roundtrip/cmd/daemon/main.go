package main

import (
	"github.com/fredbi/go-cli/cli"
	"github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands"
)

func main() {
	cli.MustOrDie(
		"daemon failed",
		commands.Root().Execute(),
	)
}
