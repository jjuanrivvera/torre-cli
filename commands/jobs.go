package commands

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
)

// jobsColumns are the default table columns for a search result list.
var jobsColumns = []string{"id", "objective", "opportunity", "remote", "status", "locations"}

// sinceDefaultScan is how many results --since scans when a date is pinned but no explicit
// --limit is given and --all is off. Torre orders by relevance, not date, so recent items
// are sparse in a small page; a wider scan surfaces them without an unbounded --all
// (DECISIONS.md).
const sinceDefaultScan = 100

func init() {
	registrars = append(registrars, func(d *deps) *cobra.Command {
		jobsCmd := &cobra.Command{
			Use:     "jobs",
			Aliases: []string{"opportunities", "opps"},
			Short:   "Search and inspect Torre job opportunities",
			Long:    "Search Torre.ai job opportunities and fetch a single opportunity's full detail.",
		}
		jobsCmd.AddCommand(newJobsSearchCmd(d), newJobsGetCmd(d))
		return jobsCmd
	})
}

func newJobsSearchCmd(d *deps) *cobra.Command {
	var f api.SearchFilters
	var since string
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search job opportunities",
		Long: `Search Torre.ai opportunities with skill/role, remote, location, organization, and
compensation filters. Results paginate with --size/--limit/--all. Machine output
(-o json/-o id/--jq) is the primary interface for an assistant; -o table is the human view.

Not every flag narrows the result set. --skill (and --experience) narrows the search, and
--since (alias --posted-after) is a hard client-side date filter. But --location and
--compensation (with --currency/--periodicity) are RANKING HINTS Torre applies server-side:
they nudge relevance/ordering, they do NOT restrict results to that location or pay. A remote
role, for example, carries no location and is not dropped by --location.

A skill search needs an experience level (Torre rejects a bare skill); --experience defaults
to "potential-to-develop" and accepts Torre's levels such as "1-plus-years",
"2-plus-years", "3-plus-years", "5-plus-years".

Results come back ordered by RELEVANCE, not date, and span years. --since (alias
--posted-after) drops anything older than a threshold — an absolute YYYY-MM-DD or a relative
Nd/Nw (last N days/weeks). Because recent items are sparse in a small relevance-ordered page,
pair --since with --all or a larger --limit; when neither is set --since widens the scan.`,
		Example: `  torre jobs search --skill golang --remote
  torre jobs search --skill "product design" --location Colombia --limit 50 -o json
  torre jobs search --skill go --since 7d --remote -o json
  torre jobs search --skill go --posted-after 2026-07-12 --all -o id
  torre jobs search --skill go --compensation 3000 --currency 'USD$' --periodicity monthly`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			size, limit := d.gf.size, d.gf.limit
			var threshold time.Time
			if since != "" {
				threshold, err = api.ParseSince(since, time.Now())
				if err != nil {
					return err
				}
				// Relevance-ordered results bury recent items; widen the scan when a date
				// is pinned but the user gave no explicit --limit and didn't ask for --all.
				if !cmd.Flags().Changed("limit") && !d.gf.all {
					limit = sinceDefaultScan
				}
			}
			results, err := c.SearchOpportunitiesAll(cmd.Context(), f, size, limit, d.gf.all)
			if err != nil {
				return err
			}
			if results == nil { // dry-run
				return nil
			}
			if since != "" {
				results = api.FilterByCreated(results, threshold)
			}
			return d.render(cmd, marshalList(results), jobsColumns)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&f.Skill, "skill", "", "skill or role text to match")
	fl.StringVar(&f.Skill, "query", "", "alias for --skill")
	_ = fl.MarkHidden("query")
	fl.StringVar(&f.Experience, "experience", "", "required experience level (default potential-to-develop)")
	fl.BoolVar(&f.Remote, "remote", false, "only remote opportunities")
	fl.StringVar(&f.Location, "location", "", "location/country ranking hint applied server-side (nudges relevance/ordering; does NOT restrict results — unlike --since)")
	fl.StringVar(&f.Organization, "organization", "", "organization name to match")
	fl.Float64Var(&f.Compensation, "compensation", 0, "compensation ranking hint applied server-side (nudges relevance/ordering; does NOT restrict results)")
	fl.StringVar(&f.Currency, "currency", "", `currency for the --compensation ranking hint (default "USD$")`)
	fl.StringVar(&f.Periodicity, "periodicity", "", "periodicity for the --compensation ranking hint: hourly|monthly|yearly (default monthly)")
	fl.StringVar(&since, "since", "", "keep only opportunities created on/after this date: absolute YYYY-MM-DD or relative Nd/Nw (e.g. 7d, 2w)")
	fl.StringVar(&since, "posted-after", "", "alias for --since")
	_ = fl.MarkHidden("posted-after")
	return annotate(cmd, kindRead)
}

func newJobsGetCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Fetch one opportunity's full detail",
		Long:  "Fetch the complete detail for a single opportunity by its id (from `torre jobs search`).",
		Example: `  torre jobs get KWN4QjAd
  torre jobs get KWN4QjAd -o yaml
  torre jobs get KWN4QjAd --jq '.compensation'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			body, err := c.GetOpportunity(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if body == nil { // dry-run
				return nil
			}
			return d.render(cmd, body, nil)
		},
	}
	return annotate(cmd, kindRead)
}

// marshalList wraps a slice of raw results into a single JSON array for the renderer.
func marshalList(items []json.RawMessage) json.RawMessage {
	if items == nil {
		items = []json.RawMessage{}
	}
	b, _ := json.Marshal(items)
	return b
}
