package contracts

import (
	"math/rand"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Value is the shared value interface used across interpreter/evaluator.
// It is intentionally anchored in runtime to avoid circular dependencies.
type Value = runtime.Value

// ClassMetaValue provides access to class metadata (class references).
// Moved to a neutral package to avoid interpreter depending on evaluator types.
type ClassMetaValue interface {
	Value
	GetClassName() string
	GetClassType() Value
	GetClassVar(name string) (Value, bool)
	GetClassConstant(name string) (Value, bool)
	HasClassMethod(name string) bool
	HasConstructor(name string) bool
	InvokeParameterlessClassMethod(name string, executor func(methodDecl any) Value) (Value, bool)
	CreateClassMethodPointer(name string, creator func(methodDecl any) Value) (Value, bool)
	InvokeConstructor(name string, executor func(methodDecl any) Value) (Value, bool)
	GetNestedClass(name string) Value
	ReadClassProperty(name string, executor func(propInfo any) Value) (Value, bool)
	GetClassInfo() any
	SetClassVar(name string, value Value) bool
	WriteClassProperty(name string, value Value, executor func(propInfo any, value Value) Value) (Value, bool)
	HasClassVar(name string) bool
}

// FunctionPointerMetadata provides execution context for function pointer invocation.
// Kept in a neutral package so both evaluator and interpreter can reference it
// without importing each other.
type FunctionPointerMetadata struct {
	Lambda     any
	Function   any
	Closure    any
	SelfObject Value
	IsLambda   bool
}

// ExternalFunctionRegistry manages external Go functions callable from DWScript.
type ExternalFunctionRegistry interface {
	Has(name string) bool
}

// EngineState holds interpreter-runtime state that must not be owned by both
// interpreter and evaluator independently.
type EngineState struct {
	SourceCode        string
	SourceFile        string
	ExternalFunctions ExternalFunctionRegistry
	UnitRegistry      *units.UnitRegistry
	InitializedUnits  map[string]bool
	SemanticInfo      *ast.SemanticInfo
	RefCountManager   runtime.RefCountManager
	MethodRegistry    *runtime.MethodRegistry
	Random            *rand.Rand
	LoadedUnits       []string
	RandomSeed        int64
}

// Evaluator is the minimal API the interpreter needs from the evaluator.
// This interface exists to avoid `internal/interp` importing `internal/interp/evaluator`.
type Evaluator interface {
	Eval(node ast.Node, ctx *runtime.ExecutionContext) Value
	ExecuteUserFunction(fn *ast.FunctionDecl, args []Value, ctx *runtime.ExecutionContext, callbacks *UserFunctionCallbacks) (Value, error)
	CurrentNode() ast.Node
	CurrentContext() *runtime.ExecutionContext
	EngineState() *EngineState
	SetCurrentNode(node ast.Node)
	SemanticInfo() *ast.SemanticInfo
	SetSemanticInfo(info *ast.SemanticInfo)
	SetSource(source, filename string)
	SourceCode() string
	SourceFile() string
	UnitRegistry() *units.UnitRegistry
	SetUnitRegistry(registry *units.UnitRegistry)
	LoadedUnits() []string
	AddLoadedUnit(unitName string)
	InitializedUnits() map[string]bool
	ExternalFunctions() ExternalFunctionRegistry
	SetExternalFunctions(reg ExternalFunctionRegistry)
	RefCountManager() runtime.RefCountManager
	Random() *rand.Rand
	RandomSeed() int64
	SetRandomSeed(seed int64)
}

// CoreEvaluator provides fallback evaluation for cross-cutting concerns.
// May be eliminated in future by migrating remaining OOP logic to evaluator.
type CoreEvaluator interface {
	// EvalNode evaluates an AST node via interpreter for OOP operations.
	// Fallback for operations not yet migrated to evaluator.
	// The ctx parameter ensures the interpreter uses the correct environment
	// (e.g., when callbacks occur from within a for loop that pushed a new scope).
	EvalNode(node ast.Node, ctx *runtime.ExecutionContext) Value

	// EvalBuiltinHelperProperty evaluates a built-in helper property (Length, Low, High, etc).
	EvalBuiltinHelperProperty(propSpec string, selfValue Value, node ast.Node) Value

	// EvalClassPropertyRead evaluates a class property read (static properties).
	EvalClassPropertyRead(classInfo any, propInfo any, node ast.Node) Value

	// EvalClassPropertyWrite evaluates a class property write (static properties).
	EvalClassPropertyWrite(classInfo any, propInfo any, value Value, node ast.Node) Value
}

// OOPEngine handles runtime object-oriented programming operations.
// Encapsulates method dispatch, constructors, type operations, and operator overloading.
type OOPEngine interface {
	CallMethod(obj Value, methodName string, args []Value, node ast.Node) Value
	CallInheritedMethod(obj Value, methodName string, args []Value) Value
	ExecuteMethodWithSelf(self Value, methodDecl any, args []Value) Value
	CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	ExecuteConstructor(obj Value, constructorName string, args []Value) error
	CallFunctionPointer(funcPtr Value, args []Value, node ast.Node) Value
	ExecuteFunctionPointerCall(metadata FunctionPointerMetadata, args []Value, node ast.Node) Value
	CreateBoundMethodPointer(obj Value, methodDecl any) Value
	CreateTypeCastWrapper(className string, obj Value) Value
	WrapInSubrange(value Value, subrangeTypeName string, node ast.Node) (Value, error)
	WrapInInterface(value Value, interfaceName string, node ast.Node) (Value, error)
	CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) Value
	CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	DispatchRecordStaticMethod(recordTypeName string, callExpr *ast.CallExpression, funcName *ast.Identifier) Value
	ExecuteRecordPropertyRead(record Value, propInfo any, indices []Value, node any) Value
	CallExternalFunction(funcName string, argExprs []ast.Expression, node ast.Node) Value
	TryBinaryOperator(operator string, left, right Value, node ast.Node) (Value, bool)
	TryUnaryOperator(operator string, operand Value, node ast.Node) (Value, bool)
	LookupClassByName(name string) ClassMetaValue
}

// ExceptionManager was removed in Task 4.2.
// Exception handling is now self-contained in the evaluator using:
// - runtime.NewException() for creating exceptions
// - runtime.NewExceptionFromObject() for wrapping objects as exceptions
// - RefCountManager.ReleaseInterface/ReleaseObject for cleanup

// DeclHandler handles type declaration processing (classes, interfaces, helpers).
type DeclHandler interface {
	NewClassInfoAdapter(name string) any
}

// User function execution callbacks.
type ImplicitConversionFunc func(value Value, targetTypeName string) (Value, bool)
type DefaultValueFunc func(returnTypeName string) Value
type FunctionNameAliasFunc func(funcName string, funcEnv *runtime.Environment) Value
type CleanupInterfaceReferencesFunc func(env *runtime.Environment)
type TryImplicitConversionReturnFunc func(returnValue Value, expectedReturnType string) (Value, bool)
type IncrementInterfaceRefCountFunc func(returnValue Value)

type UserFunctionCallbacks struct {
	ImplicitConversion   ImplicitConversionFunc
	DefaultValueGetter   DefaultValueFunc
	FunctionNameAlias    FunctionNameAliasFunc
	ReturnValueConverter TryImplicitConversionReturnFunc
	InterfaceRefCounter  IncrementInterfaceRefCountFunc
	InterfaceCleanup     CleanupInterfaceReferencesFunc
}
