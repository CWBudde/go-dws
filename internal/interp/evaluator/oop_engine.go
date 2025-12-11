package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// OOPEngine handles runtime object-oriented programming operations.
// This interface encapsulates method dispatch, constructors, type operations,
// and operator overloading without exposing the full interpreter internals.
//
// Task 3.4.3: Extracted from monolithic InterpreterAdapter (67 methods)
// to create focused interface with single responsibility.
type OOPEngine interface {
	// ===== Method Dispatch (3 methods, high usage) =====

	// CallMethod dispatches a method call on an object.
	// Used by: visitor_expressions_methods.go, visitor_expressions_identifiers.go (2 uses)
	CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value

	// CallInheritedMethod calls a parent class method.
	// Used by: visitor_expressions_methods.go (1 use)
	CallInheritedMethod(obj Value, methodName string, args []Value) Value

	// ExecuteMethodWithSelf executes a method with explicit Self context.
	// Used by: index_assignment.go (4), property_write.go (2), property_read.go (2),
	//          visitor_expressions_methods.go (1), visitor_expressions_identifiers.go (1)
	// HIGH USAGE: 10 total uses - property setters/getters with Self context
	ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value

	// ===== Implicit Self Method Dispatch (1 method, moderate usage) =====

	// CallImplicitSelfMethod calls a method on the implicit Self object.
	// Used by: visitor_expressions_functions.go (2), visitor_expressions_identifiers.go (1)
	// MODERATE USAGE: 3 uses - implicit Self context in method calls
	CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// ===== Constructors (1 method) =====

	// ExecuteConstructor invokes a constructor on an object.
	// Used by: visitor_expressions_identifiers.go (1 use)
	ExecuteConstructor(obj Value, constructorName string, args []Value) error

	// ===== Function Pointers (3 methods) =====

	// CallFunctionPointer executes a function pointer with arguments.
	// Used by: visitor_expressions_identifiers.go, visitor_expressions_functions.go (2 uses)
	CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value

	// CallUserFunction executes a user-defined function.
	// Used by: visitor_expressions_functions.go (2 uses)
	CallUserFunction(fn *ast.FunctionDecl, args []Value) Value

	// ExecuteFunctionPointerCall executes a function pointer with metadata.
	// Used by: visitor_expressions_functions.go (2 uses)
	ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value

	// CreateBoundMethodPointer creates a bound method pointer (closure).
	// Used by: visitor_expressions_identifiers.go, visitor_expressions_methods.go (2 uses)
	CreateBoundMethodPointer(obj Value, methodDecl any) Value

	// ===== Type Operations (2 methods) =====

	// CreateTypeCastWrapper wraps an object for type casting.
	// Used by: visitor_expressions_identifiers.go (2 uses)
	CreateTypeCastWrapper(className string, obj Value) Value

	// WrapInSubrange wraps a value in a subrange type.
	// Used by: visitor_declarations.go (1 use)
	WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error)

	// WrapInInterface wraps a value in an interface type.
	// Used by: visitor_declarations.go (1 use)
	WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error)

	// ===== Dispatch Helpers (4 methods) =====

	// CallQualifiedOrConstructor handles qualified calls or constructor invocations.
	// Used by: visitor_expressions_functions.go (1 use)
	CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) Value

	// CallRecordStaticMethod calls a static method on a record type.
	// Used by: visitor_expressions_functions.go (1 use)
	CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// DispatchRecordStaticMethod dispatches to a record's static method.
	// Used by: visitor_expressions_functions.go (1 use)
	DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// ExecuteRecordPropertyRead reads a record property with indices.
	// Used by: property_read.go (2 uses)
	ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value

	// CallExternalFunction calls an external Go function.
	// Used by: visitor_expressions_functions.go (1 use)
	CallExternalFunction(funcName string, argExprs []ast.Expression, node ast.Node) Value

	// ===== Operator Overloading (2 methods) =====

	// TryBinaryOperator attempts to invoke a binary operator overload.
	// Used by: binary_ops.go (2 uses)
	// Returns (result, true) if operator found, or (nil, false) if not found.
	TryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool)

	// TryUnaryOperator attempts to invoke a unary operator overload.
	// Used by: visitor_expressions.go (1 use)
	// Returns (result, true) if operator found, or (nil, false) if not found.
	TryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool)

	// ===== Class Lookup (1 method) =====

	// LookupClassByName finds a class by name and returns it as ClassMetaValue.
	// Used by: visitor_expressions_identifiers.go (1 use)
	// For typed nil class variable access.
	LookupClassByName(name string) ClassMetaValue
}

// Total: 21 methods
// High usage (3+ calls): ExecuteMethodWithSelf (10), CallImplicitSelfMethod (3)
// Moderate usage (2 calls): CallMethod, CallFunctionPointer, CallUserFunction,
//                           ExecuteFunctionPointerCall, CreateBoundMethodPointer,
//                           CreateTypeCastWrapper, ExecuteRecordPropertyRead,
//                           TryBinaryOperator
// Low usage (1 call): 11 remaining methods
