package cmd

import (
	"github.com/hashload/boss/core/gc"
	"github.com/spf13/cobra"
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Garbage collector",
	Long:  `Garbage collector to remove old cached files`,
	Run: func(cmd *cobra.Command, args []string) {
		gc.RunGC()
	},
}

func init() {
	RootCmd.AddCommand(gcCmd)
}
