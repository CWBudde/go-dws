package evaluator

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

// This file contains visitor methods for expression AST nodes.
// Phase 3.5.2: Visitor pattern implementation for expressions.
//
// Expressions evaluate to values and can be nested (e.g., binary expressions
// contain left and right sub-expressions).

// ErrorValue represents a runtime error (temporary definition to avoid circular imports).
type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() string   { return "ERROR" }
func (e *ErrorValue) String() string { return "ERROR: " + e.Message }

// newError creates a new error value with optional formatting.
// TODO: Add location information from node in Phase 3.6 (error handling improvements)
func (e *Evaluator) newError(_ ast.Node, format string, args ...interface{}) Value {
	return &ErrorValue{Message: fmt.Sprintf(format, args...)}
}

// isError checks if a value is an error.
func isError(val Value) bool {
	if val != nil {
		return val.Type() == "ERROR"
	}
	return false
}

// VisitIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) VisitIdentifier(node *ast.Identifier, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type system is available via adapter
	// (LookupClass, LookupFunction, etc.)
	// TODO: Migrate identifier lookup logic to use adapter type system methods
	return e.adapter.EvalNode(node)
}

// VisitBinaryExpression evaluates a binary expression (e.g., a + b, x == y).
func (e *Evaluator) VisitBinaryExpression(node *ast.BinaryExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Operator registry is available via adapter.GetOperatorRegistry()
	// Conversion registry available via adapter.GetConversionRegistry()
	// TODO: Migrate operator evaluation and type coercion logic
	return e.adapter.EvalNode(node)
}

// VisitUnaryExpression evaluates a unary expression (e.g., -x, not b).
func (e *Evaluator) VisitUnaryExpression(node *ast.UnaryExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Operator registry is available via adapter.GetOperatorRegistry()
	// TODO: Migrate unary operator evaluation logic
	return e.adapter.EvalNode(node)
}

// VisitAddressOfExpression evaluates an address-of expression (@funcName).
// Phase 3.5.4 - Phase 2A: Infrastructure ready, full migration pending type migration
func (e *Evaluator) VisitAddressOfExpression(node *ast.AddressOfExpression, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.10: Function lookup available via adapter.LookupFunction()
	// Full migration pending - complex logic for method pointers and function overloading
	return e.adapter.EvalNode(node)
}

// VisitGroupedExpression evaluates a grouped expression (parenthesized).
func (e *Evaluator) VisitGroupedExpression(node *ast.GroupedExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4.11: Grouped expressions just evaluate their inner expression
	// Parentheses are only for precedence, they don't change the value
	return e.Eval(node.Expression, ctx)
}

// VisitCallExpression evaluates a function call expression.
// Phase 3.5.4 - Phase 2A: Infrastructure ready, full migration pending type migration
func (e *Evaluator) VisitCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.22: Function call methods available via adapter
	// (CallFunctionPointer, CallUserFunction, CallBuiltinFunction, LookupFunction)
	// Full migration pending FunctionPointerValue migration to runtime package
	return e.adapter.EvalNode(node)
}

// VisitNewExpression evaluates a 'new' expression (object instantiation).
func (e *Evaluator) VisitNewExpression(node *ast.NewExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Class registry is available via adapter.LookupClass()
	// TODO: Migrate object instantiation and constructor dispatch logic
	return e.adapter.EvalNode(node)
}

// VisitMemberAccessExpression evaluates member access (obj.field, obj.method).
func (e *Evaluator) VisitMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2C: Property/indexing infrastructure available
	// PropertyEvalContext accessible via ctx.PropContext() for recursion prevention
	// Property dispatch uses Phase 2A (function calls) + Phase 2B (type lookups)
	// TODO: Migrate member access logic (property getters, field access, helper methods)
	return e.adapter.EvalNode(node)
}

// VisitMethodCallExpression evaluates a method call (obj.Method(args)).
// Phase 3.5.4 - Phase 2A: Infrastructure ready, full migration pending type migration
func (e *Evaluator) VisitMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.16: Method call via adapter.CallUserFunction()
	// Full migration pending ObjectInstance/ClassInfo migration to runtime package
	return e.adapter.EvalNode(node)
}

// VisitInheritedExpression evaluates an 'inherited' expression.
// Phase 3.5.4 - Phase 2A: Infrastructure ready, full migration pending type migration
func (e *Evaluator) VisitInheritedExpression(node *ast.InheritedExpression, ctx *ExecutionContext) Value {
	// TODO Phase 3.5.4.18: Parent method dispatch via adapter.CallUserFunction()
	// Full migration pending ObjectInstance/ClassInfo migration to runtime package
	return e.adapter.EvalNode(node)
}

// VisitSelfExpression evaluates a 'Self' expression.
// Phase 3.5.4.17: Migrated from Interpreter.evalSelfExpression()
// Self refers to the current instance (in instance methods) or the current class (in class methods).
// Task 9.7: Implement Self keyword
func (e *Evaluator) VisitSelfExpression(node *ast.SelfExpression, ctx *ExecutionContext) Value {
	// Get Self from the environment (should be bound when entering methods)
	selfVal, exists := ctx.Env().Get("Self")
	if !exists {
		return e.newError(node, "Self used outside method context")
	}

	// Convert interface{} to Value
	val, ok := selfVal.(Value)
	if !ok {
		return e.newError(node, "Self has invalid type")
	}

	return val
}

// VisitEnumLiteral evaluates an enum literal (EnumType.Value).
func (e *Evaluator) VisitEnumLiteral(node *ast.EnumLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.4.12: Enum literals are looked up in the environment
	// The semantic analyzer validates enum types and values exist
	if node == nil {
		return e.newError(node, "nil enum literal")
	}

	valueName := node.ValueName

	// Look up the value in the environment
	val, ok := ctx.Env().Get(valueName)
	if !ok {
		return e.newError(node, "undefined enum value '%s'", valueName)
	}

	// Environment stores interface{}, cast to Value
	// The semantic analyzer ensures this is a valid enum value
	if value, ok := val.(Value); ok {
		return value
	}

	// Should never happen if semantic analysis passed
	return e.newError(node, "enum value '%s' has invalid type", valueName)
}

// VisitRecordLiteralExpression evaluates a record literal expression.
func (e *Evaluator) VisitRecordLiteralExpression(node *ast.RecordLiteralExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2C: Record construction infrastructure available
	// Record registry accessible via adapter.LookupRecord() (Phase 2B)
	// TODO: Migrate record literal construction logic
	return e.adapter.EvalNode(node)
}

// VisitSetLiteral evaluates a set literal [value1, value2, ...].
func (e *Evaluator) VisitSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitArrayLiteralExpression evaluates an array literal [1, 2, 3].
func (e *Evaluator) VisitArrayLiteralExpression(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2C: Array construction infrastructure available
	// Type inference uses Phase 2B type system
	// TODO: Migrate array literal construction logic with type inference
	return e.adapter.EvalNode(node)
}

// VisitIndexExpression evaluates an index expression array[index].
func (e *Evaluator) VisitIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2C: Array/property indexing infrastructure available
	// Bounds checking and property indexing handled via EvalNode delegation
	// PropertyEvalContext accessible via ctx.PropContext() for property indexers
	// TODO: Migrate indexing logic (array bounds checking, property indexers, string indexing)
	return e.adapter.EvalNode(node)
}

// VisitNewArrayExpression evaluates a new array expression.
func (e *Evaluator) VisitNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitLambdaExpression evaluates a lambda expression (closure).
// Task 3.5.8: Migrated using adapter.CreateLambda()
func (e *Evaluator) VisitLambdaExpression(node *ast.LambdaExpression, ctx *ExecutionContext) Value {
	// Create lambda with current environment as closure
	// The lambda captures the current scope
	return e.adapter.CreateLambda(node, ctx.Env())
}

// VisitIsExpression evaluates an 'is' type checking expression.
func (e *Evaluator) VisitIsExpression(node *ast.IsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Class registry available via adapter.LookupClass()
	// TODO: Migrate class hierarchy checking logic - complex with boolean comparison mode
	return e.adapter.EvalNode(node)
}

// VisitAsExpression evaluates an 'as' type casting expression.
func (e *Evaluator) VisitAsExpression(node *ast.AsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Type casting infrastructure via adapter
	// TODO: Migrate type casting logic - complex with interface handling
	return e.adapter.EvalNode(node)
}

// VisitImplementsExpression evaluates an 'implements' interface checking expression.
func (e *Evaluator) VisitImplementsExpression(node *ast.ImplementsExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4 - Phase 2B: Interface registry available via adapter.LookupInterface()
	// TODO: Migrate interface checking logic
	return e.adapter.EvalNode(node)
}

// VisitIfExpression evaluates an inline if-then-else expression.
func (e *Evaluator) VisitIfExpression(node *ast.IfExpression, ctx *ExecutionContext) Value {
	// Phase 3.5.4.13: Migrated if expression evaluation with type defaults
	// Evaluate the condition
	condition := e.Eval(node.Condition, ctx)
	if isError(condition) {
		return condition
	}

	// Use isTruthy to support Variant→Boolean implicit conversion
	// If condition is true, evaluate and return consequence
	if isTruthy(condition) {
		result := e.Eval(node.Consequence, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// Condition is false
	if node.Alternative != nil {
		// Evaluate and return alternative
		result := e.Eval(node.Alternative, ctx)
		if isError(result) {
			return result
		}
		return result
	}

	// No else clause - return default value for the consequence type
	// The type should have been set during semantic analysis
	var typeAnnot *ast.TypeAnnotation
	if e.semanticInfo != nil {
		typeAnnot = e.semanticInfo.GetType(node)
	}
	if typeAnnot == nil {
		return e.newError(node, "if expression missing type annotation")
	}

	// Return default value based on type name
	typeName := strings.ToLower(typeAnnot.Name)
	switch typeName {
	case "integer", "int64":
		return &runtime.IntegerValue{Value: 0}
	case "float", "float64", "double", "real":
		return &runtime.FloatValue{Value: 0.0}
	case "string":
		return &runtime.StringValue{Value: ""}
	case "boolean", "bool":
		return &runtime.BooleanValue{Value: false}
	default:
		// For class types and other reference types, return nil
		return &runtime.NilValue{}
	}
}

// VisitOldExpression evaluates an 'old' expression (used in postconditions).
func (e *Evaluator) VisitOldExpression(node *ast.OldExpression, ctx *ExecutionContext) Value {
	// Phase 2.1: Migrated old expression evaluation
	// Get the identifier name from the old expression
	identName := node.Identifier.Value

	// Look up the old value from the context's old values stack
	oldValue, found := ctx.GetOldValue(identName)
	if !found {
		return e.newError(node, "old value for '%s' not captured (internal error)", identName)
	}

	// Return the old value (already a Value type)
	return oldValue.(Value)
}
