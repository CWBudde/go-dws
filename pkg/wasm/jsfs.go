//go:build js && wasm

package wasm

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/cwbudde/go-dws/pkg/platform"
)

// JSFileSystem adapts a JavaScript filesystem object to platform.FileSystem.
//
// The wrapped object must implement RequiredFileSystemMethods and every method
// must return synchronously; see the package documentation for the exact
// contract. Exceptions thrown by the host are converted into Go errors, and a
// method that returns a Promise fails immediately with errAsyncFileSystem
// rather than deadlocking the WASM event loop.
type JSFileSystem struct {
	obj js.Value
}

// compile-time check that the adapter satisfies the platform contract.
var _ platform.FileSystem = (*JSFileSystem)(nil)

// NewJSFileSystem validates a JavaScript filesystem object and wraps it.
// It returns an error when the value is not an object or when any required
// method is missing, so hosts learn about a malformed filesystem on
// installation instead of at first use.
func NewJSFileSystem(obj js.Value) (*JSFileSystem, error) {
	if obj.Type() != js.TypeObject {
		return nil, errNotAnObject(jsTypeName(obj))
	}
	missing := missingFileSystemMethods(func(name string) bool {
		return obj.Get(name).Type() == js.TypeFunction
	})
	if len(missing) > 0 {
		return nil, errMissingFileSystemMethods(missing)
	}
	return &JSFileSystem{obj: obj}, nil
}

// jsTypeName renders a js.Value's type for error messages.
func jsTypeName(v js.Value) string {
	switch v.Type() {
	case js.TypeUndefined:
		return "undefined"
	case js.TypeNull:
		return "null"
	case js.TypeBoolean:
		return "boolean"
	case js.TypeNumber:
		return "number"
	case js.TypeString:
		return "string"
	case js.TypeSymbol:
		return "symbol"
	case js.TypeFunction:
		return "function"
	case js.TypeObject:
		return "object"
	default:
		return "unknown"
	}
}

// isThenable reports whether a value looks like a Promise.
func isThenable(v js.Value) bool {
	if v.Type() != js.TypeObject && v.Type() != js.TypeFunction {
		return false
	}
	return v.Get("then").Type() == js.TypeFunction
}

// call invokes one filesystem method, translating JavaScript exceptions into
// Go errors and rejecting Promise results.
func (fs *JSFileSystem) call(op string, args ...any) (result js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = js.Undefined()
			err = jsPanicToError(r)
		}
	}()

	result = fs.obj.Call(op, args...)
	if isThenable(result) {
		return js.Undefined(), errAsyncFileSystem(op)
	}
	return result, nil
}

// jsPanicToError converts a recovered syscall/js panic into a Go error.
func jsPanicToError(r any) error {
	if jsErr, ok := r.(js.Error); ok {
		return errors.New(jsErr.Error())
	}
	return fmt.Errorf("javascript error: %v", r)
}

// jsValueToBytes converts a readFile result (Uint8Array, ArrayBuffer-backed
// typed array, or string) into a byte slice.
func jsValueToBytes(v js.Value) (data []byte, err error) {
	switch v.Type() {
	case js.TypeString:
		return []byte(v.String()), nil
	case js.TypeNull, js.TypeUndefined:
		return nil, errors.New("readFile returned no data")
	case js.TypeObject:
		defer func() {
			if r := recover(); r != nil {
				data = nil
				err = fmt.Errorf("readFile result is not a Uint8Array: %w", jsPanicToError(r))
			}
		}()
		length := v.Get("length")
		if length.Type() != js.TypeNumber {
			return nil, errors.New("readFile result is not a Uint8Array (no length property)")
		}
		buf := make([]byte, length.Int())
		if n := js.CopyBytesToGo(buf, v); n != len(buf) {
			return nil, fmt.Errorf("readFile copied %d of %d bytes", n, len(buf))
		}
		return buf, nil
	default:
		return nil, fmt.Errorf("readFile returned an unsupported value of type %s", jsTypeName(v))
	}
}

// bytesToUint8Array copies a byte slice into a fresh JavaScript Uint8Array.
func bytesToUint8Array(data []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(data))
	if len(data) > 0 {
		js.CopyBytesToJS(arr, data)
	}
	return arr
}

// ReadFile reads a whole file through the JavaScript host.
func (fs *JSFileSystem) ReadFile(filePath string) ([]byte, error) {
	normalized := normalizeFSPath(filePath)
	res, err := fs.call("readFile", normalized)
	if err != nil {
		return nil, fsOpError("readFile", normalized, err)
	}
	data, err := jsValueToBytes(res)
	if err != nil {
		return nil, fsOpError("readFile", normalized, err)
	}
	return data, nil
}

// WriteFile writes a whole file through the JavaScript host.
func (fs *JSFileSystem) WriteFile(filePath string, data []byte) error {
	normalized := normalizeFSPath(filePath)
	if _, err := fs.call("writeFile", normalized, bytesToUint8Array(data)); err != nil {
		return fsOpError("writeFile", normalized, err)
	}
	return nil
}

// ListDir lists a directory through the JavaScript host. Entries may be plain
// strings (names) or objects with name/size/isDir/modTime properties.
func (fs *JSFileSystem) ListDir(dirPath string) ([]platform.FileInfo, error) {
	normalized := normalizeFSPath(dirPath)
	res, err := fs.call("listDir", normalized)
	if err != nil {
		return nil, fsOpError("listDir", normalized, err)
	}
	if !js.Global().Get("Array").Call("isArray", res).Bool() {
		return nil, fsOpError("listDir", normalized,
			fmt.Errorf("expected an array, got %s", jsTypeName(res)))
	}

	length := res.Length()
	raw := make([]rawDirEntry, 0, length)
	for i := range length {
		entry, entryErr := jsDirEntry(res.Index(i))
		if entryErr != nil {
			return nil, fsOpError("listDir", normalized, fmt.Errorf("entry %d: %w", i, entryErr))
		}
		raw = append(raw, entry)
	}

	infos, err := dirEntriesToFileInfos(raw)
	if err != nil {
		return nil, fsOpError("listDir", normalized, err)
	}
	return infos, nil
}

// jsDirEntry converts a single JavaScript directory entry into its neutral form.
func jsDirEntry(v js.Value) (rawDirEntry, error) {
	switch v.Type() {
	case js.TypeString:
		return rawDirEntry{Name: v.String()}, nil
	case js.TypeObject:
		entry := rawDirEntry{}
		name := v.Get("name")
		if name.Type() != js.TypeString {
			return rawDirEntry{}, errors.New("entry object has no string \"name\" property")
		}
		entry.Name = name.String()
		if size := v.Get("size"); size.Type() == js.TypeNumber {
			entry.Size = int64(size.Float())
		}
		if isDir := v.Get("isDir"); !isDir.IsUndefined() {
			entry.IsDir = isDir.Truthy()
		}
		if modTime := v.Get("modTime"); modTime.Type() == js.TypeNumber {
			entry.ModTimeMillis = modTime.Float()
			entry.HasModTime = true
		}
		return entry, nil
	default:
		return rawDirEntry{}, fmt.Errorf("expected a string or object, got %s", jsTypeName(v))
	}
}

// Delete removes a file or empty directory through the JavaScript host.
func (fs *JSFileSystem) Delete(filePath string) error {
	normalized := normalizeFSPath(filePath)
	if _, err := fs.call("delete", normalized); err != nil {
		return fsOpError("delete", normalized, err)
	}
	return nil
}

// Exists reports whether a path exists. The platform contract has no error
// channel here, so a throwing or async host method is reported as a console
// warning and treated as "does not exist".
func (fs *JSFileSystem) Exists(filePath string) bool {
	normalized := normalizeFSPath(filePath)
	res, err := fs.call("exists", normalized)
	if err != nil {
		ConsoleWarn(fsOpError("exists", normalized, err).Error())
		return false
	}
	return res.Truthy()
}
