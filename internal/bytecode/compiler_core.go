package bytecode

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// Compiler converts AST nodes into bytecode chunks.
type Compiler struct {
	functions       map[string]functionInfo
	helpers         map[string]*HelperInfo // Helper registry (keyed by helper name)
	enclosing       *Compiler
	globals         map[string]globalVar
	chunk           *Chunk
	upvalues        []upvalue
	loopStack       []*loopContext
	locals          []local
	optimizeOptions []OptimizeOption
	scopeDepth      int
	lastLine        int
	nextSlot        uint16
	maxSlot         uint16
	nextGlobal      uint16
}

type local struct {
	typ   types.Type
	name  string
	depth int
	slot  uint16
}

type globalVar struct {
	typ   types.Type
	name  string
	index uint16
}

type upvalue struct {
	index   uint16
	isLocal bool
}

type functionInfo struct {
	fn         *FunctionObject
	constIndex uint16
	globalSlot uint16
}

// HelperInfo stores compile-time information about a helper type.
// This metadata is used during bytecode compilation and passed to the VM.
type HelperInfo struct {
	Name         string            // Helper type name (e.g., TStringHelper)
	TargetType   string            // Type being extended (e.g., String, Integer)
	ParentHelper string            // Parent helper name (for inheritance)
	Methods      map[string]uint16 // Method name -> function global slot
	Properties   []string          // Property names
	ClassVars    []string          // Class variable names
	ClassConsts  map[string]Value  // Class constant values
}

type loopKind int

const (
	loopKindWhile loopKind = iota
	loopKindRepeat
)

type loopContext struct {
	breakJumps    []int
	continueJumps []int
	kind          loopKind
	loopStart     int
}

// CompilerOption configures a new compiler instance.
type CompilerOption func(*Compiler)

// NewCompiler creates a compiler for the given chunk name.
func NewCompiler(chunkName string, opts ...CompilerOption) *Compiler {
	return newCompiler(chunkName, nil, opts...)
}

func newCompiler(chunkName string, enclosing *Compiler, opts ...CompilerOption) *Compiler {
	globals := make(map[string]globalVar)
	functions := make(map[string]functionInfo)
	helpers := make(map[string]*HelperInfo)
	if enclosing != nil {
		globals = enclosing.globals
		functions = enclosing.functions
		helpers = enclosing.helpers
	}
	c := &Compiler{
		chunk:     NewChunk(chunkName),
		globals:   globals,
		functions: functions,
		helpers:   helpers,
		enclosing: enclosing,
	}
	if enclosing != nil {
		c.optimizeOptions = enclosing.optimizeOptions
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Compiler) newChildCompiler(name string) *Compiler {
	return newCompiler(name, c)
}

// WithCompilerOptimizeOptions overrides the optimization passes used by this compiler.
func WithCompilerOptimizeOptions(opts ...OptimizeOption) CompilerOption {
	return func(c *Compiler) {
		c.optimizeOptions = opts
	}
}

// Compile compiles the provided program into bytecode.
func (c *Compiler) Compile(program *ast.Program) (*Chunk, error) {
	if program == nil {
		return nil, fmt.Errorf("bytecode compile error: nil program")
	}

	chunkName := c.chunk.Name
	c.chunk = NewChunk(chunkName)

	// Reset compiler state for reuse
	c.locals = c.locals[:0]
	c.upvalues = c.upvalues[:0]
	c.globals = make(map[string]globalVar)
	c.functions = make(map[string]functionInfo)
	c.helpers = make(map[string]*HelperInfo)
	c.loopStack = c.loopStack[:0]
	c.scopeDepth = 0
	c.nextSlot = 0
	c.maxSlot = 0
	c.nextGlobal = 0
	c.enclosing = nil

	c.initBuiltins()

	if err := c.compileProgram(program); err != nil {
		return nil, err
	}

	c.chunk.LocalCount = int(c.maxSlot)

	// Copy helper metadata to the chunk for runtime use
	if len(c.helpers) > 0 {
		c.chunk.Helpers = c.helpers
	}

	if c.needsHalt() {
		c.chunk.WriteSimple(OpHalt, c.lastLine)
	}

	c.chunk.Optimize(c.optimizeOptions...)

	return c.chunk, nil
}

// LocalCount returns the number of local slots allocated during compilation.
func (c *Compiler) LocalCount() int {
	return int(c.maxSlot)
}

func (c *Compiler) emitLoadConstant(value Value, line int) error {
	index := c.chunk.AddConstant(value)
	if index > 0xFFFF {
		return c.errorf(nil, "constant pool overflow")
	}

	switch index {
	case 0:
		c.chunk.WriteSimple(OpLoadConst0, line)
	case 1:
		c.chunk.WriteSimple(OpLoadConst1, line)
	default:
		c.chunk.Write(OpLoadConst, 0, uint16(index), line)
	}

	return nil
}

func (c *Compiler) emitValue(value Value, line int) error {
	switch value.Type {
	case ValueNil:
		c.chunk.WriteSimple(OpLoadNil, line)
		return nil
	case ValueBool:
		if value.AsBool() {
			c.chunk.WriteSimple(OpLoadTrue, line)
		} else {
			c.chunk.WriteSimple(OpLoadFalse, line)
		}
		return nil
	default:
		return c.emitLoadConstant(value, line)
	}
}

func (c *Compiler) declareLocal(ident *ast.Identifier, typ types.Type) (uint16, error) {
	if _, exists := c.resolveLocalInCurrentScope(ident.Value); exists {
		return 0, c.errorf(ident, "duplicate variable %q in current scope", ident.Value)
	}

	slot := c.nextSlot
	c.nextSlot++
	if slot+1 > c.maxSlot {
		c.maxSlot = slot + 1
	}

	c.locals = append(c.locals, local{
		name:  ident.Value,
		depth: c.scopeDepth,
		slot:  slot,
		typ:   typ,
	})

	return slot, nil
}

func (c *Compiler) declareSyntheticLocal(name string) (uint16, error) {
	ident := &ast.Identifier{Value: name}
	return c.declareLocal(ident, nil)
}

func (c *Compiler) declareGlobal(ident *ast.Identifier, typ types.Type) (uint16, error) {
	if c.globals == nil {
		c.globals = make(map[string]globalVar)
	}

	key := strings.ToLower(ident.Value)
	if _, exists := c.globals[key]; exists {
		return 0, c.errorf(ident, "duplicate global variable %q", ident.Value)
	}

	index := c.nextGlobal
	c.nextGlobal++

	c.globals[key] = globalVar{
		name:  ident.Value,
		index: index,
		typ:   typ,
	}

	return index, nil
}

func (c *Compiler) resolveLocal(name string) (local, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if strings.EqualFold(c.locals[i].name, name) {
			return c.locals[i], true
		}
	}
	return local{}, false
}

func (c *Compiler) resolveLocalInCurrentScope(name string) (local, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		loc := c.locals[i]
		if loc.depth != c.scopeDepth {
			break
		}
		if strings.EqualFold(loc.name, name) {
			return loc, true
		}
	}
	return local{}, false
}

func (c *Compiler) resolveGlobal(name string) (globalVar, bool) {
	if c.globals == nil {
		return globalVar{}, false
	}
	g, ok := c.globals[strings.ToLower(name)]
	return g, ok
}

func (c *Compiler) resolveUpvalue(name string) (uint16, bool, error) {
	if c.enclosing == nil {
		return 0, false, nil
	}

	if localInfo, ok := c.enclosing.resolveLocal(name); ok {
		return c.addUpvalue(localInfo.slot, true)
	}

	upvalueIndex, ok, err := c.enclosing.resolveUpvalue(name)
	if err != nil || !ok {
		return 0, ok, err
	}
	return c.addUpvalue(upvalueIndex, false)
}

func (c *Compiler) addUpvalue(index uint16, isLocal bool) (uint16, bool, error) {
	for i, uv := range c.upvalues {
		if uv.index == index && uv.isLocal == isLocal {
			return uint16(i), true, nil
		}
	}

	if len(c.upvalues) >= 0xFF {
		return 0, false, c.errorf(nil, "too many upvalues (max 255)")
	}

	c.upvalues = append(c.upvalues, upvalue{
		index:   index,
		isLocal: isLocal,
	})

	return uint16(len(c.upvalues) - 1), true, nil
}

func (c *Compiler) buildUpvalueDefs() []UpvalueDef {
	if len(c.upvalues) == 0 {
		return nil
	}
	defs := make([]UpvalueDef, len(c.upvalues))
	for i, uv := range c.upvalues {
		defs[i] = UpvalueDef{
			IsLocal: uv.isLocal,
			Index:   int(uv.index),
		}
	}
	return defs
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope() {
	if c.scopeDepth == 0 {
		return
	}

	for len(c.locals) > 0 && c.locals[len(c.locals)-1].depth == c.scopeDepth {
		c.locals = c.locals[:len(c.locals)-1]
	}
	c.scopeDepth--
}

func (c *Compiler) initBuiltins() {
	c.addBuiltinGlobal(builtinExceptObjectName)
	// Register built-in functions as globals
	c.addBuiltinGlobal("PrintLn")
	c.addBuiltinGlobal("Print")
	c.addBuiltinGlobal("IntToStr")
	c.addBuiltinGlobal("FloatToStr")
	c.addBuiltinGlobal("StrToInt")
	c.addBuiltinGlobal("StrToFloat")
	c.addBuiltinGlobal("StrToIntDef")
	c.addBuiltinGlobal("StrToFloatDef")
	c.addBuiltinGlobal("Length")
	c.addBuiltinGlobal("Copy")
	c.addBuiltinGlobal("SubStr")
	c.addBuiltinGlobal("SubString")
	c.addBuiltinGlobal("LeftStr")
	c.addBuiltinGlobal("RightStr")
	c.addBuiltinGlobal("MidStr")
	c.addBuiltinGlobal("StrBeginsWith")
	c.addBuiltinGlobal("StrEndsWith")
	c.addBuiltinGlobal("StrContains")
	c.addBuiltinGlobal("PosEx")
	c.addBuiltinGlobal("RevPos")
	c.addBuiltinGlobal("StrFind")
	c.addBuiltinGlobal("Ord")
	c.addBuiltinGlobal("Chr")
	// Type cast functions
	c.addBuiltinGlobal("Integer")
	c.addBuiltinGlobal("Float")
	c.addBuiltinGlobal("String")
	c.addBuiltinGlobal("Boolean")
	// Math functions (Pi is a constant, handled by semantic analyzer)
	c.addBuiltinGlobal("Sign")
	c.addBuiltinGlobal("Odd")
	c.addBuiltinGlobal("Frac")
	c.addBuiltinGlobal("Int")
	c.addBuiltinGlobal("Log10")
	c.addBuiltinGlobal("LogN")

	// MEDIUM PRIORITY Math Functions
	c.addBuiltinGlobal("Infinity")
	c.addBuiltinGlobal("NaN")
	c.addBuiltinGlobal("IsFinite")
	c.addBuiltinGlobal("IsInfinite")
	c.addBuiltinGlobal("IntPower")
	c.addBuiltinGlobal("RandSeed")
	c.addBuiltinGlobal("RandG")
	c.addBuiltinGlobal("SetRandSeed")
	c.addBuiltinGlobal("Randomize")

	// Advanced Math Functions (Phase 9.23)
	c.addBuiltinGlobal("Factorial")
	c.addBuiltinGlobal("Gcd")
	c.addBuiltinGlobal("Lcm")
	c.addBuiltinGlobal("IsPrime")
	c.addBuiltinGlobal("LeastFactor")
	c.addBuiltinGlobal("PopCount")
	c.addBuiltinGlobal("TestBit")
}

func (c *Compiler) addBuiltinGlobal(name string) {
	if c.globals == nil {
		c.globals = make(map[string]globalVar)
	}
	key := strings.ToLower(name)
	if _, exists := c.globals[key]; exists {
		return
	}
	c.globals[key] = globalVar{
		name:  name,
		index: c.nextGlobal,
	}
	c.nextGlobal++
}

func (c *Compiler) propertyNameIndex(name string, node ast.Node) (uint16, error) {
	index := c.chunk.AddConstant(StringValue(name))
	if index > 0xFFFF {
		return 0, c.errorf(node, "constant pool overflow")
	}
	return uint16(index), nil
}

func (c *Compiler) pushLoop(kind loopKind, loopStart int) *loopContext {
	ctx := &loopContext{
		kind:      kind,
		loopStart: loopStart,
	}
	c.loopStack = append(c.loopStack, ctx)
	return ctx
}

func (c *Compiler) popLoop() {
	if len(c.loopStack) == 0 {
		return
	}
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

func (c *Compiler) currentLoop() *loopContext {
	if len(c.loopStack) == 0 {
		return nil
	}
	return c.loopStack[len(c.loopStack)-1]
}

func (c *Compiler) patchLoopBreaks(ctx *loopContext) error {
	if ctx == nil || len(ctx.breakJumps) == 0 {
		return nil
	}
	target := len(c.chunk.Code)
	for _, idx := range ctx.breakJumps {
		if err := c.patchJumpToTarget(idx, target); err != nil {
			return err
		}
	}
	ctx.breakJumps = ctx.breakJumps[:0]
	return nil
}

func (c *Compiler) patchLoopContinues(ctx *loopContext, target int) error {
	if ctx == nil || len(ctx.continueJumps) == 0 {
		return nil
	}
	for _, idx := range ctx.continueJumps {
		if err := c.patchJumpToTarget(idx, target); err != nil {
			return err
		}
	}
	ctx.continueJumps = ctx.continueJumps[:0]
	return nil
}

func (c *Compiler) patchJumpToTarget(jumpIndex, target int) error {
	offset := target - jumpIndex - 1
	if offset > 32767 || offset < -32768 {
		return c.errorf(nil, "jump offset too large: %d", offset)
	}

	inst := c.chunk.Code[jumpIndex]
	c.chunk.Code[jumpIndex] = MakeInstruction(inst.OpCode(), inst.A(), uint16(offset))
	return nil
}

func (c *Compiler) ensureFunctionReturn(line int) {
	if len(c.chunk.Code) == 0 {
		c.chunk.Write(OpReturn, 0, 0, line)
		return
	}
	last := c.chunk.Code[len(c.chunk.Code)-1]
	if last.OpCode() != OpReturn {
		c.chunk.Write(OpReturn, 0, 0, line)
	}
}

func (c *Compiler) loadExceptObject(line int) error {
	global, ok := c.resolveGlobal(builtinExceptObjectName)
	if !ok {
		return c.errorf(nil, "%s global not initialized", builtinExceptObjectName)
	}
	c.chunk.Write(OpLoadGlobal, 0, global.index, line)
	return nil
}

func (c *Compiler) hasEnclosingLocal(name string) bool {
	for env := c.enclosing; env != nil; env = env.enclosing {
		if _, ok := env.resolveLocal(name); ok {
			return true
		}
	}
	return false
}

func (c *Compiler) inferExpressionType(expr ast.Expression) types.Type {
	if expr == nil {
		return nil
	}

	switch node := expr.(type) {
	case ast.TypedExpression:
		return typeFromAnnotation(node.GetType())
	case *ast.IntegerLiteral:
		return types.INTEGER
	case *ast.FloatLiteral:
		return types.FLOAT
	case *ast.StringLiteral:
		return types.STRING
	case *ast.BooleanLiteral:
		return types.BOOLEAN
	case *ast.NilLiteral:
		return types.NIL
	case *ast.Identifier:
		if localInfo, ok := c.resolveLocal(node.Value); ok {
			return localInfo.typ
		}
		if globalInfo, ok := c.resolveGlobal(node.Value); ok {
			return globalInfo.typ
		}
		return typeFromAnnotation(node.GetType())
	default:
		return nil
	}
}

func (c *Compiler) isGlobalScope() bool {
	return c.enclosing == nil && c.scopeDepth == 0
}

func (c *Compiler) needsHalt() bool {
	if len(c.chunk.Code) == 0 {
		return true
	}

	last := c.chunk.Code[len(c.chunk.Code)-1]
	return last.OpCode() != OpReturn && last.OpCode() != OpHalt
}

func (c *Compiler) errorf(node ast.Node, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	if node != nil {
		pos := node.Pos()
		if pos.Line > 0 {
			message = fmt.Sprintf("%s at %d:%d", message, pos.Line, pos.Column)
		}
	}
	return fmt.Errorf("bytecode compile error: %s", message)
}

func lineOf(node ast.Node) int {
	if node == nil {
		return 0
	}
	pos := node.Pos()
	return pos.Line
}

func typeFromAnnotation(annotation *ast.TypeAnnotation) types.Type {
	if annotation == nil {
		return nil
	}

	switch strings.ToLower(annotation.Name) {
	case "integer":
		return types.INTEGER
	case "float":
		return types.FLOAT
	case "string":
		return types.STRING
	case "boolean":
		return types.BOOLEAN
	case "variant":
		return types.VARIANT
	case "nil":
		return types.NIL
	case "void":
		return types.VOID
	default:
		return nil
	}
}

func lineOfExceptClause(clause *ast.ExceptClause) int {
	if clause == nil {
		return 0
	}
	for _, handler := range clause.Handlers {
		if handler != nil && handler.Statement != nil {
			if line := lineOf(handler.Statement); line != 0 {
				return line
			}
		}
	}
	if clause.ElseBlock != nil {
		return lineOf(clause.ElseBlock)
	}
	return 0
}

func isFloatType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.GetUnderlyingType(t).TypeKind() == types.FLOAT.TypeKind()
}

func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.GetUnderlyingType(t).TypeKind() == types.STRING.TypeKind()
}

func literalValue(expr ast.Expression) (Value, bool) {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return IntValue(node.Value), true
	case *ast.FloatLiteral:
		return FloatValue(node.Value), true
	case *ast.StringLiteral:
		return StringValue(node.Value), true
	case *ast.BooleanLiteral:
		return BoolValue(node.Value), true
	case *ast.NilLiteral:
		return NilValue(), true
	case *ast.UnaryExpression:
		operand, ok := literalValue(node.Right)
		if !ok {
			return Value{}, false
		}
		return evaluateUnary(node.Operator, operand)
	default:
		return Value{}, false
	}
}

func evaluateBinary(operator string, left, right Value) (Value, bool) {
	opLower := strings.ToLower(operator)

	// Try arithmetic operations
	if result, ok := evaluateBinaryArithmetic(opLower, left, right); ok {
		return result, true
	}

	// Try comparison operations
	if result, ok := evaluateBinaryComparison(opLower, left, right); ok {
		return result, true
	}

	// Try logical operations
	if result, ok := evaluateBinaryLogical(opLower, left, right); ok {
		return result, true
	}

	return Value{}, false
}

// evaluateBinaryArithmetic evaluates arithmetic operations (+, -, *, /, div, mod)
func evaluateBinaryArithmetic(operator string, left, right Value) (Value, bool) {
	switch operator {
	case "+":
		if left.IsString() && right.IsString() {
			return StringValue(left.AsString() + right.AsString()), true
		}
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() + right.AsFloat()), true
			}
			return IntValue(left.AsInt() + right.AsInt()), true
		}
	case "-":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() - right.AsFloat()), true
			}
			return IntValue(left.AsInt() - right.AsInt()), true
		}
	case "*":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return FloatValue(left.AsFloat() * right.AsFloat()), true
			}
			return IntValue(left.AsInt() * right.AsInt()), true
		}
	case "div":
		if left.Type == ValueInt && right.Type == ValueInt {
			divisor := right.AsInt()
			if divisor == 0 {
				return Value{}, false
			}
			return IntValue(left.AsInt() / divisor), true
		}
	case "/":
		if left.IsNumber() && right.IsNumber() {
			divisor := right.AsFloat()
			if divisor == 0 {
				return Value{}, false
			}
			return FloatValue(left.AsFloat() / divisor), true
		}
	case "mod":
		if left.Type == ValueInt && right.Type == ValueInt {
			divisor := right.AsInt()
			if divisor == 0 {
				return Value{}, false
			}
			return IntValue(left.AsInt() % divisor), true
		}
	}
	return Value{}, false
}

// evaluateBinaryComparison evaluates comparison operations (=, <>, <, <=, >, >=)
func evaluateBinaryComparison(operator string, left, right Value) (Value, bool) {
	switch operator {
	case "=":
		if eq, ok := valuesEqualForFold(left, right); ok {
			return BoolValue(eq), true
		}
	case "<>":
		if eq, ok := valuesEqualForFold(left, right); ok {
			return BoolValue(!eq), true
		}
	case "<":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() < right.AsFloat()), true
			}
			return BoolValue(left.AsInt() < right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() < right.AsString()), true
		}
	case "<=":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() <= right.AsFloat()), true
			}
			return BoolValue(left.AsInt() <= right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() <= right.AsString()), true
		}
	case ">":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() > right.AsFloat()), true
			}
			return BoolValue(left.AsInt() > right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() > right.AsString()), true
		}
	case ">=":
		if left.IsNumber() && right.IsNumber() {
			if left.Type == ValueFloat || right.Type == ValueFloat {
				return BoolValue(left.AsFloat() >= right.AsFloat()), true
			}
			return BoolValue(left.AsInt() >= right.AsInt()), true
		}
		if left.IsString() && right.IsString() {
			return BoolValue(left.AsString() >= right.AsString()), true
		}
	}
	return Value{}, false
}

// evaluateBinaryLogical evaluates logical operations (and, or, xor)
func evaluateBinaryLogical(operator string, left, right Value) (Value, bool) {
	switch operator {
	case "and":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() && right.AsBool()), true
		}
	case "or":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() || right.AsBool()), true
		}
	case "xor":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() != right.AsBool()), true
		}
	}
	return Value{}, false
}

func evaluateUnary(operator string, operand Value) (Value, bool) {
	switch strings.ToLower(operator) {
	case "+":
		if operand.IsNumber() {
			return operand, true
		}
	case "-":
		if operand.Type == ValueFloat {
			return FloatValue(-operand.AsFloat()), true
		}
		if operand.Type == ValueInt {
			return IntValue(-operand.AsInt()), true
		}
	case "not":
		if operand.Type == ValueBool {
			return BoolValue(!operand.AsBool()), true
		}
	}
	return Value{}, false
}

func valuesEqualForFold(left, right Value) (bool, bool) {
	if left.Type == right.Type {
		switch left.Type {
		case ValueNil:
			return true, true
		case ValueBool:
			return left.AsBool() == right.AsBool(), true
		case ValueInt:
			return left.AsInt() == right.AsInt(), true
		case ValueFloat:
			return left.AsFloat() == right.AsFloat(), true
		case ValueString:
			return left.AsString() == right.AsString(), true
		default:
			return false, false
		}
	}

	if left.IsNumber() && right.IsNumber() {
		return left.AsFloat() == right.AsFloat(), true
	}

	return false, false
}
