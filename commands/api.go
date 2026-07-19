package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var data, host string
		var query []string
		cmd := &cobra.Command{
			Use:   "api <METHOD> <PATH> [--host search|api] [-d body] [-q key=value ...]",
			Short: "Send a raw Torre request (escape hatch)",
			Long: `Call any Torre endpoint directly. --host selects which base the PATH is relative to:
"api" (default, https://torre.ai/api — genome, opportunity detail) or "search"
(https://search.torre.co — the _search endpoints).

This is the documented escape hatch for anything torre does not wrap as a first-class
command. It honors --dry-run, -o/--output, and --jq like every other command. Non-GET
methods are never auto-retried.`,
			Example: `  torre api GET genome/bios/torrenegra
  torre api GET suite/opportunities/KWN4QjAd
  torre api POST opportunities/_search/ --host search -q size=5 -d '{"and":[{"remote":{"term":true}}]}'`,
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				method := strings.ToUpper(args[0])
				valid := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
				if !slices.Contains(valid, method) {
					return fmt.Errorf("invalid method %q (want one of %s)", args[0], strings.Join(valid, "|"))
				}
				if host != "api" && host != "search" {
					return fmt.Errorf("invalid --host %q (want api|search)", host)
				}
				path := strings.TrimLeft(args[1], "/")

				q := url.Values{}
				for _, kv := range query {
					k, v, ok := strings.Cut(kv, "=")
					if !ok {
						return fmt.Errorf("invalid -q %q (want key=value)", kv)
					}
					q.Add(k, v)
				}

				var body []byte
				if data != "" {
					raw, err := readDataArg(cmd, data)
					if err != nil {
						return err
					}
					body = raw
				}

				c, _, err := d.getAPIClient()
				if err != nil {
					return err
				}
				status, _, respBody, err := c.Do(cmd.Context(), host, method, path, q, body)
				if err != nil {
					return err
				}
				if status == 0 { // dry-run
					return nil
				}
				if len(respBody) == 0 {
					if !d.gf.quiet {
						fmt.Fprintf(cmd.OutOrStdout(), "HTTP %d (empty body)\n", status)
					}
					return nil
				}
				if json.Valid(respBody) {
					return d.render(cmd, json.RawMessage(respBody), nil)
				}
				_, err = cmd.OutOrStdout().Write(respBody)
				return err
			},
		}
		cmd.Flags().StringVar(&host, "host", "api", "which Torre host: api|search")
		cmd.Flags().StringVarP(&data, "data", "d", "", "JSON body: inline, @file, or - for stdin")
		cmd.Flags().StringArrayVarP(&query, "query", "q", nil, "query parameter key=value (repeatable)")
		return annotate(cmd, kindWrite) // raw calls may mutate; the guard gates by METHOD (§3b.6)
	})
}
