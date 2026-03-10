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
	ExternalFunctions      ExternalFunctionRegistry
	RefCountManager        runtime.RefCountManager
	UnitRegistry           *units.UnitRegistry
	InitializedUnits       map[string]bool
	SemanticInfo           *ast.SemanticInfo
	MethodRegistry         *runtime.MethodRegistry
	Random                 *rand.Rand
	ExternalFunctionCaller func(funcName string, argExprs []ast.Expression, node ast.Node) Value
	SourceCode             string
	SourceFile             string
	LoadedUnits            []string
	RandomSeed             int64
	MaxRecursionDepth      int
}

// The old callback-style focused interfaces were removed during Phase 4.
// This package now remains as a small neutral home for cross-package runtime
// state and callback types that still make sense after the split cleanup.

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
