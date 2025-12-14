// Package tracker provides progress tracking functionality for long-running operations.
// It displays real-time status updates for dependency installations and builds.
package tracker

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

// NamePadding is the standard padding for item names in the tracker display.
const NamePadding = 30

// StatusFormatter defines how a status should be displayed.
type StatusFormatter struct {
	Icon       string
	StatusText string
}

// StatusConfig maps status values to their display format.
type StatusConfig[S comparable] map[S]StatusFormatter

// ItemProgress represents the progress state of a single tracked item.
type ItemProgress[S comparable] struct {
	Name    string
	Status  S
	Message string
}

// BaseTracker provides a generic, thread-safe progress tracking implementation.
// It uses generics to support different status types while sharing common logic.
type BaseTracker[S comparable] struct {
	items         map[string]*ItemProgress[S]
	area          *pterm.AreaPrinter
	mu            sync.Mutex
	enabled       bool
	stopped       bool
	order         []string
	defaultStatus S
	statusConfig  StatusConfig[S]
}

// Config holds configuration for creating a new BaseTracker.
type Config[S comparable] struct {
	DefaultStatus S
	StatusConfig  StatusConfig[S]
}

// New creates a new BaseTracker with the given items and configuration.
func New[S comparable](itemNames []string, config Config[S]) *BaseTracker[S] {
	if len(itemNames) == 0 {
		return &BaseTracker[S]{enabled: false}
	}

	bt := &BaseTracker[S]{
		items:         make(map[string]*ItemProgress[S]),
		order:         make([]string, 0, len(itemNames)),
		enabled:       true,
		defaultStatus: config.DefaultStatus,
		statusConfig:  config.StatusConfig,
	}

	for _, name := range itemNames {
		if _, exists := bt.items[name]; exists {
			continue
		}

		bt.items[name] = &ItemProgress[S]{
			Name:    name,
			Status:  config.DefaultStatus,
			Message: "",
		}
		bt.order = append(bt.order, name)
	}

	return bt
}

// Start begins the progress tracking display.
func (bt *BaseTracker[S]) Start() error {
	if !bt.enabled {
		return nil
	}

	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return fmt.Errorf("starting area printer: %w", err)
	}
	bt.area = area
	bt.render()

	return nil
}

// Stop ends the progress tracking display.
func (bt *BaseTracker[S]) Stop() {
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

// UpdateStatus updates the status of an item.
func (bt *BaseTracker[S]) UpdateStatus(name string, status S, message string) {
	if !bt.enabled || bt.stopped {
		return
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()

	progress, exists := bt.items[name]
	if !exists {
		return
	}

	progress.Status = status
	progress.Message = message

	bt.render()
}

// AddItem dynamically adds a new item to the tracker.
func (bt *BaseTracker[S]) AddItem(name string) {
	if !bt.enabled || bt.stopped {
		return
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()

	if _, exists := bt.items[name]; exists {
		return
	}

	bt.items[name] = &ItemProgress[S]{
		Name:    name,
		Status:  bt.defaultStatus,
		Message: "",
	}

	if slices.Contains(bt.order, name) {
		return
	}
	bt.order = append(bt.order, name)

	bt.render()
}

// IsEnabled returns whether the tracker is enabled.
func (bt *BaseTracker[S]) IsEnabled() bool {
	return bt.enabled
}

// IsStopped returns whether the tracker has been stopped.
func (bt *BaseTracker[S]) IsStopped() bool {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	return bt.stopped
}

// render updates the terminal display. Must be called with lock held.
func (bt *BaseTracker[S]) render() {
	if bt.area == nil || bt.stopped {
		return
	}

	lines := make([]string, 0, len(bt.order))
	seen := make(map[string]bool, len(bt.order))

	for _, name := range bt.order {
		if seen[name] {
			continue
		}
		seen[name] = true

		if progress := bt.items[name]; progress != nil {
			lines = append(lines, bt.formatStatus(progress))
		}
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	bt.area.Clear()
	bt.area.Update(content)
}

// formatStatus formats a single item's status for display.
func (bt *BaseTracker[S]) formatStatus(progress *ItemProgress[S]) string {
	formatter, ok := bt.statusConfig[progress.Status]
	if !ok {
		formatter = StatusFormatter{
			Icon:       pterm.Gray("?"),
			StatusText: pterm.Gray("Unknown"),
		}
	}

	name := pterm.Bold.Sprint(progress.Name)

	if progress.Message != "" {
		return fmt.Sprintf("%s %-*s%s %s",
			formatter.Icon,
			NamePadding, name,
			formatter.StatusText,
			pterm.Gray(progress.Message))
	}
	return fmt.Sprintf("%s %-*s%s", formatter.Icon, NamePadding, name, formatter.StatusText)
}

// GetStatus returns the current status of an item.
func (bt *BaseTracker[S]) GetStatus(name string) (S, bool) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if progress, exists := bt.items[name]; exists {
		return progress.Status, true
	}

	var zero S
	return zero, false
}

// Count returns the number of tracked items.
func (bt *BaseTracker[S]) Count() int {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	return len(bt.items)
}
