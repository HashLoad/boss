//go:build !windows
// +build !windows

package librarypath

import (
	"github.com/hashload/boss/msg"
)

func updateGlobalLibraryPath() {
	msg.Warn("updateGlobalLibraryPath not implemented on this platform")
}

func updateGlobalBrowsingByProject(dprojName string, setReadOnly bool) {
	msg.Warn("updateGlobalBrowsingByProject not implemented on this platform")
}
