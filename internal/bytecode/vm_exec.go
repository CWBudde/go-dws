package bytecode

import (
	"fmt"
	"math"
)

// Run executes the provided chunk and returns the resulting value.
func (vm *VM) Run(chunk *Chunk) (Value, error) {
	if chunk == nil {
		return NilValue(), fmt.Errorf("vm: nil chunk")
	}

	if err := chunk.Validate(); err != nil {
		return NilValue(), fmt.Errorf("vm: invalid chunk: %w", err)
	}

	vm.reset()

	// Load helper metadata from the chunk
	if chunk.Helpers != nil && len(chunk.Helpers) > 0 {
		vm.helpers = chunk.Helpers
	}

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
			// Support Variant→Boolean implicit conversion
			if (!left.IsBool() && !left.IsVariant()) || (!right.IsBool() && !right.IsVariant()) {
				return NilValue(), vm.typeError("AND", "Boolean or Variant", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(isTruthy(left) && isTruthy(right)))
		case OpOr:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if (!left.IsBool() && !left.IsVariant()) || (!right.IsBool() && !right.IsVariant()) {
				return NilValue(), vm.typeError("OR", "Boolean or Variant", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(isTruthy(left) || isTruthy(right)))
		case OpNot:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if !val.IsBool() && !val.IsVariant() {
				return NilValue(), vm.typeError("NOT", "Boolean or Variant", val.Type.String())
			}
			result := !isTruthy(val)
			vm.push(BoolValue(result))
		case OpXor:
			right, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			left, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if (!left.IsBool() && !left.IsVariant()) || (!right.IsBool() && !right.IsVariant()) {
				return NilValue(), vm.typeError("XOR", "Boolean or Variant", fmt.Sprintf("%s, %s", left.Type.String(), right.Type.String()))
			}
			vm.push(BoolValue(isTruthy(left) != isTruthy(right)))
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
			// Support Variant→Boolean implicit conversion
			if !cond.IsBool() && !cond.IsVariant() {
				return NilValue(), vm.typeError("JUMP_IF_FALSE", "Boolean or Variant", cond.Type.String())
			}
			if !isTruthy(cond) {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfTrue:
			cond, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if !cond.IsBool() && !cond.IsVariant() {
				return NilValue(), vm.typeError("JUMP_IF_TRUE", "Boolean or Variant", cond.Type.String())
			}
			if isTruthy(cond) {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfFalseNoPop:
			cond, err := vm.peek()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if !cond.IsBool() && !cond.IsVariant() {
				return NilValue(), vm.typeError("JUMP_IF_FALSE_NO_POP", "Boolean or Variant", cond.Type.String())
			}
			if !isTruthy(cond) {
				frame.ip += int(inst.SignedB())
			}
		case OpJumpIfTrueNoPop:
			cond, err := vm.peek()
			if err != nil {
				return NilValue(), err
			}
			// Support Variant→Boolean implicit conversion
			if !cond.IsBool() && !cond.IsVariant() {
				return NilValue(), vm.typeError("JUMP_IF_TRUE_NO_POP", "Boolean or Variant", cond.Type.String())
			}
			if isTruthy(cond) {
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
		case OpToBool:
			val, err := vm.pop()
			if err != nil {
				return NilValue(), err
			}
			vm.push(BoolValue(variantToBool(val)))
		default:
			return NilValue(), vm.runtimeError("unsupported opcode %v", inst.OpCode())
		}
	}

	return NilValue(), nil
}
