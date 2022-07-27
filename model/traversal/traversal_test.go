package traversal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/util/testutils"
)

func TestTracker_Walk(t *testing.T) {
	type spec struct {
		name           string
		t              Tracker
		root           model.Node
		graph          model.DirectedGraph
		expError       error
		expInvocations int
	}

	cases := []spec{
		{
			name: "Success/VisitRootNode",

			t: Tracker{
				budget: &Budget{
					NodeBudget: 3,
				},
				seen: map[string]struct{}{},
			},
			root: &testutils.MockNode{I: "node1"},
			graph: &mockGraph{nodes: map[string][]model.Node{
				"node1": {&testutils.MockNode{I: "node2"}}},
			},
			expInvocations: 2,
		},
		{
			name: "Success/DuplicateNodeID",
			t: Tracker{
				budget: &Budget{
					NodeBudget: 8,
				},
				seen: map[string]struct{}{},
			},
			root: &testutils.MockNode{I: "node1"},
			graph: &mockGraph{nodes: map[string][]model.Node{
				"node1": {
					&testutils.MockIterableNode{
						I:     "node2",
						Index: -1,
						Nodes: []model.Node{&testutils.MockNode{I: "node1"}}},
				},
			},
			},
			expInvocations: 2,
		},
		{
			name: "Failure/ExceededBudget",

			t: Tracker{
				budget: &Budget{
					NodeBudget: 0,
				},
				seen: map[string]struct{}{},
			},
			root: &testutils.MockNode{I: "node1"},
			graph: &mockGraph{nodes: map[string][]model.Node{
				"node1": {&testutils.MockNode{I: "node2"}}},
			},
			expInvocations: 0,
			expError:       &ErrBudgetExceeded{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actualInvocations int
			visit := func(tr Tracker, n model.Node) error {
				t.Log("Visiting " + n.ID())
				actualInvocations++
				return nil
			}

			err := c.t.Walk(c.root, c.graph, visit)
			if c.expError != nil {
				require.ErrorAs(t, err, &c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expInvocations, actualInvocations)
			}
		})
	}
}

func TestTracker_WalkNested(t *testing.T) {
	type spec struct {
		name           string
		t              Tracker
		root           model.Node
		expError       error
		expInvocations int
	}

	cases := []spec{
		{
			name: "Success/VisitNonIterableNode",

			t: Tracker{
				budget: &Budget{
					NodeBudget: 3,
				},
				seen: map[string]struct{}{},
			},
			root:           &testutils.MockNode{I: "node1"},
			expInvocations: 1,
		},
		{
			name: "Success/WithIterableNode",

			t: Tracker{
				budget: &Budget{
					NodeBudget: 8,
				},
				seen: map[string]struct{}{},
			},
			root: &testutils.MockIterableNode{
				I:     "node2",
				Index: -1,
				Nodes: []model.Node{&testutils.MockNode{I: "node3"}}},
			expInvocations: 2,
		},
		{
			name: "Failure/ExceededBudget",

			t: Tracker{
				budget: &Budget{
					NodeBudget: 0,
				},
				seen: map[string]struct{}{},
			},
			root:           &testutils.MockNode{I: "node1"},
			expInvocations: 0,
			expError:       &ErrBudgetExceeded{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var actualInvocations int
			visit := func(tr Tracker, n model.Node) error {
				t.Log("Visiting " + n.ID())
				actualInvocations++
				return nil
			}

			err := c.t.WalkNested(c.root, visit)
			if c.expError != nil {
				require.ErrorAs(t, err, &c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expInvocations, actualInvocations)
			}
		})
	}
}

// Mock tree structure
type mockGraph struct {
	nodes map[string][]model.Node
}

var _ model.DirectedGraph = &mockGraph{}

func (m *mockGraph) From(id string) []model.Node {
	return m.nodes[id]
}

func (m *mockGraph) To(_ string) []model.Node {
	return nil
}

func (m *mockGraph) Edge(_, _ string) model.Edge {
	return nil
}

func (m *mockGraph) Edges() []model.Edge {
	return nil
}

func (m *mockGraph) Nodes() []model.Node {
	return nil
}

func (m *mockGraph) HasEdgeFromTo(_, _ string) bool {
	return false
}

func (m *mockGraph) NodeByID(_ string) model.Node {
	return nil
}

// Mock Matcher

type mockMatcher struct {
	criteria string
}

var _ model.Matcher = &mockMatcher{}

func (m *mockMatcher) String() string {
	return ""
}

func (m *mockMatcher) Matches(n model.Node) bool {
	return n.ID() == m.criteria
}
