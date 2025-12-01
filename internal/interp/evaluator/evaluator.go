package evaluator

import (
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Value represents a runtime value in the DWScript interpreter.
// This is aliased from the runtime package to match builtins.Context interface.
type Value = runtime.Value

// ObjectValue is an optional interface that object instances can implement
// to provide direct access to class metadata without going through the adapter.
type ObjectValue interface {
	Value
	// ClassName returns the class name of this object instance.
	ClassName() string
	// GetClassType returns the class type (metaclass) for this object instance.
	// Returns a Value that implements ClassMetaValue interface.
	GetClassType() Value
	// HasProperty checks if this object's class has a property with the given name.
	// The check includes the entire class hierarchy.
	HasProperty(name string) bool
	// HasMethod checks if this object's class has a method with the given name.
	HasMethod(name string) bool
	// GetField retrieves the value of a field by name.
	// Returns the field value or nil if the field doesn't exist.
	GetField(name string) Value
	// GetClassVar retrieves a class variable value by name.
	// Returns the value and true if found, nil and false otherwise.
	GetClassVar(name string) (Value, bool)
	// CallInheritedMethod calls a method from the parent class.
	// The methodExecutor callback is used to execute the method once resolved.
	// Returns an error value if:
	//   - The object has no class information
	//   - The class has no parent class
	//   - The method is not found in the parent class
	// Parameters:
	//   - methodName: The name of the method to call (case-insensitive)
	//   - args: The arguments to pass to the method
	//   - methodExecutor: Callback function that executes the method with the resolved declaration
	//     The methodDecl parameter is *ast.FunctionDecl (passed as any to avoid import cycles)
	CallInheritedMethod(methodName string, args []Value, methodExecutor func(methodDecl any, args []Value) Value) Value
	// ReadProperty reads a property value from this object.
	// The propertyExecutor callback handles interpreter-dependent execution:
	//   - For field-backed: returns field value directly
	//   - For method-backed: executes getter method
	//   - For expression-backed: evaluates expression
	// Returns an error value if:
	//   - The object has no class information
	//   - The property is not found in the class hierarchy
	// Parameters:
	//   - propName: The property name (case-insensitive)
	//   - propertyExecutor: Callback function that executes property read with the resolved PropertyInfo
	//     The propInfo parameter is *types.PropertyInfo (passed as any to avoid import cycles)
	ReadProperty(propName string, propertyExecutor func(propInfo any) Value) Value
	// ReadIndexedProperty reads an indexed property value from this object.
	// The propertyExecutor callback handles interpreter-dependent execution:
	//   - Looks up the getter method from PropertyInfo
	//   - Binds Self and index parameters
	//   - Executes the getter method
	//   - Returns the result
	// Parameters:
	//   - propInfo: The property metadata (from PropertyAccessor.LookupProperty or GetDefaultProperty)
	//   - indices: The index values to pass to the getter
	//   - propertyExecutor: Callback function that executes the indexed property read
	//     The propInfo parameter is *types.PropertyInfo (passed as any to avoid import cycles)
	ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value
	// InvokeParameterlessMethod invokes a method if it has zero parameters.
	// Returns:
	//   - (result, true) if method exists and has 0 parameters (invoked via methodExecutor)
	//   - (nil, false) if method has parameters (caller should create method pointer)
	// Parameters:
	//   - methodName: The method name (case-insensitive)
	//   - methodExecutor: Callback that executes the method once resolved
	//     The methodDecl parameter is *ast.FunctionDecl (passed as any to avoid import cycles)
	InvokeParameterlessMethod(methodName string, methodExecutor func(methodDecl any) Value) (Value, bool)
	// CreateMethodPointer creates a method pointer for a method with parameters.
	// The pointerCreator callback handles creating the actual FunctionPointerValue
	// since it requires access to Interpreter's environment and type resolution.
	// Returns:
	//   - (Value, true) if method exists and has parameters (pointer created via callback)
	//   - (nil, false) if method doesn't exist or has no parameters
	// Parameters:
	//   - methodName: The method name (case-insensitive)
	//   - pointerCreator: Callback that creates the FunctionPointerValue
	//     The methodDecl parameter is *ast.FunctionDecl (passed as any to avoid import cycles)
	CreateMethodPointer(methodName string, pointerCreator func(methodDecl any) Value) (Value, bool)
}

// EnumAccessor is an optional interface for enum values.
type EnumAccessor interface {
	Value
	// GetOrdinal returns the ordinal (integer) value of the enum.
	GetOrdinal() int
}

// ExternalVarAccessor is an optional interface for external variable values.
type ExternalVarAccessor interface {
	Value
	// ExternalVarName returns the name of the external variable.
	ExternalVarName() string
}

// LazyEvaluator is an optional interface for lazy parameter thunks.
type LazyEvaluator interface {
	Value
	// Evaluate forces evaluation of the lazy parameter and returns the result.
	Evaluate() Value
}

// InterfaceInstanceValue is an optional interface that interface instances can implement
// to provide direct access to the underlying object and interface metadata without adapter.
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
type ClassMetaValue interface {
	Value
	// GetClassName returns the class name.
	GetClassName() string
	// GetClassType returns the class type (metaclass) as a ClassValue.
	GetClassType() Value
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
// This enables the evaluator to handle property access uniformly across different runtime types.
type PropertyAccessor = runtime.PropertyAccessor

// PropertyDescriptor provides metadata about a property.
// This allows the evaluator to access property metadata without knowing the specific runtime type.
type PropertyDescriptor = runtime.PropertyDescriptor

// RecordInstanceValue is an optional interface that record instances can implement
// to provide direct access to record fields and metadata without going through the adapter.
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
	// ReadIndexedProperty reads an indexed property value using the provided executor callback.
	// The propInfo is already resolved by PropertyAccessor.LookupProperty or GetDefaultProperty.
	// Parameters:
	//   - propInfo: The property implementation (types.RecordPropertyInfo from PropertyDescriptor.Impl)
	//   - indices: The index arguments to pass to the getter
	//   - propertyExecutor: Callback that executes the getter with the resolved property info
	ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value
}

// RecordTypeMetaValue is an optional interface that record type values can implement
// to provide direct access to record type metadata without going through the adapter.
type RecordTypeMetaValue interface {
	Value
	// GetRecordTypeName returns the record type name (e.g., "TPoint").
	GetRecordTypeName() string
	// HasStaticMethod checks if a static method (class function/procedure) with the given name exists.
	// The lookup is case-insensitive.
	HasStaticMethod(name string) bool
}

// SetMethodDispatcher is an optional interface that set values can implement
// to provide direct access to set mutation methods without going through the adapter.
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

// FunctionPointerCallable is an optional interface that function pointer values can implement
// to provide direct invocation without going through the adapter.
// Task 3.5.121: Enables direct function pointer invocation in VisitCallExpression and auto-invoke.
//
// Implementation note: This interface is implemented by runtime.FunctionPointerValue.
// The methods use types that avoid circular imports:
//   - IsNil, ParamCount, IsLambda, HasSelfObject: Simple value methods
//   - GetMetadata: Returns runtime.FunctionPointerMetadata (as struct value)
type FunctionPointerCallable interface {
	Value
	// IsNil returns true if this function pointer has no function or lambda assigned.
	// Used to check before invocation to raise appropriate DWScript exceptions.
	IsNil() bool
	// ParamCount returns the number of parameters this function pointer expects.
	// For lambdas, returns the lambda parameter count.
	// For regular functions, returns the function parameter count.
	// Returns 0 if neither is set.
	ParamCount() int
	// IsLambda returns true if this is a lambda/closure, false for regular function pointers.
	IsLambda() bool
	// HasSelfObject returns true if this is a method pointer with a bound Self object.
	HasSelfObject() bool
	// GetFunctionDecl returns the function AST node (*ast.FunctionDecl) for regular function pointers.
	// Returns nil for lambda closures.
	GetFunctionDecl() any
	// GetLambdaExpr returns the lambda AST node (*ast.LambdaExpression) for lambda closures.
	// Returns nil for regular function pointers.
	GetLambdaExpr() any
	// GetClosure returns the captured environment (type: *Environment).
	// For lambdas, this captures all variables from outer scopes.
	// For functions, this is typically the global environment.
	GetClosure() any
	// GetSelfObject returns the bound Self for method pointers.
	// Returns nil for non-method pointers.
	GetSelfObject() Value
}

// FunctionPointerMetadata provides the execution context for function pointer invocation.
// Task 3.5.121: Passed to the adapter's ExecuteFunctionPointerCall method.
// Note: This mirrors runtime.FunctionPointerMetadata for documentation and type conversion.
type FunctionPointerMetadata struct {
	// IsLambda indicates whether this is a lambda expression
	IsLambda bool
	// Lambda is the lambda AST node (nil for regular function pointers)
	// Type: *ast.LambdaExpression (passed as any to avoid circular import)
	Lambda any
	// Function is the function AST node (nil for lambdas)
	// Type: *ast.FunctionDecl (passed as any to avoid circular import)
	Function any
	// Closure is the captured environment (type: *Environment as interface{})
	Closure any
	// SelfObject is the bound Self for method pointers (nil for non-method pointers)
	SelfObject Value
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
	// Task 3.5.142: Parameterless auto-invoke migrated to evaluator.invokeParameterlessUserFunction()
	// Still used for: function pointer calls, method calls, explicit calls with arguments
	CallUserFunction(fn *ast.FunctionDecl, args []Value) Value

	// Task 3.5.143y: CallBuiltinFunction REMOVED - evaluator now calls builtins directly via registry
	// Evaluator implements builtins.Context and uses builtins.DefaultRegistry.Lookup() instead

	// LookupFunction finds a function by name in the function registry.
	// Returns the function declaration(s) and a boolean indicating success.
	// Multiple functions may be returned for overloaded functions.
	LookupFunction(name string) ([]*ast.FunctionDecl, bool)

	// ===== Declaration Handling =====

	// EvalMethodImplementation handles method implementation registration for classes/records.
	// Task 3.5.7: Delegated to Interpreter because it requires ClassInfo internals
	// (VMT rebuild, descendant propagation).
	// Parameters:
	//   - fn: The method implementation declaration (must have fn.ClassName != nil)
	// Returns NilValue on success, ErrorValue on failure.
	EvalMethodImplementation(fn *ast.FunctionDecl) Value

	// Phase 3.5.4 - Phase 2B: Type system access methods
	// These methods allow the Evaluator to access type registries during evaluation
	// without directly accessing Interpreter fields.

	// ===== Class Registry =====

	// LookupClass finds a class by name in the class registry.
	// Returns the class info (as any/interface{}) and a boolean indicating success.
	// The lookup is case-insensitive.
	LookupClass(name string) (any, bool)

	// ResolveClassInfoByName resolves a class by name for type resolution.
	// Task 3.5.9.4: Allows evaluator to resolve class types in property declarations.
	// Returns the class info (as any/interface{}) or nil if not found.
	ResolveClassInfoByName(name string) interface{}

	// GetClassName returns the name from a raw ClassInfo interface{}.
	// Task 3.5.9.4: Extracts class name for type construction.
	GetClassNameFromInfo(classInfo interface{}) string

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

	// ===== Task 3.5.12: Helper Declaration Adapter Methods =====

	// CreateHelperInfo creates a new HelperInfo instance.
	// Parameters use any to avoid import cycles with internal/types package.
	// targetType is expected to be types.Type from internal/types.
	// Returns the helper info as interface{} to avoid import cycles.
	CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{}

	// SetHelperParent sets the parent helper for inheritance chain.
	SetHelperParent(helper interface{}, parent interface{})

	// VerifyHelperTargetTypeMatch checks if parent helper's target type matches the given type.
	// targetType is expected to be types.Type from internal/types.
	VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool

	// GetHelperName returns the name of a helper (for parent lookup by name).
	GetHelperName(helper interface{}) string

	// AddHelperMethod registers a method in the helper.
	AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl)

	// AddHelperProperty registers a property in the helper.
	// propType is expected to be types.Type from internal/types.
	AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any)

	// AddHelperClassVar adds a class variable to the helper.
	AddHelperClassVar(helper interface{}, name string, value Value)

	// AddHelperClassConst adds a class constant to the helper.
	AddHelperClassConst(helper interface{}, name string, value Value)

	// RegisterHelperLegacy registers the helper in the legacy i.helpers map.
	// This maintains backward compatibility during migration.
	RegisterHelperLegacy(typeName string, helper interface{})

	// ===== Task 3.5.9: Interface Declaration Adapter Methods =====

	// NewInterfaceInfoAdapter creates a new InterfaceInfo instance.
	// Returns the interface info as interface{} to avoid import cycles.
	NewInterfaceInfoAdapter(name string) interface{}

	// CastToInterfaceInfo performs type assertion from any to *InterfaceInfo.
	// Returns the casted interface info and a boolean indicating success.
	CastToInterfaceInfo(iface interface{}) (interface{}, bool)

	// SetInterfaceParent sets the parent interface for inheritance.
	SetInterfaceParent(iface interface{}, parent interface{})

	// GetInterfaceName returns the name of an interface.
	GetInterfaceName(iface interface{}) string

	// GetInterfaceParent returns the parent interface for hierarchy traversal.
	GetInterfaceParent(iface interface{}) interface{}

	// AddInterfaceMethod adds a method to an interface.
	AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl)

	// AddInterfaceProperty adds a property to an interface.
	// propInfo is expected to be *types.PropertyInfo from internal/types.
	AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any)

	// ===== Operator & Conversion Registries =====

	// GetOperatorRegistry returns the operator registry for operator overload lookups.
	// Returns the registry as any/interface{} to avoid circular dependencies.
	GetOperatorRegistry() any

	// ===== Enum Type IDs =====

	// GetEnumTypeID returns the type ID for an enum type, or 0 if not found.
	GetEnumTypeID(enumName string) int

	// ===== Task 3.5.5: Type System Access Methods =====

	// Task 3.5.141: GetType removed - evaluator uses resolveTypeName() directly

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
	// DEPRECATED: Task 3.5.114 - Use ObjectValue.CallInheritedMethod + ExecuteMethodWithSelf instead.
	CallInheritedMethod(obj Value, methodName string, args []Value) Value

	// ExecuteMethodWithSelf executes a method with Self bound to the given object.
	// Task 3.5.114: Low-level method execution for inherited calls.
	// Parameters:
	//   - self: The object to bind as Self in the method environment
	//   - methodDecl: The method declaration (*ast.FunctionDecl, passed as any to avoid import cycles)
	//   - args: The arguments to pass to the method
	// Returns the method result value.
	ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value

	// ===== Object Operations =====

	// CreateObject creates a new object instance of the specified class with constructor arguments.
	// Returns the created object value and an error if the class does not exist or construction fails.
	CreateObject(className string, args []Value) (Value, error)

	// ExecuteConstructor executes a constructor method on an already-created object instance.
	// Task 3.5.126f: Callback for complex constructor execution (method body + Self binding).
	// Returns an error if constructor execution fails.
	ExecuteConstructor(obj Value, constructorName string, args []Value) error

	// CheckType checks if an object is of a specified type (implements 'is' operator).
	// Returns true if the object is compatible with the specified type name.
	CheckType(obj Value, typeName string) bool

	// GetClassMetadataFromValue extracts ClassMetadata from an object value.
	// Task 3.5.140: Helper method to extract metadata from ObjectInstance or InterfaceInstance.
	// Returns nil if the value is not an object or doesn't have class metadata.
	GetClassMetadataFromValue(obj Value) *runtime.ClassMetadata

	// GetObjectInstanceFromValue extracts ObjectInstance from a Value.
	// Task 3.5.141: Helper to extract ObjectInstance for type casting.
	// Returns the ObjectInstance interface{} if the value is an ObjectInstance, nil otherwise.
	GetObjectInstanceFromValue(val Value) interface{}

	// GetInterfaceInstanceFromValue extracts InterfaceInstance from a Value.
	// Task 3.5.141: Helper to extract InterfaceInstance for interface casting.
	// Returns (interfaceInfo, underlyingObject) if the value is an InterfaceInstance, (nil, nil) otherwise.
	GetInterfaceInstanceFromValue(val Value) (interfaceInfo interface{}, underlyingObject interface{})

	// CreateInterfaceWrapper creates an InterfaceInstance wrapper.
	// Task 3.5.141: Helper to create interface wrappers for 'as' operator.
	// Returns the InterfaceInstance wrapper or error if interface not found.
	CreateInterfaceWrapper(interfaceName string, obj Value) (Value, error)

	// CreateTypeCastWrapper creates a TypeCastValue wrapper.
	// Task 3.5.141: Helper to create TypeCastValue for TypeName(expr) casts.
	// Returns the TypeCastValue wrapper or nil if class not found.
	CreateTypeCastWrapper(className string, obj Value) Value

	// RaiseTypeCastException raises a type cast exception.
	// Task 3.5.141: Helper to raise exceptions for invalid TypeName(expr) casts.
	// This matches the behavior of castToClass which raises exceptions.
	RaiseTypeCastException(message string, node ast.Node)

	// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
	// Task 3.5.143p: Helper for Assert() function.
	// The exception includes position information from the current node.
	RaiseAssertionFailed(customMessage string)

	// CreateContractException creates an exception value for contract violations.
	// Task 3.5.142a: Bridge constructor to create exception without import cycles.
	// This is a temporary pattern similar to Task 3.5.129 bridge constructors.
	// Parameters:
	//   - className: Exception class name (e.g., "Exception")
	//   - message: Exception message with position info
	//   - node: AST node for position information (can be nil)
	//   - classMetadata: Class metadata from TypeSystem lookup (can be nil)
	//   - callStack: Call stack trace from ExecutionContext
	// Returns the created exception value as interface{}.
	CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{}

	// CleanupInterfaceReferences releases interface and object references in an environment.
	// Task 3.5.142c: Bridge method to clean up interface-held objects when scope ends.
	// This decrements reference counts and calls destructors for objects that reach zero refs.
	// Parameters:
	//   - env: Environment (interface{}) to clean up
	CleanupInterfaceReferences(env interface{})

	// CreateClassValue creates a ClassValue (metaclass reference) from a class name.
	// Task 3.5.85: Used by VisitIdentifier to return metaclass references for class names.
	// Returns the ClassValue and an error if the class is not found.
	CreateClassValue(className string) (Value, error)

	// ===== Function Pointers =====

	// CreateLambda creates a lambda/closure value from a lambda expression.
	// The closure parameter is the environment where the lambda is created.
	// Returns the lambda value.
	CreateLambda(lambda *ast.LambdaExpression, closure any) Value

	// These are now handled via FunctionPointerCallable interface type assertions.

	// ===== Method Pointers =====

	// CreateMethodPointer creates a method pointer value bound to a specific object.
	// method pointers that capture both the method and the object to call it on.
	// Parameters:
	//   - obj: The object instance (Value) to bind the method to
	//   - methodName: The name of the method to look up
	//   - closure: The environment where the method pointer is created
	// Returns the method pointer value and an error if the method is not found.
	CreateMethodPointer(obj Value, methodName string, closure any) (Value, error)

	// ExecuteFunctionPointerCall executes a function pointer with the given metadata.
	// Task 3.5.121: Low-level execution method used by FunctionPointerCallable.Invoke callback.
	// This handles the interpreter-dependent parts of function pointer invocation:
	//   - For lambdas: Environment setup, parameter binding, body execution
	//   - For method pointers: Self binding, environment creation, function call
	//   - For regular functions: Environment creation, function call
	// Parameters:
	//   - metadata: The execution metadata from FunctionPointerCallable.Invoke
	//   - args: Pre-evaluated argument values
	//   - node: AST node for error location reporting
	// Returns the function result or error value.
	ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value

	// ===== Exception Handling (Task 3.5.8) =====

	// CreateExceptionDirect creates an exception with pre-resolved dependencies.
	// Task 3.5.133: Bridge constructor for exception creation.
	// Parameters use `any` type to avoid import cycles with interp package.
	// The evaluator resolves the exception class via TypeSystem, then calls this method
	// to construct the ExceptionValue without doing class lookup itself.
	CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any

	// WrapObjectInException wraps an existing ObjectInstance in an ExceptionValue.
	// Task 3.5.134: Bridge constructor for raise statement with object instances.
	// The evaluator handles nil checking and validation, this just wraps a valid object.
	// Parameters:
	//   - objInstance: The ObjectInstance to wrap (from evaluator.Value)
	//   - pos: Position information (lexer.Position or *lexer.Position)
	//   - callStack: errors.StackTrace for the exception
	// Returns: ExceptionValue as `any` to avoid import cycles
	WrapObjectInException(objInstance Value, pos any, callStack any) any

	// ===== Environment Access (Task 3.5.9) =====
	// Task 3.5.70: GetVariable removed - use ctx.Env().Get() directly
	// Task 3.5.137: DefineVariable removed - use ctx.Env().Define() directly

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

	// Task 3.5.139h: ParseInlineArrayType removed - evaluator uses parseInlineArrayType() directly

	// Task 3.5.138: LookupSubrangeType removed - use ctx.Env().Get("__subrange_type_" + ident.Normalize(name)) directly

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

	// ===== Bridge Adapter Methods for Zero Value Creation =====

	// CreateSubrangeValueDirect creates a subrange value from subrange type metadata.
	CreateSubrangeValueDirect(subrangeType any) Value

	// CreateInterfaceInstanceDirect creates a nil interface instance from metadata.
	CreateInterfaceInstanceDirect(interfaceInfo any) Value

	// CreateTypedNilValue creates a typed nil value for a class.
	CreateTypedNilValue(className string) Value

	// ===== Property & Method Reference Adapter Methods =====

	// GetObjectFieldValue retrieves a field value from an object instance.
	// Returns the field value and true if found, nil and false otherwise.
	GetObjectFieldValue(obj Value, fieldName string) (Value, bool)

	// GetClassVariableValue retrieves a class variable value from an object's class.
	// Returns the class variable value and true if found, nil and false otherwise.
	GetClassVariableValue(obj Value, varName string) (Value, bool)

	// ReadPropertyValue reads a property value from an object.
	// Handles field-backed, method-backed, and expression-backed properties.
	// Returns the property value and an error if reading fails.
	// Note: Caller is responsible for property recursion prevention.
	// DEPRECATED: Use ObjectValue.ReadProperty() with ExecutePropertyRead callback instead.
	ReadPropertyValue(obj Value, propName string, node any) (Value, error)

	// ExecutePropertyRead executes property reading with a resolved PropertyInfo.
	// This is the callback implementation for ObjectValue.ReadProperty().
	// Parameters:
	//   - obj: The object to read the property from
	//   - propInfo: The resolved PropertyInfo (type *types.PropertyInfo)
	//   - node: AST node for error reporting
	// Returns the property value.
	ExecutePropertyRead(obj Value, propInfo any, node any) Value

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

	// CreateBoundMethodPointer creates a FunctionPointerValue for a method bound to an object.
	//
	// Parameters:
	//   - obj: The object instance to bind the method to
	//   - methodDecl: The method declaration (*ast.FunctionDecl passed as any)
	// Returns: A FunctionPointerValue with proper type information
	CreateBoundMethodPointer(obj Value, methodDecl any) Value

	// GetClassName returns the class name for an object instance.
	// Returns the class name string.
	GetClassName(obj Value) string

	// GetClassType returns the ClassValue (metaclass) for an object instance.
	// Returns the ClassValue representing the object's runtime class.
	GetClassType(obj Value) Value

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

	// ===== Method and Qualified Call Methods =====

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
	//
	// Deprecated: Task 3.5.147 - The evaluator now uses DispatchMethodCall directly.
	// This method is no longer called for RECORD, INTERFACE, or OBJECT method dispatch.
	CallMemberMethod(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression, objVal Value) Value

	// CallQualifiedOrConstructor calls a unit-qualified function or class constructor.
	// This handles:
	// - Unit-qualified calls: UnitName.FunctionName(args)
	// - Class constructor calls: TClassName.Create(args) [DEPRECATED - now uses VisitMethodCallExpression]
	// Parameters:
	//   - callExpr: The original CallExpression AST node
	//   - memberAccess: The MemberAccessExpression (unit.func or class.method)
	// Returns the call result or an error.
	//
	// Deprecated: Task 3.5.147 - Class constructor calls now use VisitMethodCallExpression directly.
	// This method is still used for unit-qualified function calls only.
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
	//
	// Deprecated: Task 3.5.146 - Use DispatchRecordStaticMethod instead.
	// This method re-fetches __CurrentRecord__ which the evaluator already has.
	CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// DispatchRecordStaticMethod dispatches a static method call on a record type.
	// Unlike CallRecordStaticMethod, this method takes the record type name directly,
	// avoiding the need to re-fetch __CurrentRecord__ from the environment.
	// The evaluator handles the lookup and validation via RecordTypeMetaValue interface.
	// Parameters:
	//   - recordTypeName: The record type name (e.g., "TPoint")
	//   - callExpr: The original CallExpression AST node
	//   - funcName: The method identifier
	// Returns the method call result or an error.
	// Task 3.5.146: Simpler adapter method that just creates MethodCallExpression and dispatches.
	DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) Value

	// ===== JSON Value Helpers =====

	// ===== Object Default Property Access =====

	// CallIndexedPropertyGetter calls an indexed property getter method on an object.
	// This is used for default property access: obj[index] -> obj.DefaultProperty[index].
	// Parameters:
	//   - obj: The object instance (ObjectInstance)
	//   - propImpl: The property implementation (types.PropertyInfo from PropertyDescriptor.Impl)
	//   - indices: The index arguments (e.g., [indexValue] for single-index properties)
	//   - node: The AST node for error reporting
	// Returns the result of the property getter method call.
	// DEPRECATED: Use ObjectValue.ReadIndexedProperty with ExecuteIndexedPropertyRead callback instead.
	CallIndexedPropertyGetter(obj Value, propImpl any, indices []Value, node any) Value

	// ExecuteIndexedPropertyRead executes an indexed property read with resolved PropertyInfo.
	// This method handles the interpreter-dependent execution:
	//   - Looks up the getter method from PropertyInfo
	//   - Binds Self and index parameters
	//   - Executes the getter method
	//   - Returns the result
	// Parameters:
	//   - obj: The object instance (ObjectInstance)
	//   - propInfo: The property metadata (*types.PropertyInfo)
	//   - indices: The index values to pass to the getter
	//   - node: The AST node for error reporting
	// Returns the result of the indexed property getter.
	ExecuteIndexedPropertyRead(obj Value, propInfo any, indices []Value, node any) Value

	// ===== Record Default Property Access =====

	// DEPRECATED: Use RecordInstanceValue.ReadIndexedProperty() with ExecuteRecordPropertyRead callback instead.
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

	// ExecuteRecordPropertyRead executes a record property getter method.
	// Task 3.5.118: Low-level callback for RecordInstanceValue.ReadIndexedProperty().
	// This is a thinner wrapper than CallRecordPropertyGetter that is called via the callback pattern.
	// Parameters:
	//   - record: The record value (RecordValue)
	//   - propInfo: The property implementation (types.RecordPropertyInfo from PropertyDescriptor.Impl)
	//   - indices: The index arguments to pass to the getter
	//   - node: The AST node for error reporting
	// Returns the result of the property getter method call.
	ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value

	// ===== Task 3.5.8: Class Declaration Adapter Methods =====

	// NewClassInfoAdapter creates a new ClassInfo with the given name.
	// Returns interface{} for adapter pattern compatibility.
	NewClassInfoAdapter(name string) interface{}

	// CastToClassInfo attempts to cast interface{} to *ClassInfo.
	// Returns the ClassInfo and true if successful, nil and false otherwise.
	CastToClassInfo(class interface{}) (interface{}, bool)

	// GetClassNameFromClassInfoInterface extracts the name from a ClassInfo interface{}.
	// Note: Different from GetClassNameFromClassInfo which takes evaluator.Value.
	GetClassNameFromClassInfoInterface(classInfo interface{}) string

	// RegisterClassEarly registers a class in the legacy class map before full initialization.
	// This enables field initializers to reference the class name.
	RegisterClassEarly(name string, classInfo interface{})

	// IsClassPartial checks if a ClassInfo is marked as partial.
	IsClassPartial(classInfo interface{}) bool

	// SetClassPartial sets the IsPartial flag on a ClassInfo.
	SetClassPartial(classInfo interface{}, isPartial bool)

	// SetClassAbstract sets the IsAbstract flag on a ClassInfo.
	SetClassAbstract(classInfo interface{}, isAbstract bool)

	// SetClassExternal sets the IsExternal flag and ExternalName on a ClassInfo.
	SetClassExternal(classInfo interface{}, isExternal bool, externalName string)

	// ClassHasNoParent checks if a ClassInfo has no parent set yet.
	// Returns true if the class has no parent, false if it already has a parent.
	ClassHasNoParent(classInfo interface{}) bool

	// DefineCurrentClassMarker defines a marker in the environment for the class being declared.
	// This enables nested type resolution to reference the enclosing class.
	DefineCurrentClassMarker(env interface{}, classInfo interface{})

	// SetClassParent sets the parent class and copies all inherited members.
	// This includes fields, methods, constructors, operators, and metadata.
	// Only sets parent if classInfo.Parent is nil (prevents overwriting).
	SetClassParent(classInfo interface{}, parentClass interface{})

	// AddInterfaceToClass adds an interface to a class's interface list.
	// Updates both ClassInfo.Interfaces slice and Metadata.Interfaces.
	AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string)

	// ===== Task 3.5.8 Phase 6: Method, Property, and Operator Adapters =====

	// AddClassMethod adds a method declaration to a ClassInfo.
	// Handles both instance and class methods, method overloading, and constructors/destructors.
	// Creates MethodMetadata and registers with MethodRegistry.
	// Returns true if method was added successfully, false otherwise.
	AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool

	// CreateMethodMetadata creates runtime MethodMetadata from an AST method declaration.
	// Registers the metadata with the MethodRegistry.
	// Returns the created MethodMetadata.
	CreateMethodMetadata(method *ast.FunctionDecl) interface{}

	// SynthesizeDefaultConstructor synthesizes an implicit parameterless constructor
	// for each constructor set that has the 'overload' directive but no parameterless overload.
	// This matches DWScript behavior where overloaded constructors implicitly include a parameterless version.
	SynthesizeDefaultConstructor(classInfo interface{})

	// AddClassProperty adds a property declaration to a ClassInfo.
	// Converts the AST PropertyDecl to PropertyInfo and stores it.
	// Returns true if property was added successfully, false otherwise.
	AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool

	// RegisterClassOperator registers an operator overload for a class.
	// Validates the binding method exists and creates operator entry.
	// Returns nil on success, error Value on failure.
	RegisterClassOperator(classInfo interface{}, opDecl *ast.OperatorDecl) Value

	// LookupClassMethod looks up a method in a ClassInfo by name.
	// If isClassMethod is true, looks in ClassMethods, otherwise in Methods.
	// Returns the method declaration and true if found, nil and false otherwise.
	LookupClassMethod(classInfo interface{}, methodName string, isClassMethod bool) (interface{}, bool)

	// SetClassConstructor sets the constructor field on a ClassInfo (legacy behavior).
	SetClassConstructor(classInfo interface{}, constructor interface{})

	// SetClassDestructor sets the destructor field on a ClassInfo (legacy behavior).
	SetClassDestructor(classInfo interface{}, destructor interface{})

	// InheritDestructorIfMissing inherits destructor from parent if no local destructor declared.
	InheritDestructorIfMissing(classInfo interface{})

	// InheritParentProperties copies parent properties to child class if not already defined.
	InheritParentProperties(classInfo interface{})

	// ===== Task 3.5.8 Phase 7: VMT and Registration Adapters =====

	// BuildVirtualMethodTable builds the virtual method table for a class.
	// Delegates to ClassInfo.buildVirtualMethodTable() which implements proper
	// virtual/override/reintroduce semantics.
	BuildVirtualMethodTable(classInfo interface{})

	// RegisterClassInTypeSystem registers a class in the TypeSystem after VMT is built.
	// Uses TypeSystem.RegisterClassWithParent() for proper hierarchy tracking.
	RegisterClassInTypeSystem(classInfo interface{}, parentName string)

	// ===== Type Conversion & Introspection Methods (Task 3.5.143g) =====
	// Note: ToInt64, ToBool, ToFloat64, GetTypeOf, GetClassOf are NOT part of this adapter interface.
	// They are part of builtins.Context interface and are implemented independently on both
	// Interpreter (in builtins_context.go) and Evaluator (in context_conversions.go).
	// The Evaluator does not delegate these methods to the adapter.
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
	currentContext    *ExecutionContext // Task 3.5.143n: For Context methods needing runtime state (call stack)
}

// Ensure Evaluator implements builtins.Context interface at compile time.
// Task 3.5.143w: Compile-time verification that all 40 Context methods are implemented.
var _ builtins.Context = (*Evaluator)(nil)

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
// Direct Environment Access Helpers
// ============================================================================
// These methods provide direct access to environment operations without going
// through the adapter. They handle the interface{} to Value type conversion.

// GetVar retrieves a variable from the execution context's environment.
// Returns the value and whether it was found.
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

// raiseMaxRecursionExceeded raises a max recursion exception.
// Task 3.5.142: Helper for parameterless function invocation.
func (e *Evaluator) raiseMaxRecursionExceeded(node ast.Node) Value {
	return e.newError(node, "maximum recursion depth exceeded")
}

// Eval evaluates an AST node and returns the result value.
// The execution context contains all execution state (environment, call stack, etc.).
//
// Phase 3.5.2: This uses the visitor pattern to dispatch to appropriate handler methods.
// The giant switch statement from Interpreter.Eval() is now here, but organized with
// visitor methods for better separation of concerns.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	// Task 3.5.143n: Set currentContext for Context interface methods (call stack access)
	e.currentContext = ctx
	defer func() { e.currentContext = nil }()

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
