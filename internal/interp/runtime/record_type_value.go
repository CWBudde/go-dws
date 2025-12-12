package runtime

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// RecordTypeValue stores record type metadata in the environment.
// It is a runtime-level "type meta value" to support static member access like TRecord.Foo.
type RecordTypeValue struct {
	RecordType *types.RecordType
	FieldDecls map[string]*ast.FieldDecl
	Metadata   *RecordMetadata

	// Deprecated: Prefer Metadata where possible.
	Methods              map[string]*ast.FunctionDecl
	StaticMethods        map[string]*ast.FunctionDecl
	ClassMethods         map[string]*ast.FunctionDecl
	MethodOverloads      map[string][]*ast.FunctionDecl
	ClassMethodOverloads map[string][]*ast.FunctionDecl
	Constants            map[string]Value
	ClassVars            map[string]Value
}

func (r *RecordTypeValue) Type() string { return "RECORD_TYPE" }

func (r *RecordTypeValue) String() string {
	if r.RecordType == nil {
		return ""
	}
	return r.RecordType.Name
}

func (r *RecordTypeValue) GetRecordType() *types.RecordType { return r.RecordType }

func (r *RecordTypeValue) GetMetadata() any { return r.Metadata }

func (r *RecordTypeValue) GetFieldDecls() map[string]*ast.FieldDecl { return r.FieldDecls }

func (r *RecordTypeValue) GetRecordTypeName() string {
	if r.RecordType != nil {
		return r.RecordType.Name
	}
	return ""
}

// HasStaticMethod checks if a static method exists (case-insensitive).
func (r *RecordTypeValue) HasStaticMethod(name string) bool {
	methodNameLower := ident.Normalize(name)
	overloads, exists := r.ClassMethodOverloads[methodNameLower]
	return exists && len(overloads) > 0
}
