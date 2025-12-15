//nolint:testpackage // Testing internal implementation details
package tracker

import (
	"testing"
)

func TestNullTracker_Start_ReturnsNil(t *testing.T) {
	tracker := NewNull[TestStatus]()

	err := tracker.Start()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNullTracker_Stop_DoesNotPanic(_ *testing.T) {
	tracker := NewNull[TestStatus]()
	// Should not panic
	tracker.Stop()
}

func TestNullTracker_UpdateStatus_DoesNotPanic(_ *testing.T) {
	tracker := NewNull[TestStatus]()
	// Should not panic
	tracker.UpdateStatus("item", StatusDone, "message")
}

func TestNullTracker_AddItem_DoesNotPanic(_ *testing.T) {
	tracker := NewNull[TestStatus]()
	// Should not panic
	tracker.AddItem("newitem")
}

func TestNullTracker_IsEnabled_ReturnsFalse(t *testing.T) {
	tracker := NewNull[TestStatus]()

	if tracker.IsEnabled() {
		t.Error("expected NullTracker.IsEnabled() to return false")
	}
}

func TestNullTracker_IsStopped_ReturnsTrue(t *testing.T) {
	tracker := NewNull[TestStatus]()

	if !tracker.IsStopped() {
		t.Error("expected NullTracker.IsStopped() to return true")
	}
}

func TestNullTracker_GetStatus_ReturnsFalse(t *testing.T) {
	tracker := NewNull[TestStatus]()

	status, ok := tracker.GetStatus("anyitem")

	if ok {
		t.Error("expected NullTracker.GetStatus() to return false")
	}

	var zero TestStatus
	if status != zero {
		t.Errorf("expected zero value status, got %v", status)
	}
}

func TestNullTracker_Count_ReturnsZero(t *testing.T) {
	tracker := NewNull[TestStatus]()

	if tracker.Count() != 0 {
		t.Errorf("expected NullTracker.Count() to return 0, got %d", tracker.Count())
	}
}

func TestNullTracker_ImplementsTrackerInterface(_ *testing.T) {
	var _ Tracker[TestStatus] = NewNull[TestStatus]()
	// If this compiles, the interface is implemented
}

func TestBaseTracker_ImplementsTrackerInterface(_ *testing.T) {
	var _ Tracker[TestStatus] = New([]string{"a"}, Config[TestStatus]{
		DefaultStatus: StatusPending,
		StatusConfig:  testStatusConfig,
	})
	// If this compiles, the interface is implemented
}
