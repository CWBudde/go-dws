package bytecode

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
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
	exceptObject      Value
	output            io.Writer
	builtins          map[string]BuiltinFunction
	stack             []Value
	frames            []callFrame
	globals           []Value
	openUpvalues      []*Upvalue
	exceptionHandlers []exceptionHandler
	finallyStack      []finallyContext
	rand              *rand.Rand
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
		rand:              rand.New(rand.NewSource(defaultSeed)),
		randSeed:          defaultSeed,
	}
	vm.registerBuiltins()
	return vm
}

// Run executes the provided chunk and returns the resulting value.
func (vm *VM) Run(chunk *Chunk) (Value, error) {
	if chunk == nil {
		return NilValue(), fmt.Errorf("vm: nil chunk")
	}

	if err := chunk.Validate(); err != nil {
		return NilValue(), fmt.Errorf("vm: invalid chunk: %w", err)
	}

	vm.reset()

	locals := make([]Value, chunk.LocalCount)
	vm.frames = append(vm.frames, callFrame{
		chunk:   chunk,
		ip:      0,
		locals:  locals,
		closure: nil,
		self:    NilValue(),
	})

	for len(vm.frames) > 0 {
		frame := &vm.frames[len(vm.frames)-1]

		if frame.ip >= len(frame.chunk.Code) {
			// Treat reaching end of instructions as implicit return nil.
			vm.closeUpvaluesForFrame(frame)
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return NilValue(), nil
			}
			continue
		}

		inst := frame.chunk.Code[frame.ip]
		frame.ip++

		switch inst.OpCode() {
		case OpLoadConst:
			constIdx := int(inst.B())
			if constIdx >= len(frame.chunk.Constants) {
				return NilValue(), vm.runtimeError("constant index %d out of range", constIdx)
			}
			vm.push(frame.chunk.Constants[constIdx])
		case OpLoadConst0:
			if len(frame.chunk.Constants) == 0 {
				return NilValue(), vm.runtimeError("LOAD_CONST_0 without constants")
			}
			vm.push(frame.chunk.Constants[0])
		case OpLoadConst1:
			if len(frame.chunk.Constants) < 2 {
				return NilValue(), vm.runtimeError("LOAD_CONST_1 requires two constants")
			}
			vm.push(frame.chunk.Constants[1])
		case OpLoadNil:
			vm.push(NilValue())
		case OpLoadTrue:
			vm.push(BoolValue(true))
		case OpLoadFalse:
			vm.push(BoolValue(false))
		case OpGetSelf:
			vm.push(frame.self)
		case OpLoadLocal:
			idx := int(inst.B())
			if idx >= len(frame.locals) {
				return NilValue(), vm.runtimeError("LOAD_LOCAL index %d out of range", idx)
			}
			vm.push(frame.locals[idx])
		case OpStoreLocal:
			idx := int(inst.B())
			if idx >= len(frame.locals) {
				return NilValue(), vm.runtimeError("STORE_LOCAL index %d out of range", idx)
			}
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			frame.locals[idx] = val
		case OpLoadGlobal:
			idx := int(inst.B())
			vm.push(vm.getGlobal(idx))
		case OpStoreGlobal:
			idx := int(inst.B())
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			vm.setGlobal(idx, val)
		case OpLoadUpvalue:
			if frame.closure == nil {
				return NilValue(), vm.runtimeError("LOAD_UPVALUE outside closure")
			}
			idx := int(inst.B())
			if idx < 0 || idx >= len(frame.closure.Upvalues) {
				return NilValue(), vm.runtimeError("LOAD_UPVALUE index %d out of range", idx)
			}
			vm.push(frame.closure.Upvalues[idx].get())
		case OpStoreUpvalue:
			if frame.closure == nil {
				return NilValue(), vm.runtimeError("STORE_UPVALUE outside closure")
			}
			idx := int(inst.B())
			if idx < 0 || idx >= len(frame.closure.Upvalues) {
				return NilValue(), vm.runtimeError("STORE_UPVALUE index %d out of range", idx)
			}
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			frame.closure.Upvalues[idx].set(val)
		case OpPop:
			if _, err := vm.pop(); err != nil {
				return NilValue(), err
			}
		case OpAddInt:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a + b }); err != nil {
				return NilValue(), err
			}
		case OpSubInt:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a - b }); err != nil {
				return NilValue(), err
			}
		case OpMulInt:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a * b }); err != nil {
				return NilValue(), err
			}
		case OpDivInt:
			if err := vm.binaryIntOpChecked(func(a, b int64) (int64, error) {
				if b == 0 {
					return 0, vm.runtimeError("integer division by zero")
				}
				return a / b, nil
			}); err != nil {
				return NilValue(), err
			}
		case OpModInt:
			if err := vm.binaryIntOpChecked(func(a, b int64) (int64, error) {
				if b == 0 {
					return 0, vm.runtimeError("integer modulo by zero")
				}
				return a % b, nil
			}); err != nil {
				return NilValue(), err
			}
		case OpNegateInt:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsInt() {
				return NilValue(), vm.typeError("NEGATE_INT", "Integer", val.Type.String())
			}
			vm.push(IntValue(-val.AsInt()))
		case OpIncInt:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsInt() {
				return NilValue(), vm.typeError("INC_INT", "Integer", val.Type.String())
			}
			vm.push(IntValue(val.AsInt() + 1))
		case OpDecInt:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsInt() {
				return NilValue(), vm.typeError("DEC_INT", "Integer", val.Type.String())
			}
			vm.push(IntValue(val.AsInt() - 1))
		case OpBitAnd:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a & b }); err != nil {
				return NilValue(), err
			}
		case OpBitOr:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a | b }); err != nil {
				return NilValue(), err
			}
		case OpBitXor:
			if err := vm.binaryIntOp(func(a, b int64) int64 { return a ^ b }); err != nil {
				return NilValue(), err
			}
		case OpBitNot:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsInt() {
				return NilValue(), vm.typeError("BIT_NOT", "Integer", val.Type.String())
			}
			vm.push(IntValue(^val.AsInt()))
		case OpAddFloat:
			if err := vm.binaryFloatOp(func(a, b float64) float64 { return a + b }); err != nil {
				return NilValue(), err
			}
		case OpSubFloat:
			if err := vm.binaryFloatOp(func(a, b float64) float64 { return a - b }); err != nil {
				return NilValue(), err
			}
		case OpMulFloat:
			if err := vm.binaryFloatOp(func(a, b float64) float64 { return a * b }); err != nil {
				return NilValue(), err
			}
		case OpDivFloat:
			if err := vm.binaryFloatOpChecked(func(a, b float64) (float64, error) {
				if b == 0 {
					return 0, vm.runtimeError("float division by zero")
				}
				return a / b, nil
			}); err != nil {
				return NilValue(), err
			}
		case OpNegateFloat:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsNumber() {
				return NilValue(), vm.typeError("NEGATE_FLOAT", "Float", val.Type.String())
			}
			vm.push(FloatValue(-val.AsFloat()))
		case OpStringConcat:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsString() || !right.IsString() {
				return NilValue(), vm.typeError("STRING_CONCAT", "String", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(StringValue(left.AsString() + right.AsString()))
		case OpTry:
			tryInfo, ok := frame.chunk.TryInfoAt(frame.ip - 1)
			if !ok {
				return NilValue(), vm.runtimeError("TRY instruction missing metadata")
			}
			handler := exceptionHandler{
				info:             tryInfo,
				frameIndex:       len(vm.frames) - 1,
				stackDepth:       len(vm.stack),
				exceptionValue:   NilValue(),
				exceptionActive:  false,
				exceptionHandled: true,
				catchCompleted:   !tryInfo.HasCatch,
				prevExceptObject: vm.exceptObject,
			}
			vm.exceptionHandlers = append(vm.exceptionHandlers, handler)
		case OpCatch:
			if len(vm.exceptionHandlers) == 0 {
				return NilValue(), vm.runtimeError("CATCH without active handler")
			}
			handler := &vm.exceptionHandlers[len(vm.exceptionHandlers)-1]
			if !handler.info.HasCatch {
				return NilValue(), vm.runtimeError("CATCH executed but handler has no catch block")
			}
			if !handler.exceptionActive {
				return NilValue(), vm.runtimeError("CATCH executed without active exception")
			}
			handler.exceptionHandled = true
			vm.setExceptObject(handler.exceptionValue)
			vm.push(handler.exceptionValue)
		case OpFinally:
			if inst.A() == 0 {
				ctx, err := vm.beginFinally()
				if err != nil {
					return NilValue(), err
				}
				vm.finallyStack = append(vm.finallyStack, ctx)
			} else {
				if len(vm.finallyStack) == 0 {
					return NilValue(), vm.runtimeError("FINALLY exit without context")
				}
				ctx := vm.finallyStack[len(vm.finallyStack)-1]
				vm.finallyStack = vm.finallyStack[:len(vm.finallyStack)-1]
				vm.setExceptObject(ctx.prevExceptObject)
				if ctx.exceptionActive && !ctx.exceptionHandled {
					if err := vm.raiseException(ctx.exceptionValue); err != nil {
						return NilValue(), err
					}
					continue
				}
			}
		case OpThrow:
			exc, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if exc.IsNil() {
				return NilValue(), vm.runtimeError("cannot raise nil exception")
			}
			vm.markTopHandlerUnhandled()
			if err := vm.raiseException(exc); err != nil {
				return NilValue(), err
			}
			continue
		case OpNewArray:
			elementCount := int(inst.B())
			if elementCount < 0 {
				return NilValue(), vm.runtimeError("NEW_ARRAY negative element count %d", elementCount)
			}
			if len(vm.stack) < elementCount {
				return NilValue(), vm.runtimeError("NEW_ARRAY requires %d values on stack", elementCount)
			}
			elements := make([]Value, elementCount)
			for i := elementCount - 1; i >= 0; i-- {
				val, err := vm.pop()
				if err != nil {
					return NilValue(), err
				}
				elements[i] = val
			}
			vm.push(ArrayValue(NewArrayInstance(elements)))
		case OpNewArraySized:
			sizeVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			size, err := vm.requireInt(sizeVal, "NEW_ARRAY_SIZED")
			if err != nil {
				return NilValue(), err
			}
			if size < 0 {
				return NilValue(), vm.runtimeError("NEW_ARRAY_SIZED negative size %d", size)
			}
			vm.push(ArrayValue(NewArrayInstanceWithLength(size)))
		case OpArrayLength:
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			arr, err := vm.requireArray(arrVal, "ARRAY_LENGTH")
			if err != nil {
				return NilValue(), err
			}
			vm.push(IntValue(int64(arr.Length())))
		case OpArrayGet:
			indexVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			idx, err := vm.requireInt(indexVal, "ARRAY_GET index")
			if err != nil {
				return NilValue(), err
			}
			arr, err := vm.requireArray(arrVal, "ARRAY_GET")
			if err != nil {
				return NilValue(), err
			}
			value, ok := arr.Get(idx)
			if !ok {
				return NilValue(), vm.runtimeError("ARRAY_GET index %d out of range (length %d)", idx, arr.Length())
			}
			vm.push(value)
		case OpArraySet:
			value, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			indexVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			idx, err := vm.requireInt(indexVal, "ARRAY_SET index")
			if err != nil {
				return NilValue(), err
			}
			arr, err := vm.requireArray(arrVal, "ARRAY_SET")
			if err != nil {
				return NilValue(), err
			}
			if !arr.Set(idx, value) {
				return NilValue(), vm.runtimeError("ARRAY_SET index %d out of range (length %d)", idx, arr.Length())
			}
		case OpArraySetLength:
			newLenVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			newLen, err := vm.requireInt(newLenVal, "ARRAY_SET_LENGTH")
			if err != nil {
				return NilValue(), err
			}
			if newLen < 0 {
				return NilValue(), vm.runtimeError("ARRAY_SET_LENGTH negative size %d", newLen)
			}
			arr, err := vm.requireArray(arrVal, "ARRAY_SET_LENGTH")
			if err != nil {
				return NilValue(), err
			}
			arr.Resize(newLen)
		case OpArrayHigh:
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			arr, err := vm.requireArray(arrVal, "ARRAY_HIGH")
			if err != nil {
				return NilValue(), err
			}
			if arr.Length() == 0 {
				vm.push(IntValue(-1))
			} else {
				vm.push(IntValue(int64(arr.Length() - 1)))
			}
		case OpArrayLow:
			arrVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if _, err := vm.requireArray(arrVal, "ARRAY_LOW"); err != nil {
				return NilValue(), err
			}
			vm.push(IntValue(0))
		case OpNewObject:
			classIdx := int(inst.B())
			className, err := vm.constantAsString(frame.chunk, classIdx, "NEW_OBJECT")
			if err != nil {
				return NilValue(), err
			}
			vm.push(ObjectValue(NewObjectInstance(className)))
		case OpGetField:
			fieldIdx := int(inst.B())
			name, err := vm.constantAsString(frame.chunk, fieldIdx, "GET_FIELD")
			if err != nil {
				return NilValue(), err
			}
			objVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !objVal.IsObject() {
				return NilValue(), vm.typeError("GET_FIELD", "Object", objVal.Type.String())
			}
			obj := objVal.AsObject()
			val, ok := obj.GetField(name)
			if !ok {
				val = NilValue()
			}
			vm.push(val)
		case OpSetField:
			fieldIdx := int(inst.B())
			name, err := vm.constantAsString(frame.chunk, fieldIdx, "SET_FIELD")
			if err != nil {
				return NilValue(), err
			}
			value, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			objVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !objVal.IsObject() {
				return NilValue(), vm.typeError("SET_FIELD", "Object", objVal.Type.String())
			}
			obj := objVal.AsObject()
			obj.SetField(name, value)
		case OpGetProperty:
			propIdx := int(inst.B())
			name, err := vm.constantAsString(frame.chunk, propIdx, "GET_PROPERTY")
			if err != nil {
				return NilValue(), err
			}
			objVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !objVal.IsObject() {
				return NilValue(), vm.typeError("GET_PROPERTY", "Object", objVal.Type.String())
			}
			obj := objVal.AsObject()
			val, ok := obj.GetProperty(name)
			if !ok {
				val = NilValue()
			}
			vm.push(val)
		case OpSetProperty:
			propIdx := int(inst.B())
			name, err := vm.constantAsString(frame.chunk, propIdx, "SET_PROPERTY")
			if err != nil {
				return NilValue(), err
			}
			value, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			objVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !objVal.IsObject() {
				return NilValue(), vm.typeError("SET_PROPERTY", "Object", objVal.Type.String())
			}
			obj := objVal.AsObject()
			obj.SetProperty(name, value)
		case OpGetClass:
			objVal, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !objVal.IsObject() {
				return NilValue(), vm.typeError("GET_CLASS", "Object", objVal.Type.String())
			}
			obj := objVal.AsObject()
			className := ""
			if obj != nil {
				className = obj.ClassName
			}
			vm.push(StringValue(className))
		case OpEqual:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			vm.push(BoolValue(vm.valuesEqual(left, right)))
		case OpNotEqual:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			vm.push(BoolValue(!vm.valuesEqual(left, right)))
		case OpGreater, OpGreaterEqual, OpLess, OpLessEqual:
			cmp, err := vm.compare(inst.OpCode())
			if err != nil {
				return NilValue(), err
			}
			vm.push(BoolValue(cmp))
		case OpCompareInt:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsInt() || !right.IsInt() {
				return NilValue(), vm.typeError("COMPARE_INT", "Integer", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			l := left.AsInt()
			r := right.AsInt()
			var result int64
			switch {
			case l < r:
				result = -1
			case l > r:
				result = 1
			default:
				result = 0
			}
			vm.push(IntValue(result))
		case OpCompareFloat:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsNumber() || !right.IsNumber() {
				return NilValue(), vm.typeError("COMPARE_FLOAT", "Number", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			l := left.AsFloat()
			r := right.AsFloat()
			var result int64
			switch {
			case l < r:
				result = -1
			case l > r:
				result = 1
			default:
				result = 0
			}
			vm.push(IntValue(result))
		case OpAnd:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsBool() || !right.IsBool() {
				return NilValue(), vm.typeError("AND", "Boolean", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(left.AsBool() && right.AsBool()))
		case OpOr:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsBool() || !right.IsBool() {
				return NilValue(), vm.typeError("OR", "Boolean", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(left.AsBool() || right.AsBool()))
		case OpNot:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsBool() {
				return NilValue(), vm.typeError("NOT", "Boolean", val.Type.String())
			}
			vm.push(BoolValue(!val.AsBool()))
		case OpXor:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !left.IsBool() || !right.IsBool() {
				return NilValue(), vm.typeError("XOR", "Boolean", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(left.AsBool() != right.AsBool()))
		case OpIsFalsey:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			// Check if value is falsey (0, 0.0, "", false, nil, empty array)
			isFalsey := false
			switch val.Type {
			case ValueInt:
				isFalsey = val.AsInt() == 0
			case ValueFloat:
				isFalsey = val.AsFloat() == 0.0
			case ValueString:
				isFalsey = val.AsString() == ""
			case ValueBool:
				isFalsey = !val.AsBool()
			case ValueNil:
				isFalsey = true
			case ValueArray:
				if arr := val.AsArray(); arr != nil {
					isFalsey = len(arr.elements) == 0
				} else {
					isFalsey = true
				}
			default:
				// Other types (objects, functions, etc.) are truthy if non-nil
				isFalsey = false
			}
			vm.push(BoolValue(isFalsey))
		case OpJump:
			frame.ip += int(inst.SignedB())
		case OpJumpIfFalse:
			cond, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !cond.IsBool() {
				return NilValue(), vm.typeError("JUMP_IF_FALSE", "Boolean", cond.Type.String())
			}
			if !cond.AsBool() {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfTrue:
			cond, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !cond.IsBool() {
				return NilValue(), vm.typeError("JUMP_IF_TRUE", "Boolean", cond.Type.String())
			}
			if cond.AsBool() {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfFalseNoPop:
			cond, err := vm.peek()
			if err != nil {
				return NilValue(), err
			}
			if !cond.IsBool() {
				return NilValue(), vm.typeError("JUMP_IF_FALSE_NO_POP", "Boolean", cond.Type.String())
			}
			if !cond.AsBool() {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfTrueNoPop:
			cond, err := vm.peek()
			if err != nil {
				return NilValue(), err
			}
			if !cond.IsBool() {
				return NilValue(), vm.typeError("JUMP_IF_TRUE_NO_POP", "Boolean", cond.Type.String())
			}
			if cond.AsBool() {
				frame.ip += int(inst.SignedB())
			}
		case OpLoop:
			frame.ip += int(inst.SignedB())
		case OpReturn:
			var ret = NilValue()
			if inst.A() != 0 {
				val, err := vm.pop()
				if err != nil {
					return NilValue(), err
				}
				ret = val
			}
			vm.closeUpvaluesForFrame(frame)
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return ret, nil
			}
			vm.push(ret)
		case OpHalt:
			if len(vm.stack) == 0 {
				return NilValue(), nil
			}
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			return val, nil
		case OpCall:
			argCount := int(inst.A())
			args, err := vm.popArgs(argCount)
			if err != nil {
				return NilValue(), err
			}
			funcIdx := int(inst.B())
			if funcIdx >= len(frame.chunk.Constants) {
				return NilValue(), vm.runtimeError("CALL constant index %d out of range", funcIdx)
			}
			callee := frame.chunk.Constants[funcIdx]
			if err := vm.callValue(callee, args); err != nil {
				return NilValue(), err
			}
			continue
		case OpCallIndirect:
			argCount := int(inst.A())
			args, err := vm.popArgs(argCount)
			if err != nil {
				return NilValue(), err
			}
			callee, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if err := vm.callValue(callee, args); err != nil {
				return NilValue(), err
			}
			continue
		case OpCallMethod, OpCallVirtual, OpInvoke:
			argCount := int(inst.A())
			args, err := vm.popArgs(argCount)
			if err != nil {
				return NilValue(), err
			}
			receiver, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			nameIdx := int(inst.B())
			ctx := "CALL_METHOD"
			if inst.OpCode() == OpInvoke {
				ctx = "INVOKE"
			}
			name, err := vm.constantAsString(frame.chunk, nameIdx, ctx)
			if err != nil {
				return NilValue(), err
			}
			if err := vm.invokeMethod(receiver, name, args); err != nil {
				return NilValue(), err
			}
			continue
		case OpCallBuiltin:
			return NilValue(), vm.runtimeError("call opcode %v not implemented", inst.OpCode())
		case OpClosure:
			upvalueCount := int(inst.A())
			funcIdx := int(inst.B())
			if funcIdx >= len(frame.chunk.Constants) {
				return NilValue(), vm.runtimeError("CLOSURE constant index %d out of range", funcIdx)
			}
			fnValue := frame.chunk.Constants[funcIdx]
			fn := fnValue.AsFunction()
			if fn == nil {
				return NilValue(), vm.runtimeError("CLOSURE constant %d is not a function", funcIdx)
			}
			if fn.UpvalueCount() != upvalueCount {
				return NilValue(), vm.runtimeError("CLOSURE expected %d upvalues but function declares %d", upvalueCount, fn.UpvalueCount())
			}
			closure := vm.makeClosure(fn)
			for i := 0; i < upvalueCount; i++ {
				def := fn.UpvalueDefs[i]
				if def.IsLocal {
					uv, err := vm.captureUpvalue(frame, def.Index)
					if err != nil {
						return NilValue(), err
					}
					closure.Upvalues[i] = uv
				} else {
					if frame.closure == nil {
						return NilValue(), vm.runtimeError("no parent closure available for upvalue %d", i)
					}
					if def.Index < 0 || def.Index >= len(frame.closure.Upvalues) {
						return NilValue(), vm.runtimeError("upvalue index %d out of range in parent closure", def.Index)
					}
					closure.Upvalues[i] = frame.closure.Upvalues[def.Index]
				}
			}
			vm.push(ClosureValue(closure))
		case OpIntToFloat:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsInt() {
				return NilValue(), vm.typeError("INT_TO_FLOAT", "Integer", val.Type.String())
			}
			vm.push(FloatValue(float64(val.AsInt())))
		case OpFloatToInt:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			if !val.IsNumber() {
				return NilValue(), vm.typeError("FLOAT_TO_INT", "Number", val.Type.String())
			}
			f := val.AsFloat()
			vm.push(IntValue(int64(math.Trunc(f))))
		default:
			return NilValue(), vm.runtimeError("unsupported opcode %v", inst.OpCode())
		}
	}

	return NilValue(), nil
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
}

// registerBuiltins registers built-in functions available to bytecode programs.
func (vm *VM) registerBuiltins() {
	vm.builtins["PrintLn"] = builtinPrintLn
	vm.builtins["Print"] = builtinPrint
	vm.builtins["IntToStr"] = builtinIntToStr
	vm.builtins["FloatToStr"] = builtinFloatToStr
	vm.builtins["StrToInt"] = builtinStrToInt
	vm.builtins["StrToFloat"] = builtinStrToFloat
	vm.builtins["StrToIntDef"] = builtinStrToIntDef
	vm.builtins["StrToFloatDef"] = builtinStrToFloatDef
	vm.builtins["Length"] = builtinLength
	vm.builtins["Copy"] = builtinCopy
	vm.builtins["SubStr"] = builtinSubStr
	vm.builtins["SubString"] = builtinSubString
	vm.builtins["LeftStr"] = builtinLeftStr
	vm.builtins["RightStr"] = builtinRightStr
	vm.builtins["MidStr"] = builtinMidStr
	vm.builtins["StrBeginsWith"] = builtinStrBeginsWith
	vm.builtins["StrEndsWith"] = builtinStrEndsWith
	vm.builtins["StrContains"] = builtinStrContains
	vm.builtins["PosEx"] = builtinPosEx
	vm.builtins["RevPos"] = builtinRevPos
	vm.builtins["StrFind"] = builtinStrFind
	vm.builtins["StrSplit"] = builtinStrSplit
	vm.builtins["StrJoin"] = builtinStrJoin
	vm.builtins["StrArrayPack"] = builtinStrArrayPack
	vm.builtins["StrBefore"] = builtinStrBefore
	vm.builtins["StrBeforeLast"] = builtinStrBeforeLast
	vm.builtins["StrAfter"] = builtinStrAfter
	vm.builtins["StrAfterLast"] = builtinStrAfterLast
	vm.builtins["StrBetween"] = builtinStrBetween
	vm.builtins["IsDelimiter"] = builtinIsDelimiter
	vm.builtins["LastDelimiter"] = builtinLastDelimiter
	vm.builtins["FindDelimiter"] = builtinFindDelimiter
	vm.builtins["Ord"] = builtinOrd
	vm.builtins["Chr"] = builtinChr
	// Type cast functions
	vm.builtins["Integer"] = builtinInteger
	vm.builtins["Float"] = builtinFloat
	vm.builtins["String"] = builtinString
	vm.builtins["Boolean"] = builtinBoolean
	// Math functions
	// Note: Pi is a constant, not a function, handled by semantic analyzer
	vm.builtins["Sign"] = builtinSign
	vm.builtins["Odd"] = builtinOdd
	vm.builtins["Frac"] = builtinFrac
	vm.builtins["Int"] = builtinInt
	vm.builtins["Log10"] = builtinLog10
	vm.builtins["LogN"] = builtinLogN

	// MEDIUM PRIORITY Math Functions
	vm.builtins["Infinity"] = builtinInfinity
	vm.builtins["NaN"] = builtinNaN
	vm.builtins["IsFinite"] = builtinIsFinite
	vm.builtins["IsInfinite"] = builtinIsInfinite
	vm.builtins["IntPower"] = builtinIntPower
	vm.builtins["RandSeed"] = builtinRandSeed
	vm.builtins["RandG"] = builtinRandG
	vm.builtins["SetRandSeed"] = builtinSetRandSeed
	vm.builtins["Randomize"] = builtinRandomize
}

// Built-in function implementations

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

func builtinIntToStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IntToStr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("IntToStr expects an integer argument")
	}
	return StringValue(fmt.Sprintf("%d", args[0].AsInt())), nil
}

func builtinFloatToStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("FloatToStr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsFloat() {
		return NilValue(), vm.runtimeError("FloatToStr expects a float argument")
	}
	return StringValue(fmt.Sprintf("%g", args[0].AsFloat())), nil
}

func builtinStrToInt(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrToInt expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToInt expects a string argument")
	}
	var val int64
	_, err := fmt.Sscanf(args[0].AsString(), "%d", &val)
	if err != nil {
		return NilValue(), vm.runtimeError("StrToInt: invalid integer string")
	}
	return IntValue(val), nil
}

func builtinStrToFloat(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrToFloat expects 1 argument, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToFloat expects a string argument")
	}
	var val float64
	_, err := fmt.Sscanf(args[0].AsString(), "%f", &val)
	if err != nil {
		return NilValue(), vm.runtimeError("StrToFloat: invalid float string")
	}
	return FloatValue(val), nil
}

func builtinStrToIntDef(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrToIntDef expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToIntDef expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrToIntDef expects an integer as second argument")
	}
	// Try to parse the string as an integer
	s := strings.TrimSpace(args[0].AsString())
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Return default value on error
		return args[1], nil
	}
	return IntValue(val), nil
}

func builtinStrToFloatDef(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrToFloatDef expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrToFloatDef expects a string as first argument")
	}
	if !args[1].IsFloat() && !args[1].IsInt() {
		return NilValue(), vm.runtimeError("StrToFloatDef expects a float as second argument")
	}
	// Try to parse the string as a float
	s := strings.TrimSpace(args[0].AsString())
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Return default value on error (coerce int to float if needed)
		if args[1].IsInt() {
			return FloatValue(float64(args[1].AsInt())), nil
		}
		return args[1], nil
	}
	return FloatValue(val), nil
}

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

func builtinCopy(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("Copy expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("Copy expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("Copy expects an integer as second argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) - 1 // DWScript uses 1-based indexing

	if start < 0 {
		start = 0
	}
	if start >= len(str) {
		return StringValue(""), nil
	}

	length := len(str) - start
	if len(args) == 3 {
		if !args[2].IsInt() {
			return NilValue(), vm.runtimeError("Copy expects an integer as third argument")
		}
		length = int(args[2].AsInt())
	}

	if start+length > len(str) {
		length = len(str) - start
	}

	return StringValue(str[start : start+length]), nil
}

func builtinSubStr(vm *VM, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return NilValue(), vm.runtimeError("SubStr expects 2 or 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("SubStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("SubStr expects an integer as second argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) - 1 // DWScript uses 1-based indexing

	if start < 0 {
		start = 0
	}
	if start >= len(str) {
		return StringValue(""), nil
	}

	// Default length is to end of string (MaxInt in DWScript)
	length := len(str) - start
	if len(args) == 3 {
		if !args[2].IsInt() {
			return NilValue(), vm.runtimeError("SubStr expects an integer as third argument")
		}
		length = int(args[2].AsInt())
	}

	if start+length > len(str) {
		length = len(str) - start
	}

	return StringValue(str[start : start+length]), nil
}

func builtinSubString(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("SubString expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("SubString expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("SubString expects an integer as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("SubString expects an integer as third argument")
	}

	str := args[0].AsString()
	start := int(args[1].AsInt()) // 1-based
	end := int(args[2].AsInt())   // 1-based, inclusive

	// Calculate length from start and end positions
	length := end - start + 1

	// Handle edge cases
	if length <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	startIdx := start - 1

	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(runes) {
		return StringValue(""), nil
	}

	endIdx := startIdx + length
	if endIdx > len(runes) {
		endIdx = len(runes)
	}

	return StringValue(string(runes[startIdx:endIdx])), nil
}

func builtinLeftStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LeftStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("LeftStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("LeftStr expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())

	if count <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	if count > len(runes) {
		count = len(runes)
	}

	return StringValue(string(runes[:count])), nil
}

func builtinRightStr(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("RightStr expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("RightStr expects a string as first argument")
	}
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("RightStr expects an integer as second argument")
	}

	str := args[0].AsString()
	count := int(args[1].AsInt())

	if count <= 0 {
		return StringValue(""), nil
	}

	// Convert to runes for UTF-8 support
	runes := []rune(str)
	strLen := len(runes)

	if count >= strLen {
		return StringValue(str), nil
	}

	start := strLen - count
	return StringValue(string(runes[start:])), nil
}

func builtinMidStr(vm *VM, args []Value) (Value, error) {
	// MidStr is an alias for SubStr
	return builtinSubStr(vm, args)
}

func builtinStrBeginsWith(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBeginsWith expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBeginsWith expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBeginsWith expects a string as second argument")
	}

	str := args[0].AsString()
	prefix := args[1].AsString()

	result := len(prefix) == 0 || (len(str) >= len(prefix) && str[:len(prefix)] == prefix)
	return BoolValue(result), nil
}

func builtinStrEndsWith(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrEndsWith expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrEndsWith expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrEndsWith expects a string as second argument")
	}

	str := args[0].AsString()
	suffix := args[1].AsString()

	result := len(suffix) == 0 || (len(str) >= len(suffix) && str[len(str)-len(suffix):] == suffix)
	return BoolValue(result), nil
}

func builtinStrContains(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrContains expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrContains expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrContains expects a string as second argument")
	}

	str := args[0].AsString()
	substr := args[1].AsString()

	// Empty substring is always contained
	if len(substr) == 0 {
		return BoolValue(true), nil
	}

	// Check if substring exists
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return BoolValue(true), nil
		}
	}

	return BoolValue(false), nil
}

func builtinPosEx(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("PosEx expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("PosEx expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("PosEx expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("PosEx expects an integer as third argument")
	}

	needle := args[0].AsString()
	haystack := args[1].AsString()
	offset := int(args[2].AsInt()) // 1-based

	// Handle invalid offset first
	if offset < 1 {
		return IntValue(0), nil
	}

	// Handle empty needle - returns 0 (not found)
	if len(needle) == 0 {
		return IntValue(0), nil
	}

	// Convert to runes for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Adjust offset to 0-based
	startIdx := offset - 1

	// If offset is beyond the string length, not found
	if startIdx >= len(haystackRunes) {
		return IntValue(0), nil
	}

	// Search for the needle starting from offset
	for i := startIdx; i <= len(haystackRunes)-len(needleRunes); i++ {
		match := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				match = false
				break
			}
		}
		if match {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// Not found
	return IntValue(0), nil
}

func builtinRevPos(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("RevPos expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("RevPos expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("RevPos expects a string as second argument")
	}

	needle := args[0].AsString()
	haystack := args[1].AsString()

	// Handle empty needle - returns length + 1
	if len(needle) == 0 {
		runes := []rune(haystack)
		return IntValue(int64(len(runes) + 1)), nil
	}

	// Convert to runes for UTF-8 support
	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)

	// Search backwards for the last occurrence
	for i := len(haystackRunes) - len(needleRunes); i >= 0; i-- {
		match := true
		for j := 0; j < len(needleRunes); j++ {
			if haystackRunes[i+j] != needleRunes[j] {
				match = false
				break
			}
		}
		if match {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// Not found
	return IntValue(0), nil
}

func builtinStrFind(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("StrFind expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrFind expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrFind expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("StrFind expects an integer as third argument")
	}

	// StrFind(str, substr, fromIndex) maps to PosEx(substr, str, fromIndex)
	// Reorder arguments
	reorderedArgs := []Value{
		args[1], // substr becomes first arg (needle)
		args[0], // str becomes second arg (haystack)
		args[2], // fromIndex stays as third arg (offset)
	}

	return builtinPosEx(vm, reorderedArgs)
}

func builtinStrSplit(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrSplit expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrSplit expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrSplit expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return array with single element (the original string)
	if len(delim) == 0 {
		elements := []Value{StringValue(str)}
		return ArrayValue(NewArrayInstance(elements)), nil
	}

	// Split the string
	parts := strings.Split(str, delim)

	// Convert to array of Value
	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = StringValue(part)
	}

	return ArrayValue(NewArrayInstance(elements)), nil
}

func builtinStrJoin(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrJoin expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsArray() {
		return NilValue(), vm.runtimeError("StrJoin expects an array as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrJoin expects a string as second argument")
	}

	arr := args[0].AsArray()
	delim := args[1].AsString()

	// Convert array elements to strings
	parts := make([]string, len(arr.elements))
	for idx, elem := range arr.elements {
		if !elem.IsString() {
			return NilValue(), vm.runtimeError("StrJoin expects array of strings, got %s at index %d", elem.Type.String(), idx)
		}
		parts[idx] = elem.AsString()
	}

	// Join the strings
	result := strings.Join(parts, delim)
	return StringValue(result), nil
}

func builtinStrArrayPack(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("StrArrayPack expects 1 argument, got %d", len(args))
	}
	if !args[0].IsArray() {
		return NilValue(), vm.runtimeError("StrArrayPack expects an array as argument")
	}

	arr := args[0].AsArray()

	// Filter out empty strings
	var packed []Value
	for _, elem := range arr.elements {
		if !elem.IsString() {
			return NilValue(), vm.runtimeError("StrArrayPack expects array of strings, got %s", elem.Type.String())
		}
		if elem.AsString() != "" {
			packed = append(packed, elem)
		}
	}

	return ArrayValue(NewArrayInstance(packed)), nil
}

func builtinStrBefore(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBefore expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBefore expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBefore expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring before delimiter
	return StringValue(str[:index]), nil
}

func builtinStrBeforeLast(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrBeforeLast expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBeforeLast expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBeforeLast expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring before last delimiter
	return StringValue(str[:index]), nil
}

func builtinStrAfter(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrAfter expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrAfter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrAfter expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the first occurrence of delimiter
	index := strings.Index(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring after delimiter
	return StringValue(str[index+len(delim):]), nil
}

func builtinStrAfterLast(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("StrAfterLast expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrAfterLast expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrAfterLast expects a string as second argument")
	}

	str := args[0].AsString()
	delim := args[1].AsString()

	// Handle empty delimiter - return empty string
	if len(delim) == 0 {
		return StringValue(""), nil
	}

	// Find the last occurrence of delimiter
	index := strings.LastIndex(str, delim)
	if index == -1 {
		// Delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Return substring after last delimiter
	return StringValue(str[index+len(delim):]), nil
}

func builtinStrBetween(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("StrBetween expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as second argument")
	}
	if !args[2].IsString() {
		return NilValue(), vm.runtimeError("StrBetween expects a string as third argument")
	}

	str := args[0].AsString()
	start := args[1].AsString()
	stop := args[2].AsString()

	// Handle empty delimiters - return empty string
	if len(start) == 0 || len(stop) == 0 {
		return StringValue(""), nil
	}

	// Find the first occurrence of start delimiter
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		// Start delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Search for stop delimiter after the start delimiter
	searchFrom := startIdx + len(start)
	if searchFrom >= len(str) {
		// No room for stop delimiter - return empty string
		return StringValue(""), nil
	}

	stopIdx := strings.Index(str[searchFrom:], stop)
	if stopIdx == -1 {
		// Stop delimiter not found - return empty string
		return StringValue(""), nil
	}

	// Adjust stopIdx to be relative to the original string
	stopIdx += searchFrom

	// Return substring between start and stop delimiters
	return StringValue(str[searchFrom:stopIdx]), nil
}

func builtinIsDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("IsDelimiter expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("IsDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("IsDelimiter expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("IsDelimiter expects an integer as third argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()
	index := args[2].AsInt() // 1-based

	// Handle invalid index
	if index < 1 {
		return BoolValue(false), nil
	}

	// Convert to rune-based indexing for UTF-8 support
	strRunes := []rune(str)

	// Check if index is within bounds (1-based)
	if int(index) > len(strRunes) {
		return BoolValue(false), nil
	}

	// Get the character at the specified position (convert to 0-based)
	ch := strRunes[index-1]

	// Check if the character is in the delimiter string
	result := strings.ContainsRune(delims, ch)
	return BoolValue(result), nil
}

func builtinLastDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LastDelimiter expects 2 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("LastDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("LastDelimiter expects a string as second argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Search from the end for any delimiter character
	for i := len(strRunes) - 1; i >= 0; i-- {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// No delimiter found
	return IntValue(0), nil
}

func builtinFindDelimiter(vm *VM, args []Value) (Value, error) {
	if len(args) != 3 {
		return NilValue(), vm.runtimeError("FindDelimiter expects 3 arguments, got %d", len(args))
	}
	if !args[0].IsString() {
		return NilValue(), vm.runtimeError("FindDelimiter expects a string as first argument")
	}
	if !args[1].IsString() {
		return NilValue(), vm.runtimeError("FindDelimiter expects a string as second argument")
	}
	if !args[2].IsInt() {
		return NilValue(), vm.runtimeError("FindDelimiter expects an integer as third argument")
	}

	delims := args[0].AsString()
	str := args[1].AsString()
	startIndex := args[2].AsInt() // 1-based

	// Handle invalid start index
	if startIndex < 1 {
		return IntValue(0), nil
	}

	// Convert to rune-based for UTF-8 support
	strRunes := []rune(str)

	// Adjust to 0-based index
	startIdx := int(startIndex) - 1

	// Check if start index is within bounds
	if startIdx >= len(strRunes) {
		return IntValue(0), nil
	}

	// Search from startIdx for any delimiter character
	for i := startIdx; i < len(strRunes); i++ {
		if strings.ContainsRune(delims, strRunes[i]) {
			// Return 1-based position
			return IntValue(int64(i + 1)), nil
		}
	}

	// No delimiter found
	return IntValue(0), nil
}

func builtinOrd(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Ord expects 1 argument, got %d", len(args))
	}
	arg := args[0]
	if arg.IsString() {
		s := arg.AsString()
		if len(s) == 0 {
			return IntValue(0), nil
		}
		return IntValue(int64(s[0])), nil
	}
	if arg.IsInt() {
		// For enums and other types
		return arg, nil
	}
	return NilValue(), vm.runtimeError("Ord expects a string or integer argument")
}

func builtinChr(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Chr expects 1 argument, got %d", len(args))
	}
	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("Chr expects an integer argument")
	}
	return StringValue(string(rune(args[0].AsInt()))), nil
}

// Type cast built-in functions

func builtinInteger(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Integer expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueInt:
		return arg, nil
	case ValueFloat:
		return IntValue(int64(arg.AsFloat())), nil
	case ValueBool:
		if arg.AsBool() {
			return IntValue(1), nil
		}
		return IntValue(0), nil
	case ValueString:
		var val int64
		_, err := fmt.Sscanf(arg.AsString(), "%d", &val)
		if err != nil {
			return NilValue(), vm.runtimeError("cannot convert string '%s' to Integer", arg.AsString())
		}
		return IntValue(val), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Integer", arg.Type.String())
	}
}

func builtinFloat(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Float expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueFloat:
		return arg, nil
	case ValueInt:
		return FloatValue(float64(arg.AsInt())), nil
	case ValueBool:
		if arg.AsBool() {
			return FloatValue(1.0), nil
		}
		return FloatValue(0.0), nil
	case ValueString:
		var val float64
		_, err := fmt.Sscanf(arg.AsString(), "%f", &val)
		if err != nil {
			return NilValue(), vm.runtimeError("cannot convert string '%s' to Float", arg.AsString())
		}
		return FloatValue(val), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Float", arg.Type.String())
	}
}

func builtinString(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("String expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueString:
		return arg, nil
	case ValueInt:
		return StringValue(fmt.Sprintf("%d", arg.AsInt())), nil
	case ValueFloat:
		return StringValue(fmt.Sprintf("%g", arg.AsFloat())), nil
	case ValueBool:
		if arg.AsBool() {
			return StringValue("True"), nil
		}
		return StringValue("False"), nil
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to String", arg.Type.String())
	}
}

func builtinBoolean(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Boolean expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	switch arg.Type {
	case ValueBool:
		return arg, nil
	case ValueInt:
		return BoolValue(arg.AsInt() != 0), nil
	case ValueFloat:
		return BoolValue(arg.AsFloat() != 0.0), nil
	case ValueString:
		s := strings.ToLower(strings.TrimSpace(arg.AsString()))
		if s == "true" || s == "1" {
			return BoolValue(true), nil
		}
		if s == "false" || s == "0" || s == "" {
			return BoolValue(false), nil
		}
		return NilValue(), vm.runtimeError("cannot convert string '%s' to Boolean", arg.AsString())
	default:
		return NilValue(), vm.runtimeError("cannot cast %s to Boolean", arg.Type.String())
	}
}

// Math Functions

func builtinPi(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Pi expects no arguments, got %d", len(args))
	}
	return FloatValue(math.Pi), nil
}

func builtinSign(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Sign expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Sign expects Float or Integer, got %s", arg.Type.String())
	}

	if floatVal > 0 {
		return IntValue(1), nil
	} else if floatVal < 0 {
		return IntValue(-1), nil
	}
	return IntValue(0), nil
}

func builtinOdd(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Odd expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	if !arg.IsInt() {
		return NilValue(), vm.runtimeError("Odd expects Integer, got %s", arg.Type.String())
	}

	return BoolValue(arg.AsInt()%2 != 0), nil
}

func builtinFrac(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Frac expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Frac expects Float or Integer, got %s", arg.Type.String())
	}

	// Fractional part = x - floor(x)
	_, frac := math.Modf(floatVal)
	return FloatValue(frac), nil
}

func builtinInt(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Int expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Int expects Float or Integer, got %s", arg.Type.String())
	}

	// Int() returns the integer part (truncated towards zero) as a Float
	return FloatValue(math.Trunc(floatVal)), nil
}

func builtinLog10(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("Log10 expects 1 argument, got %d", len(args))
	}
	arg := args[0]

	var floatVal float64
	if arg.IsFloat() {
		floatVal = arg.AsFloat()
	} else if arg.IsInt() {
		floatVal = float64(arg.AsInt())
	} else {
		return NilValue(), vm.runtimeError("Log10 expects Float or Integer, got %s", arg.Type.String())
	}

	if floatVal <= 0 {
		return NilValue(), vm.runtimeError("Log10 argument must be positive, got %f", floatVal)
	}

	return FloatValue(math.Log10(floatVal)), nil
}

func builtinLogN(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("LogN expects 2 arguments, got %d", len(args))
	}

	// First argument (x)
	var xVal float64
	if args[0].IsFloat() {
		xVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		xVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("LogN expects Float or Integer as first argument, got %s", args[0].Type.String())
	}

	// Second argument (base)
	var baseVal float64
	if args[1].IsFloat() {
		baseVal = args[1].AsFloat()
	} else if args[1].IsInt() {
		baseVal = float64(args[1].AsInt())
	} else {
		return NilValue(), vm.runtimeError("LogN expects Float or Integer as second argument, got %s", args[1].Type.String())
	}

	if xVal <= 0 {
		return NilValue(), vm.runtimeError("LogN first argument must be positive, got %f", xVal)
	}
	if baseVal <= 0 || baseVal == 1 {
		return NilValue(), vm.runtimeError("LogN base must be positive and not equal to 1, got %f", baseVal)
	}

	// LogN(x, base) = Log(x) / Log(base)
	return FloatValue(math.Log(xVal) / math.Log(baseVal)), nil
}

// MEDIUM PRIORITY Math Functions

func builtinInfinity(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Infinity expects no arguments, got %d", len(args))
	}
	return FloatValue(math.Inf(1)), nil
}

func builtinNaN(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("NaN expects no arguments, got %d", len(args))
	}
	return FloatValue(math.NaN()), nil
}

func builtinIsFinite(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IsFinite expects 1 argument, got %d", len(args))
	}

	var floatVal float64
	if args[0].IsFloat() {
		floatVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		floatVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IsFinite expects Float or Integer, got %s", args[0].Type.String())
	}

	return BoolValue(!math.IsInf(floatVal, 0) && !math.IsNaN(floatVal)), nil
}

func builtinIsInfinite(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("IsInfinite expects 1 argument, got %d", len(args))
	}

	var floatVal float64
	if args[0].IsFloat() {
		floatVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		floatVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IsInfinite expects Float or Integer, got %s", args[0].Type.String())
	}

	return BoolValue(math.IsInf(floatVal, 0)), nil
}

func builtinIntPower(vm *VM, args []Value) (Value, error) {
	if len(args) != 2 {
		return NilValue(), vm.runtimeError("IntPower expects 2 arguments, got %d", len(args))
	}

	// First argument (base) - Float or Integer
	var baseVal float64
	if args[0].IsFloat() {
		baseVal = args[0].AsFloat()
	} else if args[0].IsInt() {
		baseVal = float64(args[0].AsInt())
	} else {
		return NilValue(), vm.runtimeError("IntPower expects Float or Integer as first argument, got %s", args[0].Type.String())
	}

	// Second argument (exponent) - Integer only
	if !args[1].IsInt() {
		return NilValue(), vm.runtimeError("IntPower expects Integer as second argument, got %s", args[1].Type.String())
	}
	expVal := args[1].AsInt()

	// Calculate power using exponentiation by squaring for integer exponents
	result := 1.0
	base := baseVal
	exp := expVal

	if exp < 0 {
		base = 1.0 / base
		exp = -exp
	}

	for exp > 0 {
		if exp%2 == 1 {
			result *= base
		}
		base *= base
		exp /= 2
	}

	return FloatValue(result), nil
}

func builtinRandSeed(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("RandSeed expects no arguments, got %d", len(args))
	}
	return IntValue(vm.randSeed), nil
}

func builtinRandG(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("RandG expects no arguments, got %d", len(args))
	}

	// Generate Gaussian random number using Box-Muller transform
	u1 := vm.rand.Float64()
	u2 := vm.rand.Float64()

	// Ensure u1 is not zero or near-zero to avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	// Box-Muller transform
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)

	return FloatValue(z0), nil
}

func builtinSetRandSeed(vm *VM, args []Value) (Value, error) {
	if len(args) != 1 {
		return NilValue(), vm.runtimeError("SetRandSeed expects 1 argument, got %d", len(args))
	}

	if !args[0].IsInt() {
		return NilValue(), vm.runtimeError("SetRandSeed expects Integer, got %s", args[0].Type.String())
	}

	seed := args[0].AsInt()
	vm.randSeed = seed
	vm.rand = rand.New(rand.NewSource(seed))

	return NilValue(), nil
}

func builtinRandomize(vm *VM, args []Value) (Value, error) {
	if len(args) != 0 {
		return NilValue(), vm.runtimeError("Randomize expects no arguments, got %d", len(args))
	}

	seed := time.Now().UnixNano()
	vm.randSeed = seed
	vm.rand = rand.New(rand.NewSource(seed))

	return NilValue(), nil
}
