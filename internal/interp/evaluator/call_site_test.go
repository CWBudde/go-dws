package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func pos(line, column int) token.Position {
	return token.Position{Line: line, Column: column}
}

func identAt(name string, line, column int) *ast.Identifier {
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: token.Token{Type: token.IDENT, Literal: name, Pos: pos(line, column)},
			},
		},
		Value: name,
	}
}

// TestCallSitePos_NameToken pins the governing rule: a call site is reported at
// the callee's *name* token, never at the receiver and never at the end of the
// call expression.
func TestCallSitePos_NameToken(t *testing.T) {
	tests := []struct {
		name string
		node ast.Node
		want token.Position
	}{
		{
			name: "plain call reports the function name",
			node: &ast.CallExpression{Function: identAt("Foo", 1, 4)},
			want: pos(1, 4),
		},
		{
			name: "method call reports the method name, not the receiver",
			node: &ast.MethodCallExpression{
				Object: identAt("c", 51, 4),
				Method: identAt("Boom", 51, 6),
			},
			want: pos(51, 6),
		},
		{
			name: "member access reports the member name",
			node: &ast.MemberAccessExpression{
				Object: identAt("c", 51, 4),
				Member: identAt("Boom", 51, 6),
			},
			want: pos(51, 6),
		},
		{
			name: "call through a member access reports the member name",
			node: &ast.CallExpression{
				Function: &ast.MemberAccessExpression{
					Object: identAt("c", 51, 4),
					Member: identAt("Boom", 51, 6),
				},
			},
			want: pos(51, 6),
		},
		{
			name: "TClass.Create reports the constructor name",
			node: &ast.NewExpression{
				ClassName:      identAt("Exception", 3, 10),
				ConstructorPos: pos(3, 20),
			},
			want: pos(3, 20),
		},
		{
			name: "new TClass falls back to the class name",
			node: &ast.NewExpression{ClassName: identAt("Exception", 3, 14)},
			want: pos(3, 14),
		},
		{
			name: "non-call nodes fall back to Pos",
			node: identAt("x", 7, 2),
			want: pos(7, 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := callSitePos(tt.node)
			if got.Line != tt.want.Line || got.Column != tt.want.Column {
				t.Fatalf("callSitePos = [line: %d, column: %d], want [line: %d, column: %d]",
					got.Line, got.Column, tt.want.Line, tt.want.Column)
			}
		})
	}
}

// TestCallSitePosOf_Nil guards the frame-push helper against a missing current
// node.
func TestCallSitePosOf_Nil(t *testing.T) {
	if got := callSitePosOf(nil); got != nil {
		t.Fatalf("callSitePosOf(nil) = %v, want nil", got)
	}
}

func classNameIdent(name string) *ast.Identifier {
	return identAt(name, 1, 1)
}

// TestQualifiedRoutineName pins how a stack frame names a routine: methods are
// class-qualified whether the implementation is written out-of-line (ClassName)
// or inline in the class body (DeclaringClassName).
func TestQualifiedRoutineName(t *testing.T) {
	tests := []struct {
		name string
		fn   *ast.FunctionDecl
		want string
	}{
		{
			name: "free routine is unqualified",
			fn:   &ast.FunctionDecl{Name: classNameIdent("TestProc")},
			want: "TestProc",
		},
		{
			name: "out-of-line method uses ClassName",
			fn: &ast.FunctionDecl{
				Name:      classNameIdent("Boom"),
				ClassName: classNameIdent("TMyClass"),
			},
			want: "TMyClass.Boom",
		},
		{
			name: "inline method uses DeclaringClassName",
			fn: &ast.FunctionDecl{
				Name:               classNameIdent("TestMeth"),
				DeclaringClassName: "TTest",
			},
			want: "TTest.TestMeth",
		},
		{
			name: "ClassName wins over DeclaringClassName",
			fn: &ast.FunctionDecl{
				Name:               classNameIdent("Boom"),
				ClassName:          classNameIdent("TMyClass"),
				DeclaringClassName: "TOther",
			},
			want: "TMyClass.Boom",
		},
		{
			name: "nil routine has no name",
			fn:   nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualifiedRoutineName(tt.fn); got != tt.want {
				t.Fatalf("qualifiedRoutineName = %q, want %q", got, tt.want)
			}
		})
	}
}
