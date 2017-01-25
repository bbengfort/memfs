// Package memfs implements an in-memory replicated file system with Bazil
// FUSE and whose primary purpose is to implement all FUSE interfaces.
package memfs

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/bbengfort/sequence"
)

const (
	programName  = "memfs"
	majorVersion = 0
	minorVersion = 2
	microVersion = 0
	releaseLevel = "final"
	minBlockSize = uint64(512)
)

var (
	logger *Logger
)

func init() {
	logger, _ = InitLogger("", "DEBUG")
}

//===========================================================================
// New MemFS File System
//===========================================================================

// New MemFS file system created from a mount path and a configuration. This
// is the entry point for creating and launching all in-memory file systems.
func New(mount string, config *Config) *FileSystem {
	// Set the Log Level
	if config.Level != "" {
		logger, _ = InitLogger("", strings.ToUpper(config.Level))
	}

	// Create the file system
	fs := new(FileSystem)
	fs.MountPoint = mount
	fs.Config = config
	fs.Sequence, _ = sequence.New()

	// Set the UID and GID of the file system
	fs.uid = uint32(os.Geteuid())
	fs.gid = uint32(os.Getegid())

	// Set other system flags from the configuration
	fs.readonly = fs.Config.ReadOnly

	// Create the root directory
	fs.root = new(Dir)
	fs.root.Init("/", 0755, nil, fs)

	// Return the file system
	return fs
}

//===========================================================================
// File System Struct
//===========================================================================

// FileSystem implements the fuse.FS* interfaces as well as providing a
// lockable interaction structure to ensure concurrent accesses succeed.
type FileSystem struct {
	sync.Mutex                    // FileSystem can be locked and unlocked
	MountPoint string             // Path to the mount location on disk
	Config     *Config            // Configuration of the FileSystem
	Conn       *fuse.Conn         // Hook to the FUSE connection object
	Sequence   *sequence.Sequence // Monotonically increasing counter for inodes
	root       *Dir               // The root of the file system
	uid        uint32             // The user id of the process running the file system
	gid        uint32             // The group id of the process running the file system
	nfiles     uint64             // The number of files in the file system
	ndirs      uint64             // The number of directories in the file system
	nbytes     uint64             // The amount of data in the file system
	readonly   bool               // If the file system is readonly or not
}

// Run the FileSystem, mounting the MountPoint and connecting to FUSE
func (mfs *FileSystem) Run() error {
	var err error

	// Unmount the FS in case it was mounted with errors.
	fuse.Unmount(mfs.MountPoint)

	// Create the mount options to pass to Mount.
	opts := []fuse.MountOption{
		fuse.VolumeName("MemFS"),
		fuse.FSName("memfs"),
		fuse.Subtype("memfs"),
	}

	// If we're in readonly mode - pass to the mount options
	if mfs.readonly {
		opts = append(opts, fuse.ReadOnly())
	}

	// Mount the FS with the specified options
	if mfs.Conn, err = fuse.Mount(mfs.MountPoint, opts...); err != nil {
		return err
	}

	// Ensure that the file system is shutdown
	defer mfs.Conn.Close()
	logger.Info("mounted memfs:// on %s", mfs.MountPoint)

	// Serve the file system
	if err = fs.Serve(mfs.Conn, mfs); err != nil {
		return err
	}

	logger.Info("post serve")

	// Check if the mount process has an error to report
	<-mfs.Conn.Ready
	if mfs.Conn.MountError != nil {
		return mfs.Conn.MountError
	}

	return nil
}

// Shutdown the FileSystem unmounting the MountPoint and disconnecting FUSE.
func (mfs *FileSystem) Shutdown() error {
	logger.Info("shutting the file system down gracefully")

	if mfs.Conn == nil {
		return nil
	}

	if err := fuse.Unmount(mfs.MountPoint); err != nil {
		return err
	}

	return nil
}

//===========================================================================
// Implement fuse.FS* Methods
//===========================================================================

// Root is called to obtain the Node for the file system root. Implements the
// fuse.FS interface required of all file systems.
func (mfs *FileSystem) Root() (fs.Node, error) {
	return mfs.root, nil
}

// Destroy is called when the file system is shutting down. Implements the
// fuse.FSDestroyer interface.
//
// Linux only sends this request for block device backed (fuseblk)
// filesystems, to allow them to flush writes to disk before the
// unmount completes.
func (mfs *FileSystem) Destroy() {
	logger.Info("file system is being destroyed")
}

// GenerateInode is called to pick a dynamic inode number when it
// would otherwise be 0. Implements the fuse.FSInodeGenerator interface.
//
// Not all filesystems bother tracking inodes, but FUSE requires
// the inode to be set, and fewer duplicates in general makes UNIX
// tools work better.
//
// Operations where the nodes may return 0 inodes include Getattr,
// Setattr and ReadDir.
//
// If FS does not implement FSInodeGenerator, GenerateDynamicInode
// is used.
//
// Implementing this is useful to e.g. constrain the range of
// inode values used for dynamic inodes.
func (mfs *FileSystem) GenerateInode(parentInode uint64, name string) uint64 {
	return fs.GenerateDynamicInode(parentInode, name)
}

// Statfs is called to obtain file system metadata. Implements fuse.FSStatfser
// by writing the metadata to the resp.
func (mfs *FileSystem) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
	logger.Debug("statfs called on file system")

	// Compute the total number of available blocks
	resp.Blocks = mfs.Config.CacheSize / minBlockSize

	// Compute the number of used blocks
	numblocks := mfs.nbytes / minBlockSize
	if mfs.nbytes%minBlockSize > 0 {
		numblocks++
	}

	// Report the number of free and available blocks for the block size
	resp.Bfree = resp.Blocks - numblocks
	resp.Bavail = resp.Blocks - numblocks
	resp.Bsize = uint32(minBlockSize)

	// Report the total number of files in the file system (and those free)
	resp.Files = mfs.nfiles
	resp.Ffree = 0

	// Report the maximum length of a name and the minimum fragment size
	resp.Namelen = 2048
	resp.Frsize = uint32(minBlockSize)

	return nil
}

//===========================================================================
// Version and Package Information
//===========================================================================

// PackageVersion composes version information from the constants in this file
// and returns a string that defines current information about the package.
func PackageVersion() string {
	vstr := fmt.Sprintf("%d.%d", majorVersion, minorVersion)

	if microVersion > 0 {
		vstr += fmt.Sprintf(".%d", microVersion)
	}

	switch releaseLevel {
	case "final":
		return vstr
	case "alpha":
		return vstr + "a"
	case "beta":
		return vstr + "b"
	default:
		return vstr
	}

}
