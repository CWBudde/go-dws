package bytecode

// registerBuiltins registers all built-in functions with the VM.
// The actual implementations are split across multiple files for better organization:
// - vm_builtins_misc.go: Print, PrintLn, Length
// - vm_builtins_conversion.go: Type conversion functions
// - vm_builtins_string.go: String manipulation functions
// - vm_builtins_math.go: Math functions
func (vm *VM) registerBuiltins() {
	// Register miscellaneous functions (Print, I/O, array helpers)
	vm.registerMiscBuiltins()

	// Register type conversion functions
	vm.registerConversionBuiltins()

	// Register string manipulation functions
	vm.registerStringBuiltins()

	// Register math functions
	vm.registerMathBuiltins()
}
