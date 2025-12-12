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

	if cache.Count() != 0 {
		t.Errorf("New cache should be empty, got count %d", cache.Count())
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

	if cache.Count() != 3 {
		t.Errorf("Count() should be 3, got %d", cache.Count())
	}

	// Marking same dep twice should not increase count
	cache.MarkUpdated("dep1")
	if cache.Count() != 3 {
		t.Errorf("Count() should still be 3 after duplicate, got %d", cache.Count())
	}
}

// TestDependencyCache_Reset tests clearing the cache.
func TestDependencyCache_Reset(t *testing.T) {
	cache := installer.NewDependencyCache()

	cache.MarkUpdated("dep1")
	cache.MarkUpdated("dep2")

	if cache.Count() != 2 {
		t.Fatalf("Count() should be 2 before reset, got %d", cache.Count())
	}

	cache.Reset()

	if cache.Count() != 0 {
		t.Errorf("Count() should be 0 after Reset(), got %d", cache.Count())
	}

	if cache.IsUpdated("dep1") {
		t.Error("IsUpdated() should return false after Reset()")
	}
}

// TestDependencyCache_Concurrency tests thread safety.
func TestDependencyCache_Concurrency(t *testing.T) {
	cache := installer.NewDependencyCache()
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

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

	// Count readers
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range numOperations {
				_ = cache.Count()
			}
		}()
	}

	wg.Wait()

	// Should complete without race conditions or panics
	if cache.Count() == 0 {
		t.Error("Cache should have some entries after concurrent writes")
	}
}
