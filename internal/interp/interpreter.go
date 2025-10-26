package interp

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// Interpreter executes DWScript AST nodes and manages the runtime environment.
type Interpreter struct {
	env              *Environment                 // The current execution environment
	output           io.Writer                    // Where to write output (e.g., from PrintLn)
	functions        map[string]*ast.FunctionDecl // Registry of user-defined functions
	classes          map[string]*ClassInfo        // Registry of class definitions
	interfaces       map[string]*InterfaceInfo    // Registry of interface definitions (Task 7.118)
	currentNode      ast.Node                     // Current AST node being evaluated (for error reporting)
	globalOperators  *runtimeOperatorRegistry     // Global operator overloads (Stage 8)
	conversions      *runtimeConversionRegistry   // Global conversions
	rand             *rand.Rand                   // Random number generator for Random() and Randomize()
	exception        *ExceptionValue              // Current active exception (nil if none) (Task 8.212)
	handlerException *ExceptionValue              // Exception being handled (for bare raise)

	// Control flow signals (Task 8.235i)
	// These flags signal control flow changes (break, continue, exit) and are checked
	// after each statement. They propagate up the call stack until handled by the
	// appropriate control structure (loop for break/continue, function for exit).
	breakSignal    bool // Set by break statement, cleared by loop
	continueSignal bool // Set by continue statement, cleared by loop
	exitSignal     bool // Set by exit statement, cleared by function return
}

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	env := NewEnvironment()
	// Initialize random number generator with a default seed
	// Randomize() can be called to re-seed with current time
	source := rand.NewSource(1)
	interp := &Interpreter{
		env:             env,
		output:          output,
		functions:       make(map[string]*ast.FunctionDecl),
		classes:         make(map[string]*ClassInfo),
		interfaces:      make(map[string]*InterfaceInfo), // Task 7.118
		globalOperators: newRuntimeOperatorRegistry(),
		conversions:     newRuntimeConversionRegistry(),
		rand:            rand.New(source),
	}

	// Register built-in exception classes (Task 8.203-8.204)
	interp.registerBuiltinExceptions()

	// Initialize ExceptObject to nil (Task 8.206)
	// ExceptObject is a built-in global variable that holds the current exception
	env.Define("ExceptObject", &NilValue{})

	return interp
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

	case *ast.ConstDecl:
		return i.evalConstDecl(node)

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

	case *ast.TryStatement:
		return i.evalTryStatement(node)

	case *ast.RaiseStatement:
		return i.evalRaiseStatement(node)

	case *ast.BreakStatement:
		return i.evalBreakStatement(node)

	case *ast.ContinueStatement:
		return i.evalContinueStatement(node)

	case *ast.ExitStatement:
		return i.evalExitStatement(node)

	case *ast.FunctionDecl:
		return i.evalFunctionDeclaration(node)

	case *ast.ClassDecl:
		return i.evalClassDeclaration(node)

	case *ast.InterfaceDecl:
		return i.evalInterfaceDeclaration(node)

	case *ast.OperatorDecl:
		return i.evalOperatorDeclaration(node)

	case *ast.EnumDecl:
		return i.evalEnumDeclaration(node)

	case *ast.RecordDecl:
		return i.evalRecordDeclaration(node)

	case *ast.ArrayDecl:
		return i.evalArrayDeclaration(node)

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

	case *ast.EnumLiteral:
		return i.evalEnumLiteral(node)

	case *ast.RecordLiteral:
		return i.evalRecordLiteral(node)

	case *ast.SetLiteral:
		return i.evalSetLiteral(node)

	case *ast.IndexExpression:
		return i.evalIndexExpression(node)

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

		// Check if exception is active - if so, unwind the stack
		if i.exception != nil {
			break
		}

		// Task 8.235n: Check if exit was called at program level
		if i.exitSignal {
			i.exitSignal = false // Clear signal
			break                // Exit the program
		}
	}

	// If there's an uncaught exception, convert it to an error
	if i.exception != nil {
		exc := i.exception
		return newError("uncaught exception: %s", exc.Inspect())
	}

	return result
}

// evalVarDeclStatement evaluates a variable declaration statement.
// It defines a new variable in the current environment.
// Task 7.144: External variables are marked with a special ExternalVarValue.
func (i *Interpreter) evalVarDeclStatement(stmt *ast.VarDeclStatement) Value {
	var value Value

	// Task 7.144: Handle external variables
	if stmt.IsExternal {
		// Store a special marker for external variables
		externalName := stmt.ExternalName
		if externalName == "" {
			externalName = stmt.Name.Value
		}
		value = &ExternalVarValue{
			Name:         stmt.Name.Value,
			ExternalName: externalName,
		}
		i.env.Define(stmt.Name.Value, value)
		return value
	}

	if stmt.Value != nil {
		value = i.Eval(stmt.Value)
		if isError(value) {
			return value
		}
	} else {
		// No initializer - check if we need to initialize based on type
		if stmt.Type != nil {
			typeName := stmt.Type.Name

			// Check if this is a record type
			recordTypeKey := "__record_type_" + typeName
			if typeVal, ok := i.env.Get(recordTypeKey); ok {
				if rtv, ok := typeVal.(*RecordTypeValue); ok {
					// Initialize with empty record value
					value = NewRecordValue(rtv.RecordType)
				} else {
					value = &NilValue{}
				}
			} else {
				// Check if this is an array type
				arrayTypeKey := "__array_type_" + typeName
				if typeVal, ok := i.env.Get(arrayTypeKey); ok {
					if atv, ok := typeVal.(*ArrayTypeValue); ok {
						// Initialize with empty array value
						value = NewArrayValue(atv.ArrayType)
					} else {
						value = &NilValue{}
					}
				} else {
					// Initialize basic types with their zero values
					// Task 8.19e: Proper initialization allows implicit conversions to work with target type
					switch strings.ToLower(typeName) {
					case "integer":
						value = &IntegerValue{Value: 0}
					case "float":
						value = &FloatValue{Value: 0.0}
					case "string":
						value = &StringValue{Value: ""}
					case "boolean":
						value = &BooleanValue{Value: false}
					default:
						value = &NilValue{}
					}
				}
			}
		} else {
			value = &NilValue{}
		}
	}

	i.env.Define(stmt.Name.Value, value)
	return value
}

// evalConstDecl evaluates a const declaration statement (Task 8.257)
// Constants are immutable values stored in the environment.
// Immutability is enforced at semantic analysis time, so at runtime
// we simply evaluate the value and store it like a variable.
func (i *Interpreter) evalConstDecl(stmt *ast.ConstDecl) Value {
	// Constants must have a value
	if stmt.Value == nil {
		return newError("constant '%s' must have a value", stmt.Name.Value)
	}

	// Evaluate the constant value
	value := i.Eval(stmt.Value)
	if isError(value) {
		return value
	}

	// Store the constant in the environment
	// Note: Immutability is enforced by semantic analysis, not at runtime
	i.env.Define(stmt.Name.Value, value)
	return value
}

// evalAssignmentStatement evaluates an assignment statement.
// It updates an existing variable's value or sets an object/array element.
// Supports: x := value, obj.field := value, arr[i] := value
func (i *Interpreter) evalAssignmentStatement(stmt *ast.AssignmentStatement) Value {
	// Evaluate the value to assign
	// Special handling for record literals without type names (Task 8.74)
	var value Value
	if recordLit, ok := stmt.Value.(*ast.RecordLiteral); ok && recordLit.TypeName == "" {
		// This is an untyped record literal - get type from target variable if it's a simple identifier
		if targetIdent, ok := stmt.Target.(*ast.Identifier); ok {
			targetVar, exists := i.env.Get(targetIdent.Value)
			if exists {
				if recVal, ok := targetVar.(*RecordValue); ok {
					// Set the type name in the literal temporarily
					recordLit.TypeName = recVal.RecordType.Name
					value = i.Eval(recordLit)
					recordLit.TypeName = ""
				} else {
					value = i.Eval(stmt.Value)
				}
			} else {
				value = i.Eval(stmt.Value)
			}
		} else {
			value = i.Eval(stmt.Value)
		}
	} else {
		value = i.Eval(stmt.Value)
	}

	if isError(value) {
		return value
	}

	// Task 8.77: Records have value semantics - copy when assigning
	if recordVal, ok := value.(*RecordValue); ok {
		value = recordVal.Copy()
	}

	// Handle different target types
	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x := value
		return i.evalSimpleAssignment(target, value, stmt)

	case *ast.MemberAccessExpression:
		// Member assignment: obj.field := value or TClass.Variable := value
		return i.evalMemberAssignment(target, value, stmt)

	case *ast.IndexExpression:
		// Array index assignment: arr[i] := value (Task 8.139)
		return i.evalIndexAssignment(target, value, stmt)

	default:
		return i.newErrorWithLocation(stmt, "invalid assignment target type: %T", target)
	}
}

// evalSimpleAssignment handles simple variable assignment: x := value
func (i *Interpreter) evalSimpleAssignment(target *ast.Identifier, value Value, _ *ast.AssignmentStatement) Value {
	// Task 7.144: Check if trying to assign to an external variable
	// Task 8.19a: Apply implicit conversion if types don't match
	if existingVal, ok := i.env.Get(target.Value); ok {
		if extVar, isExternal := existingVal.(*ExternalVarValue); isExternal {
			return newError("Unsupported external variable assignment: %s", extVar.Name)
		}

		// Try implicit conversion if types don't match
		targetType := existingVal.Type()
		sourceType := value.Type()
		if targetType != sourceType {
			if converted, ok := i.tryImplicitConversion(value, targetType); ok {
				value = converted
			}
		}
	}

	// First try to set in current environment
	err := i.env.Set(target.Value, value)
	if err == nil {
		return value
	}

	// Not in environment - check if we're in a method context and this is a field/class variable
	// Check if Self is bound (instance method)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if _, exists := obj.Class.Fields[target.Value]; exists {
				obj.SetField(target.Value, value)
				return value
			}
			// Check if it's a class variable
			if _, exists := obj.Class.ClassVars[target.Value]; exists {
				obj.Class.ClassVars[target.Value] = value
				return value
			}
		}
	}

	// Check if __CurrentClass__ is bound (class method)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check if it's a class variable
			if _, exists := classInfo.ClassInfo.ClassVars[target.Value]; exists {
				classInfo.ClassInfo.ClassVars[target.Value] = value
				return value
			}
		}
	}

	// Still not found - return original error
	return newError("undefined variable: %s", target.Value)
}

// evalMemberAssignment handles member assignment: obj.field := value or TClass.Variable := value
func (i *Interpreter) evalMemberAssignment(target *ast.MemberAccessExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Check if the left side is a class identifier (for static assignment: TClass.Variable := value)
	if ident, ok := target.Object.(*ast.Identifier); ok {
		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// This is a class variable assignment
			fieldName := target.Member.Value
			if _, exists := classInfo.ClassVars[fieldName]; !exists {
				return i.newErrorWithLocation(stmt, "class variable '%s' not found in class '%s'", fieldName, ident.Value)
			}
			// Assign to the class variable
			classInfo.ClassVars[fieldName] = value
			return value
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(target.Object)
	if isError(objVal) {
		return objVal
	}

	// Task 8.76: Check if it's a record value
	if recordVal, ok := objVal.(*RecordValue); ok {
		fieldName := target.Member.Value
		// Verify field exists in the record type
		if _, exists := recordVal.RecordType.Fields[fieldName]; !exists {
			return i.newErrorWithLocation(stmt, "field '%s' not found in record '%s'", fieldName, recordVal.RecordType.Name)
		}

		// Set the field value
		recordVal.Fields[fieldName] = value
		return value
	}

	// Special case: If objVal is NilValue and target.Object is an IndexExpression,
	// we might be trying to assign to an uninitialized record array element.
	// Auto-initialize the record and retry.
	if _, isNil := objVal.(*NilValue); isNil {
		if indexExpr, ok := target.Object.(*ast.IndexExpression); ok {
			// Evaluate the array
			arrayVal := i.Eval(indexExpr.Left)
			if isError(arrayVal) {
				return arrayVal
			}

			if arrVal, ok := arrayVal.(*ArrayValue); ok {
				// Check if the element type is a record
				if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
					if recordType, ok := arrVal.ArrayType.ElementType.(*types.RecordType); ok {
						// Auto-initialize a new record
						newRecord := &RecordValue{
							RecordType: recordType,
							Fields:     make(map[string]Value),
						}

						// Assign it to the array element using evalIndexAssignment
						assignStmt := &ast.AssignmentStatement{
							Token:  stmt.Token,
							Target: indexExpr,
							Value:  &ast.Identifier{Value: "__temp__"},
						}

						// Temporarily store the record
						tempResult := i.evalIndexAssignment(indexExpr, newRecord, assignStmt)
						if isError(tempResult) {
							return tempResult
						}

						// Now retry the member assignment with the initialized record
						objVal = newRecord
					}
				}
			}
		}
	}

	// Re-check if it's a record value after potential initialization
	if recordVal, ok := objVal.(*RecordValue); ok {
		fieldName := target.Member.Value
		// Verify field exists in the record type
		if _, exists := recordVal.RecordType.Fields[fieldName]; !exists {
			return i.newErrorWithLocation(stmt, "field '%s' not found in record '%s'", fieldName, recordVal.RecordType.Name)
		}

		// Set the field value
		recordVal.Fields[fieldName] = value
		return value
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.newErrorWithLocation(stmt, "cannot assign to field of non-object type '%s'", objVal.Type())
	}

	memberName := target.Member.Value

	// Task 8.54: Check if this is a property assignment (properties take precedence over fields)
	if propInfo := obj.Class.lookupProperty(memberName); propInfo != nil {
		return i.evalPropertyWrite(obj, propInfo, value, stmt)
	}

	// Not a property - try direct field assignment
	// Verify field exists in the class
	if _, exists := obj.Class.Fields[memberName]; !exists {
		return i.newErrorWithLocation(stmt, "field '%s' not found in class '%s'", memberName, obj.Class.Name)
	}

	// Set the field value
	obj.SetField(memberName, value)
	return value
}

// evalIndexAssignment handles array index assignment: arr[i] := value (Task 8.139)
func (i *Interpreter) evalIndexAssignment(target *ast.IndexExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Evaluate the array expression
	arrayVal := i.Eval(target.Left)
	if isError(arrayVal) {
		return arrayVal
	}

	// Evaluate the index expression
	indexVal := i.Eval(target.Index)
	if isError(indexVal) {
		return indexVal
	}

	// Index must be an integer
	indexInt, ok := indexVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(stmt, "array index must be an integer, got %s", indexVal.Type())
	}
	index := int(indexInt.Value)

	// Check if left side is an array
	arrayValue, ok := arrayVal.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(stmt, "cannot index type %s", arrayVal.Type())
	}

	// Perform bounds checking and get physical index
	if arrayValue.ArrayType == nil {
		return i.newErrorWithLocation(stmt, "array has no type information")
	}

	var physicalIndex int
	if arrayValue.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arrayValue.ArrayType.LowBound
		highBound := *arrayValue.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arrayValue.Elements) {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (array length is %d)", index, len(arrayValue.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arrayValue.Elements) {
		return i.newErrorWithLocation(stmt, "array index out of bounds: physical index %d, length %d", physicalIndex, len(arrayValue.Elements))
	}

	// Update the array element
	arrayValue.Elements[physicalIndex] = value

	return value
}

// evalBlockStatement evaluates a block of statements.
func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement) Value {
	if block == nil {
		return &NilValue{}
	}

	var result Value

	for _, stmt := range block.Statements {
		result = i.Eval(stmt)

		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack (Task 8.212)
		if i.exception != nil {
			return nil
		}

		// Task 8.235o: Check for control flow signals and propagate them upward
		// These signals should propagate up to the appropriate control structure
		if i.breakSignal || i.continueSignal || i.exitSignal {
			return nil // Propagate signal upward by returning early
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
		// Task 7.144: Check if this is an external variable
		if extVar, isExternal := val.(*ExternalVarValue); isExternal {
			return i.newErrorWithLocation(node, "Unsupported external variable access: %s", extVar.Name)
		}
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

	if result, ok := i.tryBinaryOperator(expr.Operator, left, right, expr); ok {
		return result
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

	// Handle object and nil comparisons (=, <>)
	case expr.Operator == "=" || expr.Operator == "<>":
		// Check if either operand is nil or an object instance
		_, leftIsNil := left.(*NilValue)
		_, rightIsNil := right.(*NilValue)
		_, leftIsObj := left.(*ObjectInstance)
		_, rightIsObj := right.(*ObjectInstance)

		// If either is nil or an object, do object identity comparison
		if leftIsNil || rightIsNil || leftIsObj || rightIsObj {
			// Both nil
			if leftIsNil && rightIsNil {
				if expr.Operator == "=" {
					return &BooleanValue{Value: true}
				} else {
					return &BooleanValue{Value: false}
				}
			}

			// One is nil, one is not
			if leftIsNil || rightIsNil {
				if expr.Operator == "=" {
					return &BooleanValue{Value: false}
				} else {
					return &BooleanValue{Value: true}
				}
			}

			// Both are objects - compare by identity
			if expr.Operator == "=" {
				return &BooleanValue{Value: left == right}
			} else {
				return &BooleanValue{Value: left != right}
			}
		}
		// Not object/nil comparison - fall through to default case
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())

	// Check if both are records (by type assertion, not string comparison)
	// Since RecordValue.Type() now returns actual type name (e.g., "TPoint"), not "RECORD"
	default:
		if _, leftIsRecord := left.(*RecordValue); leftIsRecord {
			if _, rightIsRecord := right.(*RecordValue); rightIsRecord {
				return i.evalRecordBinaryOp(expr.Operator, left, right)
			}
		}
		return i.newErrorWithLocation(expr, "type mismatch: %s %s %s", left.Type(), expr.Operator, right.Type())
	}
}

func (i *Interpreter) tryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool) {
	operands := []Value{left, right}
	operandTypes := []string{valueTypeKey(left), valueTypeKey(right)}

	if obj, ok := left.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}
	if obj, ok := right.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}
	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}
	return nil, false
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

	if result, ok := i.tryUnaryOperator(expr.Operator, right, expr); ok {
		return result
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

func (i *Interpreter) tryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool) {
	operands := []Value{operand}
	operandTypes := []string{valueTypeKey(operand)}

	if obj, ok := operand.(*ObjectInstance); ok {
		if entry, found := obj.Class.lookupOperator(operator, operandTypes); found {
			return i.invokeRuntimeOperator(entry, operands, node), true
		}
	}

	if entry, found := i.globalOperators.lookup(operator, operandTypes); found {
		return i.invokeRuntimeOperator(entry, operands, node), true
	}

	return nil, false
}

func (i *Interpreter) invokeRuntimeOperator(entry *runtimeOperatorEntry, operands []Value, node ast.Node) Value {
	if entry == nil {
		return i.newErrorWithLocation(node, "operator not registered")
	}

	if entry.Class != nil {
		if entry.IsClassMethod {
			return i.invokeClassOperatorMethod(entry.Class, entry.BindingName, operands, node)
		}

		if entry.SelfIndex < 0 || entry.SelfIndex >= len(operands) {
			return i.newErrorWithLocation(node, "invalid operator configuration for '%s'", entry.Operator)
		}

		selfVal := operands[entry.SelfIndex]
		obj, ok := selfVal.(*ObjectInstance)
		if !ok {
			return i.newErrorWithLocation(node, "operator '%s' requires object operand", entry.Operator)
		}
		if !obj.IsInstanceOf(entry.Class) {
			return i.newErrorWithLocation(node, "operator '%s' requires instance of '%s'", entry.Operator, entry.Class.Name)
		}

		args := make([]Value, 0, len(operands)-1)
		for idx, val := range operands {
			if idx == entry.SelfIndex {
				continue
			}
			args = append(args, val)
		}

		return i.invokeInstanceOperatorMethod(obj, entry.BindingName, args, node)
	}

	return i.invokeGlobalOperator(entry, operands, node)
}

func (i *Interpreter) invokeGlobalOperator(entry *runtimeOperatorEntry, operands []Value, node ast.Node) Value {
	fn, exists := i.functions[entry.BindingName]
	if !exists {
		return i.newErrorWithLocation(node, "operator binding '%s' not found", entry.BindingName)
	}
	if len(fn.Parameters) != len(operands) {
		return i.newErrorWithLocation(node, "operator '%s' expects %d operands, got %d", entry.Operator, len(fn.Parameters), len(operands))
	}
	return i.callUserFunction(fn, operands)
}

func (i *Interpreter) invokeInstanceOperatorMethod(obj *ObjectInstance, methodName string, args []Value, node ast.Node) Value {
	method := obj.GetMethod(methodName)
	if method == nil {
		return i.newErrorWithLocation(node, "method '%s' not found in class '%s'", methodName, obj.Class.Name)
	}

	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(node, "method '%s' expects %d arguments, got %d", methodName, len(method.Parameters), len(args))
	}

	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	i.env.Define("Self", obj)

	// Bind parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		i.env.Define(method.Name.Value, &NilValue{})
	}

	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	var returnValue Value = &NilValue{}
	if method.ReturnType != nil {
		returnValue = i.extractReturnValue(method, methodEnv)
	}

	i.env = savedEnv
	return returnValue
}

func (i *Interpreter) invokeClassOperatorMethod(classInfo *ClassInfo, methodName string, args []Value, node ast.Node) Value {
	method, exists := classInfo.ClassMethods[methodName]
	if !exists {
		return i.newErrorWithLocation(node, "class method '%s' not found in class '%s'", methodName, classInfo.Name)
	}
	if len(args) != len(method.Parameters) {
		return i.newErrorWithLocation(node, "class method '%s' expects %d arguments, got %d", methodName, len(method.Parameters), len(args))
	}

	methodEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = methodEnv

	i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: classInfo})

	// Bind parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
	}

	if method.ReturnType != nil {
		i.env.Define("Result", &NilValue{})
		i.env.Define(method.Name.Value, &NilValue{})
	}

	result := i.Eval(method.Body)
	if isError(result) {
		i.env = savedEnv
		return result
	}

	var returnValue Value = &NilValue{}
	if method.ReturnType != nil {
		returnValue = i.extractReturnValue(method, methodEnv)
	}

	i.env = savedEnv
	return returnValue
}

func (i *Interpreter) extractReturnValue(method *ast.FunctionDecl, env *Environment) Value {
	resultVal, resultOk := env.Get("Result")
	methodNameVal, methodNameOk := env.Get(method.Name.Value)

	var returnValue Value
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

	// Task 8.19c: Apply implicit conversion if return type doesn't match
	if method.ReturnType != nil && returnValue.Type() != "NIL" {
		expectedReturnType := method.ReturnType.Name
		if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
			return converted
		}
	}

	return returnValue
}

// tryImplicitConversion attempts to apply an implicit conversion from source to target type.
// Returns (convertedValue, true) if conversion found and applied, (original, false) otherwise.
// Task 8.19a: Apply implicit conversions automatically at runtime.
// Task 8.19d: Support chained implicit conversions (e.g., Integer -> String -> Custom).
func (i *Interpreter) tryImplicitConversion(value Value, targetTypeName string) (Value, bool) {
	sourceTypeName := value.Type()

	// No conversion needed if types already match
	if sourceTypeName == targetTypeName {
		return value, false
	}

	// Normalize type names for conversion lookup (to match how they're registered)
	normalizedSource := normalizeTypeAnnotation(sourceTypeName)
	normalizedTarget := normalizeTypeAnnotation(targetTypeName)

	// Task 8.19a: Try direct conversion first
	entry, found := i.conversions.findImplicit(normalizedSource, normalizedTarget)
	if found {
		// Look up the conversion function
		fn, exists := i.functions[entry.BindingName]
		if !exists {
			// This should not happen if semantic analysis passed
			return value, false
		}

		// Call the conversion function with the value as argument
		args := []Value{value}
		result := i.callUserFunction(fn, args)

		if isError(result) {
			return result, false
		}

		return result, true
	}

	// Task 8.19d: Try chained conversion if direct conversion not found
	const maxConversionChainDepth = 3
	path := i.conversions.findConversionPath(normalizedSource, normalizedTarget, maxConversionChainDepth)
	if len(path) < 2 {
		return value, false
	}

	// Apply conversions sequentially along the path
	currentValue := value
	for idx := 0; idx < len(path)-1; idx++ {
		fromType := path[idx]
		toType := path[idx+1]

		// Find the conversion function for this step
		stepEntry, stepFound := i.conversions.findImplicit(fromType, toType)
		if !stepFound {
			// Path is broken - this shouldn't happen if findConversionPath worked correctly
			return value, false
		}

		// Look up the conversion function
		fn, exists := i.functions[stepEntry.BindingName]
		if !exists {
			return value, false
		}

		// Apply this conversion step
		args := []Value{currentValue}
		result := i.callUserFunction(fn, args)

		if isError(result) {
			return result, false
		}

		currentValue = result
	}

	return currentValue, true
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

	// Check if this is an instance method call within the current context (implicit Self)
	if selfVal, ok := i.env.Get("Self"); ok {
		if obj, isObj := AsObject(selfVal); isObj {
			if obj.GetMethod(funcName.Value) != nil {
				mc := &ast.MethodCallExpression{
					Token:     expr.Token,
					Object:    &ast.Identifier{Token: funcName.Token, Value: "Self"},
					Method:    funcName,
					Arguments: expr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}
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
	case "Ord":
		return i.builtinOrd(args)
	case "Integer":
		return i.builtinInteger(args)
	case "Length":
		return i.builtinLength(args)
	case "Copy":
		return i.builtinCopy(args)
	case "Concat":
		return i.builtinConcat(args)
	case "Pos":
		return i.builtinPos(args)
	case "UpperCase":
		return i.builtinUpperCase(args)
	case "LowerCase":
		return i.builtinLowerCase(args)
	case "Abs":
		return i.builtinAbs(args)
	case "Sqrt":
		return i.builtinSqrt(args)
	case "Sin":
		return i.builtinSin(args)
	case "Cos":
		return i.builtinCos(args)
	case "Tan":
		return i.builtinTan(args)
	case "Random":
		return i.builtinRandom(args)
	case "Randomize":
		return i.builtinRandomize(args)
	case "Exp":
		return i.builtinExp(args)
	case "Ln":
		return i.builtinLn(args)
	case "Round":
		return i.builtinRound(args)
	case "Trunc":
		return i.builtinTrunc(args)
	case "Low":
		return i.builtinLow(args)
	case "High":
		return i.builtinHigh(args)
	case "SetLength":
		return i.builtinSetLength(args)
	case "Add":
		return i.builtinAdd(args)
	case "Delete":
		return i.builtinDelete(args)
	case "IntToStr":
		return i.builtinIntToStr(args)
	case "StrToInt":
		return i.builtinStrToInt(args)
	case "FloatToStr":
		return i.builtinFloatToStr(args)
	case "StrToFloat":
		return i.builtinStrToFloat(args)
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

// builtinOrd implements the Ord() built-in function.
// It returns the ordinal value of an enum, boolean, or character.
// Task 8.51: Ord() function for enums
func (i *Interpreter) builtinOrd(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Ord() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
	}

	// Handle boolean values (False=0, True=1)
	if boolVal, ok := arg.(*BooleanValue); ok {
		if boolVal.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	}

	// Handle integer values (pass through)
	if intVal, ok := arg.(*IntegerValue); ok {
		return intVal
	}

	return i.newErrorWithLocation(i.currentNode, "Ord() expects enum, boolean, or integer, got %s", arg.Type())
}

// builtinInteger implements the Integer() cast function.
// It converts values to integers.
// Task 8.52: Integer() cast for enums
func (i *Interpreter) builtinInteger(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Integer() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle enum values
	if enumVal, ok := arg.(*EnumValue); ok {
		return &IntegerValue{Value: int64(enumVal.OrdinalValue)}
	}

	// Handle integer values (pass through)
	if intVal, ok := arg.(*IntegerValue); ok {
		return intVal
	}

	// Handle float values (truncate)
	if floatVal, ok := arg.(*FloatValue); ok {
		return &IntegerValue{Value: int64(floatVal.Value)}
	}

	// Handle boolean values
	if boolVal, ok := arg.(*BooleanValue); ok {
		if boolVal.Value {
			return &IntegerValue{Value: 1}
		}
		return &IntegerValue{Value: 0}
	}

	return i.newErrorWithLocation(i.currentNode, "Integer() cannot convert %s to integer", arg.Type())
}

// builtinLength implements the Length() built-in function.
// It returns the number of elements in an array or characters in a string.
// Task 8.130: Length() function for arrays
func (i *Interpreter) builtinLength(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Length() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		// Return the number of elements in the array
		// For both static and dynamic arrays, this is len(Elements)
		return &IntegerValue{Value: int64(len(arrayVal.Elements))}
	}

	// Handle string values
	if strVal, ok := arg.(*StringValue); ok {
		// Return the number of characters in the string
		return &IntegerValue{Value: int64(len(strVal.Value))}
	}

	return i.newErrorWithLocation(i.currentNode, "Length() expects array or string, got %s", arg.Type())
}

// builtinCopy implements the Copy() built-in function.
// It returns a substring of a string.
// Copy(str, index, count) - index is 1-based, count is number of characters
// Task 8.183: Copy() function for strings
func (i *Interpreter) builtinCopy(args []Value) Value {
	if len(args) != 3 {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects exactly 3 arguments, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: index (1-based)
	indexVal, ok := args[1].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects integer as second argument, got %s", args[1].Type())
	}

	// Third argument: count
	countVal, ok := args[2].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Copy() expects integer as third argument, got %s", args[2].Type())
	}

	str := strVal.Value
	index := indexVal.Value // 1-based
	count := countVal.Value

	// Handle edge cases
	// If index is <= 0, return empty string (1-based indexing, so 0 and negative are invalid)
	if index <= 0 {
		return &StringValue{Value: ""}
	}

	// If count is <= 0, return empty string
	if count <= 0 {
		return &StringValue{Value: ""}
	}

	// Convert to 0-based index for Go
	startIdx := int(index - 1)

	// If start index is beyond string length, return empty string
	if startIdx >= len(str) {
		return &StringValue{Value: ""}
	}

	// Calculate end index
	endIdx := startIdx + int(count)

	// If end index goes beyond string length, adjust it
	if endIdx > len(str) {
		endIdx = len(str)
	}

	// Extract substring
	result := str[startIdx:endIdx]

	return &StringValue{Value: result}
}

// builtinConcat implements the Concat() built-in function.
// It concatenates multiple strings together.
// Concat(str1, str2, ...) - variable number of string arguments
// Task 8.183: Concat() function for strings
func (i *Interpreter) builtinConcat(args []Value) Value {
	if len(args) == 0 {
		return i.newErrorWithLocation(i.currentNode, "Concat() expects at least 1 argument, got 0")
	}

	// Build the concatenated string
	var result strings.Builder

	for idx, arg := range args {
		strVal, ok := arg.(*StringValue)
		if !ok {
			return i.newErrorWithLocation(i.currentNode, "Concat() expects string as argument %d, got %s", idx+1, arg.Type())
		}
		result.WriteString(strVal.Value)
	}

	return &StringValue{Value: result.String()}
}

// builtinPos implements the Pos() built-in function.
// It finds the position of a substring within a string.
// Pos(substr, str) - returns 1-based position (0 if not found)
// Task 8.183: Pos() function for strings
func (i *Interpreter) builtinPos(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument: substring to find
	substrVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects string as first argument, got %s", args[0].Type())
	}

	// Second argument: string to search in
	strVal, ok := args[1].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Pos() expects string as second argument, got %s", args[1].Type())
	}

	substr := substrVal.Value
	str := strVal.Value

	// Handle empty substring - returns 1 (found at start)
	if len(substr) == 0 {
		return &IntegerValue{Value: 1}
	}

	// Find the substring
	index := strings.Index(str, substr)

	// Convert to 1-based index (or 0 if not found)
	if index == -1 {
		return &IntegerValue{Value: 0}
	}

	return &IntegerValue{Value: int64(index + 1)}
}

// builtinUpperCase implements the UpperCase() built-in function.
// It converts a string to uppercase.
// UpperCase(str) - returns uppercase version of the string
// Task 8.183: UpperCase() function for strings
func (i *Interpreter) builtinUpperCase(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "UpperCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "UpperCase() expects string as argument, got %s", args[0].Type())
	}

	return &StringValue{Value: strings.ToUpper(strVal.Value)}
}

// builtinLowerCase implements the LowerCase() built-in function.
// It converts a string to lowercase.
// LowerCase(str) - returns lowercase version of the string
// Task 8.183: LowerCase() function for strings
func (i *Interpreter) builtinLowerCase(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "LowerCase() expects exactly 1 argument, got %d", len(args))
	}

	// First argument: string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "LowerCase() expects string as argument, got %s", args[0].Type())
	}

	return &StringValue{Value: strings.ToLower(strVal.Value)}
}

// builtinAbs implements the Abs() built-in function.
// It returns the absolute value of a number.
// Abs(x) - returns absolute value (Integer → Integer, Float → Float)
// Task 8.185: Abs() function for math operations
func (i *Interpreter) builtinAbs(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Abs() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle Integer
	if intVal, ok := arg.(*IntegerValue); ok {
		if intVal.Value < 0 {
			return &IntegerValue{Value: -intVal.Value}
		}
		return &IntegerValue{Value: intVal.Value}
	}

	// Handle Float
	if floatVal, ok := arg.(*FloatValue); ok {
		return &FloatValue{Value: math.Abs(floatVal.Value)}
	}

	return i.newErrorWithLocation(i.currentNode, "Abs() expects Integer or Float as argument, got %s", arg.Type())
}

// builtinSqrt implements the Sqrt() built-in function.
// It returns the square root of a number.
// Sqrt(x) - returns square root as Float (always returns Float, even for Integer input)
// Task 8.185: Sqrt() function for math operations
func (i *Interpreter) builtinSqrt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sqrt() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Sqrt() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for negative numbers
	if value < 0 {
		return i.newErrorWithLocation(i.currentNode, "Sqrt() of negative number (%f)", value)
	}

	return &FloatValue{Value: math.Sqrt(value)}
}

// builtinSin implements the Sin() built-in function.
// It returns the sine of a number (in radians).
// Sin(x) - returns sine as Float (always returns Float, even for Integer input)
// Task 8.185: Sin() function for trigonometric operations
func (i *Interpreter) builtinSin(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Sin() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Sin() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Sin(value)}
}

// builtinCos implements the Cos() built-in function.
// It returns the cosine of a number (in radians).
// Cos(x) - returns cosine as Float (always returns Float, even for Integer input)
// Task 8.185: Cos() function for trigonometric operations
func (i *Interpreter) builtinCos(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Cos() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Cos() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Cos(value)}
}

// builtinTan implements the Tan() built-in function.
// It returns the tangent of a number (in radians).
// Tan(x) - returns tangent as Float (always returns Float, even for Integer input)
// Task 8.185: Tan() function for trigonometric operations
func (i *Interpreter) builtinTan(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Tan() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Tan() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Tan(value)}
}

// builtinRandom implements the Random() built-in function.
// It returns a random Float between 0.0 (inclusive) and 1.0 (exclusive).
// Random() - returns random Float in [0, 1)
// Task 8.185: Random() function for random number generation
func (i *Interpreter) builtinRandom(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Random() expects no arguments, got %d", len(args))
	}

	return &FloatValue{Value: i.rand.Float64()}
}

// builtinRandomize implements the Randomize() built-in procedure.
// It seeds the random number generator with the current time.
// Randomize() - seeds RNG with current time (no return value)
// Task 8.185: Randomize() procedure for random number generation
func (i *Interpreter) builtinRandomize(args []Value) Value {
	if len(args) != 0 {
		return i.newErrorWithLocation(i.currentNode, "Randomize() expects no arguments, got %d", len(args))
	}

	// Re-seed the random number generator with current time
	i.rand.Seed(time.Now().UnixNano())
	return &NilValue{}
}

// builtinExp implements the Exp() built-in function.
// It returns e raised to the power of x.
// Exp(x) - returns e^x as Float (always returns Float, even for Integer input)
// Task 8.185: Exp() function for exponential operations
func (i *Interpreter) builtinExp(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Exp() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Exp() expects Integer or Float as argument, got %s", arg.Type())
	}

	return &FloatValue{Value: math.Exp(value)}
}

// builtinLn implements the Ln() built-in function.
// It returns the natural logarithm (base e) of x.
// Ln(x) - returns natural logarithm as Float (always returns Float, even for Integer input)
// Task 8.185: Ln() function for logarithmic operations
func (i *Interpreter) builtinLn(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Ln() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - convert to float
	if intVal, ok := arg.(*IntegerValue); ok {
		value = float64(intVal.Value)
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Ln() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Check for non-positive numbers (Ln is undefined for x <= 0)
	if value <= 0 {
		return i.newErrorWithLocation(i.currentNode, "Ln() of non-positive number (%f)", value)
	}

	return &FloatValue{Value: math.Log(value)}
}

// builtinRound implements the Round() built-in function.
// It rounds a number to the nearest integer.
// Round(x) - returns rounded value as Integer (always returns Integer)
// Task 8.185: Round() function for rounding operations
func (i *Interpreter) builtinRound(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Round() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Round() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Round to nearest integer
	rounded := math.Round(value)
	return &IntegerValue{Value: int64(rounded)}
}

// builtinTrunc implements the Trunc() built-in function.
// It truncates a number towards zero (removes the decimal part).
// Trunc(x) - returns truncated value as Integer (always returns Integer)
// Task 8.185: Trunc() function for truncation operations
func (i *Interpreter) builtinTrunc(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Trunc() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]
	var value float64

	// Handle Integer - already an integer, just return it
	if intVal, ok := arg.(*IntegerValue); ok {
		return &IntegerValue{Value: intVal.Value}
	} else if floatVal, ok := arg.(*FloatValue); ok {
		// Handle Float
		value = floatVal.Value
	} else {
		return i.newErrorWithLocation(i.currentNode, "Trunc() expects Integer or Float as argument, got %s", arg.Type())
	}

	// Truncate towards zero
	truncated := math.Trunc(value)
	return &IntegerValue{Value: int64(truncated)}
}

// builtinLow implements the Low() built-in function.
// It returns the lower bound of an array.
// Task 8.132: Low() function for arrays
func (i *Interpreter) builtinLow(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "Low() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return LowBound
		// For dynamic arrays, return 0
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.LowBound)}
		}
		return &IntegerValue{Value: 0}
	}

	return i.newErrorWithLocation(i.currentNode, "Low() expects array, got %s", arg.Type())
}

// builtinHigh implements the High() built-in function.
// It returns the upper bound of an array.
// Task 8.133: High() function for arrays
func (i *Interpreter) builtinHigh(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "High() expects exactly 1 argument, got %d", len(args))
	}

	arg := args[0]

	// Handle array values
	if arrayVal, ok := arg.(*ArrayValue); ok {
		if arrayVal.ArrayType == nil {
			return i.newErrorWithLocation(i.currentNode, "array has no type information")
		}

		// For static arrays, return HighBound
		// For dynamic arrays, return Length - 1
		if arrayVal.ArrayType.IsStatic() {
			return &IntegerValue{Value: int64(*arrayVal.ArrayType.HighBound)}
		}
		// Dynamic array: High = Length - 1
		return &IntegerValue{Value: int64(len(arrayVal.Elements) - 1)}
	}

	return i.newErrorWithLocation(i.currentNode, "High() expects array, got %s", arg.Type())
}

// builtinSetLength implements the SetLength() built-in function.
// It resizes a dynamic array to the specified length.
// Task 8.131: SetLength() function for dynamic arrays
func (i *Interpreter) builtinSetLength(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be an array (we'll need the variable name to modify it)
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects array as first argument, got %s", arrayArg.Type())
	}

	// Second argument must be an integer
	lengthArg := args[1]
	lengthInt, ok := lengthArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "SetLength() expects integer as second argument, got %s", lengthArg.Type())
	}

	newLength := int(lengthInt.Value)
	if newLength < 0 {
		return i.newErrorWithLocation(i.currentNode, "SetLength() new length cannot be negative: %d", newLength)
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "SetLength() can only be used with dynamic arrays, not static arrays")
	}

	// Resize the array
	oldLength := len(arrayVal.Elements)
	if newLength == oldLength {
		// No change needed
		return &NilValue{}
	}

	if newLength > oldLength {
		// Expand: add nil elements
		for j := oldLength; j < newLength; j++ {
			arrayVal.Elements = append(arrayVal.Elements, &NilValue{})
		}
	} else {
		// Shrink: truncate
		arrayVal.Elements = arrayVal.Elements[:newLength]
	}

	return &NilValue{}
}

// builtinAdd implements the Add() built-in function.
// It appends an element to the end of a dynamic array.
// Task 8.134: Add() function for dynamic arrays
func (i *Interpreter) builtinAdd(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Add() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Add() expects array as first argument, got %s", arrayArg.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "Add() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument is the element to add
	element := args[1]

	// Append the element to the array
	arrayVal.Elements = append(arrayVal.Elements, element)

	return &NilValue{}
}

// builtinDelete implements the Delete() built-in function.
// It removes an element at the specified index from a dynamic array.
// Task 8.135: Delete() function for dynamic arrays
func (i *Interpreter) builtinDelete(args []Value) Value {
	if len(args) != 2 {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects exactly 2 arguments, got %d", len(args))
	}

	// First argument must be a dynamic array
	arrayArg := args[0]
	arrayVal, ok := arrayArg.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects array as first argument, got %s", arrayArg.Type())
	}

	// Check that it's a dynamic array
	if arrayVal.ArrayType == nil {
		return i.newErrorWithLocation(i.currentNode, "array has no type information")
	}

	if arrayVal.ArrayType.IsStatic() {
		return i.newErrorWithLocation(i.currentNode, "Delete() can only be used with dynamic arrays, not static arrays")
	}

	// Second argument must be an integer (the index)
	indexArg := args[1]
	indexInt, ok := indexArg.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "Delete() expects integer as second argument, got %s", indexArg.Type())
	}

	index := int(indexInt.Value)

	// Validate index bounds (0-based for dynamic arrays)
	if index < 0 || index >= len(arrayVal.Elements) {
		return i.newErrorWithLocation(i.currentNode, "Delete() index out of bounds: %d (array length is %d)", index, len(arrayVal.Elements))
	}

	// Remove the element at index by slicing
	// Create a new slice with the element removed
	arrayVal.Elements = append(arrayVal.Elements[:index], arrayVal.Elements[index+1:]...)

	return &NilValue{}
}

// builtinIntToStr implements the IntToStr() built-in function.
// It converts an integer to its string representation.
// IntToStr(i: Integer): String
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinIntToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "IntToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be an integer
	intVal, ok := args[0].(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "IntToStr() expects integer argument, got %s", args[0].Type())
	}

	// Convert integer to string using Go's strconv
	result := fmt.Sprintf("%d", intVal.Value)
	return &StringValue{Value: result}
}

// builtinStrToInt implements the StrToInt() built-in function.
// It converts a string to an integer, raising an error if the string is invalid.
// StrToInt(s: String): Integer
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinStrToInt(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToInt() expects string argument, got %s", args[0].Type())
	}

	// Try to parse the string as an integer
	// Use strings.TrimSpace to handle leading/trailing whitespace
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseInt for strict parsing (doesn't accept partial matches)
	intValue, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid integer", strVal.Value)
	}

	return &IntegerValue{Value: intValue}
}

// builtinFloatToStr implements the FloatToStr() built-in function.
// It converts a float to its string representation.
// FloatToStr(f: Float): String
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinFloatToStr(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "FloatToStr() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a float
	floatVal, ok := args[0].(*FloatValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "FloatToStr() expects float argument, got %s", args[0].Type())
	}

	// Convert float to string using Go's strconv
	// Use 'g' format for general representation (like DWScript's FloatToStr)
	result := fmt.Sprintf("%g", floatVal.Value)
	return &StringValue{Value: result}
}

// builtinStrToFloat implements the StrToFloat() built-in function.
// It converts a string to a float, raising an error if the string is invalid.
// StrToFloat(s: String): Float
// Task 8.187: Type conversion functions
func (i *Interpreter) builtinStrToFloat(args []Value) Value {
	if len(args) != 1 {
		return i.newErrorWithLocation(i.currentNode, "StrToFloat() expects exactly 1 argument, got %d", len(args))
	}

	// Argument must be a string
	strVal, ok := args[0].(*StringValue)
	if !ok {
		return i.newErrorWithLocation(i.currentNode, "StrToFloat() expects string argument, got %s", args[0].Type())
	}

	// Try to parse the string as a float
	s := strings.TrimSpace(strVal.Value)

	// Use strconv.ParseFloat for strict parsing (doesn't accept partial matches)
	floatValue, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return i.newErrorWithLocation(i.currentNode, "'%s' is not a valid float", strVal.Value)
	}

	return &FloatValue{Value: floatValue}
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

		// Task 8.235m: Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			continue
		}
		// Task 8.235m: Handle exit signal (exit from function while in loop)
		if i.exitSignal {
			// Don't clear the signal - let the function handle it
			break
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

		// Task 8.235m: Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			// Continue to condition check
		}
		// Task 8.235m: Handle exit signal (exit from function while in loop)
		if i.exitSignal {
			// Don't clear the signal - let the function handle it
			break
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

			// Task 8.235m: Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Task 8.235m: Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
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

			// Task 8.235m: Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Task 8.235m: Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
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

// evalBreakStatement evaluates a break statement (Task 8.235j).
// Sets the break signal to exit the innermost loop.
func (i *Interpreter) evalBreakStatement(_ *ast.BreakStatement) Value {
	i.breakSignal = true
	return &NilValue{}
}

// evalContinueStatement evaluates a continue statement (Task 8.235k).
// Sets the continue signal to skip to the next iteration of the innermost loop.
func (i *Interpreter) evalContinueStatement(_ *ast.ContinueStatement) Value {
	i.continueSignal = true
	return &NilValue{}
}

// evalExitStatement evaluates an exit statement (Task 8.235l).
// Sets the exit signal to exit the current function.
// If at program level, sets exit signal to terminate the program.
func (i *Interpreter) evalExitStatement(_ *ast.ExitStatement) Value {
	i.exitSignal = true
	// Exit doesn't return a value - the function's default return value is used
	// (or the program exits if at top level)
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
	case *RecordValue:
		r, ok := right.(*RecordValue)
		if !ok {
			return false
		}
		return i.recordsEqual(l, r)
	default:
		// For other types, use string comparison as fallback
		return left.String() == right.String()
	}
}

// evalFunctionDeclaration evaluates a function declaration.
// It registers the function in the function registry without executing its body.
// For method implementations (fn.ClassName != nil), it updates the class's Methods map.
func (i *Interpreter) evalFunctionDeclaration(fn *ast.FunctionDecl) Value {
	// Check if this is a method implementation (has a class name like TExample.Method)
	if fn.ClassName != nil {
		className := fn.ClassName.Value
		classInfo, exists := i.classes[className]
		if !exists {
			return i.newErrorWithLocation(fn, "class '%s' not found for method '%s'", className, fn.Name.Value)
		}

		// Update the method in the class (replacing the declaration with the implementation)
		if fn.IsClassMethod {
			classInfo.ClassMethods[fn.Name.Value] = fn
		} else {
			classInfo.Methods[fn.Name.Value] = fn
		}

		// Also store constructors
		if fn.IsConstructor {
			classInfo.Constructors[fn.Name.Value] = fn
			// Always update Constructor to use the implementation (which has the body)
			// This replaces the declaration that was set during class parsing
			classInfo.Constructor = fn
		}

		// Store destructor
		if fn.IsDestructor {
			classInfo.Destructor = fn
		}

		return &NilValue{}
	}

	// Store regular function in the registry
	i.functions[fn.Name.Value] = fn
	return &NilValue{}
}

// evalClassDeclaration evaluates a class declaration.
// It builds a ClassInfo from the AST and registers it in the class registry.
// Handles inheritance by copying parent fields and methods to the child class.
func (i *Interpreter) evalClassDeclaration(cd *ast.ClassDecl) Value {
	// Create new ClassInfo
	classInfo := NewClassInfo(cd.Name.Value)

	// Set abstract flag (Task 7.65)
	classInfo.IsAbstract = cd.IsAbstract

	// Set external flags (Task 7.141)
	classInfo.IsExternal = cd.IsExternal
	classInfo.ExternalName = cd.ExternalName

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

		// Copy class methods
		for methodName, methodDecl := range parentClass.ClassMethods {
			classInfo.ClassMethods[methodName] = methodDecl
		}

		// Copy constructors
		for name, constructor := range parentClass.Constructors {
			classInfo.Constructors[name] = constructor
		}

		// Copy operator overloads
		classInfo.Operators = parentClass.Operators.clone()
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

		if method.IsConstructor {
			classInfo.Constructors[method.Name.Value] = method
		}
	}

	// Identify constructor (method named "Create")
	if constructor, exists := classInfo.Methods["Create"]; exists {
		classInfo.Constructor = constructor
	}
	if cd.Constructor != nil {
		classInfo.Constructors[cd.Constructor.Name.Value] = cd.Constructor
	}

	// Identify destructor (method named "Destroy")
	if destructor, exists := classInfo.Methods["Destroy"]; exists {
		classInfo.Destructor = destructor
	}

	// Register properties (Task 8.53 - copy property metadata from AST)
	// Properties are registered after fields and methods so they can reference them
	for _, propDecl := range cd.Properties {
		if propDecl == nil {
			continue
		}

		// Convert AST property to PropertyInfo
		propInfo := i.convertPropertyDecl(propDecl)
		if propInfo != nil {
			classInfo.Properties[propDecl.Name.Value] = propInfo
		}
	}

	// Copy parent properties (child inherits all parent properties)
	if classInfo.Parent != nil {
		for propName, propInfo := range classInfo.Parent.Properties {
			// Only copy if not already defined in child class
			if _, exists := classInfo.Properties[propName]; !exists {
				classInfo.Properties[propName] = propInfo
			}
		}
	}

	// Register class operators (Stage 8)
	for _, opDecl := range cd.Operators {
		if opDecl == nil {
			continue
		}
		if errVal := i.registerClassOperator(classInfo, opDecl); isError(errVal) {
			return errVal
		}
	}

	// Register class in registry
	i.classes[classInfo.Name] = classInfo

	return &NilValue{}
}

// convertPropertyDecl converts an AST property declaration to a PropertyInfo struct.
// This extracts the property metadata for runtime property access handling. (Task 8.53)
// Note: We mark all identifiers as field access for now and will check at runtime
// whether they're actually fields or methods.
func (i *Interpreter) convertPropertyDecl(propDecl *ast.PropertyDecl) *types.PropertyInfo {
	// Resolve property type
	var propType types.Type
	switch propDecl.Type.Name {
	case "Integer":
		propType = types.INTEGER
	case "Float":
		propType = types.FLOAT
	case "String":
		propType = types.STRING
	case "Boolean":
		propType = types.BOOLEAN
	default:
		// For now, treat unknown types as NIL
		// In a full implementation, we'd look up custom types
		propType = types.NIL
	}

	propInfo := &types.PropertyInfo{
		Name:      propDecl.Name.Value,
		Type:      propType,
		IsIndexed: propDecl.IndexParams != nil && len(propDecl.IndexParams) > 0,
		IsDefault: propDecl.IsDefault,
	}

	// Determine read access kind and spec
	if propDecl.ReadSpec != nil {
		if ident, ok := propDecl.ReadSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.ReadSpec = ident.Value
			// Mark as field for now - evalPropertyRead will check both fields and methods
			propInfo.ReadKind = types.PropAccessField
		} else {
			// It's an expression
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadSpec = propDecl.ReadSpec.String()
		}
	} else {
		propInfo.ReadKind = types.PropAccessNone
	}

	// Determine write access kind and spec
	if propDecl.WriteSpec != nil {
		if ident, ok := propDecl.WriteSpec.(*ast.Identifier); ok {
			// It's an identifier - store the name, we'll check if it's a field or method at access time
			propInfo.WriteSpec = ident.Value
			// Mark as field for now - evalPropertyWrite will check both fields and methods
			propInfo.WriteKind = types.PropAccessField
		} else {
			propInfo.WriteKind = types.PropAccessNone
		}
	} else {
		propInfo.WriteKind = types.PropAccessNone
	}

	return propInfo
}

// evalInterfaceDeclaration evaluates an interface declaration (Task 7.119).
// It builds an InterfaceInfo from the AST and registers it in the interface registry.
// Handles inheritance by linking to parent interface and inheriting its methods.
func (i *Interpreter) evalInterfaceDeclaration(id *ast.InterfaceDecl) Value {
	// Create new InterfaceInfo
	interfaceInfo := NewInterfaceInfo(id.Name.Value)

	// Handle inheritance if parent interface is specified
	if id.Parent != nil {
		parentName := id.Parent.Value
		parentInterface, exists := i.interfaces[parentName]
		if !exists {
			return i.newErrorWithLocation(id, "parent interface '%s' not found", parentName)
		}

		// Set parent reference
		interfaceInfo.Parent = parentInterface

		// Note: We don't copy parent methods here because InterfaceInfo.GetMethod()
		// and AllMethods() already handle parent interface traversal
	}

	// Add methods to InterfaceInfo
	// Convert InterfaceMethodDecl nodes to FunctionDecl nodes for consistency
	for _, methodDecl := range id.Methods {
		// Create a FunctionDecl from the InterfaceMethodDecl
		// Interface methods are declarations only (no body)
		funcDecl := &ast.FunctionDecl{
			Token:      methodDecl.Token,
			Name:       methodDecl.Name,
			Parameters: methodDecl.Parameters,
			ReturnType: methodDecl.ReturnType,
			Body:       nil, // Interface methods have no body
		}

		interfaceInfo.Methods[methodDecl.Name.Value] = funcDecl
	}

	// Register interface in registry
	i.interfaces[interfaceInfo.Name] = interfaceInfo

	return &NilValue{}
}

func (i *Interpreter) evalOperatorDeclaration(decl *ast.OperatorDecl) Value {
	if decl.Kind == ast.OperatorKindClass {
		// Class operators are registered during class declaration evaluation
		return &NilValue{}
	}

	if decl.Binding == nil {
		return i.newErrorWithLocation(decl, "operator '%s' missing binding", decl.OperatorSymbol)
	}

	operandTypes := make([]string, len(decl.OperandTypes))
	for idx, operand := range decl.OperandTypes {
		opRand := operand.String()
		operandTypes[idx] = normalizeTypeAnnotation(opRand)
	}

	if decl.Kind == ast.OperatorKindConversion {
		if len(operandTypes) != 1 {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires exactly one operand", decl.OperatorSymbol)
		}
		if decl.ReturnType == nil {
			return i.newErrorWithLocation(decl, "conversion operator '%s' requires a return type", decl.OperatorSymbol)
		}
		targetType := normalizeTypeAnnotation(decl.ReturnType.String())
		entry := &runtimeConversionEntry{
			From:        operandTypes[0],
			To:          targetType,
			BindingName: decl.Binding.Value,
			Implicit:    strings.EqualFold(decl.OperatorSymbol, "implicit"),
		}
		if err := i.conversions.register(entry); err != nil {
			return i.newErrorWithLocation(decl, "conversion from %s to %s already defined", operandTypes[0], targetType)
		}
		return &NilValue{}
	}

	entry := &runtimeOperatorEntry{
		Operator:     decl.OperatorSymbol,
		OperandTypes: operandTypes,
		BindingName:  decl.Binding.Value,
	}

	if err := i.globalOperators.register(entry); err != nil {
		return i.newErrorWithLocation(decl, "operator '%s' already defined for operand types (%s)", decl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	return &NilValue{}
}

func (i *Interpreter) registerClassOperator(classInfo *ClassInfo, opDecl *ast.OperatorDecl) Value {
	if opDecl.Binding == nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' missing binding", opDecl.OperatorSymbol)
	}

	bindingName := opDecl.Binding.Value
	method, isClassMethod := classInfo.ClassMethods[bindingName]
	if !isClassMethod {
		var ok bool
		method, ok = classInfo.Methods[bindingName]
		if !ok {
			return i.newErrorWithLocation(opDecl, "binding '%s' for class operator '%s' not found in class '%s'", bindingName, opDecl.OperatorSymbol, classInfo.Name)
		}
	}

	classKey := normalizeTypeAnnotation(classInfo.Name)
	operandTypes := make([]string, 0, len(opDecl.OperandTypes)+1)
	includesClass := false
	for _, operand := range opDecl.OperandTypes {
		key := normalizeTypeAnnotation(operand.String())
		if key == classKey {
			includesClass = true
		}
		operandTypes = append(operandTypes, key)
	}
	if !includesClass {
		if strings.EqualFold(opDecl.OperatorSymbol, "in") {
			operandTypes = append(operandTypes, classKey)
		} else {
			operandTypes = append([]string{classKey}, operandTypes...)
		}
	}

	selfIndex := -1
	if !isClassMethod {
		for idx, key := range operandTypes {
			if key == classKey {
				selfIndex = idx
				break
			}
		}
		if selfIndex == -1 {
			return i.newErrorWithLocation(opDecl, "unable to determine self operand for class operator '%s'", opDecl.OperatorSymbol)
		}
	}

	entry := &runtimeOperatorEntry{
		Operator:      opDecl.OperatorSymbol,
		OperandTypes:  operandTypes,
		BindingName:   bindingName,
		Class:         classInfo,
		IsClassMethod: isClassMethod,
		SelfIndex:     selfIndex,
	}

	if err := classInfo.Operators.register(entry); err != nil {
		return i.newErrorWithLocation(opDecl, "class operator '%s' already defined for operand types (%s)", opDecl.OperatorSymbol, strings.Join(operandTypes, ", "))
	}

	if method.IsConstructor {
		classInfo.Constructors[method.Name.Value] = method
	}

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

	// Check if trying to instantiate an abstract class (Task 7.65)
	if classInfo.IsAbstract {
		return i.newErrorWithLocation(ne, "cannot instantiate abstract class '%s'", className)
	}

	// Check if trying to instantiate an external class (Task 7.141)
	// External classes are implemented outside DWScript and cannot be instantiated directly
	// Future: Provide hooks for Go FFI implementation
	if classInfo.IsExternal {
		return i.newErrorWithLocation(ne, "cannot instantiate external class '%s' - external classes are not supported", className)
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

	// Special handling for Exception.Create (Task 8.218)
	// Exception constructors are built-in and take a message parameter
	// NewExpression always implies Create constructor in DWScript
	if i.isExceptionClass(classInfo) && len(ne.Arguments) == 1 {
		// Evaluate the message argument
		msgVal := i.Eval(ne.Arguments[0])
		if isError(msgVal) {
			return msgVal
		}
		// Set the Message field
		if strVal, ok := msgVal.(*StringValue); ok {
			obj.SetField("Message", strVal)
		} else {
			obj.SetField("Message", &StringValue{Value: msgVal.String()})
		}
		return obj
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

		// Check if this identifier refers to an enum type (for scoped access: TColor.Red)
		// Look for enum type metadata stored in environment
		enumTypeKey := "__enum_type_" + ident.Value
		if enumTypeVal, ok := i.env.Get(enumTypeKey); ok {
			if _, isEnumType := enumTypeVal.(*EnumTypeValue); isEnumType {
				// This is scoped enum access: TColor.Red
				// Look up the enum value
				valueName := ma.Member.Value
				if val, exists := i.env.Get(valueName); exists {
					if enumVal, isEnum := val.(*EnumValue); isEnum {
						// Verify the value belongs to this enum type
						if enumVal.TypeName == ident.Value {
							return enumVal
						}
					}
				}
				// Enum value not found
				return i.newErrorWithLocation(ma, "enum value '%s' not found in enum '%s'", ma.Member.Value, ident.Value)
			}
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(ma.Object)
	if isError(objVal) {
		return objVal
	}

	// Task 8.75: Check if it's a record value
	if recordVal, ok := objVal.(*RecordValue); ok {
		// Access record field
		fieldValue, exists := recordVal.Fields[ma.Member.Value]
		if !exists {
			return i.newErrorWithLocation(ma, "field '%s' not found in record '%s'", ma.Member.Value, recordVal.RecordType.Name)
		}
		return fieldValue
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.newErrorWithLocation(ma, "cannot access member of non-object type '%s'", objVal.Type())
	}

	memberName := ma.Member.Value

	// Handle built-in properties/methods available on all objects (inherited from TObject)
	if memberName == "ClassName" {
		// ClassName returns the runtime type name of the object
		return &StringValue{Value: obj.Class.Name}
	}

	// Task 8.53: Check if this is a property access (properties take precedence over fields)
	if propInfo := obj.Class.lookupProperty(memberName); propInfo != nil {
		return i.evalPropertyRead(obj, propInfo, ma)
	}

	// Not a property - try direct field access
	fieldValue := obj.GetField(memberName)
	if fieldValue == nil {
		return i.newErrorWithLocation(ma, "field '%s' not found in class '%s'", memberName, obj.Class.Name)
	}

	return fieldValue
}

// evalPropertyRead evaluates a property read access. (Task 8.53)
// Handles field-backed, method-backed, and expression-backed properties.
func (i *Interpreter) evalPropertyRead(obj *ObjectInstance, propInfo *types.PropertyInfo, node ast.Node) Value {
	switch propInfo.ReadKind {
	case types.PropAccessField:
		// Task 8.53a: Field or method access - check at runtime which it is
		// First try as a field
		if _, exists := obj.Class.Fields[propInfo.ReadSpec]; exists {
			fieldValue := obj.GetField(propInfo.ReadSpec)
			if fieldValue == nil {
				return i.newErrorWithLocation(node, "property '%s' read field '%s' not found", propInfo.Name, propInfo.ReadSpec)
			}
			return fieldValue
		}

		// Not a field - try as a method (getter)
		method := obj.Class.lookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' read specifier '%s' not found as field or method", propInfo.Name, propInfo.ReadSpec)
		}

		// Call the getter method
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			i.env.Define("Result", &NilValue{})
			i.env.Define(method.Name.Value, &NilValue{})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessMethod:
		// Task 8.53b: Method access - call getter method
		// Check if method exists
		method := obj.Class.lookupMethod(propInfo.ReadSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' getter method '%s' not found", propInfo.Name, propInfo.ReadSpec)
		}

		// Call the getter method with no arguments (indexed properties handled separately)
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// For functions, initialize the Result variable
		if method.ReturnType != nil {
			i.env.Define("Result", &NilValue{})
			i.env.Define(method.Name.Value, &NilValue{})
		}

		// Execute method body
		i.Eval(method.Body)

		// Get return value
		var returnValue Value
		if method.ReturnType != nil {
			if resultVal, ok := i.env.Get("Result"); ok {
				returnValue = resultVal
			} else if methodNameVal, ok := i.env.Get(method.Name.Value); ok {
				returnValue = methodNameVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		// Restore environment
		i.env = savedEnv

		return returnValue

	case types.PropAccessExpression:
		// Task 8.53c / 8.56: Expression access - evaluate expression in context of object
		// For now, return an error as expression evaluation is complex
		return i.newErrorWithLocation(node, "expression-based property getters not yet supported")

	default:
		return i.newErrorWithLocation(node, "property '%s' has no read access", propInfo.Name)
	}
}

// evalPropertyWrite evaluates a property write access. (Task 8.54)
// Handles field-backed and method-backed property setters.
func (i *Interpreter) evalPropertyWrite(obj *ObjectInstance, propInfo *types.PropertyInfo, value Value, node ast.Node) Value {
	switch propInfo.WriteKind {
	case types.PropAccessField:
		// Task 8.54a: Field or method access - check at runtime which it is
		// First try as a field
		if _, exists := obj.Class.Fields[propInfo.WriteSpec]; exists {
			obj.SetField(propInfo.WriteSpec, value)
			return value
		}

		// Not a field - try as a method (setter)
		method := obj.Class.lookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' write specifier '%s' not found as field or method", propInfo.Name, propInfo.WriteSpec)
		}

		// Call the setter method with the value as argument
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind the value parameter (setter should have exactly one parameter)
		if len(method.Parameters) >= 1 {
			paramName := method.Parameters[0].Name.Value
			i.env.Define(paramName, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessMethod:
		// Task 8.54b: Method access - call setter method with value
		// Check if method exists
		method := obj.Class.lookupMethod(propInfo.WriteSpec)
		if method == nil {
			return i.newErrorWithLocation(node, "property '%s' setter method '%s' not found", propInfo.Name, propInfo.WriteSpec)
		}

		// Call the setter method with the value as argument
		// Create method environment with Self bound to object
		methodEnv := NewEnclosedEnvironment(i.env)
		savedEnv := i.env
		i.env = methodEnv

		// Bind Self to the object
		i.env.Define("Self", obj)

		// Bind the value parameter (setter should have exactly one parameter)
		if len(method.Parameters) >= 1 {
			i.env.Define(method.Parameters[0].Name.Value, value)
		}

		// Execute method body
		i.Eval(method.Body)

		// Restore environment
		i.env = savedEnv

		return value

	case types.PropAccessNone:
		// Read-only property
		return i.newErrorWithLocation(node, "property '%s' is read-only", propInfo.Name)

	default:
		return i.newErrorWithLocation(node, "property '%s' has no write access", propInfo.Name)
	}
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

				// Bind method parameters to arguments with implicit conversion
				for idx, param := range classMethod.Parameters {
					arg := args[idx]

					// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
					if param.Type != nil {
						paramTypeName := param.Type.Name
						if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
							arg = converted
						}
					}

					i.env.Define(param.Name.Value, arg)
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

					// Task 8.19c: Apply implicit conversion if return type doesn't match
					if returnValue.Type() != "NIL" {
						expectedReturnType := classMethod.ReturnType.Name
						if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
							returnValue = converted
						}
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

				// Bind method parameters to arguments with implicit conversion
				for idx, param := range instanceMethod.Parameters {
					arg := args[idx]

					// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
					if param.Type != nil {
						paramTypeName := param.Type.Name
						if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
							arg = converted
						}
					}

					i.env.Define(param.Name.Value, arg)
				}

				// For functions (not procedures), initialize the Result variable
				// For constructors, always initialize Result even if no explicit return type
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
					i.env.Define("Result", &NilValue{})
					// Also define the method name as an alias for Result (DWScript style)
					i.env.Define(instanceMethod.Name.Value, &NilValue{})
				}

				// Execute method body
				result := i.Eval(instanceMethod.Body)
				if isError(result) {
					i.env = savedEnv
					return result
				}

				// Extract return value (same logic as regular functions)
				var returnValue Value
				if instanceMethod.ReturnType != nil || instanceMethod.IsConstructor {
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
					} else if instanceMethod.IsConstructor {
						// Constructors return the object instance by default
						returnValue = obj
					} else {
						returnValue = &NilValue{}
					}

					// Task 8.19c: Apply implicit conversion if return type doesn't match (but not for constructors)
					if instanceMethod.ReturnType != nil && returnValue.Type() != "NIL" {
						expectedReturnType := instanceMethod.ReturnType.Name
						if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
							returnValue = converted
						}
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

	// Handle built-in methods that are available on all objects (inherited from TObject)
	if mc.Method.Value == "ClassName" {
		// ClassName returns the runtime type name of the object
		return &StringValue{Value: obj.Class.Name}
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

	// Bind method parameters to arguments with implicit conversion
	for idx, param := range method.Parameters {
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		i.env.Define(param.Name.Value, arg)
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

		// Task 8.19c: Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := method.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
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
		arg := args[idx]

		// Task 8.19b: Apply implicit conversion if parameter has a type and types don't match
		if param.Type != nil {
			paramTypeName := param.Type.Name
			if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
				arg = converted
			}
		}

		if param.ByRef {
			// By-reference parameter - we need to handle this specially
			// For now, we'll pass by value (TODO: implement proper by-ref support)
			i.env.Define(param.Name.Value, arg)
		} else {
			// By-value parameter
			i.env.Define(param.Name.Value, arg)
		}
	}

	// For functions (not procedures), initialize the Result variable
	if fn.ReturnType != nil {
		// Initialize Result based on return type
		var resultValue Value = &NilValue{}

		// Check if return type is a record
		returnTypeName := fn.ReturnType.Name
		recordTypeKey := "__record_type_" + returnTypeName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				resultValue = NewRecordValue(rtv.RecordType)
			}
		}

		i.env.Define("Result", resultValue)
		// Also define the function name as an alias for Result (DWScript style)
		i.env.Define(fn.Name.Value, resultValue)
	}

	// Execute the function body
	if fn.Body == nil {
		// Function has no body (forward declaration) - this is an error
		i.env = savedEnv
		return newError("function '%s' has no body", fn.Name.Value)
	}

	i.Eval(fn.Body)

	// Task 8.235n: Handle exit signal
	// If exit was called, clear the signal (don't propagate to caller)
	if i.exitSignal {
		i.exitSignal = false
		// Exit was called, function returns immediately with current Result value
	}

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

		// Task 8.19c: Apply implicit conversion if return type doesn't match
		if returnValue.Type() != "NIL" {
			expectedReturnType := fn.ReturnType.Name
			if converted, ok := i.tryImplicitConversion(returnValue, expectedReturnType); ok {
				returnValue = converted
			}
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
