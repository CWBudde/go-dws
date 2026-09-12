package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// File Built-in Functions
// ============================================================================
//
// These are the first built-ins to reach the host filesystem, and they do it
// only through Context.FS() — the seam an embedder controls with
// dwscript.WithPlatform. Nothing here touches the os package directly, so a
// WASM host that installs a JavaScript-backed filesystem gets the same
// behaviour as a native one.
//
// Functions in this file:
//   - LoadTextFromFile: read a whole file as a string
//   - SaveTextToFile: write a string to a file, replacing its contents

// LoadTextFromFile reads the named file and returns its contents as a string.
// LoadTextFromFile(path: String): String
//
// A missing or unreadable file is a runtime error rather than an empty string,
// so a typo in a path does not silently look like an empty file.
func LoadTextFromFile(ctx Context, args []Value) Value {
	if len(args) != 1 {
		return ctx.NewError("LoadTextFromFile expects 1 argument, got %d", len(args))
	}

	path, ok := toStringArg(ctx, args[0])
	if !ok {
		return ctx.NewError("LoadTextFromFile expects a string path, got %s", ctx.GetTypeOf(args[0]))
	}

	data, err := ctx.FS().ReadFile(path)
	if err != nil {
		return ctx.NewError("LoadTextFromFile(%q): %v", path, err)
	}

	return &runtime.StringValue{Value: string(data)}
}

// SaveTextToFile writes text to the named file, creating it if absent and
// truncating it if present.
// SaveTextToFile(path: String; text: String)
func SaveTextToFile(ctx Context, args []Value) Value {
	if len(args) != 2 {
		return ctx.NewError("SaveTextToFile expects 2 arguments, got %d", len(args))
	}

	path, ok := toStringArg(ctx, args[0])
	if !ok {
		return ctx.NewError("SaveTextToFile expects a string path, got %s", ctx.GetTypeOf(args[0]))
	}
	text, ok := toStringArg(ctx, args[1])
	if !ok {
		return ctx.NewError("SaveTextToFile expects string contents, got %s", ctx.GetTypeOf(args[1]))
	}

	if err := ctx.FS().WriteFile(path, []byte(text)); err != nil {
		return ctx.NewError("SaveTextToFile(%q): %v", path, err)
	}

	return &runtime.NilValue{}
}

// toStringArg unwraps a Variant and reports the argument as a string. A
// non-string value is rejected rather than coerced: these built-ins take
// paths, where a silent conversion would turn a type error into a file error.
func toStringArg(ctx Context, arg Value) (string, bool) {
	if arg == nil {
		return "", false
	}
	if str, ok := ctx.UnwrapVariant(arg).(*runtime.StringValue); ok {
		return str.Value, true
	}
	return "", false
}
