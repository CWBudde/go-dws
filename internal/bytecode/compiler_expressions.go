package bytecode

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

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
	case *ast.ArrayLiteralExpression:
		return c.compileArrayLiteral(node)
	case *ast.IndexExpression:
		return c.compileIndexExpression(node)
	case *ast.MemberAccessExpression:
		return c.compileMemberAccess(node)
	case *ast.NewExpression:
		return c.compileNewExpression(node)
	case *ast.LambdaExpression:
		return c.compileLambdaExpression(node)
	case *ast.MethodCallExpression:
		return c.compileMethodCallExpression(node)
	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(node)
	case *ast.UnaryExpression:
		return c.compileUnaryExpression(node)
	case *ast.IfExpression:
		return c.compileIfExpression(node)
	case *ast.IsExpression:
		return c.compileIsExpression(node)
	default:
		return c.errorf(expr, "unsupported expression type %T", expr)
	}
}

func (c *Compiler) compileIdentifier(ident *ast.Identifier) error {
	localInfo, ok := c.resolveLocal(ident.Value)
	if !ok {
		if uvIndex, ok, err := c.resolveUpvalue(ident.Value); err != nil {
			return err
		} else if ok {
			c.chunk.Write(OpLoadUpvalue, 0, uvIndex, lineOf(ident))
			return nil
		}
		if globalInfo, found := c.resolveGlobal(ident.Value); found {
			c.chunk.Write(OpLoadGlobal, 0, globalInfo.index, lineOf(ident))
			return nil
		}
		if strings.EqualFold(ident.Value, "Self") {
			c.chunk.WriteSimple(OpGetSelf, lineOf(ident))
			return nil
		}
		return c.errorf(ident, "unknown identifier %q", ident.Value)
	}

	c.chunk.Write(OpLoadLocal, 0, localInfo.slot, lineOf(ident))
	return nil
}

func (c *Compiler) compileIdentifierAssignment(ident *ast.Identifier, value ast.Expression) error {
	if err := c.compileExpression(value); err != nil {
		return err
	}

	if strings.EqualFold(ident.Value, builtinExceptObjectName) {
		return c.errorf(ident, "cannot assign to %s", builtinExceptObjectName)
	}

	if localInfo, ok := c.resolveLocal(ident.Value); ok {
		c.chunk.Write(OpStoreLocal, 0, localInfo.slot, lineOf(ident))
		return nil
	}

	if uvIndex, ok, err := c.resolveUpvalue(ident.Value); err != nil {
		return err
	} else if ok {
		c.chunk.Write(OpStoreUpvalue, 0, uvIndex, lineOf(ident))
		return nil
	}

	if globalInfo, ok := c.resolveGlobal(ident.Value); ok {
		c.chunk.Write(OpStoreGlobal, 0, globalInfo.index, lineOf(ident))
		return nil
	}

	if strings.EqualFold(ident.Value, "Self") {
		return c.errorf(ident, "cannot assign to Self")
	}

	return c.errorf(ident, "unknown variable %q", ident.Value)
}

func (c *Compiler) compileMemberAccess(expr *ast.MemberAccessExpression) error {
	if expr == nil || expr.Member == nil {
		return c.errorf(expr, "invalid member access expression")
	}

	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	nameIndex, err := c.propertyNameIndex(expr.Member.Value, expr)
	if err != nil {
		return err
	}

	c.chunk.Write(OpGetProperty, 0, nameIndex, lineOf(expr))
	return nil
}

func (c *Compiler) compileMemberAssignment(target *ast.MemberAccessExpression, value ast.Expression) error {
	if target == nil || target.Member == nil {
		return c.errorf(target, "invalid member assignment target")
	}

	if err := c.compileExpression(target.Object); err != nil {
		return err
	}

	if err := c.compileExpression(value); err != nil {
		return err
	}

	nameIndex, err := c.propertyNameIndex(target.Member.Value, target)
	if err != nil {
		return err
	}

	c.chunk.Write(OpSetProperty, 0, nameIndex, lineOf(target))
	return nil
}

func (c *Compiler) compileArrayLiteral(expr *ast.ArrayLiteralExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil array literal expression")
	}
	for _, element := range expr.Elements {
		if err := c.compileExpression(element); err != nil {
			return err
		}
	}
	count := len(expr.Elements)
	if count > 0xFFFF {
		return c.errorf(expr, "array literal has too many elements (%d)", count)
	}
	c.chunk.Write(OpNewArray, 0, uint16(count), lineOf(expr))
	return nil
}

func (c *Compiler) compileIndexExpression(expr *ast.IndexExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil index expression")
	}
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	if err := c.compileExpression(expr.Index); err != nil {
		return err
	}
	c.chunk.WriteSimple(OpArrayGet, lineOf(expr))
	return nil
}

func (c *Compiler) compileIndexAssignment(target *ast.IndexExpression, value ast.Expression) error {
	if target == nil {
		return c.errorf(nil, "nil index assignment target")
	}
	if value != nil {
		if err := c.compileExpression(value); err != nil {
			return err
		}
	} else {
		c.chunk.WriteSimple(OpLoadNil, lineOf(target))
	}
	if err := c.compileExpression(target.Left); err != nil {
		return err
	}
	if err := c.compileExpression(target.Index); err != nil {
		return err
	}
	c.chunk.WriteSimple(OpRotate3, lineOf(target))
	c.chunk.WriteSimple(OpArraySet, lineOf(target))
	return nil
}

func (c *Compiler) compileNewExpression(expr *ast.NewExpression) error {
	if expr == nil || expr.ClassName == nil {
		return c.errorf(expr, "invalid new expression")
	}
	if len(expr.Arguments) > 0 {
		return c.errorf(expr, "constructors with arguments are not supported in bytecode yet")
	}
	constIdx := c.chunk.AddConstant(StringValue(expr.ClassName.Value))
	if constIdx > 0xFFFF {
		return c.errorf(expr, "constant pool overflow")
	}
	c.chunk.Write(OpNewObject, 0, uint16(constIdx), lineOf(expr))
	return nil
}

func (c *Compiler) compileMethodCallExpression(expr *ast.MethodCallExpression) error {
	if expr == nil || expr.Method == nil {
		return c.errorf(expr, "invalid method call expression")
	}

	if err := c.compileExpression(expr.Object); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	argCount := len(expr.Arguments)
	if argCount > 0xFF {
		return c.errorf(expr, "too many arguments in method call: %d", argCount)
	}

	nameIndex, err := c.propertyNameIndex(expr.Method.Value, expr)
	if err != nil {
		return err
	}

	c.chunk.Write(OpCallMethod, byte(argCount), nameIndex, lineOf(expr))
	return nil
}

func (c *Compiler) compileBinaryExpression(expr *ast.BinaryExpression) error {
	if folded, err := c.tryFoldBinaryExpression(expr); folded {
		return err
	}

	// Special handling for coalesce operator (??) - requires short-circuit evaluation
	if strings.ToLower(expr.Operator) == "??" {
		return c.compileCoalesceExpression(expr)
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

// compileCoalesceExpression compiles the coalesce operator (??) with short-circuit evaluation.
// The bytecode:
//  1. Compile left operand (leaves value on stack)
//  2. Duplicate the value (so we can check it without losing it)
//  3. Check if it's falsey using OpIsFalsey
//  4. If not falsey (i.e., truthy), jump to end (keep left on stack)
//  5. If falsey, pop the left value, compile right operand
//  6. End: result (either left or right) is on top of stack
func (c *Compiler) compileCoalesceExpression(expr *ast.BinaryExpression) error {
	line := lineOf(expr)

	// Compile left operand
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}

	// Duplicate the left value so we can check it without consuming it
	c.chunk.WriteSimple(OpDup, line)

	// Check if the duplicated value is falsey
	c.chunk.WriteSimple(OpIsFalsey, line)

	// If falsey (true), we need to evaluate right operand
	// So jump if NOT falsey (jump if the result is false)
	// This leaves the left value on stack if it's truthy
	jumpIfTruthy := c.chunk.EmitJump(OpJumpIfFalse, line)

	// The value is falsey, so pop it and evaluate right operand
	c.chunk.WriteSimple(OpPop, line)
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// Patch the jump to skip right evaluation when left is truthy
	return c.chunk.PatchJump(jumpIfTruthy)
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

// compileIfExpression compiles an inline if-then-else conditional expression.
// Pattern:
//
//	COMPILE condition
//	JumpIfFalse elseLabel
//	COMPILE consequence
//	Jump endLabel
//	elseLabel:
//	COMPILE alternative (or emit default value)
//	endLabel:
//
// Both branches must leave exactly one value on the stack.
func (c *Compiler) compileIfExpression(expr *ast.IfExpression) error {
	// Compile the condition
	if err := c.compileExpression(expr.Condition); err != nil {
		return err
	}

	// Emit conditional jump - if false, jump to else branch
	jumpIfFalse := c.chunk.EmitJump(OpJumpIfFalse, lineOf(expr.Condition))

	// Compile the consequence (then branch)
	if err := c.compileExpression(expr.Consequence); err != nil {
		return err
	}

	// If there's an else clause or we need to emit a default value
	if expr.Alternative != nil {
		// Jump over the else branch when consequence is executed
		jumpToEnd := c.chunk.EmitJump(OpJump, lineOf(expr))

		// Patch the conditional jump to point here (else branch)
		if err := c.chunk.PatchJump(jumpIfFalse); err != nil {
			return err
		}

		// Compile the alternative (else branch)
		if err := c.compileExpression(expr.Alternative); err != nil {
			return err
		}

		// Patch the jump to end
		return c.chunk.PatchJump(jumpToEnd)
	}

	// No else clause - need to emit default value
	jumpToEnd := c.chunk.EmitJump(OpJump, lineOf(expr))

	// Patch the conditional jump to point here (default value emission)
	if err := c.chunk.PatchJump(jumpIfFalse); err != nil {
		return err
	}

	// Emit default value based on type
	if err := c.emitDefaultValue(expr, lineOf(expr)); err != nil {
		return err
	}

	// Patch the jump to end
	return c.chunk.PatchJump(jumpToEnd)
}

// emitDefaultValue emits bytecode to push a default value for the given expression type onto the stack.
func (c *Compiler) emitDefaultValue(expr *ast.IfExpression, line int) error {
	if expr.Type == nil {
		return c.errorf(expr, "if expression missing type annotation for default value")
	}

	// Get type name and emit appropriate default value
	typeName := strings.ToLower(expr.Type.Name)
	switch typeName {
	case "integer", "int64":
		return c.emitLoadConstant(IntValue(0), line)
	case "float", "float64", "double", "real":
		return c.emitLoadConstant(FloatValue(0.0), line)
	case "string":
		return c.emitLoadConstant(StringValue(""), line)
	case "boolean", "bool":
		c.chunk.WriteSimple(OpLoadFalse, line)
		return nil
	default:
		// For class types and other reference types, emit nil
		c.chunk.WriteSimple(OpLoadNil, line)
		return nil
	}
}

func (c *Compiler) compileIsExpression(expr *ast.IsExpression) error {
	line := lineOf(expr)

	// Check if this is a boolean value comparison or type check
	if expr.Right != nil {
		// Boolean value comparison: left is right
		// Convert both operands to boolean before comparing to match interpreter behavior
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		c.chunk.WriteSimple(OpToBool, line)
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}
		c.chunk.WriteSimple(OpToBool, line)
		c.chunk.WriteSimple(OpEqual, line)
		return nil
	}

	// Type checking mode - not yet fully implemented in bytecode
	// For now, we'll return an error and let the interpreter handle it
	return c.errorf(expr, "type checking with 'is' operator not yet supported in bytecode mode")
}
