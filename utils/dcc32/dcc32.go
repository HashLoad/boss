// Package dcc32 provides utilities for locating the Delphi command-line compiler (dcc32.exe).
// It searches the system PATH for installed Delphi compilers.
package dcc32

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GetDcc32DirByCmd returns the directory of the dcc32 executable found in the system path
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
	for value := range strings.SplitSeq(outputStr, "\n") {
		if len(strings.TrimSpace(value)) > 0 {
			installations = append(installations, filepath.Dir(value))
		}
	}

	return installations
}
