package cmd

import (
	"errors"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/dcc32"
	"github.com/spf13/cobra"
	"os"
)

var cmdDelphi = &cobra.Command{
	Use:   "delphi",
	Short: "Configure Delphi version",
	Long:  `Configure Delphi version to compile modules`,
	Run: func(cmd *cobra.Command, args []string) {
		msg.Info("Running in path %s", env.GlobalConfiguration.DelphiPath)
		_ = cmd.Usage()
	},
}

var cmdDelphiList = &cobra.Command{
	Use:   "list",
	Short: "List Delphi versions",
	Long:  `List Delphi versions to compile modules`,
	Run: func(cmd *cobra.Command, args []string) {
		paths := dcc32.GetDcc32DirByCmd()
		if len(paths) == 0 {
			msg.Warn("Installations not found in $PATH")
			return
		} else {
			msg.Warn("Installations found:")
			for index, path := range paths {
				msg.Info("  [%d] %s", index, path)
			}
		}
	},
}

var cmdDelphiUse = &cobra.Command{
	Use:   "use [path]",
	Short: "Use Delphi version",
	Long:  `Use Delphi version to compile modules`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if _, err := os.Stat(args[0]); os.IsNotExist(err) {
			return errors.New("invalid path")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		config := env.GlobalConfiguration
		config.DelphiPath = args[0]
		config.SaveConfiguration()
		msg.Info("Successful!")
	},
}

func init() {
	RootCmd.AddCommand(cmdDelphi)
	cmdDelphi.AddCommand(cmdDelphiList)
	cmdDelphi.AddCommand(cmdDelphiUse)
}
