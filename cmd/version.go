package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/hashload/boss/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show cli version",
	Long:    `This command show the client version`,
	Aliases: []string{"v"},
	Example: `  Print version:
  boss version`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion(true)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func printVersion(withDetails bool) {
	v := version.Get()
	if withDetails {
		jsonVersion, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(jsonVersion))
		}
	} else {
		fmt.Println(v.Version)
	}
}
