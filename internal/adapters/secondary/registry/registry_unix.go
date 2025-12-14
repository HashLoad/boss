//go:build !windows
// +build !windows

// Package registryadapter provides Unix/Linux stub implementations for registry operations.
package registryadapter

import "github.com/hashload/boss/pkg/msg"

// getDelphiVersionFromRegistry returns the delphi version from the registry
func getDelphiVersionFromRegistry() map[string]string {
	msg.Warn("⚠️ getDelphiVersionFromRegistry not implemented on this platform")

	return map[string]string{}
}

// getDetectedDelphisFromRegistry returns the detected delphi installations from the registry
func getDetectedDelphisFromRegistry() []DelphiInstallation {
	msg.Warn("⚠️ getDetectedDelphisFromRegistry not implemented on this platform")
	return []DelphiInstallation{}
}
