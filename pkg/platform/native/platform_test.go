//go:build !js && !wasm

package native

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

func TestNativeFileSystem_ReadWriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dwscript-fs-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var fs platform.FileSystem = &NativeFileSystem{}
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Hello, DWScript!")

	// Test WriteFile
	err = fs.WriteFile(testFile, testData)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	data, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("ReadFile returned wrong data: got %q, want %q", data, testData)
	}
}

func TestNativeFileSystem_Exists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dwscript-fs-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var fs platform.FileSystem = &NativeFileSystem{}
	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistentFile := filepath.Join(tempDir, "notexists.txt")

	// Create a file
	err = os.WriteFile(existingFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test Exists for existing file
	if !fs.Exists(existingFile) {
		t.Error("Exists returned false for existing file")
	}

	// Test Exists for non-existent file
	if fs.Exists(nonExistentFile) {
		t.Error("Exists returned true for non-existent file")
	}
}

func TestNativeFileSystem_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dwscript-fs-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var fs platform.FileSystem = &NativeFileSystem{}
	testFile := filepath.Join(tempDir, "delete.txt")

	// Create a file
	err = os.WriteFile(testFile, []byte("delete me"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test Delete
	err = fs.Delete(testFile)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify file is deleted
	if fs.Exists(testFile) {
		t.Error("File still exists after Delete")
	}
}

func TestNativeFileSystem_ListDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dwscript-fs-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var fs platform.FileSystem = &NativeFileSystem{}

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range testFiles {
		path := filepath.Join(tempDir, name)
		err := os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Test ListDir
	files, err := fs.ListDir(tempDir)
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}

	// Should have 3 files + 1 directory = 4 entries
	if len(files) != 4 {
		t.Errorf("ListDir returned %d entries, want 4", len(files))
	}

	// Verify we got the expected files
	fileNames := make(map[string]bool)
	for _, f := range files {
		fileNames[f.Name] = true
		if f.Name == "subdir" && !f.IsDir {
			t.Error("subdir should be marked as directory")
		}
		if strings.HasPrefix(f.Name, "file") && f.IsDir {
			t.Errorf("%s should not be marked as directory", f.Name)
		}
	}

	for _, expected := range testFiles {
		if !fileNames[expected] {
			t.Errorf("Expected file %s not found in ListDir results", expected)
		}
	}
	if !fileNames["subdir"] {
		t.Error("Expected directory 'subdir' not found in ListDir results")
	}
}

func TestNativeConsole_Print(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	console := &NativeConsole{
		output: w,
	}

	testText := "Hello, Console!"
	console.Print(testText)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != testText {
		t.Errorf("Print output = %q, want %q", output, testText)
	}
}

func TestNativeConsole_PrintLn(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	var console platform.Console = &NativeConsole{
		output: &buf,
	}

	testText := "Hello, Console!"
	console.PrintLn(testText)

	output := buf.String()
	expected := testText + "\n"

	if output != expected {
		t.Errorf("PrintLn output = %q, want %q", output, expected)
	}
}

func TestNativeConsole_ReadLine(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, _ := os.Pipe()
	console := &NativeConsole{
		input: r,
	}

	testInput := "Test input line\n"

	// Write test input in a goroutine
	go func() {
		w.Write([]byte(testInput))
		w.Close()
	}()

	// Read the line
	line, err := console.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine failed: %v", err)
	}

	expected := strings.TrimSuffix(testInput, "\n")
	if line != expected {
		t.Errorf("ReadLine = %q, want %q", line, expected)
	}
}

func TestNativePlatform_Now(t *testing.T) {
	var plat = NewNativePlatform()

	before := time.Now()
	now := plat.Now()
	after := time.Now()

	// The time returned should be between before and after
	if now.Before(before) || now.After(after) {
		t.Errorf("Now() returned unexpected time: %v (expected between %v and %v)", now, before, after)
	}
}

func TestNativePlatform_Sleep(t *testing.T) {
	var plat = NewNativePlatform()

	duration := 100 * time.Millisecond
	start := time.Now()
	plat.Sleep(duration)
	elapsed := time.Since(start)

	// Allow some tolerance (±20ms)
	if elapsed < duration-20*time.Millisecond || elapsed > duration+20*time.Millisecond {
		t.Errorf("Sleep duration = %v, want ~%v", elapsed, duration)
	}
}

func TestNativePlatform_Integration(t *testing.T) {
	// Create a complete platform instance
	var plat = NewNativePlatform()

	// Test that all components are non-nil
	if plat.FS() == nil {
		t.Error("Platform.FS() returned nil")
	}
	if plat.Console() == nil {
		t.Error("Platform.Console() returned nil")
	}

	// Test filesystem via platform
	tempDir, err := os.MkdirTemp("", "dwscript-platform-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "integration.txt")
	testData := []byte("Integration test")

	err = plat.FS().WriteFile(testFile, testData)
	if err != nil {
		t.Errorf("Platform filesystem WriteFile failed: %v", err)
	}

	data, err := plat.FS().ReadFile(testFile)
	if err != nil {
		t.Errorf("Platform filesystem ReadFile failed: %v", err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("Platform filesystem read wrong data: got %q, want %q", data, testData)
	}
}

func TestNativePlatform_ConsoleIntegration(t *testing.T) {
	var plat = NewNativePlatform()

	// Override the console to use a buffer for testing
	var buf bytes.Buffer
	console := &NativeConsole{
		output: &buf,
	}

	// Cast to concrete type to set console
	if nativePlat, ok := plat.(*NativePlatform); ok {
		nativePlat.console = console
	} else {
		t.Fatal("Platform is not *NativePlatform")
	}

	// Test console operations
	plat.Console().PrintLn("Test message")

	output := buf.String()
	expected := "Test message\n"
	if output != expected {
		t.Errorf("Console output = %q, want %q", output, expected)
	}
}

// Benchmark tests
func BenchmarkNativeFileSystem_ReadFile(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "dwscript-bench-")
	defer os.RemoveAll(tempDir)

	fs := &NativeFileSystem{}
	testFile := filepath.Join(tempDir, "bench.txt")
	testData := bytes.Repeat([]byte("test data "), 1000)
	fs.WriteFile(testFile, testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.ReadFile(testFile)
	}
}

func BenchmarkNativeFileSystem_WriteFile(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "dwscript-bench-")
	defer os.RemoveAll(tempDir)

	fs := &NativeFileSystem{}
	testData := bytes.Repeat([]byte("test data "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench"+string(rune(i))+".txt")
		fs.WriteFile(testFile, testData)
	}
}
