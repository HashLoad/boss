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

	if len(strings.ReplaceAll(outputStr, "\n", "")) == 0 {
		return []string{}
	}

	installations := []string{}
	for _, value := range strings.Split(outputStr, "\n") {
		if len(strings.TrimSpace(value)) > 0 {
			installations = append(installations, filepath.Dir(value))
		}
	}

	return installations
}
