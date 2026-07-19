package commands

import (
	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
)

// peopleColumns are the default table columns for a people search result list. Torre's
// people results wrap each person under nested keys, so the useful identity fields flatten
// to dotted paths — -o json is the primary interface here.
var peopleColumns = []string{"ggId", "name", "username", "professionalHeadline", "verified"}

func init() {
	registrars = append(registrars, func(d *deps) *cobra.Command {
		peopleCmd := &cobra.Command{
			Use:     "people",
			Aliases: []string{"person"},
			Short:   "Search Torre people",
			Long:    "Search the Torre.ai people index by skill/role, remote availability, and location.",
		}
		peopleCmd.AddCommand(newPeopleSearchCmd(d))
		return peopleCmd
	})
}

func newPeopleSearchCmd(d *deps) *cobra.Command {
	var f api.PeopleFilters
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search people by skill/role",
		Long: `Search Torre.ai people. A skill search needs an experience level (Torre rejects a
bare skill); --experience defaults to "potential-to-develop". Pair a match with
` + "`torre genome <username>`" + ` to pull a candidate's full profile.`,
		Example: `  torre people search --skill "data science" --remote
  torre people search --skill golang --location Colombia --limit 25 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			results, err := c.SearchPeopleAll(cmd.Context(), f, d.gf.size, d.gf.limit, d.gf.all)
			if err != nil {
				return err
			}
			if results == nil { // dry-run
				return nil
			}
			return d.render(cmd, marshalList(results), peopleColumns)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&f.Skill, "skill", "", "skill or role text to match")
	fl.StringVar(&f.Skill, "query", "", "alias for --skill")
	_ = fl.MarkHidden("query")
	fl.StringVar(&f.Experience, "experience", "", "required experience level (default potential-to-develop)")
	fl.BoolVar(&f.Remote, "remote", false, "only people open to remote work")
	fl.StringVar(&f.Location, "location", "", "location/country to match")
	return annotate(cmd, kindRead)
}
