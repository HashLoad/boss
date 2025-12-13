package tracker

import (
	"testing"
)

// TestStatus is a simple status type for testing.
type TestStatus int

const (
	StatusPending TestStatus = iota
	StatusRunning
	StatusDone
	StatusError
)

var testStatusConfig = StatusConfig[TestStatus]{
	StatusPending: {Icon: "⏳", StatusText: "Pending"},
	StatusRunning: {Icon: "🔄", StatusText: "Running"},
	StatusDone:    {Icon: "✓", StatusText: "Done"},
	StatusError:   {Icon: "✗", StatusText: "Error"},
}

func TestNew_WithEmptyItems_ReturnsDisabledTracker(t *testing.T) {
	tracker := New[TestStatus]([]string{}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	if tracker.IsEnabled() {
		t.Error("expected tracker to be disabled when created with empty items")
	}
}

func TestNew_WithItems_ReturnsEnabledTracker(t *testing.T) {
	tracker := New([]string{"item1", "item2"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	if !tracker.IsEnabled() {
		t.Error("expected tracker to be enabled when created with items")
	}
}

func TestNew_DuplicateItems_AreIgnored(t *testing.T) {
	tracker := New([]string{"item1", "item1", "item2"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	if tracker.Count() != 2 {
		t.Errorf("expected 2 items, got %d", tracker.Count())
	}
}

func TestBaseTracker_UpdateStatus(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	tracker.UpdateStatus("task1", StatusRunning, "processing")

	status, ok := tracker.GetStatus("task1")
	if !ok {
		t.Fatal("expected to find task1")
	}
	if status != StatusRunning {
		t.Errorf("expected status %v, got %v", StatusRunning, status)
	}
}

func TestBaseTracker_UpdateStatus_NonExistentItem(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	// Should not panic
	tracker.UpdateStatus("nonexistent", StatusRunning, "")

	_, ok := tracker.GetStatus("nonexistent")
	if ok {
		t.Error("expected nonexistent item to not be found")
	}
}

func TestBaseTracker_AddItem(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	if tracker.Count() != 1 {
		t.Fatalf("expected 1 item, got %d", tracker.Count())
	}

	tracker.AddItem("task2")

	if tracker.Count() != 2 {
		t.Errorf("expected 2 items after AddItem, got %d", tracker.Count())
	}

	status, ok := tracker.GetStatus("task2")
	if !ok {
		t.Fatal("expected to find task2")
	}
	if status != StatusPending {
		t.Errorf("expected default status %v, got %v", StatusPending, status)
	}
}

func TestBaseTracker_AddItem_Duplicate(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	tracker.AddItem("task1") // Duplicate

	if tracker.Count() != 1 {
		t.Errorf("expected 1 item (duplicate ignored), got %d", tracker.Count())
	}
}

func TestBaseTracker_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected bool
	}{
		{"empty items", []string{}, false},
		{"with items", []string{"a"}, true},
		{"multiple items", []string{"a", "b", "c"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := New(tt.items, Config[TestStatus]{
				DefaultStatus: StatusPending,
				StatusConfig:  testStatusConfig,
			})

			if tracker.IsEnabled() != tt.expected {
				t.Errorf("expected IsEnabled() = %v, got %v", tt.expected, tracker.IsEnabled())
			}
		})
	}
}

func TestBaseTracker_IsStopped(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	if tracker.IsStopped() {
		t.Error("expected tracker to not be stopped initially")
	}

	tracker.Stop()

	if !tracker.IsStopped() {
		t.Error("expected tracker to be stopped after Stop()")
	}
}

func TestBaseTracker_UpdateStatus_AfterStop(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	tracker.Stop()
	tracker.UpdateStatus("task1", StatusDone, "")

	// Status should remain as initial because tracker was stopped
	status, _ := tracker.GetStatus("task1")
	if status != StatusPending {
		t.Errorf("expected status to remain %v after stop, got %v", StatusPending, status)
	}
}

func TestBaseTracker_AddItem_AfterStop(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	tracker.Stop()
	tracker.AddItem("task2")

	if tracker.Count() != 1 {
		t.Errorf("expected count to remain 1 after adding item to stopped tracker, got %d", tracker.Count())
	}
}

func TestBaseTracker_GetStatus_NotFound(t *testing.T) {
	tracker := New([]string{"task1"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})

	_, ok := tracker.GetStatus("nonexistent")
	if ok {
		t.Error("expected ok to be false for nonexistent item")
	}
}

func TestBaseTracker_Count(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected int
	}{
		{"empty", []string{}, 0},
		{"one item", []string{"a"}, 1},
		{"three items", []string{"a", "b", "c"}, 3},
		{"duplicates removed", []string{"a", "a", "b"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := New(tt.items, Config[TestStatus]{
				DefaultStatus: StatusPending,
				StatusConfig:  testStatusConfig,
			})

			if tracker.Count() != tt.expected {
				t.Errorf("expected Count() = %d, got %d", tt.expected, tracker.Count())
			}
		})
	}
}
