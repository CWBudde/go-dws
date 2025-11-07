package bytecode

import (
	"fmt"
	"math"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// Default VM configuration constants.
const (
	defaultStackCapacity = 256
	defaultFrameCapacity = 16
)

// VM executes bytecode chunks produced by the compiler.
type VM struct {
	stack        []Value
	frames       []callFrame
	globals      []Value
	openUpvalues []*Upvalue
}

type callFrame struct {
	chunk   *Chunk
	ip      int
	locals  []Value
	closure *Closure
	self    Value
}

// NewVM creates a new VM with default configuration.
func NewVM() *VM {
	return &VM{
		stack:        make([]Value, 0, defaultStackCapacity),
		frames:       make([]callFrame, 0, defaultFrameCapacity),
		globals:      make([]Value, 0),
		openUpvalues: make([]*Upvalue, 0),
	}
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
			var ret Value = NilValue()
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
		case OpCallMethod, OpCallVirtual:
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
			name, err := vm.constantAsString(frame.chunk, nameIdx, "CALL_METHOD")
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
}

func (vm *VM) getGlobal(index int) Value {
	if index < 0 {
		return NilValue()
	}
	if index < len(vm.globals) {
		val := vm.globals[index]
		if val.Type == 0 && val.Data == nil {
			return NilValue()
		}
		return val
	}
	return NilValue()
}

func (vm *VM) setGlobal(index int, value Value) {
	if index < 0 {
		return
	}
	if index >= len(vm.globals) {
		newGlobals := make([]Value, index+1)
		copy(newGlobals, vm.globals)
		vm.globals = newGlobals
	}
	vm.globals[index] = value
}

// SetGlobal sets a global value at the given index (primarily for tests).
func (vm *VM) SetGlobal(index int, value Value) {
	vm.setGlobal(index, value)
}

// GetGlobal retrieves a global value at the given index (primarily for tests).
func (vm *VM) GetGlobal(index int) Value {
	return vm.getGlobal(index)
}

func (vm *VM) push(v Value) {
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() (Value, error) {
	if len(vm.stack) == 0 {
		return NilValue(), vm.runtimeError("stack underflow")
	}
	v := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return v, nil
}

func (vm *VM) peek() (Value, error) {
	if len(vm.stack) == 0 {
		return NilValue(), vm.runtimeError("stack underflow")
	}
	return vm.stack[len(vm.stack)-1], nil
}

func (vm *VM) binaryIntOp(fn func(a, b int64) int64) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsInt() || !right.IsInt() {
		return vm.typeError("integer operation", "Integer", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	vm.push(IntValue(fn(left.AsInt(), right.AsInt())))
	return nil
}

func (vm *VM) binaryIntOpChecked(fn func(a, b int64) (int64, error)) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsInt() || !right.IsInt() {
		return vm.typeError("integer operation", "Integer", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	result, err := fn(left.AsInt(), right.AsInt())
	if err != nil {
		return err
	}
	vm.push(IntValue(result))
	return nil
}

func (vm *VM) binaryFloatOp(fn func(a, b float64) float64) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsNumber() || !right.IsNumber() {
		return vm.typeError("float operation", "Number", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	vm.push(FloatValue(fn(left.AsFloat(), right.AsFloat())))
	return nil
}

func (vm *VM) binaryFloatOpChecked(fn func(a, b float64) (float64, error)) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}
	if !left.IsNumber() || !right.IsNumber() {
		return vm.typeError("float operation", "Number", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
	}
	result, err := fn(left.AsFloat(), right.AsFloat())
	if err != nil {
		return err
	}
	vm.push(FloatValue(result))
	return nil
}

func (vm *VM) compare(op OpCode) (bool, error) {
	right, err := vm.pop()
	if err != nil {
		return false, err
	}
	left, err := vm.pop()
	if err != nil {
		return false, err
	}

	switch {
	case left.IsNumber() && right.IsNumber():
		l := left.AsFloat()
		r := right.AsFloat()
		switch op {
		case OpGreater:
			return l > r, nil
		case OpGreaterEqual:
			return l >= r, nil
		case OpLess:
			return l < r, nil
		case OpLessEqual:
			return l <= r, nil
		}
	case left.IsString() && right.IsString():
		l := left.AsString()
		r := right.AsString()
		switch op {
		case OpGreater:
			return l > r, nil
		case OpGreaterEqual:
			return l >= r, nil
		case OpLess:
			return l < r, nil
		case OpLessEqual:
			return l <= r, nil
		}
	default:
		return false, vm.runtimeError("comparison of incompatible types %s and %s", left.Type.String(), right.Type.String())
	}

	return false, vm.runtimeError("unsupported comparison opcode %v", op)
}

func (vm *VM) valuesEqual(a, b Value) bool {
	if a.Type == b.Type {
		switch a.Type {
		case ValueNil:
			return true
		case ValueBool:
			return a.AsBool() == b.AsBool()
		case ValueInt:
			return a.AsInt() == b.AsInt()
		case ValueFloat:
			return a.AsFloat() == b.AsFloat()
		case ValueString:
			return a.AsString() == b.AsString()
		case ValueFunction:
			return a.AsFunction() == b.AsFunction()
		case ValueClosure:
			return a.AsClosure() == b.AsClosure()
		case ValueObject:
			return a.AsObject() == b.AsObject()
		default:
			return false
		}
	}

	if a.IsNumber() && b.IsNumber() {
		return a.AsFloat() == b.AsFloat()
	}

	return false
}

func (vm *VM) constantAsString(chunk *Chunk, index int, context string) (string, error) {
	if chunk == nil {
		return "", vm.runtimeError("%s without chunk", context)
	}
	if index < 0 || index >= len(chunk.Constants) {
		return "", vm.runtimeError("%s constant index %d out of range", context, index)
	}
	val := chunk.Constants[index]
	if !val.IsString() {
		return "", vm.runtimeError("%s constant %d is not a string (got %s)", context, index, val.Type.String())
	}
	return val.AsString(), nil
}

func (vm *VM) runtimeError(format string, args ...interface{}) error {
	msg := fmt.Sprintf("vm: %s", fmt.Sprintf(format, args...))
	trace := vm.buildStackTrace()
	return &RuntimeError{
		Message: msg,
		Trace:   trace,
	}
}

func (vm *VM) typeError(context, expected, actual string) error {
	return vm.runtimeError("%s expects %s but got %s", context, expected, actual)
}

func (vm *VM) invokeMethod(receiver Value, methodName string, args []Value) error {
	if !receiver.IsObject() {
		return vm.typeError("CALL_METHOD", "Object", receiver.Type.String())
	}

	obj := receiver.AsObject()
	if obj == nil {
		return vm.runtimeError("CALL_METHOD on nil object")
	}

	methodValue, ok := obj.GetField(methodName)
	if !ok {
		methodValue, ok = obj.GetProperty(methodName)
		if !ok {
			return vm.runtimeError("method %q not found on object %s", methodName, obj.ClassName)
		}
	}

	return vm.callValueWithSelf(methodValue, args, receiver)
}

func (vm *VM) popArgs(argCount int) ([]Value, error) {
	if argCount < 0 {
		return nil, vm.runtimeError("negative arg count")
	}
	if argCount == 0 {
		return nil, nil
	}
	if len(vm.stack) < argCount {
		return nil, vm.runtimeError("stack underflow when collecting arguments")
	}
	args := make([]Value, argCount)
	copy(args, vm.stack[len(vm.stack)-argCount:])
	vm.stack = vm.stack[:len(vm.stack)-argCount]
	return args, nil
}

func (vm *VM) callValue(callee Value, args []Value) error {
	return vm.callValueWithSelf(callee, args, NilValue())
}

func (vm *VM) callValueWithSelf(callee Value, args []Value, self Value) error {
	switch callee.Type {
	case ValueFunction:
		fn := callee.AsFunction()
		if fn == nil {
			return vm.runtimeError("invalid function value")
		}
		closure := vm.makeClosure(fn)
		return vm.callClosure(closure, args, self)
	case ValueClosure:
		cl := callee.AsClosure()
		if cl == nil {
			return vm.runtimeError("invalid closure value")
		}
		return vm.callClosure(cl, args, self)
	default:
		return vm.runtimeError("attempt to call non-callable value of type %s", callee.Type.String())
	}
}

func (vm *VM) callClosure(closure *Closure, args []Value, self Value) error {
	if closure == nil || closure.Function == nil {
		return vm.runtimeError("invalid closure")
	}
	fn := closure.Function
	if fn.Chunk == nil {
		return vm.runtimeError("function %s has no chunk", fn.Name)
	}
	if fn.Arity >= 0 && len(args) != fn.Arity {
		return vm.runtimeError("function %s expected %d arguments but got %d", fn.Name, fn.Arity, len(args))
	}

	localCount := fn.Chunk.LocalCount
	if localCount < len(args) {
		localCount = len(args)
	}
	locals := make([]Value, localCount)
	copy(locals, args)

	vm.frames = append(vm.frames, callFrame{
		chunk:   fn.Chunk,
		ip:      0,
		locals:  locals,
		closure: closure,
		self:    self,
	})
	return nil
}

func (vm *VM) makeClosure(fn *FunctionObject) *Closure {
	if fn == nil {
		return nil
	}
	upvalueCount := fn.UpvalueCount()
	upvalues := make([]*Upvalue, upvalueCount)
	return &Closure{
		Function: fn,
		Upvalues: upvalues,
	}
}

func (vm *VM) captureUpvalue(frame *callFrame, index int) (*Upvalue, error) {
	if frame == nil {
		return nil, vm.runtimeError("no frame available for upvalue capture")
	}
	if index < 0 || index >= len(frame.locals) {
		return nil, vm.runtimeError("upvalue local index %d out of range", index)
	}
	location := &frame.locals[index]
	for _, uv := range vm.openUpvalues {
		if uv.location == location {
			return uv, nil
		}
	}
	uv := newOpenUpvalue(location)
	vm.openUpvalues = append(vm.openUpvalues, uv)
	return uv, nil
}

func (vm *VM) closeUpvaluesForFrame(frame *callFrame) {
	if frame == nil || len(frame.locals) == 0 || len(vm.openUpvalues) == 0 {
		return
	}

	locals := frame.locals
	newOpen := vm.openUpvalues[:0]
	for _, uv := range vm.openUpvalues {
		if uv.location == nil {
			continue
		}
		closed := false
		for i := range locals {
			if uv.location == &locals[i] {
				uv.close()
				closed = true
				break
			}
		}
		if !closed {
			newOpen = append(newOpen, uv)
		}
	}
	vm.openUpvalues = newOpen
}

func (vm *VM) buildStackTrace() errors.StackTrace {
	st := errors.NewStackTrace()
	if len(vm.frames) == 0 {
		return st
	}

	for _, frame := range vm.frames {
		name := vm.frameName(&frame)
		line := frame.chunk.GetLine(max(frame.ip-1, 0))
		pos := &lexer.Position{
			Line:   line,
			Column: 1,
		}
		sf := errors.NewStackFrame(name, frame.chunk.Name, pos)
		st = append(st, sf)
	}
	return st
}

func (vm *VM) frameName(frame *callFrame) string {
	if frame == nil {
		return "<unknown>"
	}
	if frame.closure != nil && frame.closure.Function != nil && frame.closure.Function.Name != "" {
		return frame.closure.Function.Name
	}
	if frame.chunk != nil && frame.chunk.Name != "" {
		return frame.chunk.Name
	}
	return "<script>"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
