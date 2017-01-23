// Node contains shared information for both files and directories.

package memfs

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/context"

	"bazil.org/fuse"
)

//===========================================================================
// Node Types and Constructor
//===========================================================================

// XAttr is a mapping of names to binary data for file systems that support
// extended attributes or other data.
type XAttr map[string][]byte

// Entity represents a memfs.Node entity (to differentiate it from a fs.Node)
type Entity interface {
	IsDir() bool               // Returns true if the entity is a directory
	IsArchive() bool           // Returns true if the entity is an archive (version history)
	FuseType() fuse.DirentType // Returns the fuse type for listing
	Path() string              // Returns the full path to the entity
	GetNode() *Node            // Returns the node for the entity type
}

// Node contains shared data and structures for both files and directories.
type Node struct {
	ID     uint64      // Unique ID of the Node
	Name   string      // Name of the Node
	Attrs  fuse.Attr   // Node attributes and permissions
	XAttrs XAttr       // Extended attributes on the node
	Parent *Dir        // Parent directory of the Node
	fs     *FileSystem // Stored reference to the file system
}

// Init a Node with the required properties for storage in the file system.
func (n *Node) Init(name string, mode os.FileMode, parent *Dir, fs *FileSystem) {
	// Manage the Node properties
	n.ID, _ = fs.Sequence.Next()
	n.Name = name
	n.Parent = parent
	n.XAttrs = make(XAttr)
	n.fs = fs

	// Manage the fuse.Attr properties
	now := time.Now()
	n.Attrs.Inode = n.ID // inode number
	n.Attrs.Size = 0     // size in bytes
	n.Attrs.Blocks = 0   // size in 512-byte units
	n.Attrs.Atime = now  // time of last access
	n.Attrs.Mtime = now  // time of last modification
	n.Attrs.Ctime = now  // time of last inode change
	n.Attrs.Crtime = now // time of creation (OS X only)
	n.Attrs.Mode = mode  // file mode
	n.Attrs.Nlink = 1    // number of links (usually 1)
	n.Attrs.Uid = fs.uid // owner uid
	n.Attrs.Gid = fs.gid // group gid
	// n.Attrs.Rdev = 0      // device numbers
	// n.Attrs.Flags = 0     // chflags(2) flags (OS X only)
	// n.Attrs.BlockSize = 0 // preferred blocksize for filesystem I/O

	logger.Info("initialized node %d, %q", n.ID, n.Name)
}

//===========================================================================
// Node Methods
//===========================================================================

// IsDir returns if the node is a directory by inspecting the file mode.
func (n *Node) IsDir() bool {
	return (n.Attrs.Mode & os.ModeDir) != 0
}

// IsArchive returns true if the node is an archive node, that is a node
// constructed to display version history (and is therefore not writeable).
// TODO: Implement archives
func (n *Node) IsArchive() bool {
	return false
}

// FuseType returns the fuse type of the node for listing
func (n *Node) FuseType() fuse.DirentType {
	if n.IsDir() {
		return fuse.DT_Dir
	}

	return fuse.DT_File
}

// Path returns a string representation of the full path of the node.
func (n *Node) Path() string {
	if n.Parent != nil {
		return filepath.Join(n.Parent.Path(), n.Name)
	}
	return n.Name
}

// GetNode returns a pointer to the embedded Node object
func (n *Node) GetNode() *Node {
	return n
}

// String returns the full path to the node.
func (n *Node) String() string {
	return n.Path()
}

//===========================================================================
// Node Interface
//===========================================================================

// Access checks whether the calling context has permission for
// the given operations on the receiver. If so, Access should
// return nil. If not, Access should return EPERM.
//
// Note that this call affects the result of the access(2) system
// call but not the open(2) system call. If Access is not
// implemented, the Node behaves as if it always returns nil
// (permission granted), relying on checks in Open instead.
//
// https://godoc.org/bazil.org/fuse/fs#NodeAccesser
func (n *Node) Access(ctx context.Context, req *fuse.AccessRequest) error {
	logger.Debug("access called on node %d", n.ID)
	return nil // Permission always granted, relying on checks in Open.
}

// Attr fills attr with the standard metadata for the node.
//
// Fields with reasonable defaults are prepopulated. For example,
// all times are set to a fixed moment when the program started.
//
// If Inode is left as 0, a dynamic inode number is chosen.
//
// The result may be cached for the duration set in Valid.
//
// https://godoc.org/bazil.org/fuse/fs#Node
func (n *Node) Attr(ctx context.Context, attr *fuse.Attr) error {
	logger.Debug("attr called on node %d", n.ID)
	attr.Inode = n.Attrs.Inode         // inode number
	attr.Size = n.Attrs.Size           // size in bytes
	attr.Blocks = n.Attrs.Blocks       // size in 512-byte units
	attr.Atime = n.Attrs.Atime         // time of last access
	attr.Mtime = n.Attrs.Mtime         // time of last modification
	attr.Ctime = n.Attrs.Ctime         // time of last inode change
	attr.Crtime = n.Attrs.Crtime       // time of creation (OS X only)
	attr.Mode = n.Attrs.Mode           // file mode
	attr.Nlink = n.Attrs.Nlink         // number of links (usually 1)
	attr.Uid = n.Attrs.Uid             // owner uid
	attr.Gid = n.Attrs.Gid             // group gid
	attr.Rdev = n.Attrs.Rdev           // device numbers
	attr.Flags = n.Attrs.Flags         // chflags(2) flags (OS X only)
	attr.BlockSize = n.Attrs.BlockSize // preferred blocksize for filesystem I/O
	return nil
}

// Forget about this node. This node will not receive further method calls.
//
// Forget is not necessarily seen on unmount, as all nodes are implicitly
// forgotten as part part of the unmount.
//
// https://godoc.org/bazil.org/fuse/fs#NodeForgetter
//
// Currently forget does nothing except log that it was forgotten.
func (n *Node) Forget() {
	logger.Debug("forget node %d", n.ID)
}

// Getattr obtains the standard metadata for the receiver.
// It should store that metadata in resp.
//
// If this method is not implemented, the attributes will be
// generated based on Attr(), with zero values filled in.
//
// https://godoc.org/bazil.org/fuse/fs#NodeGetattrer
func (n *Node) Getattr(ctx context.Context, req *fuse.GetattrRequest, resp *fuse.GetattrResponse) error {
	logger.Debug("getting attrs on node %d", n.ID)
	resp.Attr = n.Attrs
	return nil
}

// Getxattr gets an extended attribute by the given name from the node.
//
// If there is no xattr by that name, returns fuse.ErrNoXattr.
//
// https://godoc.org/bazil.org/fuse/fs#NodeGetxattrer
func (n *Node) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	if data, ok := n.XAttrs[req.Name]; ok {
		logger.Debug("getting xattr named %s on node %d", req.Name, n.ID)
		if req.Size != 0 {
			resp.Xattr = data[:req.Size]
		} else {
			resp.Xattr = data
		}

		return nil
	}

	logger.Debug("(error) no xattr named %s on node %d", req.Name, n.ID)
	return fuse.ErrNoXattr
}

// Listxattr lists the extended attributes recorded for the node.
//
// https://godoc.org/bazil.org/fuse/fs#NodeListxattrer
func (n *Node) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	logger.Debug("listing xattr names on node %d", n.ID)

	for name := range n.XAttrs {
		resp.Append(name)
	}

	return nil
}

// Removexattr removes an extended attribute for the name.
//
// If there is no xattr by that name, returns fuse.ErrNoXattr.
//
// https://godoc.org/bazil.org/fuse/fs#NodeRemovexattrer
func (n *Node) Removexattr(ctx context.Context, req *fuse.RemovexattrRequest) error {
	if n.IsArchive() || n.fs.readonly {
		return fuse.EPERM
	}

	n.fs.Lock()
	defer n.fs.Unlock()

	if _, ok := n.XAttrs[req.Name]; ok {
		logger.Debug("removing xattr named %s on node %d", req.Name, n.ID)
		delete(n.XAttrs, req.Name)
		return nil
	}

	logger.Debug("(error) could not remove xattr named %s on node %d", req.Name, n.ID)
	return fuse.ErrNoXattr
}

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
func (n *Node) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if n.IsArchive() || n.fs.readonly {
		return fuse.EPERM
	}

	n.fs.Lock()
	defer n.fs.Unlock()

	// If a handle is set - we don't do anything with that currently.
	if req.Valid.Handle() {
		logger.Debug("(error) setting handle attr on node %d but we don't store it!", n.ID)
	}

	// If size is set, this represents a truncation for a file (for a dir?)
	if req.Valid.Size() {
		if n.IsDir() {
			// NOTE: File objects implement the actual truncation.
			logger.Debug("(error) attempting to truncate directory %d", n.ID)
		}
	}

	// Set the access time on the node
	if req.Valid.Atime() {
		logger.Debug("setting node %d Atime to %v", n.ID, req.Atime)
		n.Attrs.Atime = req.Atime
	}

	// Linux only: set the access time to now
	if req.Valid.AtimeNow() {
		logger.Debug("setting node %d Atime to now", n.ID)
		n.Attrs.Atime = time.Now()
	}

	// Set the modify time on the node
	if req.Valid.Mtime() {
		logger.Debug("setting node %d Mtime to %v", n.ID, req.Mtime)
		n.Attrs.Mtime = req.Mtime
	}

	// Linux only: set the modified time to now
	if req.Valid.MtimeNow() {
		logger.Debug("setting node %d Mtime to now", n.ID)
		n.Attrs.Mtime = time.Now()
	}

	// Set the mode on the node
	if req.Valid.Mode() {
		logger.Debug("setting node %d Mode to %v", n.ID, req.Mode)
		n.Attrs.Mode = req.Mode
	}

	// Set the uid on the node
	if req.Valid.Uid() {
		logger.Debug("setting node %d UID to %v", n.ID, req.Uid)
		n.Attrs.Uid = req.Uid
	}

	// Set the gid on the node
	if req.Valid.Gid() {
		logger.Debug("setting node %d GID to %v", n.ID, req.Gid)
		n.Attrs.Gid = req.Gid
	}

	// Linux only: set the lock owner flag - not implemented
	if req.Valid.LockOwner() {
		logger.Debug("(error) setting lock owner on node %d but we don't implement it!", n.ID)
	}

	// OS X only: set the bkuptime on the node
	if req.Valid.Bkuptime() {
		logger.Debug("(error) setting bkuptime on node %d to %v but we don't store it!", n.ID, req.Bkuptime)
	}

	// OS X only: set the chgtime on the node
	if req.Valid.Chgtime() {
		logger.Debug("(error) setting chgtime on node %d to %v but we don't store it!", n.ID, req.Chgtime)
	}

	// OS X only: set the crtime on the node
	if req.Valid.Crtime() {
		logger.Debug("setting node %d Crtime to %v", n.ID, req.Crtime)
		n.Attrs.Crtime = req.Crtime
	}

	// OS X only: set the flags on the node
	if req.Valid.Flags() {
		logger.Debug("setting node %d flags to %v", n.ID, req.Flags)
		n.Attrs.Flags = req.Flags
	}

	resp.Attr = n.Attrs
	return nil
}

// Setxattr sets an extended attribute with the given name and value.
// TODO: Use flags to fail the request if the xattr does/not already exist.
//
// https://godoc.org/bazil.org/fuse/fs#NodeSetxattrer
func (n *Node) Setxattr(ctx context.Context, req *fuse.SetxattrRequest) error {
	if n.IsArchive() || n.fs.readonly {
		return fuse.EPERM
	}

	n.fs.Lock()
	defer n.fs.Unlock()

	logger.Debug("setting xattr named %s on node %d", req.Name, n.ID)
	n.XAttrs[req.Name] = req.Xattr
	return nil
}
