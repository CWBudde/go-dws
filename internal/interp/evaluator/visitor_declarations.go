package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// VisitFunctionDecl evaluates a function declaration.
// Phase 3.5.44: Delegate to adapter for function registration.
// The Interpreter registers to both legacy i.functions map and TypeSystem.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitClassDecl evaluates a class declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for complex class registration logic.
// This includes building ClassInfo/ClassMetadata, handling inheritance, methods, properties, etc.
// This complex logic remains in Interpreter for now.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitInterfaceDecl evaluates an interface declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for complex interface registration logic.
// This includes building InterfaceInfo, handling inheritance, methods, etc.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
// Phase 3.5.44: Delegate to adapter.EvalNode() for complex operator registration logic.
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitEnumDecl evaluates an enum declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for enum registration logic.
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitRecordDecl evaluates a record declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for complex record registration logic.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitHelperDecl evaluates a helper declaration (type extension).
// Phase 3.5.44: Delegate to adapter.EvalNode() for complex helper registration logic.
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitArrayDecl evaluates an array type declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for array type registration logic.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitTypeDeclaration evaluates a type alias declaration.
// Phase 3.5.44: Delegate to adapter.EvalNode() for type alias registration logic.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	return e.adapter.EvalNode(node)
}

// VisitSetDecl evaluates a set declaration.
func (e *Evaluator) VisitSetDecl(node *ast.SetDecl, ctx *ExecutionContext) Value {
	// Set type already registered by semantic analyzer
	// Delegate to adapter for now (Phase 3 migration)
	return e.adapter.EvalNode(node)
}
