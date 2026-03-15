package interp

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// New creates the interpreter runtime engine with a fully wired internal evaluator.
//
// During Phase 4, Interpreter remains the surviving public engine type while the
// evaluator is treated as an internal implementation detail behind it.
func New(output io.Writer) *Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions creates the interpreter runtime engine with a fully wired
// internal evaluator and the provided options.
func NewWithOptions(output io.Writer, opts Options) *Interpreter {
	env := NewEnvironment()

	ts := interptypes.NewTypeSystem()
	ts.ClassInfoFactory = func(className string) any {
		return NewClassInfo(className)
	}
	ts.ClassValueFactory = func(classInfo interptypes.ClassInfo) any {
		if ci, ok := classInfo.(*ClassInfo); ok {
			return &ClassValue{ClassInfo: ci}
		}
		return nil
	}

	maxRecursionDepth := DefaultMaxRecursionDepth
	if opts != nil {
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			maxRecursionDepth = depth
		}
	}

	evalConfig := &evaluator.Config{
		MaxRecursionDepth: maxRecursionDepth,
	}

	refCountMgr := runtime.NewRefCountManager()
	eval := evaluator.NewEvaluator(
		ts,
		output,
		evalConfig,
		nil,
		nil,
		refCountMgr,
	)

	interpreter := NewWithDeps(output, opts, env, ts, eval, refCountMgr)

	// Wire the ExternalFunctionCaller callback so the evaluator can dispatch external
	// (Go-registered) functions without holding a reference to the interpreter.
	eval.EngineState().ExternalFunctionCaller = func(funcName string, argExprs []ast.Expression, node ast.Node) Value {
		return interpreter.CallExternalFunction(funcName, argExprs, node)
	}

	return interpreter
}
