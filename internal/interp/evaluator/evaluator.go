package evaluator

import (
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/contracts"
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
type ClassMetaValue = contracts.ClassMetaValue

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
type FunctionPointerMetadata = contracts.FunctionPointerMetadata

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
type ExternalFunctionRegistry = contracts.ExternalFunctionRegistry

// Evaluator evaluates DWScript AST nodes.
// Dependencies: type system, runtime services, configuration.
// Execution state is in ExecutionContext (stateless evaluator).
type Evaluator struct {
	output            io.Writer
	currentNode       ast.Node
	oopEngine         OOPEngine   // Runtime OOP operations
	coreEvaluator     CoreEvaluator
	config            *Config
	currentContext    *ExecutionContext
	typeSystem        *interptypes.TypeSystem
	engineState       *contracts.EngineState
	selfContainedMode bool // When true, Eval() won't fall back to coreEvaluator
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
	state := &contracts.EngineState{
		SourceCode:        config.SourceCode,
		SourceFile:        config.SourceFile,
		ExternalFunctions: nil,
		UnitRegistry:      unitRegistry,
		InitializedUnits:  make(map[string]bool),
		SemanticInfo:      semanticInfo,
		RefCountManager:   refCountMgr,
		MethodRegistry:    runtime.NewMethodRegistry(),
		Random:            rand.New(source),
		LoadedUnits:       make([]string, 0),
		RandomSeed:        defaultSeed,
	}

	return &Evaluator{
		typeSystem:  typeSystem,
		output:      output,
		config:      config,
		engineState: state,
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
	return e.engineState.Random
}

// RandomSeed returns the current random seed.
func (e *Evaluator) RandomSeed() int64 {
	return e.engineState.RandomSeed
}

// SetRandomSeed sets the random seed and reinitializes the generator.
func (e *Evaluator) SetRandomSeed(seed int64) {
	e.engineState.RandomSeed = seed
	source := rand.NewSource(seed)
	e.engineState.Random = rand.New(source)
}

// ExternalFunctions returns the external function registry.
func (e *Evaluator) ExternalFunctions() ExternalFunctionRegistry {
	return e.engineState.ExternalFunctions
}

// SetExternalFunctions sets the external function registry.
func (e *Evaluator) SetExternalFunctions(reg ExternalFunctionRegistry) {
	e.engineState.ExternalFunctions = reg
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
	return e.engineState.SourceCode
}

// SourceFile returns the source file path.
func (e *Evaluator) SourceFile() string {
	return e.engineState.SourceFile
}

// SetSource sets the source code and filename for enhanced error messages.
func (e *Evaluator) SetSource(source, filename string) {
	e.engineState.SourceCode = source
	e.engineState.SourceFile = filename
}

// UnitRegistry returns the unit registry.
func (e *Evaluator) UnitRegistry() *units.UnitRegistry {
	return e.engineState.UnitRegistry
}

// SetUnitRegistry sets the unit registry.
func (e *Evaluator) SetUnitRegistry(registry *units.UnitRegistry) {
	e.engineState.UnitRegistry = registry
}

// InitializedUnits returns the map of initialized units.
func (e *Evaluator) InitializedUnits() map[string]bool {
	return e.engineState.InitializedUnits
}

// LoadedUnits returns the list of loaded units.
func (e *Evaluator) LoadedUnits() []string {
	return e.engineState.LoadedUnits
}

// AddLoadedUnit adds a unit to the list of loaded units.
func (e *Evaluator) AddLoadedUnit(unitName string) {
	e.engineState.LoadedUnits = append(e.engineState.LoadedUnits, unitName)
}

// SemanticInfo returns the semantic analysis metadata.
func (e *Evaluator) SemanticInfo() *ast.SemanticInfo {
	return e.engineState.SemanticInfo
}

// SetSemanticInfo sets the semantic analysis metadata.
func (e *Evaluator) SetSemanticInfo(info *ast.SemanticInfo) {
	e.engineState.SemanticInfo = info
}

// RefCountManager returns the reference counting manager.
func (e *Evaluator) RefCountManager() runtime.RefCountManager {
	return e.engineState.RefCountManager
}

// CurrentNode returns the current AST node being evaluated (for error reporting).
func (e *Evaluator) CurrentNode() ast.Node {
	return e.currentNode
}

// CurrentContext returns the active execution context for the current evaluation.
func (e *Evaluator) CurrentContext() *ExecutionContext {
	return e.currentContext
}

// EngineState returns the shared runtime engine state.
func (e *Evaluator) EngineState() *contracts.EngineState {
	return e.engineState
}

// SetCurrentNode sets the current AST node being evaluated (for error reporting).
func (e *Evaluator) SetCurrentNode(node ast.Node) {
	e.currentNode = node
}

// SetRuntimeBridge wires the remaining runtime bridges the evaluator still needs
// during the Phase 4 collapse. Production construction should use this narrower
// bridge instead of SetFocusedInterfaces.
func (e *Evaluator) SetRuntimeBridge(oopEngine OOPEngine, coreEvaluator CoreEvaluator) {
	e.oopEngine = oopEngine
	e.coreEvaluator = coreEvaluator
}

// SetFocusedInterfaces is a compatibility/testing shim for the legacy Phase 4
// bridge API. The declaration handler argument is ignored because declaration
// processing no longer depends on that callback seam.
func (e *Evaluator) SetFocusedInterfaces(
	oopEngine OOPEngine,
	_ DeclHandler,
	coreEvaluator CoreEvaluator,
) {
	e.SetRuntimeBridge(oopEngine, coreEvaluator)
}

// EnterSelfContainedMode temporarily disables interpreter fallbacks in Eval().
// Use this when evaluation must stay entirely inside evaluator-owned visitors.
// Returns a restore function that MUST be called (typically via defer).
//
// When selfContainedMode is true, Eval() handles all AST nodes directly
// without falling back to the interpreter.
//
// Note: coreEvaluator remains available for direct calls from visitor methods
// (e.g., method_dispatch.go, helper_methods.go) - only Eval() skips fallbacks.
func (e *Evaluator) EnterSelfContainedMode() func() {
	wasSelfContained := e.selfContainedMode
	e.selfContainedMode = true

	return func() {
		e.selfContainedMode = wasSelfContained
	}
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
	return ctx.Env().Set(name, value) == nil
}

// Eval evaluates an AST node using the visitor pattern.
func (e *Evaluator) Eval(node ast.Node, ctx *ExecutionContext) Value {
	previousContext := e.currentContext
	e.currentContext = ctx
	defer func() { e.currentContext = previousContext }()

	previousNode := e.currentNode
	e.currentNode = node
	defer func() { e.currentNode = previousNode }()

	if ctx != nil && e.engineState != nil {
		ctx.SetRefCountManager(e.engineState.RefCountManager)
	}

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
		if e.coreEvaluator != nil && !e.selfContainedMode {
			return e.coreEvaluator.EvalNode(n, ctx)
		}
		return e.VisitFunctionDecl(n, ctx)
	case *ast.ClassDecl:
		if e.coreEvaluator != nil && !e.selfContainedMode {
			return e.coreEvaluator.EvalNode(n, ctx)
		}
		return e.VisitClassDecl(n, ctx)
	case *ast.InterfaceDecl:
		if e.coreEvaluator != nil && !e.selfContainedMode {
			return e.coreEvaluator.EvalNode(n, ctx)
		}
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
		if e.coreEvaluator != nil && !e.selfContainedMode {
			return e.coreEvaluator.EvalNode(n, ctx)
		}
		return e.VisitHelperDecl(n, ctx)
	case *ast.ArrayDecl:
		return e.VisitArrayDecl(n, ctx)
	case *ast.TypeDeclaration:
		return e.VisitTypeDeclaration(n, ctx)

	default:
		if e.coreEvaluator != nil && !e.selfContainedMode {
			return e.coreEvaluator.EvalNode(node, ctx)
		}
		panic("Evaluator.Eval: unknown node type and no coreEvaluator available")
	}
}

// ============================================================================
// Exception Creation
// ============================================================================

// createException creates an exception with resolved class metadata.
// Self-contained: no longer delegates to ExceptionManager.
func (e *Evaluator) createException(className, message string, pos *lexer.Position, ctx *ExecutionContext) any {
	excClass := e.typeSystem.LookupClass(className)
	if excClass == nil {
		excClass = e.typeSystem.LookupClass("Exception")
	}

	// Get metadata and create instance
	var metadata *runtime.ClassMetadata
	var instance *runtime.ObjectInstance
	if excClass != nil {
		// Type assert to IClassInfo to access GetMetadata()
		if classInfo, ok := excClass.(runtime.IClassInfo); ok {
			metadata = classInfo.GetMetadata()
			instance = runtime.NewObjectInstance(classInfo)
			instance.SetField("Message", &runtime.StringValue{Value: message})
		}
	}

	return runtime.NewException(metadata, instance, message, pos, ctx.CallStack())
}

// wrapObjectAsException wraps an existing ObjectInstance as an exception.
// Self-contained: no longer delegates to ExceptionManager.
func (e *Evaluator) wrapObjectAsException(obj Value, pos *lexer.Position, ctx *ExecutionContext) any {
	// Cast to ObjectInstance
	objInst, ok := obj.(*runtime.ObjectInstance)
	if !ok {
		// Create a simple exception if not an ObjectInstance
		return runtime.NewException(nil, nil, "Invalid exception object", pos, ctx.CallStack())
	}

	// Extract message from the object's Message field
	message := ""
	if msgVal := objInst.GetField("Message"); msgVal != nil {
		if strVal, ok := msgVal.(*runtime.StringValue); ok {
			message = strVal.Value
		}
	}

	return runtime.NewExceptionFromObject(objInst, message, pos, ctx.CallStack())
}

// cleanupInterfaceReferences releases all interface and object references when a scope ends.
// Self-contained: no longer delegates to ExceptionManager.
func (e *Evaluator) cleanupInterfaceReferences(env *runtime.Environment) {
	if env == nil || e.engineState.RefCountManager == nil {
		return
	}

	// Iterate through all variables in the environment
	env.Range(func(name string, value Value) bool {
		// Skip ReferenceValue entries (like function name aliases)
		if _, isRef := value.(*runtime.ReferenceValue); isRef {
			return true // continue
		}

		// Skip "Result" variable during function cleanup.
		// Result is the return value and will be managed by the caller.
		// Releasing it here would cause premature destructor calls.
		if name == "Result" {
			return true // continue (skip)
		}

		// Release interface references
		if intfInst, ok := value.(*runtime.InterfaceInstance); ok {
			e.engineState.RefCountManager.ReleaseInterface(intfInst)
		} else if objInst, ok := value.(*runtime.ObjectInstance); ok {
			// Release object references
			e.engineState.RefCountManager.ReleaseObject(objInst)
		} else if funcPtr, ok := value.(*runtime.FunctionPointerValue); ok {
			// Clean up method pointers that hold object references
			if objInst, isObj := funcPtr.SelfObject.(*runtime.ObjectInstance); isObj {
				e.engineState.RefCountManager.ReleaseObject(objInst)
			}
		}
		return true // continue
	})
}
