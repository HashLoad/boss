package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run cmd script",
	Long:  `Run cmd script`,
	/*Run: func(cmd *cobra.Command, args []string) {
		if pkgJson, e := models.LoadPackage("package.json", true); e != nil {
			e.Error()

		} else {

			pkgJson.Scripts.(map[string]interface{})
			print(v)
		}
	},*/
}

func init() {
	RootCmd.AddCommand(runCmd)
}
