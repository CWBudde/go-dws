package evaluator

import (
	"github.com/cwbudde/go-dws/internal/ast"
)

// This file contains visitor methods for declaration AST nodes.
// Phase 3.5.2: Visitor pattern implementation for declarations.
//
// Declarations define types, functions, classes, etc. and register them
// in the appropriate registries.

// VisitFunctionDecl evaluates a function declaration.
func (e *Evaluator) VisitFunctionDecl(node *ast.FunctionDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move function registration logic here (use typeSystem.FunctionRegistry)
	return e.adapter.EvalNode(node)
}

// VisitClassDecl evaluates a class declaration.
func (e *Evaluator) VisitClassDecl(node *ast.ClassDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move class registration logic here (use typeSystem)
	return e.adapter.EvalNode(node)
}

// VisitInterfaceDecl evaluates an interface declaration.
func (e *Evaluator) VisitInterfaceDecl(node *ast.InterfaceDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move interface registration logic here (use typeSystem)
	return e.adapter.EvalNode(node)
}

// VisitOperatorDecl evaluates an operator declaration (operator overloading).
func (e *Evaluator) VisitOperatorDecl(node *ast.OperatorDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move operator registration logic here (use typeSystem.OperatorRegistry)
	return e.adapter.EvalNode(node)
}

// VisitEnumDecl evaluates an enum declaration.
func (e *Evaluator) VisitEnumDecl(node *ast.EnumDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move enum registration logic here (use typeSystem)
	return e.adapter.EvalNode(node)
}

// VisitRecordDecl evaluates a record declaration.
func (e *Evaluator) VisitRecordDecl(node *ast.RecordDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move record registration logic here (use typeSystem)
	return e.adapter.EvalNode(node)
}

// VisitHelperDecl evaluates a helper declaration (type extension).
func (e *Evaluator) VisitHelperDecl(node *ast.HelperDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	// Future: Move helper registration logic here (use typeSystem)
	return e.adapter.EvalNode(node)
}

// VisitArrayDecl evaluates an array type declaration.
func (e *Evaluator) VisitArrayDecl(node *ast.ArrayDecl, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}

// VisitTypeDeclaration evaluates a type alias declaration.
func (e *Evaluator) VisitTypeDeclaration(node *ast.TypeDeclaration, ctx *ExecutionContext) Value {
	// Phase 3.5.2: Delegate to interpreter for now
	return e.adapter.EvalNode(node)
}
