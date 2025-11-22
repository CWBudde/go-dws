package units

import (
	"fmt"
	"os"
	"path/filepath"
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

// GetDefaultSearchPaths returns the default search paths for units.
// This includes:
//   - Current directory (".")
//   - User's DWScript library directory (if it exists)
//   - System DWScript library directory (if it exists)
func GetDefaultSearchPaths() []string {
	paths := []string{"."}

	// TODO: Add user library path (e.g., ~/.dwscript/lib)
	// TODO: Add system library path (e.g., /usr/share/dwscript/lib)

	return paths
}
