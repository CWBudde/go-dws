//go:build !js && !wasm

// Package native provides a platform implementation for native Go environments.
// It uses standard Go packages (os, io, time) to provide filesystem, console,
// and time functionality.
//
// This implementation is used when compiling for native platforms (Linux, macOS,
// Windows, etc.) and provides direct access to the operating system's facilities.
package native

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

// NativeFileSystem implements the FileSystem interface using the standard os package.
// It provides direct access to the operating system's filesystem.
type NativeFileSystem struct{}

// ReadFile reads the entire contents of a file from the native filesystem.
func (fs *NativeFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file in the native filesystem.
// If the file doesn't exist, it is created with 0644 permissions.
// If it exists, it is truncated before writing.
func (fs *NativeFileSystem) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// ListDir returns information about files and directories in the specified path.
func (fs *NativeFileSystem) ListDir(path string) ([]platform.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]platform.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// Skip entries we can't get info for
			continue
		}

		result = append(result, platform.FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	return result, nil
}

// Delete removes a file or empty directory from the native filesystem.
func (fs *NativeFileSystem) Delete(path string) error {
	return os.Remove(path)
}

// Exists checks whether a file or directory exists in the native filesystem.
func (fs *NativeFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// NativeConsole implements the Console interface using standard I/O.
// It reads from stdin and writes to stdout/stderr.
type NativeConsole struct {
	input  io.Reader
	output io.Writer
}

// Print outputs text to stdout without a trailing newline.
func (c *NativeConsole) Print(s string) {
	fmt.Fprint(c.output, s)
}

// PrintLn outputs text to stdout with a trailing newline.
func (c *NativeConsole) PrintLn(s string) {
	fmt.Fprintln(c.output, s)
}

// ReadLine reads a line of text from stdin.
// The trailing newline is removed from the returned string.
func (c *NativeConsole) ReadLine() (string, error) {
	scanner := bufio.NewScanner(c.input)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

// NativePlatform implements the Platform interface for native Go environments.
// It combines NativeFileSystem and NativeConsole with native time functions.
type NativePlatform struct {
	fs      *NativeFileSystem
	console *NativeConsole
}

// NewNativePlatform creates a new native platform instance with default
// stdin/stdout configuration.
func NewNativePlatform() platform.Platform {
	return &NativePlatform{
		fs: &NativeFileSystem{},
		console: &NativeConsole{
			input:  os.Stdin,
			output: os.Stdout,
		},
	}
}

// NewNativePlatformWithIO creates a new native platform instance with custom
// input and output streams. This is useful for testing or when you want to
// redirect I/O to different destinations.
func NewNativePlatformWithIO(input io.Reader, output io.Writer) platform.Platform {
	return &NativePlatform{
		fs: &NativeFileSystem{},
		console: &NativeConsole{
			input:  input,
			output: output,
		},
	}
}

// FS returns the native filesystem implementation.
func (p *NativePlatform) FS() platform.FileSystem {
	return p.fs
}

// Console returns the native console implementation.
func (p *NativePlatform) Console() platform.Console {
	return p.console
}

// Now returns the current time using time.Now().
func (p *NativePlatform) Now() time.Time {
	return time.Now()
}

// Sleep pauses execution for the specified duration using time.Sleep().
func (p *NativePlatform) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
