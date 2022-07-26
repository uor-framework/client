package cli

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/uor-framework/uor-client-go/cli/log"
)

// EnvConfig stores CLI runtime configuration from environment variables.
// Struct field names should match the name of the environment variable that the field is derived from.
type EnvConfig struct {
	UOR_DEV_MODE bool // true: show unimplemented stubs in --help
}

// RootOptions describe global configuration options that can be set.
type RootOptions struct {
	IOStreams genericclioptions.IOStreams
	LogLevel  string
	Logger    log.Logger
	EnvConfig
}

var clientLong = templates.LongDesc(
	`
	This client helps you build and publish UOR collections as an OCI artifact.
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
	o.EnvConfig = readEnvConfig()
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
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	f := cmd.PersistentFlags()
	f.StringVarP(&o.LogLevel, "loglevel", "l", "info",
		"Log level (debug, info, warn, error, fatal)")

	// TODO(sabre1041) Reenable/remove once build capability strategy determined
	//cmd.AddCommand(NewBuildCmd(&o))
	cmd.AddCommand(NewPushCmd(&o))
	cmd.AddCommand(NewPullCmd(&o))
	cmd.AddCommand(NewRunCmd(&o))
	cmd.AddCommand(NewVersionCmd(&o))

	return cmd
}

func readEnvConfig() EnvConfig {
	envConfig := EnvConfig{}

	devModeString := os.Getenv("UOR_DEV_MODE")
	devMode, err := strconv.ParseBool(devModeString)
	envConfig.UOR_DEV_MODE = err == nil && devMode

	return envConfig
}
