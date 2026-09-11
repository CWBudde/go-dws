package ast

import (
	"sync"

	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Semantic Metadata Architecture
// ============================================================================
//
// This file implements a separate metadata table for storing semantic analysis
// results. This design separates parsing from semantic analysis, making the AST
// immutable after parsing and enabling multiple concurrent semantic analyses.
//
// **Design Rationale**:
//
// Previously, every expression node carried a `Type *TypeAnnotation` field that
// was nil during parsing and populated during semantic analysis. This approach
// had several drawbacks:
//
// 1. **Memory overhead**: Every expression node allocated ~16 bytes for the Type
//    field, even though it was nil during parsing.
// 2. **Coupling**: The AST was tightly coupled to the semantic analyzer.
// 3. **Mutability**: The AST was modified during analysis, preventing reuse.
// 4. **Concurrency**: Multiple concurrent analyses on the same AST were not safe.
//
// **New Architecture**:
//
// The new design uses a separate SemanticInfo table that maps AST nodes to their
// semantic information (types, folded compile-time values, etc.). This provides:
//
// 1. **Separation of concerns**: Parsing produces immutable AST; analysis produces
//    separate metadata.
// 2. **Memory efficiency**: Type information only allocated for nodes that need it.
// 3. **Reusability**: Same AST can be analyzed multiple times with different contexts.
// 4. **Concurrency**: Multiple read-only analyses safe; separate SemanticInfo per
//    analysis ensures isolation.
//
// **Architecture Components**:
//
// - SemanticInfo: Main metadata table, thread-safe, maps nodes to semantic results
// - Expression → *TypeAnnotation: Maps expression nodes to their inferred types
// - *Identifier → bool: Caches compile-time predicate results per call site
//
// **Thread Safety**:
//
// SemanticInfo uses sync.RWMutex for concurrent access:
// - Multiple readers can query types simultaneously (GetType)
// - Single writer during analysis (SetType)
// - Each analysis gets its own SemanticInfo instance
//
// **Migration Path**:
//
// 1. Create SemanticInfo and API (this file)
// 2. Remove Type field from AST nodes
// 3. Update semantic analyzer to use SemanticInfo
// 4. Update interpreter to use SemanticInfo
// 5. Update bytecode compiler to use SemanticInfo
// 6. Update public API to return SemanticInfo
//
// ============================================================================

// SemanticInfo holds semantic analysis results for an AST.
// It maps AST nodes to their inferred types, folded compile-time predicate
// results, and other semantic information. This separation allows the AST to
// remain immutable after parsing while still supporting multiple semantic
// analyses.
//
// Thread Safety: SemanticInfo is safe for concurrent reads but not concurrent
// writes. Typical usage is single-threaded analysis (writes) followed by
// concurrent interpretation/compilation (reads).
type SemanticInfo struct {
	resolvedTypes    map[Node]types.Type
	types            map[Expression]*TypeAnnotation
	foldedPredicates map[*Identifier]bool
	mu               sync.RWMutex
}

// NewSemanticInfo creates a new empty semantic metadata table.
// Each semantic analysis should create its own SemanticInfo instance.
func NewSemanticInfo() *SemanticInfo {
	return &SemanticInfo{
		resolvedTypes:    make(map[Node]types.Type),
		types:            make(map[Expression]*TypeAnnotation),
		foldedPredicates: make(map[*Identifier]bool),
	}
}

// ============================================================================
// Type Information API
// ============================================================================

// GetType returns the inferred type annotation for an expression node.
// Returns nil if no type has been set for this node.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) GetType(expr Expression) *TypeAnnotation {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.types[expr]
}

// SetType associates a type annotation with an expression node.
// This is called by the semantic analyzer during type inference.
//
// Not safe for concurrent writes. Should only be called during analysis.
func (si *SemanticInfo) SetType(expr Expression, typ *TypeAnnotation) {
	si.mu.Lock()
	defer si.mu.Unlock()
	delete(si.resolvedTypes, expr)
	si.types[expr] = typ
}

// HasType returns true if a type has been set for the given expression.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) HasType(expr Expression) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	_, ok := si.types[expr]
	return ok
}

// ClearType removes type information for an expression.
// Useful for error recovery or incremental re-analysis. Resolved annotations
// remain independent node bindings until Clear, since other expressions may
// share them.
//
// Not safe for concurrent writes.
func (si *SemanticInfo) ClearType(expr Expression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	delete(si.resolvedTypes, expr)
	delete(si.types, expr)
}

// ============================================================================
// Compile-Time Predicate Cache
// ============================================================================

// FoldedPredicate returns the compile-time value the semantic analyzer folded
// for a predicate call site, keyed by the callee identifier. The second result
// reports whether a value was recorded at all.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) FoldedPredicate(ident *Identifier) (bool, bool) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	value, ok := si.foldedPredicates[ident]
	return value, ok
}

// SetFoldedPredicate records the compile-time result of a predicate intrinsic
// such as Declared or ConditionalDefined. The value is keyed by the callee
// identifier, which is unique per call site.
//
// Not safe for concurrent writes. Should only be called during analysis.
func (si *SemanticInfo) SetFoldedPredicate(ident *Identifier, value bool) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.foldedPredicates[ident] = value
}

// HasFoldedPredicate returns true if a compile-time predicate value has been
// recorded for the given identifier.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) HasFoldedPredicate(ident *Identifier) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	_, ok := si.foldedPredicates[ident]
	return ok
}

// ============================================================================
// Statistics and Debugging
// ============================================================================

// TypeCount returns the number of expressions with type information.
// Useful for statistics and testing.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) TypeCount() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.types)
}

// FoldedPredicateCount returns the number of call sites with a folded
// compile-time predicate value. Useful for statistics and testing.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) FoldedPredicateCount() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.foldedPredicates)
}

// Clear removes all semantic information.
// Useful for resetting the metadata table or memory management.
//
// Not safe for concurrent access.
func (si *SemanticInfo) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.resolvedTypes = make(map[Node]types.Type)
	si.types = make(map[Expression]*TypeAnnotation)
	si.foldedPredicates = make(map[*Identifier]bool)
}

// GetResolvedType returns the analyzer's type object for an expression or type
// annotation. Consumers must treat it as read-only after analysis. Unlike
// GetType, this preserves bounds, signatures, and nominal identity.
func (si *SemanticInfo) GetResolvedType(node Node) types.Type {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.resolvedTypes[node]
}

// SetResolvedType records a resolved type without converting it to source text.
// If node has a compatibility annotation, that annotation shares the type so
// existing annotation consumers can migrate without reconstructing it.
func (si *SemanticInfo) SetResolvedType(node Node, typ types.Type) {
	if node == nil || typ == nil {
		return
	}
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.resolvedTypes == nil {
		si.resolvedTypes = make(map[Node]types.Type)
	}
	si.resolvedTypes[node] = typ
	if expr, ok := node.(Expression); ok {
		if annot := si.types[expr]; annot != nil {
			si.resolvedTypes[annot] = typ
		}
	}
}
