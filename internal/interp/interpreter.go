package interp

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"

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
	output            io.Writer
	exception         *runtime.ExceptionValue
	handlerException  *runtime.ExceptionValue
	propContext       *PropertyEvalContext
	typeSystem        *interptypes.TypeSystem
	methodRegistry    *runtime.MethodRegistry
	records           map[string]*RecordTypeValue
	functions         map[string][]*ast.FunctionDecl
	evaluatorInstance *evaluator.Evaluator
	classes           map[string]*ClassInfo
	ctx               *evaluator.ExecutionContext
	oldValuesStack    []map[string]Value
	callStack         errors.StackTrace
	maxRecursionDepth int
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

	// Initialize ClassValueFactory to enable evaluator to create ClassValue
	ts.ClassValueFactory = func(classInfo interptypes.ClassInfo) any {
		if ci, ok := classInfo.(*ClassInfo); ok {
			return &ClassValue{ClassInfo: ci}
		}
		return nil
	}

	interp := &Interpreter{
		output:            output,
		maxRecursionDepth: DefaultMaxRecursionDepth,
		callStack:         errors.NewStackTrace(), // Initialize stack trace

		// TypeSystem (new centralized type registry)
		// This is the modern API - use this for new code
		typeSystem: ts,

		// MethodRegistry for AST-free method storage
		methodRegistry: runtime.NewMethodRegistry(),

		// Phase 3.4.1: Legacy fields for backward compatibility
		// These will be removed once migration to typeSystem is complete
		functions: make(map[string][]*ast.FunctionDecl), // Task 9.66: Support overloading
		classes:   make(map[string]*ClassInfo),
		records:   make(map[string]*RecordTypeValue),
	}

	// Task 3.8.1: Extract recursion depth from options using interface
	// This replaces the reflection hack with a clean interface-based approach
	if opts != nil {
		// Extract MaxRecursionDepth
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			interp.maxRecursionDepth = depth
		}
	}

	// Phase 3.3.1/3.3.3: Initialize execution context with call stack overflow detection
	// The context wraps the environment with a simple adapter to avoid circular dependencies.
	// The Environment is passed as interface{} to ExecutionContext to avoid import cycles.
	// Phase 3.3.3: Pass maxRecursionDepth to configure CallStack overflow detection.
	// Phase 3.1.3: Direct runtime.Environment - no adapter needed.
	// Phase 3.2: Use callbacks to unify exception handling between interpreter and evaluator.
	// Note: Getter must return untyped nil when no exception, to avoid Go's interface nil gotcha.
	interp.ctx = evaluator.NewExecutionContextWithCallbacks(
		env,
		interp.maxRecursionDepth,
		func() any { // getter: read from i.exception
			if interp.exception == nil {
				return nil // Return untyped nil, not typed nil pointer
			}
			return interp.exception
		},
		func(exc any) { // setter: write to i.exception
			if exc == nil {
				interp.exception = nil
			} else if excVal, ok := exc.(*runtime.ExceptionValue); ok {
				interp.exception = excVal
			}
		},
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

	// Task 3.5.41: Create RefCountManager for object lifecycle management
	refCountMgr := runtime.NewRefCountManager()

	// Create evaluator instance
	// Task 3.3.4: semanticInfo is nil initially, set via SetSemanticInfo if needed
	interp.evaluatorInstance = evaluator.NewEvaluator(
		ts,
		output,
		evalConfig,
		nil, // unitRegistry is set later via SetUnitRegistry if needed
		nil, // semanticInfo is set later via SetSemanticInfo if needed
		refCountMgr, // Task 3.5.41: Pass RefCountManager to evaluator
	)

	// Task 3.3.6: Initialize external functions registry in evaluator
	// Extract from options if provided, otherwise create new registry
	if opts != nil {
		if registry := opts.GetExternalFunctions(); registry != nil {
			interp.evaluatorInstance.SetExternalFunctions(registry)
		}
	}
	if interp.evaluatorInstance.ExternalFunctions() == nil {
		interp.evaluatorInstance.SetExternalFunctions(NewExternalFunctionRegistry())
	}

	// Task 3.5.41: Register destructor callback for RefCountManager
	// This callback is invoked when an object's reference count reaches 0
	refCountMgr.SetDestructorCallback(func(obj *runtime.ObjectInstance) error {
		return interp.runDestructorForRefCount(obj)
	})

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
func (i *Interpreter) GetException() *runtime.ExceptionValue {
	return i.exception
}

// SetSemanticInfo sets the semantic metadata table for this interpreter.
// The semantic info contains type annotations and symbol resolutions from analysis.
// SetSemanticInfo sets the semantic analysis metadata.
// Task 3.3.4: Now only updates evaluator (Interpreter no longer stores this field).
func (i *Interpreter) SetSemanticInfo(info *pkgast.SemanticInfo) {
	if i.evaluatorInstance != nil {
		i.evaluatorInstance.SetSemanticInfo(info)
	}
}

// GetEvaluator returns the evaluator instance.
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

// GetCallStack returns a copy of the current call stack.
// Returns stack frames in the order they were called (oldest to newest).
func (i *Interpreter) GetCallStack() errors.StackTrace {
	// Return a copy to prevent external modification
	stack := make(errors.StackTrace, len(i.callStack))
	copy(stack, i.callStack)
	return stack
}

// ===== Environment Management Helpers (Phase 3.8.1: Task 3.8.1.2) =====

// Env returns the current environment from the ExecutionContext.
// Phase 3.1.5: This is the canonical way to access the environment.
// All code should use i.Env() instead of i.env.
func (i *Interpreter) Env() *Environment {
	return i.ctx.Env()
}

// SetEnvironment switches the ExecutionContext to use a different environment.
// This is used for cases like lambda closure evaluation where we need to
// switch to a captured environment, evaluate, then switch back.
//
// Note: For scope pushing (creating new enclosed scopes), use PushScope() instead.
// SetEnvironment is specifically for environment switching, not scope management.
//
// Phase 3.1.5: Only updates ctx.env - i.env field is removed.
func (i *Interpreter) SetEnvironment(env *Environment) {
	i.ctx.SetEnv(env)
}

// PushEnvironment creates a new enclosed environment with the given parent environment,
// then updates the ExecutionContext to use the new environment.
// Returns the new environment.
//
// Note: For most scope management, prefer PushScope() which uses defer for cleanup.
// PushEnvironment is retained for special cases like closure evaluation where
// the parent environment is not necessarily the current environment.
//
// Phase 3.1.5: Only updates ctx.env - i.env field is removed.
func (i *Interpreter) PushEnvironment(parent *Environment) *Environment {
	newEnv := NewEnclosedEnvironment(parent)
	i.SetEnvironment(newEnv)
	return newEnv
}

// RestoreEnvironment restores a previously saved environment to the ExecutionContext.
// This is used to exit scopes after PushEnvironment.
//
// Phase 3.1.5: Only updates ctx.env - i.env field is removed.
func (i *Interpreter) RestoreEnvironment(saved *Environment) {
	i.SetEnvironment(saved)
}

// PushScope creates a new enclosed environment scope using the context's stack.
// Returns a cleanup function that should be deferred to restore the previous scope.
// This replaces the manual savedEnv/PushEnvironment/RestoreEnvironment pattern.
//
// Usage:
//
//	defer i.PushScope()()
//	i.Env().Define("x", value)
//	result := i.Eval(body)
//
// Phase 3.1.4: Unified scope management
// Phase 3.1.5: No i.env field - uses ctx.Env() exclusively.
func (i *Interpreter) PushScope() func() {
	i.ctx.PushEnv()
	return func() {
		i.ctx.PopEnv()
	}
}

// pushCallStack adds a new frame to the call stack with the given function name.
// The position is taken from the current node being evaluated.
// Phase 3.3.3: Delegates to ExecutionContext's CallStack.
func (i *Interpreter) pushCallStack(functionName string) {
	var pos *lexer.Position
	if i.evaluatorInstance.CurrentNode() != nil {
		nodePos := i.evaluatorInstance.CurrentNode().Pos()
		pos = &nodePos
	}
	// Also push to the old callStack field for backward compatibility
	frame := errors.NewStackFrame(functionName, i.evaluatorInstance.SourceFile(), pos)
	i.callStack = append(i.callStack, frame)

	// Phase 3.3.3: Also push to context's CallStack
	// Ignore errors here for backward compatibility; callers should check WillOverflow first
	_ = i.ctx.GetCallStack().Push(functionName, i.evaluatorInstance.SourceFile(), pos)
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

// evalViaEvaluator delegates evaluation to the evaluator and converts runtime.ErrorValue to interp.ErrorValue.
// This ensures type compatibility for tests and existing code that expect interp.ErrorValue.
func (i *Interpreter) evalViaEvaluator(node ast.Node) Value {
	result := i.evaluatorInstance.Eval(node, i.ctx)
	// Convert runtime.ErrorValue to interp.ErrorValue for type compatibility
	if runtimeErr, ok := result.(*runtime.ErrorValue); ok {
		return &ErrorValue{Message: runtimeErr.Message}
	}
	return result
}

// Eval evaluates an AST node and returns its value.
// This is the main entry point for the interpreter.
func (i *Interpreter) Eval(node ast.Node) Value {
	// Track the current node for error reporting (Task 3.3.4: now in evaluator)
	i.evaluatorInstance.SetCurrentNode(node)

	switch node := node.(type) {
	// Program - KEEP: orchestrates exception flow via i.exception
	case *ast.Program:
		return i.evalProgram(node)

	case *ast.EmptyStatement:
		return i.evalViaEvaluator(node)

	// Statements - KEEP in interpreter: exception/control flow uses i.exception
	case *ast.ExpressionStatement:
		// Evaluate the expression
		val := i.Eval(node.Expression)
		if isError(val) {
			// Enrich error with statement location to mimic DWScript call stack output
			// Task 3.8.3.0i: Prevent duplicate stack traces for the same source line
			if errVal, ok := val.(*ErrorValue); ok {
				exprPos := node.Expression.Pos()
				// Check for both "line N" and "[line: N," formats to handle all cases
				lineMarker1 := fmt.Sprintf("line %d", exprPos.Line)
				lineMarker2 := fmt.Sprintf("[line: %d,", exprPos.Line)
				loc := fmt.Sprintf("at line %d, column: %d", exprPos.Line, exprPos.Column+2)
				// Don't add stack trace if this line is already present in any format
				if !strings.Contains(errVal.Message, lineMarker1) && !strings.Contains(errVal.Message, lineMarker2) {
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
		// KEEP: Has complex static array initialization logic
		return i.evalVarDeclStatement(node)

	case *ast.ConstDecl:
		// KEEP: May have type-specific initialization
		return i.evalConstDecl(node)

	case *ast.AssignmentStatement:
		return i.evalAssignmentStatement(node)

	case *ast.BlockStatement:
		// KEEP: Block statements need interpreter's control flow handling
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
		// KEEP: Control flow needs interpreter handling
		return i.evalBreakStatement(node)

	case *ast.ContinueStatement:
		// KEEP: Control flow needs interpreter handling
		return i.evalContinueStatement(node)

	case *ast.ExitStatement:
		// KEEP: Control flow needs interpreter handling
		return i.evalExitStatement(node)

	case *ast.ReturnStatement:
		// KEEP: Control flow needs interpreter handling
		return i.evalReturnStatement(node)

	case *ast.UsesClause:
		// Uses clauses are processed before execution by the CLI/loader
		return nil

	case *ast.FunctionDecl:
		// KEEP: Has function registry interactions
		return i.evalFunctionDeclaration(node)

	case *ast.ClassDecl:
		// KEEP: Complex type system registration
		return i.evalClassDeclaration(node)

	case *ast.InterfaceDecl:
		// KEEP: Interface registration ordering
		return i.evalInterfaceDeclaration(node)

	case *ast.OperatorDecl:
		// KEEP: Operator registry interactions
		return i.evalOperatorDeclaration(node)

	case *ast.EnumDecl:
		// KEEP: Enum type registration
		return i.evalEnumDeclaration(node)

	case *ast.SetDecl:
		// Evaluator now handles this (no adapter fallback needed)
		return i.evalViaEvaluator(node)

	case *ast.RecordDecl:
		return i.evalViaEvaluator(node)

	case *ast.HelperDecl:
		// KEEP: Helper type registration
		return i.evalHelperDeclaration(node)

	case *ast.ArrayDecl:
		return i.evalViaEvaluator(node)

	case *ast.TypeDeclaration:
		// KEEP: Type alias registration
		return i.evalTypeDeclaration(node)

	// Expressions - Literals (delegate to evaluator)
	case *ast.IntegerLiteral:
		return i.evalViaEvaluator(node)

	case *ast.FloatLiteral:
		return i.evalViaEvaluator(node)

	case *ast.StringLiteral:
		return i.evalViaEvaluator(node)

	case *ast.BooleanLiteral:
		return i.evalViaEvaluator(node)

	case *ast.CharLiteral:
		return i.evalViaEvaluator(node)

	case *ast.NilLiteral:
		return i.evalViaEvaluator(node)

	case *ast.Identifier:
		return i.evalViaEvaluator(node)

	case *ast.BinaryExpression:
		return i.evalViaEvaluator(node)

	case *ast.UnaryExpression:
		return i.evalViaEvaluator(node)

	case *ast.AddressOfExpression:
		// KEEP: Method pointer creation requires interpreter state
		return i.evalAddressOfExpression(node)

	case *ast.GroupedExpression:
		return i.evalViaEvaluator(node)

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
		return i.evalViaEvaluator(node)

	case *ast.RecordLiteralExpression:
		// KEEP: Field validation and type context handling
		return i.evalRecordLiteral(node)

	case *ast.SetLiteral:
		return i.evalViaEvaluator(node)

	case *ast.ArrayLiteralExpression:
		// KEEP: Type context from assignment target
		return i.evalArrayLiteral(node)

	case *ast.IndexExpression:
		// KEEP: Has complex property access logic
		return i.evalIndexExpression(node)

	case *ast.NewArrayExpression:
		// KEEP: Array creation needs interpreter type lookup
		return i.evalNewArrayExpression(node)

	case *ast.LambdaExpression:
		return i.evalViaEvaluator(node)

	case *ast.IsExpression:
		return i.evalViaEvaluator(node)

	case *ast.AsExpression:
		// KEEP: Has complex type cast logic that may need interpreter state
		return i.evalAsExpression(node)

	case *ast.ImplementsExpression:
		return i.evalViaEvaluator(node)

	case *ast.IfExpression:
		return i.evalViaEvaluator(node)

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
