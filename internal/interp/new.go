package interp

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
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
		SourceCode:        "",
		SourceFile:        "",
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

	// Transitional Phase 4 bridge: evaluator still relies on interpreter-owned
	// runtime dispatch and fallback surfaces. Declaration callbacks are no longer
	// part of production construction.
	eval.SetRuntimeBridge(interpreter, interpreter)

	return interpreter
}
