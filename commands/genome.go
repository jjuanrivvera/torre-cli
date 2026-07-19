package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	registrars = append(registrars, func(d *deps) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "genome <username>",
			Short: "Fetch a person's public Torre genome/bio",
			Long: `Fetch a person's public genome (their Torre bio: profile, strengths, experiences,
education, interests, and more) by username — the handle in their profile URL
(torre.ai/<username>). The full object is large; use --jq or -o json to slice it, which is
ideal for an assistant computing a candidate/role match.`,
			Example: `  torre genome torrenegra
  torre genome torrenegra --jq '.person.name'
  torre genome torrenegra --jq '[.strengths[].name]' -o json
  torre genome torrenegra -o yaml`,
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				c, _, err := d.getAPIClient()
				if err != nil {
					return err
				}
				body, err := c.Genome(cmd.Context(), args[0])
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
	})
}
