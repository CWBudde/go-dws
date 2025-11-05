package bytecode

import (
	"fmt"
)

// Value represents a runtime value in the bytecode VM.
// This is a simple tagged union implementation for DWScript types.
type Value struct {
	Type ValueType
	Data interface{} // Integer (int64), Float (float64), String (string), Boolean (bool), etc.
}

// ValueType represents the type tag for a Value.
type ValueType byte

const (
	ValueNil ValueType = iota
	ValueBool
	ValueInt
	ValueFloat
	ValueString
	ValueArray
	ValueObject
	ValueFunction
	ValueClosure
)

// ValueTypeNames maps value types to their string names for debugging.
var ValueTypeNames = [...]string{
	ValueNil:      "nil",
	ValueBool:     "bool",
	ValueInt:      "int",
	ValueFloat:    "float",
	ValueString:   "string",
	ValueArray:    "array",
	ValueObject:   "object",
	ValueFunction: "function",
	ValueClosure:  "closure",
}

// String returns a string representation of the value type.
func (vt ValueType) String() string {
	if int(vt) < len(ValueTypeNames) {
		return ValueTypeNames[vt]
	}
	return "unknown"
}

// Helper constructors for common value types
func NilValue() Value {
	return Value{Type: ValueNil, Data: nil}
}

func BoolValue(b bool) Value {
	return Value{Type: ValueBool, Data: b}
}

func IntValue(i int64) Value {
	return Value{Type: ValueInt, Data: i}
}

func FloatValue(f float64) Value {
	return Value{Type: ValueFloat, Data: f}
}

func StringValue(s string) Value {
	return Value{Type: ValueString, Data: s}
}

// Type checking methods
func (v Value) IsNil() bool    { return v.Type == ValueNil }
func (v Value) IsBool() bool   { return v.Type == ValueBool }
func (v Value) IsInt() bool    { return v.Type == ValueInt }
func (v Value) IsFloat() bool  { return v.Type == ValueFloat }
func (v Value) IsString() bool { return v.Type == ValueString }
func (v Value) IsNumber() bool { return v.Type == ValueInt || v.Type == ValueFloat }

// Type conversion methods
func (v Value) AsBool() bool {
	if v.Type == ValueBool {
		return v.Data.(bool)
	}
	return false
}

func (v Value) AsInt() int64 {
	if v.Type == ValueInt {
		return v.Data.(int64)
	}
	return 0
}

func (v Value) AsFloat() float64 {
	if v.Type == ValueFloat {
		return v.Data.(float64)
	}
	if v.Type == ValueInt {
		return float64(v.Data.(int64))
	}
	return 0.0
}

func (v Value) AsString() string {
	if v.Type == ValueString {
		return v.Data.(string)
	}
	return ""
}

// String returns a human-readable representation of the value.
func (v Value) String() string {
	switch v.Type {
	case ValueNil:
		return "nil"
	case ValueBool:
		if v.Data.(bool) {
			return "true"
		}
		return "false"
	case ValueInt:
		return fmt.Sprintf("%d", v.Data.(int64))
	case ValueFloat:
		return fmt.Sprintf("%g", v.Data.(float64))
	case ValueString:
		return fmt.Sprintf("%q", v.Data.(string))
	default:
		return fmt.Sprintf("<%s>", v.Type)
	}
}

// LineInfo stores line number information for error reporting.
// Uses run-length encoding to save memory: each entry maps a range of
// instructions to a source line number.
type LineInfo struct {
	// InstructionOffset is the index of the first instruction on this line
	InstructionOffset int
	// Line is the source line number (1-based)
	Line int
}

// Chunk represents a compiled bytecode chunk with instructions and constants.
// A chunk is the basic unit of compilation - typically one function or script.
type Chunk struct {
	// Code contains the bytecode instructions
	Code []Instruction

	// Constants is the constant pool containing literal values
	Constants []Value

	// Lines maps instruction indices to source line numbers for error reporting
	// Uses run-length encoding: each LineInfo entry covers instructions from
	// InstructionOffset to the next entry's offset (or end of code)
	Lines []LineInfo

	// Name is a human-readable name for this chunk (function name, script name, etc.)
	Name string
}

// NewChunk creates a new empty bytecode chunk.
func NewChunk(name string) *Chunk {
	return &Chunk{
		Code:      make([]Instruction, 0, 64), // Start with reasonable capacity
		Constants: make([]Value, 0, 16),
		Lines:     make([]LineInfo, 0, 16),
		Name:      name,
	}
}

// WriteInstruction appends an instruction to the chunk.
// Returns the index where the instruction was written.
func (c *Chunk) WriteInstruction(instruction Instruction, line int) int {
	index := len(c.Code)
	c.Code = append(c.Code, instruction)
	c.addLineInfo(index, line)
	return index
}

// Write is a convenience method for writing an instruction with operands.
func (c *Chunk) Write(op OpCode, a byte, b uint16, line int) int {
	return c.WriteInstruction(MakeInstruction(op, a, b), line)
}

// WriteSimple is a convenience method for writing an instruction with no operands.
func (c *Chunk) WriteSimple(op OpCode, line int) int {
	return c.WriteInstruction(MakeSimpleInstruction(op), line)
}

// AddConstant adds a constant to the constant pool and returns its index.
// If the constant already exists, returns the existing index (constant deduplication).
func (c *Chunk) AddConstant(value Value) int {
	// Check if constant already exists (deduplication for simple types)
	for i, existing := range c.Constants {
		if c.valuesEqual(existing, value) {
			return i
		}
	}

	// Add new constant
	index := len(c.Constants)
	c.Constants = append(c.Constants, value)
	return index
}

// valuesEqual checks if two values are equal (for constant deduplication).
func (c *Chunk) valuesEqual(a, b Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case ValueNil:
		return true
	case ValueBool:
		return a.Data.(bool) == b.Data.(bool)
	case ValueInt:
		return a.Data.(int64) == b.Data.(int64)
	case ValueFloat:
		return a.Data.(float64) == b.Data.(float64)
	case ValueString:
		return a.Data.(string) == b.Data.(string)
	default:
		// For complex types (arrays, objects), don't deduplicate
		return false
	}
}

// GetConstant retrieves a constant by index.
func (c *Chunk) GetConstant(index int) Value {
	if index < 0 || index >= len(c.Constants) {
		return NilValue()
	}
	return c.Constants[index]
}

// addLineInfo adds line number information for an instruction.
// Uses run-length encoding: only adds a new entry if the line number changes.
func (c *Chunk) addLineInfo(instructionIndex, line int) {
	// If this is the first instruction or line changed, add new entry
	if len(c.Lines) == 0 || c.Lines[len(c.Lines)-1].Line != line {
		c.Lines = append(c.Lines, LineInfo{
			InstructionOffset: instructionIndex,
			Line:              line,
		})
	}
	// Otherwise, the previous entry covers this instruction too
}

// GetLine returns the source line number for a given instruction index.
// Returns 0 if no line information is available.
func (c *Chunk) GetLine(instructionIndex int) int {
	if len(c.Lines) == 0 {
		return 0
	}

	// Binary search to find the correct LineInfo entry
	// The entry's InstructionOffset is the first instruction on that line
	// All instructions up to the next entry (or end) are on that line
	left, right := 0, len(c.Lines)-1
	result := 0

	for left <= right {
		mid := (left + right) / 2
		if c.Lines[mid].InstructionOffset <= instructionIndex {
			result = c.Lines[mid].Line
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

// InstructionCount returns the number of instructions in the chunk.
func (c *Chunk) InstructionCount() int {
	return len(c.Code)
}

// ConstantCount returns the number of constants in the constant pool.
func (c *Chunk) ConstantCount() int {
	return len(c.Constants)
}

// PatchInstruction replaces an instruction at the given offset.
// Used for backpatching jump targets after code generation.
func (c *Chunk) PatchInstruction(offset int, instruction Instruction) {
	if offset >= 0 && offset < len(c.Code) {
		c.Code[offset] = instruction
	}
}

// PatchJump patches a jump instruction with the calculated offset.
// jumpInstruction is the index of the jump instruction to patch.
// Returns an error if the offset is out of range for 16-bit signed offset.
func (c *Chunk) PatchJump(jumpInstruction int) error {
	// Calculate jump offset: from instruction after jump to current position
	offset := len(c.Code) - jumpInstruction - 1

	// Check if offset fits in 16-bit signed integer
	if offset > 32767 || offset < -32768 {
		return fmt.Errorf("jump offset too large: %d", offset)
	}

	// Get existing instruction and update its B operand with offset
	inst := c.Code[jumpInstruction]
	op := inst.OpCode()
	a := inst.A()
	c.Code[jumpInstruction] = MakeInstruction(op, a, uint16(offset))

	return nil
}

// EmitJump emits a jump instruction with a placeholder offset.
// Returns the index of the jump instruction for later patching.
func (c *Chunk) EmitJump(op OpCode, line int) int {
	return c.Write(op, 0, 0xFFFF, line) // 0xFFFF as placeholder
}

// EmitLoop emits a backward jump to loopStart.
// Returns an error if the offset is out of range.
func (c *Chunk) EmitLoop(loopStart int, line int) error {
	offset := len(c.Code) - loopStart + 1

	if offset > 32767 {
		return fmt.Errorf("loop body too large: %d instructions", offset)
	}

	// Backward jump: encode as negative offset
	c.Write(OpLoop, 0, uint16(-offset), line)
	return nil
}

// Optimize performs simple peephole optimizations on the bytecode.
// This is optional and can be called after compilation is complete.
func (c *Chunk) Optimize() {
	// TODO: Implement peephole optimizations in task 11.18.11
	// Examples:
	// - LOAD_CONST followed by POP -> remove both
	// - Multiple POPs -> single instruction
	// - Constant folding for compile-time expressions
	// - Jump to next instruction -> remove jump
}

// Stats returns statistics about the chunk for debugging.
type ChunkStats struct {
	Name             string
	InstructionCount int
	ConstantCount    int
	CodeBytes        int // Approximate memory usage
	UniqueLines      int
}

// GetStats returns statistics about the chunk.
func (c *Chunk) GetStats() ChunkStats {
	return ChunkStats{
		Name:             c.Name,
		InstructionCount: len(c.Code),
		ConstantCount:    len(c.Constants),
		CodeBytes:        len(c.Code) * 4, // 4 bytes per instruction
		UniqueLines:      len(c.Lines),
	}
}

// String returns a human-readable representation of the chunk.
func (c *Chunk) String() string {
	stats := c.GetStats()
	return fmt.Sprintf("Chunk '%s': %d instructions, %d constants, %d lines",
		stats.Name, stats.InstructionCount, stats.ConstantCount, stats.UniqueLines)
}

// Validate checks the chunk for basic correctness.
// Returns an error if the chunk is malformed.
func (c *Chunk) Validate() error {
	// Check that constant references are valid
	for i, inst := range c.Code {
		op := inst.OpCode()

		switch op {
		case OpLoadConst:
			constIndex := int(inst.B())
			if constIndex >= len(c.Constants) {
				return fmt.Errorf("instruction %d: constant index %d out of range (have %d constants)",
					i, constIndex, len(c.Constants))
			}
		case OpLoadConst0:
			if len(c.Constants) < 1 {
				return fmt.Errorf("instruction %d: LOAD_CONST_0 but no constants", i)
			}
		case OpLoadConst1:
			if len(c.Constants) < 2 {
				return fmt.Errorf("instruction %d: LOAD_CONST_1 but only %d constants", i, len(c.Constants))
			}
		}

		// TODO: Add more validation for jumps, etc.
	}

	return nil
}
