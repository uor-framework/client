package collection

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/util/testutils"
)

func TestCollection_Root(t *testing.T) {
	type spec struct {
		name     string
		nodes    []model.Node
		edges    []model.Edge
		expID    string
		expError string
	}

	cases := []spec{
		{
			name:  "Success/RootExists",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			expID: "node3",
		},
		{
			name:  "Failure/NotRootExists",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
				&Edge{T: &testutils.MockNode{I: "node3"}, F: &testutils.MockNode{I: "node2"}},
			},
			expError: "no root found in graph",
		},
		{
			name:  "Failure/MultipleRootsExist",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
			},
			expError: "multiple roots found in graph: address, address",
		},
		{
			name:     "Failure/NoEdges",
			nodes:    []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges:    nil,
			expError: "multiple roots found in graph: address, address, address",
		},
		{
			name:     "Failure/NoNodes",
			nodes:    nil,
			edges:    nil,
			expError: "no root found in graph",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			collection := makeTestCollection(t, c.nodes, c.edges)
			root, err := collection.Root()
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expID, root.ID())
			}

		})
	}
}

func TestCollection_HasEdgeToFrom(t *testing.T) {
	type spec struct {
		name  string
		nodes []model.Node
		edges []model.Edge
		to    string
		from  string
		exp   bool
	}

	cases := []spec{
		{
			name:  "Success/EdgeExists",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			to:   "node1",
			from: "node2",
			exp:  true,
		},
		{
			name:  "Success/NoEdgeExists",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			to:   "node2",
			from: "node3",
			exp:  false,
		},
		{
			name:  "Success/EdgeExitsReverse",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
				&Edge{T: &testutils.MockNode{I: "node3"}, F: &testutils.MockNode{I: "node2"}},
			},
			to:   "node2",
			from: "node1",
			exp:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			collection := makeTestCollection(t, c.nodes, c.edges)
			actual := collection.HasEdgeFromTo(c.from, c.to)
			require.Equal(t, c.exp, actual)
		})
	}
}

func TestCollection_To(t *testing.T) {
	type spec struct {
		name   string
		nodes  []model.Node
		edges  []model.Edge
		input  string
		expIDs []string
	}

	cases := []spec{
		{
			name:  "Success/OneNodeFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			input:  "node2",
			expIDs: []string{"node1"},
		},
		{
			name:  "Success/MultipleNodesFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node2"}},
			},
			input:  "node1",
			expIDs: []string{"node2", "node3"},
		},
		{
			name:  "Success/NoNodesFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			input:  "node3",
			expIDs: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			collection := makeTestCollection(t, c.nodes, c.edges)
			nodes := collection.To(c.input)
			var actual []string
			for _, node := range nodes {
				actual = append(actual, node.ID())
			}
			sort.Strings(actual)
			require.Equal(t, c.expIDs, actual)
		})
	}
}

func TestCollection_From(t *testing.T) {
	type spec struct {
		name   string
		nodes  []model.Node
		edges  []model.Edge
		input  string
		expIDs []string
	}

	cases := []spec{
		{
			name:  "Success/OneNodeFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node2"}},
			},
			input:  "node1",
			expIDs: []string{"node2"},
		},
		{
			name:  "Success/MultipleNodesFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node3"}},
			},
			input:  "node3",
			expIDs: []string{"node1", "node2"},
		},
		{
			name:  "Success/NoNodesFound",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			input:  "node2",
			expIDs: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			collection := makeTestCollection(t, c.nodes, c.edges)
			nodes := collection.From(c.input)
			var actual []string
			for _, node := range nodes {
				actual = append(actual, node.ID())
			}
			sort.Strings(actual)
			require.Equal(t, c.expIDs, actual)
		})
	}
}

func TestCollection_Attributes(t *testing.T) {
	type spec struct {
		name          string
		nodes         []model.Node
		edges         []model.Edge
		expAttributes string
	}

	cases := []spec{
		{
			name: "Success/RootExists",
			nodes: []model.Node{
				&testutils.MockNode{
					I: "node1",
					A: &mockAttributes{
						"title": map[string]struct{}{
							"node1": {},
						},
					},
				},
				&testutils.MockNode{
					I: "node2",
					A: &mockAttributes{
						"title": map[string]struct{}{
							"node2": {},
						},
					},
				},
				&testutils.MockNode{
					I: "node3",
					A: &mockAttributes{
						"title": map[string]struct{}{
							"node3": {},
						},
					},
				},
			},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
			},
			expAttributes: "title=node3",
		},
		{
			name:  "Failure/NotRootExists",
			nodes: []model.Node{&testutils.MockNode{I: "node1"}, &testutils.MockNode{I: "node2"}, &testutils.MockNode{I: "node3"}},
			edges: []model.Edge{
				&Edge{T: &testutils.MockNode{I: "node2"}, F: &testutils.MockNode{I: "node1"}},
				&Edge{T: &testutils.MockNode{I: "node1"}, F: &testutils.MockNode{I: "node3"}},
				&Edge{T: &testutils.MockNode{I: "node3"}, F: &testutils.MockNode{I: "node2"}},
			},
			expAttributes: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			collection := makeTestCollection(t, c.nodes, c.edges)
			attr := collection.Attributes()
			if attr == nil {
				require.Len(t, c.expAttributes, 0)
			} else {
				require.Equal(t, c.expAttributes, attr.String())
			}
		})
	}
}

func makeTestCollection(t *testing.T, nodes []model.Node, edges []model.Edge) Collection {
	c := New("test")
	for _, node := range nodes {
		require.NoError(t, c.AddNode(node))
	}
	for _, edge := range edges {
		require.NoError(t, c.AddEdge(edge))
	}
	return *c
}

type mockAttributes map[string]map[string]struct{}

var _ model.Attributes = &mockAttributes{}

func (m mockAttributes) Find(key string) []string {
	valSet, exists := m[key]
	if !exists {
		return nil
	}
	var vals []string
	for val := range valSet {
		vals = append(vals, val)
	}
	return vals
}

func (m mockAttributes) Exists(key, value string) bool {
	vals, exists := m[key]
	if !exists {
		return false
	}
	_, valExists := vals[value]
	return valExists
}

func (m mockAttributes) String() string {
	out := new(strings.Builder)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for val := range m[key] {
			line := fmt.Sprintf("%s=%s,", key, val)
			out.WriteString(line)
		}
	}
	return strings.TrimSuffix(out.String(), ",")
}

func (m mockAttributes) Merge(_ model.Attributes) {
	// Not implemented
}

func (m mockAttributes) List() map[string][]string {
	return nil
}

func (m mockAttributes) Len() int {
	return len(m)
}
