package evaluator

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

// callSitePos returns the source position DWScript reports for a call site: the
// position of the *callee's name token*.
//
// DWScript identifies a stack frame (and a raise site) by the name of the thing
// being called, not by the start of the whole expression and not by its end:
//
//	Foo(1)                -> `Foo`
//	obj.Bar(1)            -> `Bar`      (not `obj`)
//	(new TTest).Baz       -> `Baz`      (not `TTest`)
//	Exception.Create('x') -> `Create`   (not `Exception`, not the closing paren)
//
// AST `Pos()` deliberately keeps reporting the start of the expression, because
// diagnostics, hints and the `*Fail` fixtures are calibrated against it; this
// helper is the call-stack-specific view instead.
//
// Nodes that are not call-like fall back to their own Pos().
func callSitePos(node ast.Node) token.Position {
	switch n := node.(type) {
	case *ast.CallExpression:
		if n.Function != nil {
			return callSitePos(n.Function)
		}
	case *ast.MethodCallExpression:
		if n.Method != nil {
			return n.Method.Pos()
		}
	case *ast.MemberAccessExpression:
		if n.Member != nil {
			return n.Member.Pos()
		}
	case *ast.NewExpression:
		return n.ConstructorNamePos()
	case *ast.ExpressionStatement:
		if n.Expression != nil {
			return callSitePos(n.Expression)
		}
	}
	return node.Pos()
}

// callSitePosOf returns a pointer to the call-site position of node, or nil when
// node is nil.
func callSitePosOf(node ast.Node) *token.Position {
	if node == nil {
		return nil
	}
	pos := callSitePos(node)
	return &pos
}

// raiseSitePos returns the position DWScript reports for `raise <expr>`.
//
// For the usual `raise EFoo.Create(...)` / `raise obj.Make(...)` spellings the
// raise site is the callee's name token. For any other expression (a plain
// object reference, for instance) DWScript reports the parser position just past
// the expression, which End() models.
func raiseSitePos(expr ast.Expression) token.Position {
	switch expr.(type) {
	case *ast.NewExpression, *ast.CallExpression, *ast.MethodCallExpression:
		return callSitePos(expr)
	}
	return expr.End()
}

// qualifiedRoutineName renders the name DWScript puts on a stack frame: a method
// is qualified with its class ("TFoo.Bar"), a free routine is not.
//
// An out-of-line implementation (`procedure TFoo.Bar;`) carries the qualifier in
// ClassName. A method declared inline in the class body has no qualifier to
// parse, so the parser records the owning class in DeclaringClassName instead.
func qualifiedRoutineName(fn *ast.FunctionDecl) string {
	if fn == nil || fn.Name == nil {
		return ""
	}
	switch {
	case fn.ClassName != nil && fn.ClassName.Value != "":
		return fn.ClassName.Value + "." + fn.Name.Value
	case fn.DeclaringClassName != "":
		return fn.DeclaringClassName + "." + fn.Name.Value
	default:
		return fn.Name.Value
	}
}
