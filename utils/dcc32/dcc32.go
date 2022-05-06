package dcc32

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func GetDcc32DirByCmd() []string {
	command := exec.Command("where", "dcc32")
	output, err := command.Output()

	if err != nil {
		return []string{}
	}

	outputStr := strings.ReplaceAll(string(output), "\t", "")
	outputStr = strings.ReplaceAll(outputStr, "\r", "")
	outputStr = strings.ReplaceAll(outputStr, "\n", "")

	if len(outputStr) == 0 {
		return []string{}
	}

	installations := strings.Split(outputStr, "\n")
	for key, value := range installations {
		installations[key] = filepath.Dir(value)
	}

	return installations
}
