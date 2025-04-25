package cmd

import (
	"strings"

	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/scripts"
	"github.com/spf13/cobra"
)

func runCmdRegister(root *cobra.Command) {
	var runScript = &cobra.Command{
		Use:               "run",
		Short:             "Run cmd script",
		Long:              `Run cmd script`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: validScripts,
		Run: func(_ *cobra.Command, args []string) {
			scripts.Run(args)
		},
	}

	root.AddCommand(runScript)
}

func validScripts(
	_ *cobra.Command,
	args []string,
	_ string,
) ([]string, cobra.ShellCompDirective) {
	packageData, err := models.LoadPackage(false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if packageData.Scripts == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var scripts []string
	for script := range packageData.Scripts {
		scripts = append(scripts, script)
	}

	if len(args) == 0 {
		return scripts, cobra.ShellCompDirectiveDefault
	}

	if len(args) == 1 {
		var completions []string
		for _, script := range scripts {
			if script != args[0] && strings.HasPrefix(script, args[0]) {
				completions = append(completions, script)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	return scripts, cobra.ShellCompDirectiveNoFileComp
}
