package domain_test

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
)

// TestNewNode tests node creation from dependency.
func TestNewNode(t *testing.T) {
	dep := domain.Dependency{
		Repository: "github.com/test/repo",
	}

	node := domain.NewNode(&dep)

	if node == nil {
		t.Fatal("NewNode() returned nil")
	}

	if node.Value == "" {
		t.Error("NewNode() should set Value")
	}

	if node.Dep.Repository != dep.Repository {
		t.Errorf("NewNode() Dep mismatch: got %s, want %s", node.Dep.Repository, dep.Repository)
	}
}

// TestNode_String tests node string representation.
func TestNode_String(t *testing.T) {
	dep := domain.Dependency{
		Repository: "github.com/test/myrepo",
	}

	node := domain.NewNode(&dep)
	str := node.String()

	if str == "" {
		t.Error("Node.String() should not be empty")
	}
}

// TestGraphItem_AddNode tests adding nodes to graph.
func TestGraphItem_AddNode(_ *testing.T) {
	g := &domain.GraphItem{}

	dep1 := domain.Dependency{Repository: "github.com/test/repo1"}
	dep2 := domain.Dependency{Repository: "github.com/test/repo2"}

	node1 := domain.NewNode(&dep1)
	node2 := domain.NewNode(&dep2)

	g.AddNode(node1)
	g.AddNode(node2)

	// Add same node again - should not duplicate
	g.AddNode(node1)
}

// TestGraphItem_AddEdge tests adding edges between nodes.
func TestGraphItem_AddEdge(_ *testing.T) {
	g := &domain.GraphItem{}

	dep1 := domain.Dependency{Repository: "github.com/test/repo1"}
	dep2 := domain.Dependency{Repository: "github.com/test/repo2"}

	node1 := domain.NewNode(&dep1)
	node2 := domain.NewNode(&dep2)

	g.AddNode(node1)
	g.AddNode(node2)

	// Add edge from node1 to node2 (node1 depends on node2)
	g.AddEdge(node1, node2)

	// Add same edge again - should not duplicate
	g.AddEdge(node1, node2)
}

// TestNodeQueue_Operations tests queue operations.
func TestNodeQueue_Operations(t *testing.T) {
	q := &domain.NodeQueue{}
	q.New()

	if !q.IsEmpty() {
		t.Error("New queue should be empty")
	}

	if q.Size() != 0 {
		t.Errorf("New queue size should be 0, got %d", q.Size())
	}

	// Add nodes
	dep1 := domain.Dependency{Repository: "github.com/test/repo1"}
	dep2 := domain.Dependency{Repository: "github.com/test/repo2"}

	node1 := domain.NewNode(&dep1)
	node2 := domain.NewNode(&dep2)

	q.Enqueue(*node1)

	if q.IsEmpty() {
		t.Error("Queue should not be empty after enqueue")
	}

	if q.Size() != 1 {
		t.Errorf("Queue size should be 1, got %d", q.Size())
	}

	q.Enqueue(*node2)

	if q.Size() != 2 {
		t.Errorf("Queue size should be 2, got %d", q.Size())
	}

	// Check front
	front := q.Front()
	if front.Value != node1.Value {
		t.Errorf("Front() should return first node: got %s, want %s", front.Value, node1.Value)
	}

	// Size should not change after Front()
	if q.Size() != 2 {
		t.Errorf("Queue size should still be 2 after Front(), got %d", q.Size())
	}

	// Dequeue
	dequeued := q.Dequeue()
	if dequeued.Value != node1.Value {
		t.Errorf("Dequeue() should return first node: got %s, want %s", dequeued.Value, node1.Value)
	}

	if q.Size() != 1 {
		t.Errorf("Queue size should be 1 after dequeue, got %d", q.Size())
	}

	// Dequeue second
	dequeued = q.Dequeue()
	if dequeued.Value != node2.Value {
		t.Errorf("Dequeue() should return second node: got %s, want %s", dequeued.Value, node2.Value)
	}

	if !q.IsEmpty() {
		t.Error("Queue should be empty after all dequeues")
	}
}
