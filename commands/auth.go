package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/torre-cli/internal/auth"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		authCmd := &cobra.Command{
			Use:   "auth",
			Short: "Manage an optional Torre bearer token",
			Long: `Torre's public endpoints (job search, opportunity detail, genome, people search)
need no credentials, so torre works with no auth at all. A bearer token is only useful for
endpoints that require one; store it here and it lands in your OS keyring, scoped to the
active profile. You can also pass a token per-invocation via the TORRE_TOKEN env var.`,
		}
		authCmd.AddCommand(newAuthLoginCmd(d), newAuthLogoutCmd(d), newAuthStatusCmd(d))
		return authCmd
	})
}

func newAuthLoginCmd(d *deps) *cobra.Command {
	var tokenFlag string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a Torre bearer token in the OS keyring",
		Long: `Capture a bearer token for the active profile and store it in the OS keyring
(encrypted-file fallback on headless hosts, keyed by $TORRE_KEYRING_PASSWORD). The token is
read from a hidden prompt so it never echoes to the terminal; use --token only in trusted,
non-interactive contexts (it can leak into shell history).`,
		Example: `  torre auth login
  torre auth login --profile work`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, cfg, err := d.resolveProfile()
			if err != nil {
				return err
			}
			token := tokenFlag
			if token == "" {
				token, err = promptSecret(cmd, "Torre bearer token: ")
				if err != nil {
					return err
				}
			}
			if token == "" {
				return fmt.Errorf("no token provided")
			}
			if err := d.store().Set(profileName, token); err != nil {
				return err
			}
			prof, _ := cfg.Profile(profileName)
			prof.HasToken = true
			if err := cfg.SetProfile(profileName, prof); err != nil {
				return err
			}
			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = profileName
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Stored a token for profile %q.\n", profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&tokenFlag, "token", "", "token value (prefer the hidden prompt; --token can leak into history)")
	return cmd
}

func newAuthLogoutCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored token for the active profile",
		Example: `  torre auth logout
  torre auth logout --profile work`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, cfg, err := d.resolveProfile()
			if err != nil {
				return err
			}
			if err := d.store().Delete(profileName); err != nil && !errors.Is(err, auth.ErrNotFound) {
				return err
			}
			if prof, ok := cfg.Profile(profileName); ok && prof.HasToken {
				prof.HasToken = false
				_ = cfg.SetProfile(profileName, prof)
				_ = cfg.Save()
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed any stored token for profile %q.\n", profileName)
			return nil
		},
	}
}

func newAuthStatusCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the active profile and whether a token is stored",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, _, err := d.resolveProfile()
			if err != nil {
				return err
			}
			store := d.store()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "profile: %s\n", profileName)
			if _, gerr := store.Get(profileName); gerr == nil {
				fmt.Fprintf(out, "token:   stored (backend: %s)\n", store.Backend())
			} else {
				fmt.Fprintln(out, "token:   none (public access — no token needed)")
			}
			return nil
		},
	}
}
