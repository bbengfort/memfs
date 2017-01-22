// Package memfs implements an in-memory replicated file system with Bazil
// FUSE and anti-entropy replication with gRPC communication.
package memfs

import "fmt"

const (
	programName  = "memfs"
	majorVersion = 0
	minorVersion = 1
	microVersion = 0
	releaseLevel = "final"
)

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
