package bytecode

import (
	"fmt"
)

// registerMiscBuiltins registers miscellaneous functions (Print, I/O, array helpers)
func (vm *VM) registerMiscBuiltins() {
	vm.builtins["PrintLn"] = builtinPrintLn
	vm.builtins["Print"] = builtinPrint
	vm.builtins["Length"] = builtinLength
}

// I/O Functions

func builtinPrintLn(vm *VM, args []Value) (Value, error) {
	if vm.output != nil {
		for i, arg := range args {
			if i > 0 {
				fmt.Fprint(vm.output, " ")
			}
			// Unquote strings for output
			if arg.IsString() {
				fmt.Fprint(vm.output, arg.AsString())
			} else {
				fmt.Fprint(vm.output, arg.String())
			}
		}
		fmt.Fprintln(vm.output)
	}
	return NilValue(), nil
}

func builtinPrint(vm *VM, args []Value) (Value, error) {
	if vm.output != nil {
		for i, arg := range args {
			if i > 0 {
				fmt.Fprint(vm.output, " ")
			}
			// Unquote strings for output
			if arg.IsString() {
				fmt.Fprint(vm.output, arg.AsString())
			} else {
				fmt.Fprint(vm.output, arg.String())
			}
		}
	}
	return NilValue(), nil
}

// Array/String Helper Functions

func builtinLength(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Length expects 1 argument, got %d", len(args))
	}
	arg := args[0]
	if arg.IsString() {
		return IntValue(int64(len(arg.AsString()))), nil
	}
	if arg.IsArray() {
		arr := arg.AsArray()
		if arr != nil {
			return IntValue(int64(len(arr.elements))), nil
		}
	}
	return NilValue(), vm.runtimeError("Length expects a string or array argument")
}
