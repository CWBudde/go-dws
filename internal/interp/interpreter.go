package interp

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"

	// Task 3.8.2: pkg/ast is imported for SemanticInfo, which holds semantic analysis
	// metadata (type annotations, symbol resolutions). This is separate from the AST
	// structure itself and is not aliased in internal/ast.
	// Task 9.18: Separate type metadata from AST nodes.
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
)

// DefaultMaxRecursionDepth is the default maximum recursion depth for function calls.
// This matches DWScript's default limit (see dwsCompiler.pas:39 cDefaultMaxRecursionDepth).
// When the call stack reaches this depth, the interpreter raises an EScriptStackOverflow exception
// to prevent infinite recursion and potential Go runtime stack overflow.
const DefaultMaxRecursionDepth = 1024

// PropertyEvalContext tracks the state during property getter/setter evaluation.
// Deprecated: Use evaluator.PropertyEvalContext instead. This is kept for backward compatibility.
type PropertyEvalContext = evaluator.PropertyEvalContext

// Interpreter executes DWScript AST nodes and manages the runtime environment.
//
// Phase 3.5.1: The Interpreter is being refactored to be a thin orchestrator.
// The evaluator field contains the evaluation logic and dependencies.
// Eventually, most of the fields below will be removed and accessed via the evaluator.
type Interpreter struct {
	currentNode          ast.Node
	output               io.Writer
	helpers              map[string][]*HelperInfo
	enumTypeIDRegistry   map[string]int
	exception            *ExceptionValue
	handlerException     *ExceptionValue
	semanticInfo         *pkgast.SemanticInfo
	unitRegistry         *units.UnitRegistry
	propContext          *PropertyEvalContext
	typeSystem           *interptypes.TypeSystem
	methodRegistry       *runtime.MethodRegistry // Task 3.5.39: AST-free method storage
	recordTypeIDRegistry map[string]int
	records              map[string]*RecordTypeValue
	interfaces           map[string]*InterfaceInfo
	functions            map[string][]*ast.FunctionDecl
	globalOperators      *runtimeOperatorRegistry
	conversions          *runtimeConversionRegistry
	env                  *Environment
	evaluatorInstance    *evaluator.Evaluator
	classes              map[string]*ClassInfo
	classTypeIDRegistry  map[string]int
	initializedUnits     map[string]bool
	externalFunctions    *ExternalFunctionRegistry
	rand                 *rand.Rand
	ctx                  *evaluator.ExecutionContext
	sourceCode           string
	sourceFile           string
	oldValuesStack       []map[string]Value
	loadedUnits          []string
	callStack            errors.StackTrace
	nextEnumTypeID       int
	randSeed             int64
	nextRecordTypeID     int
	maxRecursionDepth    int
	nextClassTypeID      int
}

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions creates a new Interpreter with options.
// If options is nil, default options are used.
// Task 3.8.1: Uses Options interface to avoid circular dependency and remove reflection hack.
func NewWithOptions(output io.Writer, opts Options) *Interpreter {
	env := NewEnvironment()
	// Initialize random number generator with a default seed
	// Randomize() can be called to re-seed with current time
	const defaultSeed = int64(1)
	source := rand.NewSource(defaultSeed)

	// Phase 3.4.1: Initialize TypeSystem
	// The TypeSystem is the new centralized type registry that manages all type information
	// including classes, records, interfaces, functions, helpers, operators, and conversions.
	//
	// Migration Strategy (Gradual Transition):
	// - The old fields (functions, classes, records, etc.) are kept for backward compatibility
	// - Existing code continues to work unchanged during the transition period
	// - New code should use typeSystem methods (e.g., typeSystem.RegisterClass, typeSystem.LookupClass)
	// - Old code will be gradually refactored to use typeSystem in future tasks
	// - Once migration is complete, the old fields will be removed (future Phase 4+ work)
	ts := interptypes.NewTypeSystem()

	interp := &Interpreter{
		env:               env,
		output:            output,
		rand:              rand.New(source),
		randSeed:          defaultSeed,
		loadedUnits:       make([]string, 0),
		initializedUnits:  make(map[string]bool),
		maxRecursionDepth: DefaultMaxRecursionDepth,
		callStack:         errors.NewStackTrace(), // Initialize stack trace

		// Phase 3.4.1: TypeSystem (new centralized type registry)
		// This is the modern API - use this for new code
		typeSystem: ts,

		// Task 3.5.39: MethodRegistry for AST-free method storage
		methodRegistry: runtime.NewMethodRegistry(),

		// Phase 3.4.1: Legacy fields for backward compatibility
		// These will be removed once migration to typeSystem is complete
		functions:            make(map[string][]*ast.FunctionDecl), // Task 9.66: Support overloading
		classes:              make(map[string]*ClassInfo),
		records:              make(map[string]*RecordTypeValue),
		interfaces:           make(map[string]*InterfaceInfo),
		globalOperators:      newRuntimeOperatorRegistry(),
		conversions:          newRuntimeConversionRegistry(),
		helpers:              make(map[string][]*HelperInfo),
		classTypeIDRegistry:  make(map[string]int), // Task 9.25: RTTI type ID registry
		recordTypeIDRegistry: make(map[string]int), // Task 9.25: RTTI type ID registry
		enumTypeIDRegistry:   make(map[string]int), // Task 9.25: RTTI type ID registry
		nextClassTypeID:      1000,                 // Task 9.25: Start class IDs at 1000
		nextRecordTypeID:     200000,               // Task 9.25: Start record IDs at 200000
		nextEnumTypeID:       300000,               // Task 9.25: Start enum IDs at 300000
	}

	// Task 3.8.1: Extract external functions and recursion depth from options using interface
	// This replaces the reflection hack with a clean interface-based approach
	if opts != nil {
		// Extract ExternalFunctions
		if registry := opts.GetExternalFunctions(); registry != nil {
			interp.externalFunctions = registry
		}

		// Extract MaxRecursionDepth
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			interp.maxRecursionDepth = depth
		}
	}

	// Ensure we have a registry even if not provided
	if interp.externalFunctions == nil {
		interp.externalFunctions = NewExternalFunctionRegistry()
	}

	// Phase 3.3.1/3.3.3: Initialize execution context with call stack overflow detection
	// The context wraps the environment with a simple adapter to avoid circular dependencies.
	// The Environment is passed as interface{} to ExecutionContext to avoid import cycles.
	// Phase 3.3.3: Pass maxRecursionDepth to configure CallStack overflow detection.
	interp.ctx = evaluator.NewExecutionContextWithMaxDepth(
		evaluator.NewEnvironmentAdapter(env),
		interp.maxRecursionDepth,
	)

	// Phase 3.5.1: Initialize Evaluator
	// The Evaluator holds evaluation logic and dependencies (type system, runtime services, config)
	// Note: unitRegistry can be nil initially - it's set via SetUnitRegistry if needed

	// Create evaluator config
	evalConfig := &evaluator.Config{
		MaxRecursionDepth: interp.maxRecursionDepth,
		SourceCode:        "",
		SourceFile:        "",
	}

	// Create evaluator instance
	// Task 3.5.76: semanticInfo passed via constructor for explicit dependency injection
	interp.evaluatorInstance = evaluator.NewEvaluator(
		ts,
		output,
		evalConfig,
		nil,                 // unitRegistry is set later via SetUnitRegistry if needed
		interp.semanticInfo, // Task 3.5.76: pass semanticInfo to constructor
	)

	// Set external functions if available
	if interp.externalFunctions != nil {
		interp.evaluatorInstance.SetExternalFunctions(interp.externalFunctions)
	}

	// Set the adapter so the evaluator can delegate back to the interpreter during migration
	interp.evaluatorInstance.SetAdapter(interp)

	// Register built-in exception classes
	interp.registerBuiltinExceptions()

	// Register built-in interfaces
	interp.registerBuiltinInterfaces()

	// Register built-in array helpers
	interp.initArrayHelpers()

	// Register built-in helpers for primitive types
	interp.initIntrinsicHelpers()

	// Register built-in enum helpers
	interp.initEnumHelpers()

	// Initialize ExceptObject to nil
	// ExceptObject is a built-in global variable that holds the current exception
	env.Define("ExceptObject", &NilValue{})

	// Register built-in type meta-values
	// These allow type names to be used as runtime values, e.g., High(Integer)
	env.Define("Integer", NewTypeMetaValue(types.INTEGER, "Integer"))
	env.Define("Float", NewTypeMetaValue(types.FLOAT, "Float"))
	env.Define("String", NewTypeMetaValue(types.STRING, "String"))
	env.Define("Boolean", NewTypeMetaValue(types.BOOLEAN, "Boolean"))

	// Register mathematical constants
	env.Define("PI", &FloatValue{Value: math.Pi})
	env.Define("NaN", &FloatValue{Value: math.NaN()})
	env.Define("Infinity", &FloatValue{Value: math.Inf(1)})

	// Task 9.4.1: Register Variant special values
	env.Define("Null", NewNullValue())
	env.Define("Unassigned", NewUnassignedValue())

	return interp
}

// GetException returns the current active exception, or nil if none.
// This is used by the CLI to detect and report unhandled exceptions.
func (i *Interpreter) GetException() *ExceptionValue {
	return i.exception
}

// SetSemanticInfo sets the semantic metadata table for this interpreter.
// The semantic info contains type annotations and symbol resolutions from analysis.
// Task 9.18: Separate type metadata from AST nodes.
func (i *Interpreter) SetSemanticInfo(info *pkgast.SemanticInfo) {
	i.semanticInfo = info

	// Phase 3.5.1: Also update the evaluator's semantic info
	if i.evaluatorInstance != nil {
		i.evaluatorInstance.SetSemanticInfo(info)
	}
}

// GetEvaluator returns the evaluator instance.
// Phase 3.5.1: This provides access to the evaluation engine for advanced use cases.
func (i *Interpreter) GetEvaluator() *evaluator.Evaluator {
	return i.evaluatorInstance
}

// EvalNode implements the evaluator.InterpreterAdapter interface.
// This allows the Evaluator to delegate back to the Interpreter during migration.
// Phase 3.5.1: This is temporary and will be removed once all evaluation logic
// is moved to the Evaluator.
func (i *Interpreter) EvalNode(node ast.Node) evaluator.Value {
	// Delegate to the legacy Eval method
	// The cast is safe because our Value type matches evaluator.Value interface
	return i.Eval(node)
}

// Phase 3.5.4 - Phase 2A: Function call system adapter methods
// These methods implement the InterpreterAdapter interface for function calls.

// convertEvaluatorArgs converts a slice of evaluator.Value to interp.Value.
// This is used by adapter methods when delegating to internal functions.
func convertEvaluatorArgs(args []evaluator.Value) []Value {
	interpArgs := make([]Value, len(args))
	for idx, arg := range args {
		interpArgs[idx] = arg
	}
	return interpArgs
}

// CallFunctionPointer executes a function pointer with given arguments.
func (i *Interpreter) CallFunctionPointer(funcPtr evaluator.Value, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert evaluator.Value to interp.Value (they're the same interface)
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return i.newErrorWithLocation(node, "invalid function pointer type: expected FunctionPointerValue, got %T", funcPtr)
	}

	return i.callFunctionPointer(fp, convertEvaluatorArgs(args), node)
}

// CallUserFunction executes a user-defined function.
func (i *Interpreter) CallUserFunction(fn *ast.FunctionDecl, args []evaluator.Value) evaluator.Value {
	return i.callUserFunction(fn, convertEvaluatorArgs(args))
}

// CallBuiltinFunction executes a built-in function by name.
func (i *Interpreter) CallBuiltinFunction(name string, args []evaluator.Value) evaluator.Value {
	return i.callBuiltinFunction(name, convertEvaluatorArgs(args))
}

// LookupFunction finds a function by name in the function registry.
func (i *Interpreter) LookupFunction(name string) ([]*ast.FunctionDecl, bool) {
	// DWScript is case-insensitive, so normalize to lowercase
	normalizedName := ident.Normalize(name)
	functions, ok := i.functions[normalizedName]
	return functions, ok
}

// ===== Task 3.5.96: Method and Qualified Call Methods =====

// CallMemberMethod calls a method on an object (record, interface, or object instance).
// Task 3.5.96: Enables evaluator to delegate member method calls without using EvalNode.
func (i *Interpreter) CallMemberMethod(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression, objVal evaluator.Value) evaluator.Value {
	// This method encapsulates the complex logic from evalCallExpression lines 82-120

	// Check if this is a record method call
	if recVal, ok := objVal.(*RecordValue); ok {
		return i.evalRecordMethodCall(recVal, memberAccess, callExpr.Arguments, memberAccess.Object)
	}

	// Check if this is an interface method call
	if ifaceInst, ok := objVal.(*InterfaceInstance); ok {
		// Dispatch to the underlying object
		if ifaceInst.Object == nil {
			return i.newErrorWithLocation(callExpr, "cannot call method on nil interface")
		}
		// Call the method on the underlying object by temporarily swapping the variable
		if objIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
			savedVal, exists := i.env.Get(objIdent.Value)
			if exists {
				// Temporarily set to underlying object
				_ = i.env.Set(objIdent.Value, ifaceInst.Object)
				// Use defer to ensure restoration even if method call panics or returns early
				defer func() { _ = i.env.Set(objIdent.Value, savedVal) }()

				// Create a method call expression
				mc := &ast.MethodCallExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: callExpr.Token,
						},
					},
					Object:    memberAccess.Object,
					Method:    memberAccess.Member,
					Arguments: callExpr.Arguments,
				}
				return i.evalMethodCall(mc)
			}
		}
		return i.newErrorWithLocation(callExpr, "interface method call requires identifier")
	}

	// If it's an object, create a MethodCallExpression and evaluate it
	// This handles regular object method calls
	if objVal.Type() == "OBJECT" {
		mc := &ast.MethodCallExpression{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: callExpr.Token,
				},
			},
			Object:    memberAccess.Object,
			Method:    memberAccess.Member,
			Arguments: callExpr.Arguments,
		}
		return i.evalMethodCall(mc)
	}

	return i.newErrorWithLocation(callExpr, "cannot call method on value of type %s", objVal.Type())
}

// CallQualifiedOrConstructor calls a unit-qualified function or class constructor.
// Task 3.5.96: Enables evaluator to delegate qualified calls without using EvalNode.
func (i *Interpreter) CallQualifiedOrConstructor(callExpr *ast.CallExpression, memberAccess *ast.MemberAccessExpression) evaluator.Value {
	// This method encapsulates the complex logic from evalCallExpression lines 122-201

	// Check if the left side is a unit identifier (for qualified access: UnitName.FunctionName)
	if unitIdent, ok := memberAccess.Object.(*ast.Identifier); ok {
		// This could be a unit-qualified call: UnitName.FunctionName()
		if i.unitRegistry != nil {
			if _, exists := i.unitRegistry.GetUnit(unitIdent.Value); exists {
				// Resolve the qualified function
				fn, err := i.ResolveQualifiedFunction(unitIdent.Value, memberAccess.Member.Value)
				if err == nil {
					// Prepare arguments - lazy parameters get LazyThunks, var parameters get References
					args := make([]Value, len(callExpr.Arguments))
					for idx, arg := range callExpr.Arguments {
						// Check parameter flags
						isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
						isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

						if isLazy {
							// For lazy parameters, create a LazyThunk
							args[idx] = NewLazyThunk(arg, i.env, i)
						} else if isByRef {
							// For var parameters, create a reference
							if argIdent, ok := arg.(*ast.Identifier); ok {
								if val, exists := i.env.Get(argIdent.Value); exists {
									if refVal, isRef := val.(*ReferenceValue); isRef {
										args[idx] = refVal // Pass through existing reference
									} else {
										args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
									}
								} else {
									args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
								}
							} else {
								return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
							}
						} else {
							// For regular parameters, evaluate immediately
							val := i.Eval(arg)
							if isError(val) {
								return val
							}
							args[idx] = val
						}
					}
					return i.callUserFunction(fn, args)
				}
				// Function not found in unit
				return i.newErrorWithLocation(callExpr, "function '%s' not found in unit '%s'", memberAccess.Member.Value, unitIdent.Value)
			}
		}

		// Check if this is a class constructor call (TClass.Create(...))
		var classInfo *ClassInfo
		for className, class := range i.classes {
			if ident.Equal(className, unitIdent.Value) {
				classInfo = class
				break
			}
		}
		if classInfo != nil {
			// This is a class constructor/method call - convert to MethodCallExpression
			mc := &ast.MethodCallExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: callExpr.Token,
					},
				},
				Object:    unitIdent,
				Method:    memberAccess.Member,
				Arguments: callExpr.Arguments,
			}
			return i.evalMethodCall(mc)
		}
	}

	return i.newErrorWithLocation(callExpr, "cannot call member expression that is not a method or unit-qualified function")
}

// ===== Task 3.5.97: User Function Call Methods =====

// CallUserFunctionWithOverloads calls a user-defined function with overload resolution.
// Task 3.5.97a: Enables evaluator to call user functions without using EvalNode.
func (i *Interpreter) CallUserFunctionWithOverloads(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 210-265

	funcNameLower := ident.Normalize(funcName.Value)
	overloads, exists := i.functions[funcNameLower]
	if !exists || len(overloads) == 0 {
		return i.newErrorWithLocation(callExpr, "function '%s' not found", funcName.Value)
	}

	// Resolve overload based on argument types and get cached evaluated arguments
	fn, cachedArgs, err := i.resolveOverload(funcNameLower, overloads, callExpr.Arguments)
	if err != nil {
		return i.newErrorWithLocation(callExpr, "%s", err.Error())
	}

	// Prepare arguments - lazy parameters get LazyThunks, var parameters get References
	args := make([]Value, len(callExpr.Arguments))
	for idx, arg := range callExpr.Arguments {
		// Check parameter flags
		isLazy := idx < len(fn.Parameters) && fn.Parameters[idx].IsLazy
		isByRef := idx < len(fn.Parameters) && fn.Parameters[idx].ByRef

		if isLazy {
			// For lazy parameters, create a LazyThunk
			args[idx] = NewLazyThunk(arg, i.env, i)
		} else if isByRef {
			// For var parameters, create a reference
			if argIdent, ok := arg.(*ast.Identifier); ok {
				if val, exists := i.env.Get(argIdent.Value); exists {
					if refVal, isRef := val.(*ReferenceValue); isRef {
						args[idx] = refVal // Pass through existing reference
					} else {
						args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
					}
				} else {
					args[idx] = &ReferenceValue{Env: i.env, VarName: argIdent.Value}
				}
			} else {
				return i.newErrorWithLocation(arg, "var parameter requires a variable, got %T", arg)
			}
		} else {
			// For regular parameters, use the cached value from overload resolution
			args[idx] = cachedArgs[idx]
		}
	}
	return i.callUserFunction(fn, args)
}

// CallImplicitSelfMethod calls a method on the implicit Self object.
// Task 3.5.97b: Enables evaluator to call implicit Self methods without using EvalNode.
func (i *Interpreter) CallImplicitSelfMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 267-291

	selfVal, ok := i.env.Get("Self")
	if !ok {
		return i.newErrorWithLocation(callExpr, "no Self context for implicit method call")
	}

	obj, isObj := AsObject(selfVal)
	if !isObj {
		return i.newErrorWithLocation(callExpr, "Self is not an object")
	}

	if obj.GetMethod(funcName.Value) == nil {
		return i.newErrorWithLocation(callExpr, "method '%s' not found on Self", funcName.Value)
	}

	// Create a method call expression: Self.MethodName(args)
	mc := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: callExpr.Token,
			},
		},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: funcName.Token,
				},
			},
			Value: "Self",
		},
		Method:    funcName,
		Arguments: callExpr.Arguments,
	}
	return i.evalMethodCall(mc)
}

// CallRecordStaticMethod calls a static method within a record context.
// Task 3.5.97c: Enables evaluator to call record static methods without using EvalNode.
func (i *Interpreter) CallRecordStaticMethod(callExpr *ast.CallExpression, funcName *ast.Identifier) evaluator.Value {
	// This method encapsulates the logic from evalCallExpression lines 293+

	recordVal, ok := i.env.Get("__CurrentRecord__")
	if !ok {
		return i.newErrorWithLocation(callExpr, "no __CurrentRecord__ context for static method call")
	}

	rtv, isRecord := recordVal.(*RecordTypeValue)
	if !isRecord {
		return i.newErrorWithLocation(callExpr, "__CurrentRecord__ is not a RecordTypeValue")
	}

	// Check if the function name matches a static method (case-insensitive)
	methodNameLower := ident.Normalize(funcName.Value)
	overloads, exists := rtv.ClassMethodOverloads[methodNameLower]
	if !exists || len(overloads) == 0 {
		return i.newErrorWithLocation(callExpr, "static method '%s' not found on record type '%s'", funcName.Value, rtv.RecordType.Name)
	}

	// Found a static method - convert to qualified call
	mc := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: callExpr.Token,
			},
		},
		Object: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: funcName.Token,
				},
			},
			Value: rtv.RecordType.Name,
		},
		Method:    funcName,
		Arguments: callExpr.Arguments,
	}
	return i.evalMethodCall(mc)
}

// WrapJSONValueInVariant wraps a jsonvalue.Value in a VariantValue containing a JSONValue.
// Task 3.5.99b: Implements InterpreterAdapter.WrapJSONValueInVariant for JSON indexing support.
func (i *Interpreter) WrapJSONValueInVariant(jv any) evaluator.Value {
	// Convert any to *jsonvalue.Value
	jsonVal, ok := jv.(*jsonvalue.Value)
	if !ok && jv != nil {
		// Invalid type passed - return error
		return &ErrorValue{Message: "invalid type passed to WrapJSONValueInVariant: expected *jsonvalue.Value"}
	}

	// Use the existing jsonValueToVariant helper
	return jsonValueToVariant(jsonVal)
}

// CallIndexedPropertyGetter calls an indexed property getter method on an object.
// Task 3.5.99c: Implements InterpreterAdapter.CallIndexedPropertyGetter for object default property access.
func (i *Interpreter) CallIndexedPropertyGetter(obj evaluator.Value, propImpl any, indices []evaluator.Value, node any) evaluator.Value {
	// Convert obj to ObjectInstance
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects ObjectInstance"}
	}

	// Convert propImpl to *types.PropertyInfo
	propInfo, ok := propImpl.(*types.PropertyInfo)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects *types.PropertyInfo"}
	}

	// Convert node to ast.Node
	astNode, ok := node.(ast.Node)
	if !ok {
		return &ErrorValue{Message: "CallIndexedPropertyGetter expects ast.Node"}
	}

	// Convert []evaluator.Value to []Value (they're the same underlying interface)
	// evaluator.Value is an alias for the local Value interface in the interp package
	convertedIndices := make([]Value, len(indices))
	for i, idx := range indices {
		convertedIndices[i] = idx
	}

	// Delegate to the existing evalIndexedPropertyRead method
	return i.evalIndexedPropertyRead(objInst, propInfo, convertedIndices, astNode)
}

// CallRecordPropertyGetter calls a record property getter method.
// Task 3.5.99e: Implements InterpreterAdapter.CallRecordPropertyGetter for record default property access.
func (i *Interpreter) CallRecordPropertyGetter(record evaluator.Value, propImpl any, indices []evaluator.Value, node any) evaluator.Value {
	// Convert record to RecordValue
	recordVal, ok := record.(*RecordValue)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects RecordValue"}
	}

	// Convert propImpl to *types.RecordPropertyInfo
	propInfo, ok := propImpl.(*types.RecordPropertyInfo)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects *types.RecordPropertyInfo"}
	}

	// Convert node to ast.Node (specifically *ast.IndexExpression for now)
	indexExpr, ok := node.(*ast.IndexExpression)
	if !ok {
		return &ErrorValue{Message: "CallRecordPropertyGetter expects *ast.IndexExpression"}
	}

	// Check if the property has a read accessor
	if propInfo.ReadField == "" {
		return i.newErrorWithLocation(indexExpr, "default property is write-only")
	}

	// Get the getter method
	getterMethod := recordVal.GetMethod(propInfo.ReadField)
	if getterMethod == nil {
		return i.newErrorWithLocation(indexExpr, "default property read accessor '%s' is not a method", propInfo.ReadField)
	}

	// Convert []evaluator.Value to []Value
	convertedIndices := make([]Value, len(indices))
	for idx, val := range indices {
		convertedIndices[idx] = val
	}

	// Create a synthetic method call expression: record.GetterMethod(index)
	// We need to bind the index value(s) in the environment temporarily
	methodCall := &ast.MethodCallExpression{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{Token: indexExpr.Token},
		},
		Object: indexExpr.Left,
		Method: &ast.Identifier{
			Value: propInfo.ReadField,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		},
		Arguments: make([]ast.Expression, len(indices)),
	}

	// Create temporary identifiers for each index argument
	for idx := range indices {
		tempVarName := fmt.Sprintf("__temp_default_index_%d__", idx)
		methodCall.Arguments[idx] = &ast.Identifier{
			Value: tempVarName,
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{Token: indexExpr.Token},
			},
		}
		// Bind the index value in the environment
		i.env.Define(tempVarName, convertedIndices[idx])
	}

	// Call the getter method
	return i.evalMethodCall(methodCall)
}

// Phase 3.5.4 - Phase 2B: Type system access adapter methods
// These methods implement the InterpreterAdapter interface for type system access.

// ===== Class Registry =====

// LookupClass finds a class by name in the class registry.
// Task 3.5.46: Delegates to TypeSystem instead of using legacy map.
func (i *Interpreter) LookupClass(name string) (any, bool) {
	class := i.typeSystem.LookupClass(name)
	if class == nil {
		return nil, false
	}
	return class, true
}

// HasClass checks if a class with the given name exists.
// Task 3.5.46: Delegates to TypeSystem instead of using legacy map.
func (i *Interpreter) LookupRecord(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	record, ok := i.records[normalizedName]
	if !ok {
		return nil, false
	}
	return record, true
}

// HasRecord checks if a record type with the given name exists.
func (i *Interpreter) LookupInterface(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	iface, ok := i.interfaces[normalizedName]
	if !ok {
		return nil, false
	}
	return iface, true
}

// HasInterface checks if an interface with the given name exists.
func (i *Interpreter) LookupHelpers(typeName string) []any {
	normalizedName := ident.Normalize(typeName)
	helpers, ok := i.helpers[normalizedName]
	if !ok {
		return nil
	}
	// Convert []*HelperInfo to []any
	result := make([]any, len(helpers))
	for idx, helper := range helpers {
		result[idx] = helper
	}
	return result
}

// HasHelpers checks if a type has helper methods defined.
func (i *Interpreter) GetOperatorRegistry() any {
	return i.globalOperators
}

// GetConversionRegistry returns the conversion registry for type conversion lookups.
func (i *Interpreter) GetEnumTypeID(enumName string) int {
	normalizedName := ident.Normalize(enumName)
	typeID, ok := i.enumTypeIDRegistry[normalizedName]
	if !ok {
		return 0
	}
	return typeID
}

// ===== Task 3.5.5: Type System Adapter Method Implementations =====

// GetType resolves a type by name using the type system.
func (i *Interpreter) GetType(name string) (any, error) {
	typ, err := i.resolveType(name)
	if err != nil {
		return nil, err
	}
	return typ, nil
}

// ResolveType resolves a type from an AST type annotation.
func (i *Interpreter) ConvertValue(value evaluator.Value, targetTypeName string) (evaluator.Value, error) {
	// Convert from evaluator.Value to internal Value
	internalValue := value.(Value)

	// Try implicit conversion first
	if converted, ok := i.tryImplicitConversion(internalValue, targetTypeName); ok {
		return converted, nil
	}

	// Conversion failed
	return nil, fmt.Errorf("cannot convert %s to %s", value.Type(), targetTypeName)
}

// CreateDefaultValue creates a zero/default value for a given type name.
func (i *Interpreter) ParseInlineArrayType(typeName string) (any, error) {
	arrType := i.parseInlineArrayType(typeName)
	if arrType == nil {
		return nil, fmt.Errorf("invalid inline array type: %s", typeName)
	}
	return arrType, nil
}

// ParseInlineSetType parses inline set type signatures.
func (i *Interpreter) LookupSubrangeType(name string) (any, bool) {
	normalizedName := ident.Normalize(name)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	return typeVal, ok
}

// BoxVariant wraps a value in a Variant container.
func (i *Interpreter) BoxVariant(value evaluator.Value) evaluator.Value {
	return boxVariant(value.(Value))
}

// TryImplicitConversion attempts an implicit type conversion.
func (i *Interpreter) TryImplicitConversion(value evaluator.Value, targetTypeName string) (evaluator.Value, bool) {
	converted, ok := i.tryImplicitConversion(value.(Value), targetTypeName)
	if ok {
		return converted, true
	}
	return value, false
}

// WrapInSubrange wraps an integer value in a subrange type with validation.
func (i *Interpreter) WrapInSubrange(value evaluator.Value, subrangeTypeName string, node ast.Node) (evaluator.Value, error) {
	normalizedName := ident.Normalize(subrangeTypeName)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	if !ok {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	stv, ok := typeVal.(*SubrangeTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a subrange type", subrangeTypeName)
	}

	// Extract integer value
	var intValue int
	if intVal, ok := value.(*IntegerValue); ok {
		intValue = int(intVal.Value)
	} else if srcSubrange, ok := value.(*SubrangeValue); ok {
		intValue = srcSubrange.Value
	} else {
		return nil, fmt.Errorf("cannot convert %s to subrange type %s", value.Type(), subrangeTypeName)
	}

	// Create subrange value and validate
	subrangeVal := &SubrangeValue{
		Value:        0, // Will be set by ValidateAndSet
		SubrangeType: stv.SubrangeType,
	}
	if err := subrangeVal.ValidateAndSet(intValue); err != nil {
		return nil, err
	}
	return subrangeVal, nil
}

// WrapInInterface wraps an object value in an interface instance.
func (i *Interpreter) WrapInInterface(value evaluator.Value, interfaceName string, node ast.Node) (evaluator.Value, error) {
	ifaceInfo, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Check if the value is already an InterfaceInstance
	if _, alreadyInterface := value.(*InterfaceInstance); alreadyInterface {
		return value, nil
	}

	// Check if the value is an ObjectInstance
	objInst, isObj := value.(*ObjectInstance)
	if !isObj {
		return nil, fmt.Errorf("cannot wrap %s in interface %s", value.Type(), interfaceName)
	}

	// Validate that the object's class implements the interface
	if !classImplementsInterface(objInst.Class, ifaceInfo) {
		return nil, fmt.Errorf("class '%s' does not implement interface '%s'",
			objInst.Class.Name, ifaceInfo.Name)
	}

	// Wrap the object in an InterfaceInstance
	return NewInterfaceInstance(ifaceInfo, objInst), nil
}

// EvalArrayLiteralWithExpectedType evaluates an array literal with expected type context.
func (i *Interpreter) EvalArrayLiteralWithExpectedType(lit ast.Node, expectedTypeName string) evaluator.Value {
	arrayLit, ok := lit.(*ast.ArrayLiteralExpression)
	if !ok {
		return i.newErrorWithLocation(lit, "expected array literal expression")
	}

	// Resolve expected type
	arrType, errVal := i.arrayTypeByName(expectedTypeName, lit)
	if errVal != nil {
		return errVal
	}

	return i.evalArrayLiteralWithExpected(arrayLit, arrType)
}

// ClassImplementsInterface checks if a class implements an interface.
func (i *Interpreter) CreateExternalVar(varName, externalName string) evaluator.Value {
	return &ExternalVarValue{
		Name:         varName,
		ExternalName: externalName,
	}
}

// ResolveArrayTypeNode resolves an array type from an AST ArrayTypeNode.
func (i *Interpreter) ResolveArrayTypeNode(arrayNode ast.Node) (any, error) {
	arrNode, ok := arrayNode.(*ast.ArrayTypeNode)
	if !ok {
		return nil, fmt.Errorf("expected ArrayTypeNode")
	}

	arrType := i.resolveArrayTypeNode(arrNode)
	if arrType == nil {
		return nil, fmt.Errorf("failed to resolve array type")
	}
	return arrType, nil
}

// CreateRecordZeroValue creates a zero-initialized record value.
func (i *Interpreter) CreateRecordZeroValue(recordTypeName string) (evaluator.Value, error) {
	normalizedName := ident.Normalize(recordTypeName)
	recordTypeKey := "__record_type_" + normalizedName
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil, fmt.Errorf("record type '%s' not found", recordTypeName)
	}

	rtv, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a record type", recordTypeName)
	}

	return i.createRecordValue(rtv.RecordType, rtv.Methods), nil
}

// CreateArrayZeroValue creates a zero-initialized array value.
// Task 3.5.69c: Migrated to use TypeSystem instead of environment lookup.
func (i *Interpreter) CreateArrayZeroValue(arrayTypeName string) (evaluator.Value, error) {
	arrayType := i.typeSystem.LookupArrayType(arrayTypeName)
	if arrayType == nil {
		return nil, fmt.Errorf("array type '%s' not found", arrayTypeName)
	}

	return NewArrayValue(arrayType), nil
}

// CreateSetZeroValue creates an empty set value.
func (i *Interpreter) CreateSetZeroValue(setTypeName string) (evaluator.Value, error) {
	setType := i.parseInlineSetType(setTypeName)
	if setType == nil {
		return nil, fmt.Errorf("invalid set type: %s", setTypeName)
	}
	return NewSetValue(setType), nil
}

// CreateSubrangeZeroValue creates a zero-initialized subrange value.
func (i *Interpreter) CreateSubrangeZeroValue(subrangeTypeName string) (evaluator.Value, error) {
	normalizedName := ident.Normalize(subrangeTypeName)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	if !ok {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	stv, ok := typeVal.(*SubrangeTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a subrange type", subrangeTypeName)
	}

	return &SubrangeValue{
		Value:        stv.SubrangeType.LowBound,
		SubrangeType: stv.SubrangeType,
	}, nil
}

// CreateInterfaceZeroValue creates a nil interface instance.
func (i *Interpreter) CreateInterfaceZeroValue(interfaceName string) (evaluator.Value, error) {
	ifaceInfo, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	return &InterfaceInstance{
		Interface: ifaceInfo,
		Object:    nil,
	}, nil
}

// CreateClassZeroValue creates a typed nil value for a class.
// Task 3.5.46: Use TypeSystem for class lookup.
func (i *Interpreter) CreateClassZeroValue(className string) (evaluator.Value, error) {
	if !i.typeSystem.HasClass(className) {
		return nil, fmt.Errorf("class '%s' not found", className)
	}

	return &NilValue{ClassType: className}, nil
}

// ===== Task 3.5.40: Record Literal Adapter Method Implementations =====

// CreateRecordValue creates a record value with field initialization.
func (i *Interpreter) CreateRecordValue(recordTypeName string, fieldValues map[string]evaluator.Value) (evaluator.Value, error) {
	normalizedName := ident.Normalize(recordTypeName)
	recordTypeKey := "__record_type_" + normalizedName
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil, fmt.Errorf("record type '%s' not found", recordTypeName)
	}

	recordTypeValue, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a record type", recordTypeName)
	}

	recordType := recordTypeValue.RecordType

	// Create the record value with methods
	// Task 3.5.42: Updated to use RecordMetadata
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
		Metadata:   recordTypeValue.Metadata,
		Methods:    recordTypeValue.Methods, // Deprecated: backward compatibility
	}

	// Copy provided field values (already evaluated)
	for fieldName, fieldValue := range fieldValues {
		fieldNameLower := ident.Normalize(fieldName)
		// Validate field exists
		if _, exists := recordType.Fields[fieldNameLower]; !exists {
			return nil, fmt.Errorf("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
		}
		// Convert evaluator.Value to internal Value
		recordValue.Fields[fieldNameLower] = fieldValue.(Value)
	}

	// Initialize remaining fields with field initializers or default values
	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	for fieldName, fieldType := range recordType.Fields {
		if _, exists := recordValue.Fields[fieldName]; !exists {
			var fieldValue Value

			// Check if field has an initializer expression
			if fieldDecl, hasDecl := recordTypeValue.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				// Evaluate the field initializer
				fieldValue = i.Eval(fieldDecl.InitValue)
				if isError(fieldValue) {
					return nil, fmt.Errorf("error evaluating field initializer for '%s': %s", fieldName, fieldValue.(*ErrorValue).Message)
				}
			}

			// If no initializer, use getZeroValueForType
			if fieldValue == nil {
				fieldValue = getZeroValueForType(fieldType, methodsLookup)

				// Handle interface-typed fields specially
				if intfValue := i.initializeInterfaceField(fieldType); intfValue != nil {
					fieldValue = intfValue
				}
			}

			recordValue.Fields[fieldName] = fieldValue
		}
	}

	return recordValue, nil
}

// GetRecordFieldDeclarations retrieves field declarations for a record type.
func (i *Interpreter) GetZeroValueForType(typeInfo any) evaluator.Value {
	t, ok := typeInfo.(types.Type)
	if !ok {
		return &NilValue{}
	}

	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	return getZeroValueForType(t, methodsLookup)
}

// InitializeInterfaceField creates a nil interface instance for interface-typed fields.
func (i *Interpreter) MatchesExceptionType(exc interface{}, typeExpr ast.TypeExpression) bool {
	excVal, ok := exc.(*ExceptionValue)
	if !ok {
		return false
	}
	return i.matchesExceptionType(excVal, typeExpr)
}

// GetExceptionInstance returns the ObjectInstance from an exception.
func (i *Interpreter) GetExceptionInstance(exc interface{}) evaluator.Value {
	excVal, ok := exc.(*ExceptionValue)
	if !ok {
		return nil
	}
	return excVal.Instance
}

// CreateExceptionFromObject creates an ExceptionValue from an object instance.
func (i *Interpreter) CreateExceptionFromObject(obj evaluator.Value, ctx *evaluator.ExecutionContext, pos any) interface{} {
	// Should be an object instance
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("runtime error: raise requires exception object, got %s", obj.Type()))
	}

	// Get the class info
	classInfo := objInst.Class

	// Extract message from the object's Message field
	message := ""
	if msgVal, ok := objInst.Fields["Message"]; ok {
		if strVal, ok := msgVal.(*StringValue); ok {
			message = strVal.Value
		}
	}

	// Capture current call stack from context
	callStack := make(errors.StackTrace, len(ctx.CallStack()))
	copy(callStack, ctx.CallStack())

	// Get position
	var excPos *lexer.Position
	if p, ok := pos.(lexer.Position); ok {
		excPos = &p
	} else if p, ok := pos.(*lexer.Position); ok {
		excPos = p
	}

	return &ExceptionValue{
		ClassInfo: classInfo,
		Message:   message,
		Instance:  objInst,
		Position:  excPos,
		CallStack: callStack,
	}
}

// EvalBlockStatement evaluates a block statement in the given context.
func (i *Interpreter) EvalBlockStatement(block *ast.BlockStatement, ctx *evaluator.ExecutionContext) {
	// Sync context state to interpreter
	i.syncFromContext(ctx)
	defer i.syncToContext(ctx)

	i.evalBlockStatement(block)
}

// EvalStatement evaluates a single statement in the given context.
func (i *Interpreter) EvalStatement(stmt ast.Statement, ctx *evaluator.ExecutionContext) {
	// Sync context state to interpreter
	i.syncFromContext(ctx)
	defer i.syncToContext(ctx)

	i.Eval(stmt)
}

// syncFromContext syncs execution state from context to interpreter.
func (i *Interpreter) syncFromContext(ctx *evaluator.ExecutionContext) {
	// Sync exception state
	if exc := ctx.Exception(); exc != nil {
		if excVal, ok := exc.(*ExceptionValue); ok {
			i.exception = excVal
		}
	} else {
		i.exception = nil
	}

	// Sync handler exception
	if hexc := ctx.HandlerException(); hexc != nil {
		if excVal, ok := hexc.(*ExceptionValue); ok {
			i.handlerException = excVal
		}
	} else {
		i.handlerException = nil
	}
}

// syncToContext syncs execution state from interpreter to context.
func (i *Interpreter) syncToContext(ctx *evaluator.ExecutionContext) {
	// Sync exception state back
	ctx.SetException(i.exception)
	ctx.SetHandlerException(i.handlerException)
}

// ===== Task 3.5.6: Array and Collection Adapter Method Implementations =====

// CreateArray creates an array from a list of elements with a specified element type.
func (i *Interpreter) CreateArray(elementType any, elements []evaluator.Value) evaluator.Value {
	// Convert elementType to types.Type
	var typedElementType types.Type
	if elementType != nil {
		if t, ok := elementType.(types.Type); ok {
			typedElementType = t
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Create array type (dynamic array has nil bounds)
	arrayType := &types.ArrayType{
		ElementType: typedElementType,
		LowBound:    nil,
		HighBound:   nil,
	}

	// Create array value
	arrayVal := NewArrayValue(arrayType)

	// Add elements (append to Elements slice)
	arrayVal.Elements = append(arrayVal.Elements, internalElements...)

	return arrayVal
}

// CreateDynamicArray allocates a new dynamic array of a given size and element type.
func (i *Interpreter) CreateArrayValue(arrayType any, elements []evaluator.Value) evaluator.Value {
	// Convert arrayType to *types.ArrayType
	var typedArrayType *types.ArrayType
	if arrayType != nil {
		if at, ok := arrayType.(*types.ArrayType); ok {
			typedArrayType = at
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Create and return the array value
	return &ArrayValue{
		ArrayType: typedArrayType,
		Elements:  internalElements,
	}
}

// GetArrayElement retrieves an element from an array at the given index.
func (i *Interpreter) CallMethod(obj evaluator.Value, methodName string, args []evaluator.Value, node ast.Node) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Task 3.5.111c: Handle CLASS_INFO values (class method calls)
	// Pattern: ClassInfoValue.Method() where Self is already a ClassInfoValue
	if classInfoVal, ok := internalObj.(*ClassInfoValue); ok {
		classInfo := classInfoVal.ClassInfo
		if classInfo == nil {
			return newError("ClassInfoValue has no class information")
		}

		// Look up class method (case-insensitive)
		classMethodOverloads := i.getMethodOverloadsInHierarchy(classInfo, methodName, true)
		if len(classMethodOverloads) == 0 {
			return newError("class method '%s' not found in class '%s'", methodName, classInfo.Name)
		}

		// Create a synthetic MethodCallExpression for overload resolution
		// We need this to support overloaded methods with different parameter types
		argExprs := make([]ast.Expression, len(internalArgs))
		for idx := range internalArgs {
			// We don't have actual expressions, so we pass nil
			// The overload resolver will fall back to argument count matching
			argExprs[idx] = nil
		}

		// Simple case: single overload or select by argument count
		var classMethod *ast.FunctionDecl
		if len(classMethodOverloads) == 1 {
			classMethod = classMethodOverloads[0]
		} else {
			// Find matching overload by argument count
			for _, m := range classMethodOverloads {
				if len(m.Parameters) == len(internalArgs) {
					classMethod = m
					break
				}
			}
			if classMethod == nil {
				return newError("no matching overload for class method '%s' in class '%s' with %d arguments",
					methodName, classInfo.Name, len(internalArgs))
			}
		}

		// Execute class method with Self bound to ClassInfoValue
		savedEnv := i.env
		methodEnv := NewEnclosedEnvironment(i.env)
		i.env = methodEnv

		// Check recursion depth
		if i.ctx.GetCallStack().WillOverflow() {
			i.env = savedEnv
			return i.raiseMaxRecursionExceeded()
		}

		// Push to call stack
		fullMethodName := classInfo.Name + "." + methodName
		i.pushCallStack(fullMethodName)
		defer i.popCallStack()

		// Bind Self to ClassInfoValue for class methods
		i.env.Define("Self", classInfoVal)
		i.env.Define("__CurrentClass__", classInfoVal)

		// Add class constants
		i.bindClassConstantsToEnv(classInfo)

		// Bind parameters with implicit conversion
		for idx, param := range classMethod.Parameters {
			arg := internalArgs[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.env.Define(param.Name.Value, arg)
		}

		// Initialize Result for functions
		if classMethod.ReturnType != nil {
			returnType := i.resolveTypeFromAnnotation(classMethod.ReturnType)
			defaultVal := i.getDefaultValue(returnType)
			i.env.Define("Result", defaultVal)
			i.env.Define(classMethod.Name.Value, &ReferenceValue{Env: i.env, VarName: "Result"})
		}

		// Execute method body
		result := i.Eval(classMethod.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		// Extract return value
		var returnValue Value
		if classMethod.ReturnType != nil {
			resultVal, resultOk := i.env.Get("Result")
			methodNameVal, methodNameOk := i.env.Get(classMethod.Name.Value)

			if resultOk && resultVal.Type() != "NIL" {
				returnValue = resultVal
			} else if methodNameOk && methodNameVal.Type() != "NIL" {
				returnValue = methodNameVal
			} else if resultOk {
				returnValue = resultVal
			} else {
				returnValue = &NilValue{}
			}
		} else {
			returnValue = &NilValue{}
		}

		i.env = savedEnv
		return returnValue
	}

	// Task 3.5.111e: Handle ClassValue (metaclass) constructor calls
	// Pattern: classVar.Create(args) where classVar is a "class of TBase" metaclass variable
	if classVal, ok := internalObj.(*ClassValue); ok {
		runtimeClass := classVal.ClassInfo
		if runtimeClass == nil {
			return newError("invalid class reference")
		}

		// Look up constructor in the runtime class
		constructorOverloads := i.getMethodOverloadsInHierarchy(runtimeClass, methodName, false)
		if len(constructorOverloads) == 0 {
			return newError("constructor '%s' not found in class '%s'", methodName, runtimeClass.Name)
		}

		// Simple overload resolution by argument count
		var constructor *ast.FunctionDecl
		if len(constructorOverloads) == 1 {
			constructor = constructorOverloads[0]
		} else {
			for _, c := range constructorOverloads {
				if len(c.Parameters) == len(internalArgs) {
					constructor = c
					break
				}
			}
			if constructor == nil {
				return newError("no matching overload for constructor '%s' in class '%s' with %d arguments",
					methodName, runtimeClass.Name, len(internalArgs))
			}
		}

		// Check argument count
		if len(internalArgs) != len(constructor.Parameters) {
			return newError("wrong number of arguments for constructor '%s': expected %d, got %d",
				methodName, len(constructor.Parameters), len(internalArgs))
		}

		// Create new instance of the runtime class
		newInstance := NewObjectInstance(runtimeClass)

		// Initialize all fields with default values
		for fieldName, fieldType := range runtimeClass.Fields {
			var defaultValue Value
			if fieldDecl, hasDecl := runtimeClass.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				// Use field initializer
				savedEnv := i.env
				tempEnv := NewEnclosedEnvironment(i.env)
				for constName, constValue := range runtimeClass.ConstantValues {
					tempEnv.Define(constName, constValue)
				}
				i.env = tempEnv
				defaultValue = i.Eval(fieldDecl.InitValue)
				i.env = savedEnv
				if isError(defaultValue) {
					return defaultValue
				}
			} else {
				defaultValue = getZeroValueForType(fieldType, nil)
			}
			newInstance.SetField(fieldName, defaultValue)
		}

		// Execute constructor
		savedEnv := i.env
		methodEnv := NewEnclosedEnvironment(i.env)
		i.env = methodEnv

		// Check recursion depth
		if i.ctx.GetCallStack().WillOverflow() {
			i.env = savedEnv
			return i.raiseMaxRecursionExceeded()
		}

		// Push to call stack
		fullMethodName := runtimeClass.Name + "." + methodName
		i.pushCallStack(fullMethodName)
		defer i.popCallStack()

		// Bind Self to the new instance
		i.env.Define("Self", newInstance)
		i.env.Define("__CurrentClass__", &ClassInfoValue{ClassInfo: runtimeClass})

		// Add class constants
		i.bindClassConstantsToEnv(runtimeClass)

		// Bind constructor parameters with implicit conversion
		for idx, param := range constructor.Parameters {
			arg := internalArgs[idx]
			if param.Type != nil {
				paramTypeName := param.Type.String()
				if converted, ok := i.tryImplicitConversion(arg, paramTypeName); ok {
					arg = converted
				}
			}
			i.env.Define(param.Name.Value, arg)
		}

		// Execute constructor body
		result := i.Eval(constructor.Body)
		if isError(result) {
			i.env = savedEnv
			return result
		}

		i.env = savedEnv

		// Constructors return the new instance (Result is Self)
		return newInstance
	}

	// Original OBJECT handling
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("not an object: %s", internalObj.Type()))
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		panic("object has no class information")
	}

	// Find method (case-insensitive) using the existing helper
	method := classInfo.lookupMethod(methodName)
	if method == nil {
		panic(fmt.Sprintf("method '%s' not found in class '%s'", methodName, classInfo.Name))
	}

	// Call the method using existing infrastructure
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("Self", objVal)
	i.env = tempEnv

	result := i.callUserFunction(method, internalArgs)

	i.env = savedEnv
	return result
}

// CallInheritedMethod executes an inherited (parent) method with the given arguments.
func (i *Interpreter) CallInheritedMethod(obj evaluator.Value, methodName string, args []evaluator.Value) evaluator.Value {
	// Convert to internal types
	internalObj := obj.(Value)
	internalArgs := convertEvaluatorArgs(args)

	// Get object instance
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return newError("inherited requires Self to be an object instance, got %s", internalObj.Type())
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return newError("object has no class information")
	}

	// Check parent class
	if classInfo.Parent == nil {
		return newError("class '%s' has no parent class", classInfo.Name)
	}

	parentInfo := classInfo.Parent

	// Find method in parent (case-insensitive)
	methodNameLower := ident.Normalize(methodName)
	method, exists := parentInfo.Methods[methodNameLower]
	if !exists {
		return newError("method, property, or field '%s' not found in parent class '%s'", methodName, parentInfo.Name)
	}

	// Call the method using existing infrastructure
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	tempEnv.Define("Self", objVal)
	i.env = tempEnv

	result := i.callUserFunction(method, internalArgs)

	i.env = savedEnv
	return result
}

// CreateObject creates a new object instance of the specified class with constructor arguments.
func (i *Interpreter) CreateObject(className string, args []evaluator.Value) (evaluator.Value, error) {
	// Convert arguments
	internalArgs := convertEvaluatorArgs(args)

	// Task 3.5.46: Look up class via TypeSystem (case-insensitive)
	classInfoIface := i.typeSystem.LookupClass(className)
	if classInfoIface == nil {
		return nil, fmt.Errorf("class '%s' not found", className)
	}
	classInfo, ok := classInfoIface.(*ClassInfo)
	if !ok {
		return nil, fmt.Errorf("class '%s' has invalid type", className)
	}

	// Check if trying to instantiate an abstract class
	if classInfo.IsAbstract {
		return nil, fmt.Errorf("Trying to create an instance of an abstract class")
	}

	// Check if trying to instantiate an external class
	if classInfo.IsExternal {
		return nil, fmt.Errorf("cannot instantiate external class '%s' - external classes are not supported", className)
	}

	// Create new object instance
	obj := NewObjectInstance(classInfo)

	// Initialize fields with default values
	savedEnv := i.env
	tempEnv := NewEnclosedEnvironment(i.env)
	i.env = tempEnv

	for fieldName, fieldType := range classInfo.Fields {
		var fieldValue Value
		if fieldDecl, hasDecl := classInfo.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
			fieldValue = i.Eval(fieldDecl.InitValue)
			if isError(fieldValue) {
				i.env = savedEnv
				return nil, fmt.Errorf("failed to initialize field '%s': %v", fieldName, fieldValue)
			}
		} else {
			fieldValue = getZeroValueForType(fieldType, nil)
		}
		obj.SetField(fieldName, fieldValue)
	}

	i.env = savedEnv

	// Call constructor if it exists
	constructorNameLower := ident.Normalize("Create")
	if constructor, exists := classInfo.Constructors[constructorNameLower]; exists {
		tempEnv := NewEnclosedEnvironment(i.env)
		tempEnv.Define("Self", obj)
		i.env = tempEnv

		result := i.callUserFunction(constructor, internalArgs)

		i.env = savedEnv

		// Propagate constructor errors
		if isError(result) {
			return nil, fmt.Errorf("constructor failed: %v", result)
		}
	} else if len(internalArgs) > 0 {
		return nil, fmt.Errorf("no constructor found for class '%s' with %d arguments", className, len(internalArgs))
	}

	return obj, nil
}

// CheckType checks if an object is of a specified type (implements 'is' operator).
// Task 3.5.34: Extended to support both class hierarchy and interface implementation checking.
func (i *Interpreter) CheckType(obj evaluator.Value, typeName string) bool {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil - nil is not an instance of any type
	if _, isNil := internalObj.(*NilValue); isNil {
		return false
	}

	// Check if it's an object
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return false
	}

	// Get class info
	classInfo := objVal.Class
	if classInfo == nil {
		return false
	}

	// Check if the object's class matches (case-insensitive)
	if ident.Equal(classInfo.Name, typeName) {
		return true
	}

	// Check parent class hierarchy
	current := classInfo.Parent
	for current != nil {
		if ident.Equal(current.Name, typeName) {
			return true
		}
		current = current.Parent
	}

	// Task 3.5.34: Check if the target is an interface and if the object's class implements it
	if iface, exists := i.interfaces[ident.Normalize(typeName)]; exists {
		return classImplementsInterface(objVal.Class, iface)
	}

	return false
}

// CastType performs type casting (implements 'as' operator).
// Task 3.5.35: Extended to fully support type casting with interface wrapping/unwrapping.
//
// Handles the following cases:
// 1. nil → any type: returns nil
// 2. interface → class: extracts underlying object (with type check)
// 3. interface → interface: creates new interface wrapper (with implementation check)
// 4. object → class: validates class hierarchy
// 5. object → interface: creates interface wrapper (with implementation check)
func (i *Interpreter) CastType(obj evaluator.Value, typeName string) (evaluator.Value, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil - nil can be cast to any type
	if _, isNil := internalObj.(*NilValue); isNil {
		return &NilValue{}, nil
	}

	// Handle interface-to-object/interface casting
	if intfInst, ok := internalObj.(*InterfaceInstance); ok {
		// Task 3.5.46: Check if target is a class via TypeSystem
		if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
			targetClass, _ := targetClassIface.(*ClassInfo)
			// Interface-to-class casting: extract the underlying object
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				return nil, fmt.Errorf("cannot cast nil interface to class '%s'", targetClass.Name)
			}

			// Check if the underlying object's class is compatible with the target class
			if !isClassCompatible(underlyingObj.Class, targetClass) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to class '%s'", underlyingObj.Class.Name, targetClass.Name)
			}

			// Cast is valid - return the underlying object
			return underlyingObj, nil
		}

		// Check if target is an interface
		if targetIface, isInterface := i.interfaces[ident.Normalize(typeName)]; isInterface {
			// Interface-to-interface casting
			underlyingObj := intfInst.Object
			if underlyingObj == nil {
				// DWScript: nil interface cast to interface yields nil interface wrapper
				return &InterfaceInstance{Interface: targetIface, Object: nil}, nil
			}

			// Check if the underlying object's class implements the target interface
			if !classImplementsInterface(underlyingObj.Class, targetIface) {
				return nil, fmt.Errorf("cannot cast interface of '%s' to interface '%s'", underlyingObj.Class.Name, targetIface.Name)
			}

			// Create and return new interface instance
			return NewInterfaceInstance(targetIface, underlyingObj), nil
		}

		return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
	}

	// Handle object casting
	objVal, ok := internalObj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("'as' operator requires object instance, got %s", internalObj.Type())
	}

	// Task 3.5.46: Try class-to-class casting first via TypeSystem
	if targetClassIface := i.typeSystem.LookupClass(typeName); targetClassIface != nil {
		targetClass, _ := targetClassIface.(*ClassInfo)
		// Validate that the object's actual runtime type is compatible with the target
		if !isClassCompatible(objVal.Class, targetClass) {
			return nil, fmt.Errorf("instance of type '%s' cannot be cast to class '%s'", objVal.Class.Name, targetClass.Name)
		}

		// Cast is valid - return the same object
		return objVal, nil
	}

	// Try interface casting
	if iface, exists := i.interfaces[ident.Normalize(typeName)]; exists {
		// Validate that the object's class implements the interface
		if !classImplementsInterface(objVal.Class, iface) {
			return nil, fmt.Errorf("class '%s' does not implement interface '%s'", objVal.Class.Name, iface.Name)
		}

		// Create and return the interface instance
		return NewInterfaceInstance(iface, objVal), nil
	}

	return nil, fmt.Errorf("type '%s' not found (neither class nor interface)", typeName)
}

// CastToClass performs class type casting for TypeName(expr) expressions.
// Task 3.5.94: Adapter method for type cast migration. Uses existing castToClass logic.
func (i *Interpreter) CastToClass(val evaluator.Value, className string, node ast.Expression) evaluator.Value {
	// Convert to internal type
	internalVal := val.(Value)

	// Look up the class
	classInfo := i.lookupClassInfo(className)
	if classInfo == nil {
		return nil // Not a class type, caller will try other options
	}

	// Use the existing castToClass method
	return i.castToClass(internalVal, classInfo, node)
}

// CheckImplements checks if an object/class implements an interface.
// Task 3.5.36: Adapter method for 'implements' operator.
// Supports ObjectInstance, ClassValue, and ClassInfoValue inputs.
func (i *Interpreter) CheckImplements(obj evaluator.Value, interfaceName string) (bool, error) {
	// Convert to internal type
	internalObj := obj.(Value)

	// Handle nil - nil implements no interfaces
	if _, isNil := internalObj.(*NilValue); isNil {
		return false, nil
	}

	// Extract ClassInfo from different value types
	var classInfo *ClassInfo

	if objInst, ok := internalObj.(*ObjectInstance); ok {
		// Object instance - extract class
		classInfo = objInst.Class
	} else if classVal, ok := internalObj.(*ClassValue); ok {
		// Class reference (e.g., from metaclass variable: var cls: class of TParent)
		classInfo = classVal.ClassInfo
	} else if classInfoVal, ok := internalObj.(*ClassInfoValue); ok {
		// Class type identifier (e.g., TMyImplementation in: if TMyImplementation implements IMyInterface then)
		classInfo = classInfoVal.ClassInfo
	} else {
		return false, fmt.Errorf("'implements' operator requires object instance or class type, got %s", internalObj.Type())
	}

	// Guard against nil ClassInfo (e.g., uninitialized metaclass variables)
	if classInfo == nil {
		return false, nil
	}

	// Look up the interface in the registry
	iface, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return false, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	// Check if the class implements the interface
	// 'implements' operator in DWScript only considers explicitly declared interfaces,
	// not interfaces inherited through other interfaces.
	return classExplicitlyImplementsInterface(classInfo, iface), nil
}

// CreateClassValue creates a ClassValue (metaclass reference) from a class name.
// Task 3.5.85: Adapter method for returning metaclass references from VisitIdentifier.
func (i *Interpreter) CreateClassValue(className string) (evaluator.Value, error) {
	// Look up the class in the registry (case-insensitive)
	for name, classInfo := range i.classes {
		if ident.Equal(name, className) {
			return &ClassValue{ClassInfo: classInfo}, nil
		}
	}
	return nil, fmt.Errorf("class '%s' not found", className)
}

// CreateFunctionPointer creates a function pointer value from a function declaration.
// Task 3.5.8: Adapter method for function pointer creation.
func (i *Interpreter) CreateFunctionPointer(fn *ast.FunctionDecl, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Function: fn,
		Closure:  env,
	}
}

// CreateLambda creates a lambda/closure value from a lambda expression.
// Task 3.5.8: Adapter method for lambda creation.
func (i *Interpreter) CreateLambda(lambda *ast.LambdaExpression, closure any) evaluator.Value {
	// Convert closure to Environment
	var env *Environment
	if closure != nil {
		env = closure.(*Environment)
	}

	return &FunctionPointerValue{
		Lambda:  lambda,
		Closure: env,
	}
}

// IsFunctionPointer checks if a value is a function pointer.
// Task 3.5.8: Adapter method for function pointer type checking.
func (i *Interpreter) IsFunctionPointer(value evaluator.Value) bool {
	_, ok := value.(*FunctionPointerValue)
	return ok
}

// GetFunctionPointerParamCount returns the number of parameters a function pointer expects.
// Task 3.5.8: Adapter method for function pointer parameter count.
func (i *Interpreter) GetFunctionPointerParamCount(funcPtr evaluator.Value) int {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return 0
	}

	if fp.Function != nil {
		return len(fp.Function.Parameters)
	} else if fp.Lambda != nil {
		return len(fp.Lambda.Parameters)
	}

	return 0
}

// IsFunctionPointerNil checks if a function pointer is nil (unassigned).
// Task 3.5.8: Adapter method for function pointer nil checking.
func (i *Interpreter) IsFunctionPointerNil(funcPtr evaluator.Value) bool {
	fp, ok := funcPtr.(*FunctionPointerValue)
	if !ok {
		return false
	}

	// A function pointer is nil if both Function and Lambda are nil
	return fp.Function == nil && fp.Lambda == nil
}

// CreateMethodPointer creates a method pointer value bound to a specific object.
// Task 3.5.37: Adapter method for method pointer creation from @object.MethodName expressions.
func (i *Interpreter) CreateMethodPointer(objVal evaluator.Value, methodName string, closure any) (evaluator.Value, error) {
	// Extract the object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return nil, fmt.Errorf("method pointer requires an object instance, got %s", objVal.Type())
	}

	// Look up the method in the class hierarchy (case-insensitive)
	method := obj.Class.lookupMethod(methodName)
	if method == nil {
		return nil, fmt.Errorf("undefined method: %s.%s", obj.Class.Name, methodName)
	}

	// Convert closure to Environment
	// Handle both direct *Environment and *EnvironmentAdapter (from evaluator)
	var env *Environment
	if closure != nil {
		if adapter, ok := closure.(*evaluator.EnvironmentAdapter); ok {
			env = adapter.Underlying().(*Environment)
		} else if envVal, ok := closure.(*Environment); ok {
			env = envVal
		}
	}

	// Build parameter types for the function pointer type
	paramTypes := make([]types.Type, len(method.Parameters))
	for idx, param := range method.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if method.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(method.ReturnType)
	}

	// Create the method pointer type
	methodPtr := types.NewMethodPointerType(paramTypes, returnType)
	pointerType := &methodPtr.FunctionPointerType

	// Create and return the function pointer value with SelfObject bound
	return NewFunctionPointerValue(method, env, objVal, pointerType), nil
}

// CreateFunctionPointerFromName creates a function pointer for a named function.
// Task 3.5.37: Adapter method for function pointer creation from @FunctionName expressions.
func (i *Interpreter) CreateFunctionPointerFromName(funcName string, closure any) (evaluator.Value, error) {
	// Look up the function in the function registry (case-insensitive)
	overloads, exists := i.functions[ident.Normalize(funcName)]
	if !exists || len(overloads) == 0 {
		return nil, fmt.Errorf("undefined function or procedure: %s", funcName)
	}

	// For overloaded functions, use the first overload
	// Note: Function pointers cannot represent overload sets, only single functions
	function := overloads[0]

	// Convert closure to Environment
	// Handle both direct *Environment and *EnvironmentAdapter (from evaluator)
	var env *Environment
	if closure != nil {
		if adapter, ok := closure.(*evaluator.EnvironmentAdapter); ok {
			env = adapter.Underlying().(*Environment)
		} else if envVal, ok := closure.(*Environment); ok {
			env = envVal
		}
	}

	// Build parameter types for the function pointer type
	paramTypes := make([]types.Type, len(function.Parameters))
	for idx, param := range function.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		} else {
			paramTypes[idx] = &types.IntegerType{} // Default fallback
		}
	}

	// Get return type
	var returnType types.Type
	if function.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(function.ReturnType)
	}

	// Create the function pointer type
	var pointerType *types.FunctionPointerType
	if returnType != nil {
		pointerType = types.NewFunctionPointerType(paramTypes, returnType)
	} else {
		pointerType = types.NewProcedurePointerType(paramTypes)
	}

	// Create and return the function pointer value (no SelfObject)
	return NewFunctionPointerValue(function, env, nil, pointerType), nil
}

// CreateRecord creates a record value from field values.
// Task 3.5.8: Adapter method for record construction.
func (i *Interpreter) RaiseException(className string, message string, pos any) {
	// Convert pos to lexer.Position if provided
	var position *lexer.Position
	if pos != nil {
		if p, ok := pos.(*lexer.Position); ok {
			position = p
		}
	}

	// Call the internal raiseException method
	i.raiseException(className, message, position)
}

// Task 3.5.70: GetVariable removed - evaluator now uses ctx.Env().Get() directly

// DefineVariable defines a new variable in the execution context.
// Task 3.5.9: Adapter method for environment access.
func (i *Interpreter) DefineVariable(name string, value evaluator.Value, ctx *evaluator.ExecutionContext) {
	// Convert to internal Value type
	internalValue := value.(Value)

	// Define in the context's environment
	ctx.Env().Define(name, internalValue)
}

// ===== Task 3.5.19: Binary Operator Adapter Methods (Fix for PR #219) =====
//
// These adapter methods delegate to the Interpreter's binary operator implementation
// WITHOUT re-evaluating the operands. This fixes the double-evaluation bug where
// operands with side effects (function calls, increments, etc.) were executed twice.

// EvalVariantBinaryOp handles binary operations with Variant operands using pre-evaluated values.
func (i *Interpreter) EvalVariantBinaryOp(op string, left, right evaluator.Value, node ast.Node) evaluator.Value {
	// The Interpreter's evalVariantBinaryOp already works with pre-evaluated values
	return i.evalVariantBinaryOp(op, left, right, node)
}

// EvalInOperator evaluates the 'in' operator for membership testing using pre-evaluated values.
func (i *Interpreter) EvalInOperator(value, container evaluator.Value, node ast.Node) evaluator.Value {
	// The Interpreter's evalInOperator already works with pre-evaluated values
	return i.evalInOperator(value, container, node)
}

// EvalEqualityComparison handles = and <> operators for complex types using pre-evaluated values.
func (i *Interpreter) EvalEqualityComparison(op string, left, right evaluator.Value, node ast.Node) evaluator.Value {
	// This is extracted from eval BinaryExpression to handle complex type comparisons
	// with pre-evaluated operands (fixing double-evaluation bug in PR #219)

	// Check if either operand is nil or an object instance
	_, leftIsNil := left.(*NilValue)
	_, rightIsNil := right.(*NilValue)
	_, leftIsObj := left.(*ObjectInstance)
	_, rightIsObj := right.(*ObjectInstance)
	leftIntf, leftIsIntf := left.(*InterfaceInstance)
	rightIntf, rightIsIntf := right.(*InterfaceInstance)
	leftClass, leftIsClass := left.(*ClassValue)
	rightClass, rightIsClass := right.(*ClassValue)

	// Handle RTTITypeInfoValue comparisons (for TypeOf results)
	leftRTTI, leftIsRTTI := left.(*RTTITypeInfoValue)
	rightRTTI, rightIsRTTI := right.(*RTTITypeInfoValue)
	if leftIsRTTI && rightIsRTTI {
		// Compare by TypeID (unique identifier for each type)
		result := leftRTTI.TypeID == rightRTTI.TypeID
		if op == "=" {
			return &BooleanValue{Value: result}
		}
		return &BooleanValue{Value: !result}
	}

	// Handle ClassValue (metaclass) comparisons
	if leftIsClass || rightIsClass {
		// Both are ClassValue - compare by ClassInfo identity
		if leftIsClass && rightIsClass {
			result := leftClass.ClassInfo == rightClass.ClassInfo
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		// One is ClassValue, one is nil
		if leftIsNil || rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: false}
			}
			return &BooleanValue{Value: true}
		}
	}

	// Handle InterfaceInstance comparisons
	if leftIsIntf || rightIsIntf {
		// Both are interfaces - compare underlying objects
		if leftIsIntf && rightIsIntf {
			result := leftIntf.Object == rightIntf.Object
			if op == "=" {
				return &BooleanValue{Value: result}
			}
			return &BooleanValue{Value: !result}
		}
		// One is interface, one is nil
		if leftIsNil || rightIsNil {
			var intfIsNil bool
			if leftIsIntf {
				intfIsNil = leftIntf.Object == nil
			} else {
				intfIsNil = rightIntf.Object == nil
			}
			if op == "=" {
				return &BooleanValue{Value: intfIsNil}
			}
			return &BooleanValue{Value: !intfIsNil}
		}
	}

	// If either is nil or an object, do object identity comparison
	if leftIsNil || rightIsNil || leftIsObj || rightIsObj {
		// Both nil
		if leftIsNil && rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: true}
			}
			return &BooleanValue{Value: false}
		}

		// One is nil, one is not
		if leftIsNil || rightIsNil {
			if op == "=" {
				return &BooleanValue{Value: false}
			}
			return &BooleanValue{Value: true}
		}

		// Both are objects - compare by identity
		if op == "=" {
			return &BooleanValue{Value: left == right}
		}
		return &BooleanValue{Value: left != right}
	}

	// Check if both are records
	if _, leftIsRecord := left.(*RecordValue); leftIsRecord {
		if _, rightIsRecord := right.(*RecordValue); rightIsRecord {
			return i.evalRecordBinaryOp(op, left, right)
		}
	}

	// Not a supported equality comparison type
	return i.newErrorWithLocation(node, "type mismatch: %s %s %s", left.Type(), op, right.Type())
}

// ===== Task 3.5.21: Complex Value Retrieval Adapter Method Implementations =====
//
// These adapter methods allow the Evaluator to handle complex value types
// (ReferenceValue) that require special processing when accessed as identifiers.
//
// Task 3.5.71: IsReferenceValue removed - evaluator uses val.Type() == "REFERENCE" directly
// Task 3.5.73: IsExternalVar, IsLazyThunk, EvaluateLazyThunk, GetExternalVarName removed
//              - evaluator uses ExternalVarAccessor and LazyEvaluator interfaces directly

// DereferenceValue dereferences a var parameter reference.
// Returns the actual value and an error if dereferencing fails.
// Panics if the value is not a ReferenceValue.
func (i *Interpreter) DereferenceValue(value evaluator.Value) (evaluator.Value, error) {
	refVal, ok := value.(*ReferenceValue)
	if !ok {
		panic("DereferenceValue called on non-ReferenceValue value")
	}
	actualVal, err := refVal.Dereference()
	if err != nil {
		return nil, err
	}
	return actualVal, nil
}

// CreateLazyThunk creates a lazy parameter thunk from an unevaluated expression.
// Task 3.5.23: Enables lazy parameter evaluation in user function calls.
func (i *Interpreter) CreateLazyThunk(expr ast.Expression, env any) evaluator.Value {
	// Convert environment from any to *Environment
	environment, ok := env.(*Environment)
	if !ok {
		panic(fmt.Sprintf("CreateLazyThunk: env must be *Environment, got %T", env))
	}
	return NewLazyThunk(expr, environment, i)
}

// CreateReferenceValue creates a var parameter reference.
// Task 3.5.23: Enables pass-by-reference semantics for var parameters.
func (i *Interpreter) CreateReferenceValue(varName string, env any) evaluator.Value {
	// Convert environment from any to *Environment
	environment, ok := env.(*Environment)
	if !ok {
		panic(fmt.Sprintf("CreateReferenceValue: env must be *Environment, got %T", env))
	}
	return &ReferenceValue{
		Env:     environment,
		VarName: varName,
	}
}

// ===== Task 3.5.22: Property & Method Reference Adapter Method Implementations =====
//
// These adapter methods allow the Evaluator to access object fields, properties,
// methods, and class metadata when handling identifier lookups in method contexts.

// Task 3.5.71: IsObjectInstance removed - evaluator uses val.Type() == "OBJECT" directly

// GetObjectFieldValue retrieves a field value from an object instance.
func (i *Interpreter) GetObjectFieldValue(obj evaluator.Value, fieldName string) (evaluator.Value, bool) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, false
	}
	fieldValue := objInst.GetField(fieldName)
	if fieldValue == nil {
		return nil, false
	}
	return fieldValue, true
}

// GetClassVariableValue retrieves a class variable value from an object's class.
func (i *Interpreter) GetClassVariableValue(obj evaluator.Value, varName string) (evaluator.Value, bool) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, false
	}
	// Case-insensitive lookup to match DWScript semantics
	for name, value := range objInst.Class.ClassVars {
		if ident.Equal(name, varName) {
			return value, true
		}
	}
	return nil, false
}

// Task 3.5.72: HasProperty removed - ObjectInstance implements evaluator.ObjectValue directly

// ReadPropertyValue reads a property value from an object.
func (i *Interpreter) ReadPropertyValue(obj evaluator.Value, propName string, node any) (evaluator.Value, error) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot read property from non-object value")
	}

	propInfo := objInst.Class.lookupProperty(propName)
	if propInfo == nil {
		return nil, fmt.Errorf("property '%s' not found", propName)
	}

	// Use the existing evalPropertyRead method
	astNode, ok := node.(ast.Node)
	if !ok {
		astNode = nil
	}
	return i.evalPropertyRead(objInst, propInfo, astNode), nil
}

// Task 3.5.72: HasMethod removed - ObjectInstance implements evaluator.ObjectValue directly

// IsMethodParameterless checks if a method has zero parameters.
func (i *Interpreter) IsMethodParameterless(obj evaluator.Value, methodName string) bool {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return false
	}
	method, exists := objInst.Class.Methods[strings.ToLower(methodName)]
	if !exists {
		return false
	}
	return len(method.Parameters) == 0
}

// CreateMethodCall creates a synthetic method call expression for auto-invocation.
func (i *Interpreter) CreateMethodCall(obj evaluator.Value, methodName string, node any) evaluator.Value {
	// Create a synthetic method call and evaluate it
	// We create identifiers without token information since this is synthetic
	selfIdent := &ast.Identifier{Value: "Self"}
	methodIdent := &ast.Identifier{Value: methodName}

	// Copy token information from the original node if available
	if astNode, ok := node.(*ast.Identifier); ok {
		selfIdent.Token = astNode.Token
		methodIdent.Token = astNode.Token
	}

	syntheticCall := &ast.MethodCallExpression{
		Object:    selfIdent,
		Method:    methodIdent,
		Arguments: []ast.Expression{},
	}

	return i.evalMethodCall(syntheticCall)
}

// CreateMethodPointerFromObject creates a method pointer for a method with parameters.
func (i *Interpreter) CreateMethodPointerFromObject(obj evaluator.Value, methodName string) (evaluator.Value, error) {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("cannot create method pointer from non-object value")
	}

	method, exists := objInst.Class.Methods[strings.ToLower(methodName)]
	if !exists {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}

	// Build the pointer type
	paramTypes := make([]types.Type, len(method.Parameters))
	for idx, param := range method.Parameters {
		if param.Type != nil {
			paramTypes[idx] = i.getTypeFromAnnotation(param.Type)
		}
	}
	var returnType types.Type
	if method.ReturnType != nil {
		returnType = i.getTypeFromAnnotation(method.ReturnType)
	}
	pointerType := types.NewFunctionPointerType(paramTypes, returnType)

	return NewFunctionPointerValue(method, i.env, objInst, pointerType), nil
}

// GetClassName returns the class name for an object instance.
func (i *Interpreter) GetClassName(obj evaluator.Value) string {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return ""
	}
	return objInst.Class.Name
}

// GetClassType returns the ClassValue (metaclass) for an object instance.
func (i *Interpreter) GetClassType(obj evaluator.Value) evaluator.Value {
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		return nil
	}
	return &ClassValue{ClassInfo: objInst.Class}
}

// Task 3.5.71: IsClassInfoValue removed - evaluator uses val.Type() == "CLASSINFO" directly

// GetClassNameFromClassInfo returns the class name from a ClassInfoValue.
func (i *Interpreter) GetClassNameFromClassInfo(classInfo evaluator.Value) string {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassNameFromClassInfo called on non-ClassInfoValue value")
	}
	return classInfoVal.ClassInfo.Name
}

// GetClassTypeFromClassInfo returns the ClassValue from a ClassInfoValue.
func (i *Interpreter) GetClassTypeFromClassInfo(classInfo evaluator.Value) evaluator.Value {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassTypeFromClassInfo called on non-ClassInfoValue value")
	}
	return &ClassValue{ClassInfo: classInfoVal.ClassInfo}
}

// GetClassVariableFromClassInfo retrieves a class variable from ClassInfoValue.
func (i *Interpreter) GetClassVariableFromClassInfo(classInfo evaluator.Value, varName string) (evaluator.Value, bool) {
	classInfoVal, ok := classInfo.(*ClassInfoValue)
	if !ok {
		panic("GetClassVariableFromClassInfo called on non-ClassInfoValue value")
	}
	// Case-insensitive lookup to match DWScript semantics
	for name, value := range classInfoVal.ClassInfo.ClassVars {
		if ident.Equal(name, varName) {
			return value, true
		}
	}
	return nil, false
}

// GetCallStack returns a copy of the current call stack.
// Returns stack frames in the order they were called (oldest to newest).
func (i *Interpreter) GetCallStack() errors.StackTrace {
	// Return a copy to prevent external modification
	stack := make(errors.StackTrace, len(i.callStack))
	copy(stack, i.callStack)
	return stack
}

// pushCallStack adds a new frame to the call stack with the given function name.
// The position is taken from the current node being evaluated.
// Phase 3.3.3: Delegates to ExecutionContext's CallStack.
func (i *Interpreter) pushCallStack(functionName string) {
	var pos *lexer.Position
	if i.currentNode != nil {
		nodePos := i.currentNode.Pos()
		pos = &nodePos
	}
	// Also push to the old callStack field for backward compatibility
	frame := errors.NewStackFrame(functionName, i.sourceFile, pos)
	i.callStack = append(i.callStack, frame)

	// Phase 3.3.3: Also push to context's CallStack
	// Ignore errors here for backward compatibility; callers should check WillOverflow first
	_ = i.ctx.GetCallStack().Push(functionName, i.sourceFile, pos)
}

// popCallStack removes the most recent frame from the call stack.
// Phase 3.3.3: Delegates to ExecutionContext's CallStack.
func (i *Interpreter) popCallStack() {
	if len(i.callStack) > 0 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
	// Phase 3.3.3: Also pop from context's CallStack
	i.ctx.GetCallStack().Pop()
}

// Eval evaluates an AST node and returns its value.
// This is the main entry point for the interpreter.
func (i *Interpreter) Eval(node ast.Node) Value {
	// Track the current node for error reporting
	i.currentNode = node

	switch node := node.(type) {
	// Program
	case *ast.Program:
		return i.evalProgram(node)

	// Statements
	case *ast.ExpressionStatement:
		// Evaluate the expression
		val := i.Eval(node.Expression)
		if isError(val) {
			// Enrich error with statement location to mimic DWScript call stack output
			if errVal, ok := val.(*ErrorValue); ok {
				exprPos := node.Expression.Pos()
				lineMarker := fmt.Sprintf("line %d", exprPos.Line)
				loc := fmt.Sprintf("at line %d, column: %d", exprPos.Line, exprPos.Column+2)
				if !strings.Contains(errVal.Message, lineMarker) {
					errVal.Message = errVal.Message + "\n " + loc
				}
			}
			return val
		}

		// Auto-invoke parameterless function pointers stored in variables
		// In DWScript, when a variable holds a function pointer with no parameters
		// and is used as a statement, it's automatically invoked
		// Example: var fp := @SomeProc; fp; // auto-invokes SomeProc
		if funcPtr, isFuncPtr := val.(*FunctionPointerValue); isFuncPtr {
			// Determine parameter count
			paramCount := 0
			if funcPtr.Function != nil {
				paramCount = len(funcPtr.Function.Parameters)
			} else if funcPtr.Lambda != nil {
				paramCount = len(funcPtr.Lambda.Parameters)
			}

			// If it has zero parameters, auto-invoke it
			if paramCount == 0 {
				// Check if the function pointer is nil (not assigned)
				if funcPtr.Function == nil && funcPtr.Lambda == nil {
					// Raise an exception that can be caught by try-except
					i.raiseException("Exception", "Function pointer is nil", &node.Token.Pos)
					return &NilValue{}
				}
				return i.callFunctionPointer(funcPtr, []Value{}, node)
			}
		}

		return val

	case *ast.VarDeclStatement:
		return i.evalVarDeclStatement(node)

	case *ast.ConstDecl:
		return i.evalConstDecl(node)

	case *ast.AssignmentStatement:
		return i.evalAssignmentStatement(node)

	case *ast.BlockStatement:
		return i.evalBlockStatement(node)

	case *ast.IfStatement:
		return i.evalIfStatement(node)

	case *ast.WhileStatement:
		return i.evalWhileStatement(node)

	case *ast.RepeatStatement:
		return i.evalRepeatStatement(node)

	case *ast.ForStatement:
		return i.evalForStatement(node)

	case *ast.ForInStatement:
		return i.evalForInStatement(node)

	case *ast.CaseStatement:
		return i.evalCaseStatement(node)

	case *ast.TryStatement:
		return i.evalTryStatement(node)

	case *ast.RaiseStatement:
		return i.evalRaiseStatement(node)

	case *ast.BreakStatement:
		return i.evalBreakStatement(node)

	case *ast.ContinueStatement:
		return i.evalContinueStatement(node)

	case *ast.ExitStatement:
		return i.evalExitStatement(node)

	case *ast.ReturnStatement:
		// Handle return statements in lambda shorthand syntax
		return i.evalReturnStatement(node)

	case *ast.UsesClause:
		// Uses clauses are processed before execution by the CLI/loader
		// At runtime, they're no-ops since units are already loaded
		return nil

	case *ast.FunctionDecl:
		return i.evalFunctionDeclaration(node)

	case *ast.ClassDecl:
		return i.evalClassDeclaration(node)

	case *ast.InterfaceDecl:
		return i.evalInterfaceDeclaration(node)

	case *ast.OperatorDecl:
		return i.evalOperatorDeclaration(node)

	case *ast.EnumDecl:
		return i.evalEnumDeclaration(node)

	case *ast.SetDecl:
		return i.evalSetDeclaration(node)

	case *ast.RecordDecl:
		return i.evalRecordDeclaration(node)

	case *ast.HelperDecl:
		return i.evalHelperDeclaration(node)

	case *ast.ArrayDecl:
		return i.evalArrayDeclaration(node)

	case *ast.TypeDeclaration:
		return i.evalTypeDeclaration(node)

	// Expressions
	case *ast.IntegerLiteral:
		return &IntegerValue{Value: node.Value}

	case *ast.FloatLiteral:
		return &FloatValue{Value: node.Value}

	case *ast.StringLiteral:
		return &StringValue{Value: node.Value}

	case *ast.BooleanLiteral:
		return &BooleanValue{Value: node.Value}

	case *ast.CharLiteral:
		// Character literals are treated as single-character strings
		return &StringValue{Value: string(node.Value)}

	case *ast.NilLiteral:
		return &NilValue{}

	case *ast.Identifier:
		return i.evalIdentifier(node)

	case *ast.BinaryExpression:
		return i.evalBinaryExpression(node)

	case *ast.UnaryExpression:
		return i.evalUnaryExpression(node)

	case *ast.AddressOfExpression:
		return i.evalAddressOfExpression(node)

	case *ast.GroupedExpression:
		return i.Eval(node.Expression)

	case *ast.CallExpression:
		return i.evalCallExpression(node)

	case *ast.NewExpression:
		return i.evalNewExpression(node)

	case *ast.MemberAccessExpression:
		return i.evalMemberAccess(node)

	case *ast.MethodCallExpression:
		return i.evalMethodCall(node)

	case *ast.InheritedExpression:
		return i.evalInheritedExpression(node)

	case *ast.SelfExpression:
		return i.evalSelfExpression(node)

	case *ast.EnumLiteral:
		return i.evalEnumLiteral(node)

	case *ast.RecordLiteralExpression:
		return i.evalRecordLiteral(node)

	case *ast.SetLiteral:
		return i.evalSetLiteral(node)

	case *ast.ArrayLiteralExpression:
		return i.evalArrayLiteral(node)

	case *ast.IndexExpression:
		return i.evalIndexExpression(node)

	case *ast.NewArrayExpression:
		return i.evalNewArrayExpression(node)

	case *ast.LambdaExpression:
		// Evaluate lambda expression to create closure
		return i.evalLambdaExpression(node)

	case *ast.IsExpression:
		// Task 9.40: Evaluate 'is' type checking operator
		return i.evalIsExpression(node)

	case *ast.AsExpression:
		// Task 9.48: Evaluate 'as' type casting operator
		return i.evalAsExpression(node)

	case *ast.ImplementsExpression:
		// Task 9.48: Evaluate 'implements' interface checking operator
		return i.evalImplementsExpression(node)

	case *ast.IfExpression:
		// Task 9.217: Evaluate inline if-then-else expressions
		return i.evalIfExpression(node)

	case *ast.OldExpression:
		// Evaluate 'old' expressions in postconditions
		identName := node.Identifier.Value
		oldValue, found := i.getOldValue(identName)
		if !found {
			return newError("old value for '%s' not captured (internal error)", identName)
		}
		return oldValue

	default:
		return newError("unknown node type: %T", node)
	}
}

// EvalWithExpectedType evaluates a node with an expected type for better type inference.
// This is primarily used for array literals in function calls where the parameter type is known.
// If expectedType is nil, this falls back to regular Eval().
func (i *Interpreter) EvalWithExpectedType(node ast.Node, expectedType types.Type) Value {
	// Special handling for array literals with expected array type
	if arrayLit, ok := node.(*ast.ArrayLiteralExpression); ok {
		if arrayType, ok := expectedType.(*types.ArrayType); ok {
			return i.evalArrayLiteralWithExpected(arrayLit, arrayType)
		}
	}

	// For all other cases, use regular Eval
	return i.Eval(node)
}
