package traversal

import (
	"errors"

	"github.com/uor-framework/uor-client-go/model"
)

// ErrSkip allows a node to be intentionally skipped.
var ErrSkip = errors.New("skip")

// Tracker contains information stored during graph traversal
// such as the tree structure to guide traversal direction, budgeting information,
// and visited nodes.
type Tracker struct {
	// Path defines a series of steps
	// during traversal over nodes.
	Path
	// Budget tracks traversal maximums such as maximum visited nodes.
	budget *Budget
	// Seen track which nodes have been visited by ID.
	seen map[string]struct{}
}

// NewTracker returns a new Tracker instance.
func NewTracker(root model.Node, budget *Budget) Tracker {
	t := Tracker{
		budget: budget,
		Path:   NewPath(root),
		seen:   map[string]struct{}{},
	}
	return t
}

// VisitFunc is a read-only visitor for model.Node.
type VisitFunc func(Tracker, model.Node) error

// Walk traverses a series of Node per the graph.
func Walk(start model.Node, graph model.DirectedGraph, fn VisitFunc) error {
	tracker := NewTracker(start, nil)
	return tracker.Walk(start, graph, fn)
}

// WalkNested visits the current nodes and visits
// nested nodes.
func WalkNested(start model.Node, fn VisitFunc) error {
	tracker := NewTracker(start, nil)
	return tracker.WalkNested(start, fn)
}

// Walk performs traversal using an iterative DFS algorithm to
// visit as many nodes as possible until the node budget is hit
// or the whole graph is traversed.
func (t Tracker) Walk(start model.Node, graph model.DirectedGraph, fn VisitFunc) error {
	t.Path = NewPath(start)
	return t.walkIterative(start, graph, func(t Tracker, n model.Node) error {
		return fn(t, n)
	})
}

// WalkNested walks a tree of Nodes, visiting each of them,
// and calling the given VisitFn on all of them. WalkNested can be used inside the
// Walk function to visit subtree of individual nodes.
func (t Tracker) WalkNested(start model.Node, fn VisitFunc) error {
	t.Path = NewPath(start)
	return t.walkNested(start, func(t Tracker, n model.Node) error {
		return fn(t, n)
	})
}

// walkIterative uses an iterative DFS algorithm to traverse the graph
// of model.Node types. Each node is visited and the given VisitFunc
// is called on all of them.
func (t Tracker) walkIterative(n model.Node, graph model.DirectedGraph, fn VisitFunc) error {
	if n == nil {
		return nil
	}

	// Starting simple using a slice to implement a stack.
	stack := []model.Node{n}

	for len(stack) != 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, seen := t.seen[n.ID()]; seen {
			continue
		}

		if t.budget != nil {
			if t.budget.NodeBudget <= 0 {
				return &ErrBudgetExceeded{Node: n}
			}
			t.budget.NodeBudget--
		}

		// Visit the current node.
		t.seen[n.ID()] = struct{}{}
		if err := fn(t, n); err != nil {
			if errors.Is(err, ErrSkip) {
				continue
			}
			return err
		}

		// Iterate over adjacent nodes per the graph.
		for _, node := range graph.From(n.ID()) {
			t.Path.Add(n, node)
			stack = append(stack, node)
		}
	}
	return nil
}

// walkNested uses an iterative DFS algorithm to traverse the tree
// of model.Node types. Each node is visited and the given VisitFunc
// is called on all of them.
func (t Tracker) walkNested(n model.Node, fn VisitFunc) error {
	if n == nil {
		return nil
	}

	// Starting simple using a slice to implement a stack.
	stack := []model.Node{n}

	for len(stack) != 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, seen := t.seen[n.ID()]; seen {
			continue
		}

		if t.budget != nil {
			if t.budget.NodeBudget <= 0 {
				return &ErrBudgetExceeded{Node: n}
			}
			t.budget.NodeBudget--
		}

		// Visit the current node.
		t.seen[n.ID()] = struct{}{}
		if err := fn(t, n); err != nil {
			if errors.Is(err, ErrSkip) {
				continue
			}
			return err
		}

		// Add nodes to stack if the node type implements an iterator
		// (i.e. this a node of nodes)
		itr, ok := n.(model.Iterator)
		if ok {
			for itr.Next() {
				if err := itr.Error(); err != nil {
					return err
				}
				currNode := itr.Node()
				t.Path.Add(n, currNode)
				stack = append(stack, itr.Node())
			}
		}
	}
	return nil
}
