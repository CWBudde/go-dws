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
	line := d.chunk.GetLine(offset)
	if offset > 0 && line == d.chunk.GetLine(offset-1) {
		// Same line as previous instruction, use pipe
		fmt.Fprintf(d.writer, "%04d    | ", offset)
	} else {
		// New line
		fmt.Fprintf(d.writer, "%04d %4d ", offset, line)
	}

	// Disassemble based on opcode type
	switch op {
	// Simple instructions (no operands)
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
		OpArrayLength, OpArrayHigh, OpArrayLow,
		OpStringConcat, OpStringLength,
		OpIntToFloat, OpFloatToInt, OpIntToString, OpFloatToString, OpBoolToString,
		OpGetClass, OpGetSelf,
		OpHalt, OpPrint, OpAssert, OpDebugger,
		OpBreak, OpContinue:
		d.simpleInstruction(inst)

	// Constant instructions
	case OpLoadConst, OpLoadConst0, OpLoadConst1:
		d.constantInstruction(inst, offset)

	// Local variable instructions
	case OpLoadLocal, OpStoreLocal:
		d.byteInstruction(inst, "local")

	// Global variable instructions
	case OpLoadGlobal, OpStoreGlobal:
		d.byteInstruction(inst, "global")

	// Upvalue instructions
	case OpLoadUpvalue, OpStoreUpvalue:
		d.byteInstruction(inst, "upvalue")

	// Jump instructions
	case OpJump, OpJumpIfTrue, OpJumpIfFalse, OpJumpIfTrueNoPop, OpJumpIfFalseNoPop:
		d.jumpInstruction(inst, offset, 1)

	// Loop instruction (backward jump)
	case OpLoop:
		d.jumpInstruction(inst, offset, -1)

	// For loop instructions
	case OpForPrep, OpForLoop:
		d.forInstruction(inst, offset)

	// Function call instructions
	case OpCall, OpCallMethod, OpCallVirtual, OpCallBuiltin, OpCallIndirect, OpTailCall:
		d.callInstruction(inst)

	// Return instruction
	case OpReturn, OpExit:
		d.returnInstruction(inst)

	// Closure instruction
	case OpClosure:
		d.closureInstruction(inst)

	// Array instructions
	case OpNewArray:
		d.byteInstruction(inst, "elements")

	case OpNewArraySized, OpArrayGet, OpArraySet, OpArraySetLength:
		d.simpleInstruction(inst)

	// Object instructions
	case OpNewObject:
		d.byteInstruction(inst, "class")

	case OpGetField, OpSetField:
		d.byteInstruction(inst, "field")

	case OpGetProperty, OpSetProperty:
		d.byteInstruction(inst, "property")

	case OpInstanceOf, OpCastObject:
		d.byteInstruction(inst, "class")

	case OpInvoke:
		d.invokeInstruction(inst)

	// String instructions
	case OpStringGet, OpStringSlice:
		d.simpleInstruction(inst)

	// Type conversion
	case OpVariantToType:
		d.byteInstruction(inst, "type")

	// Exception handling
	case OpTry, OpCatch, OpFinally:
		d.jumpInstruction(inst, offset, 1)

	case OpThrow:
		d.simpleInstruction(inst)

	// Case instruction
	case OpCase:
		d.byteInstruction(inst, "jumpTable")

	default:
		fmt.Fprintf(d.writer, "UNKNOWN_OP %d\n", op)
	}
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
