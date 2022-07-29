package cli

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/uor-framework/uor-client-go/cli/log"
)

// RootOptions describe global configuration options that can be set.
type RootOptions struct {
	IOStreams genericclioptions.IOStreams
	LogLevel  string
	Logger    log.Logger
	cacheDir  string
}

var clientLong = templates.LongDesc(
	`
	The UOR client helps you build, publish, and retrieve UOR collections as an OCI artifact.

	The workflow to publish a collection is to gather files for a collection in a directory workspace 
	and use the build sub-command. During the build process, the tag for the
	remote destination is specified. 
	
	This build action will store the collection in a build cache. This location can be specified with the UOR_CACHE environment 
	variable. The default location is ~/.uor/cache. 
	
	After the collection has been stored, it can be retrieved and pushed the to registry with the push sub-command.

	Collections can be retrieved from the cache or the remote location (if not stored) with the pull sub-command. The pull sub-command also
	allows for filtering of the collection with the attributes flag.
	`,
)

// NewRootCmd creates a new cobra.Command for the command root.
func NewRootCmd() *cobra.Command {
	o := RootOptions{}

	o.IOStreams = genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	cmd := &cobra.Command{
		Use:           filepath.Base(os.Args[0]),
		Short:         "UOR Client",
		Long:          clientLong,
		SilenceErrors: false,
		SilenceUsage:  false,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			logger, err := log.NewLogger(o.IOStreams.Out, o.LogLevel)
			if err != nil {
				return err
			}
			o.Logger = logger

			cacheEnv := os.Getenv("UOR_CACHE")
			if cacheEnv != "" {
				o.cacheDir = cacheEnv
			} else {
				home, err := homedir.Dir()
				if err != nil {
					return err
				}
				o.cacheDir = filepath.Join(home, ".uor", "cache")
			}

			return os.MkdirAll(o.cacheDir, 0750)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&o.LogLevel, "loglevel", "l", "info",
		"Log level (debug, info, warn, error, fatal)")

	cmd.AddCommand(NewRenderCmd(&o))
	cmd.AddCommand(NewInspectCmd(&o))
	cmd.AddCommand(NewBuildCmd(&o))
	cmd.AddCommand(NewPushCmd(&o))
	cmd.AddCommand(NewPullCmd(&o))
	cmd.AddCommand(NewVersionCmd(&o))

	return cmd
}
