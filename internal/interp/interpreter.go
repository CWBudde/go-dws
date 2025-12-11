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

	pkgast "github.com/cwbudde/go-dws/pkg/ast" // SemanticInfo (type annotations)
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
// Thin orchestrator delegating evaluation logic to the Evaluator.
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

// Ensure Interpreter implements the four focused interfaces.
var (
	_ evaluator.OOPEngine        = (*Interpreter)(nil)
	_ evaluator.DeclHandler      = (*Interpreter)(nil)
	_ evaluator.ExceptionManager = (*Interpreter)(nil)
	_ evaluator.CoreEvaluator    = (*Interpreter)(nil)
)

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions creates a new Interpreter with options.
// If options is nil, default options are used.
func NewWithOptions(output io.Writer, opts Options) *Interpreter {
	env := NewEnvironment()

	// Initialize TypeSystem - centralized type registry for classes, records, interfaces, functions, helpers, operators
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

		// Legacy fields for backward compatibility (will be removed after typeSystem migration)
		functions: make(map[string][]*ast.FunctionDecl), // Supports overloading
		classes:   make(map[string]*ClassInfo),
		records:   make(map[string]*RecordTypeValue),
	}

	// Extract recursion depth from options if provided
	if opts != nil {
		// Extract MaxRecursionDepth
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			interp.maxRecursionDepth = depth
		}
	}

	// Initialize execution context with call stack overflow detection.
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

	// Initialize Evaluator with evaluation logic and dependencies
	evalConfig := &evaluator.Config{
		MaxRecursionDepth: interp.maxRecursionDepth,
		SourceCode:        "",
		SourceFile:        "",
	}

	// Create RefCountManager for object lifecycle management
	refCountMgr := runtime.NewRefCountManager()

	// Create evaluator instance (semanticInfo is set later via SetSemanticInfo if needed)
	interp.evaluatorInstance = evaluator.NewEvaluator(
		ts,
		output,
		evalConfig,
		nil,         // unitRegistry is set later via SetUnitRegistry if needed
		nil,         // semanticInfo is set later via SetSemanticInfo if needed
		refCountMgr, // Pass RefCountManager to evaluator
	)

	// Initialize external functions registry in evaluator
	if opts != nil {
		if registry := opts.GetExternalFunctions(); registry != nil {
			interp.evaluatorInstance.SetExternalFunctions(registry)
		}
	}
	if interp.evaluatorInstance.ExternalFunctions() == nil {
		interp.evaluatorInstance.SetExternalFunctions(NewExternalFunctionRegistry())
	}

	// Register destructor callback - invoked when reference count reaches 0
	refCountMgr.SetDestructorCallback(func(obj *runtime.ObjectInstance) error {
		return interp.runDestructorForRefCount(obj)
	})

	// Set focused interfaces so evaluator can delegate to interpreter
	interp.evaluatorInstance.SetFocusedInterfaces(interp, interp, interp, interp)

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

	// Register Variant special values
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
func (i *Interpreter) SetSemanticInfo(info *pkgast.SemanticInfo) {
	if i.evaluatorInstance != nil {
		i.evaluatorInstance.SetSemanticInfo(info)
	}
}

// GetEvaluator returns the evaluator instance.
func (i *Interpreter) GetEvaluator() *evaluator.Evaluator {
	return i.evaluatorInstance
}

// EvalNode implements the evaluator.CoreEvaluator interface.
// Allows Evaluator to delegate back to Interpreter for OOP operations.
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

// ===== Environment Management Helpers =====

// Env returns the current environment from the ExecutionContext.
func (i *Interpreter) Env() *Environment {
	return i.ctx.Env()
}

// SetEnvironment switches the ExecutionContext to use a different environment.
// Used for lambda closure evaluation and environment switching.
func (i *Interpreter) SetEnvironment(env *Environment) {
	i.ctx.SetEnv(env)
}

// PushEnvironment creates a new enclosed environment with the given parent,
// then updates the ExecutionContext to use it. Returns the new environment.
func (i *Interpreter) PushEnvironment(parent *Environment) *Environment {
	newEnv := NewEnclosedEnvironment(parent)
	i.SetEnvironment(newEnv)
	return newEnv
}

// RestoreEnvironment restores a previously saved environment to the ExecutionContext.
func (i *Interpreter) RestoreEnvironment(saved *Environment) {
	i.SetEnvironment(saved)
}

// PushScope creates a new enclosed environment scope using the context's stack.
// Returns a cleanup function that should be deferred to restore the previous scope.
//
// Usage:
//
//	defer i.PushScope()()
//	i.Env().Define("x", value)
//	result := i.Eval(body)
func (i *Interpreter) PushScope() func() {
	i.ctx.PushEnv()
	return func() {
		i.ctx.PopEnv()
	}
}

// pushCallStack adds a new frame to the call stack with the given function name.
func (i *Interpreter) pushCallStack(functionName string) {
	var pos *lexer.Position
	if i.evaluatorInstance.CurrentNode() != nil {
		nodePos := i.evaluatorInstance.CurrentNode().Pos()
		pos = &nodePos
	}
	frame := errors.NewStackFrame(functionName, i.evaluatorInstance.SourceFile(), pos)
	i.callStack = append(i.callStack, frame)
	_ = i.ctx.GetCallStack().Push(functionName, i.evaluatorInstance.SourceFile(), pos)
}

// popCallStack removes the most recent frame from the call stack.
func (i *Interpreter) popCallStack() {
	if len(i.callStack) > 0 {
		i.callStack = i.callStack[:len(i.callStack)-1]
	}
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
// Main entry point for the interpreter.
func (i *Interpreter) Eval(node ast.Node) Value {
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
			// Enrich error with statement location
			// Prevent duplicate stack traces for the same source line
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

		// Auto-invoke parameterless function pointers in statements
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
