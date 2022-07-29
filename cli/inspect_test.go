package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/require"
	"github.com/uor-framework/uor-client-go/cli/log"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestInspectComplete(t *testing.T) {
	type spec struct {
		name     string
		args     []string
		opts     *InspectOptions
		expOpts  *InspectOptions
		expError string
	}

	cases := []spec{
		{
			name: "Valid/CorrectNumberOfArguments",
			args: []string{"test-registry.com/image:latest"},
			expOpts: &InspectOptions{
				Source: "test-registry.com/image:latest",
			},
			opts: &InspectOptions{},
		},
		{
			name:     "Invalid/NotEnoughArguments",
			args:     []string{},
			expOpts:  &InspectOptions{},
			opts:     &InspectOptions{},
			expError: "bug: expecting one argument",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.opts.Complete(c.args)
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expOpts, c.opts)
			}
		})
	}
}

func TestInspectRun(t *testing.T) {
	testlogr, err := log.NewLogger(ioutil.Discard, "debug")
	require.NoError(t, err)

	server := httptest.NewServer(registry.New())
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	type spec struct {
		name        string
		opts        *InspectOptions
		annotations map[string]string
		expRes      string
		expError    string
	}

	cases := []spec{
		{
			name: "Success/AttributesMatch",
			opts: &InspectOptions{
				RootOptions: &RootOptions{
					IOStreams: genericclioptions.IOStreams{
						Out:    os.Stdout,
						In:     os.Stdin,
						ErrOut: os.Stderr,
					},
					Logger: testlogr,
				},
				Source: fmt.Sprintf("%s/success:latest", u.Host),
				Attributes: map[string]string{
					"size": "small",
				},
			},
			annotations: map[string]string{
				"size": "small",
			},
			expRes: "Listing matching descriptors for source:  " + u.Host + "/success:latest\nName" +
				"                                      Digest" +
				"                                                                   Size  MediaType\nhello.txt" +
				"                                 sha256:03ba204e50d126e4674c005e04d82e84c21366780af1f43bd54a37816b6ab340" +
				"  13    application/vnd.oci.image.layer.v1.tar\n",
		},
		{
			name: "Success/NoAttributesMatch",
			opts: &InspectOptions{
				RootOptions: &RootOptions{
					IOStreams: genericclioptions.IOStreams{
						Out:    os.Stdout,
						In:     os.Stdin,
						ErrOut: os.Stderr,
					},
					Logger: testlogr,
				},
				Source: fmt.Sprintf("%s/success:latest", u.Host),
				Attributes: map[string]string{
					"size": "small",
				},
			},
			expRes: "Listing matching descriptors for source:  " + u.Host + "/success:latest\nName" +
				"                                      Digest  Size  MediaType\n",
		},
		{
			name: "Failure/NotStored",
			opts: &InspectOptions{
				RootOptions: &RootOptions{
					IOStreams: genericclioptions.IOStreams{
						Out:    os.Stdout,
						In:     os.Stdin,
						ErrOut: os.Stderr,
					},
					Logger:   testlogr,
					cacheDir: "testdata/cache",
				},
				Source: "localhost:5001/client-fake:latest",
			},
			expError: "descriptor for reference localhost:5001/client-fake:latest is not stored",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := filepath.Join(t.TempDir(), "cache")
			require.NoError(t, os.MkdirAll(cache, 0750))

			if c.opts.cacheDir == "" {
				c.opts.cacheDir = cache
				err := prepCache(c.opts.Source, cache, c.annotations)
				require.NoError(t, err)
			}

			out := new(strings.Builder)
			c.opts.IOStreams.Out = out

			err := c.opts.Run(context.TODO())
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expRes, out.String())
			}
		})
	}
}
