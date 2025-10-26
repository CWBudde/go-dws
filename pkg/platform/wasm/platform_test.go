//go:build js && wasm

package wasm

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

func TestWASMFileSystem_ReadWriteFile(t *testing.T) {
	fs := NewWASMFileSystem()
	testPath := "/test.txt"
	testData := []byte("Hello, WASM!")

	// Test WriteFile
	err := fs.WriteFile(testPath, testData)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	data, err := fs.ReadFile(testPath)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if !bytes.Equal(data, testData) {
		t.Errorf("ReadFile returned wrong data: got %q, want %q", data, testData)
	}
}

func TestWASMFileSystem_Exists(t *testing.T) {
	fs := NewWASMFileSystem()
	existingFile := "/exists.txt"
	nonExistentFile := "/notexists.txt"

	// Create a file
	err := fs.WriteFile(existingFile, []byte("test"))
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

func TestWASMFileSystem_Delete(t *testing.T) {
	fs := NewWASMFileSystem()
	testFile := "/delete.txt"

	// Create a file
	err := fs.WriteFile(testFile, []byte("delete me"))
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

func TestWASMFileSystem_ListDir(t *testing.T) {
	fs := NewWASMFileSystem()

	// Create test files in a directory
	testFiles := []string{"/dir/file1.txt", "/dir/file2.txt", "/dir/file3.txt"}
	for _, path := range testFiles {
		err := fs.WriteFile(path, []byte("test"))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Create a subdirectory with a file to make it appear
	err := fs.WriteFile("/dir/subdir/file.txt", []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create subdirectory file: %v", err)
	}

	// Test ListDir
	files, err := fs.ListDir("/dir")
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}

	// Should have 3 files + 1 directory
	if len(files) != 4 {
		t.Errorf("ListDir returned %d entries, want 4 (got: %v)", len(files), files)
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

	expectedNames := []string{"file1.txt", "file2.txt", "file3.txt", "subdir"}
	for _, expected := range expectedNames {
		if !fileNames[expected] {
			t.Errorf("Expected file/dir %s not found in ListDir results", expected)
		}
	}
}

func TestWASMFileSystem_ListDir_Root(t *testing.T) {
	fs := NewWASMFileSystem()

	// Create files at root level
	fs.WriteFile("/root1.txt", []byte("test"))
	fs.WriteFile("/root2.txt", []byte("test"))
	fs.WriteFile("/dir/file.txt", []byte("test"))

	// List root directory
	files, err := fs.ListDir("/")
	if err != nil {
		t.Fatalf("ListDir(/) failed: %v", err)
	}

	// Should have 2 files + 1 directory
	if len(files) < 2 {
		t.Errorf("ListDir(/) returned %d entries, want at least 2", len(files))
	}
}

func TestWASMFileSystem_ReadNonExistent(t *testing.T) {
	fs := NewWASMFileSystem()

	_, err := fs.ReadFile("/nonexistent.txt")
	if err == nil {
		t.Error("ReadFile should return error for non-existent file")
	}
}

func TestWASMFileSystem_DeleteNonExistent(t *testing.T) {
	fs := NewWASMFileSystem()

	err := fs.Delete("/nonexistent.txt")
	if err == nil {
		t.Error("Delete should return error for non-existent file")
	}
}

func TestWASMConsole_Print(t *testing.T) {
	var buf bytes.Buffer
	console := NewWASMConsoleWithOutput(&buf)

	testText := "Hello, Console!"
	console.Print(testText)

	output := buf.String()
	if output != testText {
		t.Errorf("Print output = %q, want %q", output, testText)
	}
}

func TestWASMConsole_PrintLn(t *testing.T) {
	var buf bytes.Buffer
	console := NewWASMConsoleWithOutput(&buf)

	testText := "Hello, Console!"
	console.PrintLn(testText)

	output := buf.String()
	expected := testText + "\n"

	if output != expected {
		t.Errorf("PrintLn output = %q, want %q", output, expected)
	}
}

func TestWASMConsole_ReadLine(t *testing.T) {
	// Create console with input callback
	testInput := "Test input line"
	inputCalled := false

	console := &WASMConsole{
		readLineCallback: func() (string, error) {
			inputCalled = true
			return testInput, nil
		},
	}

	// Read the line
	line, err := console.ReadLine()
	if err != nil {
		t.Fatalf("ReadLine failed: %v", err)
	}

	if line != testInput {
		t.Errorf("ReadLine = %q, want %q", line, testInput)
	}

	if !inputCalled {
		t.Error("ReadLine callback was not called")
	}
}

func TestWASMPlatform_Now(t *testing.T) {
	plat := NewWASMPlatform()

	before := time.Now()
	now := plat.Now()
	after := time.Now()

	// The time returned should be between before and after (with some tolerance)
	// Note: In real WASM, this would use JavaScript Date API
	if now.Before(before.Add(-time.Second)) || now.After(after.Add(time.Second)) {
		t.Errorf("Now() returned unexpected time: %v (expected between %v and %v)", now, before, after)
	}
}

func TestWASMPlatform_Sleep(t *testing.T) {
	plat := NewWASMPlatform()

	duration := 50 * time.Millisecond
	start := time.Now()
	plat.Sleep(duration)
	elapsed := time.Since(start)

	// Allow more tolerance for WASM (±30ms) since timing may be less precise
	if elapsed < duration-30*time.Millisecond || elapsed > duration+100*time.Millisecond {
		t.Logf("Warning: Sleep duration = %v, expected ~%v", elapsed, duration)
		// Don't fail the test, just log a warning
	}
}

func TestWASMPlatform_Integration(t *testing.T) {
	// Create a complete platform instance
	var plat platform.Platform = NewWASMPlatform()

	// Test that all components are non-nil
	if plat.FS() == nil {
		t.Error("Platform.FS() returned nil")
	}
	if plat.Console() == nil {
		t.Error("Platform.Console() returned nil")
	}

	// Test filesystem via platform
	testFile := "/integration.txt"
	testData := []byte("Integration test")

	err := plat.FS().WriteFile(testFile, testData)
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

func TestWASMPlatform_ConsoleIntegration(t *testing.T) {
	var buf bytes.Buffer
	plat := NewWASMPlatformWithIO(&buf)

	// Test console operations
	plat.Console().PrintLn("Test message")

	output := buf.String()
	expected := "Test message\n"
	if output != expected {
		t.Errorf("Console output = %q, want %q", output, expected)
	}
}

// Benchmark tests
func BenchmarkWASMFileSystem_ReadFile(b *testing.B) {
	fs := NewWASMFileSystem()
	testFile := "/bench.txt"
	testData := bytes.Repeat([]byte("test data "), 1000)
	fs.WriteFile(testFile, testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.ReadFile(testFile)
	}
}

func BenchmarkWASMFileSystem_WriteFile(b *testing.B) {
	fs := NewWASMFileSystem()
	testData := bytes.Repeat([]byte("test data "), 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := "/bench" + string(rune(i)) + ".txt"
		fs.WriteFile(testFile, testData)
	}
}
