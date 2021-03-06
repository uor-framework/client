package content

import (
	"oras.land/oras-go/v2/content"
)

// Store defines the methods for adding, inspecting, and removing
// OCI content from a storage location. The interface wraps oras
// Storage and TagResolver interfaces for use with `oras` Copy methods.
type Store interface {
	// Storage represents a content-addressable storage where contents are
	// accessed via Descriptors.
	content.Storage
	// TagResolver defines methods for indexing tags.
	content.TagResolver
}
