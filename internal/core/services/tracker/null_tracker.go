package tracker

// NullTracker implements a no-op tracker that satisfies the Tracker interface.
// This follows the Null Object Pattern to eliminate nil checks throughout the codebase.
type NullTracker[S comparable] struct{}

// NewNull creates a new NullTracker.
func NewNull[S comparable]() *NullTracker[S] {
	return &NullTracker[S]{}
}

// Start is a no-op.
func (n *NullTracker[S]) Start() error { return nil }

// Stop is a no-op.
func (n *NullTracker[S]) Stop() {}

// UpdateStatus is a no-op.
func (n *NullTracker[S]) UpdateStatus(string, S, string) {}

// AddItem is a no-op.
func (n *NullTracker[S]) AddItem(string) {}

// IsEnabled always returns false.
func (n *NullTracker[S]) IsEnabled() bool { return false }

// IsStopped always returns true.
func (n *NullTracker[S]) IsStopped() bool { return true }

// GetStatus always returns zero value and false.
func (n *NullTracker[S]) GetStatus(string) (S, bool) {
	var zero S
	return zero, false
}

// Count always returns 0.
func (n *NullTracker[S]) Count() int { return 0 }
