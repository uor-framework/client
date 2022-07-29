package traversal

import (
	"github.com/uor-framework/uor-client-go/model"
)

// Path describes a series of steps across a DAG of model.Node.
type Path struct {
	prev  map[string]string
	index map[string]model.Node
}

// NewPath returns a Path with an initial node.
func NewPath(n model.Node) Path {
	return Path{
		prev: map[string]string{
			n.ID(): "none",
		},
		index: map[string]model.Node{
			n.ID(): n,
		},
	}
}

// Add adds the current node to the path with the previous
// node to indicate position.
func (p Path) Add(prev model.Node, curr model.Node) Path {
	p.prev[curr.ID()] = prev.ID()
	if _, ok := p.index[curr.ID()]; !ok {
		p.index[curr.ID()] = curr
	}

	if _, ok := p.index[prev.ID()]; !ok {
		p.index[prev.ID()] = curr
	}
	return p
}

// Len returns the length of the path.
func (p Path) Len() int {
	return len(p.index)
}

// Prev return the previous node in the path
// for the specified node.
func (p Path) Prev(n model.Node) model.Node {
	parentID := p.prev[n.ID()]
	return p.index[parentID]

}

// List returns a path from the specified end node to
// the initial node.
func (p Path) List(end model.Node) []model.Node {
	path := []model.Node{end}
	for next := p.prev[end.ID()]; next != "none"; next = p.prev[next] {
		nextNode := p.index[next]
		path = append(path, nextNode)
	}

	// Reverse the path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}
