package platform_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/platform"
	"github.com/cwbudde/go-dws/pkg/platform/native"
)

// TestPlatformInterface verifies that native platform implements the Platform interface.
func TestPlatformInterface(t *testing.T) {
	var _ = native.NewNativePlatform()

	plat := native.NewNativePlatform()
	if plat == nil {
		t.Fatal("NewNativePlatform returned nil")
	}

	// Verify all interface methods are available
	if plat.FS() == nil {
		t.Error("Platform.FS() returned nil")
	}
	if plat.Console() == nil {
		t.Error("Platform.Console() returned nil")
	}

	// Test basic operations
	now := plat.Now()
	if now.IsZero() {
		t.Error("Platform.Now() returned zero time")
	}
}

// TestFileSystemInterface verifies that FileSystem interface works correctly.
func TestFileSystemInterface(t *testing.T) {
	plat := native.NewNativePlatform()
	var _ = plat.FS()

	fs := plat.FS()
	if fs == nil {
		t.Fatal("Platform.FS() returned nil")
	}

	// Verify all FileSystem methods are available
	exists := fs.Exists("/nonexistent")
	if exists {
		t.Error("Exists returned true for non-existent file")
	}
}

// TestConsoleInterface verifies that Console interface works correctly.
func TestConsoleInterface(t *testing.T) {
	plat := native.NewNativePlatform()
	var _ = plat.Console()

	console := plat.Console()
	if console == nil {
		t.Fatal("Platform.Console() returned nil")
	}

	// Verify Console methods are available (just checking they don't panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Console methods panicked: %v", r)
		}
	}()

	console.Print("")
	console.PrintLn("")
}

// TestFileInfoStruct verifies the FileInfo struct.
func TestFileInfoStruct(t *testing.T) {
	info := platform.FileInfo{
		Name:  "test.txt",
		Size:  1024,
		IsDir: false,
	}

	if info.Name != "test.txt" {
		t.Errorf("FileInfo.Name = %q, want %q", info.Name, "test.txt")
	}
	if info.Size != 1024 {
		t.Errorf("FileInfo.Size = %d, want %d", info.Size, 1024)
	}
	if info.IsDir {
		t.Error("FileInfo.IsDir should be false")
	}
}
