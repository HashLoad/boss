//go:build !windows
// +build !windows

package registry

func getDelphiVersionFromRegistry() map[string]string {
	return map[string]string{}
}
