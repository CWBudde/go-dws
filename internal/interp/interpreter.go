package interp

import (
	"io"
	"math"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/contracts"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// DefaultMaxRecursionDepth is the default maximum recursion depth for function calls.
// This matches DWScript's default limit (see dwsCompiler.pas:39 cDefaultMaxRecursionDepth).
// When the call stack reaches this depth, the interpreter raises an EScriptStackOverflow exception
// to prevent infinite recursion and potential Go runtime stack overflow.
const DefaultMaxRecursionDepth = 1024

// PropertyEvalContext tracks the state during property getter/setter evaluation.
type PropertyEvalContext = runtime.PropertyEvalContext

type evaluatorShim interface {
	Eval(node ast.Node, ctx *runtime.ExecutionContext) Value
	ExecuteUserFunction(fn *ast.FunctionDecl, args []Value, ctx *runtime.ExecutionContext, callbacks *contracts.UserFunctionCallbacks) (Value, error)
	CurrentNode() ast.Node
	CurrentContext() *runtime.ExecutionContext
	EngineState() *contracts.EngineState
	SetCurrentNode(node ast.Node)
}

// Interpreter executes DWScript AST nodes and manages the runtime environment.
// Thin orchestrator delegating evaluation logic to the Evaluator.
type Interpreter struct {
	output            io.Writer
	engineState       *contracts.EngineState
	typeSystem        *interptypes.TypeSystem
	methodRegistry    *runtime.MethodRegistry
	evaluatorInstance evaluatorShim
	ctx               *runtime.ExecutionContext
}

// NewWithDeps creates an Interpreter with its core dependencies provided by a higher-level runner.
// This avoids `internal/interp` importing `internal/interp/evaluator`.
func NewWithDeps(
	output io.Writer,
	opts Options,
	env *runtime.Environment,
	typeSystem *interptypes.TypeSystem,
	eval evaluatorShim,
	refCountMgr runtime.RefCountManager,
) *Interpreter {
	interp := &Interpreter{
		output:            output,
		engineState:       eval.EngineState(),
		typeSystem:        typeSystem,
		methodRegistry:    runtime.NewMethodRegistry(),
		evaluatorInstance: eval,
	}
	interp.engineState.MethodRegistry = interp.methodRegistry

	if opts != nil {
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			interp.engineState.MaxRecursionDepth = depth
		}
	}
	if interp.engineState.MaxRecursionDepth <= 0 {
		interp.engineState.MaxRecursionDepth = DefaultMaxRecursionDepth
	}

	interp.ctx = runtime.NewExecutionContextWithMaxDepth(env, interp.engineState.MaxRecursionDepth)
	interp.ctx.SetRefCountManager(refCountMgr)

	if interp.evaluatorInstance != nil {
		if opts != nil {
			if registry := opts.GetExternalFunctions(); registry != nil {
				interp.engineState.ExternalFunctions = registry
			}
		}
		if interp.engineState.ExternalFunctions == nil {
			interp.engineState.ExternalFunctions = NewExternalFunctionRegistry()
		}
	}

	refCountMgr.SetDestructorCallback(func(obj *runtime.ObjectInstance) error {
		return interp.runDestructorForRefCount(obj)
	})

	interp.registerBuiltinExceptions()
	interp.registerBuiltinInterfaces()
	interp.initArrayHelpers()
	interp.initIntrinsicHelpers()
	interp.initEnumHelpers()

	env.Define("ExceptObject", &NilValue{})
	env.Define("Integer", NewTypeMetaValue(types.INTEGER, "Integer"))
	env.Define("Float", NewTypeMetaValue(types.FLOAT, "Float"))
	env.Define("String", NewTypeMetaValue(types.STRING, "String"))
	env.Define("Boolean", NewTypeMetaValue(types.BOOLEAN, "Boolean"))
	env.Define("PI", &FloatValue{Value: math.Pi})
	env.Define("NaN", &FloatValue{Value: math.NaN()})
	env.Define("Infinity", &FloatValue{Value: math.Inf(1)})
	env.Define("Null", NewNullValue())
	env.Define("Unassigned", NewUnassignedValue())

	return interp
}

// GetException returns the current active exception, or nil if none.
// This is used by the CLI to detect and report unhandled exceptions.
func (i *Interpreter) GetException() *runtime.ExceptionValue {
	exc, _ := i.ctx.Exception().(*runtime.ExceptionValue)
	return exc
}

// SetSemanticInfo sets the semantic metadata table for this interpreter.
// The semantic info contains type annotations and symbol resolutions from analysis.
func (i *Interpreter) SetSemanticInfo(info *ast.SemanticInfo) {
	i.engineState.SemanticInfo = info
}

// GetCallStack returns a copy of the current call stack.
// Returns stack frames in the order they were called (oldest to newest).
func (i *Interpreter) GetCallStack() errors.StackTrace {
	return i.ctx.CallStack()
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
	_ = i.ctx.GetCallStack().Push(functionName, i.sourceFile(), pos)
}

// popCallStack removes the most recent frame from the call stack.
func (i *Interpreter) popCallStack() {
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
	return i.evalViaEvaluator(node)
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
