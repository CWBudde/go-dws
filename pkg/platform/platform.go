// Package platform provides a platform abstraction layer for go-dws,
// enabling the DWScript interpreter to run on different platforms
// (native Go, WebAssembly, etc.) with consistent behavior.
//
// The platform abstraction layer separates platform-dependent code (file I/O,
// console operations, time functions) from the core interpreter logic. This
// enables the same DWScript code to run identically on both native Go and
// WebAssembly environments.
//
// Core interfaces:
//   - FileSystem: abstraction for file I/O operations
//   - Console: abstraction for input/output operations
//   - Platform: overall platform abstraction combining filesystem, console, and time
//
// Implementations:
//   - platform/native: uses standard Go packages (os, io, time)
//   - platform/wasm: uses JavaScript APIs via syscall/js with virtual filesystem
//
// See docs/plans/2025-10-26-wasm-compilation-design.md for the complete design.
package platform

import "time"

// FileInfo represents basic information about a file or directory.
// This is a simplified version of os.FileInfo for cross-platform compatibility.
type FileInfo struct {
	ModTime time.Time
	Name    string
	Size    int64
	IsDir   bool
}

// FileSystem provides an abstraction for filesystem operations.
// This interface enables the DWScript interpreter to work with both real
// filesystems (native) and virtual filesystems (WASM).
//
// Implementations:
//   - Native: Direct access to OS filesystem via os package
//   - WASM: In-memory virtual filesystem (map[string][]byte) or IndexedDB
type FileSystem interface {
	// ReadFile reads the entire contents of a file.
	// Returns the file data or an error if the file doesn't exist or can't be read.
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file, creating it if it doesn't exist.
	// If the file exists, it is truncated before writing.
	// Returns an error if the write operation fails.
	WriteFile(path string, data []byte) error

	// ListDir returns a list of files and directories in the specified directory.
	// Returns an error if the directory doesn't exist or can't be read.
	ListDir(path string) ([]FileInfo, error)

	// Delete removes a file or empty directory.
	// Returns an error if the file/directory doesn't exist or can't be deleted.
	Delete(path string) error

	// Exists checks whether a file or directory exists at the given path.
	// Returns true if the path exists, false otherwise.
	Exists(path string) bool
}

// Console provides an abstraction for input/output operations.
// This interface enables the DWScript interpreter to work with different
// I/O mechanisms (terminal, browser console, custom callbacks).
//
// Implementations:
//   - Native: Uses os.Stdin, os.Stdout for terminal I/O
//   - WASM: Bridges to JavaScript console.log() or custom callbacks
type Console interface {
	// Print outputs text without a trailing newline.
	// This is used for inline output and prompts.
	Print(s string)

	// PrintLn outputs text with a trailing newline.
	// This is the most common output method in DWScript programs.
	PrintLn(s string)

	// ReadLine reads a line of input from the console.
	// Returns the input string (without newline) or an error.
	//
	// Native: Reads from stdin
	// WASM: May use window.prompt() or a custom callback
	ReadLine() (string, error)
}

// Platform provides the overall platform abstraction, combining filesystem,
// console, and time-related functionality.
//
// This is the main interface used by the DWScript interpreter to interact
// with platform-specific features. The interpreter receives a Platform
// instance and uses it for all platform-dependent operations.
type Platform interface {
	// FS returns the filesystem implementation for this platform.
	FS() FileSystem

	// Console returns the console implementation for this platform.
	Console() Console

	// Now returns the current time.
	// Native: Uses time.Now()
	// WASM: Uses JavaScript Date API
	Now() time.Time

	// Sleep pauses execution for the specified duration.
	// Native: Uses time.Sleep()
	// WASM: Uses JavaScript setTimeout with Promise/channel bridge
	Sleep(duration time.Duration)
}
