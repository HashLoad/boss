package cmd

import (
	"github.com/hashload/boss/core/gc"
	"github.com/hashload/boss/msg"
	"github.com/spf13/cobra"
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Garbage collector",
	Long:  `Garbage collector to remove old cached files`,
	Run: func(cmd *cobra.Command, args []string) {
		msg.Info("Running GC...")
		gc.RunGC()
	},
}

func init() {
	RootCmd.AddCommand(gcCmd)
}
