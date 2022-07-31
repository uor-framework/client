package builder

import (
	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/model/nodes/collection"
)

var _ model.NodeBuilder = &collectionBuilder{}

type collectionBuilder struct {
	nodes []model.Node
	edges []model.Edge
}

// New returns a builder for collection nodes.
func New(nodes []model.Node, edges []model.Edge) model.NodeBuilder {
	return &collectionBuilder{
		nodes: nodes,
		edges: edges,
	}
}

// Build completes any required actions for assembly
// before return the final immutable collection.
// At node build time create and attach the iterator.
func (b *collectionBuilder) Build(id string) (model.Node, error) {
	c := collection.New(id)
	for _, node := range b.nodes {
		if err := c.AddNode(node); err != nil {
			return nil, err
		}
	}
	for _, edge := range b.edges {
		if err := c.AddEdge(edge); err != nil {
			return nil, err
		}
	}
	itr := collection.NewByAttributesIterator(c.Nodes())
	c.ByAttributesIterator = itr
	return c, nil
}
