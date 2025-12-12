package contracts

import (
	"math/rand"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
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

// Evaluator is the minimal API the interpreter needs from the evaluator.
// This interface exists to avoid `internal/interp` importing `internal/interp/evaluator`.
type Evaluator interface {
	Eval(node ast.Node, ctx *runtime.ExecutionContext) Value
	ExecuteUserFunction(fn *ast.FunctionDecl, args []Value, ctx *runtime.ExecutionContext, callbacks *UserFunctionCallbacks) (Value, error)
	CurrentNode() ast.Node
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
	EvalNode(node ast.Node) Value

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
	CallUserFunction(fn *ast.FunctionDecl, args []Value) Value
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

// ExceptionManager handles exception creation, propagation, and cleanup.
type ExceptionManager interface {
	CreateExceptionDirect(classMetadata any, message string, pos any, callStack any) any
	WrapObjectInException(objInstance Value, pos any, callStack any) any
	CreateContractException(className, message string, node ast.Node, classMetadata interface{}, callStack interface{}) interface{}
	RaiseTypeCastException(message string, node ast.Node)
	RaiseAssertionFailed(customMessage string)
	CleanupInterfaceReferences(env interface{})
}

// DeclHandler handles type declaration processing (classes, interfaces, helpers).
type DeclHandler interface {
	NewClassInfoAdapter(name string) any
	CastToClassInfo(class any) (any, bool)
	IsClassPartial(classInfo any) bool
	SetClassPartial(classInfo any, isPartial bool)
	SetClassAbstract(classInfo any, isAbstract bool)
	SetClassExternal(classInfo any, isExternal bool, externalName string)
	ClassHasNoParent(classInfo any) bool
	DefineCurrentClassMarker(env any, classInfo any)
	SetClassParent(classInfo any, parentClass any)
	AddInterfaceToClass(classInfo any, interfaceInfo any, interfaceName string)
	AddClassMethod(classInfo any, method *ast.FunctionDecl, className string) bool
	SynthesizeDefaultConstructor(classInfo any)
	AddClassProperty(classInfo any, propDecl *ast.PropertyDecl) bool
	RegisterClassOperator(classInfo any, opDecl *ast.OperatorDecl) Value
	LookupClassMethod(classInfo any, methodName string, isClassMethod bool) (any, bool)
	SetClassConstructor(classInfo any, constructor any)
	SetClassDestructor(classInfo any, destructor any)
	InheritDestructorIfMissing(classInfo any)
	InheritParentProperties(classInfo any)
	BuildVirtualMethodTable(classInfo any)
	RegisterClassInTypeSystem(classInfo any, parentName string)
	AddClassConstant(classInfo any, constDecl *ast.ConstDecl, value Value)
	GetClassConstantValues(classInfo any) map[string]Value
	InheritClassConstants(classInfo any, parentClass any)
	AddClassField(classInfo any, fieldDecl *ast.FieldDecl, fieldType types.Type)
	AddClassVar(classInfo any, name string, value Value)
	AddNestedClass(parentClass any, nestedName string, nestedClass any)

	NewInterfaceInfoAdapter(name string) any
	CastToInterfaceInfo(iface any) (any, bool)
	SetInterfaceParent(iface any, parent any)
	GetInterfaceName(iface any) string
	GetInterfaceParent(iface any) any
	AddInterfaceMethod(iface any, normalizedName string, method *ast.FunctionDecl)
	AddInterfaceProperty(iface any, normalizedName string, propInfo any)

	CreateHelperInfo(name string, targetType any, isRecordHelper bool) any
	SetHelperParent(helper any, parent any)
	VerifyHelperTargetTypeMatch(parent any, targetType any) bool
	GetHelperName(helper any) string
	AddHelperMethod(helper any, normalizedName string, method *ast.FunctionDecl)
	AddHelperProperty(helper any, prop *ast.PropertyDecl, propType any)
	AddHelperClassVar(helper any, name string, value Value)
	AddHelperClassConst(helper any, name string, value Value)
	RegisterHelperLegacy(typeName string, helper any)

	EvalMethodImplementation(fn *ast.FunctionDecl) Value
}

// User function execution callbacks.
type ImplicitConversionFunc func(value Value, targetTypeName string) (Value, bool)
type DefaultValueFunc func(returnTypeName string) Value
type FunctionNameAliasFunc func(funcName string, funcEnv *runtime.Environment) Value
type CleanupInterfaceReferencesFunc func(env *runtime.Environment)
type TryImplicitConversionReturnFunc func(returnValue Value, expectedReturnType string) (Value, bool)
type IncrementInterfaceRefCountFunc func(returnValue Value)
type EnvSyncerFunc func(funcEnv *runtime.Environment) func()

type UserFunctionCallbacks struct {
	ImplicitConversion   ImplicitConversionFunc
	DefaultValueGetter   DefaultValueFunc
	FunctionNameAlias    FunctionNameAliasFunc
	ReturnValueConverter TryImplicitConversionReturnFunc
	InterfaceRefCounter  IncrementInterfaceRefCountFunc
	InterfaceCleanup     CleanupInterfaceReferencesFunc
	EnvSyncer            EnvSyncerFunc
}
