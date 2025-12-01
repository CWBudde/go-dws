// Package evaluator provides the visitor-based evaluation engine for DWScript.
// This file contains metadata builders for record types.
package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// buildRecordMetadata builds RecordMetadata from AST declarations.
// Task 3.5.10: Moved from Interpreter to Evaluator to eliminate adapter dependency.
// Task 3.5.42: Helper to create AST-free metadata for records.
func (e *Evaluator) buildRecordMetadata(
	recordName string,
	recordType *types.RecordType,
	methods map[string]*ast.FunctionDecl,
	staticMethods map[string]*ast.FunctionDecl,
	constants map[string]Value,
	classVars map[string]Value,
) *runtime.RecordMetadata {
	metadata := runtime.NewRecordMetadata(recordName, recordType)

	// Convert instance methods to MethodMetadata
	for methodName, methodDecl := range methods {
		methodMeta := e.buildMethodMetadata(methodDecl)
		metadata.Methods[methodName] = methodMeta
		metadata.MethodOverloads[methodName] = []*runtime.MethodMetadata{methodMeta}
	}

	// Convert static methods to MethodMetadata
	for methodName, methodDecl := range staticMethods {
		methodMeta := e.buildMethodMetadata(methodDecl)
		methodMeta.IsClassMethod = true
		metadata.StaticMethods[methodName] = methodMeta
		metadata.StaticMethodOverloads[methodName] = []*runtime.MethodMetadata{methodMeta}
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

// buildMethodMetadata converts an AST FunctionDecl to MethodMetadata.
// Task 3.5.10: Moved from Interpreter to Evaluator to eliminate adapter dependency.
// Task 3.5.42: Helper to extract metadata from AST method declarations.
func (e *Evaluator) buildMethodMetadata(decl *ast.FunctionDecl) *runtime.MethodMetadata {
	// Build parameter metadata
	params := make([]runtime.ParameterMetadata, len(decl.Parameters))
	for idx, param := range decl.Parameters {
		typeName := ""
		if param.Type != nil {
			typeName = param.Type.String()
		}
		params[idx] = runtime.ParameterMetadata{
			Name:         param.Name.Value,
			TypeName:     typeName,
			Type:         nil, // Will be resolved later if needed
			ByRef:        param.ByRef,
			DefaultValue: param.DefaultValue,
		}
	}

	// Determine return type
	returnTypeName := ""
	if decl.ReturnType != nil {
		returnTypeName = decl.ReturnType.String()
	}

	return &runtime.MethodMetadata{
		Name:           decl.Name.Value,
		Parameters:     params,
		ReturnTypeName: returnTypeName,
		ReturnType:     nil, // Will be resolved later if needed
		Body:           decl.Body,
		IsClassMethod:  decl.IsClassMethod,
		IsConstructor:  decl.IsConstructor,
		IsDestructor:   decl.IsDestructor,
	}
}
