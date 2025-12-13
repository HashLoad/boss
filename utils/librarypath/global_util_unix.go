//go:build !windows
// +build !windows

package librarypath

import (
	"github.com/hashload/boss/pkg/msg"
)

// updateGlobalLibraryPath updates the global library path
func updateGlobalLibraryPath() {
	msg.Warn("updateGlobalLibraryPath not implemented on this platform")
}

// updateGlobalBrowsingByProject updates the global browsing path by project
func updateGlobalBrowsingByProject(_ string, _ bool) {
	msg.Warn("updateGlobalBrowsingByProject not implemented on this platform")
}
