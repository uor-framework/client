package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/uor-framework/uor-client-go/attributes"
	"github.com/uor-framework/uor-client-go/content/layout"
	"github.com/uor-framework/uor-client-go/util/examples"
)

// PullOptions describe configuration options that can
// be set using the pull subcommand.
type InspectOptions struct {
	*RootOptions
	Source     string
	Attributes map[string]string
}

var clientInspectExamples = []examples.Example{
	{
		RootCommand:   filepath.Base(os.Args[0]),
		Descriptions:  []string{"Inspect artifacts."},
		CommandString: "inspect localhost:5000/myartifacts:latest",
	},
	{
		RootCommand:   filepath.Base(os.Args[0]),
		Descriptions:  []string{"Inspect artifacts with attributes filtering."},
		CommandString: "inspect localhost:5000/myartifacts:latest --attributes \"size=small\"",
	},
}

// NewPullCmd creates a new cobra.Command for the pull subcommand.
func NewInspectCmd(rootOpts *RootOptions) *cobra.Command {
	o := InspectOptions{RootOptions: rootOpts}

	cmd := &cobra.Command{
		Use:           "inspect SRC",
		Short:         "Print UOR collection information",
		Example:       examples.FormatExamples(clientInspectExamples...),
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Complete(args))
			cobra.CheckErr(o.Validate())
			cobra.CheckErr(o.Run(cmd.Context()))
		},
	}

	cmd.Flags().StringToStringVarP(&o.Attributes, "attributes", "", o.Attributes, "list of key,value pairs (e.g. key=value) for "+
		"retrieving artifacts by attributes")

	return cmd
}

func (o *InspectOptions) Complete(args []string) error {
	if len(args) < 1 {
		return errors.New("bug: expecting one argument")
	}
	o.Source = args[0]
	return nil
}

func (o *InspectOptions) Validate() error {
	return nil
}

func (o *InspectOptions) Run(ctx context.Context) error {
	cache, err := layout.New(ctx, o.cacheDir)
	if err != nil {
		return err
	}

	o.Logger.Debugf("Resolving source %s to descriptor with %d attributes", o.Source, len(o.Attributes))

	var matcher attributes.PartialAttributeMatcher = o.Attributes
	descs, err := cache.ResolveByAttribute(ctx, o.Source, matcher)
	if err != nil {
		return err
	}

	return o.formatDescriptors(o.IOStreams.Out, descs)
}

func (o *InspectOptions) formatDescriptors(w io.Writer, descs []ocispec.Descriptor) error {
	tw := tabwriter.NewWriter(o.IOStreams.Out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintf(tw, "Listing matching descriptors for source:\t%s\n", o.Source); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "Name\tDigest\tSize\tMediaType"); err != nil {
		return err
	}
	for _, desc := range descs {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", desc.Annotations[ocispec.AnnotationTitle], desc.Digest, desc.Size, desc.MediaType); err != nil {
			return err
		}
	}

	return tw.Flush()

}
