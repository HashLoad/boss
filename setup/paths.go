package setup

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
)

// BuildMessage creates a message with instructions to add paths to the shell.
func BuildMessage(path []string) string {
	if runtime.GOOS == "windows" {
		advice := "\nTo add the path permanently, run the following command in the terminal:\n\n" +
			"Press Win + R, type 'sysdm.cpl' and press Enter\n" +
			"Click on the 'Advanced' tab and then on 'Environment Variables'\n" +
			"In the 'System Variables' section, click on 'Path' and then on 'Edit'\n" +
			"Click on 'New' and add the following path:\n" +
			"\n"

		for _, p := range path {
			advice += p + "\n"
		}
	}

	shellFile := ".bashrc"

	shell := os.Getenv("SHELL")

	if strings.HasSuffix(shell, "fish") {
		shellFile = ".config/fish/config.fish"
	}

	if strings.HasSuffix(shell, "zsh") {
		shellFile = ".zshrc"
	}

	pathStr := strings.Join(path, ":")

	return "\nTo add the path permanently, run the following command in the terminal:\n\n" +
		"echo 'export PATH=$PATH:" + pathStr + "' >> ~/" + shellFile + "\n" +
		"source ~/" + shellFile + "\n"
}

func InitializePath() {
	if env.GlobalConfiguration().Advices.SetupPath {
		return
	}

	paths := []string{
		consts.EnvBossBin,
		env.GetGlobalBinPath(),
		env.GetGlobalEnvBpl(),
		env.GetGlobalEnvDcu(),
		env.GetGlobalEnvDcp(),
	}

	var needAdd = false
	currentPath, err := os.Getwd()
	if err != nil {
		msg.Die("Failed to load current working directory \n %s", err.Error())
		return
	}

	splitPath := strings.Split(currentPath, ";")

	for _, path := range paths {
		if !utils.Contains(splitPath, path) {
			splitPath = append(splitPath, path)
			needAdd = true
			msg.Info("Adding path %s", path)
		}
	}

	if needAdd {
		newPath := strings.Join(splitPath, ";")
		currentPathEnv := os.Getenv(PATH)
		err := os.Setenv(PATH, currentPathEnv+";"+newPath)
		if err != nil {
			msg.Die("Failed to update PATH \n %s", err.Error())
			return
		}

		msg.Warn("Please restart your console after complete.")

		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			msg.Info(BuildMessage(paths))

			spinner, _ := pterm.DefaultSpinner.Start("Sleeping for 5 seconds")
			if spinner != nil {
				time.Sleep(5 * time.Second)
				_ = spinner.Stop()
			}

			env.GlobalConfiguration().Advices.SetupPath = true
			env.GlobalConfiguration().SaveConfiguration()
		}
	}
}
