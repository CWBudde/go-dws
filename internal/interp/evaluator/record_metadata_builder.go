// Package evaluator provides the visitor-based evaluation engine for DWScript.
// This file contains metadata builders for record types.
package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// buildRecordMetadata builds RecordMetadata from AST declarations.
func (e *Evaluator) buildRecordMetadata(
	recordName string,
	recordType *types.RecordType,
	methods map[string]*ast.FunctionDecl,
	staticMethods map[string]*ast.FunctionDecl,
	methodOverloads map[string][]*ast.FunctionDecl,
	staticMethodOverloads map[string][]*ast.FunctionDecl,
	constants map[string]Value,
	classVars map[string]Value,
	ctx *ExecutionContext,
) *runtime.RecordMetadata {
	metadata := runtime.NewRecordMetadata(recordName, recordType)
	callables := make(map[*ast.FunctionDecl]*runtime.MethodMetadata)
	methodMetadata := func(decl *ast.FunctionDecl) *runtime.MethodMetadata {
		if cached := callables[decl]; cached != nil {
			return cached
		}
		result := runtime.MethodMetadataFromAST(decl, e.metadataTypeResolver(ctx))
		callables[decl] = result
		return result
	}

	// Convert instance methods to MethodMetadata (keeping all overloads)
	for methodName, methodDecl := range methods {
		metadata.Methods[methodName] = methodMetadata(methodDecl)
	}
	for methodName, decls := range methodOverloads {
		metas := make([]*runtime.MethodMetadata, 0, len(decls))
		for _, decl := range decls {
			metas = append(metas, methodMetadata(decl))
		}
		metadata.MethodOverloads[methodName] = metas
	}

	// Convert static methods to MethodMetadata (keeping all overloads)
	for methodName, methodDecl := range staticMethods {
		methodMeta := methodMetadata(methodDecl)
		methodMeta.IsClassMethod = true
		metadata.StaticMethods[methodName] = methodMeta
	}
	for methodName, decls := range staticMethodOverloads {
		metas := make([]*runtime.MethodMetadata, 0, len(decls))
		for _, decl := range decls {
			methodMeta := methodMetadata(decl)
			methodMeta.IsClassMethod = true
			metas = append(metas, methodMeta)
		}
		metadata.StaticMethodOverloads[methodName] = metas
	}

	// Copy constants and class vars
	for k, v := range constants {
		metadata.Constants[k] = v
	}
	for k, v := range classVars {
		metadata.ClassVars[k] = v
	}

	return metadata
}

// metadataTypeResolver fills runtime callable types from semantic identities or
// structured unchecked annotations in the declaration's environment.
func (e *Evaluator) metadataTypeResolver(ctx *ExecutionContext) runtime.TypeResolver {
	return func(annotation ast.TypeExpression) types.Type {
		resolved, err := e.ResolveTypeFromAnnotation(annotation, ctx)
		if err != nil {
			return nil
		}
		return resolved
	}
}

func (e *Evaluator) resolveClassCallableTypes(metadata *runtime.ClassMetadata, declaration *ast.FunctionDecl, ctx *ExecutionContext) {
	if metadata == nil || declaration == nil || declaration.Name == nil {
		return
	}
	name := ident.Normalize(declaration.Name.Value)
	candidates := []*runtime.MethodMetadata{metadata.Methods[name], metadata.ClassMethods[name], metadata.Constructors[name], metadata.Destructor}
	candidates = append(candidates, metadata.MethodOverloads[name]...)
	candidates = append(candidates, metadata.ClassMethodOverloads[name]...)
	candidates = append(candidates, metadata.ConstructorOverloads[name]...)
	for _, candidate := range candidates {
		if candidate != nil && (candidate.Declaration == declaration || candidate.SourceDeclaration == declaration) {
			candidate.ResolveTypes(e.metadataTypeResolver(ctx))
			return
		}
	}
}
