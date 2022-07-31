package cli

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/content/file"

	"github.com/uor-framework/uor-client-go/attributes"
	"github.com/uor-framework/uor-client-go/content/layout"
	"github.com/uor-framework/uor-client-go/model"
	"github.com/uor-framework/uor-client-go/model/nodes/basic"
	"github.com/uor-framework/uor-client-go/model/nodes/collection"
	"github.com/uor-framework/uor-client-go/model/nodes/descriptor"
	"github.com/uor-framework/uor-client-go/ocimanifest"
	"github.com/uor-framework/uor-client-go/registryclient/orasclient"
	"github.com/uor-framework/uor-client-go/util/examples"
	"github.com/uor-framework/uor-client-go/util/workspace"
)

// PullOptions describe configuration options that can
// be set using the pull subcommand.
type PullOptions struct {
	*RootOptions
	Source     string
	Output     string
	Insecure   bool
	PullAll    bool
	PlainHTTP  bool
	Configs    []string
	Attributes map[string]string
}

var clientPullExamples = examples.Example{
	RootCommand:   filepath.Base(os.Args[0]),
	Descriptions:  []string{"Pull artifacts."},
	CommandString: "pull localhost:5000/myartifacts:latest",
}

// NewPullCmd creates a new cobra.Command for the pull subcommand.
func NewPullCmd(rootOpts *RootOptions) *cobra.Command {
	o := PullOptions{RootOptions: rootOpts}

	cmd := &cobra.Command{
		Use:           "pull SRC",
		Short:         "Pull a UOR collection based on content or attribute address",
		Example:       examples.FormatExamples(clientPullExamples),
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Complete(args))
			cobra.CheckErr(o.Validate())
			cobra.CheckErr(o.Run(cmd.Context()))
		},
	}

	cmd.Flags().StringArrayVarP(&o.Configs, "configs", "c", o.Configs, "auth config paths when contacting registries")
	cmd.Flags().BoolVarP(&o.Insecure, "insecure", "", o.Insecure, "allow connections to SSL registry without certs")
	cmd.Flags().BoolVarP(&o.PlainHTTP, "plain-http", "", o.PlainHTTP, "use plain http and not https when contacting registries")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "output location for artifacts")
	cmd.Flags().StringToStringVarP(&o.Attributes, "attributes", "", o.Attributes, "list of key,value pairs (e.g. key=value) for "+
		"retrieving artifacts by attributes")
	cmd.Flags().BoolVar(&o.PullAll, "pull-all", o.PullAll, "pull all linked collections")

	return cmd
}

func (o *PullOptions) Complete(args []string) error {
	if len(args) < 1 {
		return errors.New("bug: expecting one argument")
	}
	o.Source = args[0]
	if o.Output == "" {
		o.Output = "."
	}
	return nil
}

func (o *PullOptions) Validate() error {
	if _, err := os.Stat(o.Output); err != nil {
		if err := os.MkdirAll(o.Output, 0750); err != nil {
			return err
		}
	}
	return nil
}

func (o *PullOptions) Run(ctx context.Context) error {
	var pullFn pullFunc
	if o.Attributes != nil {
		pullFn = withAttributes
	} else {
		pullFn = func(ctx context.Context, _ PullOptions) (ocispec.Descriptor, error) {
			desc, _, err := o.pullCollection(ctx, o.Output)
			return desc, err
		}
	}
	return o.run(ctx, pullFn)
}

func (o *PullOptions) run(ctx context.Context, pullFn pullFunc) error {
	desc, err := pullFn(ctx, *o)
	if err != nil {
		return err
	}

	o.Logger.Infof("Artifact %s from %s pulled to %s\n", desc.Digest, o.Source, o.Output)
	return nil
}

type pullFunc func(context.Context, PullOptions) (ocispec.Descriptor, error)

func withAttributes(ctx context.Context, o PullOptions) (ocispec.Descriptor, error) {
	cleanup, dir, err := o.mktempDir(o.Output)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	defer cleanup()

	desc, layerDescs, err := o.pullCollection(ctx, dir)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	// Convert descriptor annotations to type Attribute.
	attributesByFile := make(map[string]model.Attributes, len(layerDescs))
	for _, ldesc := range layerDescs {
		filename, ok := ldesc.Annotations[ocispec.AnnotationTitle]
		if !ok {
			continue
		}
		attr := descriptor.AnnotationsToAttributes(ldesc.Annotations)
		o.Logger.Debugf("Adding attributes %s for file %s", attr.String(), filename)
		attributesByFile[filename] = attr
	}

	space, err := workspace.NewLocalWorkspace(dir)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	var nodes []model.Node
	err = space.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("traversing %s: %v", path, err)
		}
		if info == nil {
			return fmt.Errorf("no file info")
		}

		if info.IsDir() {
			return nil
		}

		attr, ok := attributesByFile[path]
		if !ok {
			o.Logger.Debugf("No attributes found for file %s", path)
			return nil
		}

		node := basic.NewNode(path, attr)

		node.Location = space.Path(path)
		nodes = append(nodes, node)
		o.Logger.Debugf("Node %s added", path)
		return nil
	})
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	// Iterator through the collection using an iterator instead
	// of constructing a tree data structure for now since
	// we are handling one collection at first.
	itr := collection.NewByAttributesIterator(nodes)

	return desc, o.moveToResults(itr, o.Attributes)
}

// moveToResult will iterate through the collection
func (o *PullOptions) moveToResults(itr model.Iterator, matcher attributes.PartialAttributeMatcher) error {
	for itr.Next() {
		node := itr.Node()
		if matcher.Matches(node) {
			o.Logger.Debugf("Found match: %q", node.ID())
			newLoc := filepath.Join(o.Output, node.ID())
			cleanLoc := filepath.Clean(newLoc)
			if err := os.MkdirAll(filepath.Dir(cleanLoc), 0750); err != nil {
				return err
			}
			if err := os.Rename(node.Address(), cleanLoc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *PullOptions) pullCollection(ctx context.Context, output string) (ocispec.Descriptor, []ocispec.Descriptor, error) {
	var layerDescs []ocispec.Descriptor
	var mu sync.Mutex
	layerFn := func(_ context.Context, desc ocispec.Descriptor) error {
		mu.Lock()
		defer mu.Unlock()
		layerDescs = append(layerDescs, desc)
		return nil
	}

	cache, err := layout.NewWithContext(ctx, o.cacheDir)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	client, err := orasclient.NewClient(
		orasclient.WithPostCopy(layerFn),
		orasclient.SkipTLSVerify(o.Insecure),
		orasclient.WithAuthConfigs(o.Configs),
		orasclient.WithPlainHTTP(o.PlainHTTP),
		orasclient.WithCache(cache),
	)
	if err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("error configuring client: %v", err)
	}
	defer func() {
		if err := client.Destroy(); err != nil {
			o.Logger.Errorf(err.Error())
		}
	}()

	pullSource := func(source string) (ocispec.Descriptor, error) {
		// TODO(jpower432): Write an method to pull blobs
		// by attribute from the cache to a content.Store.
		desc, err := client.Pull(ctx, source, file.New(output))
		if err != nil {
			return desc, fmt.Errorf("client pull error for reference %s: %v", o.Source, err)
		}

		o.Logger.Debugf("Pulled down %s for reference %s", desc.Digest, source)

		// The cache will be populated by the pull command
		// Ensure the resource is captured in the index.json by
		// tagging the reference.
		if err := cache.Tag(ctx, desc, source); err != nil {
			return desc, err
		}
		return desc, nil
	}

	desc, err := pullSource(o.Source)
	if err != nil {
		return desc, layerDescs, err
	}

	// Resolve source links and all linked collections by BFS.
	if o.PullAll {
		o.Logger.Infof("Resolving linked collections for reference %s", o.Source)
		visitedRefs := map[string]struct{}{o.Source: {}}
		linkedRefs, err := cache.ResolveLinks(ctx, o.Source)
		if err := o.checkResolvedLinksError(o.Source, err); err != nil {
			return desc, layerDescs, err
		}

		for len(linkedRefs) != 0 {
			currRef := linkedRefs[0]
			linkedRefs = linkedRefs[1:]
			if _, ok := visitedRefs[currRef]; ok {
				continue
			}
			visitedRefs[currRef] = struct{}{}

			// No need to log this descriptor digest
			// since it will be logged when pulled.
			_, err := pullSource(currRef)
			if err != nil {
				return desc, layerDescs, err
			}
			o.Logger.Infof("Resolving linked collections for reference %s", currRef)
			currLinks, err := cache.ResolveLinks(ctx, currRef)
			if err := o.checkResolvedLinksError(currRef, err); err != nil {
				return desc, layerDescs, err
			}
			linkedRefs = append(linkedRefs, currLinks...)
		}
	}

	return desc, layerDescs, nil
}

// checkResolvedLinksError logs errors when no collection is
// found.
func (o *PullOptions) checkResolvedLinksError(ref string, err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, ocimanifest.ErrNoCollectionLinks) {
		return err
	}
	o.Logger.Infof("No linked collections found for %s", ref)
	return nil
}

// mkTempDir will make a temporary dir and return the name
// and cleanup method.
func (o *PullOptions) mktempDir(parent string) (func(), string, error) {
	dir, err := ioutil.TempDir(parent, "collection.*")
	return func() {
		if err := os.RemoveAll(dir); err != nil {
			o.Logger.Fatalf(err.Error())
		}
	}, dir, err
}
