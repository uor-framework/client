package layout

import (
	"context"
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/v1/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"

	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/model/nodes/collection"
	"github.com/uor-framework/uor-client-go/model/nodes/descriptor"
)

// ManifestToCollection converts a UOR managed OCI manifest to a Collection.
func ManifestToCollection(ctx context.Context, graph *collection.Collection, fetcher content.Fetcher, manifest ocispec.Descriptor) error {
	return manifestToCollection(ctx, graph, fetcher, manifest)
}

// manifestToCollection recursively adds nodes to the index based on media type.
func manifestToCollection(ctx context.Context, graph *collection.Collection, fetcher content.Fetcher, node ocispec.Descriptor) error {
	switch node.MediaType {
	case string(types.DockerManifestList), ocispec.MediaTypeImageIndex:
		c, err := content.FetchAll(ctx, fetcher, node)
		if err != nil {
			return err
		}

		var index ocispec.Index
		if err := json.Unmarshal(c, &index); err != nil {
			return err
		}
		return indexNode(graph, node, index.Manifests)
	default:
		c, err := content.FetchAll(ctx, fetcher, node)
		if err != nil {
			return err
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(c, &manifest); err != nil {
			return err
		}

		return indexNode(graph, node, append([]ocispec.Descriptor{manifest.Config}, manifest.Layers...))
	}
}

// addManifest will add a single manifest to the collection
func addManifest(ctx context.Context, graph *collection.Collection, fetcher content.Fetcher, manifest ocispec.Descriptor) error {
	successors, err := content.Successors(ctx, fetcher, manifest)
	if err != nil {
		return err
	}
	return indexNode(graph, manifest, successors)
}

// indexNode indexes relationships between child and parent nodes.
func indexNode(graph *collection.Collection, node ocispec.Descriptor, successors []ocispec.Descriptor) error {
	n, err := addOrGetNode(graph, node)
	if err != nil {
		return err
	}
	for _, successor := range successors {
		s, err := addOrGetNode(graph, successor)
		if err != nil {
			return err
		}
		e := collection.NewEdge(n, s)
		if err := graph.AddEdge(e); err != nil {
			return err
		}
	}
	return nil
}

// addOrGetNode will return the node if it exists in the graph or will create a new
// descriptor node.
func addOrGetNode(graph *collection.Collection, desc ocispec.Descriptor) (model.Node, error) {
	n := graph.NodeByID(desc.Digest.String())
	if n != nil {
		return n, nil
	}
	n = descriptor.NewNode(desc.Digest.String(), desc)
	if err := graph.AddNode(n); err != nil {
		return nil, err
	}
	return n, nil
}
