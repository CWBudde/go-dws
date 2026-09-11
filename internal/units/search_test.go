package units

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindUnit(t *testing.T) {
	// Create temporary test directory structure
	tempDir := t.TempDir()

	// Create test files with different naming conventions
	testFiles := []string{
		"MyUnit.dws",
		"Myunit.dws", // For case-insensitive search test
		"lowercase.dws",
		"UPPERCASE.dws",
		"TestUnit.pas",
		"Another.pas",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("// test"), 0644)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name        string
		unitName    string
		expectedExt string
		searchPaths []string
		shouldFind  bool
	}{
		{
			name:        "Find exact match .dws",
			unitName:    "MyUnit",
			searchPaths: []string{tempDir},
			shouldFind:  true,
			expectedExt: ".dws",
		},
		{
			name:        "Find lowercase",
			unitName:    "lowercase",
			searchPaths: []string{tempDir},
			shouldFind:  true,
			expectedExt: ".dws",
		},
		{
			name:        "Find uppercase",
			unitName:    "UPPERCASE",
			searchPaths: []string{tempDir},
			shouldFind:  true,
			expectedExt: ".dws",
		},
		{
			name:        "Find .pas file",
			unitName:    "TestUnit",
			searchPaths: []string{tempDir},
			shouldFind:  true,
			expectedExt: ".pas",
		},
		{
			name:        "Case insensitive search (capitalized)",
			unitName:    "myunit",
			searchPaths: []string{tempDir},
			shouldFind:  true, // Will find "Myunit.dws" created by capitalization logic
			expectedExt: ".dws",
		},
		{
			name:        "Not found",
			unitName:    "NonExistent",
			searchPaths: []string{tempDir},
			shouldFind:  false,
		},
		{
			name:        "Empty search paths uses current dir",
			unitName:    "MyUnit",
			searchPaths: []string{},
			shouldFind:  false, // Won't find it in current dir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindUnit(tt.unitName, tt.searchPaths)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("expected to find unit, got error: %v", err)
				}

				if path == "" {
					t.Error("expected non-empty path")
				}

				if !strings.HasSuffix(path, tt.expectedExt) {
					t.Errorf("expected path to end with %s, got: %s", tt.expectedExt, path)
				}

				// Verify file exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("returned path does not exist: %s", path)
				}
			} else {
				if err == nil {
					t.Error("expected error when unit not found")
				}

				if !strings.Contains(err.Error(), "not found") {
					t.Errorf("expected 'not found' error, got: %v", err)
				}
			}
		})
	}
}

func TestFindUnit_MultipleSearchPaths(t *testing.T) {
	// Create two temporary directories
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	// Create unit in tempDir2 only
	unitPath := filepath.Join(tempDir2, "TestUnit.dws")
	err := os.WriteFile(unitPath, []byte("// test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Search in both directories (tempDir1 first, but file is in tempDir2)
	path, err := FindUnit("TestUnit", []string{tempDir1, tempDir2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != unitPath {
		t.Errorf("expected path %s, got %s", unitPath, path)
	}
}

func TestFindUnit_PrefersDws(t *testing.T) {
	tempDir := t.TempDir()

	// Create both .dws and .pas files
	dwsPath := filepath.Join(tempDir, "MyUnit.dws")
	pasPath := filepath.Join(tempDir, "MyUnit.pas")

	os.WriteFile(dwsPath, []byte("// dws"), 0644)
	os.WriteFile(pasPath, []byte("// pas"), 0644)

	// Should prefer .dws over .pas
	path, err := FindUnit("MyUnit", []string{tempDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != dwsPath {
		t.Errorf("expected to find .dws file first, got: %s", path)
	}
}

func TestFindUnit_CaseVariations(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different casing that our search algorithm can find
	// On case-sensitive filesystems, we can only find exact matches or standard conventions
	files := []string{
		"MySpecialUnit.dws", // Exact match
		"Myspecialunit.dws", // Capitalized first letter (what we generate for lowercase search)
		"myspecialunit.dws", // All lowercase
		"MYSPECIALUNIT.dws", // All uppercase
	}

	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if err := os.WriteFile(path, []byte("// test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name       string
		searchName string
		shouldFind bool
	}{
		{"Exact case", "MySpecialUnit", true},
		{"All lowercase", "myspecialunit", true},
		{"All uppercase", "MYSPECIALUNIT", true},
		{"Capitalized", "Myspecialunit", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindUnit(tt.searchName, []string{tempDir})

			if tt.shouldFind {
				if err != nil {
					t.Errorf("expected to find unit, got error: %v", err)
				}
				if path == "" {
					t.Error("expected non-empty path")
				}
			} else {
				if err == nil {
					t.Error("expected error")
				}
			}
		})
	}
}

func TestFindUnit_InvalidPaths(t *testing.T) {
	tests := []struct {
		name        string
		searchPaths []string
	}{
		{"Non-existent directory", []string{"/non/existent/path"}},
		{"Empty path in list", []string{"", "/another/path"}},
		{"File instead of directory", []string{"/etc/hosts"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindUnit("TestUnit", tt.searchPaths)
			if err == nil {
				t.Error("expected error when searching invalid paths")
			}
		})
	}
}

func TestFindUnitInPath(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	unitPath := filepath.Join(tempDir, "TestUnit.dws")
	err := os.WriteFile(unitPath, []byte("// test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test the convenience function
	path, err := FindUnitInPath("TestUnit", tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != unitPath {
		t.Errorf("expected path %s, got %s", unitPath, path)
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{"Existing file", testFile, true},
		{"Non-existent file", filepath.Join(tempDir, "nonexistent.txt"), false},
		{"Directory", tempDir, false}, // Should return false for directories
		{"Empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fileExists(tt.path)
			if result != tt.exists {
				t.Errorf("fileExists(%s) = %v, expected %v", tt.path, result, tt.exists)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestAddSearchPath(t *testing.T) {
	t.Run("Add new path", func(t *testing.T) {
		paths := []string{"/path1"}
		newPaths, err := AddSearchPath(paths, "/path2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(newPaths) != 2 {
			t.Errorf("expected 2 paths, got %d", len(newPaths))
		}
	})

	t.Run("Add duplicate path", func(t *testing.T) {
		tempDir := t.TempDir()
		paths := []string{tempDir}
		newPaths, err := AddSearchPath(paths, tempDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(newPaths) != 1 {
			t.Errorf("expected 1 path (no duplicate), got %d", len(newPaths))
		}
	})

	t.Run("Add relative path", func(t *testing.T) {
		paths := []string{}
		newPaths, err := AddSearchPath(paths, ".")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(newPaths) != 1 {
			t.Errorf("expected 1 path, got %d", len(newPaths))
		}

		// Should be converted to absolute
		if !filepath.IsAbs(newPaths[0]) {
			t.Error("expected absolute path")
		}
	})
}

func TestGetDefaultSearchPaths(t *testing.T) {
	extra := t.TempDir()
	t.Setenv(SearchPathEnvVar, extra+string(filepath.ListSeparator)+filepath.Join(extra, "missing"))

	paths := GetDefaultSearchPaths()

	if len(paths) == 0 {
		t.Fatal("expected at least one default search path")
	}

	// First path should be current directory
	if paths[0] != "." {
		t.Errorf("expected first path to be '.', got '%s'", paths[0])
	}

	// The existing DWSCRIPT_PATH entry must be present, the missing one must not.
	absExtra, err := filepath.Abs(extra)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	if len(paths) < 2 || paths[1] != absExtra {
		t.Errorf("expected DWSCRIPT_PATH entry '%s' right after '.', got %v", absExtra, paths)
	}
	for _, p := range paths {
		if p == filepath.Join(absExtra, "missing") {
			t.Errorf("non-existent DWSCRIPT_PATH entry should be skipped, got %v", paths)
		}
	}

	// Every entry but "." must be absolute, and no duplicates are allowed.
	seen := map[string]bool{}
	for _, p := range paths {
		if p != "." && !filepath.IsAbs(p) {
			t.Errorf("expected absolute path, got '%s'", p)
		}
		if seen[p] {
			t.Errorf("duplicate search path '%s' in %v", p, paths)
		}
		seen[p] = true
	}
}

func TestDefaultSearchPaths(t *testing.T) {
	// existsIn returns an exists func accepting exactly the listed directories.
	existsIn := func(dirs ...string) func(string) bool {
		set := make(map[string]bool, len(dirs))
		for _, d := range dirs {
			set[d] = true
		}
		return func(path string) bool { return set[path] }
	}

	abs := func(t *testing.T, path string) string {
		t.Helper()
		a, err := filepath.Abs(path)
		if err != nil {
			t.Fatalf("filepath.Abs(%q): %v", path, err)
		}
		return a
	}

	home := abs(t, "testhome")
	userLib := filepath.Join(home, ".dwscript", "lib")
	sysA := abs(t, filepath.Join("sys", "a"))
	sysB := abs(t, filepath.Join("sys", "b"))
	envA := abs(t, filepath.Join("env", "a"))
	envB := abs(t, filepath.Join("env", "b"))
	sep := string(filepath.ListSeparator)

	tests := []struct {
		exists  func(string) bool
		name    string
		home    string
		envPath string
		sysDirs []string
		want    []string
	}{
		{
			name:   "no home and nothing exists",
			exists: existsIn(),
			want:   []string{"."},
		},
		{
			name:    "home without the lib directory",
			home:    home,
			sysDirs: []string{sysA},
			exists:  existsIn(home),
			want:    []string{"."},
		},
		{
			name:   "home with the lib directory",
			home:   home,
			exists: existsIn(userLib),
			want:   []string{".", userLib},
		},
		{
			name:    "no home skips the user library",
			sysDirs: []string{sysA},
			exists:  existsIn(userLib, sysA),
			want:    []string{".", sysA},
		},
		{
			name:    "env entries come before user and system",
			home:    home,
			envPath: envA + sep + envB,
			sysDirs: []string{sysA, sysB},
			exists:  existsIn(envA, envB, userLib, sysA, sysB),
			want:    []string{".", envA, envB, userLib, sysA, sysB},
		},
		{
			name:    "system directories keep their order",
			sysDirs: []string{sysA, sysB},
			exists:  existsIn(sysA, sysB),
			want:    []string{".", sysA, sysB},
		},
		{
			name:    "missing env entries are skipped",
			envPath: envA + sep + envB,
			exists:  existsIn(envB),
			want:    []string{".", envB},
		},
		{
			name:    "empty env entries are ignored",
			envPath: sep + envA + sep + sep,
			exists:  existsIn(envA),
			want:    []string{".", envA},
		},
		{
			name:    "duplicates are deduplicated",
			home:    home,
			envPath: envA + sep + envA + sep + userLib,
			sysDirs: []string{sysA, sysA},
			exists:  existsIn(envA, userLib, sysA),
			want:    []string{".", envA, userLib, sysA},
		},
		{
			name:    "empty system directory entries are ignored",
			sysDirs: []string{"", sysA},
			exists:  existsIn(sysA),
			want:    []string{".", sysA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultSearchPaths(tt.home, tt.envPath, tt.sysDirs, tt.exists)

			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("expected %v, got %v", tt.want, got)
				}
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "unit.dws")
	if err := os.WriteFile(file, []byte("unit Unit;"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !dirExists(tempDir) {
		t.Error("expected dirExists to be true for a directory")
	}
	if dirExists(file) {
		t.Error("expected dirExists to be false for a regular file")
	}
	if dirExists(filepath.Join(tempDir, "nope")) {
		t.Error("expected dirExists to be false for a missing path")
	}
}

func TestFindUnit_ErrorMessage(t *testing.T) {
	tempDir := t.TempDir()

	_, err := FindUnit("NonExistent", []string{tempDir})
	if err == nil {
		t.Fatal("expected error")
	}

	errMsg := err.Error()

	// Should contain useful information
	if !strings.Contains(errMsg, "not found") {
		t.Error("error message should mention 'not found'")
	}

	if !strings.Contains(errMsg, "NonExistent") {
		t.Error("error message should mention the unit name")
	}

	// Should show some attempted paths
	if !strings.Contains(errMsg, "searched") {
		t.Error("error message should mention search attempts")
	}
}

func TestFindUnit_RelativeAndAbsolutePaths(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	unitPath := filepath.Join(tempDir, "TestUnit.dws")
	err := os.WriteFile(unitPath, []byte("// test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	t.Run("Absolute path", func(t *testing.T) {
		path, err := FindUnit("TestUnit", []string{tempDir})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !filepath.IsAbs(path) {
			t.Error("expected absolute path to be returned")
		}
	})

	t.Run("Relative path", func(t *testing.T) {
		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)

		os.Chdir(tempDir)

		path, err := FindUnit("TestUnit", []string{"."})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should still return absolute path
		if !filepath.IsAbs(path) {
			t.Error("expected absolute path to be returned even with relative search path")
		}
	})
}
