package passes

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
)

// PassContext contains shared state and resources used across all semantic analysis passes.
// It serves as the communication medium between passes, allowing later passes to build
// on the results of earlier passes.
type PassContext struct {
	// ============================================================================
	// Core Registries (Read/Write by all passes)
	// ============================================================================

	// Symbols is the global symbol table for variables, constants, and functions
	Symbols *semantic.SymbolTable

	// TypeRegistry tracks all user-defined and built-in types
	TypeRegistry *semantic.TypeRegistry

	// GlobalOperators manages operator overloading registrations
	GlobalOperators *types.OperatorRegistry

	// ConversionRegistry manages type conversion rules
	ConversionRegistry *types.ConversionRegistry

	// SemanticInfo stores AST annotations and metadata (e.g., resolved types)
	SemanticInfo *pkgast.SemanticInfo

	// ============================================================================
	// Secondary Registries (Read/Write by specific passes)
	// ============================================================================

	// UnitSymbols maps unit names to their symbol tables (for multi-file projects)
	UnitSymbols map[string]*semantic.SymbolTable

	// Helpers maps type names to their helper type extensions
	Helpers map[string][]*types.HelperType

	// Subranges maps subrange type names to their definitions
	Subranges map[string]*types.SubrangeType

	// FunctionPointers maps function pointer type names to their definitions
	FunctionPointers map[string]*types.FunctionPointerType

	// ============================================================================
	// Error Collection (Write by all passes)
	// ============================================================================

	// Errors collects string-formatted error messages (legacy format)
	Errors []string

	// StructuredErrors collects detailed structured error objects
	StructuredErrors []*semantic.SemanticError

	// ============================================================================
	// Pass Execution Context (Read/Write by specific passes)
	// ============================================================================

	// CurrentFunction tracks the function being analyzed (for return validation)
	CurrentFunction interface{} // *ast.FunctionDecl

	// CurrentClass tracks the class being analyzed (for member access validation)
	CurrentClass *types.ClassType

	// CurrentRecord tracks the record being analyzed (for field initialization)
	CurrentRecord *types.RecordType

	// CurrentProperty tracks the property being analyzed (for getter/setter validation)
	CurrentProperty string

	// ============================================================================
	// Source Code Context (Read-only for all passes)
	// ============================================================================

	// SourceCode is the original source text (for error reporting)
	SourceCode string

	// SourceFile is the path to the source file being analyzed
	SourceFile string

	// ============================================================================
	// State Flags (Read/Write by specific passes)
	// ============================================================================

	// LoopDepth tracks nesting level of loops (for break/continue validation)
	LoopDepth int

	// InExceptionHandler indicates if we're inside a try/except block
	InExceptionHandler bool

	// InFinallyBlock indicates if we're inside a finally block
	InFinallyBlock bool

	// InLoop indicates if we're inside any loop construct
	InLoop bool

	// InLambda indicates if we're inside a lambda/anonymous function
	InLambda bool

	// InClassMethod indicates if we're inside a class method
	InClassMethod bool

	// InPropertyExpr indicates if we're inside a property expression
	InPropertyExpr bool
}

// NewPassContext creates a new pass context with initialized registries.
func NewPassContext() *PassContext {
	return &PassContext{
		Symbols:            semantic.NewSymbolTable(),
		TypeRegistry:       semantic.NewTypeRegistry(),
		UnitSymbols:        make(map[string]*semantic.SymbolTable),
		Errors:             make([]string, 0),
		StructuredErrors:   make([]*semantic.SemanticError, 0),
		Subranges:          make(map[string]*types.SubrangeType),
		FunctionPointers:   make(map[string]*types.FunctionPointerType),
		Helpers:            make(map[string][]*types.HelperType),
		GlobalOperators:    types.NewOperatorRegistry(),
		ConversionRegistry: types.NewConversionRegistry(),
		SemanticInfo:       pkgast.NewSemanticInfo(),
	}
}

// AddError adds a formatted error message to the error list.
func (ctx *PassContext) AddError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ctx.Errors = append(ctx.Errors, msg)
}

// AddStructuredError adds a structured error to the error list.
func (ctx *PassContext) AddStructuredError(err *semantic.SemanticError) {
	ctx.StructuredErrors = append(ctx.StructuredErrors, err)
}

// HasErrors returns true if any errors have been collected.
func (ctx *PassContext) HasErrors() bool {
	return len(ctx.Errors) > 0 || len(ctx.StructuredErrors) > 0
}

// HasCriticalErrors returns true if any critical (non-hint, non-warning) errors exist.
func (ctx *PassContext) HasCriticalErrors() bool {
	// Check string errors (exclude hints)
	for _, err := range ctx.Errors {
		if !strings.HasPrefix(err, "Hint:") {
			return true
		}
	}

	// Check structured errors (exclude warnings)
	for _, err := range ctx.StructuredErrors {
		if !err.IsWarning() {
			return true
		}
	}

	return false
}

// ErrorCount returns the total number of errors (including warnings and hints).
func (ctx *PassContext) ErrorCount() int {
	return len(ctx.Errors) + len(ctx.StructuredErrors)
}

// CriticalErrorCount returns the number of critical errors (excluding warnings and hints).
func (ctx *PassContext) CriticalErrorCount() int {
	count := 0

	// Count non-hint string errors
	for _, err := range ctx.Errors {
		if !strings.HasPrefix(err, "Hint:") {
			count++
		}
	}

	// Count non-warning structured errors
	for _, err := range ctx.StructuredErrors {
		if !err.IsWarning() {
			count++
		}
	}

	return count
}
