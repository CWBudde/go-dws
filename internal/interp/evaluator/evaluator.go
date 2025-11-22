package evaluator

import (
	"io"
	"math/rand"

	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"

	// Task 3.8.2: pkg/ast is imported for SemanticInfo, which holds semantic analysis
	// metadata (type annotations, symbol resolutions). This is separate from the AST
	// structure itself and is not aliased in internal/ast.
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
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

	// ===== Task 3.5.5: Type System Access Methods =====

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
	semanticInfo      *pkgast.SemanticInfo
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
func (e *Evaluator) SemanticInfo() *pkgast.SemanticInfo {
	return e.semanticInfo
}

// SetSemanticInfo sets the semantic analysis metadata.
func (e *Evaluator) SetSemanticInfo(info *pkgast.SemanticInfo) {
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
