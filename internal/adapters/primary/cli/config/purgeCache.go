package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashload/boss/internal/core/services/gc"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// RegisterCmd registers the cache command
func RegisterCmd(cmd *cobra.Command) {
	purgeCacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Configure cache",
	}

	rmCacheCmd := &cobra.Command{
		Use:     "rm",
		Short:   "Remove cache",
		Aliases: []string{"purge", "clean"},
		Long:    "Remove all cached modules. This will free up disk space but modules will need to be re-downloaded.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return removeCacheWithConfirmation()
		},
	}

	purgeCacheCmd.AddCommand(rmCacheCmd)

	cmd.AddCommand(purgeCacheCmd)
}

// removeCacheWithConfirmation removes the cache with confirmation
func removeCacheWithConfirmation() error {
	modulesDir := env.GetModulesDir()

	var totalSize int64
	err := filepath.Walk(modulesDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		msg.Warn("Could not calculate cache size: %s", err)
	}

	sizeStr := formatBytes(totalSize)

	entries, _ := os.ReadDir(modulesDir)
	moduleCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			moduleCount++
		}
	}

	pterm.Warning.Printfln("This will remove ALL cached modules")
	pterm.Info.Printfln("  Modules: %d", moduleCount)
	pterm.Info.Printfln("  Size: %s", sizeStr)
	pterm.Info.Printfln("  Path: %s\n", modulesDir)

	result, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).
		WithDefaultText("Are you sure you want to continue?").
		Show()

	if !result {
		msg.Info("Cache purge cancelled")
		return nil
	}

	return gc.RunGC(true)
}

// formatBytes formats bytes to string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
