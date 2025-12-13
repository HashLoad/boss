package installer

import (
	"fmt"
	"sync"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/pterm/pterm"
)

type DependencyStatus int

const (
	StatusWaiting DependencyStatus = iota
	StatusCloning
	StatusDownloading
	StatusChecking
	StatusInstalling
	StatusCompleted
	StatusSkipped
	StatusFailed
)

type ProgressTracker struct {
	dependencies map[string]*DependencyProgress
	area         *pterm.AreaPrinter
	mu           sync.Mutex
	enabled      bool
	stopped      bool
	order        []string
}

type DependencyProgress struct {
	Name    string
	Status  DependencyStatus
	Message string
}

func NewProgressTracker(deps []domain.Dependency) *ProgressTracker {
	if len(deps) == 0 {
		return &ProgressTracker{enabled: false}
	}

	pt := &ProgressTracker{
		dependencies: make(map[string]*DependencyProgress),
		order:        make([]string, 0, len(deps)),
		enabled:      true,
	}

	for _, dep := range deps {
		name := dep.Name()

		if _, exists := pt.dependencies[name]; exists {
			continue
		}

		pt.dependencies[name] = &DependencyProgress{
			Name:    name,
			Status:  StatusWaiting,
			Message: "",
		}
		pt.order = append(pt.order, name)
	}

	return pt
}

func (pt *ProgressTracker) Start() error {
	if !pt.enabled {
		return nil
	}

	area, _ := pterm.DefaultArea.Start()
	pt.area = area

	pt.render()

	return nil
}

func (pt *ProgressTracker) Stop() {
	if !pt.enabled {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.stopped = true
	if pt.area != nil {
		_ = pt.area.Stop()
	}
}

func (pt *ProgressTracker) UpdateStatus(depName string, status DependencyStatus, message string) {
	if !pt.enabled || pt.stopped {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	progress, exists := pt.dependencies[depName]
	if !exists {
		return
	}

	progress.Status = status
	progress.Message = message

	pt.render()
}

func (pt *ProgressTracker) render() {
	if pt.area == nil || pt.stopped {
		return
	}

	var lines []string
	seen := make(map[string]bool)

	for _, name := range pt.order {

		if seen[name] {
			continue
		}
		seen[name] = true

		progress := pt.dependencies[name]
		if progress != nil {
			lines = append(lines, pt.formatStatus(progress))
		}
	}

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	pt.area.Clear()
	pt.area.Update(content)
}

func (pt *ProgressTracker) formatStatus(progress *DependencyProgress) string {
	var icon string
	var statusText string

	switch progress.Status {
	case StatusWaiting:
		icon = pterm.LightYellow("⏳")
		statusText = pterm.Gray("Waiting...")
	case StatusCloning:
		icon = pterm.LightCyan("📥")
		statusText = pterm.LightCyan("Cloning...")
	case StatusDownloading:
		icon = pterm.LightCyan("⬇️")
		statusText = pterm.LightCyan("Downloading...")
	case StatusChecking:
		icon = pterm.LightBlue("🔍")
		statusText = pterm.LightBlue("Checking...")
	case StatusInstalling:
		icon = pterm.LightMagenta("⚙️")
		statusText = pterm.LightMagenta("Installing...")
	case StatusCompleted:
		icon = pterm.LightGreen("✓")
		statusText = pterm.LightGreen("Installed")
	case StatusSkipped:
		icon = pterm.Gray("→")
		statusText = pterm.Gray("Skipped")
	case StatusFailed:
		icon = pterm.LightRed("✗")
		statusText = pterm.LightRed("Failed")
	}

	name := pterm.Bold.Sprint(progress.Name)
	padding := 30 - len(progress.Name)
	if padding < 1 {
		padding = 1
	}

	spaces := ""
	for i := 0; i < padding; i++ {
		spaces += " "
	}

	if progress.Message != "" {
		return fmt.Sprintf("%s %s%s%s %s", icon, name, spaces, statusText, pterm.Gray(progress.Message))
	}
	return fmt.Sprintf("%s %s%s%s", icon, name, spaces, statusText)
}

func (pt *ProgressTracker) IsEnabled() bool {
	return pt.enabled
}

// AddDependency adds a transitive dependency to the tracking list if it doesn't exist.
func (pt *ProgressTracker) AddDependency(depName string) {
	if !pt.enabled || pt.stopped {
		return
	}

	pt.mu.Lock()
	defer pt.mu.Unlock()

	if _, exists := pt.dependencies[depName]; exists {
		return
	}

	pt.dependencies[depName] = &DependencyProgress{
		Name:    depName,
		Status:  StatusWaiting,
		Message: "",
	}

	for _, existing := range pt.order {
		if existing == depName {
			return
		}
	}
	pt.order = append(pt.order, depName)

	pt.render()
}

// Helper methods for common status updates.
func (pt *ProgressTracker) SetWaiting(depName string) {
	pt.UpdateStatus(depName, StatusWaiting, "")
}

func (pt *ProgressTracker) SetCloning(depName string) {
	pt.UpdateStatus(depName, StatusCloning, "")
}

func (pt *ProgressTracker) SetDownloading(depName string, message string) {
	pt.UpdateStatus(depName, StatusDownloading, message)
}

func (pt *ProgressTracker) SetChecking(depName string, message string) {
	pt.UpdateStatus(depName, StatusChecking, message)
}

func (pt *ProgressTracker) SetInstalling(depName string) {
	pt.UpdateStatus(depName, StatusInstalling, "")
}

func (pt *ProgressTracker) SetCompleted(depName string) {
	pt.UpdateStatus(depName, StatusCompleted, "")
}

func (pt *ProgressTracker) SetSkipped(depName string, reason string) {
	pt.UpdateStatus(depName, StatusSkipped, reason)
}

func (pt *ProgressTracker) SetFailed(depName string, err error) {
	pt.UpdateStatus(depName, StatusFailed, err.Error())
}
