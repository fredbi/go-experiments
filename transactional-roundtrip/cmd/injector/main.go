package main

import (
	"github.com/fredbi/go-cli/cli"
)

func main() {
	cli.MustOrDie(
		"daemon failed",
		Injector().Execute(),
	)
}
