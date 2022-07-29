package ocimanifest

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/uor-framework/uor-client-go/builder/api/v1alpha1"
	"github.com/uor-framework/uor-client-go/registryclient"
)

const (
	// AnnotationSchema is the reference to the
	// default schema of the collection.
	AnnotationSchema = "uor.schema"
	// AnnotationSchemaLinks is the reference to linked
	// schemas for a collection. This will define all referenced
	// schemas for the collection and sub-collection. The tree will
	// be fully resolved.
	AnnotationSchemaLinks = "uor.schema.linked"
	// AnnotationCollectionLinks references the collections
	// that are linked to a collection node. The will only
	// reference adjacent collection and will not descend
	// into sub-collections.
	AnnotationCollectionLinks = "uor.collections.linked"
	// Separator is the value used to denote a list of
	// schema or collection in a manifest.
	Separator = ","
	// UORConfigMediaType is the manifest config media type
	// for UOR OCI manifests.
	UORConfigMediaType = "application/vnd.uor.config.v1+json"
)

var (
	// ErrNoKnownSchema denotes that no schema
	// annotation is set on the manifest.
	ErrNoKnownSchema = errors.New("no schema")
	// ErrNoCollectionLinks denotes that the manifest
	// does contain annotation that set collection links.
	ErrNoCollectionLinks = errors.New("no collection links")
)

// FetchSchema fetches schema information for a reference.
func FetchSchema(ctx context.Context, reference string, client registryclient.Remote) (string, []string, error) {
	_, manBytes, err := client.GetManifest(ctx, reference)
	if err != nil {
		return "", nil, err
	}

	var manifest ocispec.Manifest
	if err := json.NewDecoder(manBytes).Decode(&manifest); err != nil {
		return "", nil, err
	}

	schema, ok := manifest.Annotations[AnnotationSchema]
	if !ok {
		return "", nil, ErrNoKnownSchema
	}

	links, ok := manifest.Annotations[AnnotationSchemaLinks]
	if !ok {
		return schema, nil, nil
	}

	return schema, []string{links}, nil
}

// UpdateLayerDescriptors updates layers descriptor annotations with user provided key,value pairs
func UpdateLayerDescriptors(descs []ocispec.Descriptor, cfg v1alpha1.DataSetConfiguration) ([]ocispec.Descriptor, error) {
	var updateDescs []ocispec.Descriptor
	for _, desc := range descs {
		filename, ok := desc.Annotations[ocispec.AnnotationTitle]
		if !ok {
			// skip any descriptor with no name attached
			continue
		}
		for _, file := range cfg.Files {
			// If the config has a grouping declared, make a valid regex.
			if strings.Contains(file.File, "*") && !strings.Contains(file.File, ".*") {
				file.File = strings.Replace(file.File, "*", ".*", -1)
			} else {
				file.File = strings.Replace(file.File, file.File, "^"+file.File+"$", -1)
			}
			namesearch, err := regexp.Compile(file.File)
			if err != nil {
				return nil, err
			}

			if namesearch.Match([]byte(filename)) {
				// Get the k/v pairs from the config and add them to the descriptor annotations.
				for k, v := range file.Attributes {
					desc.Annotations[k] = v
				}
			}
		}
		updateDescs = append(updateDescs, desc)
	}

	return updateDescs, nil
}
