package layout

import (
	"context"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	cacheDir := "testdata/valid"
	l, err := New(cacheDir)
	require.NoError(t, err)
	type spec struct {
		name     string
		desc     ocispec.Descriptor
		expRes   bool
		expError string
	}

	cases := []spec{
		{
			name: "Success/DescExists",
			desc: ocispec.Descriptor{
				MediaType: "application/vnd.uor.config.v1+json",
				Size:      2,
				Digest:    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			},
			expRes: true,
		},
		{
			name: "Success/DescDoesNotExist",
			desc: ocispec.Descriptor{
				MediaType: "application/vnd.uor.config.v1+json",
				Size:      3,
				Digest:    "sha256:44136fa356b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			},
			expRes: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := l.Exists(context.TODO(), c.desc)
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expRes, res)
			}
		})
	}
}

func TestTag(t *testing.T) {
	cacheDir := t.TempDir()
	source := "testdata/valid"
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Do not copy in the index file. We are generating a new one for this test.
		if info.Name() == indexFile {
			return nil
		}
		relPath := strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		switch m := info.Mode(); {
		case m&fs.ModeSymlink != 0:
			dst, err := os.Readlink(path)
			if err != nil {
				return err
			}
			id := filepath.Base(dst)
			if err := os.Symlink(id, filepath.Join(cacheDir, relPath)); err != nil {
				return err
			}
		case m.IsDir():
			return os.Mkdir(filepath.Join(cacheDir, relPath), 0750)
		default:
			newSource := filepath.Join(source, relPath)
			cleanSource := filepath.Clean(newSource)
			data, err := ioutil.ReadFile(cleanSource)
			if err != nil {
				return err
			}
			newDest := filepath.Join(cacheDir, relPath)
			cleanDest := filepath.Clean(newDest)
			return ioutil.WriteFile(cleanDest, data, 0600)
		}
		return nil
	})
	require.NoError(t, err)

	l, err := New(cacheDir)
	require.NoError(t, err)

	desc := ocispec.Descriptor{Digest: "sha256:2e30f6131ce2164ed5ef017845130727291417d60a1be6fad669bdc4473289cd"}

	require.NoError(t, l.Tag(context.TODO(), desc, "test/test:tag"))

	require.Error(t, l.Tag(context.TODO(), ocispec.Descriptor{}, "test"))
	require.Error(t, l.Tag(context.TODO(), ocispec.Descriptor{}, "test/repo@sha256:2e30f6131"))

	_, err = os.Stat(filepath.Join(cacheDir, indexFile))
	require.NoError(t, err)

	ii, err := l.Index()
	require.NoError(t, err)
	require.Len(t, ii.Manifests, 1)
	desc = ii.Manifests[0]
	refName := desc.Annotations[ocispec.AnnotationRefName]

	require.Equal(t, "test/test:tag", refName)
}

func TestSaveIndex(t *testing.T) {
	cacheDir := t.TempDir()

	l, err := New(cacheDir)
	require.NoError(t, err)

	l.resolver.Store("test", ocispec.Descriptor{})
	require.NoError(t, l.SaveIndex())

	_, err = os.Stat(filepath.Join(cacheDir, indexFile))
	require.NoError(t, err)

	ii, err := l.Index()
	require.NoError(t, err)
	require.Len(t, ii.Manifests, 1)
	desc := ii.Manifests[0]
	refName := desc.Annotations[ocispec.AnnotationRefName]

	require.Equal(t, "test", refName)
}

func TestResolve(t *testing.T) {
	cacheDir := "testdata/valid"
	l, err := New(cacheDir)
	require.NoError(t, err)

	type spec struct {
		name     string
		ref      string
		expDesc  ocispec.Descriptor
		expError string
	}

	cases := []spec{
		{
			name: "Success/RefExists",
			ref:  "localhost:5001/test:latest",
			expDesc: ocispec.Descriptor{
				MediaType:   "application/vnd.oci.image.manifest.v1+json",
				Size:        1013,
				Digest:      "sha256:473f7d69dbc51105aff4bb2f7ec80e27402d2f40c3e9a076e8c773b15969eadf",
				Annotations: map[string]string{"org.opencontainers.image.ref.name": "localhost:5001/test:latest"},
			},
		},
		{
			name:     "Failure/RefDoesNotExist",
			ref:      "localhost:5001/notexists:latest",
			expError: "descriptor for reference localhost:5001/notexists:latest is not stored",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			desc, err := l.Resolve(context.TODO(), c.ref)
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expDesc, desc)
			}
		})
	}
}

func TestLoadIndex(t *testing.T) {
	cacheDir := "testdata/valid"
	ctx := context.TODO()
	l, err := New(cacheDir)
	require.NoError(t, err)

	require.NoError(t, l.loadIndex(ctx))

	ii, err := l.Index()
	require.NoError(t, err)

	require.Len(t, ii.Manifests, 1)
	require.Equal(t, "sha256:473f7d69dbc51105aff4bb2f7ec80e27402d2f40c3e9a076e8c773b15969eadf", ii.Manifests[0].Digest.String())
}

func TestValidateOCILayoutFile(t *testing.T) {
	type spec struct {
		name     string
		cacheDir string
		expError string
	}

	cases := []spec{
		{
			name:     "Success/OCILayoutFileWithCorrectVersion",
			cacheDir: "testdata/valid",
		},
		{
			name:     "Failure/WrongVersion",
			cacheDir: "testdata/invalid",
			expError: "unsupported version",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := &Layout{rootPath: c.cacheDir}
			err := l.validateOCILayoutFile()
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPredecessors(t *testing.T) {
	cacheDir := "testdata/valid"
	expected := []ocispec.Descriptor{{
		MediaType:   "application/vnd.oci.image.manifest.v1+json",
		Digest:      "sha256:473f7d69dbc51105aff4bb2f7ec80e27402d2f40c3e9a076e8c773b15969eadf",
		Size:        1013,
		Annotations: map[string]string{"org.opencontainers.image.ref.name": "localhost:5001/test:latest"},
	}}
	ctx := context.TODO()
	l, err := New(cacheDir)
	require.NoError(t, err)

	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    "sha256:5c29ebcf4a3e7ac6dca6dcea98b4fa98de57c4aca65fa0b49989fbeab1dfdf84",
		Size:      32,
	}
	pre, err := l.Predecessors(ctx, desc)
	require.NoError(t, err)
	require.Equal(t, expected, pre)

}

func TestResolveLinks(t *testing.T) {
	type spec struct {
		name     string
		cacheDir string
		ref      string
		expRes   []string
		expError string
	}

	cases := []spec{
		{
			name:     "Success/OCILayoutFileWithCorrectVersion",
			cacheDir: "testdata/attributes",
			ref:      "localhost:5001/test3:latest",
			expRes:   []string{"localhost:5001/test1:latest"},
		},
		{
			name:     "Failure/NoCollectionLinks",
			cacheDir: "testdata/valid",
			ref:      "localhost:5001/test:latest",
			expError: "no collection links",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.TODO()
			l, err := NewWithContext(ctx, c.cacheDir)
			require.NoError(t, err)
			res, err := l.ResolveLinks(ctx, c.ref)
			if c.expError != "" {
				require.EqualError(t, err, c.expError)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expRes, res)
			}
		})
	}
}
