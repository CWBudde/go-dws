package bytecode

import (
	"fmt"
	"strings"
)

// Value represents a runtime value in the bytecode VM.
// This is a simple tagged union implementation for DWScript types.
type Value struct {
	Data interface{}
	Type ValueType
}

// FunctionObject represents a compiled function in bytecode form.
type FunctionObject struct {
	Chunk       *Chunk
	Name        string
	UpvalueDefs []UpvalueDef
	Arity       int
}

// NewFunctionObject creates a new function object.
func NewFunctionObject(name string, chunk *Chunk, arity int) *FunctionObject {
	return &FunctionObject{
		Name:  name,
		Chunk: chunk,
		Arity: arity,
	}
}

// UpvalueCount returns the number of upvalues this function expects.
func (fn *FunctionObject) UpvalueCount() int {
	if fn == nil {
		return 0
	}
	return len(fn.UpvalueDefs)
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
	ValueRecord // Task 9.7: Record value type
	ValueFunction
	ValueClosure
	ValueBuiltin
	ValueVariant
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
	ValueRecord:   "record", // Task 9.7
	ValueFunction: "function",
	ValueClosure:  "closure",
	ValueBuiltin:  "builtin",
	ValueVariant:  "variant",
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

// ArrayValue constructs a Value representing an array instance.
func ArrayValue(arr *ArrayInstance) Value {
	return Value{Type: ValueArray, Data: arr}
}

// ObjectValue constructs a Value representing an object instance.
func ObjectValue(obj *ObjectInstance) Value {
	return Value{Type: ValueObject, Data: obj}
}

// UpvalueDef describes how a function captures an upvalue when a closure is created.
type UpvalueDef struct {
	IsLocal bool // true if capturing a local variable from the creating frame
	Index   int  // index of the local or upvalue to capture
}

// Closure represents a function together with its captured upvalues.
type Closure struct {
	Function *FunctionObject
	Upvalues []*Upvalue
}

// Upvalue represents a captured variable that may reference a stack slot (open)
// or hold a closed-over value.
type Upvalue struct {
	location *Value
	closed   Value
}

// FunctionValue constructs a Value representing a function.
func FunctionValue(fn *FunctionObject) Value {
	return Value{Type: ValueFunction, Data: fn}
}

// ClosureValue constructs a Value representing a closure.
func ClosureValue(cl *Closure) Value {
	return Value{Type: ValueClosure, Data: cl}
}

// BuiltinInfo holds information about a built-in function.
type BuiltinInfo struct {
	Func interface{}
	Name string
}

// BuiltinValue constructs a Value representing a built-in function.
func BuiltinValue(name string) Value {
	return Value{Type: ValueBuiltin, Data: name}
}

// VariantValue constructs a Value representing a Variant (wraps another value).
func VariantValue(wrapped Value) Value {
	return Value{Type: ValueVariant, Data: wrapped}
}

// Type checking methods
func (v Value) IsNil() bool    { return v.Type == ValueNil }
func (v Value) IsBool() bool   { return v.Type == ValueBool }
func (v Value) IsInt() bool    { return v.Type == ValueInt }
func (v Value) IsFloat() bool  { return v.Type == ValueFloat }
func (v Value) IsString() bool { return v.Type == ValueString }
func (v Value) IsArray() bool  { return v.Type == ValueArray }
func (v Value) IsNumber() bool { return v.Type == ValueInt || v.Type == ValueFloat }
func (v Value) IsFunction() bool {
	return v.Type == ValueFunction
}
func (v Value) IsClosure() bool {
	return v.Type == ValueClosure
}
func (v Value) IsObject() bool {
	return v.Type == ValueObject
}
func (v Value) IsRecord() bool {
	return v.Type == ValueRecord
}
func (v Value) IsBuiltin() bool {
	return v.Type == ValueBuiltin
}
func (v Value) IsVariant() bool {
	return v.Type == ValueVariant
}

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

// AsArray returns the underlying array instance if the value is an array.
func (v Value) AsArray() *ArrayInstance {
	if v.Type == ValueArray {
		if arr, ok := v.Data.(*ArrayInstance); ok {
			return arr
		}
	}
	return nil
}

// AsFunction returns the underlying function object if the value represents a function.
func (v Value) AsFunction() *FunctionObject {
	if v.Type == ValueFunction {
		if fn, ok := v.Data.(*FunctionObject); ok {
			return fn
		}
	}
	return nil
}

// AsClosure returns the underlying closure if the value is a closure.
func (v Value) AsClosure() *Closure {
	if v.Type == ValueClosure {
		if cl, ok := v.Data.(*Closure); ok {
			return cl
		}
	}
	return nil
}

// AsObject returns the underlying object instance if the value is an object.
func (v Value) AsObject() *ObjectInstance {
	if v.Type == ValueObject {
		if obj, ok := v.Data.(*ObjectInstance); ok {
			return obj
		}
	}
	return nil
}
func (v Value) AsRecord() *RecordInstance {
	if v.Type == ValueRecord {
		if rec, ok := v.Data.(*RecordInstance); ok {
			return rec
		}
	}
	return nil
}

// AsVariant returns the wrapped value if this is a Variant.
func (v Value) AsVariant() Value {
	if v.Type == ValueVariant {
		if wrapped, ok := v.Data.(Value); ok {
			return wrapped
		}
	}
	return NilValue()
}

// AsBuiltin returns the built-in function name if the value is a built-in.
func (v Value) AsBuiltin() string {
	if v.Type == ValueBuiltin {
		if name, ok := v.Data.(string); ok {
			return name
		}
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
	case ValueArray:
		if arr := v.AsArray(); arr != nil {
			return arr.String()
		}
		return "[]"
	case ValueFunction:
		if fn := v.AsFunction(); fn != nil {
			if fn.Name != "" {
				return fmt.Sprintf("<function %s>", fn.Name)
			}
			return "<function>"
		}
		return "<function>"
	case ValueClosure:
		if cl := v.AsClosure(); cl != nil && cl.Function != nil {
			if cl.Function.Name != "" {
				return fmt.Sprintf("<closure %s>", cl.Function.Name)
			}
			return "<closure>"
		}
		return "<closure>"
	case ValueObject:
		if obj := v.AsObject(); obj != nil {
			if obj.ClassName != "" {
				return fmt.Sprintf("<object %s>", obj.ClassName)
			}
			return "<object>"
		}
		return "<object>"
	case ValueRecord:
		if rec := v.AsRecord(); rec != nil {
			if rec.TypeName != "" {
				return fmt.Sprintf("<record %s>", rec.TypeName)
			}
			return "<record>"
		}
		return "<record>"
	case ValueBuiltin:
		name := v.AsBuiltin()
		if name != "" {
			return fmt.Sprintf("<builtin %s>", name)
		}
		return "<builtin>"
	default:
		return fmt.Sprintf("<%s>", v.Type)
	}
}

// newOpenUpvalue creates an upvalue referencing the given stack slot.
func newOpenUpvalue(slot *Value) *Upvalue {
	return &Upvalue{location: slot}
}

// get returns the current value stored in the upvalue.
func (uv *Upvalue) get() Value {
	if uv == nil {
		return NilValue()
	}
	if uv.location != nil {
		return *uv.location
	}
	return uv.closed
}

// set writes a value to the upvalue.
func (uv *Upvalue) set(val Value) {
	if uv == nil {
		return
	}
	if uv.location != nil {
		*uv.location = val
	} else {
		uv.closed = val
	}
}

// close seals the upvalue, copying the referenced stack value if necessary.
func (uv *Upvalue) close() {
	if uv == nil {
		return
	}
	if uv.location != nil {
		uv.closed = *uv.location
		uv.location = nil
	}
}

// ArrayInstance represents a dynamically sized array with value semantics for elements.
type ArrayInstance struct {
	elements []Value
}

// NewArrayInstance creates a new array initialized with the provided elements.
// The slice is copied to preserve value semantics for literals.
func NewArrayInstance(elements []Value) *ArrayInstance {
	if len(elements) == 0 {
		return &ArrayInstance{elements: make([]Value, 0)}
	}
	copyBuf := make([]Value, len(elements))
	copy(copyBuf, elements)
	return &ArrayInstance{elements: copyBuf}
}

// NewArrayInstanceWithLength allocates an array with the requested length.
// Elements are initialized to NilValue.
func NewArrayInstanceWithLength(length int) *ArrayInstance {
	if length < 0 {
		length = 0
	}
	elements := make([]Value, length)
	for i := range elements {
		elements[i] = NilValue()
	}
	return &ArrayInstance{elements: elements}
}

// Length returns the current number of elements in the array.
func (a *ArrayInstance) Length() int {
	if a == nil {
		return 0
	}
	return len(a.elements)
}

// Get returns the element at the specified index.
// The bool return reports if the index was within bounds.
func (a *ArrayInstance) Get(index int) (Value, bool) {
	if a == nil {
		return NilValue(), false
	}
	if index < 0 || index >= len(a.elements) {
		return NilValue(), false
	}
	return a.elements[index], true
}

// Set writes the element at the specified index.
// Returns false if the index was out of range.
func (a *ArrayInstance) Set(index int, value Value) bool {
	if a == nil {
		return false
	}
	if index < 0 || index >= len(a.elements) {
		return false
	}
	a.elements[index] = value
	return true
}

// Resize changes the logical length of the array, filling new slots with nil.
func (a *ArrayInstance) Resize(length int) {
	if a == nil {
		return
	}
	if length < 0 {
		length = 0
	}
	current := len(a.elements)
	switch {
	case length < current:
		a.elements = a.elements[:length]
	case length > current:
		extra := make([]Value, length-current)
		for i := range extra {
			extra[i] = NilValue()
		}
		a.elements = append(a.elements, extra...)
	}
}

// String formats the array value for debugging output.
func (a *ArrayInstance) String() string {
	if a == nil {
		return "[]"
	}
	if len(a.elements) == 0 {
		return "[]"
	}
	var builder strings.Builder
	builder.WriteByte('[')
	for i, elem := range a.elements {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(elem.String())
	}
	builder.WriteByte(']')
	return builder.String()
}

// ObjectInstance represents a simple object with mutable fields and properties.
// This lightweight structure is sufficient for current VM needs and can be
// extended with richer metadata in later milestones.
type ObjectInstance struct {
	fields    map[string]Value
	props     map[string]Value
	ClassName string
}

// NewObjectInstance creates a new object instance with optional class name.
func NewObjectInstance(className string) *ObjectInstance {
	return &ObjectInstance{
		ClassName: className,
		fields:    make(map[string]Value),
		props:     make(map[string]Value),
	}
}

// GetField retrieves a field value by name (case-insensitive).
func (o *ObjectInstance) GetField(name string) (Value, bool) {
	if o == nil {
		return NilValue(), false
	}
	val, ok := o.fields[strings.ToLower(name)]
	if !ok {
		return NilValue(), false
	}
	return val, true
}

// SetField stores a field value by name (case-insensitive).
func (o *ObjectInstance) SetField(name string, value Value) {
	if o == nil {
		return
	}
	if o.fields == nil {
		o.fields = make(map[string]Value)
	}
	o.fields[strings.ToLower(name)] = value
}

// GetProperty retrieves a property value by name (case-insensitive).
func (o *ObjectInstance) GetProperty(name string) (Value, bool) {
	if o == nil {
		return NilValue(), false
	}
	// Try properties first
	val, ok := o.props[strings.ToLower(name)]
	if ok {
		return val, true
	}
	// Fall back to fields (Task 9.5.5: fields are accessed via property syntax)
	val, ok = o.fields[strings.ToLower(name)]
	if ok {
		return val, true
	}
	return NilValue(), false
}

// SetProperty stores a property value by name (case-insensitive).
func (o *ObjectInstance) SetProperty(name string, value Value) {
	if o == nil {
		return
	}
	if o.props == nil {
		o.props = make(map[string]Value)
	}
	o.props[strings.ToLower(name)] = value
}

// ============================================================================
// RecordInstance (Task 9.7)
// ============================================================================

// RecordInstance represents a record value in the bytecode VM.
// Records are value types (unlike classes which are reference types) and
// contain fields that can be accessed via member expressions.
type RecordInstance struct {
	fields   map[string]Value // Field values (case-insensitive keys)
	TypeName string           // Record type name (e.g., "TPoint")
}

// NewRecordInstance creates a new record instance.
func NewRecordInstance(typeName string) *RecordInstance {
	return &RecordInstance{
		TypeName: typeName,
		fields:   make(map[string]Value),
	}
}

// GetField retrieves a field value by name (case-insensitive).
func (r *RecordInstance) GetField(name string) (Value, bool) {
	if r == nil {
		return NilValue(), false
	}
	val, ok := r.fields[strings.ToLower(name)]
	if !ok {
		return NilValue(), false
	}
	return val, true
}

// SetField stores a field value by name (case-insensitive).
func (r *RecordInstance) SetField(name string, value Value) {
	if r == nil {
		return
	}
	if r.fields == nil {
		r.fields = make(map[string]Value)
	}
	r.fields[strings.ToLower(name)] = value
}

// RecordValue constructs a Value representing a record instance.
func RecordValue(rec *RecordInstance) Value {
	return Value{Type: ValueRecord, Data: rec}
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

// FieldMetadata stores metadata for a class field including its initializer.
type FieldMetadata struct {
	Name        string
	Initializer *Chunk // Compiled bytecode for field initializer expression (nil if no initializer)
}

// ClassMetadata stores metadata for a class including field initializers.
type ClassMetadata struct {
	Name   string
	Fields []*FieldMetadata
}

// RecordMetadata stores metadata for record types including static methods (Task 9.1).
// Records can have static methods (class functions) that are called on the type itself,
// not on instances. This metadata maps method names to their constant indices for direct calls.
type RecordMetadata struct {
	Name    string
	Methods map[string]uint16  // Method name (lowercase) -> constant index (for OpCall)
	Fields  []*FieldMetadata   // For future field initializer support
}

// Chunk represents a compiled bytecode chunk with instructions and constants.
// A chunk is the basic unit of compilation - typically one function or script.
type Chunk struct {
	tryInfos   map[int]TryInfo
	Name       string
	Code       []Instruction
	Constants  []Value
	Lines      []LineInfo
	Helpers    map[string]*HelperInfo    // Helper metadata for runtime method resolution
	Classes    map[string]*ClassMetadata // Class metadata for field initialization
	Records    map[string]*RecordMetadata // Record metadata for static methods (Task 9.1)
	LocalCount int
}

// TryInfo describes the catch/finally targets for a try instruction.
type TryInfo struct {
	CatchTarget   int
	FinallyTarget int
	HasCatch      bool
	HasFinally    bool
}

// NewChunk creates a new empty bytecode chunk.
func NewChunk(name string) *Chunk {
	return &Chunk{
		Code:       make([]Instruction, 0, 64), // Start with reasonable capacity
		Constants:  make([]Value, 0, 16),
		Lines:      make([]LineInfo, 0, 16),
		Helpers:    make(map[string]*HelperInfo),
		Classes:    make(map[string]*ClassMetadata),
		Records:    make(map[string]*RecordMetadata), // Task 9.1
		LocalCount: 0,
		Name:       name,
		tryInfos:   make(map[int]TryInfo),
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

// SetTryInfo records metadata for the try instruction at the given index.
func (c *Chunk) SetTryInfo(index int, info TryInfo) {
	if c == nil {
		return
	}
	if c.tryInfos == nil {
		c.tryInfos = make(map[int]TryInfo)
	}
	c.tryInfos[index] = info
}

// TryInfoAt retrieves try metadata for the instruction at the given index.
func (c *Chunk) TryInfoAt(index int) (TryInfo, bool) {
	if c == nil || c.tryInfos == nil {
		return TryInfo{}, false
	}
	info, ok := c.tryInfos[index]
	return info, ok
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

// Optimize performs bytecode optimizations using the pass manager.
// Optional options can be provided to enable/disable individual passes.
func (c *Chunk) Optimize(opts ...OptimizeOption) {
	if c == nil || len(c.Code) == 0 {
		return
	}

	config := defaultOptimizeConfig()
	for _, opt := range opts {
		opt(&config)
	}

	newChunkOptimizer(c, config).run()
}

// Stats returns statistics about the chunk for debugging.
type ChunkStats struct {
	Name             string
	InstructionCount int
	ConstantCount    int
	CodeBytes        int // Approximate memory usage
	UniqueLines      int
	LocalCount       int
}

// GetStats returns statistics about the chunk.
func (c *Chunk) GetStats() ChunkStats {
	return ChunkStats{
		Name:             c.Name,
		InstructionCount: len(c.Code),
		ConstantCount:    len(c.Constants),
		CodeBytes:        len(c.Code) * 4, // 4 bytes per instruction
		UniqueLines:      len(c.Lines),
		LocalCount:       c.LocalCount,
	}
}

// String returns a human-readable representation of the chunk.
func (c *Chunk) String() string {
	stats := c.GetStats()
	return fmt.Sprintf("Chunk '%s': %d instructions, %d constants, %d locals, %d lines",
		stats.Name, stats.InstructionCount, stats.ConstantCount, stats.LocalCount, stats.UniqueLines)
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
