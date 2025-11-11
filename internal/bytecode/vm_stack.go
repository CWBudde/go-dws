package bytecode

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

func (vm *VM) trimStack(depth int) {
	if depth < 0 {
		depth = 0
	}
	if len(vm.stack) > depth {
		vm.stack = vm.stack[:depth]
	}
}
