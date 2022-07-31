package content

import (
	"context"

	"oras.land/oras-go/v2/content"
)

// Store defines the methods for adding, inspecting, and removing
// OCI content from a storage location. The interface wraps `oras``
// Storage and TagResolver interfaces for use with `oras` Copy methods.
type Store interface {
	// Storage represents a content-addressable storage where contents are
	// accessed via Descriptors.
	content.Storage
	// TagResolver defines methods for indexing tags.
	content.TagResolver
}

// GraphStore defines the methods for adding, inspecting, and removing
// OCI content from a storage location. The interface wraps `oras`
// Storage, TagResolver, and PredecessorFinder interfaces for use with `oras` extended Copy methods.
type GraphStore interface {
	Store
	// PredecessorFinder returns the nodes directly pointing to the current node.
	content.PredecessorFinder
	// ResolveLinks returns all sub-collections references that are linked
	// to the root node.
	ResolveLinks(context.Context, string) ([]string, error)
}
