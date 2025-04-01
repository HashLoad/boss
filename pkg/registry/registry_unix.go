//go:build !windows
// +build !windows

package registry

import "github.com/hashload/boss/pkg/msg"

func getDelphiVersionFromRegistry() map[string]string {
	msg.Warn("getDelphiVersionFromRegistry not implemented on this platform")

	return map[string]string{}
}
