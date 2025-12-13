package installer

import (
	"os"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/dcp"
)

// LocalInstall installs dependencies locally.
func LocalInstall(options InstallOptions, pkg *domain.Package) {
	// TODO noSave
	EnsureDependency(pkg, options.Args)
	if err := DoInstall(options, pkg); err != nil {
		msg.Err("%s", err)
		os.Exit(1)
	}
	dcp.InjectDpcs(pkg, pkg.Lock)
}
