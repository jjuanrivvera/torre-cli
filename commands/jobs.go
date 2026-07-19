package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
)

// jobsColumns are the default table columns for a search result list.
var jobsColumns = []string{"id", "objective", "opportunity", "remote", "status", "locations"}

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
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search job opportunities",
		Long: `Search Torre.ai opportunities with skill/role, remote, location, organization, and
compensation filters. Results paginate with --size/--limit/--all. Machine output
(-o json/-o id/--jq) is the primary interface for an assistant; -o table is the human view.

A skill search needs an experience level (Torre rejects a bare skill); --experience defaults
to "potential-to-develop" and accepts Torre's levels such as "1-plus-years",
"2-plus-years", "3-plus-years", "5-plus-years".`,
		Example: `  torre jobs search --skill golang --remote
  torre jobs search --skill "product design" --location Colombia --limit 50 -o json
  torre jobs search --skill go --compensation 3000 --currency 'USD$' --periodicity monthly
  torre jobs search --skill go --remote -o id | head`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			results, err := c.SearchOpportunitiesAll(cmd.Context(), f, d.gf.size, d.gf.limit, d.gf.all)
			if err != nil {
				return err
			}
			if results == nil { // dry-run
				return nil
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
	fl.StringVar(&f.Location, "location", "", "location/country to match (e.g. Colombia)")
	fl.StringVar(&f.Organization, "organization", "", "organization name to match")
	fl.Float64Var(&f.Compensation, "compensation", 0, "minimum compensation amount")
	fl.StringVar(&f.Currency, "currency", "", `compensation currency (default "USD$")`)
	fl.StringVar(&f.Periodicity, "periodicity", "", "compensation periodicity: hourly|monthly|yearly (default monthly)")
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
