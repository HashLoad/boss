// Package installer provides local dependency installation.
package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/dcp"
)

// LocalInstall installs dependencies locally.
func LocalInstall(config env.ConfigProvider, options InstallOptions, pkg *domain.Package) {
	// TODO noSave
	EnsureDependency(pkg, options.Args)
	if err := DoInstall(config, options, pkg); err != nil {
		msg.Die("❌ %s", err)
	}
	dcp.InjectDpcs(pkg, pkg.Lock)
}
