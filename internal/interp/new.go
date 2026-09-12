package interp

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp/contracts"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
	"github.com/cwbudde/go-dws/pkg/platform"
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

	maxRecursionDepth := DefaultMaxRecursionDepth
	var hostPlatform platform.Platform
	if opts != nil {
		if depth := opts.GetMaxRecursionDepth(); depth > 0 {
			maxRecursionDepth = depth
		}
		hostPlatform = opts.GetPlatform()
	}
	if hostPlatform == nil {
		hostPlatform = contracts.DefaultPlatform()
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

	eval.EngineState().Platform = hostPlatform

	interpreter := NewWithDeps(output, opts, env, ts, eval, refCountMgr)

	// Wire host-function invocation after the evaluator has prepared runtime values.
	eval.EngineState().ExternalFunctionCaller = func(funcName string, args []Value) Value {
		return interpreter.CallExternalFunction(funcName, args)
	}

	return interpreter
}
