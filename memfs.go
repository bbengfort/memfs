// Package memfs implements an in-memory replicated file system with Bazil
// FUSE and anti-entropy replication with gRPC communication.
package memfs

import (
	"fmt"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/bbengfort/sequence"
)

const (
	programName  = "memfs"
	majorVersion = 0
	minorVersion = 1
	microVersion = 0
	releaseLevel = "final"
)

var (
	logger *Logger
)

func init() {
	logger, _ = InitLogger("", "INFO")
}

//===========================================================================
// New MemFS File System
//===========================================================================

// New MemFS file system created from a mount path and a configuration. This
// is the entry point for creating and launching all in-memory file systems.
func New(mount string, config *Config) *FileSystem {
	fs := new(FileSystem)
	fs.MountPoint = mount
	fs.Config = config
	fs.Sequence, _ = sequence.New()

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
}

// Run the FileSystem, mounting the MountPoint and connecting to FUSE
func (mfs *FileSystem) Run() error {
	var err error

	// Unmount the FS in case it was mounted with errors.
	fuse.Unmount(mfs.MountPoint)

	// Mount the FS with the specified options
	if mfs.Conn, err = fuse.Mount(
		mfs.MountPoint,
	); err != nil {
		return err
	}

	// Ensure that the file system is shutdown
	defer mfs.Conn.Close()
	logger.Info("mounted memfs:// on %s", mfs.MountPoint)

	// Serve the file system
	if err = fs.Serve(mfs.Conn, mfs); err != nil {
		return err
	}

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

	// if err := mfs.Conn.Close(); err != nil {
	// 	return err
	// }

	if err := fuse.Unmount(mfs.MountPoint); err != nil {
		return err
	}

	return nil
}

//===========================================================================
// Implement fuse.FS* Methods
//===========================================================================

// Root returns the root directory
func (mfs *FileSystem) Root() (fs.Node, error) {
	return Dir{}, nil
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
