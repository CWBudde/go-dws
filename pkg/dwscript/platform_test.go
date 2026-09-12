package dwscript

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/pkg/platform"
)

// memFS is an in-memory FileSystem that records every path a script touches,
// so a test can assert that the file built-ins went through the installed
// platform rather than the real disk.
type memFS struct {
	files map[string][]byte
	reads []string
	wrote []string
}

func newMemFS() *memFS {
	return &memFS{files: map[string][]byte{}}
}

func (m *memFS) ReadFile(path string) ([]byte, error) {
	m.reads = append(m.reads, path)
	data, ok := m.files[path]
	if !ok {
		return nil, errors.New("no such file: " + path)
	}
	return data, nil
}

func (m *memFS) WriteFile(path string, data []byte) error {
	m.wrote = append(m.wrote, path)
	m.files[path] = append([]byte(nil), data...)
	return nil
}

func (m *memFS) ListDir(string) ([]platform.FileInfo, error) { return nil, nil }
func (m *memFS) Delete(path string) error                    { delete(m.files, path); return nil }
func (m *memFS) Exists(path string) bool                     { _, ok := m.files[path]; return ok }

// memPlatform pairs memFS with the remaining platform services, which these
// tests do not exercise.
type memPlatform struct{ fs *memFS }

func (p *memPlatform) FS() platform.FileSystem { return p.fs }
func (p *memPlatform) Console() platform.Console {
	return nil
}
func (p *memPlatform) Now() time.Time      { return time.Unix(0, 0) }
func (p *memPlatform) Sleep(time.Duration) {}

func TestWithPlatform_FileBuiltinsUseTheInstalledFilesystem(t *testing.T) {
	fs := newMemFS()

	var out bytes.Buffer
	engine, err := New(WithPlatform(&memPlatform{fs: fs}), WithOutput(&out))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const source = `
SaveTextToFile('greeting.txt', 'hello');
PrintLn(LoadTextFromFile('greeting.txt'));
`

	if _, err := engine.Eval(source); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if got := out.String(); got != "hello\n" {
		t.Errorf("output = %q, want %q", got, "hello\n")
	}
	// The write must have landed in the installed filesystem, not on disk.
	if got := string(fs.files["greeting.txt"]); got != "hello" {
		t.Errorf("in-memory file = %q, want %q", got, "hello")
	}
	if len(fs.wrote) != 1 || fs.wrote[0] != "greeting.txt" {
		t.Errorf("writes = %v, want exactly [greeting.txt]", fs.wrote)
	}
	if len(fs.reads) != 1 || fs.reads[0] != "greeting.txt" {
		t.Errorf("reads = %v, want exactly [greeting.txt]", fs.reads)
	}
}

// A missing file must surface as an error rather than an empty string, so a
// bad path cannot be mistaken for an empty file.
func TestWithPlatform_MissingFileIsAnError(t *testing.T) {
	engine, err := New(WithPlatform(&memPlatform{fs: newMemFS()}), WithOutput(&bytes.Buffer{}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = engine.Eval(`PrintLn(LoadTextFromFile('absent.txt'));`)
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if !strings.Contains(err.Error(), "absent.txt") {
		t.Errorf("error %q does not name the missing path", err)
	}
}

// An engine created without WithPlatform still reports a usable platform, so
// callers never have to nil-check Engine.FS().
func TestEngine_DefaultPlatformIsNeverNil(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if engine.Platform() == nil {
		t.Fatal("Platform() = nil, want the build default")
	}
	if engine.FS() == nil {
		t.Fatal("FS() = nil, want the default platform's filesystem")
	}
}

func TestWithPlatform_RejectsNil(t *testing.T) {
	if _, err := New(WithPlatform(nil)); err == nil {
		t.Fatal("expected WithPlatform(nil) to be rejected")
	}
}
