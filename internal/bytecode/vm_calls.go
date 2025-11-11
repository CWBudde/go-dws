package bytecode

import (
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

// invokeMethod invokes a method on an object receiver.
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
