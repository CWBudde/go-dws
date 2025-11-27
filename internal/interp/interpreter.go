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
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/internal/units"
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
