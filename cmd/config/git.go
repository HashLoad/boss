package config

import (
	"strings"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:     "git",
	Short:   "Configure Git",
	Example: "boss config git mode",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()

	},
}

func boolToMode(embedded bool) string {
	if embedded {
		return "embedded"
	}

	return "native"
}

var gitModeCmd = &cobra.Command{
	Use:           "mode [type]",
	Short:         "Configure Git mode",
	ValidArgs:     []string{"native", "embedded", "default"},
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.OnlyValidArgs(cmd, args)
		if err == nil {
			err = cobra.ExactArgs(1)(cmd, args)
		}
		if err != nil {
			msg.Warn(err.Error())
			msg.Info("Current: %s\n\nValid args:\n\t%s\n",
				boolToMode(env.GlobalConfiguration.GitEmbedded),
				strings.Join(cmd.ValidArgs, "\n\t"))
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "native":
			{
				env.GlobalConfiguration.GitEmbedded = false

			}
		case "embedded", "default":
			{
				env.GlobalConfiguration.GitEmbedded = true
			}

		}
		msg.Info("Using %s git", boolToMode(env.GlobalConfiguration.GitEmbedded))
		env.GlobalConfiguration.SaveConfiguration()

	},
}

func init() {
	CmdConfig.AddCommand(gitCmd)
	gitCmd.AddCommand(gitModeCmd)
}
