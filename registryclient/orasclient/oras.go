package orasclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/gabriel-vasile/mimetype"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"github.com/uor-framework/uor-client-go/content"
	"github.com/uor-framework/uor-client-go/registryclient"
	"github.com/uor-framework/uor-client-go/registryclient/orasclient/internal/cache"
)

type orasClient struct {
	insecure      bool
	plainHTTP     bool
	configs       []string
	copyOpts      oras.CopyOptions
	artifactStore *file.Store
	cache         content.Store
	destroy       func() error
	outputDir     string
}

var _ registryclient.Client = &orasClient{}

// AddFiles loads one or more files to create OCI descriptors with a specific
// media type and pushes them into underlying storage.
func (c *orasClient) AddFiles(ctx context.Context, mediaType string, files ...string) ([]ocispec.Descriptor, error) {
	if err := c.checkFileStore(); err != nil {
		return nil, err
	}
	descs, err := loadFiles(ctx, c.artifactStore, mediaType, files...)
	if err != nil {
		return nil, fmt.Errorf("unable to load files: %w", err)
	}
	return descs, nil
}

// AddContent creates and stores a descriptor from content in bytes, a media type, and
// annotations.
func (c *orasClient) AddContent(ctx context.Context, mediaType string, content []byte, annotations map[string]string) (ocispec.Descriptor, error) {
	if err := c.checkFileStore(); err != nil {
		return ocispec.Descriptor{}, err
	}
	configDesc := ocispec.Descriptor{
		MediaType:   mediaType,
		Digest:      digest.FromBytes(content),
		Size:        int64(len(content)),
		Annotations: annotations,
	}

	return configDesc, c.artifactStore.Push(ctx, configDesc, bytes.NewReader(content))
}

// AddManifest creates and stores a manifest.
// This is generated from the config descriptor and artifact descriptors.
func (c *orasClient) AddManifest(ctx context.Context, ref string, configDesc ocispec.Descriptor, manifestAnnotations map[string]string, descriptors ...ocispec.Descriptor) (ocispec.Descriptor, error) {
	if err := c.checkFileStore(); err != nil {
		return ocispec.Descriptor{}, err
	}
	if descriptors == nil {
		descriptors = []ocispec.Descriptor{}
	}

	// Keep descriptor order deterministic
	sort.Slice(descriptors, func(i, j int) bool {
		return descriptors[i].Digest < descriptors[j].Digest
	})

	var packOpts oras.PackOptions
	packOpts.ConfigDescriptor = &configDesc
	packOpts.ManifestAnnotations = manifestAnnotations

	manifestDesc, err := oras.Pack(ctx, c.artifactStore, descriptors, packOpts)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	return manifestDesc, c.artifactStore.Tag(ctx, manifestDesc, ref)
}

// Save save the OCI artifact to local store location (e.g. cache)
func (c *orasClient) Save(ctx context.Context, ref string, store content.Store) (ocispec.Descriptor, error) {
	return oras.Copy(ctx, c.artifactStore, ref, store, ref, c.copyOpts)
}

// Pull performs a copy of OCI artifacts to a local location from a remote location.
func (c *orasClient) Pull(ctx context.Context, ref string, store content.Store) (ocispec.Descriptor, error) {
	var from oras.Target
	repo, err := c.setupRepo(ref)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("could not create registry target: %w", err)
	}
	from = repo

	if c.cache != nil {
		from = cache.New(repo, c.cache)
	}

	return oras.Copy(ctx, from, ref, store, ref, c.copyOpts)
}

// Push performs a copy of OCI artifacts to a remote location.
func (c *orasClient) Push(ctx context.Context, store content.Store, ref string) (ocispec.Descriptor, error) {
	repo, err := c.setupRepo(ref)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("could not create registry target: %w", err)
	}

	return oras.Copy(ctx, store, ref, repo, ref, c.copyOpts)
}

// GetManifest returns the manifest the reference resolves to.
func (c *orasClient) GetManifest(ctx context.Context, reference string) (ocispec.Descriptor, io.ReadCloser, error) {
	repo, err := c.setupRepo(reference)
	if err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("could not create registry target: %w", err)
	}
	return repo.FetchReference(ctx, reference)
}

// Store returns the source storage being used to stored
// the OCI artifact.
func (c *orasClient) Store() (content.Store, error) {
	return c.artifactStore, nil
}

// Destroy cleans up any temporary on-disk resources used to track descriptors.
func (c *orasClient) Destroy() error {
	return c.destroy()
}

// checkFileStore ensure that the file store
// has been initialized.
func (c *orasClient) checkFileStore() error {
	if c.artifactStore == nil {
		return errors.New("file store uninitialized")
	}
	return nil
}

// setupRepo configures the client to access the remote repository.
func (c *orasClient) setupRepo(ref string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("could not create registry target: %w", err)
	}
	repo.PlainHTTP = c.plainHTTP
	authC, err := c.authClient()
	if err != nil {
		return nil, err
	}
	repo.Client = authC
	return repo, nil
}

// authClient gather authorization information
// for registry access from provided and default configuration
// files.
func (c *orasClient) authClient() (*auth.Client, error) {
	client := &auth.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: c.insecure,
				},
			},
		},
		Cache: auth.NewCache(),
	}

	store, err := NewAuthStore(c.configs...)
	if err != nil {
		return nil, err
	}
	client.Credential = store.Credential
	return client, nil
}

// loadFiles stores files in a file store and creates descriptors representing each file in the store.
func loadFiles(ctx context.Context, store *file.Store, mediaType string, files ...string) ([]ocispec.Descriptor, error) {
	var descs []ocispec.Descriptor
	var skipMediaTypeDetection bool
	var err error

	if mediaType != "" {
		skipMediaTypeDetection = true
	}
	for _, fileRef := range files {
		name := filepath.Clean(fileRef)
		if !filepath.IsAbs(name) {
			// convert to slash-separated path unless it is absolute path
			name = filepath.ToSlash(name)
		}

		if !skipMediaTypeDetection {
			mediaType, err = getDefaultMediaType(fileRef)
			if err != nil {
				return nil, fmt.Errorf("file %q: error dectecting media type: %v", name, err)
			}
		}

		desc, err := store.Add(ctx, name, mediaType, fileRef)
		if err != nil {
			return nil, err
		}
		descs = append(descs, desc)
	}
	return descs, nil
}

// getDefaultMediaType detects the media type of the
// file based on content.
func getDefaultMediaType(file string) (string, error) {
	mType, err := mimetype.DetectFile(file)
	if err != nil {
		return "", err
	}
	return mType.String(), nil
}
