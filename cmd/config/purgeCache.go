package config

import (
	"github.com/hashload/boss/pkg/gc"
	"github.com/spf13/cobra"
)

func RegisterCmd(cmd *cobra.Command) {
	purgeCacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Configure cache",
	}

	cleanAll := false

	rmCacheCmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"rm"},
		Short:   "Clean cache based on TTL and usage",
		RunE: func(_ *cobra.Command, _ []string) error {
			return gc.CleanupCache(true, cleanAll)
		},
	}

	rmCacheCmd.Flags().BoolVarP(&cleanAll, "all", "a", false, "clean all cache")

	purgeCacheCmd.AddCommand(rmCacheCmd)

	cmd.AddCommand(purgeCacheCmd)
}
