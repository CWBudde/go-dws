package wasm

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNormalizeFSPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "/"},
		{"whitespace only is a real name", "   ", "/   "},
		{"root", "/", "/"},
		{"relative", "a/b.txt", "/a/b.txt"},
		{"absolute", "/a/b.txt", "/a/b.txt"},
		{"backslashes", `C:\dir\file.txt`, "/C:/dir/file.txt"},
		{"dot segments", "/a/./b/../c.txt", "/a/c.txt"},
		{"trailing slash", "/a/b/", "/a/b"},
		{"double slash", "//a//b", "/a/b"},
		{"escape above root", "/../../etc/passwd", "/etc/passwd"},
		{"relative escape", "../secret", "/secret"},
		{"leading and trailing spaces are preserved", "/ reports ", "/ reports "},
		{"spaces inside a segment are preserved", "/my dir/a b.txt", "/my dir/a b.txt"},
		{"tab in a name is preserved", "/a\tb.txt", "/a\tb.txt"},
		{"spaced relative path", " notes ", "/ notes "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFSPath(tt.input); got != tt.want {
				t.Errorf("normalizeFSPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMissingFileSystemMethods(t *testing.T) {
	tests := []struct {
		name    string
		present []string
		want    []string
	}{
		{"complete", RequiredFileSystemMethods, nil},
		{"empty object", nil, RequiredFileSystemMethods},
		{
			"partial",
			[]string{"readFile", "exists"},
			[]string{"writeFile", "listDir", "delete"},
		},
		{
			"wrong case is not accepted",
			[]string{"ReadFile", "writeFile", "listDir", "delete", "exists"},
			[]string{"readFile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			present := make(map[string]bool, len(tt.present))
			for _, name := range tt.present {
				present[name] = true
			}
			got := missingFileSystemMethods(func(name string) bool { return present[name] })
			if len(got) != len(tt.want) {
				t.Fatalf("missing = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("missing = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestErrMissingFileSystemMethodsMentionsNames(t *testing.T) {
	err := errMissingFileSystemMethods([]string{"writeFile", "delete"})
	msg := err.Error()
	for _, want := range []string{"writeFile", "delete", "readFile"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not mention %q", msg, want)
		}
	}
}

func TestErrAsyncFileSystemIsExplicit(t *testing.T) {
	msg := errAsyncFileSystem("readFile").Error()
	if !strings.Contains(msg, "readFile") {
		t.Errorf("error %q does not name the operation", msg)
	}
	if !strings.Contains(msg, "Promise") || !strings.Contains(msg, "synchronous") {
		t.Errorf("error %q does not explain the synchronous requirement", msg)
	}
}

func TestFSOpErrorWrapsCause(t *testing.T) {
	cause := errors.New("boom")
	err := fsOpError("readFile", "/a.txt", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("fsOpError did not wrap the cause: %v", err)
	}
	if got, want := err.Error(), "readFile /a.txt: boom"; got != want {
		t.Errorf("err = %q, want %q", got, want)
	}
}

func TestDirEntryToFileInfo(t *testing.T) {
	t.Run("plain file", func(t *testing.T) {
		info, err := dirEntryToFileInfo(rawDirEntry{Name: "a.txt", Size: 12})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "a.txt" || info.Size != 12 || info.IsDir {
			t.Errorf("unexpected info: %+v", info)
		}
		if !info.ModTime.IsZero() {
			t.Errorf("ModTime should stay zero when the host supplies none, got %v", info.ModTime)
		}
	})

	t.Run("directory size is cleared", func(t *testing.T) {
		info, err := dirEntryToFileInfo(rawDirEntry{Name: "sub", Size: 99, IsDir: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !info.IsDir || info.Size != 0 {
			t.Errorf("unexpected info: %+v", info)
		}
	})

	t.Run("mod time from javascript millis", func(t *testing.T) {
		info, err := dirEntryToFileInfo(rawDirEntry{
			Name:          "a.txt",
			ModTimeMillis: 1700000000000,
			HasModTime:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.UnixMilli(1700000000000).UTC()
		if !info.ModTime.Equal(want) {
			t.Errorf("ModTime = %v, want %v", info.ModTime, want)
		}
	})

	t.Run("full path is reduced to the base name", func(t *testing.T) {
		info, err := dirEntryToFileInfo(rawDirEntry{Name: "/dir/sub/a.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "a.txt" {
			t.Errorf("Name = %q, want %q", info.Name, "a.txt")
		}
	})

	t.Run("empty name is rejected", func(t *testing.T) {
		if _, err := dirEntryToFileInfo(rawDirEntry{Name: ""}); err == nil {
			t.Fatal("expected an error for an empty entry name")
		}
	})

	t.Run("negative size is rejected", func(t *testing.T) {
		if _, err := dirEntryToFileInfo(rawDirEntry{Name: "a.txt", Size: -1}); err == nil {
			t.Fatal("expected an error for a negative size")
		}
	})
}

// TestDirEntryToFileInfo_Whitespace pins the rule that directory-entry names
// carry their whitespace: only slashes are stripped, so " draft " stays a
// distinct name from "draft".
func TestDirEntryToFileInfo_Whitespace(t *testing.T) {
	t.Run("whitespace in a name is preserved", func(t *testing.T) {
		for _, name := range []string{" draft ", "  ", "a b.txt"} {
			info, err := dirEntryToFileInfo(rawDirEntry{Name: name})
			if err != nil {
				t.Fatalf("dirEntryToFileInfo(%q): unexpected error: %v", name, err)
			}
			if info.Name != name {
				t.Errorf("Name = %q, want %q", info.Name, name)
			}
		}
	})

	t.Run("spaced base name survives a full path", func(t *testing.T) {
		info, err := dirEntryToFileInfo(rawDirEntry{Name: "/ reports / a.txt "})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != " a.txt " {
			t.Errorf("Name = %q, want %q", info.Name, " a.txt ")
		}
	})
}

func TestDirEntriesToFileInfos(t *testing.T) {
	infos, err := dirEntriesToFileInfos([]rawDirEntry{
		{Name: "a.txt", Size: 1},
		{Name: "sub", IsDir: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) != 2 || infos[0].Name != "a.txt" || !infos[1].IsDir {
		t.Fatalf("unexpected listing: %+v", infos)
	}

	_, err = dirEntriesToFileInfos([]rawDirEntry{{Name: "ok"}, {Name: ""}})
	if err == nil {
		t.Fatal("expected an error for a malformed entry")
	}
	if !strings.Contains(err.Error(), "entry 1") {
		t.Errorf("error %q does not identify the offending entry", err)
	}
}

func TestErrFileSystemAccessorWrapsCause(t *testing.T) {
	cause := errors.New("boom")
	err := errFileSystemAccessor("readFile", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("errFileSystemAccessor did not wrap the cause: %v", err)
	}
	if msg := err.Error(); !strings.Contains(msg, "readFile") {
		t.Errorf("error %q does not name the offending method", msg)
	}
}

func TestErrNotAnObject(t *testing.T) {
	if msg := errNotAnObject("function").Error(); !strings.Contains(msg, "function") {
		t.Errorf("error %q does not mention the offending type", msg)
	}
}
