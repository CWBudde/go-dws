package evaluator

import (
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/interp/builtins"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Value represents a runtime value in the DWScript interpreter.
type Value = runtime.Value

// ObjectValue provides direct access to class metadata without the adapter.
type ObjectValue interface {
	Value
	ClassName() string
	GetClassType() Value
	HasProperty(name string) bool
	HasMethod(name string) bool
	// GetMethodDecl retrieves method declaration by name from the class hierarchy.
	// Returns *ast.FunctionDecl (passed as any) or nil if not found.
	GetMethodDecl(name string) any
	GetField(name string) Value
	GetClassVar(name string) (Value, bool)
	// CallInheritedMethod calls a parent class method.
	// methodExecutor callback executes the resolved method (*ast.FunctionDecl as any).
	CallInheritedMethod(methodName string, args []Value, methodExecutor func(methodDecl any, args []Value) Value) Value
	// ReadProperty reads a property value using the propertyExecutor callback.
	// Supports field-backed, method-backed, and expression-backed properties.
	ReadProperty(propName string, propertyExecutor func(propInfo any) Value) Value
	// ReadIndexedProperty reads indexed property via propertyExecutor callback.
	ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value
	// WriteProperty writes a property value via propertyExecutor callback.
	// Supports field-backed and method-backed properties.
	WriteProperty(propName string, value Value, propertyExecutor func(propInfo any, value Value) Value) Value
	// WriteIndexedProperty writes indexed property via propertyExecutor callback.
	WriteIndexedProperty(propInfo any, indices []Value, value Value, propertyExecutor func(propInfo any, indices []Value, value Value) Value) Value
	// InvokeParameterlessMethod invokes zero-parameter methods via methodExecutor callback.
	// Returns (result, true) if method exists and has 0 parameters, (nil, false) otherwise.
	InvokeParameterlessMethod(methodName string, methodExecutor func(methodDecl any) Value) (Value, bool)
	// CreateMethodPointer creates method pointer via pointerCreator callback.
	// Returns (Value, true) if method exists and has parameters, (nil, false) otherwise.
	CreateMethodPointer(methodName string, pointerCreator func(methodDecl any) Value) (Value, bool)
}

// EnumAccessor provides access to enum ordinal values.
type EnumAccessor interface {
	Value
	GetOrdinal() int
}

// ExternalVarAccessor provides access to external variable names.
type ExternalVarAccessor interface {
	Value
	ExternalVarName() string
}

// LazyEvaluator supports lazy parameter evaluation.
type LazyEvaluator interface {
	Value
	Evaluate() Value
}

// InterfaceInstanceValue provides access to interface instance metadata.
type InterfaceInstanceValue interface {
	Value
	// GetUnderlyingObjectValue returns the wrapped object (nil if interface wraps nil).
	GetUnderlyingObjectValue() Value
	InterfaceName() string
	HasInterfaceMethod(name string) bool
	HasInterfaceProperty(name string) bool
}

// ClassMetaValue provides access to class metadata (class references).
type ClassMetaValue interface {
	Value
	GetClassName() string
	GetClassType() Value
	GetClassVar(name string) (Value, bool)
	GetClassConstant(name string) (Value, bool)
	HasClassMethod(name string) bool
	HasConstructor(name string) bool
}

// TypeCastAccessor wraps objects with their static type from cast expressions.
// Example: TBase(child).ClassVar accesses TBase's class variable, not TChild's.
type TypeCastAccessor interface {
	Value
	GetStaticTypeName() string
	GetWrappedValue() Value
	GetStaticClassVar(name string) (Value, bool)
}

// NilAccessor provides typed nil support.
// Typed nil values can access class variables but not instance members.
type NilAccessor interface {
	Value
	// GetTypedClassName returns the class type name for typed nil, "" for untyped nil.
	GetTypedClassName() string
}

// PropertyAccessor enables uniform property access across runtime types.
type PropertyAccessor = runtime.PropertyAccessor

// PropertyDescriptor provides property metadata.
type PropertyDescriptor = runtime.PropertyDescriptor

// RecordInstanceValue provides access to record fields and metadata.
type RecordInstanceValue interface {
	Value
	GetRecordTypeName() string
	GetRecordField(name string) (Value, bool)
	HasRecordMethod(name string) bool
	HasRecordProperty(name string) bool

	// NEW: Retrieve the AST declaration for a record method.
	// Returns the method declaration and true if found, nil and false otherwise.
	// The name comparison is case-insensitive (DWScript convention).
	GetRecordMethod(name string) (*ast.FunctionDecl, bool)

	// ReadIndexedProperty reads indexed property via propertyExecutor callback.
	ReadIndexedProperty(propInfo any, indices []Value, propertyExecutor func(propInfo any, indices []Value) Value) Value
}

// RecordTypeMetaValue provides access to record type metadata.
type RecordTypeMetaValue interface {
	Value
	GetRecordTypeName() string
	HasStaticMethod(name string) bool
}

// SetMethodDispatcher provides set mutation operations.
type SetMethodDispatcher interface {
	Value
	AddElement(ordinal int)
	RemoveElement(ordinal int)
	GetSetElementTypeName() string
}

// EnumTypeMetaDispatcher provides enum type operations.
type EnumTypeMetaDispatcher interface {
	Value
	IsEnumTypeMeta() bool
	EnumLow() int
	EnumHigh() int
	// EnumByName supports simple ('Red') and qualified ('TColor.Red') names.
	EnumByName(name string) int
	// GetEnumValue looks up an enum value by name and returns it as a runtime Value.
	// Returns nil if the name is not found. Used for member access like TColor.Red.
	GetEnumValue(name string) Value
}

// FunctionPointerCallable enables direct function pointer invocation.
type FunctionPointerCallable interface {
	Value
	IsNil() bool
	ParamCount() int
	IsLambda() bool
	HasSelfObject() bool
	GetFunctionDecl() any // Returns *ast.FunctionDecl (nil for lambdas)
	GetLambdaExpr() any   // Returns *ast.LambdaExpression (nil for functions)
	GetClosure() any      // Returns *Environment
	GetSelfObject() Value
}

// FunctionPointerMetadata provides execution context for function pointer invocation.
type FunctionPointerMetadata struct {
	Lambda     any
	Function   any
	Closure    any
	SelfObject Value
	IsLambda   bool
}

// Config holds evaluator configuration options.
type Config struct {
	SourceCode        string
	SourceFile        string
	MaxRecursionDepth int
}

// DefaultConfig returns default configuration (matches DWScript defaults).
func DefaultConfig() *Config {
	return &Config{
		MaxRecursionDepth: 1024,
		SourceCode:        "",
		SourceFile:        "",
	}
}

// ExternalFunctionRegistry manages external Go functions callable from DWScript.
type ExternalFunctionRegistry interface {
	// Has checks if an external function with the given name is registered.
	Has(name string) bool
}

// InterpreterAdapter provides the boundary between evaluator and interpreter.
// The evaluator is independent and operates on ExecutionContext and runtime values.
// The interpreter owns OOP semantics (classes, Self context, method dispatch).
type InterpreterAdapter interface {
	// ===== Core Execution =====

	// EvalNode evaluates a node via interpreter for OOP operations.
	// Handles Self/class context and complex OOP operations.
	EvalNode(node ast.Node) Value

	// CallFunctionPointer executes a function pointer with arguments.
	CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value

	// CallUserFunction executes a user-defined function.
	CallUserFunction(fn *ast.FunctionDecl, args []Value) Value

	// ===== Declaration Handling =====

	// EvalMethodImplementation registers method implementations for classes/records.
	// Handles ClassInfo internals (VMT rebuild, descendant propagation).
	EvalMethodImplementation(fn *ast.FunctionDecl) Value

	// ===== Helper Declaration =====

	CreateHelperInfo(name string, targetType any, isRecordHelper bool) interface{}
	SetHelperParent(helper interface{}, parent interface{})
	VerifyHelperTargetTypeMatch(parent interface{}, targetType any) bool
	GetHelperName(helper interface{}) string
	AddHelperMethod(helper interface{}, normalizedName string, method *ast.FunctionDecl)
	AddHelperProperty(helper interface{}, prop *ast.PropertyDecl, propType any)
	AddHelperClassVar(helper interface{}, name string, value Value)
	AddHelperClassConst(helper interface{}, name string, value Value)
	RegisterHelperLegacy(typeName string, helper interface{})

	// ===== Interface Declaration =====

	NewInterfaceInfoAdapter(name string) interface{}
	CastToInterfaceInfo(iface interface{}) (interface{}, bool)
	SetInterfaceParent(iface interface{}, parent interface{})
	GetInterfaceName(iface interface{}) string
	GetInterfaceParent(iface interface{}) interface{}
	AddInterfaceMethod(iface interface{}, normalizedName string, method *ast.FunctionDecl)
	AddInterfaceProperty(iface interface{}, normalizedName string, propInfo any)

	// ===== Method Calls =====

	CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value
	CallInheritedMethod(obj Value, methodName string, args []Value) Value
	ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value

	// ===== Object Operations =====

	ExecuteConstructor(obj Value, constructorName string, args []Value) error
	CreateTypeCastWrapper(className string, obj Value) Value
	RaiseTypeCastException(message string, node ast.Node)
	RaiseAssertionFailed(customMessage string)
	CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{}
	CleanupInterfaceReferences(env interface{})

	// ===== Method Pointers =====

	ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value

	// ===== Exception Handling =====

	CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any
	WrapObjectInException(objInstance Value, pos any, callStack any) any

	// ===== Variable Declaration =====

	WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error)
	WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error)

	// ===== Property & Method References =====

	CreateBoundMethodPointer(obj Value, methodDecl any) Value

	// ===== Dispatch Methods =====

	CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) Value
	CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value

	// ===== Class Declaration =====

	NewClassInfoAdapter(name string) interface{}
	CastToClassInfo(class interface{}) (interface{}, bool)
	IsClassPartial(classInfo interface{}) bool
	SetClassPartial(classInfo interface{}, isPartial bool)
	SetClassAbstract(classInfo interface{}, isAbstract bool)
	SetClassExternal(classInfo interface{}, isExternal bool, externalName string)
	ClassHasNoParent(classInfo interface{}) bool
	DefineCurrentClassMarker(env interface{}, classInfo interface{})
	SetClassParent(classInfo interface{}, parentClass interface{})
	AddInterfaceToClass(classInfo interface{}, interfaceInfo interface{}, interfaceName string)
	AddClassMethod(classInfo interface{}, method *ast.FunctionDecl, className string) bool
	SynthesizeDefaultConstructor(classInfo interface{})
	AddClassProperty(classInfo interface{}, propDecl *ast.PropertyDecl) bool
	RegisterClassOperator(classInfo interface{}, opDecl *ast.OperatorDecl) Value
	LookupClassMethod(classInfo interface{}, methodName string, isClassMethod bool) (interface{}, bool)
	SetClassConstructor(classInfo interface{}, constructor interface{})
	SetClassDestructor(classInfo interface{}, destructor interface{})
	InheritDestructorIfMissing(classInfo interface{})
	InheritParentProperties(classInfo interface{})
	BuildVirtualMethodTable(classInfo interface{})
	RegisterClassInTypeSystem(classInfo interface{}, parentName string)

	// ===== Helper Properties =====

	EvalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value

	// ===== Operator Overloading =====

	// TryBinaryOperator attempts to find and invoke a binary operator overload.
	// Returns (result, true) if operator found, or (nil, false) if not found.
	TryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool)

	// TryUnaryOperator attempts to find and invoke a unary operator overload.
	// Returns (result, true) if operator found, or (nil, false) if not found.
	TryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool)
}

// Evaluator evaluates DWScript AST nodes.
// Dependencies: type system, runtime services, configuration.
// Execution state is in ExecutionContext (stateless evaluator).
type Evaluator struct {
	output            io.Writer
	externalFunctions ExternalFunctionRegistry
	currentNode       ast.Node
	adapter           InterpreterAdapter
	refCountMgr       runtime.RefCountManager
	config            *Config
	rand              *rand.Rand
	unitRegistry      *units.UnitRegistry
	initializedUnits  map[string]bool
	semanticInfo      *ast.SemanticInfo
	currentContext    *ExecutionContext
	typeSystem        *interptypes.TypeSystem
	loadedUnits       []string
	randSeed          int64
}

// Ensure Evaluator implements builtins.Context interface.
var _ builtins.Context = (*Evaluator)(nil)

// NewEvaluator creates a new Evaluator with explicit dependency injection.
func NewEvaluator(
	typeSystem *interptypes.TypeSystem,
	output io.Writer,
	config *Config,
	unitRegistry *units.UnitRegistry,
	semanticInfo *ast.SemanticInfo,
	refCountMgr runtime.RefCountManager,
) *Evaluator {
	if config == nil {
		config = DefaultConfig()
	}

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
		refCountMgr:      refCountMgr,
	}
}

// TypeSystem returns the type system instance.
func (e *Evaluator) TypeSystem() *interptypes.TypeSystem {
	return e.typeSystem
}

// FunctionRegistry returns the function registry.
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

// RefCountManager returns the reference counting manager.
func (e *Evaluator) RefCountManager() runtime.RefCountManager {
	return e.refCountMgr
}

// CurrentNode returns the current AST node being evaluated (for error reporting).
func (e *Evaluator) CurrentNode() ast.Node {
	return e.currentNode
}

// SetCurrentNode sets the current AST node being evaluated (for error reporting).
func (e *Evaluator) SetCurrentNode(node ast.Node) {
	e.currentNode = node
}

// SetAdapter sets the interpreter adapter.
func (e *Evaluator) SetAdapter(adapter InterpreterAdapter) {
	e.adapter = adapter
}

// ============================================================================
// Direct Environment Access
// ============================================================================

// GetVar retrieves a variable from the execution context's environment.
func (e *Evaluator) GetVar(ctx *ExecutionContext, name string) (Value, bool) {
	val, found := ctx.Env().Get(name)
	if !found {
		return nil, false
	}
	if v, ok := val.(Value); ok {
		return v, true
	}
	return nil, false
}

// DefineVar defines a new variable in the execution context's environment.
func (e *Evaluator) DefineVar(ctx *ExecutionContext, name string, value Value) {
	ctx.Env().Define(name, value)
}

// SetVar updates an existing variable in the execution context's environment.
func (e *Evaluator) SetVar(ctx *ExecutionContext, name string, value Value) bool {
	return ctx.Env().Set(name, value)
}

// raiseMaxRecursionExceeded raises a max recursion exception.
func (e *Evaluator) raiseMaxRecursionExceeded(node ast.Node) Value {
	return e.newError(node, "maximum recursion depth exceeded")
}

// Eval evaluates an AST node using the visitor pattern.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	e.currentContext = ctx
	defer func() { e.currentContext = nil }()

	ctx.SetEvaluator(e)
	e.currentNode = node

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
	case *ast.EmptyStatement:
		return e.VisitEmptyStatement(n, ctx)
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
		if e.adapter != nil {
			return e.adapter.EvalNode(node)
		}
		panic("Evaluator.Eval: unknown node type and no adapter available")
	}
}

// ============================================================================
// Exception Creation
// ============================================================================

// createException creates an exception with resolved class metadata.
func (e *Evaluator) createException(className, message string, pos *lexer.Position, ctx *ExecutionContext) any {
	excClass := e.typeSystem.LookupClass(className)
	if excClass == nil {
		excClass = e.typeSystem.LookupClass("Exception")
	}
	return e.adapter.CreateExceptionDirect(excClass, message, pos, ctx.CallStack())
}

// wrapObjectAsException wraps an existing ObjectInstance as an exception.
func (e *Evaluator) wrapObjectAsException(obj Value, pos *lexer.Position, ctx *ExecutionContext) any {
	return e.adapter.WrapObjectInException(obj, pos, ctx.CallStack())
}
