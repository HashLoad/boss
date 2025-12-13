package installer_test

import (
	"sync"
	"testing"

	"github.com/hashload/boss/internal/core/services/installer"
)

// TestDependencyCache_NewDependencyCache tests cache initialization.
func TestDependencyCache_NewDependencyCache(t *testing.T) {
	cache := installer.NewDependencyCache()

	if cache == nil {
		t.Fatal("NewDependencyCache() returned nil")
	}

	// New cache should report nothing as updated
	if cache.IsUpdated("any-dep") {
		t.Error("New cache should have no dependencies marked as updated")
	}
}

// TestDependencyCache_IsUpdated tests checking update status.
func TestDependencyCache_IsUpdated(t *testing.T) {
	cache := installer.NewDependencyCache()

	// Initially not updated
	if cache.IsUpdated("test-dep") {
		t.Error("IsUpdated() should return false for new dependency")
	}

	// After marking
	cache.MarkUpdated("test-dep")
	if !cache.IsUpdated("test-dep") {
		t.Error("IsUpdated() should return true after MarkUpdated()")
	}

	// Other deps still not updated
	if cache.IsUpdated("other-dep") {
		t.Error("IsUpdated() should return false for different dependency")
	}
}

// TestDependencyCache_MarkUpdated tests marking dependencies.
func TestDependencyCache_MarkUpdated(t *testing.T) {
	cache := installer.NewDependencyCache()

	cache.MarkUpdated("dep1")
	cache.MarkUpdated("dep2")
	cache.MarkUpdated("dep3")

	if !cache.IsUpdated("dep1") || !cache.IsUpdated("dep2") || !cache.IsUpdated("dep3") {
		t.Error("All marked dependencies should be updated")
	}

	// Marking same dep twice should not cause issues
	cache.MarkUpdated("dep1")
	if !cache.IsUpdated("dep1") {
		t.Error("Dependency should still be marked after duplicate MarkUpdated()")
	}
}

// TestDependencyCache_Concurrency tests thread safety.
func TestDependencyCache_Concurrency(t *testing.T) {
	cache := installer.NewDependencyCache()
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Writers
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for range numOperations {
				cache.MarkUpdated("dep-" + string(rune('A'+id%26)))
			}
		}(i)
	}

	// Readers
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for range numOperations {
				_ = cache.IsUpdated("dep-" + string(rune('A'+id%26)))
			}
		}(i)
	}

	wg.Wait()

	// Should complete without race conditions or panics
	// At least one dependency should be marked
	hasAny := false
	for i := range 26 {
		if cache.IsUpdated("dep-" + string(rune('A'+i))) {
			hasAny = true
			break
		}
	}
	if !hasAny {
		t.Error("Cache should have some entries after concurrent writes")
	}
}
