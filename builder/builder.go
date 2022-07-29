package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/opencontainers/go-digest"

	"github.com/uor-framework/uor-client-go/builder/parser"
	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/model/nodes/collection"
	"github.com/uor-framework/uor-client-go/util/workspace"
)

// Builder defines methods for building UOR datasets.
type Builder interface {
	Run(context.Context, *collection.Collection, workspace.Workspace) error
}

// Compatibility renders and writes templates from the source workspace
// in compatibility mode.
// Compatibility mode renders the template for each node representing a
// file replacing any links with the content address of the linked file.
type Compatibility struct {
	// Source workspace to find files
	Source workspace.Workspace
	// Links association with the node ID
	Links map[string]map[string]interface{}
	// Templates association with the node ID
	Templates map[string]template.Template
}

var _ Builder = &Compatibility{}

// NewCompatibilityBuilder creates a new Builder from the source
// workspace.
func NewCompatibilityBuilder(source workspace.Workspace) Compatibility {
	return Compatibility{
		Source:    source,
		Links:     map[string]map[string]interface{}{},
		Templates: map[string]template.Template{},
	}
}

// Run traverses the graph with build nodes to render the file templates to the destination workspace.
// All nodes are expected have an underlying concrete type of BuildNode.
func (b Compatibility) Run(ctx context.Context, c *collection.Collection, destination workspace.Workspace) error {
	root, err := c.Root()
	if err != nil {
		return fmt.Errorf("error calculating root node: %v", err)
	}

	// Links stores the calculated result of each sub-problem (i.e. link hashes)
	links := make(map[string]interface{})
	return b.makeTemplates(ctx, c, root, destination, links)
}

// makeTemplates does recursive DFS traversal of the graph to generate digest values and template files.
func (b Compatibility) makeTemplates(ctx context.Context, c *collection.Collection, start model.Node, destination workspace.Workspace, links map[string]interface{}) error {
	if start == nil {
		return nil
	}

	// Template and hash each child node to
	// calculate parent node information
	for _, n := range c.From(start.ID()) {
		if _, found := links[n.Address()]; found {
			continue
		}

		if err := b.makeTemplates(ctx, c, n, destination, links); err != nil {
			return err
		}
	}

	buf := new(bytes.Buffer)
	if b.isBuildable(start.ID()) {
		// Update all links data with currently accumulated
		// digest values and render the new file from template.
		nodeLinks, ok := b.Links[start.ID()]
		if !ok {
			return fmt.Errorf("buildable node %s has no values", start.Address())
		}
		b.Links[start.ID()] = mergeLinkData(nodeLinks, links)
		if err := b.render(buf, start.ID()); err != nil {
			return err
		}
	} else {
		if err := b.Source.ReadObject(ctx, start.Address(), buf); err != nil {
			return err
		}
	}

	if err := destination.WriteObject(ctx, start.Address(), buf.Bytes()); err != nil {
		return err
	}

	// Must calculate the digest after writing the content of
	// the buffer to file because the FromReader method consumes the data.
	dgst, err := digest.FromReader(buf)
	if err != nil {
		return err
	}

	templateValue := parser.ConvertFilenameForGoTemplateValue(start.Address())
	links[templateValue] = dgst
	return nil
}

// isBuildable will check the node ID able to be built with
// the render method.
func (b Compatibility) isBuildable(id string) bool {
	tmpl, ok := b.Templates[id]
	if !ok {
		return false
	}
	return tmpl != (template.Template{})
}

// render will render the template for the current state of the
// node at the specified ID.
func (b Compatibility) render(w io.Writer, id string) error {
	tmpl, ok := b.Templates[id]
	if !ok {
		return fmt.Errorf("no template associated with node %v", id)
	}

	values, ok := b.Links[id]
	if !ok {
		return fmt.Errorf("no links associated with node %v", id)
	}

	return tmpl.Execute(w, values)
}

// mergeLinkData will merge any references to in-content links with
// the currently calculated values.
func mergeLinkData(in, curr map[string]interface{}) map[string]interface{} {
	for key := range in {
		currentVal, ok := curr[key]
		if ok {
			in[key] = currentVal
		}
	}
	return in
}
