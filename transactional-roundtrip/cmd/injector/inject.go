package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/fredbi/go-cli/cli/cli-utils/resolve"
	configkeys "github.com/fredbi/go-experiments/transactional-roundtrip/cmd/daemon/commands/config-keys"
	"github.com/fredbi/go-experiments/transactional-roundtrip/pkg/repos"
	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const amountMax = int64(1_000_000_000) // max precision in DB column is numeric(15,2)

//go:embed fixtures/*.ndjson
var embeddedFixtures embed.FS

func inject(c *cobra.Command, files []string) error {
	parentCtx, zlg, cfg := resolve.InjectedZapConfig(c,
		resolve.WithZapLoggerDefaulter(zap.NewNop),
		resolve.WithConfigDefaulter(configkeys.DefaultConfig),
	)
	zlg = zlg.Named(c.Name())
	client := http.DefaultClient

	repeat := cfg.GetInt("injector.repeat")
	randomize := repeat > 1
	baseURL := cfg.GetString("injector.url")
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "message")
	target := u.String()
	zlg.Info("posting to URL", zap.String("target_url", target))

	// pick from local disk or from embedded fixtures
	var fsys fs.FS
	if cfg.GetBool("injector.UseEmbedded") {
		fsys = embeddedFixtures
	} else {
		fsys = os.DirFS(".")
	}

	for _, file := range files {
		f, err := fsys.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open input: %w", err)
		}

		r := bufio.NewScanner(f)
		for r.Scan() { // scan a JSON line from the input fixture
			var inputMsg repos.InputPayload
			if err := json.Unmarshal(r.Bytes(), &inputMsg); err != nil {
				return fmt.Errorf("failed to unmarshal json content: %w (%s)", err, r.Text())
			}

			for n := 0; n < repeat; n++ {
				if randomize {
					// shuffle content
					inputMsg.Amount = *repos.NewDecimal(rand.Int63n(amountMax), 2) //#nosec
				}

				body := new(bytes.Buffer)
				enc := json.NewEncoder(body)
				if err := enc.Encode(inputMsg); err != nil {
					return fmt.Errorf("failed to marshal json content: %w", err)
				}

				ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
				if err != nil {
					cancel()

					return fmt.Errorf("failed to build request: %w", err)
				}

				resp, err := client.Do(req)
				if err != nil {
					cancel()

					return err
				}
				cancel()

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("request was not successful: %d", resp.StatusCode)
				}
			}
		}

		if err := r.Err(); err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}

	}

	return nil
}
