package librarypath

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
)

func UpdateLibraryPath() {
	if env.Global {
		msg.Warn("::TODO::")
	} else {
		updateDprojLibraryPath()
	}

}
