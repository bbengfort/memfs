// Implements Node and Handler methods for files

package memfs

import (
	"os"
	"time"

	"bazil.org/fuse"
	"golang.org/x/net/context"
)

//===========================================================================
// Dir Type and Constructor
//===========================================================================

// File implements Node and Handler interfaces for file (data containing)
// objects in MemFs. Data is allocated directly in the file object, and is
// not chunked or broken up until transport.
type File struct {
	Node
	Data []byte // Actual data contained by the File
}

// Init the file and create the data array
func (f *File) Init(name string, mode os.FileMode, parent *Dir, memfs *FileSystem) {
	// Init the embedded node.
	f.Node.Init(name, mode, parent, memfs)

	// Make the data array
	f.Data = make([]byte, 0, 0)
}

//===========================================================================
// File Methods
//===========================================================================

// GetNode returns a pointer to the embedded Node object
func (f *File) GetNode() *Node {
	return &f.Node
}

//===========================================================================
// File fuse.Node* Interface
//===========================================================================

// Setattr sets the standard metadata for the receiver.
//
// Note, this is also used to communicate changes in the size of
// the file, outside of Writes.
//
// req.Valid is a bitmask of what fields are actually being set.
// For example, the method should not change the mode of the file
// unless req.Valid.Mode() is true.
//
// https://godoc.org/bazil.org/fuse/fs#NodeSetattrer
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	// If size is set, this represents a truncation for a file (for a dir?)
	if req.Valid.Size() {
		f.fs.Lock() // Only lock if we're going to change the size.

		f.Attrs.Size = req.Size
		f.Data = f.Data[:req.Size]
		logger.Debug("truncate size from %d to %d on file %d", f.Attrs.Size, req.Size, f.ID)

		f.fs.Unlock() // Must unlock before Node.Setattr is called!
	}

	// Now use the embedded Node's Setattr method.
	return f.Node.Setattr(ctx, req, resp)
}

// Fsync must be defined or edting with vim or emacs fails.
// Implements NodeFsyncer, which has no associated documentation.
//
// https://godoc.org/bazil.org/fuse/fs#NodeFsyncer
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	logger.Debug("fsync on file %d", f.ID)
	return nil
}

//===========================================================================
// File fuse.Handle* Interface
//===========================================================================

// Flush is called each time the file or directory is closed. Because there
// can be multiple file descriptors referring to a single opened file, Flush
// can be called multiple times.
//
// Because this is an in-memory system, Flush is basically ignored.
//
// https://godoc.org/bazil.org/fuse/fs#HandleFlusher
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	logger.Debug("flush file %d", f.ID)
	return nil
}

// ReadAll the data from a file. Implements HandleReadAller which has no
// associated documentation.
//
// https://godoc.org/bazil.org/fuse/fs#HandleReadAller
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	f.fs.Lock()
	defer f.fs.Unlock()

	// Set the access time on the file.
	f.Attrs.Atime = time.Now()

	// Return the data with no error.
	logger.Debug("read all file %d", f.ID)
	return f.Data, nil
}

// Read requests to read data from the handle.
//
// There is a page cache in the kernel that normally submits only page-aligned
// reads spanning one or more pages. However, you should not rely on this. To
// see individual requests as submitted by the file system clients, set
// OpenDirectIO.
//
// NOTE: that reads beyond the size of the file as reported by Attr are not
// even attempted (except in OpenDirectIO mode).
//
// https://godoc.org/bazil.org/fuse/fs#HandleReader
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.fs.Lock()
	defer f.fs.Unlock()

	// Find the end of the data slice to return.
	to := uint64(req.Offset) + uint64(req.Size)
	if to > f.Attrs.Size {
		to = f.Attrs.Size
	}

	// Set the access time on the file.
	f.Attrs.Atime = time.Now()

	// Set the data on the response object.
	resp.Data = f.Data[req.Offset:to]

	logger.Debug("read %d bytes from offset %d in file %d", req.Size, req.Offset, f.ID)
	return nil
}

// Release the handle to the file. No associated documentation.
//
// https://godoc.org/bazil.org/fuse/fs#HandleReleaser
func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	logger.Debug("release handle on file %d", f.ID)
	return nil
}

// Write requests to write data into the handle at the given offset.
// Store the amount of data written in resp.Size.
//
// There is a writeback page cache in the kernel that normally submits only
// page-aligned writes spanning one or more pages. However, you should not
// rely on this. To see individual requests as submitted by the file system
// clients, set OpenDirectIO.
//
// Writes that grow the file are expected to update the file size (as seen
// through Attr). Note that file size changes are communicated also through
// Setattr.
//
// https://godoc.org/bazil.org/fuse/fs#HandleWriter
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	f.fs.Lock()
	defer f.fs.Unlock()

	olen := uint64(len(f.Data))   // original data length
	wlen := uint64(len(req.Data)) // data write length
	off := uint64(req.Offset)     // offset of the write
	lim := off + wlen             // The final length of the data

	// Ensure the original size is the same as the set size (debugging)
	if olen != f.Attrs.Size {
		msg := "bad size match: %d vs %d"
		logger.Error(msg, olen, f.Attrs.Size)
	}

	// If the amount of data being written is greater than the amount of data
	// currently being stored, allocate a new array with sufficient size and
	// copy the original data to that buffer.
	if lim > olen {
		buf := make([]byte, lim)

		var to uint64
		if off < olen {
			to = off
		} else {
			to = olen
		}

		copy(buf[0:to], f.Data[0:to])
		f.Data = buf

		// Update the attrs on the file
		f.Attrs.Size = lim

		// Update the file system state
		f.fs.nbytes += lim - olen
	}

	// Copy the data from the request into our data buffer
	copy(f.Data[off:lim], req.Data[:])

	// Set the attributes on the file
	// TODO: What if the size of the data (lim) <= olen? Should we truncate?
	f.Attrs.Mtime = time.Now()

	// Set the attributes on the response
	resp.Size = int(wlen)

	logger.Debug("wrote %d bytes offset by %d to file %d", wlen, off, f.ID)
	return nil
}
