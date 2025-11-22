package semantic

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Pass represents a single semantic analysis pass.
// The multi-pass architecture allows for:
// - Proper handling of forward declarations
// - Clear separation of concerns
// - Incremental analysis and caching
// - Better error messages with complete context
type Pass interface {
	// Name returns the name of this pass for logging and debugging
	Name() string

	// Run executes this pass on the given program.
	// The pass should:
	// - Read and write to the shared PassContext
	// - Collect any errors in the context's error list
	// - NOT modify the AST structure (only annotate it)
	// Returns an error only for fatal internal errors (not semantic errors)
	Run(program *ast.Program, ctx *PassContext) error
}

// PassManager coordinates the execution of multiple passes in order.
type PassManager struct {
	passes []Pass
}

// NewPassManager creates a new pass manager with the given passes.
// Passes will be executed in the order they are provided.
func NewPassManager(passes ...Pass) *PassManager {
	return &PassManager{
		passes: passes,
	}
}

// RunAll executes all passes in sequence.
// If any pass returns an error, execution stops and the error is returned.
// Semantic errors are collected in the PassContext, not returned as errors.
func (pm *PassManager) RunAll(program *ast.Program, ctx *PassContext) error {
	for _, pass := range pm.passes {
		if err := pass.Run(program, ctx); err != nil {
			return err
		}
		// If critical errors were found, stop processing
		if ctx.HasCriticalErrors() {
			break
		}
	}
	return nil
}

// AddPass adds a pass to the manager.
// The pass will be executed after all previously added passes.
func (pm *PassManager) AddPass(pass Pass) {
	pm.passes = append(pm.passes, pass)
}

// Passes returns the list of registered passes.
func (pm *PassManager) Passes() []Pass {
	return pm.passes
}
