package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/api"
	"github.com/jjuanrivvera/torre-cli/internal/config"
	"github.com/jjuanrivvera/torre-cli/internal/version"
)

// doctorCheck is one diagnostic result.
type doctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var jsonOut bool
		cmd := &cobra.Command{
			Use:   "doctor",
			Short: "Diagnose configuration, keyring, and Torre connectivity",
			Long: `Run local and remote health checks: config file, keyring backend, optional token,
and a live search against Torre. Exits non-zero when a check fails, so it is scriptable.`,
			Example: `  torre doctor
  torre doctor --json`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				checks := d.runDoctor(cmd)
				failed := false
				for _, c := range checks {
					if !c.OK {
						failed = true
					}
				}
				if jsonOut {
					b, err := json.MarshalIndent(checks, "", "  ")
					if err != nil {
						return err
					}
					fmt.Fprintln(cmd.OutOrStdout(), string(b))
				} else {
					for _, c := range checks {
						mark := "✓"
						if !c.OK {
							mark = "✗"
						}
						fmt.Fprintf(cmd.OutOrStdout(), "%s %-12s %s\n", mark, c.Name, c.Detail)
					}
				}
				if failed {
					return fmt.Errorf("doctor found problems")
				}
				return nil
			},
		}
		cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
		return cmd
	})
}

func (d *deps) runDoctor(cmd *cobra.Command) []doctorCheck {
	var checks []doctorCheck
	add := func(name string, ok bool, detail string) {
		checks = append(checks, doctorCheck{Name: name, OK: ok, Detail: detail})
	}

	add("version", true, version.String())

	cfgPath, err := config.Path()
	if err != nil {
		add("config", false, err.Error())
		return checks
	}
	cfg, err := d.loadConfig()
	if err != nil {
		add("config", false, err.Error())
		return checks
	}
	add("config", true, cfgPath)

	profileName := cfg.ResolveProfileName(d.gf.profile)
	add("profile", true, profileName)

	store := d.store()
	if _, gerr := store.Get(profileName); gerr == nil {
		add("token", true, "stored (backend: "+store.Backend()+")")
	} else {
		add("token", true, "none (public access)")
	}

	c, _, err := d.getAPIClient()
	if err != nil {
		add("connectivity", false, err.Error())
		return checks
	}
	resp, _, err := c.SearchOpportunities(cmd.Context(), api.SearchFilters{Skill: "engineer"}, 1, 0)
	switch {
	case err != nil:
		add("connectivity", false, err.Error())
	case resp == nil:
		add("connectivity", true, "dry-run")
	default:
		add("connectivity", true, fmt.Sprintf("search OK (%s)", c.SearchBaseURL()))
	}
	return checks
}
