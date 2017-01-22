// Implements node methods for files

package memfs

import (
	"bazil.org/fuse"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

// Attr sets attributes and permissions on the file
func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

// ReadAll data from the file.
func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}
