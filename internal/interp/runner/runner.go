package runner

import (
	"io"

	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	interptypes "github.com/cwbudde/go-dws/internal/interp/types"
)

// New creates a new Interpreter with a fresh global environment.
func New(output io.Writer) *interp.Interpreter {
	return NewWithOptions(output, nil)
}

// NewWithOptions wires up interpreter + evaluator while keeping `internal/interp` free of evaluator imports.
func NewWithOptions(output io.Writer, opts interp.Options) *interp.Interpreter {
	env := interp.NewEnvironment()

	ts := interptypes.NewTypeSystem()
	ts.ClassValueFactory = func(classInfo interptypes.ClassInfo) any {
		if ci, ok := classInfo.(*interp.ClassInfo); ok {
			return &interp.ClassValue{ClassInfo: ci}
		}
		return nil
	}

	maxRecursionDepth := interp.DefaultMaxRecursionDepth
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

	interpreter := interp.NewWithDeps(output, opts, env, ts, eval, refCountMgr)

	// Allow evaluator to delegate OO/decl/exception helpers back to interpreter.
	eval.SetFocusedInterfaces(interpreter, interpreter, interpreter, interpreter)

	return interpreter
}
