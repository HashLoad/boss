// Package cmd provides the entry point for the CLI application.
// It delegates to the cli adapter for actual command handling.
package cmd

import "github.com/hashload/boss/internal/adapters/primary/cli"

// Execute runs the CLI application.
func Execute() error {
	return cli.Execute()
}
