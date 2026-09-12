package builtins

import (
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/platform"
)

// fakeFS is a minimal in-memory FileSystem. ReadFile can be told to fail so the
// error path is covered without depending on a real unreadable file.
type fakeFS struct {
	files   map[string][]byte
	readErr error
}

func newFakeFS() *fakeFS { return &fakeFS{files: map[string][]byte{}} }

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	data, ok := f.files[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}

func (f *fakeFS) WriteFile(path string, data []byte) error {
	f.files[path] = append([]byte(nil), data...)
	return nil
}

func (f *fakeFS) ListDir(string) ([]platform.FileInfo, error) { return nil, nil }
func (f *fakeFS) Delete(path string) error                    { delete(f.files, path); return nil }
func (f *fakeFS) Exists(path string) bool                     { _, ok := f.files[path]; return ok }

var _ platform.FileSystem = (*fakeFS)(nil)

func contextWithFS(fs *fakeFS) *mockContext {
	ctx := newMockContext()
	ctx.fs = fs
	return ctx
}

func str(s string) Value { return &runtime.StringValue{Value: s} }

func TestSaveTextToFile_And_LoadTextFromFile(t *testing.T) {
	fs := newFakeFS()
	ctx := contextWithFS(fs)

	if got := SaveTextToFile(ctx, []Value{str("notes.txt"), str("body")}); isMockError(got) {
		t.Fatalf("SaveTextToFile failed: %s", got.String())
	}
	if got := string(fs.files["notes.txt"]); got != "body" {
		t.Errorf("stored %q, want %q", got, "body")
	}

	loaded := LoadTextFromFile(ctx, []Value{str("notes.txt")})
	if isMockError(loaded) {
		t.Fatalf("LoadTextFromFile failed: %s", loaded.String())
	}
	if got := loaded.String(); got != "body" {
		t.Errorf("loaded %q, want %q", got, "body")
	}
}

// A read failure must reach the script as an error naming the path, not as an
// empty string that looks like a legitimately empty file.
func TestLoadTextFromFile_ReadErrorNamesThePath(t *testing.T) {
	fs := newFakeFS()
	fs.readErr = errors.New("permission denied")
	ctx := contextWithFS(fs)

	got := LoadTextFromFile(ctx, []Value{str("secret.txt")})
	if !isMockError(got) {
		t.Fatalf("expected an error, got %q", got.String())
	}
	if !strings.Contains(ctx.lastError, "secret.txt") {
		t.Errorf("error %q does not name the path", ctx.lastError)
	}
	if !strings.Contains(ctx.lastError, "permission denied") {
		t.Errorf("error %q drops the underlying cause", ctx.lastError)
	}
}

func TestFileBuiltins_RejectBadArguments(t *testing.T) {
	fs := newFakeFS()

	tests := []struct {
		call func(Context) Value
		name string
	}{
		{name: "load with no arguments", call: func(c Context) Value { return LoadTextFromFile(c, nil) }},
		{name: "load with two arguments", call: func(c Context) Value {
			return LoadTextFromFile(c, []Value{str("a"), str("b")})
		}},
		{name: "load with a non-string path", call: func(c Context) Value {
			return LoadTextFromFile(c, []Value{&runtime.IntegerValue{Value: 1}})
		}},
		{name: "save with one argument", call: func(c Context) Value {
			return SaveTextToFile(c, []Value{str("a")})
		}},
		{name: "save with a non-string path", call: func(c Context) Value {
			return SaveTextToFile(c, []Value{&runtime.IntegerValue{Value: 1}, str("b")})
		}},
		{name: "save with non-string contents", call: func(c Context) Value {
			return SaveTextToFile(c, []Value{str("a"), &runtime.IntegerValue{Value: 1}})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contextWithFS(fs)
			if got := tt.call(ctx); !isMockError(got) {
				t.Fatalf("expected an error, got %q", got.String())
			}
		})
	}
}

func isMockError(v Value) bool {
	_, ok := v.(*mockErrorValue)
	return ok
}
