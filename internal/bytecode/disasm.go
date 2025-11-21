package bytecode

import (
	"fmt"
	"io"
	"strings"
)

// Disassembler provides human-readable bytecode disassembly for debugging.
type Disassembler struct {
	writer io.Writer
	chunk  *Chunk
}

// NewDisassembler creates a new disassembler for the given chunk.
func NewDisassembler(chunk *Chunk, writer io.Writer) *Disassembler {
	return &Disassembler{
		writer: writer,
		chunk:  chunk,
	}
}

// Disassemble prints a complete disassembly of the chunk.
func (d *Disassembler) Disassemble() {
	fmt.Fprintf(d.writer, "== %s ==\n", d.chunk.Name)
	fmt.Fprintf(d.writer, "Instructions: %d, Constants: %d, Lines: %d\n\n",
		len(d.chunk.Code), len(d.chunk.Constants), len(d.chunk.Lines))

	// Print constants pool
	if len(d.chunk.Constants) > 0 {
		fmt.Fprintf(d.writer, "Constants Pool:\n")
		for i, constant := range d.chunk.Constants {
			fmt.Fprintf(d.writer, "  [%04d] %s\n", i, constant.String())
		}
		fmt.Fprintf(d.writer, "\n")
	}

	// Print bytecode
	fmt.Fprintf(d.writer, "Bytecode:\n")
	for offset := 0; offset < len(d.chunk.Code); offset++ {
		d.DisassembleInstruction(offset)
	}

	fmt.Fprintf(d.writer, "\n")
}

// DisassembleInstruction prints a single instruction at the given offset.
func (d *Disassembler) DisassembleInstruction(offset int) {
	if offset < 0 || offset >= len(d.chunk.Code) {
		fmt.Fprintf(d.writer, "Invalid offset: %d\n", offset)
		return
	}

	inst := d.chunk.Code[offset]
	op := inst.OpCode()

	// Print offset and line number
	d.printInstructionHeader(offset)

	// Dispatch to appropriate handler based on opcode category
	if d.tryDisassembleSimpleOp(inst, op) {
		return
	}
	if d.tryDisassembleConstantOp(inst, op, offset) {
		return
	}
	if d.tryDisassembleVarOp(inst, op) {
		return
	}
	if d.tryDisassembleJumpOp(inst, op, offset) {
		return
	}
	if d.tryDisassembleCallOp(inst, op) {
		return
	}
	if d.tryDisassembleArrayOp(inst, op) {
		return
	}
	if d.tryDisassembleObjectOp(inst, op) {
		return
	}
	if d.tryDisassembleStringOp(inst, op) {
		return
	}
	if d.tryDisassembleExceptionOp(inst, op, offset) {
		return
	}
	if d.tryDisassembleMiscOp(inst, op) {
		return
	}

	// Unknown opcode
	fmt.Fprintf(d.writer, "UNKNOWN_OP %d\n", op)
}

// printInstructionHeader prints the offset and line number prefix for an instruction.
func (d *Disassembler) printInstructionHeader(offset int) {
	line := d.chunk.GetLine(offset)
	if offset > 0 && line == d.chunk.GetLine(offset-1) {
		fmt.Fprintf(d.writer, "%04d    | ", offset)
	} else {
		fmt.Fprintf(d.writer, "%04d %4d ", offset, line)
	}
}

// tryDisassembleSimpleOp attempts to disassemble simple instructions that take no operands.
func (d *Disassembler) tryDisassembleSimpleOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpLoadNil, OpLoadTrue, OpLoadFalse,
		OpAddInt, OpSubInt, OpMulInt, OpDivInt, OpModInt, OpNegateInt, OpIncInt, OpDecInt,
		OpBitAnd, OpBitOr, OpBitXor, OpBitNot,
		OpAddFloat, OpSubFloat, OpMulFloat, OpDivFloat, OpNegateFloat,
		OpPower, OpSqrt, OpAbs, OpFloor, OpCeil, OpRound, OpTrunc,
		OpEqual, OpNotEqual, OpLess, OpLessEqual, OpGreater, OpGreaterEqual,
		OpCompareInt, OpCompareFloat,
		OpNot, OpAnd, OpOr, OpXor,
		OpShl, OpShr, OpSar, OpRotl,
		OpPop, OpDup, OpDup2, OpSwap, OpRotate3,
		OpArrayLength, OpArrayHigh, OpArrayLow, OpArrayCount, OpArrayDelete, OpArrayIndexOf,
		OpStringConcat, OpStringLength,
		OpIntToFloat, OpFloatToInt, OpIntToString, OpFloatToString, OpBoolToString, OpToBool,
		OpGetClass, OpGetSelf,
		OpHalt, OpPrint, OpAssert, OpDebugger,
		OpBreak, OpContinue:
		d.simpleInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleConstantOp attempts to disassemble constant loading instructions.
func (d *Disassembler) tryDisassembleConstantOp(inst Instruction, op OpCode, offset int) bool {
	switch op {
	case OpLoadConst, OpLoadConst0, OpLoadConst1:
		d.constantInstruction(inst, offset)
		return true
	}
	return false
}

// tryDisassembleVarOp attempts to disassemble variable load and store instructions (local, global, upvalue).
func (d *Disassembler) tryDisassembleVarOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpLoadLocal, OpStoreLocal:
		d.byteInstruction(inst, "local")
		return true
	case OpLoadGlobal, OpStoreGlobal:
		d.byteInstruction(inst, "global")
		return true
	case OpLoadUpvalue, OpStoreUpvalue:
		d.byteInstruction(inst, "upvalue")
		return true
	}
	return false
}

// tryDisassembleJumpOp attempts to disassemble jump and loop instructions.
func (d *Disassembler) tryDisassembleJumpOp(inst Instruction, op OpCode, offset int) bool {
	switch op {
	case OpJump, OpJumpIfTrue, OpJumpIfFalse, OpJumpIfTrueNoPop, OpJumpIfFalseNoPop:
		d.jumpInstruction(inst, offset, 1)
		return true
	case OpLoop:
		d.jumpInstruction(inst, offset, -1)
		return true
	case OpForPrep, OpForLoop:
		d.forInstruction(inst, offset)
		return true
	}
	return false
}

// tryDisassembleCallOp attempts to disassemble function call and return instructions.
func (d *Disassembler) tryDisassembleCallOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpCall, OpCallMethod, OpCallVirtual, OpCallBuiltin, OpCallIndirect, OpTailCall:
		d.callInstruction(inst)
		return true
	case OpReturn, OpExit:
		d.returnInstruction(inst)
		return true
	case OpClosure:
		d.closureInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleArrayOp attempts to disassemble array-related instructions.
func (d *Disassembler) tryDisassembleArrayOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpNewArray:
		d.byteInstruction(inst, "elements")
		return true
	case OpNewSet:
		d.byteInstruction(inst, "elements")
		return true
	case OpNewArrayMultiDim:
		// Show both dimension count (A) and type index (B)
		dimCount := inst.A()
		typeIndex := inst.B()
		fmt.Fprintf(d.writer, "%-20s dims=%d typeIdx=%d\n", inst.String(), dimCount, typeIndex)
		return true
	case OpNewArraySized:
		// Show type index (B)
		d.byteInstruction(inst, "typeIdx")
		return true
	case OpArrayGet, OpArraySet, OpArraySetLength:
		d.simpleInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleObjectOp attempts to disassemble object-related instructions (fields, properties, methods).
func (d *Disassembler) tryDisassembleObjectOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpNewObject:
		d.byteInstruction(inst, "class")
		return true
	case OpNewRecord:
		// Task 9.7
		d.byteInstruction(inst, "type")
		return true
	case OpGetField, OpSetField:
		d.byteInstruction(inst, "field")
		return true
	case OpGetProperty, OpSetProperty:
		d.byteInstruction(inst, "property")
		return true
	case OpInstanceOf, OpCastObject:
		d.byteInstruction(inst, "class")
		return true
	case OpInvoke:
		d.invokeInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleStringOp attempts to disassemble string operation instructions.
func (d *Disassembler) tryDisassembleStringOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpStringGet, OpStringSlice:
		d.simpleInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleExceptionOp attempts to disassemble exception handling instructions.
func (d *Disassembler) tryDisassembleExceptionOp(inst Instruction, op OpCode, offset int) bool {
	switch op {
	case OpTry, OpCatch, OpFinally:
		d.jumpInstruction(inst, offset, 1)
		return true
	case OpThrow:
		d.simpleInstruction(inst)
		return true
	}
	return false
}

// tryDisassembleMiscOp attempts to disassemble miscellaneous instructions that don't fit other categories.
func (d *Disassembler) tryDisassembleMiscOp(inst Instruction, op OpCode) bool {
	switch op {
	case OpCase:
		d.byteInstruction(inst, "jumpTable")
		return true
	case OpVariantToType:
		d.byteInstruction(inst, "type")
		return true
	}
	return false
}

// simpleInstruction prints an instruction with no operands.
func (d *Disassembler) simpleInstruction(inst Instruction) {
	fmt.Fprintf(d.writer, "%s\n", inst.String())
}

// constantInstruction prints an instruction that references the constant pool.
func (d *Disassembler) constantInstruction(inst Instruction, offset int) {
	op := inst.OpCode()
	var constIndex int

	switch op {
	case OpLoadConst:
		constIndex = int(inst.B())
	case OpLoadConst0:
		constIndex = 0
	case OpLoadConst1:
		constIndex = 1
	}

	constant := d.chunk.GetConstant(constIndex)
	fmt.Fprintf(d.writer, "%-20s %4d '%s'\n", inst.String(), constIndex, constant.String())
}

// byteInstruction prints an instruction with a single byte operand.
func (d *Disassembler) byteInstruction(inst Instruction, operandName string) {
	index := inst.B()
	fmt.Fprintf(d.writer, "%-20s %4d  ; %s index\n", inst.String(), index, operandName)
}

// jumpInstruction prints a jump instruction.
func (d *Disassembler) jumpInstruction(inst Instruction, offset int, sign int) {
	jumpOffset := int(inst.SignedB())
	target := offset + 1 + sign*jumpOffset

	fmt.Fprintf(d.writer, "%-20s %4d -> %04d\n", inst.String(), jumpOffset, target)
}

// forInstruction prints a for-loop instruction.
func (d *Disassembler) forInstruction(inst Instruction, offset int) {
	loopVar := inst.A()
	jumpOffset := int(inst.SignedB())
	target := offset + 1 + jumpOffset

	fmt.Fprintf(d.writer, "%-20s var=%d, jump=%d -> %04d\n",
		inst.String(), loopVar, jumpOffset, target)
}

// callInstruction prints a function call instruction.
func (d *Disassembler) callInstruction(inst Instruction) {
	argCount := inst.A()
	funcIndex := inst.B()

	fmt.Fprintf(d.writer, "%-20s args=%d, func=%d\n", inst.String(), argCount, funcIndex)
}

// returnInstruction prints a return instruction.
func (d *Disassembler) returnInstruction(inst Instruction) {
	hasReturnValue := inst.A() != 0
	if hasReturnValue {
		fmt.Fprintf(d.writer, "%-20s (with value)\n", inst.String())
	} else {
		fmt.Fprintf(d.writer, "%-20s\n", inst.String())
	}
}

// closureInstruction prints a closure creation instruction.
func (d *Disassembler) closureInstruction(inst Instruction) {
	upvalueCount := inst.A()
	funcIndex := inst.B()

	fmt.Fprintf(d.writer, "%-20s upvalues=%d, func=%d\n",
		inst.String(), upvalueCount, funcIndex)
}

// invokeInstruction prints a method invocation instruction.
func (d *Disassembler) invokeInstruction(inst Instruction) {
	argCount := inst.A()
	methodIndex := inst.B()

	fmt.Fprintf(d.writer, "%-20s args=%d, method=%d\n",
		inst.String(), argCount, methodIndex)
}

// DisassembleRange disassembles a range of instructions.
func (d *Disassembler) DisassembleRange(start, end int) {
	if start < 0 {
		start = 0
	}
	if end > len(d.chunk.Code) {
		end = len(d.chunk.Code)
	}

	fmt.Fprintf(d.writer, "== %s (instructions %d-%d) ==\n\n", d.chunk.Name, start, end-1)

	for offset := start; offset < end; offset++ {
		d.DisassembleInstruction(offset)
	}

	fmt.Fprintf(d.writer, "\n")
}

// DisassembleToString returns the disassembly as a string.
func DisassembleToString(chunk *Chunk) string {
	var sb strings.Builder
	d := NewDisassembler(chunk, &sb)
	d.Disassemble()
	return sb.String()
}

// DisassembleInstructionToString returns a single instruction disassembly as a string.
func DisassembleInstructionToString(chunk *Chunk, offset int) string {
	var sb strings.Builder
	d := NewDisassembler(chunk, &sb)
	d.DisassembleInstruction(offset)
	return sb.String()
}
