package semantic

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	pkgast "github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// BuiltinChecker provides an interface for checking and analyzing built-in functions.
// This allows the passes to validate built-in function calls without direct coupling to the Analyzer.
type BuiltinChecker interface {
	// AnalyzeBuiltin analyzes a built-in function call.
	// Returns (resultType, true) if the function is a recognized built-in,
	// or (nil, false) if it's not a built-in function.
	AnalyzeBuiltin(name string, args []pkgast.Expression, callExpr *pkgast.CallExpression) (types.Type, bool)

	// IsBuiltinFunction checks if a function name is a built-in function.
	// Task 6.1.2.1: This method checks if a function is built-in WITHOUT analyzing arguments.
	// Returns (returnType, true) if the function is a recognized built-in,
	// or (nil, false) if it's not a built-in function.
	// Use this when you only need to check if something is a built-in and get its return type,
	// without triggering argument analysis (which would use the wrong scope).
	IsBuiltinFunction(name string) (types.Type, bool)
}

// ScopeKind identifies the type of scope (for debugging and special handling)
type ScopeKind int

const (
	// ScopeGlobal is the outermost scope containing global declarations
	ScopeGlobal ScopeKind = iota
	// ScopeFunction is a function/procedure/method scope
	ScopeFunction
	// ScopeBlock is a nested block scope (begin/end, if/else, loops, etc.)
	ScopeBlock
)

// Scope represents a lexical scope for symbol resolution.
// Scopes are organized in a parent chain, allowing inner scopes to access
// symbols from outer scopes while shadowing is permitted.
type Scope struct {
	// Symbols maps normalized identifier names to their types
	// Names are normalized using ident.Normalize for case-insensitive lookup
	Symbols map[string]types.Type

	// Parent is the enclosing scope (nil for global scope)
	Parent *Scope

	// Kind identifies the type of scope (global, function, block)
	Kind ScopeKind
}

// NewScope creates a new scope with the given kind and parent.
func NewScope(kind ScopeKind, parent *Scope) *Scope {
	return &Scope{
		Kind:    kind,
		Symbols: make(map[string]types.Type),
		Parent:  parent,
	}
}

// Define adds a symbol to this scope.
// The name is normalized for case-insensitive lookup.
// ident.Normalize converts the name to lowercase for normalization.
func (s *Scope) Define(name string, typ types.Type) {
	normalized := ident.Normalize(name)
	s.Symbols[normalized] = typ
}

// Lookup searches for a symbol in this scope only (does not check parent).
// Returns (type, true) if found, (nil, false) otherwise.
// The name is normalized for case-insensitive lookup.
func (s *Scope) Lookup(name string) (types.Type, bool) {
	normalized := ident.Normalize(name)
	typ, found := s.Symbols[normalized]
	return typ, found
}

// LookupChain searches for a symbol in this scope and all parent scopes.
// Returns (type, true) if found in any scope, (nil, false) otherwise.
// The name is normalized for case-insensitive lookup.
func (s *Scope) LookupChain(name string) (types.Type, bool) {
	// Search current scope
	if typ, found := s.Lookup(name); found {
		return typ, true
	}

	// Search parent scopes
	if s.Parent != nil {
		return s.Parent.LookupChain(name)
	}

	return nil, false
}

// PassContext contains shared state and resources used across all semantic analysis passes.
// It serves as the communication medium between passes, allowing later passes to build
// on the results of earlier passes.
type PassContext struct {
	BuiltinChecker     BuiltinChecker                        // Built-in function analyzer
	CurrentFunction    any                                   // Current function (*ast.FunctionDecl)
	UnitSymbols        map[string]*SymbolTable               // Unit symbol tables
	ConversionRegistry *types.ConversionRegistry             // Type conversion registry
	SemanticInfo       *pkgast.SemanticInfo                  // AST annotations
	GlobalOperators    *types.OperatorRegistry               // Operator overload registry
	TypeRegistry       *TypeRegistry                         // Type registry
	Helpers            map[string][]*types.HelperType        // Helper type registry
	Subranges          map[string]*types.SubrangeType        // Subrange type registry
	FunctionPointers   map[string]*types.FunctionPointerType // Function pointer type registry
	Symbols            *SymbolTable                          // Global symbol table
	CurrentRecord      *types.RecordType                     // Current record being analyzed
	CurrentClass       *types.ClassType                      // Current class being analyzed
	SourceFile         string                                // Source file path
	CurrentForLoopVar  string                                // Current for loop variable
	SourceCode         string                                // Original source text
	CurrentProperty    string                                // Current property being analyzed
	StructuredErrors   []*SemanticError                      // Structured error objects
	Errors             []string                              // Error messages (legacy)
	ScopeStack         []*Scope                              // Scope chain for local variables
	LoopDepth          int                                   // Loop nesting level
	InExceptionHandler bool                                  // Inside try/except block
	InFinallyBlock     bool                                  // Inside finally block
	InLoop             bool                                  // Inside any loop construct
	InLambda           bool                                  // Inside lambda/anonymous function
	InClassMethod      bool                                  // Inside class method
	InPropertyExpr     bool                                  // Inside property expression
}

// NewPassContext creates a new pass context with initialized registries.
func NewPassContext() *PassContext {
	// Create the global scope as the root of the scope chain
	globalScope := NewScope(ScopeGlobal, nil)

	return &PassContext{
		Symbols:            NewSymbolTable(),
		TypeRegistry:       NewTypeRegistry(),
		UnitSymbols:        make(map[string]*SymbolTable),
		Errors:             make([]string, 0),
		StructuredErrors:   make([]*SemanticError, 0),
		Subranges:          make(map[string]*types.SubrangeType),
		FunctionPointers:   make(map[string]*types.FunctionPointerType),
		Helpers:            make(map[string][]*types.HelperType),
		GlobalOperators:    types.NewOperatorRegistry(),
		ConversionRegistry: types.NewConversionRegistry(),
		SemanticInfo:       pkgast.NewSemanticInfo(),
		ScopeStack:         []*Scope{globalScope}, // Initialize with global scope
	}
}

// AddError adds a formatted error message to the error list.
func (ctx *PassContext) AddError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ctx.Errors = append(ctx.Errors, msg)
}

// AddStructuredError adds a structured error to the error list.
func (ctx *PassContext) AddStructuredError(err *SemanticError) {
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

// ============================================================================
// Scope Management Methods
// ============================================================================

// PushScope creates and pushes a new scope onto the scope stack.
// The new scope becomes the current scope and has the previous current scope as its parent.
func (ctx *PassContext) PushScope(kind ScopeKind) {
	parent := ctx.CurrentScope()
	newScope := NewScope(kind, parent)
	ctx.ScopeStack = append(ctx.ScopeStack, newScope)
}

// PopScope removes the current scope from the scope stack.
// This should be called when exiting a function, method, or block.
// Panics if attempting to pop the global scope.
func (ctx *PassContext) PopScope() {
	if len(ctx.ScopeStack) <= 1 {
		panic("cannot pop global scope")
	}
	ctx.ScopeStack = ctx.ScopeStack[:len(ctx.ScopeStack)-1]
}

// CurrentScope returns the current (innermost) scope.
// This is always valid as the global scope is always present.
func (ctx *PassContext) CurrentScope() *Scope {
	if len(ctx.ScopeStack) == 0 {
		panic("scope stack is empty")
	}
	return ctx.ScopeStack[len(ctx.ScopeStack)-1]
}

// LookupInScopes searches for a symbol in the current scope and all parent scopes.
// Returns (type, true) if found, (nil, false) otherwise.
// This is the primary method for resolving identifiers during semantic analysis.
func (ctx *PassContext) LookupInScopes(name string) (types.Type, bool) {
	return ctx.CurrentScope().LookupChain(name)
}

// DefineInCurrentScope adds a symbol to the current scope.
// This is used when declaring local variables, parameters, or constants.
func (ctx *PassContext) DefineInCurrentScope(name string, typ types.Type) {
	ctx.CurrentScope().Define(name, typ)
}

// GlobalScope returns the global (outermost) scope.
// This is always at index 0 of the scope stack.
func (ctx *PassContext) GlobalScope() *Scope {
	if len(ctx.ScopeStack) == 0 {
		panic("scope stack is empty")
	}
	return ctx.ScopeStack[0]
}
