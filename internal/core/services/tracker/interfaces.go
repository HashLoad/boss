package tracker

// Tracker defines the interface for progress tracking.
// Both BaseTracker and NullTracker implement this interface,
// allowing consumers to use either without nil checks.
type Tracker[S comparable] interface {
	// Start begins the progress tracking display.
	Start() error

	// Stop ends the progress tracking display.
	Stop()

	// UpdateStatus updates the status of an item.
	UpdateStatus(name string, status S, message string)

	// AddItem dynamically adds a new item to the tracker.
	AddItem(name string)

	// IsEnabled returns whether the tracker is enabled.
	IsEnabled() bool

	// IsStopped returns whether the tracker has been stopped.
	IsStopped() bool

	// GetStatus returns the current status of an item.
	GetStatus(name string) (S, bool)

	// Count returns the number of tracked items.
	Count() int
}

// Compile-time interface compliance checks.
var (
	_ Tracker[int] = (*BaseTracker[int])(nil)
	_ Tracker[int] = (*NullTracker[int])(nil)
)
