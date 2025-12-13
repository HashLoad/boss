package compiler

import (
	"github.com/hashload/boss/internal/core/services/tracker"
	"github.com/pterm/pterm"
)

// BuildStatus represents the build status of a package.
type BuildStatus int

const (
	BuildStatusWaiting BuildStatus = iota
	BuildStatusBuilding
	BuildStatusSuccess
	BuildStatusFailed
	BuildStatusSkipped
)

// buildStatusConfig defines how each build status should be displayed.
var buildStatusConfig = tracker.StatusConfig[BuildStatus]{
	BuildStatusWaiting: {
		Icon:       pterm.LightYellow("⏳"),
		StatusText: pterm.Gray("Waiting..."),
	},
	BuildStatusBuilding: {
		Icon:       pterm.LightCyan("🔨"),
		StatusText: pterm.LightCyan("Building..."),
	},
	BuildStatusSuccess: {
		Icon:       pterm.LightGreen("✓"),
		StatusText: pterm.LightGreen("Built"),
	},
	BuildStatusFailed: {
		Icon:       pterm.LightRed("✗"),
		StatusText: pterm.LightRed("Failed"),
	},
	BuildStatusSkipped: {
		Icon:       pterm.Gray("→"),
		StatusText: pterm.Gray("Skipped"),
	},
}

// BuildTracker wraps the generic BaseTracker for package compilation.
// It provides convenience methods with semantic names for build operations.
type BuildTracker struct {
	tracker.Tracker[BuildStatus]
}

// NewBuildTracker creates a new BuildTracker for the given package names.
func NewBuildTracker(packageNames []string) *BuildTracker {
	if len(packageNames) == 0 {
		return &BuildTracker{
			Tracker: tracker.NewNull[BuildStatus](),
		}
	}

	seen := make(map[string]bool)
	names := make([]string, 0, len(packageNames))
	for _, name := range packageNames {
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}

	return &BuildTracker{
		Tracker: tracker.New(names, tracker.Config[BuildStatus]{
			DefaultStatus: BuildStatusWaiting,
			StatusConfig:  buildStatusConfig,
		}),
	}
}

// SetBuilding sets the status to building with the current project name.
func (bt *BuildTracker) SetBuilding(name string, project string) {
	bt.UpdateStatus(name, BuildStatusBuilding, project)
}

// SetSuccess sets the status to success.
func (bt *BuildTracker) SetSuccess(name string) {
	bt.UpdateStatus(name, BuildStatusSuccess, "")
}

// SetFailed sets the status to failed with a message.
func (bt *BuildTracker) SetFailed(name string, message string) {
	bt.UpdateStatus(name, BuildStatusFailed, message)
}

// SetSkipped sets the status to skipped with a reason.
func (bt *BuildTracker) SetSkipped(name string, reason string) {
	bt.UpdateStatus(name, BuildStatusSkipped, reason)
}
