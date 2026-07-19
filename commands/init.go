package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		return &cobra.Command{
			Use:     "init",
			Aliases: []string{"setup"},
			Short:   "First-run setup wizard",
			Long: `Walk through torre setup. Because Torre's public API needs no credentials, setup is
mostly informational: it confirms connectivity and optionally stores a bearer token for
endpoints that require one.`,
			Example: `  torre init`,
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				out := cmd.OutOrStdout()
				profileName, cfg, err := d.resolveProfile()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "torre setup — profile %q\n", profileName)
				fmt.Fprintln(out, "Torre's public endpoints need no token; you can start immediately:")
				fmt.Fprintln(out, "  torre jobs search --skill golang --remote")

				// A closed/empty stdin (non-interactive, piped) means "no" — never fail setup on it.
				ans, _ := promptLine(cmd, "Store an optional bearer token now? [y/N]: ")
				if strings.EqualFold(strings.TrimSpace(ans), "y") {
					token, err := promptSecret(cmd, "Torre bearer token: ")
					if err != nil {
						return err
					}
					if token != "" {
						if err := d.store().Set(profileName, token); err != nil {
							return err
						}
						prof, _ := cfg.Profile(profileName)
						prof.HasToken = true
						_ = cfg.SetProfile(profileName, prof)
						if cfg.CurrentProfile == "" {
							cfg.CurrentProfile = profileName
						}
						if err := cfg.Save(); err != nil {
							return err
						}
						fmt.Fprintf(out, "Stored a token for profile %q.\n", profileName)
					}
				}

				// A connectivity smoke test, honoring --dry-run.
				c, _, err := d.getAPIClient()
				if err != nil {
					return err
				}
				resp, _, err := c.SearchOpportunities(cmd.Context(), api.SearchFilters{Skill: "engineer"}, 1, 0)
				switch {
				case err != nil:
					fmt.Fprintf(cmd.ErrOrStderr(), "connectivity check failed: %v\n", err)
				case resp != nil:
					fmt.Fprintln(out, "Connectivity OK. You're ready.")
				}
				return nil
			},
		}
	})
}
