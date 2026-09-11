//go:build js && wasm

package wasm

import (
	"bytes"
	"strings"
	"syscall/js"
	"testing"

	platformwasm "github.com/cwbudde/go-dws/pkg/platform/wasm"
)

// newJSObject builds a JavaScript object whose properties are Go-backed
// functions, releasing them when the test ends.
func newJSObject(t *testing.T, methods map[string]func(args []js.Value) any) js.Value {
	t.Helper()
	obj := js.Global().Get("Object").New()
	for name, fn := range methods {
		impl := fn
		jsFunc := js.FuncOf(func(_ js.Value, args []js.Value) any { return impl(args) })
		t.Cleanup(jsFunc.Release)
		obj.Set(name, jsFunc)
	}
	return obj
}

// jsFunction compiles a real JavaScript function, which is the only way to get
// host code that genuinely throws or returns a Promise.
func jsFunction(body string) js.Value {
	return js.Global().Get("Function").New("path", "data", body)
}

// newMemoryJSFS builds a working synchronous filesystem object backed by a Go map.
func newMemoryJSFS(t *testing.T, files map[string]string) (js.Value, map[string]string) {
	t.Helper()
	store := make(map[string]string, len(files))
	for k, v := range files {
		store[k] = v
	}
	seen := make(map[string]string) // records the paths the host was called with

	obj := newJSObject(t, map[string]func(args []js.Value) any{
		"readFile": func(args []js.Value) any {
			p := args[0].String()
			seen["readFile"] = p
			data, ok := store[p]
			if !ok {
				return js.Null()
			}
			return bytesToUint8Array([]byte(data))
		},
		"writeFile": func(args []js.Value) any {
			p := args[0].String()
			seen["writeFile"] = p
			buf := make([]byte, args[1].Get("length").Int())
			js.CopyBytesToGo(buf, args[1])
			store[p] = string(buf)
			return js.Undefined()
		},
		"listDir": func(args []js.Value) any {
			seen["listDir"] = args[0].String()
			arr := js.Global().Get("Array").New()
			arr.Call("push", "plain.txt")
			entry := js.Global().Get("Object").New()
			entry.Set("name", "sub")
			entry.Set("isDir", true)
			arr.Call("push", entry)
			sized := js.Global().Get("Object").New()
			sized.Set("name", "sized.txt")
			sized.Set("size", 7)
			sized.Set("modTime", 1700000000000)
			arr.Call("push", sized)
			return arr
		},
		"delete": func(args []js.Value) any {
			p := args[0].String()
			seen["delete"] = p
			delete(store, p)
			return js.Undefined()
		},
		"exists": func(args []js.Value) any {
			p := args[0].String()
			seen["exists"] = p
			_, ok := store[p]
			return ok
		},
	})
	t.Cleanup(func() {
		for k, v := range store {
			files[k] = v
		}
	})
	_ = seen
	return obj, seen
}

func TestNewJSFileSystemRejectsNonObjects(t *testing.T) {
	for _, v := range []js.Value{js.ValueOf("nope"), js.ValueOf(42), js.Null(), js.Undefined()} {
		if _, err := NewJSFileSystem(v); err == nil {
			t.Errorf("expected an error for %s", jsTypeName(v))
		}
	}
}

func TestNewJSFileSystemRejectsIncompleteObjects(t *testing.T) {
	obj := newJSObject(t, map[string]func(args []js.Value) any{
		"readFile": func([]js.Value) any { return js.Null() },
		"exists":   func([]js.Value) any { return false },
	})

	_, err := NewJSFileSystem(obj)
	if err == nil {
		t.Fatal("expected an error for an incomplete filesystem object")
	}
	for _, want := range []string{"writeFile", "listDir", "delete"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention missing method %q", err, want)
		}
	}
}

func TestJSFileSystemRoundTrip(t *testing.T) {
	files := map[string]string{"/hello.txt": "world"}
	obj, seen := newMemoryJSFS(t, files)

	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	data, err := fs.ReadFile("hello.txt") // relative on purpose
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.Equal(data, []byte("world")) {
		t.Errorf("ReadFile = %q, want %q", data, "world")
	}
	if seen["readFile"] != "/hello.txt" {
		t.Errorf("host received path %q, want %q (paths must be normalized)", seen["readFile"], "/hello.txt")
	}

	if err := fs.WriteFile("/dir/../out.txt", []byte("payload")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if seen["writeFile"] != "/out.txt" {
		t.Errorf("host received path %q, want %q", seen["writeFile"], "/out.txt")
	}
	if !fs.Exists("/out.txt") {
		t.Error("Exists(/out.txt) = false after WriteFile")
	}
	written, err := fs.ReadFile("/out.txt")
	if err != nil || string(written) != "payload" {
		t.Errorf("round trip = %q, %v", written, err)
	}

	if err := fs.Delete("/out.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if fs.Exists("/out.txt") {
		t.Error("Exists(/out.txt) = true after Delete")
	}
}

func TestJSFileSystemReadFileMissingReturnsError(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{})
	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	if _, err := fs.ReadFile("/nope.txt"); err == nil {
		t.Fatal("expected an error when the host returns null")
	} else if !strings.Contains(err.Error(), "/nope.txt") {
		t.Errorf("error %q does not mention the path", err)
	}
}

func TestJSFileSystemListDir(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{})
	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	infos, err := fs.ListDir("/dir/")
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(infos) != 3 {
		t.Fatalf("ListDir returned %d entries, want 3: %+v", len(infos), infos)
	}
	if infos[0].Name != "plain.txt" || infos[0].IsDir {
		t.Errorf("string entry decoded as %+v", infos[0])
	}
	if infos[1].Name != "sub" || !infos[1].IsDir {
		t.Errorf("directory entry decoded as %+v", infos[1])
	}
	if infos[2].Size != 7 || infos[2].ModTime.IsZero() {
		t.Errorf("sized entry decoded as %+v", infos[2])
	}
}

func TestJSFileSystemRejectsPromises(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{"/a.txt": "x"})
	obj.Set("readFile", jsFunction("return Promise.resolve(new Uint8Array(0));"))

	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	_, err = fs.ReadFile("/a.txt")
	if err == nil {
		t.Fatal("expected an error for a Promise-returning readFile")
	}
	if !strings.Contains(err.Error(), "Promise") || !strings.Contains(err.Error(), "synchronous") {
		t.Errorf("error %q does not explain the synchronous requirement", err)
	}
}

func TestJSFileSystemMapsThrownErrors(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{})
	obj.Set("readFile", jsFunction("throw new Error('permission denied');"))
	obj.Set("exists", jsFunction("throw new Error('nope');"))

	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	_, err = fs.ReadFile("/a.txt")
	if err == nil {
		t.Fatal("expected an error for a throwing readFile")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error %q does not carry the JavaScript message", err)
	}

	// Exists has no error channel: a throwing host method means "no".
	if fs.Exists("/a.txt") {
		t.Error("Exists = true for a throwing host method, want false")
	}
}

func TestJSFileSystemRejectsBadResults(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{})
	obj.Set("listDir", jsFunction("return 'not an array';"))
	obj.Set("readFile", jsFunction("return 42;"))

	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	if _, err := fs.ListDir("/"); err == nil {
		t.Error("expected an error when listDir does not return an array")
	}
	if _, err := fs.ReadFile("/a.txt"); err == nil {
		t.Error("expected an error when readFile returns a number")
	}
}

func TestJSFileSystemAcceptsStringResults(t *testing.T) {
	obj, _ := newMemoryJSFS(t, map[string]string{})
	obj.Set("readFile", jsFunction("return 'plain text';"))

	fs, err := NewJSFileSystem(obj)
	if err != nil {
		t.Fatalf("NewJSFileSystem: %v", err)
	}

	data, err := fs.ReadFile("/a.txt")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "plain text" {
		t.Errorf("ReadFile = %q, want %q", data, "plain text")
	}
}

func TestContextInstallFileSystem(t *testing.T) {
	plat, ok := platformwasm.NewWASMPlatform().(*platformwasm.WASMPlatform)
	if !ok {
		t.Fatal("NewWASMPlatform did not return *WASMPlatform")
	}
	ctx := &Context{platform: plat}

	// A valid object replaces the built-in virtual filesystem.
	obj, _ := newMemoryJSFS(t, map[string]string{"/a.txt": "from host"})
	if err := ctx.installFileSystem(obj); err != nil {
		t.Fatalf("installFileSystem: %v", err)
	}
	if !plat.HasCustomFileSystem() {
		t.Fatal("platform did not record the custom filesystem")
	}
	data, err := plat.FS().ReadFile("/a.txt")
	if err != nil || string(data) != "from host" {
		t.Fatalf("platform FS did not route to the host: %q, %v", data, err)
	}

	// null restores the built-in virtual filesystem.
	if err := ctx.installFileSystem(js.Null()); err != nil {
		t.Fatalf("installFileSystem(null): %v", err)
	}
	if plat.HasCustomFileSystem() {
		t.Fatal("null did not reset the custom filesystem")
	}
	if plat.FS().Exists("/a.txt") {
		t.Error("built-in virtual filesystem unexpectedly sees host files")
	}

	// An invalid object is rejected and leaves the platform untouched.
	bad := newJSObject(t, map[string]func(args []js.Value) any{
		"readFile": func([]js.Value) any { return js.Null() },
	})
	if err := ctx.installFileSystem(bad); err == nil {
		t.Fatal("expected an error for an incomplete filesystem object")
	}
	if plat.HasCustomFileSystem() {
		t.Error("a rejected filesystem must not be installed")
	}
}
