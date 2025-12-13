package installer

import (
	"testing"
	"time"

	"github.com/hashload/boss/internal/core/domain"
)

func TestProgressTracker(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping interactive progress tracker test")
	}

	// Create fake dependencies
	deps := []domain.Dependency{
		{Repository: "github.com/hashload/horse"},
		{Repository: "github.com/hashload/dataset-serialize"},
		{Repository: "github.com/hashload/jhonson"},
		{Repository: "github.com/hashload/redis-client"},
		{Repository: "github.com/hashload/boss-core"},
	}

	tracker := NewProgressTracker(deps)

	if err := tracker.Start(); err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Simulate installation progress
	time.Sleep(500 * time.Millisecond)

	tracker.SetCloning("horse")
	time.Sleep(1 * time.Second)

	tracker.SetCloning("dataset-serialize")
	tracker.SetChecking("horse", "resolving version")
	time.Sleep(1 * time.Second)

	tracker.SetInstalling("horse")
	tracker.SetCloning("jhonson")
	tracker.SetChecking("dataset-serialize", "resolving version")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("horse")
	tracker.SetInstalling("dataset-serialize")
	tracker.SetCloning("redis-client")
	tracker.SetChecking("jhonson", "resolving version")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("dataset-serialize")
	tracker.SetInstalling("jhonson")
	tracker.SetCloning("boss-core")
	tracker.SetChecking("redis-client", "resolving version")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("jhonson")
	tracker.SetSkipped("redis-client", "already up to date")
	tracker.SetInstalling("boss-core")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("boss-core")
	time.Sleep(2 * time.Second)
}

func TestProgressTrackerWithDynamicDependencies(t *testing.T) {
	// Skip in CI/non-interactive environments
	if testing.Short() {
		t.Skip("Skipping interactive progress tracker test with dynamic dependencies")
	}

	// Create fake dependencies
	deps := []domain.Dependency{
		{Repository: "github.com/hashload/horse"},
		{Repository: "github.com/hashload/dataset-serialize"},
	}

	tracker := NewProgressTracker(deps)

	if err := tracker.Start(); err != nil {
		t.Fatalf("Failed to start tracker: %v", err)
	}
	defer tracker.Stop()

	// Simulate installation progress with dynamic dependency discovery
	time.Sleep(500 * time.Millisecond)

	tracker.SetCloning("horse")
	time.Sleep(1 * time.Second)

	// Simulate discovering transitive dependencies
	tracker.AddDependency("dcc")
	tracker.AddDependency("other-dep")
	tracker.SetInstalling("horse")
	tracker.SetCloning("dcc")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("horse")
	tracker.SetInstalling("dcc")
	tracker.SetCloning("other-dep")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("dcc")
	tracker.SetChecking("other-dep", "resolving version")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("other-dep")
	tracker.SetCloning("dataset-serialize")
	time.Sleep(1 * time.Second)

	// Discover more transitive dependencies
	tracker.AddDependency("redis-client")
	tracker.AddDependency("crypto")
	tracker.SetInstalling("dataset-serialize")
	tracker.SetCloning("redis-client")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("dataset-serialize")
	tracker.SetInstalling("redis-client")
	tracker.SetCloning("crypto")
	time.Sleep(1 * time.Second)

	tracker.SetCompleted("redis-client")
	tracker.SetCompleted("crypto")
	time.Sleep(2 * time.Second)
}
