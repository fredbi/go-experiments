// go:!build
package main

import (
	"log"
	"path/filepath"

	"github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands"
	"github.com/spf13/cobra/doc"
)

func main() {
	cmd := commands.Root().Command
	if err := doc.GenMarkdownTree(cmd, filepath.Join("..", "docs")); err != nil {
		log.Fatal(err)
	}
}
