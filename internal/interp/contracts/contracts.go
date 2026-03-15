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
	GetClassInfo() runtime.IClassInfo
	SetClassVar(name string, value Value) bool
	WriteClassProperty(name string, value Value, executor func(propInfo any, value Value) Value) (Value, bool)
	HasClassVar(name string) bool
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
// This package now remains as a small neutral home for shared engine state and
// a minimal set of cross-package coordination types only.
