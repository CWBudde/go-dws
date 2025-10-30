package interp

import (
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/units"
)

// Interpreter executes DWScript AST nodes and manages the runtime environment.
type Interpreter struct {
	currentNode      ast.Node                     // Current AST node being evaluated (for error reporting)
	output           io.Writer                    // Where to write output (e.g., from PrintLn)
	rand             *rand.Rand                   // Random number generator for Random() and Randomize()
	exception        *ExceptionValue              // Current active exception (nil if none)
	interfaces       map[string]*InterfaceInfo    // Registry of interface definitions
	functions        map[string]*ast.FunctionDecl // Registry of user-defined functions
	globalOperators  *runtimeOperatorRegistry     // Global operator overloads
	conversions      *runtimeConversionRegistry   // Global conversions
	env              *Environment                 // The current execution environment
	classes          map[string]*ClassInfo        // Registry of class definitions
	handlerException *ExceptionValue              // Exception being handled (for bare raise)
	callStack        []string                     // Stack of currently executing function names (for stack traces)
	initializedUnits map[string]bool              // Track which units have been initialized
	unitRegistry     *units.UnitRegistry          // Registry for managing loaded units
	loadedUnits      []string                     // Units loaded in order (for initialization/finalization)

	// These flags signal control flow changes (break, continue, exit) and are checked
	// after each statement. They propagate up the call stack until handled by the
	// appropriate control structure (loop for break/continue, function for exit).
	exitSignal     bool // Set by break statement, cleared by loop
	continueSignal bool // Set by continue statement, cleared by loop
	breakSignal    bool // Set by exit statement, cleared by function return

	helpers map[string][]*HelperInfo // Task 9.86-9.87: Helper types (type name -> list of helpers)
}

// New creates a new Interpreter with a fresh global environment.
// The output writer is where built-in functions like PrintLn will write.
func New(output io.Writer) *Interpreter {
	env := NewEnvironment()
	// Initialize random number generator with a default seed
	// Randomize() can be called to re-seed with current time
	source := rand.NewSource(1)
	interp := &Interpreter{
		env:              env,
		output:           output,
		functions:        make(map[string]*ast.FunctionDecl),
		classes:          make(map[string]*ClassInfo),
		interfaces:       make(map[string]*InterfaceInfo), // Task 7.118
		globalOperators:  newRuntimeOperatorRegistry(),
		conversions:      newRuntimeConversionRegistry(),
		rand:             rand.New(source),
		loadedUnits:      make([]string, 0),
		initializedUnits: make(map[string]bool),
	}

	// Register built-in exception classes (Task 8.203-8.204)
	interp.registerBuiltinExceptions()

	// Register built-in array helpers (Task 9.171)
	interp.initArrayHelpers()

	// Initialize ExceptObject to nil
	// ExceptObject is a built-in global variable that holds the current exception
	env.Define("ExceptObject", &NilValue{})

	return interp
}

// GetException returns the current active exception, or nil if none.
// This is used by the CLI to detect and report unhandled exceptions.
func (i *Interpreter) GetException() *ExceptionValue {
	return i.exception
}

// GetCallStack returns a copy of the current call stack.
// Returns function names in the order they were called (oldest to newest).
func (i *Interpreter) GetCallStack() []string {
	// Return a copy to prevent external modification
	stack := make([]string, len(i.callStack))
	copy(stack, i.callStack)
	return stack
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
		return i.Eval(node.Expression)

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

	case *ast.EnumLiteral:
		return i.evalEnumLiteral(node)

	case *ast.RecordLiteralExpression:
		return i.evalRecordLiteral(node)

	case *ast.SetLiteral:
		return i.evalSetLiteral(node)

	case *ast.IndexExpression:
		return i.evalIndexExpression(node)

	case *ast.NewArrayExpression:
		return i.evalNewArrayExpression(node)

	case *ast.LambdaExpression:
		// Evaluate lambda expression to create closure
		return i.evalLambdaExpression(node)

	default:
		return newError("unknown node type: %T", node)
	}
}
