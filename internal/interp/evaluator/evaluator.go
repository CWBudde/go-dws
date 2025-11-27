package evaluator

import (
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

// ObjectValue is an optional interface that object instances can implement
// to provide direct access to class metadata without going through the adapter.
// Task 3.5.72: Enables direct property/method existence checks.
// Task 3.5.86: Extended with GetField and GetClassVar for member access.
type ObjectValue interface {
	Value
	// ClassName returns the class name of this object instance.
	ClassName() string
	// HasProperty checks if this object's class has a property with the given name.
	// The check includes the entire class hierarchy.
	HasProperty(name string) bool
	// HasMethod checks if this object's class has a method with the given name.
	HasMethod(name string) bool
	// GetField retrieves the value of a field by name.
	// Returns the field value or nil if the field doesn't exist.
	// Task 3.5.86: Enables direct field access without adapter.
	GetField(name string) Value
	// GetClassVar retrieves a class variable value by name.
	// Returns the value and true if found, nil and false otherwise.
	// Task 3.5.86: Enables direct class variable access without adapter.
	GetClassVar(name string) (Value, bool)
}

// EnumAccessor is an optional interface for enum values.
// Task 3.5.89: Enables direct access to enum ordinal value without adapter.
type EnumAccessor interface {
	Value
	// GetOrdinal returns the ordinal (integer) value of the enum.
	GetOrdinal() int
}

// ExternalVarAccessor is an optional interface for external variable values.
// Task 3.5.73: Enables direct access to external variable name without adapter.
type ExternalVarAccessor interface {
	Value
	// ExternalVarName returns the name of the external variable.
	ExternalVarName() string
}

// LazyEvaluator is an optional interface for lazy parameter thunks.
// Task 3.5.73: Enables direct lazy evaluation without adapter.
type LazyEvaluator interface {
	Value
	// Evaluate forces evaluation of the lazy parameter and returns the result.
	Evaluate() Value
}

// InterfaceInstanceValue is an optional interface that interface instances can implement
// to provide direct access to the underlying object and interface metadata without adapter.
// Task 3.5.87: Enables direct interface member access verification.
type InterfaceInstanceValue interface {
	Value
	// GetUnderlyingObjectValue returns the object wrapped by this interface instance.
	// Returns nil if the interface instance wraps a nil object.
	// Note: Returns Value to avoid circular imports; caller should type-assert to ObjectValue.
	// Note: Named differently from GetUnderlyingObject() to allow coexistence with
	// concrete return type method for backwards compatibility.
	GetUnderlyingObjectValue() Value
	// InterfaceName returns the name of the interface type.
	// Used for error messages and debugging.
	InterfaceName() string
	// HasInterfaceMethod checks if the interface declares a method with the given name.
	// The check includes parent interfaces.
	HasInterfaceMethod(name string) bool
	// HasInterfaceProperty checks if the interface declares a property with the given name.
	// The check includes parent interfaces.
	HasInterfaceProperty(name string) bool
}

// ClassMetaValue is an optional interface that class references (ClassValue, ClassInfoValue)
// can implement to provide direct access to class metadata without going through the adapter.
// Task 3.5.88: Enables direct class member access for CLASS and CLASS_INFO value types.
type ClassMetaValue interface {
	Value
	// GetClassName returns the class name.
	GetClassName() string
	// GetClassVar retrieves a class variable value by name from the class hierarchy.
	// Returns the value and true if found, nil and false otherwise.
	// The lookup is case-insensitive.
	GetClassVar(name string) (Value, bool)
	// GetClassConstant retrieves a class constant value by name from the class hierarchy.
	// Returns the value and true if found, nil and false otherwise.
	// The lookup is case-insensitive.
	GetClassConstant(name string) (Value, bool)
	// HasClassMethod checks if a class method with the given name exists.
	// The lookup is case-insensitive and includes the entire class hierarchy.
	HasClassMethod(name string) bool
	// HasConstructor checks if a constructor with the given name exists.
	// The lookup is case-insensitive and includes the entire class hierarchy.
	HasConstructor(name string) bool
}

// TypeCastAccessor is an optional interface for type cast values.
// Task 3.5.89: Enables direct access to the static type and wrapped value without adapter.
// TypeCastValue wraps an object with its static type from a type cast expression.
// Example: TBase(childObj).ClassVar should access TBase's class variable, not TChild's.
type TypeCastAccessor interface {
	Value
	// GetStaticTypeName returns the static type name from the cast (e.g., "TBase").
	GetStaticTypeName() string
	// GetWrappedValue returns the actual value wrapped by the type cast.
	// This is the runtime object (could be ObjectInstance, NilValue, etc.).
	GetWrappedValue() Value
	// GetStaticClassVar retrieves a class variable from the static type's class hierarchy.
	// This is the key operation for type-cast member access: TBase(child).ClassVar
	// must access TBase's class variable, not TChild's.
	// Returns the value and true if found, nil and false otherwise.
	GetStaticClassVar(name string) (Value, bool)
}

// NilAccessor is an optional interface for nil values.
// Task 3.5.90: Enables direct access to the typed class name for nil values.
// Typed nil values (e.g., `var b: TBase := nil`) can access class variables
// but not instance members.
type NilAccessor interface {
	Value
	// GetTypedClassName returns the class type name for typed nil values.
	// Returns "" for untyped nil values.
	// Example: For `var b: TBase := nil`, returns "TBase".
	GetTypedClassName() string
}

// PropertyAccessor is an optional interface for values that support property access.
// Task 3.5.99a: Provides common abstraction for property lookup on objects, interfaces, and records.
// This enables the evaluator to handle property access uniformly across different runtime types.
type PropertyAccessor interface {
	Value
	// LookupProperty searches for a property by name in the type hierarchy.
	// Returns a PropertyDescriptor with metadata needed for property access.
	// Returns nil if the property is not found.
	// The lookup is case-insensitive and includes parent types where applicable.
	LookupProperty(name string) *PropertyDescriptor

	// GetDefaultProperty returns the default property for this type, if any.
	// Default properties allow indexing syntax: obj[index] instead of obj.Property[index].
	// Returns nil if no default property is defined.
	GetDefaultProperty() *PropertyDescriptor
}

// PropertyDescriptor provides metadata about a property.
// Task 3.5.99a: Abstracts property info across classes, interfaces, and records.
// This allows the evaluator to access property metadata without knowing the specific runtime type.
type PropertyDescriptor struct {
	Name      string // Property name
	IsIndexed bool   // True if this is an indexed property (e.g., property Items[Index: Integer]: String)
	IsDefault bool   // True if this is the default property

	// For implementation reference:
	// - Objects: pointer to types.PropertyInfo
	// - Interfaces: pointer to types.PropertyInfo
	// - Records: pointer to types.RecordPropertyInfo
	// We store as `any` to avoid circular imports and maintain type flexibility
	Impl any
}

// RecordInstanceValue is an optional interface that record instances can implement
// to provide direct access to record fields and metadata without going through the adapter.
// Task 3.5.91: Enables direct record field access in VisitMemberAccessExpression.
type RecordInstanceValue interface {
	Value
	// GetRecordField retrieves a field value by name (case-insensitive lookup).
	// Returns the field value and true if found, nil and false otherwise.
	GetRecordField(name string) (Value, bool)
	// GetRecordTypeName returns the record type name (e.g., "TPoint").
	// Returns "RECORD" if the type name is not available.
	GetRecordTypeName() string
	// HasRecordMethod checks if a method with the given name exists on this record type.
	// The lookup is case-insensitive.
	HasRecordMethod(name string) bool
	// HasRecordProperty checks if a property with the given name exists.
	// Note: Records in DWScript don't have properties (unlike classes), so this
	// typically returns false. Included for consistency with other value interfaces.
	HasRecordProperty(name string) bool
}

// SetMethodDispatcher is an optional interface that set values can implement
// to provide direct access to set mutation methods without going through the adapter.
// Task 3.5.111a: Enables direct set method dispatch (Include, Exclude) in VisitMethodCallExpression.
type SetMethodDispatcher interface {
	Value
	// AddElement adds an element with the given ordinal value to the set.
	// This mutates the set in place (used for Include method).
	AddElement(ordinal int)
	// RemoveElement removes an element with the given ordinal value from the set.
	// This mutates the set in place (used for Exclude method).
	RemoveElement(ordinal int)
	// GetSetElementTypeName returns the element type name for error messages.
	// Returns "Unknown" if the element type cannot be determined.
	GetSetElementTypeName() string
}

// EnumTypeMetaDispatcher is an optional interface that type meta values wrapping
// enum types can implement to provide direct access to enum type methods.
// Task 3.5.111b: Enables direct enum type method dispatch (Low, High, ByName) in VisitMethodCallExpression.
type EnumTypeMetaDispatcher interface {
	Value
	// IsEnumTypeMeta returns true if this type meta wraps an enum type.
	IsEnumTypeMeta() bool
	// EnumLow returns the lowest ordinal value of the enum type.
	EnumLow() int
	// EnumHigh returns the highest ordinal value of the enum type.
	EnumHigh() int
	// EnumByName looks up an enum value by name (case-insensitive).
	// Supports both simple names ('Red') and qualified names ('TColor.Red').
	// Returns the ordinal value if found, or 0 if not found (DWScript behavior).
	EnumByName(name string) int
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

	// LookupFunction finds a function by name in the function registry.
	// Returns the function declaration(s) and a boolean indicating success.
	// Multiple functions may be returned for overloaded functions.
	LookupFunction(name string) ([]*ast.FunctionDecl, bool)

	// Phase 3.5.4 - Phase 2B: Type system access methods
	// These methods allow the Evaluator to access type registries during evaluation
	// without directly accessing Interpreter fields.

	// ===== Class Registry =====

	// LookupClass finds a class by name in the class registry.
	// Returns the class info (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupClass(name string) (any, bool)

	// ===== Record Registry =====

	// LookupRecord finds a record type by name in the record registry.
	// Returns the record type value (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupRecord(name string) (any, bool)

	// ===== Interface Registry =====

	// LookupInterface finds an interface by name in the interface registry.
	// Returns the interface info (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupInterface(name string) (any, bool)

	// ===== Helper Registry =====

	// LookupHelpers finds helper methods for a type by name.
	// Returns a slice of helper info (each element as any/interface{}).
	// The lookup is case-insensitive.
	LookupHelpers(typeName string) []any

	// ===== Operator & Conversion Registries =====

	// GetOperatorRegistry returns the operator registry for operator overload lookups.
	// Returns the registry as any/interface{} to avoid circular dependencies.
	GetOperatorRegistry() any

	// ===== Enum Type IDs =====

	// GetEnumTypeID returns the type ID for an enum type, or 0 if not found.
	GetEnumTypeID(enumName string) int

	// ===== Task 3.5.5: Type System Access Methods =====

	// GetType resolves a type by name.
	// Returns the resolved type and an error if the type is not found.
	// The lookup is case-insensitive.
	GetType(name string) (any, error)

	// ConvertValue performs implicit or explicit type conversion.
	// Returns the converted value or an error if conversion is not possible.
	ConvertValue(value Value, targetTypeName string) (Value, error)

	// ===== Task 3.5.6: Array and Collection Adapter Methods =====

	// CreateArray creates an array from a list of elements with a specified element type.
	// Returns the created array value.
	CreateArray(elementType any, elements []Value) Value

	// CreateArrayValue creates an ArrayValue with the specified array type and elements.
	// Task 3.5.83: Direct array construction without re-evaluation.
	// Parameters:
	//   - arrayType: The *types.ArrayType for the array (passed as any to avoid import cycles)
	//   - elements: The pre-evaluated element values
	// Returns the created ArrayValue.
	CreateArrayValue(arrayType any, elements []Value) Value

	// ===== Task 3.5.7: Property, Field, and Member Access Adapter Methods =====

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

	// CastToClass performs class type casting for TypeName(expr) expressions.
	// Task 3.5.94: Delegates class casting logic to the Interpreter during type cast migration.
	// Returns the casted value with proper static type preservation.
	CastToClass(val Value, className string, node ast.Expression) Value

	// CheckImplements checks if an object/class implements an interface (implements 'implements' operator).
	// Task 3.5.36: Supports ObjectInstance, ClassValue, and ClassInfoValue inputs.
	// Returns true if the class implements the specified interface.
	CheckImplements(obj Value, interfaceName string) (bool, error)

	// CreateClassValue creates a ClassValue (metaclass reference) from a class name.
	// Task 3.5.85: Used by VisitIdentifier to return metaclass references for class names.
	// Returns the ClassValue and an error if the class is not found.
	CreateClassValue(className string) (Value, error)

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

	// ===== Exception Handling (Task 3.5.8) =====

	// RaiseException raises an exception with the given class name and message.
	// The pos parameter provides source location information for error reporting.
	RaiseException(className string, message string, pos any)

	// ===== Environment Access (Task 3.5.9) =====
	// Task 3.5.70: GetVariable removed - use ctx.Env().Get() directly

	// DefineVariable defines a new variable in the execution context.
	// This creates a new binding in the current scope.
	DefineVariable(name string, value Value, ctx *ExecutionContext)

	// Phase 3.5.4 - Phase 2C: Property & Indexing System infrastructure
	// Property and indexing operations are available through existing infrastructure:
	//
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

	// ===== Task 3.5.21: Complex Value Retrieval Adapter Methods =====

	// Task 3.5.71: IsReferenceValue removed - use val.Type() == "REFERENCE" directly
	// Task 3.5.73: IsExternalVar, IsLazyThunk, EvaluateLazyThunk, GetExternalVarName removed
	//              - use ExternalVarAccessor and LazyEvaluator interfaces directly

	// DereferenceValue dereferences a var parameter reference.
	// Returns the actual value and an error if dereferencing fails.
	// Panics if the value is not a ReferenceValue (check with IsReferenceValue first).
	DereferenceValue(value Value) (Value, error)

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
	// Task 3.5.71: IsObjectInstance removed - use val.Type() == "OBJECT" directly

	// GetObjectFieldValue retrieves a field value from an object instance.
	// Returns the field value and true if found, nil and false otherwise.
	GetObjectFieldValue(obj Value, fieldName string) (Value, bool)

	// GetClassVariableValue retrieves a class variable value from an object's class.
	// Returns the class variable value and true if found, nil and false otherwise.
	GetClassVariableValue(obj Value, varName string) (Value, bool)

	// Task 3.5.72: HasProperty removed - use ObjectValue interface directly

	// ReadPropertyValue reads a property value from an object.
	// Handles field-backed, method-backed, and expression-backed properties.
	// Returns the property value and an error if reading fails.
	// Note: Caller is responsible for property recursion prevention.
	ReadPropertyValue(obj Value, propName string, node any) (Value, error)

	// Task 3.5.72: HasMethod removed - use ObjectValue interface directly

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

	// Task 3.5.71: IsClassInfoValue removed - use val.Type() == "CLASSINFO" directly

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

	// EvalBlockStatement evaluates a block statement in the given context.
	// Returns nil after evaluating all statements.
	EvalBlockStatement(block *ast.BlockStatement, ctx *ExecutionContext)

	// EvalStatement evaluates a single statement in the given context.
	// Used by exception handlers to evaluate the handler statement.
	EvalStatement(stmt ast.Statement, ctx *ExecutionContext)

	// ===== Task 3.5.96: Method and Qualified Call Methods =====

	// CallMemberMethod calls a method on an object (record, interface, or object instance).
	// This handles:
	// - Record method calls: recVal.Method(args)
	// - Interface method calls: ifaceVal.Method(args) - dispatches to underlying object
	// - Object method calls: objVal.Method(args)
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - memberAccess: The MemberAccessExpression (obj.member)
	//   - objVal: The evaluated object value
	// Returns the method call result or an error.
	CallMemberMethod(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression, objVal Value) Value

	// CallQualifiedOrConstructor calls a unit-qualified function or class constructor.
	// This handles:
	// - Unit-qualified calls: UnitName.FunctionName(args)
	// - Class constructor calls: TClassName.Create(args)
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - memberAccess: The MemberAccessExpression (unit.func or class.method)
	// Returns the call result or an error.
	CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) Value

	// ===== Task 3.5.97: User Function Call Methods =====

	// CallUserFunctionWithOverloads calls a user-defined function with overload resolution.
	// This handles:
	// - Single function calls: MyFunction(args)
	// - Overloaded function calls: OverloadedFunc(args) - resolves based on argument types
	// - Parameter preparation: lazy parameters get LazyThunks, var parameters get References
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - funcName: The function identifier
	// Returns the function call result or an error.
	// Task 3.5.97a: Encapsulates overload resolution and parameter preparation.
	CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// CallImplicitSelfMethod calls a method on the implicit Self object.
	// This handles:
	// - MethodName() inside instance methods → Self.MethodName()
	// - Converts simple CallExpression to MethodCallExpression
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - funcName: The method identifier
	// Returns the method call result or an error.
	// Task 3.5.97b: Enables implicit method calls without EvalNode.
	CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// CallRecordStaticMethod calls a static method within a record context.
	// This handles:
	// - MethodName() inside record static methods → TRecord.MethodName()
	// - Looks up method in __CurrentRecord__ context
	// - Converts to MethodCallExpression with record type
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - funcName: The method identifier
	// Returns the method call result or an error.
	// Task 3.5.97c: Enables record static method calls without EvalNode.
	CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// ===== Task 3.5.99b: JSON Value Helpers =====

	// WrapJSONValueInVariant wraps a jsonvalue.Value in a VariantValue containing a JSONValue.
	// This creates the necessary JSONValue wrapper and boxes it in a Variant.
	// The JSON value becomes available for variant operations.
	// Parameters:
	//   - jv: The jsonvalue.Value to wrap (nil creates a JSON null)
	// Returns a VariantValue containing a JSONValue.
	// Task 3.5.99b: Enables JSON indexing without circular imports.
	WrapJSONValueInVariant(jv any) Value

	// ===== Task 3.5.99c: Object Default Property Access =====

	// CallIndexedPropertyGetter calls an indexed property getter method on an object.
	// This is used for default property access: obj[index] -> obj.DefaultProperty[index].
	// Parameters:
	//   - obj: The object instance (ObjectInstance)
	//   - propImpl: The property implementation (types.PropertyInfo from PropertyDescriptor.Impl)
	//   - indices: The index arguments (e.g., [indexValue] for single-index properties)
	//   - node: The AST node for error reporting
	// Returns the result of the property getter method call.
	// Task 3.5.99c: Enables object default property indexing in evaluator.
	CallIndexedPropertyGetter(obj Value, propImpl any, indices []Value, node any) Value

	// ===== Task 3.5.99e: Record Default Property Access =====

	// CallRecordPropertyGetter calls a record property getter method.
	// This is used for record default property access: record[index] -> record.GetProperty(index).
	// Parameters:
	//   - record: The record value (RecordValue)
	//   - propImpl: The property implementation (types.RecordPropertyInfo from PropertyDescriptor.Impl)
	//   - indices: The index arguments (e.g., [indexValue] for single-index properties)
	//   - node: The AST node for error reporting
	// Returns the result of the property getter method call.
	// Task 3.5.99e: Enables record default property indexing in evaluator.
	CallRecordPropertyGetter(record Value, propImpl any, indices []Value, node any) Value
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
// Task 3.5.76: semanticInfo is now passed via constructor (like TypeRegistry)
// for explicit dependency injection.
func NewEvaluator(
	typeSystem *interptypes.TypeSystem,
	output io.Writer,
	config *Config,
	unitRegistry *units.UnitRegistry,
	semanticInfo *ast.SemanticInfo,
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
		semanticInfo:     semanticInfo,
	}
}

// TypeSystem returns the type system instance.
func (e *Evaluator) TypeSystem() *interptypes.TypeSystem {
	return e.typeSystem
}

// FunctionRegistry returns the function registry for direct function lookups.
// Task 3.5.62: Provides direct access to FunctionRegistry without going through adapter.
func (e *Evaluator) FunctionRegistry() *interptypes.FunctionRegistry {
	return e.typeSystem.Functions()
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

// ============================================================================
// Task 3.5.63: Direct Environment Access Helpers
// ============================================================================
// These methods provide direct access to environment operations without going
// through the adapter. They handle the interface{} to Value type conversion.

// GetVar retrieves a variable from the execution context's environment.
// Returns the value and whether it was found.
// Task 3.5.63: Direct environment access without adapter.
func (e *Evaluator) GetVar(ctx *ExecutionContext, name string) (Value, bool) {
	val, found := ctx.Env().Get(name)
	if !found {
		return nil, false
	}
	// The environment stores interface{}, but we know it's always a Value
	if v, ok := val.(Value); ok {
		return v, true
	}
	return nil, false
}

// DefineVar defines a new variable in the execution context's environment.
// Task 3.5.63: Direct environment access without adapter.
func (e *Evaluator) DefineVar(ctx *ExecutionContext, name string, value Value) {
	ctx.Env().Define(name, value)
}

// SetVar updates an existing variable in the execution context's environment.
// Returns true if the variable existed and was updated, false otherwise.
// Task 3.5.63: Direct environment access without adapter.
func (e *Evaluator) SetVar(ctx *ExecutionContext, name string, value Value) bool {
	return ctx.Env().Set(name, value)
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
		// Phase 3.5.2: Unknown node type - delegate to adapter if available
		// This provides a safety net during the migration
		if e.adapter != nil {
			return e.adapter.EvalNode(node)
		}
		// If no adapter, this is an error (unknown node type)
		panic("Evaluator.Eval: unknown node type and no adapter available")
	}
}
