// Package installer provides progress tracking for installations.
package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/tracker"
	"github.com/pterm/pterm"
)

// DependencyStatus represents the installation status of a dependency.
type DependencyStatus int

const (
	// StatusWaiting indicates the dependency is waiting to be processed.
	StatusWaiting DependencyStatus = iota
	// StatusCloning indicates the dependency is being cloned.
	StatusCloning
	// StatusDownloading indicates the dependency is being downloaded.
	StatusDownloading
	// StatusUpdating indicates the dependency is being updated.
	StatusUpdating
	// StatusChecking indicates the dependency is being checked.
	StatusChecking
	// StatusInstalling indicates the dependency is being installed.
	StatusInstalling
	// StatusCompleted indicates the dependency installation is completed.
	StatusCompleted
	// StatusSkipped indicates the dependency was skipped.
	StatusSkipped
	// StatusFailed indicates the dependency installation failed.
	StatusFailed
	// StatusWarning indicates a warning occurred during installation.
	StatusWarning
)

// dependencyStatusConfig defines how each status should be displayed.
//
//nolint:gochecknoglobals // Dependency status configuration
var dependencyStatusConfig = tracker.StatusConfig[DependencyStatus]{
	StatusWaiting: {
		Icon:       pterm.LightYellow("⏳"),
		StatusText: pterm.Gray("Waiting..."),
	},
	StatusCloning: {
		Icon:       pterm.LightCyan("🧬"),
		StatusText: pterm.LightCyan("Cloning..."),
	},
	StatusDownloading: {
		Icon:       pterm.LightCyan("📥"),
		StatusText: pterm.LightCyan("Downloading..."),
	},
	StatusUpdating: {
		Icon:       pterm.LightCyan("🔁"),
		StatusText: pterm.LightCyan("Updating..."),
	},
	StatusChecking: {
		Icon:       pterm.LightBlue("🔎"),
		StatusText: pterm.LightBlue("Checking..."),
	},
	StatusInstalling: {
		Icon:       pterm.LightMagenta("🔥"),
		StatusText: pterm.LightMagenta("Installing..."),
	},
	StatusCompleted: {
		Icon:       pterm.LightGreen("📦"),
		StatusText: pterm.LightGreen("Installed"),
	},
	StatusSkipped: {
		Icon:       pterm.Gray("⏩"),
		StatusText: pterm.Gray("Skipped"),
	},
	StatusFailed: {
		Icon:       pterm.LightRed("⛓️‍💥"),
		StatusText: pterm.LightRed("Failed"),
	},
	StatusWarning: {
		Icon:       pterm.LightYellow("⚠️"),
		StatusText: pterm.LightYellow("Warning"),
	},
}

// ProgressTracker wraps the generic BaseTracker for dependency installation.
// It provides convenience methods with semantic names for installation operations.
type ProgressTracker struct {
	tracker.Tracker[DependencyStatus]
}

// NewProgressTracker creates a new ProgressTracker for the given dependencies.
func NewProgressTracker(deps []domain.Dependency) *ProgressTracker {
	names := make([]string, 0, len(deps))
	seen := make(map[string]bool)

	for _, dep := range deps {
		name := dep.Name()
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}

	if len(names) == 0 {
		return &ProgressTracker{
			Tracker: tracker.NewNull[DependencyStatus](),
		}
	}

	return &ProgressTracker{
		Tracker: tracker.New(names, tracker.Config[DependencyStatus]{
			DefaultStatus: StatusWaiting,
			StatusConfig:  dependencyStatusConfig,
		}),
	}
}

// AddDependency adds a transitive dependency to the tracking list.
func (pt *ProgressTracker) AddDependency(depName string) {
	pt.AddItem(depName)
}

// SetWaiting sets the status to waiting.
func (pt *ProgressTracker) SetWaiting(depName string) {
	pt.UpdateStatus(depName, StatusWaiting, "")
}

// SetCloning sets the status to cloning.
func (pt *ProgressTracker) SetCloning(depName string) {
	pt.UpdateStatus(depName, StatusCloning, "")
}

// SetDownloading sets the status to downloading with a message.
func (pt *ProgressTracker) SetDownloading(depName string, message string) {
	pt.UpdateStatus(depName, StatusDownloading, message)
}

// SetUpdating sets the status to updating with a message.
func (pt *ProgressTracker) SetUpdating(depName string, message string) {
	pt.UpdateStatus(depName, StatusUpdating, message)
}

// SetChecking sets the status to checking with a message.
func (pt *ProgressTracker) SetChecking(depName string, message string) {
	pt.UpdateStatus(depName, StatusChecking, message)
}

// SetInstalling sets the status to installing.
func (pt *ProgressTracker) SetInstalling(depName string) {
	pt.UpdateStatus(depName, StatusInstalling, "")
}

// SetCompleted sets the status to completed.
func (pt *ProgressTracker) SetCompleted(depName string) {
	pt.UpdateStatus(depName, StatusCompleted, "")
}

// SetSkipped sets the status to skipped with a reason.
func (pt *ProgressTracker) SetSkipped(depName string, reason string) {
	pt.UpdateStatus(depName, StatusSkipped, reason)
}

// SetFailed sets the status to failed with an error.
func (pt *ProgressTracker) SetFailed(depName string, err error) {
	pt.UpdateStatus(depName, StatusFailed, err.Error())
}

// SetWarning sets the status to warning with a message.
func (pt *ProgressTracker) SetWarning(depName string, message string) {
	pt.UpdateStatus(depName, StatusWarning, message)
}
