package bytecode

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// Compiler converts AST nodes into bytecode chunks.
type Compiler struct {
	chunk      *Chunk
	locals     []local
	scopeDepth int
	nextSlot   uint16
	maxSlot    uint16
	lastLine   int
}

type local struct {
	name  string
	depth int
	slot  uint16
	typ   types.Type
}

// NewCompiler creates a compiler for the given chunk name.
func NewCompiler(chunkName string) *Compiler {
	return &Compiler{
		chunk: NewChunk(chunkName),
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
	c.scopeDepth = 0
	c.nextSlot = 0
	c.maxSlot = 0

	if err := c.compileProgram(program); err != nil {
		return nil, err
	}

	c.chunk.LocalCount = int(c.maxSlot)

	if c.needsHalt() {
		c.chunk.WriteSimple(OpHalt, c.lastLine)
	}

	return c.chunk, nil
}

// LocalCount returns the number of local slots allocated during compilation.
func (c *Compiler) LocalCount() int {
	return int(c.maxSlot)
}

func (c *Compiler) compileProgram(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	if stmt == nil {
		return nil
	}

	c.lastLine = lineOf(stmt)

	switch node := stmt.(type) {
	case *ast.BlockStatement:
		return c.compileBlock(node)
	case *ast.VarDeclStatement:
		return c.compileVarDecl(node)
	case *ast.AssignmentStatement:
		return c.compileAssignment(node)
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(node)
	case *ast.IfStatement:
		return c.compileIf(node)
	case *ast.WhileStatement:
		return c.compileWhile(node)
	case *ast.RepeatStatement:
		return c.compileRepeat(node)
	case *ast.ReturnStatement:
		return c.compileReturn(node)
	default:
		return c.errorf(stmt, "unsupported statement type %T", stmt)
	}
}

func (c *Compiler) compileBlock(block *ast.BlockStatement) error {
	c.beginScope()
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.endScope()
	return nil
}

func (c *Compiler) compileVarDecl(stmt *ast.VarDeclStatement) error {
	if len(stmt.Names) == 0 {
		return c.errorf(stmt, "variable declaration without names")
	}

	for _, name := range stmt.Names {
		localType := typeFromAnnotation(stmt.Type)
		if localType == nil && stmt.Value != nil {
			localType = c.inferExpressionType(stmt.Value)
		}
		slot, err := c.declareLocal(name, localType)
		if err != nil {
			return err
		}

		if stmt.Value != nil {
			if err := c.compileExpression(stmt.Value); err != nil {
				return err
			}
		} else {
			c.chunk.WriteSimple(OpLoadNil, lineOf(stmt))
		}

		c.chunk.Write(OpStoreLocal, 0, slot, lineOf(name))
	}

	return nil
}

func (c *Compiler) compileAssignment(stmt *ast.AssignmentStatement) error {
	targetIdent, ok := stmt.Target.(*ast.Identifier)
	if !ok {
		return c.errorf(stmt.Target, "unsupported assignment target %T", stmt.Target)
	}

	localInfo, ok := c.resolveLocal(targetIdent.Value)
	if !ok {
		return c.errorf(stmt.Target, "unknown variable %q", targetIdent.Value)
	}

	if stmt.Operator != lexer.ASSIGN {
		return c.errorf(stmt, "unsupported assignment operator %s", stmt.Operator)
	}

	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	c.chunk.Write(OpStoreLocal, 0, localInfo.slot, lineOf(stmt))
	return nil
}

func (c *Compiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	if stmt.Expression == nil {
		return nil
	}

	if err := c.compileExpression(stmt.Expression); err != nil {
		return err
	}

	c.chunk.WriteSimple(OpPop, lineOf(stmt))
	return nil
}

func (c *Compiler) compileIf(stmt *ast.IfStatement) error {
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	jumpIfFalse := c.chunk.EmitJump(OpJumpIfFalse, lineOf(stmt.Condition))

	if err := c.compileStatement(stmt.Consequence); err != nil {
		return err
	}

	if stmt.Alternative != nil {
		jumpToEnd := c.chunk.EmitJump(OpJump, lineOf(stmt))
		if err := c.chunk.PatchJump(jumpIfFalse); err != nil {
			return err
		}

		if err := c.compileStatement(stmt.Alternative); err != nil {
			return err
		}

		return c.chunk.PatchJump(jumpToEnd)
	}

	return c.chunk.PatchJump(jumpIfFalse)
}

func (c *Compiler) compileWhile(stmt *ast.WhileStatement) error {
	loopStart := len(c.chunk.Code)

	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	exitJump := c.chunk.EmitJump(OpJumpIfFalse, lineOf(stmt.Condition))

	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	if err := c.chunk.EmitLoop(loopStart, lineOf(stmt)); err != nil {
		return err
	}

	return c.chunk.PatchJump(exitJump)
}

func (c *Compiler) compileRepeat(stmt *ast.RepeatStatement) error {
	loopStart := len(c.chunk.Code)

	if err := c.compileStatement(stmt.Body); err != nil {
		return err
	}

	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	exitJump := c.chunk.EmitJump(OpJumpIfTrue, lineOf(stmt.Condition))

	if err := c.chunk.EmitLoop(loopStart, lineOf(stmt)); err != nil {
		return err
	}

	return c.chunk.PatchJump(exitJump)
}

func (c *Compiler) compileReturn(stmt *ast.ReturnStatement) error {
	if stmt.ReturnValue != nil {
		if err := c.compileExpression(stmt.ReturnValue); err != nil {
			return err
		}
		c.chunk.Write(OpReturn, 1, 0, lineOf(stmt))
		return nil
	}

	c.chunk.Write(OpReturn, 0, 0, lineOf(stmt))
	return nil
}

func (c *Compiler) compileExpression(expr ast.Expression) error {
	if expr == nil {
		return nil
	}

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return c.emitLoadConstant(IntValue(node.Value), lineOf(node))
	case *ast.FloatLiteral:
		return c.emitLoadConstant(FloatValue(node.Value), lineOf(node))
	case *ast.StringLiteral:
		return c.emitLoadConstant(StringValue(node.Value), lineOf(node))
	case *ast.BooleanLiteral:
		if node.Value {
			c.chunk.WriteSimple(OpLoadTrue, lineOf(node))
		} else {
			c.chunk.WriteSimple(OpLoadFalse, lineOf(node))
		}
		return nil
	case *ast.NilLiteral:
		c.chunk.WriteSimple(OpLoadNil, lineOf(node))
		return nil
	case *ast.Identifier:
		return c.compileIdentifier(node)
	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(node)
	case *ast.UnaryExpression:
		return c.compileUnaryExpression(node)
	default:
		return c.errorf(expr, "unsupported expression type %T", expr)
	}
}

func (c *Compiler) compileIdentifier(ident *ast.Identifier) error {
	localInfo, ok := c.resolveLocal(ident.Value)
	if !ok {
		return c.errorf(ident, "unknown identifier %q", ident.Value)
	}

	c.chunk.Write(OpLoadLocal, 0, localInfo.slot, lineOf(ident))
	return nil
}

func (c *Compiler) compileCallExpression(expr *ast.CallExpression) error {
	if err := c.compileExpression(expr.Function); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	argCount := len(expr.Arguments)
	if argCount > 0xFF {
		return c.errorf(expr, "too many arguments in function call: %d", argCount)
	}

	c.chunk.Write(OpCallIndirect, byte(argCount), 0, lineOf(expr))
	return nil
}

func (c *Compiler) compileBinaryExpression(expr *ast.BinaryExpression) error {
	if folded, err := c.tryFoldBinaryExpression(expr); folded {
		return err
	}

	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	resultType := c.inferExpressionType(expr)
	if resultType == nil {
		resultType = c.inferExpressionType(expr.Left)
	}
	if resultType == nil {
		resultType = c.inferExpressionType(expr.Right)
	}

	line := lineOf(expr)
	switch strings.ToLower(expr.Operator) {
	case "+":
		return c.emitNumericBinaryOp(resultType, line, OpAddInt, OpAddFloat, OpStringConcat)
	case "-":
		return c.emitNumericBinaryOp(resultType, line, OpSubInt, OpSubFloat, 0)
	case "*":
		return c.emitNumericBinaryOp(resultType, line, OpMulInt, OpMulFloat, 0)
	case "div":
		return c.emitNumericBinaryOp(resultType, line, OpDivInt, 0, 0)
	case "/":
		return c.emitNumericBinaryOp(resultType, line, 0, OpDivFloat, 0)
	case "mod":
		return c.emitNumericBinaryOp(resultType, line, OpModInt, 0, 0)
	case "=":
		c.chunk.WriteSimple(OpEqual, line)
	case "<>":
		c.chunk.WriteSimple(OpNotEqual, line)
	case "<":
		c.chunk.WriteSimple(OpLess, line)
	case "<=":
		c.chunk.WriteSimple(OpLessEqual, line)
	case ">":
		c.chunk.WriteSimple(OpGreater, line)
	case ">=":
		c.chunk.WriteSimple(OpGreaterEqual, line)
	case "and":
		c.chunk.WriteSimple(OpAnd, line)
	case "or":
		c.chunk.WriteSimple(OpOr, line)
	default:
		return c.errorf(expr, "unsupported binary operator %q", expr.Operator)
	}

	return nil
}

func (c *Compiler) compileUnaryExpression(expr *ast.UnaryExpression) error {
	if folded, err := c.tryFoldUnaryExpression(expr); folded {
		return err
	}

	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	exprType := c.inferExpressionType(expr)
	if exprType == nil {
		exprType = c.inferExpressionType(expr.Right)
	}

	line := lineOf(expr)
	switch strings.ToLower(expr.Operator) {
	case "+":
		return nil
	case "-":
		if isFloatType(exprType) {
			c.chunk.WriteSimple(OpNegateFloat, line)
		} else {
			c.chunk.WriteSimple(OpNegateInt, line)
		}
	case "not":
		c.chunk.WriteSimple(OpNot, line)
	default:
		return c.errorf(expr, "unsupported unary operator %q", expr.Operator)
	}

	return nil
}

func (c *Compiler) tryFoldBinaryExpression(expr *ast.BinaryExpression) (bool, error) {
	leftVal, leftOk := literalValue(expr.Left)
	rightVal, rightOk := literalValue(expr.Right)
	if !leftOk || !rightOk {
		return false, nil
	}

	result, ok := evaluateBinary(expr.Operator, leftVal, rightVal)
	if !ok {
		return false, nil
	}

	return true, c.emitValue(result, lineOf(expr))
}

func (c *Compiler) tryFoldUnaryExpression(expr *ast.UnaryExpression) (bool, error) {
	operandVal, ok := literalValue(expr.Right)
	if !ok {
		return false, nil
	}

	result, ok := evaluateUnary(expr.Operator, operandVal)
	if !ok {
		return false, nil
	}

	return true, c.emitValue(result, lineOf(expr))
}

func (c *Compiler) emitNumericBinaryOp(resultType types.Type, line int, intOp, floatOp OpCode, stringOp OpCode) error {
	switch {
	case stringOp != 0 && isStringType(resultType):
		c.chunk.WriteSimple(stringOp, line)
	case isFloatType(resultType):
		if floatOp == 0 {
			return c.errorf(nil, "no float opcode available for operation")
		}
		c.chunk.WriteSimple(floatOp, line)
	default:
		if intOp == 0 {
			return c.errorf(nil, "no integer opcode available for operation")
		}
		c.chunk.WriteSimple(intOp, line)
	}
	return nil
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
		return typeFromAnnotation(node.GetType())
	default:
		return nil
	}
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
	switch strings.ToLower(operator) {
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
	case "and":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() && right.AsBool()), true
		}
	case "or":
		if left.Type == ValueBool && right.Type == ValueBool {
			return BoolValue(left.AsBool() || right.AsBool()), true
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
