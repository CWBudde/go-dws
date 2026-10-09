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

// EnumElementBinding is an immutable compile-time enum constant identity.
// EnumType is canonical and read-only after analysis; no runtime value is shared.
type EnumElementBinding struct {
	EnumType           *types.EnumType
	Name               string
	DeprecationMessage string
	Ordinal            int
	IsDeprecated       bool
}

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
	enumElements              map[Expression]EnumElementBinding
	resolvedTypes             map[Node]types.Type
	types                     map[Expression]*TypeAnnotation
	foldedPredicates          map[*Identifier]bool
	propertyReads             map[*MethodCallExpression]*MemberAccessExpression
	implicitPropertyReads     map[*CallExpression]*ImplicitPropertyReadBinding
	inheritedPropertyReads    map[*InheritedExpression]*InheritedPropertyReadBinding
	indexedPropertyReads      map[*IndexExpression]*IndexedPropertyReadBinding
	resolvedIndexedProperties map[*IndexExpression]bool
	implicitCalls             map[Expression]bool
	defaultNamespace          map[Expression]bool
	mu                        sync.RWMutex
}

// NewSemanticInfo creates a new empty semantic metadata table.
// Each semantic analysis should create its own SemanticInfo instance.
func NewSemanticInfo() *SemanticInfo {
	return &SemanticInfo{
		resolvedTypes:             make(map[Node]types.Type),
		enumElements:              make(map[Expression]EnumElementBinding),
		types:                     make(map[Expression]*TypeAnnotation),
		foldedPredicates:          make(map[*Identifier]bool),
		propertyReads:             make(map[*MethodCallExpression]*MemberAccessExpression),
		implicitPropertyReads:     make(map[*CallExpression]*ImplicitPropertyReadBinding),
		inheritedPropertyReads:    make(map[*InheritedExpression]*InheritedPropertyReadBinding),
		indexedPropertyReads:      make(map[*IndexExpression]*IndexedPropertyReadBinding),
		resolvedIndexedProperties: make(map[*IndexExpression]bool),
		implicitCalls:             make(map[Expression]bool),
		defaultNamespace:          make(map[Expression]bool),
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
	si.enumElements = make(map[Expression]EnumElementBinding)
	si.types = make(map[Expression]*TypeAnnotation)
	si.foldedPredicates = make(map[*Identifier]bool)
	si.propertyReads = make(map[*MethodCallExpression]*MemberAccessExpression)
	si.implicitPropertyReads = make(map[*CallExpression]*ImplicitPropertyReadBinding)
	si.inheritedPropertyReads = make(map[*InheritedExpression]*InheritedPropertyReadBinding)
	si.indexedPropertyReads = make(map[*IndexExpression]*IndexedPropertyReadBinding)
	si.resolvedIndexedProperties = make(map[*IndexExpression]bool)
	si.implicitCalls = make(map[Expression]bool)
	si.defaultNamespace = make(map[Expression]bool)
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

// SetImplicitCall records a deliberate zero-argument call reading for a bare
// expression. The evaluator must invoke the original callable once and preserve
// its result, even when the result is itself callable. Set only during analysis.
func (si *SemanticInfo) SetImplicitCall(expr Expression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.implicitCalls[expr] = true
}

// IsImplicitCall reports an analyzed call reading, independently of result type.
// It is safe for concurrent reads after analysis.
func (si *SemanticInfo) IsImplicitCall(expr Expression) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.implicitCalls[expr]
}

// SetDefaultNamespace records a receiver bound to the standard result unit.
// Its lexical identity must survive caller environments during execution.
func (si *SemanticInfo) SetDefaultNamespace(expr Expression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.defaultNamespace == nil {
		si.defaultNamespace = make(map[Expression]bool)
	}
	si.defaultNamespace[expr] = true
}

// IsDefaultNamespace reports a compile-time result-unit receiver binding.
// It is safe for concurrent reads after analysis.
func (si *SemanticInfo) IsDefaultNamespace(expr Expression) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.defaultNamespace[expr]
}

// SetEnumElementBinding records a constant selected during source lookup.
func (si *SemanticInfo) SetEnumElementBinding(expr Expression, binding EnumElementBinding) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.enumElements == nil {
		si.enumElements = make(map[Expression]EnumElementBinding)
	}
	si.enumElements[expr] = binding
}

// EnumElementBinding returns the selected constant, independently of value type.
func (si *SemanticInfo) EnumElementBinding(expr Expression) (EnumElementBinding, bool) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	binding, ok := si.enumElements[expr]
	return binding, ok
}

// SetPropertyRead binds an empty legacy property call to its resolved read.
// The read and its receiver are immutable after semantic analysis.
func (si *SemanticInfo) SetPropertyRead(call *MethodCallExpression, read *MemberAccessExpression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.propertyReads == nil {
		si.propertyReads = make(map[*MethodCallExpression]*MemberAccessExpression)
	}
	si.propertyReads[call] = read
}

// PropertyRead returns the property read bound to a legacy empty call, if any.
// It is safe for concurrent reads.
func (si *SemanticInfo) PropertyRead(call *MethodCallExpression) *MemberAccessExpression {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.propertyReads[call]
}

// ImplicitPropertyReadBinding preserves the declaring property owner and the
// receiver for a checked unqualified compatibility call. Owner selects the
// descriptor; the receiver retains virtual getter dispatch.
type ImplicitPropertyReadBinding struct {
	Read  *MemberAccessExpression
	Owner string
}

// SetImplicitPropertyRead binds a direct empty property call to its implicit-Self read.
// The read is immutable after semantic analysis.
func (si *SemanticInfo) SetImplicitPropertyRead(call *CallExpression, read *ImplicitPropertyReadBinding) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.implicitPropertyReads == nil {
		si.implicitPropertyReads = make(map[*CallExpression]*ImplicitPropertyReadBinding)
	}
	si.implicitPropertyReads[call] = read
}

// ImplicitPropertyRead returns the bound implicit-Self property read, if any.
// It is safe for concurrent reads.
func (si *SemanticInfo) ImplicitPropertyRead(call *CallExpression) *ImplicitPropertyReadBinding {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.implicitPropertyReads[call]
}

// InheritedPropertyReadBinding retains the selected property and its resolved
// accessor identity, independently from dynamic Self used for virtual dispatch.
type InheritedPropertyReadBinding struct {
	Read     *MemberAccessExpression
	Property *types.PropertyInfo
	Owner    string
}

// SetInheritedPropertyRead binds a named inherited scalar read to its parent descriptor.
// The binding retains implicit Self so a virtual accessor keeps dynamic dispatch.
func (si *SemanticInfo) SetInheritedPropertyRead(expr *InheritedExpression, read *InheritedPropertyReadBinding) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.inheritedPropertyReads == nil {
		si.inheritedPropertyReads = make(map[*InheritedExpression]*InheritedPropertyReadBinding)
	}
	si.inheritedPropertyReads[expr] = read
}

// InheritedPropertyRead returns the selected parent property read, if any.
func (si *SemanticInfo) InheritedPropertyRead(expr *InheritedExpression) *InheritedPropertyReadBinding {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.inheritedPropertyReads[expr]
}

// IndexedPropertyReadBinding retains the selected descriptor and one declared
// bracket group's arguments for a checked compatibility read.
type IndexedPropertyReadBinding struct {
	Read    *InheritedPropertyReadBinding
	Indices []Expression
}

// SetIndexedPropertyRead binds an explicitly named indexed compatibility read.
// The binding and its expressions are immutable after analysis.
func (si *SemanticInfo) SetIndexedPropertyRead(expr *IndexExpression, read *IndexedPropertyReadBinding) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.indexedPropertyReads == nil {
		si.indexedPropertyReads = make(map[*IndexExpression]*IndexedPropertyReadBinding)
	}
	si.indexedPropertyReads[expr] = read
}

// IndexedPropertyRead returns a bound indexed compatibility read, if any.
// It is safe for concurrent reads.
func (si *SemanticInfo) IndexedPropertyRead(expr *IndexExpression) *IndexedPropertyReadBinding {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.indexedPropertyReads[expr]
}

// MarkResolvedIndexedProperty records that analysis actually selected an indexed
// property for this exact bracket group. The marker holds no execution state and
// is immutable after analysis; it settles provisional parser index diagnostics.
func (si *SemanticInfo) MarkResolvedIndexedProperty(expr *IndexExpression) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if si.resolvedIndexedProperties == nil {
		si.resolvedIndexedProperties = make(map[*IndexExpression]bool)
	}
	si.resolvedIndexedProperties[expr] = true
}

// IsResolvedIndexedProperty reports a group marked by actual property selection.
// Concurrent reads are safe.
func (si *SemanticInfo) IsResolvedIndexedProperty(expr *IndexExpression) bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.resolvedIndexedProperties[expr]
}
