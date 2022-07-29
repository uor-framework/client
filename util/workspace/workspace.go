package workspace

import (
	"context"
	"io"
	"path/filepath"
)

// Workspace defines methods for accessing and publishing
// files in a local context.
type Workspace interface {
	// ReadObject reads the provided object from disk.
	ReadObject(context.Context, string, interface{}) error
	// WriteObject writes the provided object to disk.
	WriteObject(context.Context, string, interface{}) error
	// GetWriter returns an os.File as a writer.
	GetWriter(context.Context, string) (io.Writer, error)
	// Walk will traverse the workspace directory.
	Walk(filepath.WalkFunc) error
	// NewDirectory creates a new workspace under the current workspace.
	NewDirectory(string) (Workspace, error)
	// DeleteDirectory will delete a directory under a workspace.
	DeleteDirectory(string) error
	// Path generates a path of a file with the workspace directory.
	Path(...string) string
}
