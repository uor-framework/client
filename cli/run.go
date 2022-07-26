package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
	"os"
)

// RunOptions describe configuration options that can
// be set using the run subcommand.
type RunOptions struct {
	*RootOptions
	Config string
}

var clientRunExamples = templates.Examples(
	`
	# Push artifacts
	client run ./config.yaml
	`,
)

// NewRunCmd creates a new cobra.Command for the run subcommand.
func NewRunCmd(rootOpts *RootOptions) *cobra.Command {
	o := RunOptions{RootOptions: rootOpts}

	cmd := &cobra.Command{
		Hidden:        !o.UOR_DEV_MODE, // TODO delete after implementation
		Use:           "run <CONFIG>",
		Short:         "Run instructions against a UOR collection",
		Example:       clientRunExamples,
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Complete(args))
			cobra.CheckErr(o.Validate())
			cobra.CheckErr(o.Run(cmd.Context()))
		},
	}

	return cmd
}

func (o *RunOptions) Complete(args []string) error {
	if len(args) < 1 {
		return errors.New("bug: expecting 1 argument")
	}
	o.Config = args[0]
	return nil
}

func (o *RunOptions) Validate() error {
	configInfo, err := os.Stat(o.Config)
	if err != nil {
		return fmt.Errorf("unable to read config file: %v", err)
	}
	if !configInfo.Mode().IsRegular() {
		return errors.New("config file must be a regular file")
	}
	return nil
}

func (o *RunOptions) Run(ctx context.Context) error {
	o.Logger.Infof("stub: command run not yet implemented\n")
	return nil
}
