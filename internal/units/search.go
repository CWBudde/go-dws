package units

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// FindUnit searches for a unit file by name in the given search paths.
// It tries common DWScript file extensions (.dws, .pas) and supports both
// relative and absolute paths.
//
// Search order:
//  1. Current directory (if "." is in paths)
//  2. Each specified path in order
//  3. Tries both UnitName.dws and UnitName.pas
//
// Returns:
//   - The absolute path to the unit file if found
//   - An error if the unit file cannot be found in any search path
//
// Example:
//
//	path, err := FindUnit("MyUnit", []string{".", "./lib", "/usr/share/dwscript"})
func FindUnit(name string, paths []string) (string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Common DWScript file extensions
	extensions := []string{".dws", ".pas"}

	// Track all attempted paths for error message
	attempted := []string{}

	// Try each search path
	for _, searchPath := range paths {
		// Handle empty path
		if searchPath == "" {
			continue
		}

		// Convert to absolute path for consistency
		absPath, err := filepath.Abs(searchPath)
		if err != nil {
			// Skip invalid paths
			continue
		}

		// Check if the path exists and is a directory
		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			// Skip non-existent or non-directory paths
			continue
		}

		// Try each extension
		for _, ext := range extensions {
			// Build the full file path
			fileName := name + ext
			fullPath := filepath.Join(absPath, fileName)
			attempted = append(attempted, fullPath)

			// Check if file exists
			if fileExists(fullPath) {
				return fullPath, nil
			}

			// Also try with capitalized first letter (common convention)
			if len(name) > 0 {
				capitalizedName := strings.ToUpper(name[:1]) + ident.Normalize(name[1:])
				if capitalizedName != name {
					capitalizedFileName := capitalizedName + ext
					capitalizedFullPath := filepath.Join(absPath, capitalizedFileName)
					attempted = append(attempted, capitalizedFullPath)

					if fileExists(capitalizedFullPath) {
						return capitalizedFullPath, nil
					}
				}
			}

			// Also try all lowercase (another common convention)
			lowercaseName := ident.Normalize(name)
			if lowercaseName != name {
				lowercaseFileName := lowercaseName + ext
				lowercaseFullPath := filepath.Join(absPath, lowercaseFileName)
				attempted = append(attempted, lowercaseFullPath)

				if fileExists(lowercaseFullPath) {
					return lowercaseFullPath, nil
				}
			}

			// Also try all uppercase
			uppercaseName := strings.ToUpper(name)
			if uppercaseName != name {
				uppercaseFileName := uppercaseName + ext
				uppercaseFullPath := filepath.Join(absPath, uppercaseFileName)
				attempted = append(attempted, uppercaseFullPath)

				if fileExists(uppercaseFullPath) {
					return uppercaseFullPath, nil
				}
			}
		}
	}

	// Unit not found in any search path
	return "", fmt.Errorf(
		"unit file not found: '%s' (searched %d locations: %s)",
		name,
		len(attempted),
		strings.Join(attempted[:min(5, len(attempted))], ", "), // Show first 5 attempts
	)
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a path exists and is a directory.
// It is the directory-side counterpart of fileExists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FindUnitInPath searches for a unit file in a single directory path.
// This is a convenience function that calls FindUnit with a single search path.
func FindUnitInPath(name, path string) (string, error) {
	return FindUnit(name, []string{path})
}

// AddSearchPath adds a new search path to a list of search paths if it's not already present.
// The path is converted to an absolute path before adding.
func AddSearchPath(paths []string, newPath string) ([]string, error) {
	absPath, err := filepath.Abs(newPath)
	if err != nil {
		return paths, fmt.Errorf("invalid search path '%s': %w", newPath, err)
	}

	// Check if already in the list
	for _, p := range paths {
		if p == absPath {
			return paths, nil // Already present
		}
	}

	// Add to the list
	return append(paths, absPath), nil
}

// SearchPathEnvVar is the environment variable holding extra unit search paths.
// Its value is a list of directories separated by the platform list separator
// (":" on Unix, ";" on Windows), searched after the current directory and before
// the user and system library directories.
const SearchPathEnvVar = "DWSCRIPT_PATH"

// GetDefaultSearchPaths returns the default search paths for units, in order:
//
//  1. Current directory (".")
//  2. Each existing directory listed in the DWSCRIPT_PATH environment variable
//  3. The user's DWScript library directory, ~/.dwscript/lib (if it exists)
//  4. The system DWScript library directories (if they exist):
//     /usr/local/share/dwscript/lib then /usr/share/dwscript/lib on Unix,
//     %ProgramData%\dwscript\lib on Windows
//
// Every entry except "." is made absolute and de-duplicated, and is only
// included when the directory actually exists.
func GetDefaultSearchPaths() []string {
	// os.UserHomeDir fails when no home directory is configured; an empty
	// home simply skips the user library path.
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	return defaultSearchPaths(home, os.Getenv(SearchPathEnvVar), systemLibraryDirs(), dirExists)
}

// systemLibraryDirs returns the OS-specific system library directories, most
// specific first. A runtime.GOOS switch is used rather than build tags because
// the whole list is a handful of string constants: keeping it in one function
// makes the full cross-platform ordering reviewable in one place, and the
// interesting logic lives in defaultSearchPaths, which takes the list as a
// parameter and is therefore testable on every OS.
func systemLibraryDirs() []string {
	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			return nil
		}
		return []string{filepath.Join(programData, "dwscript", "lib")}
	}

	return []string{
		filepath.Join("/usr", "local", "share", "dwscript", "lib"),
		filepath.Join("/usr", "share", "dwscript", "lib"),
	}
}

// defaultSearchPaths builds the default search path list from explicit inputs.
// It is the testable core of GetDefaultSearchPaths: home is the user's home
// directory ("" when unknown), envPath the raw DWSCRIPT_PATH value, sysDirs the
// system library directories in priority order, and exists reports whether a
// directory is present.
func defaultSearchPaths(home, envPath string, sysDirs []string, exists func(string) bool) []string {
	paths := []string{"."}

	add := func(dir string) {
		if dir == "" || !exists(dir) {
			return
		}
		updated, err := AddSearchPath(paths, dir)
		if err != nil {
			return
		}
		paths = updated
	}

	for _, dir := range filepath.SplitList(envPath) {
		add(dir)
	}

	if home != "" {
		add(filepath.Join(home, ".dwscript", "lib"))
	}

	for _, dir := range sysDirs {
		add(dir)
	}

	return paths
}
