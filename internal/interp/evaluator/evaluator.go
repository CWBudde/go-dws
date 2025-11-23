package evaluator

import (
	"fmt"
	"io"
	"math/rand"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Value represents a runtime value in the DWScript interpreter.
// This is temporarily defined here to avoid circular imports during the refactoring.
// In the final architecture, this will be properly organized.
type Value interface {
	Type() string
	String() string
}

// Config holds configuration options for the evaluator.
type Config struct {
	SourceCode        string
	SourceFile        string
	MaxRecursionDepth int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		MaxRecursionDepth: 1024, // Matches DWScript default
		SourceCode:        "",
		SourceFile:        "",
	}
}

// ExternalFunctionRegistry manages external (Go) functions that can be called from DWScript.
// This is temporarily defined here to avoid circular imports during the refactoring.
type ExternalFunctionRegistry interface {
	// Placeholder for now - will be properly defined later
}

// InterpreterAdapter is a temporary interface to allow the Evaluator to delegate
// back to the Interpreter during the migration process.
// Phase 3.5.1: This will be removed once all evaluation logic is moved to Evaluator.
// Phase 3.5.4 - Phase 2A: Extended to include function call methods.
type InterpreterAdapter interface {
	// EvalNode evaluates a node using the legacy Interpreter.Eval method.
	EvalNode(node ast.Node) Value

	// EvalNodeWithContext evaluates a node using the legacy Interpreter.Eval method
	// with proper environment synchronization from the ExecutionContext.
	// This ensures that scoped environments (from loops, functions, etc.) are respected.
	// Phase 3.5.44: Added to fix scope desync when adapter fallbacks are called from scoped contexts.
	EvalNodeWithContext(node ast.Node, ctx *ExecutionContext) Value

	// Phase 3.5.4 - Phase 2A: Function call system methods
	// These methods allow the Evaluator to call functions during evaluation
	// without directly accessing Interpreter fields.

	// CallFunctionPointer executes a function pointer with given arguments.
	// The funcPtr should be a FunctionPointerValue containing the function/lambda and closure.
	CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value

	// CallUserFunction executes a user-defined function.
	CallUserFunction(fn *ast.FunctionDecl, args []Value) Value

	// CallBuiltinFunction executes a built-in function by name.
	CallBuiltinFunction(name string, args []Value) Value

	// IsBuiltinFunction checks if a name refers to a built-in function.
	// This is used to avoid unnecessary function call attempts for undefined identifiers.
	IsBuiltinFunction(name string) bool

	// LookupFunction finds a function by name in the function registry.
	// Returns the function declaration(s) and a boolean indicating success.
	// Multiple functions may be returned for overloaded functions.
	LookupFunction(name string) ([]*ast.FunctionDecl, bool)

	// ===== Task 3.5.46: Specific Call Adapter Methods =====
	// These methods replace generic EvalNodeWithContext calls with specific,
	// well-named methods that describe what they do.

	// EvalCallExpression evaluates a call expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for complex call cases that require:
	// - Function pointer calls with closure restoration
	// - Method dispatch (object, record, interface, class)
	// - Overload resolution for user functions
	// - Lazy and var parameter handling
	// - Unit-qualified function calls
	// - Constructor calls
	// - Type casts
	// - Default() function
	// This replaces generic EvalNodeWithContext for CallExpression nodes.
	EvalCallExpression(node *ast.CallExpression, ctx *ExecutionContext) Value

	// EvalMemberAccessExpression evaluates a member access expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for complex member access cases that require:
	// - Unit-qualified access (UnitName.Symbol)
	// - Static class access (TClass.ClassVar, TClass.ClassName)
	// - Enum type access (TColor.Red, TColor.Low, TColor.High)
	// - Record type static access (TPoint.cOrigin)
	// - Record instance access (record.Field, record.Method)
	// - Object instance access (obj.Field, obj.Method, obj.Property)
	// - Interface instance access (interface.Method)
	// - Type cast value handling (TBase(child).ClassVar)
	// - Nil object handling (nil.ClassVar)
	// - Enum value properties (enumVal.Value)
	// - Class/metaclass access (ClassValue.Member)
	// This replaces generic EvalNodeWithContext for MemberAccessExpression nodes.
	EvalMemberAccessExpression(node *ast.MemberAccessExpression, ctx *ExecutionContext) Value

	// EvalMethodCallExpression evaluates a method call expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for method call cases including helper methods.
	// This replaces generic EvalNodeWithContext for MethodCallExpression default cases.
	EvalMethodCallExpression(node *ast.MethodCallExpression, ctx *ExecutionContext) Value

	// EvalSetLiteral evaluates a set literal expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for set literal evaluation.
	// This replaces generic EvalNodeWithContext for SetLiteral nodes.
	EvalSetLiteral(node *ast.SetLiteral, ctx *ExecutionContext) Value

	// EvalArrayLiteral evaluates an array literal expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for array literal evaluation with type inference.
	// This replaces generic EvalNodeWithContext for ArrayLiteralExpression nodes.
	EvalArrayLiteral(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value

	// EvalNewArrayExpression evaluates a new array expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for array construction.
	// This replaces generic EvalNodeWithContext for NewArrayExpression nodes.
	EvalNewArrayExpression(node *ast.NewArrayExpression, ctx *ExecutionContext) Value

	// EvalIndexExpression evaluates an index expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for array/string indexing.
	// This replaces generic EvalNodeWithContext for IndexExpression nodes.
	EvalIndexExpression(node *ast.IndexExpression, ctx *ExecutionContext) Value

	// EvalRangeExpression evaluates a range expression using the interpreter's full logic.
	// Task 3.5.46: Specific adapter for range evaluation.
	// This replaces generic EvalNodeWithContext for RangeExpression nodes.
	EvalRangeExpression(node *ast.RangeExpression, ctx *ExecutionContext) Value

	// ===== Task 3.5.47: Statement and Binary Op Adapter Methods =====

	// EvalAssignment evaluates an assignment statement using the interpreter's full logic.
	// Task 3.5.47: Specific adapter for assignment statement evaluation that handles:
	// - Simple assignment: x := value
	// - Member assignment: obj.field := value, record.field := value
	// - Index assignment: arr[i] := value, obj.Property[x, y] := value
	// - Compound operators: +=, -=, *=, /= with type coercion and operator overloads
	// - ReferenceValue (var parameters), external variables, subrange validation
	// - Implicit type conversions, variant boxing, object reference counting
	// - Property setter dispatch with recursion prevention
	// This replaces generic EvalNodeWithContext for AssignmentStatement nodes.
	EvalAssignment(node *ast.AssignmentStatement, ctx *ExecutionContext) Value

	// EvalSetBinaryOperation evaluates a binary operation on set values using the interpreter's full logic.
	// Task 3.5.47: Specific adapter for set binary operations that handles:
	// - Union (+), difference (-), intersection (*) operations
	// - SetValue type and storage backends (bitmask vs map)
	// - SetType information for type checking
	// - Calls interpreter's evalBinarySetOperation method
	// This replaces generic EvalNodeWithContext for set binary operations.
	EvalSetBinaryOperation(op string, left, right Value, node ast.Node, ctx *ExecutionContext) Value

	// EvalVariantUnaryNot evaluates the NOT operator on a Variant value using the interpreter's full logic.
	// Task 3.5.47: Specific adapter for Variant NOT operations that handles:
	// - Variant unwrapping
	// - Underlying type determination
	// - Applying NOT to unwrapped value (boolean or bitwise)
	// - Special handling for nil/unassigned variants
	// This replaces generic EvalNodeWithContext for Variant NOT operations.
	EvalVariantUnaryNot(operand Value, node ast.Node, ctx *ExecutionContext) Value

	// ===== Task 3.5.50: Declaration Adapter Methods =====

	// EvalClassDeclaration evaluates a class declaration using the interpreter's full logic.
	// Task 3.5.50: Specific adapter for class declaration that handles:
	// - ClassInfo creation and initialization
	// - Inheritance (parent class lookup, field/method copying, VMT building)
	// - Interface implementation tracking
	// - Class constants, fields (instance and class vars), properties
	// - Methods (instance, class, virtual, override), constructors, destructors
	// - Operator overload registration
	// - Type ID assignment and metadata creation
	// This replaces generic EvalNodeWithContext for ClassDecl nodes.
	EvalClassDeclaration(node *ast.ClassDecl, ctx *ExecutionContext) Value

	// EvalInterfaceDeclaration evaluates an interface declaration using the interpreter's full logic.
	// Task 3.5.50: Specific adapter for interface declaration that handles:
	// - InterfaceInfo creation and initialization
	// - Parent interface inheritance (method and property inheritance)
	// - Method signature registration (InterfaceMethodDecl to FunctionDecl conversion)
	// - Property declaration registration
	// - Interface registration in both TypeSystem and legacy map
	// This replaces generic EvalNodeWithContext for InterfaceDecl nodes.
	EvalInterfaceDeclaration(node *ast.InterfaceDecl, ctx *ExecutionContext) Value

	// EvalHelperDeclaration evaluates a helper declaration using the interpreter's full logic.
	// Task 3.5.50: Specific adapter for helper declaration that handles:
	// - HelperInfo creation and initialization
	// - Target type resolution from AST type annotation
	// - Parent helper inheritance (method and property copying)
	// - Method registration (user-defined and builtin)
	// - Property registration (getter/setter handling)
	// - Class variable and class constant initialization
	// - Helper registration in TypeSystem by type name (normalized and simple)
	// This replaces generic EvalNodeWithContext for HelperDecl nodes.
	EvalHelperDeclaration(node *ast.HelperDecl, ctx *ExecutionContext) Value

	// ===== Task 3.5.51: Function/Operator Declaration Adapter Methods =====

	// EvalFunctionDeclaration evaluates a function declaration using the interpreter's full logic.
	// Task 3.5.51: Specific adapter for function declaration that handles:
	// - Function registration in i.functions map (case-insensitive)
	// - Method implementation (ClassName != nil): updates ClassInfo or RecordTypeValue methods
	// - Support for function overloading (multiple functions per name)
	// - Replacement of interface declarations (no body) with implementations (has body)
	// - Constructor and destructor registration for class methods
	// - VMT rebuilding after method implementation
	// This replaces generic EvalNodeWithContext for FunctionDecl nodes.
	EvalFunctionDeclaration(node *ast.FunctionDecl, ctx *ExecutionContext) Value

	// EvalOperatorDeclaration evaluates an operator declaration using the interpreter's full logic.
	// Task 3.5.51: Specific adapter for operator declaration that handles:
	// - Class operators: skipped (handled during class declaration)
	// - Conversion operators (implicit/explicit): registers in i.conversions
	// - Global operators: registers in i.globalOperators with operand types
	// - Binding function name normalization (case-insensitive)
	// This replaces generic EvalNodeWithContext for OperatorDecl nodes.
	EvalOperatorDeclaration(node *ast.OperatorDecl, ctx *ExecutionContext) Value

	// Phase 3.5.4 - Phase 2B: Type system access methods
	// These methods allow the Evaluator to access type registries during evaluation
	// without directly accessing Interpreter fields.

	// ===== Class Registry =====

	// LookupClass finds a class by name in the class registry.
	// Returns the class info (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupClass(name string) (any, bool)

	// HasClass checks if a class with the given name exists.
	HasClass(name string) bool

	// GetClassTypeID returns the type ID for a class, or 0 if not found.
	GetClassTypeID(className string) int

	// ===== Record Registry =====

	// LookupRecord finds a record type by name in the record registry.
	// Returns the record type value (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupRecord(name string) (any, bool)

	// HasRecord checks if a record type with the given name exists.
	HasRecord(name string) bool

	// GetRecordTypeID returns the type ID for a record type, or 0 if not found.
	GetRecordTypeID(recordName string) int

	// ===== Interface Registry =====

	// LookupInterface finds an interface by name in the interface registry.
	// Returns the interface info (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupInterface(name string) (any, bool)

	// HasInterface checks if an interface with the given name exists.
	HasInterface(name string) bool

	// ===== Helper Registry =====

	// LookupHelpers finds helper methods for a type by name.
	// Returns a slice of helper info (each element as any/interface{}).
	// The lookup is case-insensitive.
	LookupHelpers(typeName string) []any

	// HasHelpers checks if a type has helper methods defined.
	HasHelpers(typeName string) bool

	// ===== Operator & Conversion Registries =====

	// GetOperatorRegistry returns the operator registry for operator overload lookups.
	// Returns the registry as any/interface{} to avoid circular dependencies.
	GetOperatorRegistry() any

	// GetConversionRegistry returns the conversion registry for type conversion lookups.
	// Returns the registry as any/interface{} to avoid circular dependencies.
	GetConversionRegistry() any

	// ===== Enum Type IDs =====

	// GetEnumTypeID returns the type ID for an enum type, or 0 if not found.
	GetEnumTypeID(enumName string) int

	// GetType resolves a type by name.
	// Returns the resolved type and an error if the type is not found.
	// The lookup is case-insensitive.
	GetType(name string) (any, error)

	// ResolveType resolves a type from an AST type annotation.
	// Returns the resolved type and an error if the type cannot be resolved.
	ResolveType(typeAnnotation *ast.TypeAnnotation) (any, error)

	// IsTypeCompatible checks if a value is compatible with a target type.
	// This is used for type checking in assignments and parameter passing.
	IsTypeCompatible(from Value, toTypeName string) bool

	// InferArrayElementType infers the element type from array literal elements.
	// Returns the inferred type or an error if elements have incompatible types.
	InferArrayElementType(elements []Value) (any, error)

	// InferRecordType infers the record type name from field values.
	// Returns the record type name or an error if it cannot be inferred.
	InferRecordType(fields map[string]Value) (string, error)

	// ConvertValue performs implicit or explicit type conversion.
	// Returns the converted value or an error if conversion is not possible.
	ConvertValue(value Value, targetTypeName string) (Value, error)

	// CreateDefaultValue creates a zero/default value for a given type name.
	// Returns the default value or nil if the type is not recognized.
	CreateDefaultValue(typeName string) Value

	// IsEnumType checks if a given name refers to an enum type.
	// The lookup is case-insensitive.
	IsEnumType(typeName string) bool

	// IsRecordType checks if a given name refers to a record type.
	// The lookup is case-insensitive.
	IsRecordType(typeName string) bool

	// IsArrayType checks if a given name refers to an array type.
	// The lookup is case-insensitive.
	IsArrayType(typeName string) bool

	// ===== Task 3.5.6: Array and Collection Adapter Methods =====

	// CreateArray creates an array from a list of elements with a specified element type.
	// Returns the created array value.
	CreateArray(elementType any, elements []Value) Value

	// CreateDynamicArray allocates a new dynamic array of a given size and element type.
	// Returns the created array value.
	CreateDynamicArray(elementType any, size int) Value

	// CreateArrayWithExpectedType creates an array from elements with type-aware construction.
	// Uses the expected array type for proper element type inference and coercion.
	CreateArrayWithExpectedType(elements []Value, expectedType any) Value

	// GetArrayElement retrieves an element from an array at the given index.
	// Performs bounds checking and returns an error if index is out of range.
	GetArrayElement(array Value, index Value) (Value, error)

	// SetArrayElement sets an element in an array at the given index.
	// Performs bounds checking and returns an error if index is out of range.
	SetArrayElement(array Value, index Value, value Value) error

	// GetArrayLength returns the length of an array.
	// Returns 0 for non-array values.
	GetArrayLength(array Value) int

	// CreateSet creates a set from a list of elements with a specified element type.
	// Returns the created set value.
	CreateSet(elementType any, elements []Value) Value

	// EvaluateSetRange expands a range expression (e.g., 1..10, 'a'..'z') into ordinal values.
	// Returns a slice of ordinal values or an error if the range cannot be evaluated.
	EvaluateSetRange(start Value, end Value) ([]int, error)

	// AddToSet adds an element to a set.
	// Returns an error if the element cannot be added.
	AddToSet(set Value, element Value) error

	// GetStringChar retrieves a character from a string at the given index (1-based).
	// Returns an error if index is out of range.
	GetStringChar(str Value, index Value) (Value, error)

	// ===== Task 3.5.7: Property, Field, and Member Access Adapter Methods =====

	// ===== Field Access =====

	// GetObjectField retrieves a field value from an object.
	// Returns the field value and an error if the field does not exist.
	GetObjectField(obj Value, fieldName string) (Value, error)

	// SetObjectField sets a field value in an object.
	// Returns an error if the field does not exist or the value is incompatible.
	SetObjectField(obj Value, fieldName string, value Value) error

	// GetRecordField retrieves a field value from a record.
	// Returns the field value and an error if the field does not exist.
	GetRecordField(record Value, fieldName string) (Value, error)

	// SetRecordField sets a field value in a record.
	// Returns an error if the field does not exist or the value is incompatible.
	SetRecordField(record Value, fieldName string, value Value) error

	// ===== Property Access =====

	// GetPropertyValue retrieves a property value from an object.
	// Returns the property value and an error if the property does not exist.
	GetPropertyValue(obj Value, propName string) (Value, error)

	// SetPropertyValue sets a property value in an object.
	// Returns an error if the property does not exist or the value is incompatible.
	SetPropertyValue(obj Value, propName string, value Value) error

	// GetIndexedProperty retrieves an indexed property value from an object.
	// Returns the property value and an error if the property does not exist or indices are invalid.
	GetIndexedProperty(obj Value, propName string, indices []Value) (Value, error)

	// SetIndexedProperty sets an indexed property value in an object.
	// Returns an error if the property does not exist, indices are invalid, or value is incompatible.
	SetIndexedProperty(obj Value, propName string, indices []Value, value Value) error

	// ===== Method Calls =====

	// CallMethod executes a method on an object with the given arguments.
	// Returns the method result value.
	CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value

	// CallInheritedMethod executes an inherited (parent) method with the given arguments.
	// Returns the method result value.
	CallInheritedMethod(obj Value, methodName string, args []Value) Value

	// ===== Object Operations =====

	// CreateObject creates a new object instance of the specified class with constructor arguments.
	// Returns the created object value and an error if the class does not exist or construction fails.
	CreateObject(className string, args []Value) (Value, error)

	// CheckType checks if an object is of a specified type (implements 'is' operator).
	// Returns true if the object is compatible with the specified type name.
	CheckType(obj Value, typeName string) bool

	// CastType performs type casting (implements 'as' operator).
	// Returns the casted value and an error if the cast fails.
	CastType(obj Value, typeName string) (Value, error)

	// CheckImplements checks if an object/class implements an interface (implements 'implements' operator).
	// Task 3.5.36: Supports ObjectInstance, ClassValue, and ClassInfoValue inputs.
	// Returns true if the class implements the specified interface.
	CheckImplements(obj Value, interfaceName string) (bool, error)

	// ===== Function Pointers (Task 3.5.8) =====

	// CreateFunctionPointer creates a function pointer value from a function declaration.
	// The closure parameter is the environment where the function pointer is created.
	// Returns the function pointer value.
	CreateFunctionPointer(fn *ast.FunctionDecl, closure any) Value

	// CreateLambda creates a lambda/closure value from a lambda expression.
	// The closure parameter is the environment where the lambda is created.
	// Returns the lambda value.
	CreateLambda(lambda *ast.LambdaExpression, closure any) Value

	// IsFunctionPointer checks if a value is a function pointer.
	IsFunctionPointer(value Value) bool

	// GetFunctionPointerParamCount returns the number of parameters a function pointer expects.
	// Returns 0 for non-function-pointer values.
	GetFunctionPointerParamCount(funcPtr Value) int

	// IsFunctionPointerNil checks if a function pointer is nil (unassigned).
	// Returns true if the function pointer has no function or lambda assigned.
	IsFunctionPointerNil(funcPtr Value) bool

	// ===== Method Pointers (Task 3.5.37) =====

	// CreateMethodPointer creates a method pointer value bound to a specific object.
	// Task 3.5.37: Used by address-of expression (@object.MethodName) to create
	// method pointers that capture both the method and the object to call it on.
	// Parameters:
	//   - obj: The object instance (Value) to bind the method to
	//   - methodName: The name of the method to look up
	//   - closure: The environment where the method pointer is created
	// Returns the method pointer value and an error if the method is not found.
	CreateMethodPointer(obj Value, methodName string, closure any) (Value, error)

	// CreateFunctionPointerFromName creates a function pointer for a named function.
	// Task 3.5.37: Used by address-of expression (@FunctionName) to create
	// function pointers from standalone functions.
	// Parameters:
	//   - funcName: The name of the function to look up (case-insensitive)
	//   - closure: The environment where the function pointer is created
	// Returns the function pointer value and an error if the function is not found.
	CreateFunctionPointerFromName(funcName string, closure any) (Value, error)

	// ===== Record Operations (Task 3.5.7) =====

	// CreateRecord creates a record value from field values.
	// Returns the record value and an error if the record type doesn't exist or fields are invalid.
	CreateRecord(recordType string, fields map[string]Value) (Value, error)

	// ===== Assignment Helpers (Task 3.5.7) =====

	// SetVariable assigns a value to a variable in the execution context.
	// Returns an error if the assignment fails.
	SetVariable(name string, value Value, ctx *ExecutionContext) error

	// CanAssign checks if an AST node can be used as an lvalue (assignment target).
	// Returns true if the node is a valid lvalue.
	CanAssign(target ast.Node) bool

	// ===== Exception Handling (Task 3.5.8) =====

	// RaiseException raises an exception with the given class name and message.
	// The pos parameter provides source location information for error reporting.
	RaiseException(className string, message string, pos any)

	// ===== Environment Access (Task 3.5.9) =====

	// GetVariable retrieves a variable value from the execution context.
	// Returns the value and true if found, nil and false otherwise.
	GetVariable(name string, ctx *ExecutionContext) (Value, bool)

	// DefineVariable defines a new variable in the execution context.
	// This creates a new binding in the current scope.
	DefineVariable(name string, value Value, ctx *ExecutionContext)

	// CreateEnclosedEnvironment creates a new execution context with an enclosed environment.
	// The new environment has the current environment as its parent (for scoping).
	// Returns a new ExecutionContext with the enclosed environment.
	CreateEnclosedEnvironment(ctx *ExecutionContext) *ExecutionContext

	// PropertyEvalContext: Available via ExecutionContext.PropContext() for recursion prevention
	// Property dispatch: Available via EvalNode delegation (uses Phase 2A function calls + Phase 2B type lookups)
	// Array indexing: Available via EvalNode delegation (bounds checking integrated)
	// Record operations: Available via Phase 2B record registry + EvalNode delegation
	// Helper operations: Available via Phase 2B helper registry + EvalNode delegation
	//
	// These complex operations compose existing infrastructure (Phase 2A + Phase 2B + ExecutionContext)
	// and are properly handled through EvalNode delegation. No additional adapter methods needed.

	// ===== Task 3.5.19: Binary Operator Adapter Methods (Fix for PR #219) =====
	//
	// These methods delegate binary operator evaluation to the Interpreter WITHOUT re-evaluating operands.
	// This fixes the double-evaluation bug where operands were evaluated once in the Evaluator,
	// then re-evaluated again when calling adapter.EvalNode(node).

	// EvalVariantBinaryOp handles binary operations with Variant operands using pre-evaluated values.
	// This prevents double-evaluation of operands with side effects.
	EvalVariantBinaryOp(op string, left, right Value, node ast.Node) Value

	// EvalInOperator evaluates the 'in' operator for membership testing using pre-evaluated values.
	// This prevents double-evaluation of operands with side effects.
	EvalInOperator(value, container Value, node ast.Node) Value

	// EvalEqualityComparison handles = and <> operators for complex types using pre-evaluated values.
	// This prevents double-evaluation of operands with side effects.
	EvalEqualityComparison(op string, left, right Value, node ast.Node) Value

	// ===== Task 3.5.38: Variable Declaration Adapter Methods =====

	// ParseInlineArrayType parses inline array type signatures like "array of Integer" or "array[1..10] of String".
	// Returns the array type (as any/interface{}) and an error if parsing fails.
	ParseInlineArrayType(typeName string) (any, error)

	// ParseInlineSetType parses inline set type signatures like "set of TColor".
	// Returns the set type (as any/interface{}) and an error if parsing fails.
	ParseInlineSetType(typeName string) (any, error)

	// LookupSubrangeType finds a subrange type by name in the subrange type registry.
	// Returns the subrange type value (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupSubrangeType(name string) (any, bool)

	// BoxVariant wraps a value in a Variant container.
	// Returns the variant value containing the wrapped value.
	BoxVariant(value Value) Value

	// TryImplicitConversion attempts an implicit type conversion from value to targetTypeName.
	// Returns the converted value and true if conversion succeeded, or original value and false if not.
	TryImplicitConversion(value Value, targetTypeName string) (Value, bool)

	// WrapInSubrange wraps an integer value in a subrange type with validation.
	// Returns the subrange value and an error if validation fails.
	WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error)

	// WrapInInterface wraps an object value in an interface instance.
	// Validates that the object's class implements the interface.
	// Returns the interface instance and an error if validation fails.
	WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error)

	// EvalArrayLiteralWithExpectedType evaluates an array literal with expected type context.
	// This allows proper element type inference and coercion.
	// Returns the evaluated array value.
	EvalArrayLiteralWithExpectedType(lit ast.Node, expectedTypeName string) Value

	// ClassImplementsInterface checks if a class implements an interface.
	// Returns true if the class (by name) implements the interface (by name).
	// Both lookups are case-insensitive.
	ClassImplementsInterface(className, interfaceName string) bool

	// CreateExternalVar creates an external variable marker.
	// External variables are placeholders that map to Go-side external functions.
	// Returns the external variable value.
	CreateExternalVar(varName, externalName string) Value

	// ResolveArrayTypeNode resolves an array type from an AST ArrayTypeNode.
	// This handles nested arrays and complex bound expressions.
	// Returns the array type (as any/interface{}) and an error if resolution fails.
	ResolveArrayTypeNode(arrayNode ast.Node) (any, error)

	// CreateRecordZeroValue creates a zero-initialized record value for a given record type.
	// All fields are initialized to their respective zero values.
	// Returns the record value and an error if the record type doesn't exist.
	CreateRecordZeroValue(recordTypeName string) (Value, error)

	// CreateArrayZeroValue creates a zero-initialized array value for a given array type name.
	// Returns the array value and an error if the array type doesn't exist.
	CreateArrayZeroValue(arrayTypeName string) (Value, error)

	// CreateSetZeroValue creates an empty set value for a given set type signature.
	// Returns the set value and an error if the set type cannot be created.
	CreateSetZeroValue(setTypeName string) (Value, error)

	// CreateSubrangeZeroValue creates a zero-initialized subrange value for a given subrange type.
	// Returns the subrange value and an error if the subrange type doesn't exist.
	CreateSubrangeZeroValue(subrangeTypeName string) (Value, error)

	// CreateInterfaceZeroValue creates a nil interface instance for a given interface type.
	// Returns the interface instance and an error if the interface type doesn't exist.
	CreateInterfaceZeroValue(interfaceName string) (Value, error)

	// CreateClassZeroValue creates a typed nil value for a given class type.
	// This allows accessing class variables via nil instances.
	// Returns the typed nil value and an error if the class doesn't exist.
	CreateClassZeroValue(className string) (Value, error)

	// ===== Task 3.5.40: Record Literal Adapter Methods =====

	// CreateRecordValue creates a record value with field initialization.
	// fieldValues is a map of field names to evaluated values.
	// Missing fields are initialized with default values or field initializers.
	// Returns the record value and an error if the record type doesn't exist or fields are invalid.
	CreateRecordValue(recordTypeName string, fieldValues map[string]Value) (Value, error)

	// GetRecordFieldDeclarations retrieves field declarations for a record type.
	// Returns field declarations map and a boolean indicating success.
	GetRecordFieldDeclarations(recordTypeName string) (any, bool)

	// GetZeroValueForType creates a zero/default value for a given type.
	// Handles nested records, arrays, and all DWScript types.
	// Returns the zero value for the type.
	GetZeroValueForType(typeInfo any) Value

	// InitializeInterfaceField creates a nil interface instance for interface-typed fields.
	// Returns an InterfaceInstance value or nil if the type is not an interface.
	InitializeInterfaceField(fieldType any) Value

	// ===== Task 3.5.21: Complex Value Retrieval Adapter Methods =====

	// IsExternalVar checks if a value is an ExternalVarValue.
	// Returns true if the value is an external variable marker.
	IsExternalVar(value Value) bool

	// IsLazyThunk checks if a value is a LazyThunk.
	// Returns true if the value is a lazy parameter that needs evaluation.
	IsLazyThunk(value Value) bool

	// IsReferenceValue checks if a value is a ReferenceValue.
	// Returns true if the value is a var parameter reference that needs dereferencing.
	IsReferenceValue(value Value) bool

	// EvaluateLazyThunk forces evaluation of a lazy parameter.
	// Returns the evaluated value.
	// Panics if the value is not a LazyThunk (check with IsLazyThunk first).
	EvaluateLazyThunk(value Value) Value

	// DereferenceValue dereferences a var parameter reference.
	// Returns the actual value and an error if dereferencing fails.
	// Panics if the value is not a ReferenceValue (check with IsReferenceValue first).
	DereferenceValue(value Value) (Value, error)

	// GetExternalVarName returns the name of an external variable.
	// Returns the external variable name.
	// Panics if the value is not an ExternalVarValue (check with IsExternalVar first).
	GetExternalVarName(value Value) string

	// CreateLazyThunk creates a lazy parameter thunk from an unevaluated expression.
	// Lazy parameters are re-evaluated each time they are accessed (Jensen's Device pattern).
	// Parameters:
	//   - expr: The AST expression to evaluate lazily
	//   - env: The environment captured from the call site
	// Returns the lazy thunk value.
	// Panics if the env parameter is not of type *Environment.
	CreateLazyThunk(expr ast.Expression, env any) Value

	// CreateReferenceValue creates a var parameter reference.
	// Var parameters allow pass-by-reference semantics.
	// Parameters:
	//   - varName: The name of the variable to reference
	//   - env: The environment containing the variable
	// Returns the reference value.
	// Panics if the env parameter is not of type *Environment.
	CreateReferenceValue(varName string, env any) Value

	// ===== Task 3.5.22: Property & Method Reference Adapter Methods =====

	// IsObjectInstance checks if a value is an ObjectInstance.
	// Returns true if the value is an object instance.
	IsObjectInstance(value Value) bool

	// GetObjectFieldValue retrieves a field value from an object instance.
	// Returns the field value and true if found, nil and false otherwise.
	GetObjectFieldValue(obj Value, fieldName string) (Value, bool)

	// GetClassVariableValue retrieves a class variable value from an object's class.
	// Returns the class variable value and true if found, nil and false otherwise.
	GetClassVariableValue(obj Value, varName string) (Value, bool)

	// HasProperty checks if an object has a property with the given name.
	// Returns true if the property exists (case-insensitive lookup).
	HasProperty(obj Value, propName string) bool

	// ReadPropertyValue reads a property value from an object.
	// Handles field-backed, method-backed, and expression-backed properties.
	// Returns the property value and an error if reading fails.
	// Note: Caller is responsible for property recursion prevention.
	ReadPropertyValue(obj Value, propName string, node any) (Value, error)

	// HasMethod checks if an object has a method with the given name.
	// Returns true if the method exists (case-insensitive lookup).
	HasMethod(obj Value, methodName string) bool

	// IsMethodParameterless checks if a method has zero parameters.
	// Returns true if the method exists and has no parameters.
	// Returns false if the method doesn't exist or has parameters.
	IsMethodParameterless(obj Value, methodName string) bool

	// CreateMethodCall creates a synthetic method call expression for auto-invocation.
	// Used when a parameterless method is referenced without parentheses.
	// Returns the result of calling the method.
	CreateMethodCall(obj Value, methodName string, node any) Value

	// CreateMethodPointer creates a method pointer for a method with parameters.
	// Used when a method with parameters is referenced without parentheses.
	// Returns a FunctionPointerValue bound to the object.
	CreateMethodPointerFromObject(obj Value, methodName string) (Value, error)

	// GetClassName returns the class name for an object instance.
	// Returns the class name string.
	GetClassName(obj Value) string

	// GetClassType returns the ClassValue (metaclass) for an object instance.
	// Returns the ClassValue representing the object's runtime class.
	GetClassType(obj Value) Value

	// IsClassInfoValue checks if a value is a ClassInfoValue.
	// Returns true if the value represents class metadata (from __CurrentClass__).
	IsClassInfoValue(value Value) bool

	// GetClassNameFromClassInfo returns the class name from a ClassInfoValue.
	// Returns the class name string.
	// Panics if the value is not a ClassInfoValue.
	GetClassNameFromClassInfo(classInfo Value) string

	// GetClassTypeFromClassInfo returns the ClassValue from a ClassInfoValue.
	// Returns the ClassValue (metaclass reference).
	// Panics if the value is not a ClassInfoValue.
	GetClassTypeFromClassInfo(classInfo Value) Value

	// GetClassVariableFromClassInfo retrieves a class variable from ClassInfoValue.
	// Returns the class variable value and true if found, nil and false otherwise.
	// Panics if the value is not a ClassInfoValue.
	GetClassVariableFromClassInfo(classInfo Value, varName string) (Value, bool)

	// IsClassValue checks if a value is a ClassValue (metaclass reference).
	// Returns true if the value is a class name used as a value (e.g., var c := TMyClass).
	IsClassValue(value Value) bool

	// CreateClassValueFromName creates a ClassValue from a class name.
	// Task 3.5.46: Used by VisitIdentifier when an identifier resolves to a class name.
	// Returns the ClassValue (metaclass reference) and true if the class exists,
	// nil and false otherwise.
	CreateClassValueFromName(className string) (Value, bool)

	// ===== Task 3.5.29: Exception Handling Adapter Methods =====

	// MatchesExceptionType checks if an exception matches a handler's type.
	// The exception should be an *ExceptionValue. The typeExpr is the exception type
	// from the handler (e.g., "Exception", "EDivByZero").
	// Returns true if the exception type matches or inherits from the handler type.
	// Returns true if typeExpr is nil (bare handler catches all).
	MatchesExceptionType(exc interface{}, typeExpr ast.TypeExpression) bool

	// GetExceptionInstance returns the ObjectInstance from an exception.
	// The exception should be an *ExceptionValue.
	// Returns the instance Value or nil if not an exception.
	GetExceptionInstance(exc interface{}) Value

	// CreateExceptionFromObject creates an ExceptionValue from an object instance.
	// Used by raise statement to create exception from user-provided object.
	// Parameters:
	//   - obj: The object instance (Value) that should be an exception object
	//   - ctx: The execution context for call stack capture
	//   - pos: Position information for error reporting (token.Position or similar)
	// Returns the exception value (interface{}) to be set in context.
	CreateExceptionFromObject(obj Value, ctx *ExecutionContext, pos any) interface{}

	// SyncException synchronizes exception state from interpreter to context.
	// This must be called after operations that may raise exceptions
	// (e.g., CallBuiltinFunction, CallUserFunction) because the interpreter
	// stores exceptions on its own field which doesn't automatically propagate
	// to the ExecutionContext.
	SyncException(ctx *ExecutionContext)

	// EvalBlockStatement evaluates a block statement in the given context.
	// Returns nil after evaluating all statements.
	EvalBlockStatement(block *ast.BlockStatement, ctx *ExecutionContext)

	// EvalStatement evaluates a single statement in the given context.
	// Used by exception handlers to evaluate the handler statement.
	EvalStatement(stmt ast.Statement, ctx *ExecutionContext)

	// ===== Task 3.5.49: Record and Type Declaration Adapter Methods =====

	// BuildRecordTypeValue creates a RecordTypeValue from record declaration components.
	// This encapsulates the creation of interp-specific types that the evaluator cannot create directly.
	// Parameters:
	//   - recordName: The name of the record type
	//   - recordType: The types.RecordType with field definitions
	//   - fieldDecls: Map of field names to field declarations (for initializers)
	//   - methods: Map of method names to method declarations
	//   - staticMethods: Map of static method names to method declarations
	//   - constants: Map of constant names to evaluated values
	//   - classVars: Map of class variable names to evaluated values
	// Returns the RecordTypeValue (as any) for registration.
	BuildRecordTypeValue(
		recordName string,
		recordType any,
		fieldDecls map[string]*ast.FieldDecl,
		methods map[string]*ast.FunctionDecl,
		staticMethods map[string]*ast.FunctionDecl,
		constants map[string]Value,
		classVars map[string]Value,
	) any

	// RegisterRecordTypeInEnvironment registers a record type in the environment.
	// This handles the special key naming convention and stores the RecordTypeValue.
	// Parameters:
	//   - recordName: The name of the record type
	//   - recordTypeValue: The RecordTypeValue (as any) to register
	//   - ctx: The execution context with the environment
	RegisterRecordTypeInEnvironment(recordName string, recordTypeValue any, ctx *ExecutionContext)

	// RegisterArrayTypeInEnvironment registers an array type in the interpreter's environment.
	// Task 3.5.48: This stores in i.env directly to ensure IsArrayType and CreateArrayZeroValue
	// can find the type during variable declarations.
	// Parameters:
	//   - arrayName: The name of the array type
	//   - arrayTypeValue: The ArrayTypeValue (as Value) to register
	RegisterArrayTypeInEnvironment(arrayName string, arrayTypeValue Value)

	// BuildTypeAliasValue creates a TypeAliasValue from type alias components.
	// Parameters:
	//   - aliasName: The name of the type alias
	//   - aliasedType: The underlying type being aliased
	// Returns the TypeAliasValue (as any) for registration.
	BuildTypeAliasValue(aliasName string, aliasedType any) any

	// RegisterTypeAliasInEnvironment registers a type alias in the environment.
	// Parameters:
	//   - aliasName: The name of the type alias
	//   - typeAliasValue: The TypeAliasValue (as any) to register
	//   - ctx: The execution context with the environment
	RegisterTypeAliasInEnvironment(aliasName string, typeAliasValue any, ctx *ExecutionContext)

	// BuildSubrangeTypeValue creates a SubrangeTypeValue from subrange components.
	// Parameters:
	//   - typeName: The name of the subrange type
	//   - lowBound: The low bound value (IntegerValue)
	//   - highBound: The high bound value (IntegerValue)
	// Returns the SubrangeTypeValue (as any) for registration, or an error.
	BuildSubrangeTypeValue(typeName string, lowBound, highBound Value) (any, error)

	// RegisterSubrangeTypeInEnvironment registers a subrange type in the environment.
	// Parameters:
	//   - typeName: The name of the subrange type
	//   - subrangeTypeValue: The SubrangeTypeValue (as any) to register
	//   - ctx: The execution context with the environment
	RegisterSubrangeTypeInEnvironment(typeName string, subrangeTypeValue any, ctx *ExecutionContext)

	// ResolveTypeFromExpression resolves a type from an AST type expression.
	// This handles type names, inline types, and complex type expressions.
	// Returns the resolved types.Type (as any) or nil if resolution fails.
	ResolveTypeFromExpression(typeExpr ast.TypeExpression) any

	// GetValueType returns the types.Type for a runtime value.
	// Used for type inference from initializer values.
	// Returns the type (as any) or nil if the type cannot be determined.
	GetValueType(value Value) any

	// ===== Task 3.5.52: Call Expression Adapter Methods =====
	// These methods handle complex call scenarios that require interpreter state.

	// CallFunctionPointerWithArgs calls a function pointer with unevaluated arguments.
	// Handles lazy params, var params, and closure environment restoration.
	CallFunctionPointerWithArgs(funcPtr Value, args []ast.Expression, ctx *ExecutionContext) Value

	// CallRecordMethod calls a method on a record value.
	CallRecordMethod(record Value, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallInterfaceMethod calls a method on an interface value.
	CallInterfaceMethod(iface Value, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallObjectMethod calls a method on an object value.
	CallObjectMethod(obj Value, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallUnitFunction calls a unit-qualified function (UnitName.FunctionName).
	CallUnitFunction(unitName, funcName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallClassMethod calls a class method or constructor (TClass.MethodName).
	CallClassMethod(className, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallUserFunctionWithOverloads calls a user-defined function with potential overloads.
	CallUserFunctionWithOverloads(funcName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallImplicitSelfMethod calls a method using implicit Self reference.
	CallImplicitSelfMethod(selfVal Value, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallRecordStaticMethod calls a record's static method.
	CallRecordStaticMethod(recordVal Value, methodName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallBuiltinWithVarParam calls a builtin function that modifies its arguments.
	CallBuiltinWithVarParam(funcName string, args []ast.Expression, ctx *ExecutionContext) Value

	// CallExternalFunction calls an external (Go) function.
	CallExternalFunction(funcName string, args []ast.Expression, ctx *ExecutionContext) Value

	// EvalDefaultFunction evaluates a Default(TypeName) call.
	EvalDefaultFunction(arg ast.Expression, ctx *ExecutionContext) Value

	// EvalTypeCast evaluates a type cast expression (TypeName(expression)).
	EvalTypeCast(typeName string, arg ast.Expression, ctx *ExecutionContext) Value

	// HasExternalFunction checks if an external function with the given name exists.
	HasExternalFunction(funcName string) bool

	// ===== Task 3.5.53: Member Access Adapter Methods =====

	// EvalObjectMemberAccess evaluates member access on an object (methods, properties).
	EvalObjectMemberAccess(node *ast.MemberAccessExpression, obj Value, memberName string, ctx *ExecutionContext) Value

	// EvalInterfaceMemberAccess evaluates member access on an interface.
	EvalInterfaceMemberAccess(node *ast.MemberAccessExpression, iface Value, memberName string, ctx *ExecutionContext) Value

	// EvalClassMemberAccess evaluates member access on a class/class info value.
	EvalClassMemberAccess(node *ast.MemberAccessExpression, classVal Value, memberName string, ctx *ExecutionContext) Value

	// EvalTypeCastMemberAccess evaluates member access on a type cast value.
	EvalTypeCastMemberAccess(node *ast.MemberAccessExpression, typeCastVal Value, memberName string, ctx *ExecutionContext) Value

	// EvalNilMemberAccess evaluates member access on nil (class variables may be accessible).
	EvalNilMemberAccess(node *ast.MemberAccessExpression, memberName string, ctx *ExecutionContext) Value

	// EvalRecordMemberAccess evaluates member access on a record.
	EvalRecordMemberAccess(node *ast.MemberAccessExpression, record Value, memberName string, ctx *ExecutionContext) Value

	// EvalEnumMemberAccess evaluates member access on an enum value (e.g., enumVal.Value).
	EvalEnumMemberAccess(node *ast.MemberAccessExpression, enumVal Value, memberName string, ctx *ExecutionContext) Value

	// ===== Task 3.5.54: Collection Expression Adapter Methods =====

	// EvalSetLiteralElements evaluates a set literal with pre-evaluated elements.
	EvalSetLiteralElements(node *ast.SetLiteral, ctx *ExecutionContext) Value

	// EvalEmptyArrayLiteral evaluates an empty array literal (needs type annotation).
	EvalEmptyArrayLiteral(node *ast.ArrayLiteralExpression, ctx *ExecutionContext) Value

	// EvalArrayLiteralWithElements evaluates an array literal with pre-evaluated elements.
	EvalArrayLiteralWithElements(node *ast.ArrayLiteralExpression, elements []Value, ctx *ExecutionContext) Value

	// EvalIndexExpressionWithValues evaluates indexing with pre-evaluated base and index.
	EvalIndexExpressionWithValues(node *ast.IndexExpression, base Value, index Value, ctx *ExecutionContext) Value

	// EvalNewArrayWithDimensions evaluates new array with pre-evaluated dimensions.
	EvalNewArrayWithDimensions(node *ast.NewArrayExpression, dimensions []int, ctx *ExecutionContext) Value

	// EvalRangeExpressionValues evaluates a range expression (used in set literals and case statements).
	EvalRangeExpressionValues(node *ast.RangeExpression, ctx *ExecutionContext) Value
}

// Evaluator is responsible for evaluating DWScript AST nodes.
// It holds the necessary dependencies for evaluation (type system, runtime services, configuration)
// but does NOT hold execution state. Execution state is passed via ExecutionContext.
//
// Phase 3.5.1: This struct separates concerns:
// - TypeSystem: manages types, classes, records, interfaces, operators, conversions
// - Runtime services: I/O, random numbers, external functions
// - Configuration: recursion limits, source file information
// - Unit system: manages unit loading and initialization
//
// The Evaluator is stateless with respect to execution - all execution state
// (environment, call stack, exceptions, control flow) is in ExecutionContext.
type Evaluator struct {
	output            io.Writer
	externalFunctions ExternalFunctionRegistry
	currentNode       ast.Node
	adapter           InterpreterAdapter
	typeSystem        *interptypes.TypeSystem
	rand              *rand.Rand
	config            *Config
	unitRegistry      *units.UnitRegistry
	initializedUnits  map[string]bool
	semanticInfo      *ast.SemanticInfo
	loadedUnits       []string
	randSeed          int64
}

// NewEvaluator creates a new Evaluator with the given dependencies.
func NewEvaluator(
	typeSystem *interptypes.TypeSystem,
	output io.Writer,
	config *Config,
	unitRegistry *units.UnitRegistry,
) *Evaluator {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize random number generator with a default seed
	const defaultSeed = int64(1)
	source := rand.NewSource(defaultSeed)

	return &Evaluator{
		typeSystem:       typeSystem,
		output:           output,
		rand:             rand.New(source),
		randSeed:         defaultSeed,
		config:           config,
		unitRegistry:     unitRegistry,
		initializedUnits: make(map[string]bool),
		loadedUnits:      make([]string, 0),
	}
}

// TypeSystem returns the type system instance.
func (e *Evaluator) TypeSystem() *interptypes.TypeSystem {
	return e.typeSystem
}

// Output returns the output writer.
func (e *Evaluator) Output() io.Writer {
	return e.output
}

// SetOutput sets the output writer.
func (e *Evaluator) SetOutput(w io.Writer) {
	e.output = w
}

// Random returns the random number generator.
func (e *Evaluator) Random() *rand.Rand {
	return e.rand
}

// RandomSeed returns the current random seed.
func (e *Evaluator) RandomSeed() int64 {
	return e.randSeed
}

// SetRandomSeed sets the random seed and reinitializes the generator.
func (e *Evaluator) SetRandomSeed(seed int64) {
	e.randSeed = seed
	source := rand.NewSource(seed)
	e.rand = rand.New(source)
}

// ExternalFunctions returns the external function registry.
func (e *Evaluator) ExternalFunctions() ExternalFunctionRegistry {
	return e.externalFunctions
}

// SetExternalFunctions sets the external function registry.
func (e *Evaluator) SetExternalFunctions(reg ExternalFunctionRegistry) {
	e.externalFunctions = reg
}

// Config returns the configuration.
func (e *Evaluator) Config() *Config {
	return e.config
}

// SetConfig sets the configuration.
func (e *Evaluator) SetConfig(cfg *Config) {
	e.config = cfg
}

// MaxRecursionDepth returns the maximum recursion depth.
func (e *Evaluator) MaxRecursionDepth() int {
	return e.config.MaxRecursionDepth
}

// SourceCode returns the source code being executed.
func (e *Evaluator) SourceCode() string {
	return e.config.SourceCode
}

// SourceFile returns the source file path.
func (e *Evaluator) SourceFile() string {
	return e.config.SourceFile
}

// UnitRegistry returns the unit registry.
func (e *Evaluator) UnitRegistry() *units.UnitRegistry {
	return e.unitRegistry
}

// SetUnitRegistry sets the unit registry.
// Phase 3.5.1: Allows Interpreter to update the registry during migration.
func (e *Evaluator) SetUnitRegistry(registry *units.UnitRegistry) {
	e.unitRegistry = registry
}

// InitializedUnits returns the map of initialized units.
func (e *Evaluator) InitializedUnits() map[string]bool {
	return e.initializedUnits
}

// LoadedUnits returns the list of loaded units.
func (e *Evaluator) LoadedUnits() []string {
	return e.loadedUnits
}

// AddLoadedUnit adds a unit to the list of loaded units.
func (e *Evaluator) AddLoadedUnit(unitName string) {
	e.loadedUnits = append(e.loadedUnits, unitName)
}

// SemanticInfo returns the semantic analysis metadata.
func (e *Evaluator) SemanticInfo() *ast.SemanticInfo {
	return e.semanticInfo
}

// SetSemanticInfo sets the semantic analysis metadata.
func (e *Evaluator) SetSemanticInfo(info *ast.SemanticInfo) {
	e.semanticInfo = info
}

// CurrentNode returns the current AST node being evaluated (for error reporting).
func (e *Evaluator) CurrentNode() ast.Node {
	return e.currentNode
}

// SetCurrentNode sets the current AST node being evaluated (for error reporting).
func (e *Evaluator) SetCurrentNode(node ast.Node) {
	e.currentNode = node
}

// SetAdapter sets the interpreter adapter for delegation during migration.
// Phase 3.5.1: This is temporary and will be removed once migration is complete.
func (e *Evaluator) SetAdapter(adapter InterpreterAdapter) {
	e.adapter = adapter
}

// Eval evaluates an AST node and returns the result value.
// The execution context contains all execution state (environment, call stack, etc.).
//
// Phase 3.5.2: This uses the visitor pattern to dispatch to appropriate handler methods.
// The giant switch statement from Interpreter.Eval() is now here, but organized with
// visitor methods for better separation of concerns.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	// Track current node for error reporting
	e.currentNode = node

	// Phase 3.5.2: Visitor pattern dispatch
	// Dispatch to the appropriate visitor method based on node type
	switch n := node.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return e.VisitIntegerLiteral(n, ctx)
	case *ast.FloatLiteral:
		return e.VisitFloatLiteral(n, ctx)
	case *ast.StringLiteral:
		return e.VisitStringLiteral(n, ctx)
	case *ast.BooleanLiteral:
		return e.VisitBooleanLiteral(n, ctx)
	case *ast.CharLiteral:
		return e.VisitCharLiteral(n, ctx)
	case *ast.NilLiteral:
		return e.VisitNilLiteral(n, ctx)

	// Identifiers
	case *ast.Identifier:
		return e.VisitIdentifier(n, ctx)

	// Expressions
	case *ast.BinaryExpression:
		return e.VisitBinaryExpression(n, ctx)
	case *ast.UnaryExpression:
		return e.VisitUnaryExpression(n, ctx)
	case *ast.AddressOfExpression:
		return e.VisitAddressOfExpression(n, ctx)
	case *ast.GroupedExpression:
		return e.VisitGroupedExpression(n, ctx)
	case *ast.CallExpression:
		return e.VisitCallExpression(n, ctx)
	case *ast.NewExpression:
		return e.VisitNewExpression(n, ctx)
	case *ast.MemberAccessExpression:
		return e.VisitMemberAccessExpression(n, ctx)
	case *ast.MethodCallExpression:
		return e.VisitMethodCallExpression(n, ctx)
	case *ast.InheritedExpression:
		return e.VisitInheritedExpression(n, ctx)
	case *ast.SelfExpression:
		return e.VisitSelfExpression(n, ctx)
	case *ast.EnumLiteral:
		return e.VisitEnumLiteral(n, ctx)
	case *ast.RecordLiteralExpression:
		return e.VisitRecordLiteralExpression(n, ctx)
	case *ast.SetLiteral:
		return e.VisitSetLiteral(n, ctx)
	case *ast.ArrayLiteralExpression:
		return e.VisitArrayLiteralExpression(n, ctx)
	case *ast.IndexExpression:
		return e.VisitIndexExpression(n, ctx)
	case *ast.NewArrayExpression:
		return e.VisitNewArrayExpression(n, ctx)
	case *ast.LambdaExpression:
		return e.VisitLambdaExpression(n, ctx)
	case *ast.IsExpression:
		return e.VisitIsExpression(n, ctx)
	case *ast.AsExpression:
		return e.VisitAsExpression(n, ctx)
	case *ast.ImplementsExpression:
		return e.VisitImplementsExpression(n, ctx)
	case *ast.IfExpression:
		return e.VisitIfExpression(n, ctx)
	case *ast.OldExpression:
		return e.VisitOldExpression(n, ctx)
	case *ast.RangeExpression:
		return e.VisitRangeExpression(n, ctx)

	// Statements
	case *ast.Program:
		return e.VisitProgram(n, ctx)
	case *ast.ExpressionStatement:
		return e.VisitExpressionStatement(n, ctx)
	case *ast.VarDeclStatement:
		return e.VisitVarDeclStatement(n, ctx)
	case *ast.ConstDecl:
		return e.VisitConstDecl(n, ctx)
	case *ast.AssignmentStatement:
		return e.VisitAssignmentStatement(n, ctx)
	case *ast.BlockStatement:
		return e.VisitBlockStatement(n, ctx)
	case *ast.IfStatement:
		return e.VisitIfStatement(n, ctx)
	case *ast.WhileStatement:
		return e.VisitWhileStatement(n, ctx)
	case *ast.RepeatStatement:
		return e.VisitRepeatStatement(n, ctx)
	case *ast.ForStatement:
		return e.VisitForStatement(n, ctx)
	case *ast.ForInStatement:
		return e.VisitForInStatement(n, ctx)
	case *ast.CaseStatement:
		return e.VisitCaseStatement(n, ctx)
	case *ast.TryStatement:
		return e.VisitTryStatement(n, ctx)
	case *ast.RaiseStatement:
		return e.VisitRaiseStatement(n, ctx)
	case *ast.BreakStatement:
		return e.VisitBreakStatement(n, ctx)
	case *ast.ContinueStatement:
		return e.VisitContinueStatement(n, ctx)
	case *ast.ExitStatement:
		return e.VisitExitStatement(n, ctx)
	case *ast.ReturnStatement:
		return e.VisitReturnStatement(n, ctx)
	case *ast.UsesClause:
		return e.VisitUsesClause(n, ctx)

	// Declarations
	case *ast.FunctionDecl:
		return e.VisitFunctionDecl(n, ctx)
	case *ast.ClassDecl:
		return e.VisitClassDecl(n, ctx)
	case *ast.InterfaceDecl:
		return e.VisitInterfaceDecl(n, ctx)
	case *ast.OperatorDecl:
		return e.VisitOperatorDecl(n, ctx)
	case *ast.EnumDecl:
		return e.VisitEnumDecl(n, ctx)
	case *ast.SetDecl:
		return e.VisitSetDecl(n, ctx)
	case *ast.RecordDecl:
		return e.VisitRecordDecl(n, ctx)
	case *ast.HelperDecl:
		return e.VisitHelperDecl(n, ctx)
	case *ast.ArrayDecl:
		return e.VisitArrayDecl(n, ctx)
	case *ast.TypeDeclaration:
		return e.VisitTypeDeclaration(n, ctx)

	default:
		// Phase 3.5.48: All known node types are handled above.
		// Unknown node types indicate a bug (missing case) or an invalid AST.
		panic(fmt.Sprintf("Evaluator.Eval: unknown node type %T", node))
	}
}
