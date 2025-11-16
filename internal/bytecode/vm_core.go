package bytecode

import (
	"io"
	"math/rand"
)

// Default VM configuration constants.
const (
	defaultStackCapacity = 256
	defaultFrameCapacity = 16
)

// BuiltinFunction represents a built-in function callable from bytecode.
type BuiltinFunction func(vm *VM, args []Value) (Value, error)

// VM executes bytecode chunks produced by the compiler.
type VM struct {
	output            io.Writer
	builtins          map[string]BuiltinFunction
	helpers           map[string]*HelperInfo
	rand              *rand.Rand
	exceptObject      Value
	stack             []Value
	frames            []callFrame
	globals           []Value
	openUpvalues      []*Upvalue
	exceptionHandlers []exceptionHandler
	finallyStack      []finallyContext
	randSeed          int64
}

// NewVM creates a new VM with default configuration.
func NewVM() *VM {
	return NewVMWithOutput(nil)
}

// NewVMWithOutput creates a new VM with the specified output writer.
// If output is nil, output operations will be no-ops.
func NewVMWithOutput(output io.Writer) *VM {
	defaultSeed := int64(1)
	vm := &VM{
		stack:             make([]Value, 0, defaultStackCapacity),
		frames:            make([]callFrame, 0, defaultFrameCapacity),
		globals:           make([]Value, 0),
		openUpvalues:      make([]*Upvalue, 0),
		exceptionHandlers: make([]exceptionHandler, 0),
		finallyStack:      make([]finallyContext, 0),
		exceptObject:      NilValue(),
		output:            output,
		builtins:          make(map[string]BuiltinFunction),
		helpers:           make(map[string]*HelperInfo),
		rand:              rand.New(rand.NewSource(defaultSeed)),
		randSeed:          defaultSeed,
	}
	vm.registerBuiltins()
	return vm
}

func (vm *VM) reset() {
	vm.stack = vm.stack[:0]
	vm.frames = vm.frames[:0]
	vm.openUpvalues = vm.openUpvalues[:0]
	vm.exceptionHandlers = vm.exceptionHandlers[:0]
	vm.finallyStack = vm.finallyStack[:0]
	vm.exceptObject = NilValue()
	vm.setGlobal(builtinExceptObjectIndex, vm.exceptObject)

	// Initialize built-in functions as globals
	// The order must match the order in compiler's initBuiltins()
	vm.setGlobal(1, BuiltinValue("PrintLn"))
	vm.setGlobal(2, BuiltinValue("Print"))
	vm.setGlobal(3, BuiltinValue("IntToStr"))
	vm.setGlobal(4, BuiltinValue("FloatToStr"))
	vm.setGlobal(5, BuiltinValue("StrToInt"))
	vm.setGlobal(6, BuiltinValue("StrToFloat"))
	vm.setGlobal(7, BuiltinValue("StrToIntDef"))
	vm.setGlobal(8, BuiltinValue("StrToFloatDef"))
	vm.setGlobal(9, BuiltinValue("Length"))
	vm.setGlobal(10, BuiltinValue("Copy"))
	vm.setGlobal(11, BuiltinValue("SubStr"))
	vm.setGlobal(12, BuiltinValue("SubString"))
	vm.setGlobal(13, BuiltinValue("LeftStr"))
	vm.setGlobal(14, BuiltinValue("RightStr"))
	vm.setGlobal(15, BuiltinValue("MidStr"))
	vm.setGlobal(16, BuiltinValue("StrBeginsWith"))
	vm.setGlobal(17, BuiltinValue("StrEndsWith"))
	vm.setGlobal(18, BuiltinValue("StrContains"))
	vm.setGlobal(19, BuiltinValue("PosEx"))
	vm.setGlobal(20, BuiltinValue("RevPos"))
	vm.setGlobal(21, BuiltinValue("StrFind"))
	vm.setGlobal(22, BuiltinValue("Ord"))
	vm.setGlobal(23, BuiltinValue("Chr"))
	// Type cast functions
	vm.setGlobal(24, BuiltinValue("Integer"))
	vm.setGlobal(25, BuiltinValue("Float"))
	vm.setGlobal(26, BuiltinValue("String"))
	vm.setGlobal(27, BuiltinValue("Boolean"))
	// Math functions (Pi is a constant, handled separately)
	vm.setGlobal(28, BuiltinValue("Sign"))
	vm.setGlobal(29, BuiltinValue("Odd"))
	vm.setGlobal(30, BuiltinValue("Frac"))
	vm.setGlobal(31, BuiltinValue("Int"))
	vm.setGlobal(32, BuiltinValue("Log10"))
	vm.setGlobal(33, BuiltinValue("LogN"))

	// MEDIUM PRIORITY Math Functions
	vm.setGlobal(34, BuiltinValue("Infinity"))
	vm.setGlobal(35, BuiltinValue("NaN"))
	vm.setGlobal(36, BuiltinValue("IsFinite"))
	vm.setGlobal(37, BuiltinValue("IsInfinite"))
	vm.setGlobal(38, BuiltinValue("IntPower"))
	vm.setGlobal(39, BuiltinValue("RandSeed"))
	vm.setGlobal(40, BuiltinValue("RandG"))
	vm.setGlobal(41, BuiltinValue("SetRandSeed"))
	vm.setGlobal(42, BuiltinValue("Randomize"))

	// Advanced Math Functions (Phase 9.23)
	vm.setGlobal(43, BuiltinValue("Factorial"))
	vm.setGlobal(44, BuiltinValue("Gcd"))
	vm.setGlobal(45, BuiltinValue("Lcm"))
	vm.setGlobal(46, BuiltinValue("IsPrime"))
	vm.setGlobal(47, BuiltinValue("LeastFactor"))
	vm.setGlobal(48, BuiltinValue("PopCount"))
	vm.setGlobal(49, BuiltinValue("TestBit"))
}
