package setup

import (
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

func updateVersion(newVersion int64) {
	env.GlobalConfiguration().ConfigVersion = newVersion
	env.GlobalConfiguration().SaveConfiguration()
}

func needUpdate(toVersion int64) bool {
	return env.GlobalConfiguration().ConfigVersion < toVersion
}

func executeUpdate(version int64, update func()) {
	if needUpdate(version) {
		msg.Debug("\t\tRunning update to version %d", version)
		update()
		updateVersion(version)
	} else {
		msg.Debug("\t\tUpdate to version %d already performed", version)
	}
}

func migration() {
	executeUpdate(1, one)
	executeUpdate(2, two)
	executeUpdate(3, three)
	executeUpdate(4, cleanup)
	executeUpdate(5, cleanup)
	executeUpdate(6, six)
	executeUpdate(7, func() {
		seven()
		cleanup()
	})
}
