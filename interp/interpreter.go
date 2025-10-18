package interp

import (
	"fmt"
	"io"
	"strings"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// Interpreter executes DWScript AST nodes and manages the runtime environment.
type Interpreter struct {
	env         *Environment                 // The current execution environment
	output      io.Writer                    // Where to write output (e.g., from PrintLn)
	functions   map[string]*ast.FunctionDecl // Registry of user-defined functions
	classes     map[string]*ClassInfo        // Registry of class definitions
	currentNode ast.Node                     // Current AST node being evaluated (for error reporting)
}

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	env := NewEnvironment()
	return &Interpreter{
		env:       env,
		output:    output,
		functions: make(map[string]*ast.FunctionDecl),
		classes:   make(map[string]*ClassInfo),
	}
}

// Eval evaluates an AST node and returns its value.
// This is the main entry point for the interpreter.
func (i *Interpreter) Eval(node ast.Node) Value {
	// Track the current node for error reporting
	i.currentNode = node

	switch node := node.(type) {
	// Program
	case *ast.Program:
		return i.evalProgram(node)

	// Statements
	case *ast.ExpressionStatement:
		return i.Eval(node.Expression)

	case *ast.VarDeclStatement:
		return i.evalVarDeclStatement(node)

	case *ast.AssignmentStatement:
		return i.evalAssignmentStatement(node)

	case *ast.BlockStatement:
		return i.evalBlockStatement(node)

	case *ast.IfStatement:
		return i.evalIfStatement(node)

	case *ast.WhileStatement:
		return i.evalWhileStatement(node)

	case *ast.RepeatStatement:
		return i.evalRepeatStatement(node)

	case *ast.ForStatement:
		return i.evalForStatement(node)

	case *ast.CaseStatement:
		return i.evalCaseStatement(node)

	case *ast.FunctionDecl:
		return i.evalFunctionDeclaration(node)

	case *ast.ClassDecl:
		return i.evalClassDeclaration(node)

	// Expressions
	case *ast.IntegerLiteral:
		return &IntegerValue{Value: node.Value}

	case *ast.FloatLiteral:
		return &FloatValue{Value: node.Value}

	case *ast.StringLiteral:
		return &StringValue{Value: node.Value}

	case *ast.BooleanLiteral:
		return &BooleanValue{Value: node.Value}

	case *ast.NilLiteral:
		return &NilValue{}

	case *ast.Identifier:
		return i.evalIdentifier(node)

	case *ast.BinaryExpression:
		return i.evalBinaryExpression(node)

	case *ast.UnaryExpression:
		return i.evalUnaryExpression(node)

	case *ast.GroupedExpression:
		return i.Eval(node.Expression)

	case *ast.CallExpression:
		return i.evalCallExpression(node)

	case *ast.NewExpression:
		return i.evalNewExpression(node)

	case *ast.MemberAccessExpression:
		return i.evalMemberAccess(node)

	case *ast.MethodCallExpression:
		return i.evalMethodCall(node)

	default:
		return newError("unknown node type: %T", node)
	}
}

// evalProgram evaluates all statements in the program.
func (i *Interpreter) evalProgram(program *ast.Program) Value {
	var result Value

	for _, stmt := range program.Statements {
		result = i.Eval(stmt)

		// If we hit an error, stop execution
		if isError(result) {
			return result
		}
	}

	return result
}

// evalVarDeclStatement evaluates a variable declaration statement.
// It defines a new variable in the current environment.
func (i *Interpreter) evalVarDeclStatement(stmt *ast.VarDeclStatement) Value {
	var value Value

	if stmt.Value != nil {
		value = i.Eval(stmt.Value)
		if isError(value) {
			return value
		}
	} else {
		// No initializer, use nil
		value = &NilValue{}
	}

	i.env.Define(stmt.Name.Value, value)
	return value
}

// evalAssignmentStatement evaluates an assignment statement.
// It updates an existing variable's value or sets an object field.
func (i *Interpreter) evalAssignmentStatement(stmt *ast.AssignmentStatement) Value {
	value := i.Eval(stmt.Value)
	if isError(value) {
		return value
	}

	// Check if this is a member assignment (obj.field := value or TClass.Variable := value)
	// The parser encodes member assignments as "obj.field" in the Name field
	nameValue := stmt.Name.Value
	if strings.Contains(nameValue, ".") {
		// This is a member assignment: obj.field := value or TClass.Variable := value
		parts := strings.SplitN(nameValue, ".", 2)
		if len(parts) != 2 {
			return newError("invalid member assignment format: %s", nameValue)
		}

		objectName := parts[0]
		fieldName := parts[1]

		// Check if objectName refers to a class (static assignment: TClass.Variable := value)
		if classInfo, exists := i.classes[objectName]; exists {
			// This is a class variable assignment
			if _, exists := classInfo.ClassVars[fieldName]; !exists {
				return newError("class variable '%s' not found in class '%s'", fieldName, objectName)
			}
			// Assign to the class variable
			classInfo.ClassVars[fieldName] = value
			return value
		}

		// Not a class name - try object instance field assignment
		// Get the object from the environment
		objVal, ok := i.env.Get(objectName)
		if !ok {
			return newError("undefined variable: %s", objectName)
		}

		// Check if it's an object instance
		obj, ok := AsObject(objVal)
		if !ok {
			return newError("cannot assign to field of non-object type '%s'", objVal.Type())
		}

		// Verify field exists in the class
		if _, exists := obj.Class.Fields[fieldName]; !exists {
			return newError("field '%s' not found in class '%s'", fieldName, obj.Class.Name)
		}

		// Set the field value
		obj.SetField(fieldName, value)
		return value
	}

	// Simple variable assignment
	// First try to set in current environment
	err := i.env.Set(stmt.Name.Value, value)
	if err == nil {
		return value
	}

	// Not in environment - check if we're in a method context and this is a field/class variable
	// Check if Self is bound (instance method)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if _, exists := obj.Class.Fields[stmt.Name.Value]; exists {
				obj.SetField(stmt.Name.Value, value)
				return value
			}
			// Check if it's a class variable
			if _, exists := obj.Class.ClassVars[stmt.Name.Value]; exists {
				obj.Class.ClassVars[stmt.Name.Value] = value
				return value
			}
		}
	}

	// Check if __CurrentClass__ is bound (class method)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check if it's a class variable
			if _, exists := classInfo.ClassInfo.ClassVars[stmt.Name.Value]; exists {
				classInfo.ClassInfo.ClassVars[stmt.Name.Value] = value
				return value
			}
		}
	}

	// Still not found - return original error
	return newError("undefined variable: %s", stmt.Name.Value)
}

// evalBlockStatement evaluates a block of statements.
func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement) Value {
	var result Value

	for _, stmt := range block.Statements {
		result = i.Eval(stmt)

		if isError(result) {
			return result
		}
	}

	return result
}

// evalIdentifier looks up an identifier in the environment.
// Special handling for "Self" keyword in method contexts.
// Also handles class variable access from within methods.
func (i *Interpreter) evalIdentifier(node *ast.Identifier) Value {
	// Special case for Self keyword
	if node.Value == "Self" {
		val, ok := i.env.Get("Self")
		if !ok {
			return i.newErrorWithLocation(node, "Self used outside method context")
		}
		return val
	}

	// First, try to find in current environment
	val, ok := i.env.Get(node.Value)
	if ok {
		return val
	}

	// Not found in environment - check if we're in a method context (Self is bound)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		// We're in an instance method context - check for instance fields first
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if fieldValue := obj.GetField(node.Value); fieldValue != nil {
				return fieldValue
			}

			// Check if it's a class variable
			if classVarValue, exists := obj.Class.ClassVars[node.Value]; exists {
				return classVarValue
			}
		}
	}

	// Check if we're in a class method context (__CurrentClass__ marker)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check if it's a class variable
			if classVarValue, exists := classInfo.ClassInfo.ClassVars[node.Value]; exists {
				return classVarValue
			}
		}
	}

	// Still not found - return error
	return i.newErrorWithLocation(node, "undefined variable '%s'", node.Value)
}

// evalBinaryExpression evaluates a binary expression.
func (i *Interpreter) evalBinaryExpression(expr *ast.BinaryExpression) Value {
	left := i.Eval(expr.Left)
	if isError(left) {
		return left
	}

	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}

	// Handle operations based on operand types
	switch {
	case left.Type() == "INTEGER" && right.Type() == "INTEGER":
		return i.evalIntegerBinaryOp(expr.Operator, left, right)

	case left.Type() == "FLOAT" || right.Type() == "FLOAT":
		return i.evalFloatBinaryOp(expr.Operator, left, right)

	case left.Type() == "STRING" && right.Type() == "STRING":
		return i.evalStringBinaryOp(expr.Operator, left, right)

	case left.Type() == "BOOLEAN" && right.Type() == "BOOLEAN":
		return i.evalBooleanBinaryOp(expr.Operator, left, right)

	default:
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())
	}
}

// evalIntegerBinaryOp evaluates binary operations on integers.
func (i *Interpreter) evalIntegerBinaryOp(op string, left, right Value) Value {
	// Safe type assertions with error handling
	leftInt, ok := left.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected integer, got %s", left.Type())
	}
	rightInt, ok := right.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected integer, got %s", right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch op {
	case "+":
		return &IntegerValue{Value: leftVal + rightVal}
	case "-":
		return &IntegerValue{Value: leftVal - rightVal}
	case "*":
		return &IntegerValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return i.newErrorWithLocation(i.currentNode, "division by zero")
		}
		// Integer division in DWScript uses / for float division
		// We'll convert to float for division
		return &FloatValue{Value: float64(leftVal) / float64(rightVal)}
	case "div":
		if rightVal == 0 {
			return i.newErrorWithLocation(i.currentNode, "division by zero")
		}
		return &IntegerValue{Value: leftVal / rightVal}
	case "mod":
		if rightVal == 0 {
			return i.newErrorWithLocation(i.currentNode, "division by zero")
		}
		return &IntegerValue{Value: leftVal % rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalFloatBinaryOp evaluates binary operations on floats.
// Handles mixed integer/float operations by converting to float.
func (i *Interpreter) evalFloatBinaryOp(op string, left, right Value) Value {
	var leftVal, rightVal float64

	// Convert left operand to float
	switch v := left.(type) {
	case *FloatValue:
		leftVal = v.Value
	case *IntegerValue:
		leftVal = float64(v.Value)
	default:
		return newError("type error in float operation")
	}

	// Convert right operand to float
	switch v := right.(type) {
	case *FloatValue:
		rightVal = v.Value
	case *IntegerValue:
		rightVal = float64(v.Value)
	default:
		return newError("type error in float operation")
	}

	switch op {
	case "+":
		return &FloatValue{Value: leftVal + rightVal}
	case "-":
		return &FloatValue{Value: leftVal - rightVal}
	case "*":
		return &FloatValue{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return i.newErrorWithLocation(i.currentNode, "division by zero")
		}
		return &FloatValue{Value: leftVal / rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalStringBinaryOp evaluates binary operations on strings.
func (i *Interpreter) evalStringBinaryOp(op string, left, right Value) Value {
	// Safe type assertions with error handling
	leftStr, ok := left.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected string, got %s", left.Type())
	}
	rightStr, ok := right.(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected string, got %s", right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	switch op {
	case "+":
		return &StringValue{Value: leftVal + rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	case "<":
		return &BooleanValue{Value: leftVal < rightVal}
	case ">":
		return &BooleanValue{Value: leftVal > rightVal}
	case "<=":
		return &BooleanValue{Value: leftVal <= rightVal}
	case ">=":
		return &BooleanValue{Value: leftVal >= rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalBooleanBinaryOp evaluates binary operations on booleans.
func (i *Interpreter) evalBooleanBinaryOp(op string, left, right Value) Value {
	// Safe type assertions with error handling
	leftBool, ok := left.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected boolean, got %s", left.Type())
	}
	rightBool, ok := right.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected boolean, got %s", right.Type())
	}

	leftVal := leftBool.Value
	rightVal := rightBool.Value

	switch op {
	case "and":
		return &BooleanValue{Value: leftVal && rightVal}
	case "or":
		return &BooleanValue{Value: leftVal || rightVal}
	case "xor":
		return &BooleanValue{Value: leftVal != rightVal}
	case "=":
		return &BooleanValue{Value: leftVal == rightVal}
	case "<>":
		return &BooleanValue{Value: leftVal != rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

// evalUnaryExpression evaluates a unary expression.
func (i *Interpreter) evalUnaryExpression(expr *ast.UnaryExpression) Value {
	right := i.Eval(expr.Right)
	if isError(right) {
		return right
	}

	switch expr.Operator {
	case "-":
		return i.evalMinusUnaryOp(right)
	case "+":
		return i.evalPlusUnaryOp(right)
	case "not":
		return i.evalNotUnaryOp(right)
	default:
		return newError("unknown operator: %s%s", expr.Operator, right.Type())
	}
}

// evalMinusUnaryOp evaluates the unary minus operator.
func (i *Interpreter) evalMinusUnaryOp(right Value) Value {
	switch v := right.(type) {
	case *IntegerValue:
		return &IntegerValue{Value: -v.Value}
	case *FloatValue:
		return &FloatValue{Value: -v.Value}
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary minus, got %s", right.Type())
	}
}

// evalPlusUnaryOp evaluates the unary plus operator.
func (i *Interpreter) evalPlusUnaryOp(right Value) Value {
	switch right.(type) {
	case *IntegerValue, *FloatValue:
		return right
	default:
		return i.newErrorWithLocation(i.currentNode, "expected integer or float for unary plus, got %s", right.Type())
	}
}

// evalNotUnaryOp evaluates the not operator.
func (i *Interpreter) evalNotUnaryOp(right Value) Value {
	boolVal, ok := right.(*BooleanValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "expected boolean for NOT operator, got %s", right.Type())
	}
	return &BooleanValue{Value: !boolVal.Value}
}

// evalCallExpression evaluates a function call expression.
func (i *Interpreter) evalCallExpression(expr *ast.CallExpression) Value {
	// Get the function name
	funcName, ok := expr.Function.(*ast.Identifier)
	if !ok {
		return newError("function call requires identifier, got %T", expr.Function)
	}

	// Check if it's a user-defined function first
	if fn, exists := i.functions[funcName.Value]; exists {
		// Evaluate all arguments
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}
		return i.callUserFunction(fn, args)
	}

	// Otherwise, try built-in functions
	// Evaluate all arguments
	args := make([]Value, len(expr.Arguments))
	for idx, arg := range expr.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	return i.callBuiltin(funcName.Value, args)
}

// callBuiltin calls a built-in function by name.
func (i *Interpreter) callBuiltin(name string, args []Value) Value {
	switch name {
	case "PrintLn":
		return i.builtinPrintLn(args)
	case "Print":
		return i.builtinPrint(args)
	default:
		return i.newErrorWithLocation(i.currentNode, "undefined function: %s", name)
	}
}

// builtinPrintLn implements the PrintLn built-in function.
// It prints all arguments followed by a newline.
func (i *Interpreter) builtinPrintLn(args []Value) Value {
	for idx, arg := range args {
		if idx > 0 {
			fmt.Fprint(i.output, " ")
		}
		fmt.Fprint(i.output, arg.String())
	}
	fmt.Fprintln(i.output)
	return &NilValue{}
}

// builtinPrint implements the Print built-in function.
// It prints all arguments without a newline.
func (i *Interpreter) builtinPrint(args []Value) Value {
	for idx, arg := range args {
		if idx > 0 {
			fmt.Fprint(i.output, " ")
		}
		fmt.Fprint(i.output, arg.String())
	}
	return &NilValue{}
}

// evalIfStatement evaluates an if statement.
// It evaluates the condition, converts it to a boolean, and executes
// the consequence if true, or the alternative if false (and present).
func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) Value {
	// Evaluate the condition
	condition := i.Eval(stmt.Condition)
	if isError(condition) {
		return condition
	}

	// Convert condition to boolean
	if isTruthy(condition) {
		return i.Eval(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return i.Eval(stmt.Alternative)
	}

	// No alternative and condition was false - return nil
	return &NilValue{}
}

// isTruthy determines if a value is considered "true" for conditional logic.
// In DWScript, only boolean true is truthy. Everything else requires explicit conversion.
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *BooleanValue:
		return v.Value
	default:
		// In DWScript, only booleans can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// For now, treat non-booleans as false
		return false
	}
}

// evalWhileStatement evaluates a while loop.
// It repeatedly evaluates the condition and executes the body while the condition is true.
func (i *Interpreter) evalWhileStatement(stmt *ast.WhileStatement) Value {
	var result Value = &NilValue{}

	for {
		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true
		if !isTruthy(condition) {
			break
		}

		// Execute the body
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}
	}

	return result
}

// evalRepeatStatement evaluates a repeat-until loop.
// The body executes at least once, then repeats until the condition becomes true.
// This differs from while loops: the body always executes at least once,
// and the loop continues UNTIL the condition is true (not WHILE it's true).
func (i *Interpreter) evalRepeatStatement(stmt *ast.RepeatStatement) Value {
	var result Value = &NilValue{}

	for {
		// Execute the body first (repeat-until always executes at least once)
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true - if so, exit the loop
		// Note: repeat UNTIL condition, so we break when condition is TRUE
		if isTruthy(condition) {
			break
		}
	}

	return result
}

// evalForStatement evaluates a for loop.
// DWScript for loops iterate from start to end (or downto), with the loop variable
// scoped to the loop body. The loop variable is automatically created and managed.
func (i *Interpreter) evalForStatement(stmt *ast.ForStatement) Value {
	var result Value = &NilValue{}

	// Evaluate start value
	startVal := i.Eval(stmt.Start)
	if isError(startVal) {
		return startVal
	}

	// Evaluate end value
	endVal := i.Eval(stmt.End)
	if isError(endVal) {
		return endVal
	}

	// Both start and end must be integers for for loops
	startInt, ok := startVal.(*IntegerValue)
	if !ok {
		return newError("for loop start value must be integer, got %s", startVal.Type())
	}

	endInt, ok := endVal.(*IntegerValue)
	if !ok {
		return newError("for loop end value must be integer, got %s", endVal.Type())
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	loopEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = loopEnv

	// Define the loop variable in the loop environment
	loopVarName := stmt.Variable.Value

	// Execute the loop based on direction
	if stmt.Direction == ast.ForTo {
		// Ascending loop: for i := start to end
		for current := startInt.Value; current <= endInt.Value; current++ {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}
		}
	} else {
		// Descending loop: for i := start downto end
		for current := startInt.Value; current >= endInt.Value; current-- {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}
		}
	}

	// Restore the original environment after the loop
	i.env = savedEnv

	return result
}

// evalCaseStatement evaluates a case statement.
// It evaluates the case expression, then checks each branch in order.
// The first branch with a matching value has its statement executed.
// If no branch matches and there's an else clause, it's executed.
func (i *Interpreter) evalCaseStatement(stmt *ast.CaseStatement) Value {
	// Evaluate the case expression
	caseValue := i.Eval(stmt.Expression)
	if isError(caseValue) {
		return caseValue
	}

	// Check each case branch in order
	for _, branch := range stmt.Cases {
		// Check each value in this branch
		for _, branchVal := range branch.Values {
			// Evaluate the branch value
			branchValue := i.Eval(branchVal)
			if isError(branchValue) {
				return branchValue
			}

			// Check if values match
			if i.valuesEqual(caseValue, branchValue) {
				// Execute this branch's statement
				return i.Eval(branch.Statement)
			}
		}
	}

	// No branch matched - execute else clause if present
	if stmt.Else != nil {
		return i.Eval(stmt.Else)
	}

	// No match and no else clause - return nil
	return &NilValue{}
}

// valuesEqual compares two values for equality.
// This is used by case statements to match values.
func (i *Interpreter) valuesEqual(left, right Value) bool {
	// Handle same type comparisons
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *IntegerValue:
		r, ok := right.(*IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *FloatValue:
		r, ok := right.(*FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *StringValue:
		r, ok := right.(*StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *BooleanValue:
		r, ok := right.(*BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *NilValue:
		return true // nil == nil
	default:
		// For other types, use string comparison as fallback
		return left.String() == right.String()
	}
}

// evalFunctionDeclaration evaluates a function declaration.
// It registers the function in the function registry without executing its body.
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
	// Store the function in the registry
	i.functions[fn.Name.Value] = fn
	return &NilValue{}
}

// evalClassDeclaration evaluates a class declaration.
// It builds a ClassInfo from the AST and registers it in the class registry.
// Handles inheritance by copying parent fields and methods to the child class.
func (i *Interpreter) evalClassDeclaration(cd *ast.ClassDecl) Value {
	// Create new ClassInfo
	classInfo := NewClassInfo(cd.Name.Value)

	// Handle inheritance if parent class is specified
	if cd.Parent != nil {
		parentName := cd.Parent.Value
		parentClass, exists := i.classes[parentName]
		if !exists {
			return i.newErrorWithLocation(cd, "parent class '%s' not found", parentName)
		}

		// Set parent reference
		classInfo.Parent = parentClass

		// Copy parent fields (child inherits all parent fields)
		for fieldName, fieldType := range parentClass.Fields {
			classInfo.Fields[fieldName] = fieldType
		}

		// Copy parent methods (child inherits all parent methods)
		// Child methods with same name will override these
		for methodName, methodDecl := range parentClass.Methods {
			classInfo.Methods[methodName] = methodDecl
		}
	}

	// Add own fields to ClassInfo
	for _, field := range cd.Fields {
		// Get the field type from the type annotation
		if field.Type == nil {
			return i.newErrorWithLocation(field, "field '%s' has no type annotation", field.Name.Value)
		}

		// Map type names to types.Type
		var fieldType types.Type
		switch field.Type.Name {
		case "Integer":
			fieldType = types.INTEGER
		case "Float":
			fieldType = types.FLOAT
		case "String":
			fieldType = types.STRING
		case "Boolean":
			fieldType = types.BOOLEAN
		default:
			// Check if it's a class type
			if _, exists := i.classes[field.Type.Name]; exists {
				// It's a class type - for now we'll use NIL as a placeholder
				// This will need proper ClassType support later
				fieldType = types.NIL
			} else {
				return i.newErrorWithLocation(field, "unknown type '%s' for field '%s'", field.Type.Name, field.Name.Value)
			}
		}

		// Check if this is a class variable (static field) or instance field
		if field.IsClassVar {
			// Initialize class variable with default value based on type - Task 7.62
			var defaultValue Value
			switch fieldType {
			case types.INTEGER:
				defaultValue = &IntegerValue{Value: 0}
			case types.FLOAT:
				defaultValue = &FloatValue{Value: 0.0}
			case types.STRING:
				defaultValue = &StringValue{Value: ""}
			case types.BOOLEAN:
				defaultValue = &BooleanValue{Value: false}
			default:
				defaultValue = &NilValue{}
			}
			classInfo.ClassVars[field.Name.Value] = defaultValue
		} else {
			// Store instance field type in ClassInfo
			classInfo.Fields[field.Name.Value] = fieldType
		}
	}

	// Add own methods to ClassInfo (these override parent methods if same name)
	for _, method := range cd.Methods {
		// Check if this is a class method (static method) or instance method
		if method.IsClassMethod {
			// Store in ClassMethods map - Task 7.61
			classInfo.ClassMethods[method.Name.Value] = method
		} else {
			// Store in instance Methods map
			classInfo.Methods[method.Name.Value] = method
		}
	}

	// Identify constructor (method named "Create")
	if constructor, exists := classInfo.Methods["Create"]; exists {
		classInfo.Constructor = constructor
	}

	// Identify destructor (method named "Destroy")
	if destructor, exists := classInfo.Methods["Destroy"]; exists {
		classInfo.Destructor = destructor
	}

	// Register class in registry
	i.classes[classInfo.Name] = classInfo

	return &NilValue{}
}

// evalNewExpression evaluates object instantiation (TClassName.Create(...)).
// It looks up the class, creates an object instance, initializes fields, and calls the constructor.
func (i *Interpreter) evalNewExpression(ne *ast.NewExpression) Value {
	// Look up class in registry
	className := ne.ClassName.Value
	classInfo, exists := i.classes[className]
	if !exists {
		return i.newErrorWithLocation(ne, "class '%s' not found", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize all fields with default values based on their types
	for fieldName, fieldType := range classInfo.Fields {
		var defaultValue Value
		switch fieldType {
		case types.INTEGER:
			defaultValue = &IntegerValue{Value: 0}
		case types.FLOAT:
			defaultValue = &FloatValue{Value: 0.0}
		case types.STRING:
			defaultValue = &StringValue{Value: ""}
		case types.BOOLEAN:
			defaultValue = &BooleanValue{Value: false}
		default:
			defaultValue = &NilValue{}
		}
		obj.SetField(fieldName, defaultValue)
	}

	// Call constructor if present
	if classInfo.Constructor != nil {
		// Evaluate constructor arguments
		args := make([]Value, len(ne.Arguments))
		for idx, arg := range ne.Arguments {
			val := i.Eval(arg)
			if isError(val) {
				return val
			}
			args[idx] = val
		}

		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind constructor parameters to arguments
		for idx, param := range classInfo.Constructor.Parameters {
			if idx < len(args) {
				i.env.Define(param.Name.Value, args[idx])
			}
		}

		// For constructors with return types, initialize the Result variable
		// This allows constructors to use "Result := Self" to return the object
		if classInfo.Constructor.ReturnType != nil {
			i.env.Define("Result", obj)
			i.env.Define(classInfo.Constructor.Name.Value, obj)
		}

		// Execute constructor body
		result := i.Eval(classInfo.Constructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// Restore environment
		i.env = savedEnv
	}

	return obj
}

// evalMemberAccess evaluates field access (obj.field) or class variable access (TClass.Variable).
// It evaluates the object expression and retrieves the field value.
// For class variable access, it checks if the left side is a class name.
func (i *Interpreter) evalMemberAccess(ma *ast.MemberAccessExpression) Value {
	// Check if the left side is a class identifier (for static access: TClass.Variable)
	if ident, ok := ma.Object.(*ast.Identifier); ok {
		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// This is static access: TClass.Variable
			// Look up the class variable
			if classVarValue, exists := classInfo.ClassVars[ma.Member.Value]; exists {
				return classVarValue
			}
			// Not a class variable - this is an error
			return i.newErrorWithLocation(ma, "class variable '%s' not found in class '%s'", ma.Member.Value, classInfo.Name)
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(ma.Object)
	if isError(objVal) {
		return objVal
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.newErrorWithLocation(ma, "cannot access member of non-object type '%s'", objVal.Type())
	}

	// Get the field value
	fieldValue := obj.GetField(ma.Member.Value)
	if fieldValue == nil {
		return i.newErrorWithLocation(ma, "field '%s' not found in class '%s'", ma.Member.Value, obj.Class.Name)
	}

	return fieldValue
}

// evalMethodCall evaluates a method call (obj.Method(...)) or class method call (TClass.Method(...)).
// It looks up the method in the object's class hierarchy and executes it with Self bound to the object.
// For class methods, Self is not bound as they are static methods.
func (i *Interpreter) evalMethodCall(mc *ast.MethodCallExpression) Value {
	// Check if the left side is a class identifier (for static method call: TClass.Method())
	if ident, ok := mc.Object.(*ast.Identifier); ok {
		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// Check if this is a class method (static method) or instance method called as constructor
			classMethod, isClassMethod := classInfo.ClassMethods[mc.Method.Value]
			instanceMethod, isInstanceMethod := classInfo.Methods[mc.Method.Value]

			if isClassMethod {
				// This is a class method - execute it without Self binding
				// Evaluate method arguments
				args := make([]Value, len(mc.Arguments))
				for idx, arg := range mc.Arguments {
					val := i.Eval(arg)
					if isError(val) {
						return val
					}
					args[idx] = val
				}

				// Check argument count matches parameter count
				if len(args) != len(classMethod.Parameters) {
					return i.newErrorWithLocation(mc, "wrong number of arguments for class method '%s': expected %d, got %d",
						mc.Method.Value, len(classMethod.Parameters), len(args))
				}

				// Create method environment (NO Self binding for class methods)
				methodEnv := NewEnclosedEnvironment(i.env)
				savedEnv := i.env
				i.env = methodEnv

				// Bind __CurrentClass__ so class variables can be accessed
				i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

				// Bind method parameters to arguments
				for idx, param := range classMethod.Parameters {
					i.env.Define(param.Name.Value, args[idx])
				}

				// For functions (not procedures), initialize the Result variable
				if classMethod.ReturnType != nil {
					i.env.Define("Result", &NilValue{})
					// Also define the method name as an alias for Result (DWScript style)
					i.env.Define(classMethod.Name.Value, &NilValue{})
				}

				// Execute method body
				result := i.Eval(classMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if classMethod.ReturnType != nil {
					// Check both Result and method name variable
					resultVal, resultOk := i.env.Get("Result")
					methodNameVal, methodNameOk := i.env.Get(classMethod.Name.Value)

					// Use whichever variable is not nil, preferring Result if both are set
					if resultOk && resultVal.Type() != "NIL" {
						returnValue = resultVal
					} else if methodNameOk && methodNameVal.Type() != "NIL" {
						returnValue = methodNameVal
					} else if resultOk {
						returnValue = resultVal
					} else if methodNameOk {
						returnValue = methodNameVal
					} else {
						returnValue = &NilValue{}
					}
				} else {
					// Procedure - no return value
					returnValue = &NilValue{}
				}

				// Restore environment
				i.env = savedEnv

				return returnValue
			} else if isInstanceMethod {
				// This is an instance method being called from the class (e.g., TClass.Create())
				// Create a new instance and call the method on it
				obj := NewObjectInstance(classInfo)

				// Initialize all fields with default values
				for fieldName, fieldType := range classInfo.Fields {
					var defaultValue Value
					switch fieldType {
					case types.INTEGER:
						defaultValue = &IntegerValue{Value: 0}
					case types.FLOAT:
						defaultValue = &FloatValue{Value: 0.0}
					case types.STRING:
						defaultValue = &StringValue{Value: ""}
					case types.BOOLEAN:
						defaultValue = &BooleanValue{Value: false}
					default:
						defaultValue = &NilValue{}
					}
					obj.SetField(fieldName, defaultValue)
				}

				// Evaluate method arguments
				args := make([]Value, len(mc.Arguments))
				for idx, arg := range mc.Arguments {
					val := i.Eval(arg)
					if isError(val) {
						return val
					}
					args[idx] = val
				}

				// Check argument count matches parameter count
				if len(args) != len(instanceMethod.Parameters) {
					return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
						mc.Method.Value, len(instanceMethod.Parameters), len(args))
				}

				// Create method environment with Self bound to new object
				methodEnv := NewEnclosedEnvironment(i.env)
				savedEnv := i.env
				i.env = methodEnv

				// Bind Self to the object
				i.env.Define("Self", obj)

				// Bind method parameters to arguments
				for idx, param := range instanceMethod.Parameters {
					i.env.Define(param.Name.Value, args[idx])
				}

				// For functions (not procedures), initialize the Result variable
				if instanceMethod.ReturnType != nil {
					i.env.Define("Result", &NilValue{})
					// Also define the method name as an alias for Result (DWScript style)
					i.env.Define(instanceMethod.Name.Value, &NilValue{})
				} else {
					// DEBUG: This shouldn't happen for Create which has a return type
					// But let's check
					fmt.Fprintf(i.output, "DEBUG: Method %s has no return type!\n", instanceMethod.Name.Value)
				}

				// Execute method body
				result := i.Eval(instanceMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if instanceMethod.ReturnType != nil {
					// Check both Result and method name variable
					resultVal, resultOk := i.env.Get("Result")
					methodNameVal, methodNameOk := i.env.Get(instanceMethod.Name.Value)

					// Use whichever variable is not nil, preferring Result if both are set
					if resultOk && resultVal.Type() != "NIL" {
						returnValue = resultVal
					} else if methodNameOk && methodNameVal.Type() != "NIL" {
						returnValue = methodNameVal
					} else if resultOk {
						returnValue = resultVal
					} else if methodNameOk {
						returnValue = methodNameVal
					} else {
						returnValue = &NilValue{}
					}
				} else {
					// Procedure - no return value
					returnValue = &NilValue{}
				}

				// Restore environment
				i.env = savedEnv

				return returnValue
			} else {
				// Neither class method nor instance method found
				return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, classInfo.Name)
			}
		}
	}

	// Not static method call - evaluate the object expression for instance method call
	objVal := i.Eval(mc.Object)
	if isError(objVal) {
		return objVal
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.newErrorWithLocation(mc, "cannot call method on non-object type '%s'", objVal.Type())
	}

	// Look up method in object's class
	method := obj.GetMethod(mc.Method.Value)
	if method == nil {
		return i.newErrorWithLocation(mc, "method '%s' not found in class '%s'", mc.Method.Value, obj.Class.Name)
	}

	// Evaluate method arguments
	args := make([]Value, len(mc.Arguments))
	for idx, arg := range mc.Arguments {
		val := i.Eval(arg)
		if isError(val) {
			return val
		}
		args[idx] = val
	}

	// Check argument count matches parameter count
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(mc, "wrong number of arguments for method '%s': expected %d, got %d",
			mc.Method.Value, len(method.Parameters), len(args))
	}

	// Create method environment with Self bound to object
	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	// Bind Self to the object
	i.env.Define("Self", obj)

	// Bind method parameters to arguments
	for idx, param := range method.Parameters {
		i.env.Define(param.Name.Value, args[idx])
	}

	// For functions (not procedures), initialize the Result variable
	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		// Also define the method name as an alias for Result (DWScript style)
		i.env.Define(method.Name.Value, &NilValue{})
	}

	// Execute method body
	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	// Extract return value (same logic as regular functions)
	var returnValue Value
	if method.ReturnType != nil {
		// Check both Result and method name variable
		resultVal, resultOk := i.env.Get("Result")
		methodNameVal, methodNameOk := i.env.Get(method.Name.Value)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if methodNameOk && methodNameVal.Type() != "NIL" {
			returnValue = methodNameVal
		} else if resultOk {
			returnValue = resultVal
		} else if methodNameOk {
			returnValue = methodNameVal
		} else {
			returnValue = &NilValue{}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Restore environment
	i.env = savedEnv

	return returnValue
}

// callUserFunction calls a user-defined function.
// It creates a new environment, binds parameters to arguments, executes the body,
// and extracts the return value from the Result variable or function name variable.
func (i *Interpreter) callUserFunction(fn *ast.FunctionDecl, args []Value) Value {
	// Check argument count matches parameter count
	if len(args) != len(fn.Parameters) {
		return newError("wrong number of arguments: expected %d, got %d",
			len(fn.Parameters), len(args))
	}

	// Create a new environment for the function scope
	funcEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = funcEnv

	// Bind parameters to arguments
	for idx, param := range fn.Parameters {
		if param.ByRef {
			// By-reference parameter - we need to handle this specially
			// For now, we'll pass by value (TODO: implement proper by-ref support)
			i.env.Define(param.Name.Value, args[idx])
		} else {
			// By-value parameter
			i.env.Define(param.Name.Value, args[idx])
		}
	}

	// For functions (not procedures), initialize the Result variable
	if fn.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		// Also define the function name as an alias for Result (DWScript style)
		i.env.Define(fn.Name.Value, &NilValue{})
	}

	// Execute the function body
	i.Eval(fn.Body)

	// Extract return value
	var returnValue Value
	if fn.ReturnType != nil {
		// Check both Result and function name variable
		// Prioritize whichever one was actually set (not nil)
		resultVal, resultOk := i.env.Get("Result")
		fnNameVal, fnNameOk := i.env.Get(fn.Name.Value)

		// Use whichever variable is not nil, preferring Result if both are set
		if resultOk && resultVal.Type() != "NIL" {
			returnValue = resultVal
		} else if fnNameOk && fnNameVal.Type() != "NIL" {
			returnValue = fnNameVal
		} else if resultOk {
			// Result exists but is nil - use it
			returnValue = resultVal
		} else if fnNameOk {
			// Function name exists but is nil - use it
			returnValue = fnNameVal
		} else {
			// Neither exists (shouldn't happen)
			returnValue = &NilValue{}
		}
	} else {
		// Procedure - no return value
		returnValue = &NilValue{}
	}

	// Restore the original environment
	i.env = savedEnv

	return returnValue
}

// ErrorValue represents a runtime error.
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new ErrorValue.
func newError(format string, args ...interface{}) *ErrorValue {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// newErrorWithLocation creates a new ErrorValue with location information from a node.
func (i *Interpreter) newErrorWithLocation(node ast.Node, format string, args ...interface{}) *ErrorValue {
	message := fmt.Sprintf(format, args...)

	// Try to get location information from the node's token
	if node != nil {
		tokenLiteral := node.TokenLiteral()
		if tokenLiteral != "" {
			// Extract token information - we need to get the actual token from the node
			location := i.getLocationFromNode(node)
			if location != "" {
				message = fmt.Sprintf("%s at %s", message, location)
			}
		}
	}

	return &ErrorValue{Message: message}
}

// getLocationFromNode extracts location information from an AST node
func (i *Interpreter) getLocationFromNode(node ast.Node) string {
	// Try to extract token information from various node types
	switch n := node.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.IntegerLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.FloatLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.StringLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.BooleanLiteral:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.BinaryExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.UnaryExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.CallExpression:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.VarDeclStatement:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	case *ast.AssignmentStatement:
		return fmt.Sprintf("line %d, column %d", n.Token.Pos.Line, n.Token.Pos.Column)
	}
	return ""
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}
