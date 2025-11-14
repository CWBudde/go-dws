package interp

import (
	"fmt"
)

// I/O Built-in Functions
// Print, PrintLn (and their Write equivalents)

// builtinPrintLn implements the PrintLn built-in function.
// It prints all arguments followed by a newline.
// Like DWScript's WriteLn, arguments are concatenated directly.
func (i *Interpreter) builtinPrintLn(args []Value) Value {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output == nil {
		return &NilValue{}
	}
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			fmt.Fprint(i.output, "<nil>")
		} else {
			fmt.Fprint(i.output, arg.String())
		}
	}
	fmt.Fprintln(i.output)
	return &NilValue{}
}

// builtinPrint implements the Print built-in function.
// It prints all arguments without a newline.
// Like DWScript's Write, arguments are concatenated directly.
func (i *Interpreter) builtinPrint(args []Value) Value {
	// If output is nil, silently discard output (some tests use New(nil))
	if i.output == nil {
		return &NilValue{}
	}
	for _, arg := range args {
		// Handle nil arguments
		if arg == nil {
			fmt.Fprint(i.output, "<nil>")
		} else {
			fmt.Fprint(i.output, arg.String())
		}
	}
	return &NilValue{}
}
