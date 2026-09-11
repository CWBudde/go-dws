package wasm

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

// RequiredFileSystemMethods lists the methods a JavaScript object must provide
// to be usable as a custom filesystem. All of them must return synchronously.
var RequiredFileSystemMethods = []string{"readFile", "writeFile", "listDir", "delete", "exists"}

// missingFileSystemMethods returns the required method names for which isFunc
// reports false, in declaration order. The result is nil when the object
// satisfies the contract.
func missingFileSystemMethods(isFunc func(name string) bool) []string {
	var missing []string
	for _, name := range RequiredFileSystemMethods {
		if !isFunc(name) {
			missing = append(missing, name)
		}
	}
	return missing
}

// normalizeFSPath canonicalizes a script-supplied path into the absolute,
// slash-separated form used by the virtual filesystem and handed to the
// JavaScript host: backslashes become slashes, "." and ".." are resolved, and
// the result always starts with a single slash. The empty path maps to "/".
//
// Only separators and the empty string are special. Whitespace is part of a
// path, not decoration: POSIX and the browser storage APIs all accept names
// with leading or trailing spaces, so "/ reports " must keep addressing
// "/ reports " and not silently collapse onto "/reports".
func normalizeFSPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}

// errMissingFileSystemMethods reports a filesystem object that does not
// implement the full contract.
func errMissingFileSystemMethods(missing []string) error {
	return fmt.Errorf("custom filesystem is missing required method(s): %s (required: %s)",
		strings.Join(missing, ", "), strings.Join(RequiredFileSystemMethods, ", "))
}

// errFileSystemAccessor reports a filesystem object whose property accessor
// threw while a required method was being read, which happens with Proxy
// traps and getters. Validation treats it like any other malformed
// filesystem so the host still gets an ArgumentError.
func errFileSystemAccessor(name string, err error) error {
	return fmt.Errorf("reading required method %q of the custom filesystem failed: %w", name, err)
}

// errNotAnObject reports a non-object passed where a filesystem was expected.
func errNotAnObject(kind string) error {
	return fmt.Errorf("custom filesystem must be an object, got %s", kind)
}

// errAsyncFileSystem reports a filesystem method that returned a Promise.
// Async filesystems cannot be supported: platform.FileSystem is synchronous
// and awaiting a Promise from Go would deadlock the WASM event loop.
func errAsyncFileSystem(op string) error {
	return fmt.Errorf("custom filesystem method %q returned a Promise; "+
		"filesystem methods must be synchronous (pre-load async storage into memory)", op)
}

// fsOpError wraps a failure of a single filesystem operation with its operation
// name and path, mirroring the shape of *os.PathError messages.
func fsOpError(op, fsPath string, err error) error {
	return fmt.Errorf("%s %s: %w", op, fsPath, err)
}

// rawDirEntry is the platform-neutral form of one directory entry as reported
// by a JavaScript filesystem, before it is converted to platform.FileInfo.
type rawDirEntry struct {
	// Name is the entry name relative to the listed directory.
	Name string
	// ModTimeMillis is a JavaScript timestamp (milliseconds since the epoch).
	// It is only honored when HasModTime is true.
	ModTimeMillis float64
	// Size is the entry size in bytes; ignored for directories.
	Size int64
	// IsDir marks directory entries.
	IsDir bool
	// HasModTime reports whether the host supplied a modification time.
	HasModTime bool
}

// dirEntryToFileInfo converts one raw entry into a platform.FileInfo,
// rejecting entries without a usable name. Like normalizeFSPath, it strips
// separators but never whitespace: a file literally named " draft " keeps
// that name.
func dirEntryToFileInfo(e rawDirEntry) (platform.FileInfo, error) {
	name := strings.Trim(e.Name, "/")
	if name == "" {
		return platform.FileInfo{}, errors.New("directory entry has an empty name")
	}
	if strings.Contains(name, "/") {
		// Hosts sometimes return full paths; keep only the final element so
		// callers always see names relative to the listed directory.
		name = path.Base(name)
	}

	info := platform.FileInfo{
		Name:  name,
		Size:  e.Size,
		IsDir: e.IsDir,
	}
	if info.IsDir {
		info.Size = 0
	}
	if info.Size < 0 {
		return platform.FileInfo{}, fmt.Errorf("directory entry %q has a negative size (%d)", name, e.Size)
	}
	if e.HasModTime {
		info.ModTime = time.UnixMilli(int64(e.ModTimeMillis)).UTC()
	}
	return info, nil
}

// dirEntriesToFileInfos converts a whole listing, failing on the first
// malformed entry so hosts get an actionable error instead of silent gaps.
func dirEntriesToFileInfos(entries []rawDirEntry) ([]platform.FileInfo, error) {
	result := make([]platform.FileInfo, 0, len(entries))
	for i, e := range entries {
		info, err := dirEntryToFileInfo(e)
		if err != nil {
			return nil, fmt.Errorf("entry %d: %w", i, err)
		}
		result = append(result, info)
	}
	return result, nil
}
