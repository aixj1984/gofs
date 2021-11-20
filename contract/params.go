package contract

import (
	"strconv"
)

const (
	// FsDir is dir, see FsDirValue
	FsDir = "dir"
	// FsSize file size, bytes
	FsSize = "size"
	// FsHash file hash value
	FsHash = "hash"
	// FsCtime file create time
	FsCtime = "ctime"
	// FsAtime file last access time
	FsAtime = "atime"
	// FsMtime file last modify time
	FsMtime = "mtime"
	// FsPath file path
	FsPath = "path"
)

// FsDirValue the optional value of FsDir
type FsDirValue int

const (
	// FsIsDir current path is a dir
	FsIsDir FsDirValue = 1
	// FsNotDir current path is not a dir
	FsNotDir FsDirValue = 0
	// FsUnknown current path is unknown file type
	FsUnknown FsDirValue = -1
)

// String parse the current value to string
func (v FsDirValue) String() string {
	return strconv.Itoa(int(v))
}

// Is is current value equal to target
func (v FsDirValue) Is(t string) bool {
	return v.String() == t
}

// Not is current value not equal to target
func (v FsDirValue) Not(t string) bool {
	return v.String() != t
}
