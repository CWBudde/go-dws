package ast

import (
	"sync"
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
// semantic information (types, symbols, etc.). This provides:
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
// - SemanticInfo: Main metadata table, thread-safe, maps nodes to types/symbols
// - Expression → *TypeAnnotation: Maps expression nodes to their inferred types
// - *Identifier → Symbol: Maps identifiers to their symbol table entries (future)
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
// It maps AST nodes to their inferred types, resolved symbols, and other
// semantic information. This separation allows the AST to remain immutable
// after parsing while still supporting multiple semantic analyses.
//
// Thread Safety: SemanticInfo is safe for concurrent reads but not concurrent
// writes. Typical usage is single-threaded analysis (writes) followed by
// concurrent interpretation/compilation (reads).
type SemanticInfo struct {
	types   map[Expression]*TypeAnnotation
	symbols map[*Identifier]interface{}
	mu      sync.RWMutex
}

// NewSemanticInfo creates a new empty semantic metadata table.
// Each semantic analysis should create its own SemanticInfo instance.
func NewSemanticInfo() *SemanticInfo {
	return &SemanticInfo{
		types:   make(map[Expression]*TypeAnnotation),
		symbols: make(map[*Identifier]interface{}),
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
// Useful for error recovery or incremental re-analysis.
//
// Not safe for concurrent writes.
func (si *SemanticInfo) ClearType(expr Expression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	delete(si.types, expr)
}

// ============================================================================
// Symbol Information API
// ============================================================================

// GetSymbol returns the resolved symbol for an identifier node.
// Returns nil if no symbol has been set for this identifier.
//
// Thread-safe for concurrent reads.
//
// NOTE: Currently returns any to avoid circular dependency.
// TODO: Refine to return proper Symbol type
func (si *SemanticInfo) GetSymbol(ident *Identifier) any {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.symbols[ident]
}

// SetSymbol associates a symbol with an identifier node.
// This is called by the semantic analyzer during name resolution.
//
// Not safe for concurrent writes. Should only be called during analysis.
//
// NOTE: Currently accepts any to avoid circular dependency.
// TODO: Refine to return proper Symbol type
func (si *SemanticInfo) SetSymbol(ident *Identifier, symbol any) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.symbols[ident] = symbol
}

// HasSymbol returns true if a symbol has been set for the given identifier.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) HasSymbol(ident *Identifier) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	_, ok := si.symbols[ident]
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

// SymbolCount returns the number of identifiers with symbol information.
// Useful for statistics and testing.
//
// Thread-safe for concurrent reads.
func (si *SemanticInfo) SymbolCount() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.symbols)
}

// Clear removes all semantic information.
// Useful for resetting the metadata table or memory management.
//
// Not safe for concurrent access.
func (si *SemanticInfo) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.types = make(map[Expression]*TypeAnnotation)
	si.symbols = make(map[*Identifier]interface{})
}
