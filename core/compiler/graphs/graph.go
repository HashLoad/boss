package graphs

import (
	"fmt"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"strings"
	"sync"
)

type Node struct {
	Value string
	Dep   models.Dependency
}

func NewNode(dependency *models.Dependency) *Node {
	return &Node{Dep: *dependency, Value: strings.ToLower(dependency.GetName())}
}

func (n *Node) String() string {
	return fmt.Sprintf("%s", n.Dep.GetName())
}

type GraphItem struct {
	nodes     []*Node
	depends   map[string][]*Node
	usedBy    map[string][]*Node
	lockMutex sync.RWMutex
}

func (g *GraphItem) lock() {
	g.lockMutex.Lock()
}

func (g *GraphItem) unlock() {
	g.lockMutex.Unlock()
}

func (g *GraphItem) AddNode(n *Node) {
	g.lock()
	if !contains(g.nodes, n) {
		g.nodes = append(g.nodes, n)
	}
	g.unlock()
}

func contains(a []*Node, x *Node) bool {
	for _, n := range a {
		if x.Value == n.Value {
			return true
		}
	}
	return false
}

func containsOne(a []*Node, b []*Node) bool {
	for _, n := range a {
		for _, x := range b {
			if x.Value == n.Value {
				return true
			}
		}
	}
	return false
}

func containsAll(list []*Node, in []*Node) bool {
	var check = 0
	for _, n := range in {
		for _, x := range list {
			if x.Value == n.Value {
				check++
			}
		}
	}
	return check == len(in)
}

func (g *GraphItem) AddEdge(nLeft, nRight *Node) {
	g.lock()
	if g.depends == nil {
		g.depends = make(map[string][]*Node)
		g.usedBy = make(map[string][]*Node)
	}
	if !contains(g.depends[nLeft.Value], nRight) {
		g.depends[nLeft.Value] = append(g.depends[nLeft.Value], nRight)
	}
	if !contains(g.usedBy[nRight.Value], nLeft) {
		g.usedBy[nRight.Value] = append(g.usedBy[nRight.Value], nLeft)
	}
	g.unlock()
}

func (g *GraphItem) String() {
	g.lock()

	for index := 0; index < len(g.nodes); index++ {
		var node = g.nodes[index]
		var response = ""
		response += g.nodes[index].String() + " -> \n\t\tDepends: "
		nears := g.depends[node.Value]
		for _, near := range nears {
			response += near.String() + " - "
		}

		response += "\n\t\tUsed by: "
		nears = g.usedBy[node.Value]
		for _, near := range nears {
			response += near.String() + " - "
		}
		msg.Info(response)
	}
	g.unlock()
}

func removeNode(nodes []*Node, key int) []*Node {
	if key == len(nodes) {
		return nodes[:key]
	} else {
		return append(nodes[:key], nodes[key+1:]...)
	}
}

func (g *GraphItem) Queue(pkg *models.Package, allDeps bool) NodeQueue {
	g.lock()
	queue := NodeQueue{}
	queue.New()
	nodes := g.nodes
	for key := 0; key < len(nodes); key++ {
		if !pkg.Lock.GetInstalled(nodes[key].Dep).Changed && !allDeps {
			nodes = removeNode(nodes, key)
			key--
		}
	}

	var redo = true
	for {
		if !redo {
			break
		}
		redo = false
		for _, node := range nodes {
			usedBy := g.usedBy[node.Value]
			if !containsAll(nodes, usedBy) {
				for _, consumerNode := range usedBy {
					installed := pkg.Lock.GetInstalled(consumerNode.Dep)
					installed.Changed = true
					pkg.Lock.SetInstalled(consumerNode.Dep, installed)
					if !contains(nodes, consumerNode) {
						redo = true
						nodes = append(nodes, consumerNode)
					}
				}
			}
		}
	}

	for {
		if len(nodes) == 0 {
			break
		}

		for key := 0; key < len(nodes); key++ {
			node := nodes[key]
			if !containsOne(g.depends[node.Value], nodes) {
				queue.Enqueue(*node)
				nodes = removeNode(nodes, key)
				key--
			}
		}
	}
	g.unlock()
	return queue
}

type NodeQueue struct {
	items []Node
	lock  sync.RWMutex
}

func (s *NodeQueue) New() *NodeQueue {
	s.lock.Lock()
	s.items = []Node{}
	s.lock.Unlock()
	return s
}

func (s *NodeQueue) Enqueue(t Node) {
	s.lock.Lock()
	s.items = append(s.items, t)
	s.lock.Unlock()
}

func (s *NodeQueue) Dequeue() *Node {
	s.lock.Lock()
	item := s.items[0]
	s.items = s.items[1:len(s.items)]
	s.lock.Unlock()
	return &item
}

func (s *NodeQueue) Front() *Node {
	s.lock.RLock()
	item := s.items[0]
	s.lock.RUnlock()
	return &item
}

func (s *NodeQueue) IsEmpty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items) == 0
}

func (s *NodeQueue) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items)
}
