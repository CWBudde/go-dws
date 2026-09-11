package cmd

import (
	"path/filepath"

	"github.com/cwbudde/go-dws/internal/units"
)

// evalFilename is the pseudo-filename the run command gives inline -e code. It
// has no directory of its own, so it contributes no script-relative unit path.
const evalFilename = "<eval>"

// resolveUnitSearchPaths builds the ordered unit search path list for a script.
//
// The precedence is:
//
//  1. The script's own directory, when no -I path was given. (With explicit -I
//     paths the caller asked for a specific search order, so the script
//     directory is not injected ahead of them.)
//  2. Every -I path, in the order given on the command line.
//  3. The default paths from units.GetDefaultSearchPaths: the current
//     directory, DWSCRIPT_PATH, ~/.dwscript/lib and the system library
//     directories.
//
// Duplicates are dropped, comparing directories by their absolute path so that
// e.g. "." and an explicit path to the current directory collapse into one
// entry while the original spelling is preserved.
func resolveUnitSearchPaths(filename string) []string {
	var paths []string
	seen := make(map[string]bool)

	add := func(dir string) {
		if dir == "" {
			return
		}
		key := dir
		if abs, err := filepath.Abs(dir); err == nil {
			key = abs
		}
		if seen[key] {
			return
		}
		seen[key] = true
		paths = append(paths, dir)
	}

	if len(unitSearchPaths) == 0 && filename != evalFilename {
		add(filepath.Dir(filename))
	}
	for _, dir := range unitSearchPaths {
		add(dir)
	}
	for _, dir := range units.GetDefaultSearchPaths() {
		add(dir)
	}

	return paths
}
