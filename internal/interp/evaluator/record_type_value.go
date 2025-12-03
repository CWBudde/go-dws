// Package evaluator provides the visitor-based evaluation engine for DWScript.
// This file contains RecordTypeValue for storing record type metadata.
package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// RecordTypeValue is an internal value type used to store record type metadata
// in the interpreter's environment.
type RecordTypeValue struct {
	RecordType *types.RecordType
	FieldDecls map[string]*ast.FieldDecl // Field declarations (for initializers) - Task 9.5
	Metadata   *runtime.RecordMetadata   // Runtime metadata (methods, constants, etc.) - Task 3.5.42

	// Deprecated: Use Metadata.Methods instead. Will be removed in Phase 3.5.44.
	// Kept temporarily for backward compatibility during migration.
	Methods              map[string]*ast.FunctionDecl   // Instance methods: Method name -> AST declaration
	StaticMethods        map[string]*ast.FunctionDecl   // Static methods: Method name -> AST declaration (class function/procedure)
	ClassMethods         map[string]*ast.FunctionDecl   // Alias for StaticMethods (for compatibility)
	MethodOverloads      map[string][]*ast.FunctionDecl // Instance method overloads
	ClassMethodOverloads map[string][]*ast.FunctionDecl // Static method overloads
	Constants            map[string]Value               // Record constants (evaluated at declaration) - Task 9.12.2
	ClassVars            map[string]Value               // Class variables (shared across all instances) - Task 9.12.2
}

// Type returns "RECORD_TYPE".
func (r *RecordTypeValue) Type() string {
	return "RECORD_TYPE"
}

// String returns the record type name.
func (r *RecordTypeValue) String() string {
	return r.RecordType.Name
}

// GetRecordType returns the underlying RecordType.
func (r *RecordTypeValue) GetRecordType() *types.RecordType {
	return r.RecordType
}

// GetMetadata returns the RecordMetadata for this record type.
func (r *RecordTypeValue) GetMetadata() any {
	return r.Metadata
}

// GetFieldDecls returns the FieldDecls map for this record type.
func (r *RecordTypeValue) GetFieldDecls() map[string]*ast.FieldDecl {
	return r.FieldDecls
}

// GetRecordTypeName returns the record type name (e.g., "TPoint").
func (r *RecordTypeValue) GetRecordTypeName() string {
	if r.RecordType != nil {
		return r.RecordType.Name
	}
	return ""
}

// HasStaticMethod checks if a static method (class function/procedure) with the given name exists.
// The lookup is case-insensitive.
func (r *RecordTypeValue) HasStaticMethod(name string) bool {
	methodNameLower := ident.Normalize(name)
	overloads, exists := r.ClassMethodOverloads[methodNameLower]
	return exists && len(overloads) > 0
}
