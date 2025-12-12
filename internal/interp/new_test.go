package interp

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
)

// New creates a fully-wired Interpreter for tests.
//
// Production code should use internal/interp/runner (or pkg/dwscript), but many
// internal/interp package tests rely on a convenient constructor.
func New(output io.Writer) *Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions creates a fully-wired Interpreter for tests with options.
func NewWithOptions(output io.Writer, opts Options) *Interpreter {
	env := NewEnvironment()

	ts := interptypes.NewTypeSystem()
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

	// Allow evaluator to delegate OO/decl/exception helpers back to interpreter.
	eval.SetFocusedInterfaces(interpreter, interpreter, interpreter, interpreter)

	return interpreter
}
