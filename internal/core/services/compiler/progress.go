package compiler

import (
	"fmt"
	"sync"

	"github.com/pterm/pterm"
)

type BuildStatus int

const (
	BuildStatusWaiting BuildStatus = iota
	BuildStatusBuilding
	BuildStatusSuccess
	BuildStatusFailed
	BuildStatusSkipped
)

type BuildTracker struct {
	packages map[string]*BuildProgress
	area     *pterm.AreaPrinter
	mu       sync.Mutex
	enabled  bool
	stopped  bool
	order    []string
}

type BuildProgress struct {
	Name    string
	Status  BuildStatus
	Message string
}

func NewBuildTracker(packageNames []string) *BuildTracker {
	if len(packageNames) == 0 {
		return &BuildTracker{enabled: false}
	}

	bt := &BuildTracker{
		packages: make(map[string]*BuildProgress),
		order:    make([]string, 0, len(packageNames)),
		enabled:  true,
	}

	for _, name := range packageNames {
		if _, exists := bt.packages[name]; exists {
			continue
		}

		bt.packages[name] = &BuildProgress{
			Name:   name,
			Status: BuildStatusWaiting,
		}
		bt.order = append(bt.order, name)
	}

	return bt
}

func (bt *BuildTracker) Start() error {
	if !bt.enabled {
		return nil
	}

	area, _ := pterm.DefaultArea.Start()
	bt.area = area
	bt.render()

	return nil
}

func (bt *BuildTracker) Stop() {
	if !bt.enabled {
		return
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()

	bt.stopped = true
	if bt.area != nil {
		_ = bt.area.Stop()
	}
}

func (bt *BuildTracker) UpdateStatus(name string, status BuildStatus, message string) {
	if !bt.enabled || bt.stopped {
		return
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()

	progress, exists := bt.packages[name]
	if !exists {
		return
	}

	progress.Status = status
	progress.Message = message
	bt.render()
}

func (bt *BuildTracker) render() {
	if bt.area == nil || bt.stopped {
		return
	}

	var lines []string
	for _, name := range bt.order {
		progress := bt.packages[name]
		if progress != nil {
			lines = append(lines, bt.formatStatus(progress))
		}
	}

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	bt.area.Clear()
	bt.area.Update(content)
}

func (bt *BuildTracker) formatStatus(progress *BuildProgress) string {
	var icon string
	var statusText string

	switch progress.Status {
	case BuildStatusWaiting:
		icon = pterm.LightYellow("⏳")
		statusText = pterm.Gray("Waiting...")
	case BuildStatusBuilding:
		icon = pterm.LightCyan("🔨")
		statusText = pterm.LightCyan("Building...")
	case BuildStatusSuccess:
		icon = pterm.LightGreen("✓")
		statusText = pterm.LightGreen("Built")
	case BuildStatusFailed:
		icon = pterm.LightRed("✗")
		statusText = pterm.LightRed("Failed")
	case BuildStatusSkipped:
		icon = pterm.Gray("→")
		statusText = pterm.Gray("Skipped")
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

func (bt *BuildTracker) IsEnabled() bool {
	return bt.enabled
}

func (bt *BuildTracker) SetBuilding(name string, project string) {
	bt.UpdateStatus(name, BuildStatusBuilding, project)
}

func (bt *BuildTracker) SetSuccess(name string) {
	bt.UpdateStatus(name, BuildStatusSuccess, "")
}

func (bt *BuildTracker) SetFailed(name string, message string) {
	bt.UpdateStatus(name, BuildStatusFailed, message)
}

func (bt *BuildTracker) SetSkipped(name string, reason string) {
	bt.UpdateStatus(name, BuildStatusSkipped, reason)
}
