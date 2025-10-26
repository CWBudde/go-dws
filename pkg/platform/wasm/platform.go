//go:build js && wasm

// Package wasm provides a platform implementation for WebAssembly environments.
// It uses an in-memory virtual filesystem and bridges to JavaScript APIs for
// console I/O and time functions.
//
// This implementation is used when compiling for WASM target (GOOS=js GOARCH=wasm).
// The virtual filesystem stores all files in memory as a map[string][]byte and
// does not persist data. Future versions may add IndexedDB support for persistence.
package wasm

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

// WASMFileSystem implements a virtual in-memory filesystem for WASM environments.
// Files are stored in a map and do not persist across page reloads.
//
// The filesystem uses Unix-style paths (forward slashes) and supports basic
// directory navigation. Directories are virtual (implicitly created by file paths).
type WASMFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewWASMFileSystem creates a new virtual filesystem.
func NewWASMFileSystem() *WASMFileSystem {
	return &WASMFileSystem{
		files: make(map[string][]byte),
	}
}

// ReadFile reads the contents of a file from the virtual filesystem.
func (fs *WASMFileSystem) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, exists := fs.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// WriteFile writes data to a file in the virtual filesystem.
// The file is created if it doesn't exist.
func (fs *WASMFileSystem) WriteFile(filePath string, data []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Store a copy to prevent external modification
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	fs.files[filePath] = dataCopy
	return nil
}

// ListDir returns information about files in a directory.
// Directories are virtual - they're implicitly created by file paths.
func (fs *WASMFileSystem) ListDir(dirPath string) ([]platform.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize directory path
	if dirPath == "" {
		dirPath = "/"
	}
	if !strings.HasSuffix(dirPath, "/") && dirPath != "/" {
		dirPath += "/"
	}
	if !strings.HasPrefix(dirPath, "/") {
		dirPath = "/" + dirPath
	}

	// Find all files in this directory (not in subdirectories)
	entries := make(map[string]platform.FileInfo)

	for filePath, data := range fs.files {
		// Check if file is in this directory
		if !strings.HasPrefix(filePath, dirPath) {
			continue
		}

		// Get relative path from directory
		relativePath := strings.TrimPrefix(filePath, dirPath)
		if relativePath == "" {
			continue
		}

		// Check if this is a direct child or in a subdirectory
		parts := strings.Split(relativePath, "/")
		if len(parts) == 0 {
			continue
		}

		entryName := parts[0]

		if len(parts) == 1 {
			// This is a file directly in the directory
			if _, exists := entries[entryName]; !exists {
				entries[entryName] = platform.FileInfo{
					Name:    entryName,
					Size:    int64(len(data)),
					IsDir:   false,
					ModTime: time.Now(),
				}
			}
		} else {
			// This is in a subdirectory - add the subdirectory entry
			if _, exists := entries[entryName]; !exists {
				entries[entryName] = platform.FileInfo{
					Name:    entryName,
					Size:    0,
					IsDir:   true,
					ModTime: time.Now(),
				}
			}
		}
	}

	// Convert map to slice
	result := make([]platform.FileInfo, 0, len(entries))
	for _, info := range entries {
		result = append(result, info)
	}

	return result, nil
}

// Delete removes a file from the virtual filesystem.
func (fs *WASMFileSystem) Delete(filePath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.files[filePath]; !exists {
		return fmt.Errorf("file not found: %s", filePath)
	}

	delete(fs.files, filePath)
	return nil
}

// Exists checks whether a file exists in the virtual filesystem.
func (fs *WASMFileSystem) Exists(filePath string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	_, exists := fs.files[filePath]
	return exists
}

// WASMConsole implements console I/O for WASM environments.
// In a real WASM build, this would bridge to JavaScript console.log(),
// window.prompt(), or custom callbacks provided by the JavaScript side.
type WASMConsole struct {
	output           io.Writer
	readLineCallback func() (string, error)
}

// NewWASMConsole creates a new WASM console with default behavior.
// In a real WASM environment, this would set up JavaScript bridges.
func NewWASMConsole() *WASMConsole {
	return &WASMConsole{
		readLineCallback: func() (string, error) {
			return "", fmt.Errorf("ReadLine not implemented in WASM environment")
		},
	}
}

// NewWASMConsoleWithOutput creates a WASM console with a custom output writer.
// This is useful for testing or custom output handling.
func NewWASMConsoleWithOutput(output io.Writer) *WASMConsole {
	return &WASMConsole{
		output: output,
		readLineCallback: func() (string, error) {
			return "", fmt.Errorf("ReadLine not implemented in WASM environment")
		},
	}
}

// SetReadLineCallback sets a custom callback for reading input.
// In a real WASM build, this would be called from JavaScript.
func (c *WASMConsole) SetReadLineCallback(callback func() (string, error)) {
	c.readLineCallback = callback
}

// Print outputs text without a trailing newline.
// In a real WASM build, this would call console.log() via syscall/js.
func (c *WASMConsole) Print(s string) {
	if c.output != nil {
		fmt.Fprint(c.output, s)
	}
	// In real WASM: js.Global().Get("console").Call("log", s)
}

// PrintLn outputs text with a trailing newline.
// In a real WASM build, this would call console.log() via syscall/js.
func (c *WASMConsole) PrintLn(s string) {
	if c.output != nil {
		fmt.Fprintln(c.output, s)
	}
	// In real WASM: js.Global().Get("console").Call("log", s)
}

// ReadLine reads a line of input.
// In a real WASM build, this would use window.prompt() or a custom callback.
func (c *WASMConsole) ReadLine() (string, error) {
	if c.readLineCallback != nil {
		return c.readLineCallback()
	}
	return "", fmt.Errorf("ReadLine not available in WASM")
	// In real WASM: return js.Global().Call("prompt", "").String(), nil
}

// WASMPlatform implements the Platform interface for WASM environments.
// It combines WASMFileSystem and WASMConsole with time functions.
type WASMPlatform struct {
	fs      *WASMFileSystem
	console *WASMConsole
}

// NewWASMPlatform creates a new WASM platform instance with default configuration.
func NewWASMPlatform() platform.Platform {
	return &WASMPlatform{
		fs:      NewWASMFileSystem(),
		console: NewWASMConsole(),
	}
}

// NewWASMPlatformWithIO creates a new WASM platform with custom output.
// This is useful for testing or custom I/O handling.
func NewWASMPlatformWithIO(output io.Writer) platform.Platform {
	return &WASMPlatform{
		fs:      NewWASMFileSystem(),
		console: NewWASMConsoleWithOutput(output),
	}
}

// NewWASMPlatformWithCallbacks creates a WASM platform with custom JavaScript callbacks.
// This would be used in real WASM builds to connect to JavaScript functionality.
func NewWASMPlatformWithCallbacks(outputCallback func(string), inputCallback func() (string, error)) platform.Platform {
	// Create custom console that uses callbacks
	console := NewWASMConsole()
	if inputCallback != nil {
		console.SetReadLineCallback(inputCallback)
	}

	return &WASMPlatform{
		fs:      NewWASMFileSystem(),
		console: console,
	}
}

// FS returns the virtual filesystem implementation.
func (p *WASMPlatform) FS() platform.FileSystem {
	return p.fs
}

// Console returns the WASM console implementation.
func (p *WASMPlatform) Console() platform.Console {
	return p.console
}

// Now returns the current time.
// In a real WASM build, this would use JavaScript Date API via syscall/js.
func (p *WASMPlatform) Now() time.Time {
	return time.Now()
	// In real WASM:
	// ms := js.Global().Get("Date").Call("now").Int()
	// return time.Unix(0, ms*int64(time.Millisecond))
}

// Sleep pauses execution for the specified duration.
// In a real WASM build, this would use JavaScript setTimeout via syscall/js
// with a channel to make it blocking.
func (p *WASMPlatform) Sleep(duration time.Duration) {
	time.Sleep(duration)
	// In real WASM with proper async handling:
	// done := make(chan struct{})
	// js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	//     close(done)
	//     return nil
	// }), int(duration.Milliseconds()))
	// <-done
}

// GetFileSystem returns the underlying WASMFileSystem for direct access.
// This is useful for advanced operations not covered by the interface.
func (p *WASMPlatform) GetFileSystem() *WASMFileSystem {
	return p.fs
}
