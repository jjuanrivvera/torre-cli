// Package commands wires the cobra command tree. root.go owns the global flags, the shared
// Torre client factory, and the single render() path used by every command. The tree is
// built fresh per NewRootCmd() call so tests never leak flag state across cases.
package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
	"github.com/jjuanrivvera/torre-cli/internal/auth"
	"github.com/jjuanrivvera/torre-cli/internal/config"
	"github.com/jjuanrivvera/torre-cli/internal/output"
)

// globalFlags holds the persistent flag values for one command tree.
type globalFlags struct {
	outputFormat  string
	profile       string
	baseURL       string // overrides the app-API host (torre.ai/api)
	searchBaseURL string // overrides the search host (search.torre.co)
	dryRun        bool
	showToken     bool
	verbose       bool
	noColor       bool
	columns       []string
	quiet         bool
	jq            string

	// list flags (read by search commands)
	all   bool
	limit int
	size  int
}

// deps carries the per-tree state into every command builder.
type deps struct {
	gf *globalFlags

	// overridable in tests
	loadConfig func() (*config.Config, error)
	store      func() auth.Store
	// newClient builds the API client; tests inject one pointed at an httptest server.
	newClient func(searchBase, apiBase string, opts ...api.Option) *api.Client
	// out overrides where dry-run curls go (tests capture it; default os.Stdout).
	out io.Writer
}

func newDeps() *deps {
	return &deps{
		gf:         &globalFlags{},
		loadConfig: config.Load,
		store: func() auth.Store {
			dir, err := config.Dir()
			if err != nil {
				dir = "."
			}
			return auth.New(dir)
		},
		newClient: api.New,
	}
}

// NewRootCmd assembles the full command tree.
func NewRootCmd() *cobra.Command { return newRootCmd(newDeps()) }

// registrars build the resource commands; resource files append from init().
var registrars []func(d *deps) *cobra.Command

// metaRegistrars register the non-resource commands (auth, config, doctor, …).
var metaRegistrars []func(d *deps) *cobra.Command

func newRootCmd(d *deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "torre",
		Short: "A fast, scriptable CLI for Torre.ai jobs and profiles",
		Long: `torre is a read-only, agent-friendly client for the Torre.ai public API: search job
opportunities, fetch an opportunity's detail, search people, and pull a person's public
genome/bio — all with machine-first output (JSON/YAML/CSV, -o id, --jq) so an AI assistant
or a shell pipeline can consume it directly.

Torre's public endpoints need no credentials, so it works out of the box. A bearer token is
optional (torre auth login) for any endpoint that requires one.

Examples:
  torre jobs search --skill "golang" --remote -o json
  torre jobs search --skill "product design" --location Colombia --limit 50
  torre jobs get KWN4QjAd
  torre genome torrenegra --jq '.person.name'
  torre people search --skill "data science" --remote -o table`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if d.gf.outputFormat != "" && !output.Format(d.gf.outputFormat).Valid() {
				return fmt.Errorf("unknown output format %q (want table|json|yaml|csv|id)", d.gf.outputFormat)
			}
			if d.gf.profile != "" {
				if err := config.ValidateProfileName(d.gf.profile); err != nil {
					return err
				}
			}
			return nil
		},
	}
	registerGlobalFlags(root, d.gf)

	for _, build := range registrars {
		root.AddCommand(build(d))
	}
	for _, build := range metaRegistrars {
		root.AddCommand(build(d))
	}
	return root
}

func registerGlobalFlags(root *cobra.Command, gf *globalFlags) {
	pf := root.PersistentFlags()
	pf.StringVarP(&gf.outputFormat, "output", "o", "", "output format: table|json|yaml|csv|id")
	// Torre is a single fixed public API; a "profile" only scopes an optional stored token
	// and host overrides. --profile is the natural name here (no per-API rename needed).
	pf.StringVar(&gf.profile, "profile", "", "named profile to use")
	pf.StringVar(&gf.baseURL, "base-url", "", "override the Torre app-API host (default https://torre.ai/api)")
	pf.StringVar(&gf.searchBaseURL, "search-base-url", "", "override the Torre search host (default https://search.torre.co)")
	pf.BoolVar(&gf.dryRun, "dry-run", false, "print the equivalent curl and make no request")
	pf.BoolVar(&gf.showToken, "show-token", false, "reveal the bearer token in dry-run output")
	pf.BoolVarP(&gf.verbose, "verbose", "v", false, "verbose request logging (stderr)")
	pf.BoolVar(&gf.noColor, "no-color", false, "disable colored output")
	pf.StringSliceVar(&gf.columns, "columns", nil, "comma-separated columns to show")
	pf.BoolVar(&gf.quiet, "quiet", false, "suppress non-essential chatter")
	pf.StringVar(&gf.jq, "jq", "", "gojq expression applied to the response before rendering")

	pf.BoolVar(&gf.all, "all", false, "page through all results (search commands)")
	pf.IntVar(&gf.limit, "limit", 0, "max items to return across pages (search commands)")
	pf.IntVar(&gf.size, "size", 20, "results per page (search commands)")
}

// resolveProfile returns the active profile name and config.
func (d *deps) resolveProfile() (string, *config.Config, error) {
	cfg, err := d.loadConfig()
	if err != nil {
		return "", nil, err
	}
	return cfg.ResolveProfileName(d.gf.profile), cfg, nil
}

// getAPIClient builds a Torre client for the ACTIVE profile, honoring flag > env > config >
// default for both hosts, and wiring an optional keyring-backed bearer token.
func (d *deps) getAPIClient() (*api.Client, *config.Config, error) {
	profileName, cfg, err := d.resolveProfile()
	if err != nil {
		return nil, nil, err
	}
	prof, _ := cfg.Profile(profileName)

	apiBase := config.FirstNonEmpty(d.gf.baseURL, os.Getenv("TORRE_BASE_URL"), prof.APIBaseURL, api.DefaultAPIBaseURL)
	searchBase := config.FirstNonEmpty(d.gf.searchBaseURL, os.Getenv("TORRE_SEARCH_BASE_URL"), prof.SearchBaseURL, api.DefaultSearchBaseURL)
	for _, u := range []string{apiBase, searchBase} {
		if err := config.ValidateBaseURL(u); err != nil {
			return nil, nil, err
		}
	}

	opts := []api.Option{
		api.WithDryRun(d.gf.dryRun, d.stdout()),
		api.WithUserAgent(os.Getenv("TORRE_USER_AGENT")),
	}
	// Wire the optional bearer token: env first, then the keyring for this profile.
	store := d.store()
	opts = append(opts, api.WithToken(func(_ context.Context) (string, error) {
		if t := os.Getenv("TORRE_TOKEN"); t != "" {
			return t, nil
		}
		t, gerr := store.Get(profileName)
		if gerr != nil {
			if errors.Is(gerr, auth.ErrNotFound) {
				return "", nil // no token — public path
			}
			return "", gerr
		}
		return t, nil
	}))

	c := d.newClient(searchBase, apiBase, opts...)
	c.ShowToken = d.gf.showToken
	c.Verbose = d.gf.verbose
	c.VerboseOut = os.Stderr
	return c, cfg, nil
}

func (d *deps) stdout() io.Writer {
	if d.out != nil {
		return d.out
	}
	return os.Stdout
}

// render is the single output path for every command.
func (d *deps) render(cmd *cobra.Command, v any, defaultColumns []string) error {
	raw, ok := v.(json.RawMessage)
	if !ok {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		raw = b
	}
	format := output.Format(config.FirstNonEmpty(d.gf.outputFormat, string(output.FormatTable)))
	cols := normalizeColumns(d.gf.columns)
	if len(cols) == 0 && format != output.FormatID {
		cols = defaultColumns
	}
	return output.Render(raw, output.Options{
		Format:  format,
		Columns: cols,
		NoColor: d.gf.noColor,
		Quiet:   d.gf.quiet,
		JQ:      d.gf.jq,
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
	})
}

func normalizeColumns(cols []string) []string {
	var out []string
	for _, c := range cols {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return out
}
