// Package config provides Git configuration commands.
package config

import (
	"strings"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

// boolToMode converts boolean to mode string
func boolToMode(embedded bool) string {
	if embedded {
		return "embedded"
	}

	return "native"
}

// registryGitCmd registers the git command
func registryGitCmd(root *cobra.Command) {
	gitCmd := &cobra.Command{
		Use:     "git",
		Short:   "Configure Git",
		Example: "boss config git mode",
	}

	gitModeCmd := &cobra.Command{
		Use:       "mode [type]",
		Short:     "Configure Git mode",
		ValidArgs: []string{"native", "embedded", "default"},
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.OnlyValidArgs(cmd, args)
			if err == nil {
				err = cobra.ExactArgs(1)(cmd, args)
			}
			if err != nil {
				msg.Warn(err.Error())
				msg.Info("Current: %s\n\nValid args:\n\t%s\n",
					boolToMode(env.GlobalConfiguration().GitEmbedded),
					strings.Join(cmd.ValidArgs, "\n\t"))
				return err
			}
			return nil
		},
		Run: func(_ *cobra.Command, args []string) {
			env.GlobalConfiguration().GitEmbedded = args[0] != "native"

			msg.Info("Using %s git", boolToMode(env.GlobalConfiguration().GitEmbedded))
			env.GlobalConfiguration().SaveConfiguration()
		},
	}

	root.AddCommand(gitCmd)
	gitCmd.AddCommand(gitModeCmd)
}
