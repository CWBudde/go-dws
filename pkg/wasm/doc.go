// Package wasm provides WebAssembly-specific functionality for go-dws,
// including JavaScript/Go interop and browser API bindings.
//
// Most of the package is guarded by the "js && wasm" build tags because it
// depends on syscall/js. The host-testable parts (path normalization, custom
// filesystem validation, and error mapping in fsbridge.go) are deliberately
// kept free of build tags so they can be unit-tested on any platform.
//
// # Custom filesystems
//
// A JavaScript host can supply its own filesystem implementation, either via
// the "fs" option of init() or via setFileSystem(). The object must expose
// five methods that all return synchronously:
//
//	readFile(path)        -> Uint8Array | string
//	writeFile(path, data) -> any (ignored)
//	listDir(path)         -> Array<string | {name, size, isDir, modTime}>
//	delete(path)          -> any (ignored)
//	exists(path)          -> boolean
//
// Synchronous is a hard requirement: platform.FileSystem is a synchronous Go
// interface, and blocking a Go goroutine on a JavaScript Promise deadlocks the
// single-threaded WASM event loop. A method that returns a thenable therefore
// fails with an explicit error instead of hanging. Hosts that only have async
// storage (IndexedDB, fetch) must pre-load into memory and serve the cached
// copy synchronously.
package wasm
