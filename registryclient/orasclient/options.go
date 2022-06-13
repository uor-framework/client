package orasclient

import (
	"github.com/uor-framework/client/registryclient"
	"oras.land/oras-go/pkg/content"
)

// TODO(jpower432): Allow configuration for relevant ORAS copy options

// ClientOption is a function that configures
// options on the client config.
type ClientOption func(o *ClientConfig) error

// ClientConfig contains configuration data for the registry client.
type ClientConfig struct {
	output    string
	configs   []string
	plainHTTP bool
	insecure  bool
}

func (c *ClientConfig) apply(options []ClientOption) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return err
		}
	}
	return nil
}

// NewClient returns a new ORAS client implementation
func NewClient(options ...ClientOption) (registryclient.Client, error) {
	client := &orasClient{
		fileStore: content.NewFile(""),
	}

	config := &ClientConfig{}
	if err := config.apply(options); err != nil {
		return client, err
	}

	client.registryOpts.Insecure = config.insecure
	client.registryOpts.Configs = config.configs
	client.registryOpts.PlainHTTP = config.plainHTTP
	client.outputDir = config.output
	return client, nil
}

// WithAuthConfigs adds configuration files
// with registry authorization information.
func WithAuthConfigs(configs []string) ClientOption {
	return func(config *ClientConfig) error {
		config.configs = configs
		return nil
	}
}

// SkipTLSVerify disables TLS certificate checking.
func SkipTLSVerify(insecure bool) ClientOption {
	return func(config *ClientConfig) error {
		config.insecure = insecure
		return nil
	}
}

// WithPlainHTTP uses the HTTP protocol with the registry.
func WithPlainHTTP(plainHTTP bool) ClientOption {
	return func(config *ClientConfig) error {
		config.plainHTTP = plainHTTP
		return nil
	}
}

// WithOutputDir will copy any pulled artifact to this directory
func WithOutputDir(dir string) ClientOption {
	return func(config *ClientConfig) error {
		config.output = dir
		return nil
	}
}
