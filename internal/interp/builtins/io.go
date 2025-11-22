package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// ============================================================================
// I/O Built-in Functions
// ============================================================================
//
// This file contains I/O functions that have been migrated from internal/interp
// to use the Context interface pattern.
//
// Functions in this file:
//   - Print: Print arguments without newline
//   - PrintLn: Print arguments with newline
//
// These functions use Context helper methods (Write, WriteLine) to access
// the output writer without creating circular dependencies.

// Print prints all arguments without a newline.
// Print(...values)
//
// Arguments are converted to strings and concatenated directly.
// Nil arguments are rendered as "<nil>".
//
// This corresponds to DWScript's Write() function.
func Print(ctx Context, args []Value) Value {
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			ctx.Write("<nil>")
		} else {
			ctx.Write(arg.String())
		}
	}
	return &runtime.NilValue{}
}

// PrintLn prints all arguments followed by a newline.
// PrintLn(...values)
//
// Arguments are converted to strings and concatenated directly.
// Nil arguments are rendered as "<nil>".
// A newline is appended after all arguments.
//
// This corresponds to DWScript's WriteLn() function.
func PrintLn(ctx Context, args []Value) Value {
	// Build the output string from all arguments
	var output string
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			output += "<nil>"
		} else {
			output += arg.String()
		}
	}
	// Write with newline
	ctx.WriteLine(output)
	return &runtime.NilValue{}
}
