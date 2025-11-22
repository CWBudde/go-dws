package bytecode

import (
	"math"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// callFrame represents a single function call frame in the VM's call stack.
type callFrame struct {
	self    Value
	chunk   *Chunk
	closure *Closure
	locals  []Value
	ip      int
}

// exceptionHandler tracks exception handling state for try/catch/finally blocks.
type exceptionHandler struct {
	exceptionValue   Value
	prevExceptObject Value
	info             TryInfo
	frameIndex       int
	stackDepth       int
	exceptionActive  bool
	exceptionHandled bool
	catchCompleted   bool
}

// finallyContext tracks state when executing a finally block.
type finallyContext struct {
	exceptionValue   Value
	prevExceptObject Value
	exceptionActive  bool
	exceptionHandled bool
}

// invokeMethod invokes a method on a receiver value.
// For objects, looks up the method in the object's fields/properties.
// For other types (Integer, String, Float, etc.), looks up helper methods.
func (vm *VM) invokeMethod(receiver Value, methodName string, args []Value) error {
	// Handle both initialized arrays and nil arrays (uninitialized dynamic arrays)
	if receiver.IsArray() || receiver.IsNil() {
		arr := receiver.AsArray()
		// For nil arrays, we need to initialize them for mutation methods
		if arr == nil && (methodName == "Add" || methodName == "add" || methodName == "SetLength" || methodName == "setlength") {
			// Initialize as empty array
			arr = NewArrayInstance(nil)
			// Need to update the receiver's value in the global or local variable
			// This is a limitation - for now, methods on nil arrays will fail
			return vm.runtimeError("cannot call %s on nil array - array must be initialized first", methodName)
		}
		// PR #78: Guard against nil arrays for all helper methods to prevent panics
		if arr == nil {
			return vm.runtimeError("cannot call %s on nil array - array must be initialized first", methodName)
		}
		switch methodName {
		case "Add", "add":
			if len(args) != 1 {
				return vm.runtimeError("Array.Add expects 1 argument, got %d", len(args))
			}
			// Append by resizing and setting the last element
			currentLen := arr.Length()
			arr.Resize(currentLen + 1)
			arr.Set(currentLen, args[0])
			vm.push(NilValue())
			return nil
		case "Delete", "delete":
			if len(args) < 1 || len(args) > 2 {
				return vm.runtimeError("Array.Delete expects 1 or 2 arguments, got %d", len(args))
			}
			if !args[0].IsInt() {
				return vm.runtimeError("Array.Delete index must be Integer")
			}
			index := int(args[0].AsInt())
			count := 1
			if len(args) == 2 {
				if !args[1].IsInt() {
					return vm.runtimeError("Array.Delete count must be Integer")
				}
				count = int(args[1].AsInt())
			}

			arrayLen := arr.Length()
			if index < 0 || index >= arrayLen {
				return vm.runtimeError("Array.Delete index %d out of bounds (0..%d)", index, arrayLen-1)
			}
			if count < 0 {
				return vm.runtimeError("Array.Delete count must be non-negative, got %d", count)
			}

			endIndex := index + count
			if endIndex > arrayLen {
				endIndex = arrayLen
			}

			// Delete elements by creating new slice
			newElements := make([]Value, 0, arrayLen-(endIndex-index))
			for i := 0; i < index; i++ {
				val, _ := arr.Get(i)
				newElements = append(newElements, val)
			}
			for i := endIndex; i < arrayLen; i++ {
				val, _ := arr.Get(i)
				newElements = append(newElements, val)
			}

			arr.Resize(len(newElements))
			for i, val := range newElements {
				arr.Set(i, val)
			}
			vm.push(NilValue())
			return nil
		case "IndexOf", "indexof":
			if len(args) < 1 || len(args) > 2 {
				return vm.runtimeError("Array.IndexOf expects 1 or 2 arguments, got %d", len(args))
			}
			valueToFind := args[0]
			startIndex := 0
			if len(args) == 2 {
				if !args[1].IsInt() {
					return vm.runtimeError("Array.IndexOf startIndex must be Integer")
				}
				startIndex = int(args[1].AsInt())
			}

			arrayLen := arr.Length()
			// PR #78: Allow startIndex == arrayLen (searches 0 elements, returns -1)
			if startIndex < 0 || startIndex > arrayLen {
				vm.push(IntValue(-1))
				return nil
			}

			found := false
			for i := startIndex; i < arrayLen; i++ {
				elem, _ := arr.Get(i)
				if vm.valuesEqual(elem, valueToFind) {
					vm.push(IntValue(int64(i)))
					found = true
					break
				}
			}
			if !found {
				vm.push(IntValue(-1))
			}
			return nil
		case "SetLength", "setlength":
			if len(args) != 1 {
				return vm.runtimeError("Array.SetLength expects 1 argument, got %d", len(args))
			}
			if !args[0].IsInt() {
				return vm.runtimeError("Array.SetLength length must be Integer")
			}
			newLength := int(args[0].AsInt())
			if newLength < 0 {
				return vm.runtimeError("Array.SetLength length must be non-negative, got %d", newLength)
			}
			arr.Resize(newLength)
			vm.push(NilValue())
			return nil
		}
		// Fall through to generic helper method handling if not a built-in
	}

	// callStringHelper is a helper function to reduce code duplication in string helper methods.
	// It calls a builtin function with the receiver and args, handles errors, and pushes the result.
	callStringHelper := func(builtin func(*VM, []Value) (Value, error), args []Value) error {
		result, err := builtin(vm, append([]Value{receiver}, args...))
		if err != nil {
			return err
		}
		vm.push(result)
		return nil
	}

	// Handle String helper methods
	if receiver.IsString() {
		str := receiver.AsString()
		methodNameLower := strings.ToLower(methodName)

		switch methodNameLower {
		case "toupper", "uppercase":
			if len(args) != 0 {
				return vm.runtimeError("String.ToUpper expects 0 arguments, got %d", len(args))
			}
			return callStringHelper(builtinUpperCase, args)

		case "tolower", "lowercase":
			if len(args) != 0 {
				return vm.runtimeError("String.ToLower expects 0 arguments, got %d", len(args))
			}
			return callStringHelper(builtinLowerCase, args)

		// TODO: Implement Trim builtin in VM
		// case "trim":

		case "tointeger":
			if len(args) != 0 {
				return vm.runtimeError("String.ToInteger expects 0 arguments, got %d", len(args))
			}
			return callStringHelper(builtinStrToInt, args)

		case "tofloat":
			if len(args) != 0 {
				return vm.runtimeError("String.ToFloat expects 0 arguments, got %d", len(args))
			}
			return callStringHelper(builtinStrToFloat, args)

		case "tostring":
			if len(args) != 0 {
				return vm.runtimeError("String.ToString expects 0 arguments, got %d", len(args))
			}
			vm.push(receiver) // Identity operation for strings
			return nil

		case "startswith":
			if len(args) != 1 {
				return vm.runtimeError("String.StartsWith expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrBeginsWith, args)

		case "endswith":
			if len(args) != 1 {
				return vm.runtimeError("String.EndsWith expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrEndsWith, args)

		case "contains":
			if len(args) != 1 {
				return vm.runtimeError("String.Contains expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrContains, args)

		case "indexof":
			if len(args) != 1 {
				return vm.runtimeError("String.IndexOf expects 1 argument, got %d", len(args))
			}
			// Use PosEx(substr, str, 1) - starts search from position 1
			result, err := builtinPosEx(vm, []Value{args[0], receiver, IntValue(1)})
			if err != nil {
				return err
			}
			vm.push(result)
			return nil

		case "copy":
			if len(args) < 1 || len(args) > 2 {
				return vm.runtimeError("String.Copy expects 1 or 2 arguments, got %d", len(args))
			}
			// Copy(str, start, [length])
			if len(args) == 1 {
				// Copy from start to end (use MaxInt32 as length)
				result, err := builtinCopy(vm, []Value{receiver, args[0], IntValue(math.MaxInt32)})
				if err != nil {
					return err
				}
				vm.push(result)
			} else {
				result, err := builtinCopy(vm, []Value{receiver, args[0], args[1]})
				if err != nil {
					return err
				}
				vm.push(result)
			}
			return nil

		case "before":
			if len(args) != 1 {
				return vm.runtimeError("String.Before expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrBefore, args)

		case "after":
			if len(args) != 1 {
				return vm.runtimeError("String.After expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrAfter, args)

		case "split":
			if len(args) != 1 {
				return vm.runtimeError("String.Split expects 1 argument, got %d", len(args))
			}
			return callStringHelper(builtinStrSplit, args)

		case "length":
			if len(args) != 0 {
				return vm.runtimeError("String.Length expects 0 arguments, got %d", len(args))
			}
			vm.push(IntValue(int64(len(str))))
			return nil
		}
		// Fall through to generic helper method handling for other methods
	}

	// Handle object method calls
	if receiver.IsObject() {
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

	// Handle helper method calls for primitive types (Integer, String, Float, etc.)
	helper := vm.findHelperForValue(receiver)
	if helper == nil {
		return vm.runtimeError("no helper found for type %s, method %q", receiver.Type.String(), methodName)
	}

	// Look up the method in the helper (case-insensitive)
	methodSlot, ok := findHelperMethodCaseInsensitive(helper.Methods, methodName)
	if !ok {
		return vm.runtimeError("method %q not found in helper for type %s", methodName, receiver.Type.String())
	}

	// Load the method function from globals
	if int(methodSlot) >= len(vm.globals) {
		return vm.runtimeError("helper method %q references invalid global slot %d", methodName, methodSlot)
	}

	methodValue := vm.globals[methodSlot]
	if methodValue.IsNil() {
		return vm.runtimeError("helper method %q not initialized", methodName)
	}

	// Call the method with the receiver as Self
	return vm.callValueWithSelf(methodValue, args, receiver)
}

// findHelperForValue finds the first helper that applies to the given value's type.
func (vm *VM) findHelperForValue(value Value) *HelperInfo {
	// Map Value types to type names that helpers might extend
	var typeName string
	switch value.Type {
	case ValueInt:
		typeName = "Integer"
	case ValueFloat:
		typeName = "Float"
	case ValueString:
		typeName = "String"
	case ValueBool:
		typeName = "Boolean"
	case ValueArray:
		typeName = "Array"
	default:
		return nil
	}

	// Look for helpers that extend this type (case-insensitive)
	for _, helper := range vm.helpers {
		if equalsCaseInsensitive(helper.TargetType, typeName) {
			return helper
		}
	}

	return nil
}

// findHelperMethodCaseInsensitive searches for a method in the helper's Methods map.
func findHelperMethodCaseInsensitive(methods map[string]uint16, methodName string) (uint16, bool) {
	for key, slot := range methods {
		if equalsCaseInsensitive(key, methodName) {
			return slot, true
		}
	}
	return 0, false
}

// equalsCaseInsensitive compares two strings case-insensitively.
func equalsCaseInsensitive(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		// Convert to lowercase
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// popArgs pops the specified number of arguments from the stack and returns them as a slice.
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

// callValue calls a callable value (function, closure, or builtin) with the given arguments.
func (vm *VM) callValue(callee Value, args []Value) error {
	return vm.callValueWithSelf(callee, args, NilValue())
}

// callValueWithSelf calls a callable value with a specific self/receiver value.
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
	case ValueBuiltin:
		name := callee.AsBuiltin()
		if name == "" {
			return vm.runtimeError("invalid built-in function")
		}
		builtinFunc, ok := vm.builtins[name]
		if !ok {
			return vm.runtimeError("built-in function %q not found", name)
		}
		result, err := builtinFunc(vm, args)
		if err != nil {
			return err
		}
		vm.push(result)
		return nil
	default:
		return vm.runtimeError("attempt to call non-callable value of type %s", callee.Type.String())
	}
}

// callClosure calls a closure with the given arguments and self value.
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

// makeClosure creates a new closure from a function object.
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

// captureUpvalue captures a local variable as an upvalue for closure support.
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

// closeUpvaluesForFrame closes all open upvalues that reference the given frame's locals.
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

// buildStackTrace constructs a stack trace from the current call frames.
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

// frameName returns a descriptive name for a call frame.
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

// setExceptObject sets the current exception object in the VM.
func (vm *VM) setExceptObject(val Value) {
	vm.exceptObject = val
	vm.setGlobal(builtinExceptObjectIndex, val)
}

// currentFrame returns the top frame of the call stack.
func (vm *VM) currentFrame() *callFrame {
	if len(vm.frames) == 0 {
		return nil
	}
	return &vm.frames[len(vm.frames)-1]
}

// unwindFramesTo unwinds the call stack to the specified frame index.
func (vm *VM) unwindFramesTo(target int) {
	if target < 0 {
		target = 0
	}
	for len(vm.frames)-1 > target {
		frame := &vm.frames[len(vm.frames)-1]
		vm.closeUpvaluesForFrame(frame)
		vm.frames = vm.frames[:len(vm.frames)-1]
	}
}

// beginFinally prepares to execute a finally block by popping the exception handler.
func (vm *VM) beginFinally() (finallyContext, error) {
	if len(vm.exceptionHandlers) == 0 {
		return finallyContext{}, vm.runtimeError("FINALLY without matching TRY")
	}
	handler := vm.exceptionHandlers[len(vm.exceptionHandlers)-1]
	vm.exceptionHandlers = vm.exceptionHandlers[:len(vm.exceptionHandlers)-1]
	ctx := finallyContext{
		exceptionValue:   handler.exceptionValue,
		exceptionActive:  handler.exceptionActive,
		exceptionHandled: handler.exceptionHandled,
		prevExceptObject: handler.prevExceptObject,
	}
	if handler.exceptionActive {
		vm.setExceptObject(handler.exceptionValue)
	} else {
		vm.setExceptObject(NilValue())
	}
	return ctx, nil
}

// markTopHandlerUnhandled marks the top exception handler as unhandled.
func (vm *VM) markTopHandlerUnhandled() {
	if len(vm.exceptionHandlers) == 0 {
		return
	}
	handler := &vm.exceptionHandlers[len(vm.exceptionHandlers)-1]
	handler.exceptionHandled = false
}

// raiseException raises an exception and unwinds the stack to find an appropriate handler.
func (vm *VM) raiseException(exc Value) error {
	for len(vm.exceptionHandlers) > 0 {
		idx := len(vm.exceptionHandlers) - 1
		handler := &vm.exceptionHandlers[idx]
		vm.unwindFramesTo(handler.frameIndex)
		vm.trimStack(handler.stackDepth)
		handler.exceptionValue = exc
		if !handler.exceptionActive {
			handler.exceptionActive = true
			handler.exceptionHandled = !handler.info.HasCatch
			handler.catchCompleted = !handler.info.HasCatch
		}
		if handler.info.HasCatch && !handler.catchCompleted {
			handler.catchCompleted = true
			frame := vm.currentFrame()
			if frame == nil {
				break
			}
			frame.ip = handler.info.CatchTarget
			return nil
		}
		if handler.info.HasFinally && handler.info.FinallyTarget >= 0 {
			frame := vm.currentFrame()
			if frame == nil {
				break
			}
			frame.ip = handler.info.FinallyTarget
			return nil
		}
		vm.exceptionHandlers = vm.exceptionHandlers[:idx]
	}
	return vm.runtimeError("unhandled exception: %s", exc.String())
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
