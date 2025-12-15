package env

// ConfigAccessor provides helper functions to access configuration
// with better testability. These functions wrap the global singleton
// but can be easily mocked or replaced in tests.
type ConfigAccessor struct {
	provider ConfigProvider
}

// NewConfigAccessor creates a new accessor with the given provider.
func NewConfigAccessor(provider ConfigProvider) *ConfigAccessor {
	return &ConfigAccessor{provider: provider}
}

// GetDelphiPath returns the configured Delphi path.
func (a *ConfigAccessor) GetDelphiPath() string {
	return a.provider.GetDelphiPath()
}

// GetGitEmbedded returns whether embedded git is enabled.
func (a *ConfigAccessor) GetGitEmbedded() bool {
	return a.provider.GetGitEmbedded()
}
