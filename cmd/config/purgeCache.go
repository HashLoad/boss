package config

import (
	"fmt"
	"github.com/hashload/boss/core"
	"github.com/hashload/boss/utils"
	"github.com/spf13/cobra"
	"strings"
)

var purgeCacheCmd = &cobra.Command{
	Use:           "cache",
	Short:         "Configure cache",
	ValidArgs:     []string{"rm"},
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.OnlyValidArgs(cmd, args)
		if err == nil {
			err = cobra.ExactArgs(1)(cmd, args)
		}
		if err != nil {
			println(err.Error())
			println()
			fmt.Printf("Valid args:\n\t%s\n", strings.Join(cmd.ValidArgs, "\n\t"))
			println()
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if utils.Contains(args, "rm") {
			core.RunGC(true)
		}
	},
}

func init() {
	CmdConfig.AddCommand(purgeCacheCmd)
}
