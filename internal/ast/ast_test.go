package ast

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestProgram tests the Program node.
func TestProgram(t *testing.T) {
	// Empty program
	prog := &Program{Statements: []Statement{}}
	if prog.TokenLiteral() != "" {
		t.Errorf("empty program TokenLiteral() = %q, want empty string", prog.TokenLiteral())
	}
	if prog.String() != "" {
		t.Errorf("empty program String() = %q, want empty string", prog.String())
	}

	// Program with statements
	prog = &Program{
		Statements: []Statement{
			&ExpressionStatement{
				BaseNode:   NewTestBaseNode(lexer.INT, "42"),
				Expression: NewTestIntegerLiteral(42),
			},
		},
	}
	if prog.TokenLiteral() != "42" {
		t.Errorf("program TokenLiteral() = %q, want %q", prog.TokenLiteral(), "42")
	}
	if prog.String() != "42" {
		t.Errorf("program String() = %q, want %q", prog.String(), "42")
	}
}

// TestIdentifier tests the Identifier node.
func TestIdentifier(t *testing.T) {
	ident := NewTestIdentifier("myVar")

	if ident.TokenLiteral() != "myVar" {
		t.Errorf("TokenLiteral() = %q, want %q", ident.TokenLiteral(), "myVar")
	}
	if ident.String() != "myVar" {
		t.Errorf("String() = %q, want %q", ident.String(), "myVar")
	}
}

// TestIntegerLiteral tests the IntegerLiteral node.
func TestIntegerLiteral(t *testing.T) {
	tests := []struct {
		literal string
		value   int64
	}{
		{"0", 0},
		{"42", 42},
		{"-5", -5},
		{"1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.literal, func(t *testing.T) {
			node := NewTestIntegerLiteral(tt.value)

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.literal {
				t.Errorf("String() = %q, want %q", node.String(), tt.literal)
			}
		})
	}
}

// TestFloatLiteral tests the FloatLiteral node.
func TestFloatLiteral(t *testing.T) {
	tests := []struct {
		literal string
		value   float64
	}{
		{"0.0", 0.0},
		{"3.14", 3.14},
		{"2.5", 2.5},
		{"1.0e10", 1.0e10},
	}

	for _, tt := range tests {
		t.Run(tt.literal, func(t *testing.T) {
			node := NewTestFloatLiteral(tt.value, tt.literal)

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.literal {
				t.Errorf("String() = %q, want %q", node.String(), tt.literal)
			}
		})
	}
}

// TestStringLiteral tests the StringLiteral node.
func TestStringLiteral(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		value   string
		want    string
	}{
		{"simple", "'hello'", "hello", "\"hello\""},
		{"empty", "''", "", "\"\""},
		{"with spaces", "'hello world'", "hello world", "\"hello world\""},
		{"escaped", "'it''s'", "it's", "\"it's\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestStringLiteral(tt.value, tt.literal)

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestBooleanLiteral tests the BooleanLiteral node.
func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		literal string
		value   bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.literal, func(t *testing.T) {
			node := NewTestBooleanLiteral(tt.value)

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.literal {
				t.Errorf("String() = %q, want %q", node.String(), tt.literal)
			}
		})
	}
}

// TestCharLiteral tests the CharLiteral node.
func TestCharLiteral(t *testing.T) {
	tests := []struct {
		name    string
		literal string
		value   rune
	}{
		{"decimal", "#65", 'A'},
		{"hex", "#$41", 'A'},
		{"carriage return", "#13", '\r'},
		{"line feed", "#10", '\n'},
		{"lowercase hex", "#$61", 'a'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &CharLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: NewTestBaseNode(lexer.CHAR, tt.literal)}, Value: tt.value}

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.literal {
				t.Errorf("String() = %q, want %q", node.String(), tt.literal)
			}
			if node.Value != tt.value {
				t.Errorf("Value = %q, want %q", node.Value, tt.value)
			}

			// Test position tracking
			expectedPos := lexer.Position{Line: 0, Column: 0}
			if node.Pos() != expectedPos {
				t.Errorf("Pos() = %v, want %v", node.Pos(), expectedPos)
			}

			// Test type annotation
			if node.GetType() != nil {
				t.Errorf("GetType() = %v, want nil", node.GetType())
			}
			typ := NewTestTypeAnnotation("String")
			node.SetType(typ)
			if node.GetType() != typ {
				t.Errorf("GetType() after SetType = %v, want %v", node.GetType(), typ)
			}
		})
	}
}

// TestNilLiteral tests the NilLiteral node.
func TestNilLiteral(t *testing.T) {
	node := &NilLiteral{TypedExpressionBase: TypedExpressionBase{BaseNode: NewTestBaseNode(lexer.NIL, "nil")}}

	if node.TokenLiteral() != "nil" {
		t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), "nil")
	}
	if node.String() != "nil" {
		t.Errorf("String() = %q, want %q", node.String(), "nil")
	}
}

// TestBinaryExpression tests the BinaryExpression node.
func TestBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		left     Expression
		operator string
		right    Expression
		want     string
	}{
		{
			name:     "simple addition",
			left:     NewTestIntegerLiteral(1),
			operator: "+",
			right:    NewTestIntegerLiteral(2),
			want:     "(1 + 2)",
		},
		{
			name:     "multiplication",
			left:     NewTestIntegerLiteral(3),
			operator: "*",
			right:    NewTestIntegerLiteral(4),
			want:     "(3 * 4)",
		},
		{
			name:     "comparison",
			left:     NewTestIdentifier("x"),
			operator: "<",
			right:    NewTestIntegerLiteral(10),
			want:     "(x < 10)",
		},
		{
			name: "nested expression",
			left: NewTestBinaryExpression(
				NewTestIntegerLiteral(1),
				"+",
				NewTestIntegerLiteral(2),
			),
			operator: "*",
			right:    NewTestIntegerLiteral(3),
			want:     "((1 + 2) * 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestBinaryExpression(tt.left, tt.operator, tt.right)

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestUnaryExpression tests the UnaryExpression node.
func TestUnaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		right    Expression
		want     string
	}{
		{
			name:     "negation",
			operator: "-",
			right:    NewTestIntegerLiteral(5),
			want:     "(-5)",
		},
		{
			name:     "not",
			operator: "not",
			right:    NewTestBooleanLiteral(true),
			want:     "(not true)",
		},
		{
			name:     "unary plus",
			operator: "+",
			right:    NewTestIntegerLiteral(3),
			want:     "(+3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestUnaryExpression(tt.operator, tt.right)

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestGroupedExpression tests the GroupedExpression node.
func TestGroupedExpression(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
		want string
	}{
		{
			name: "simple",
			expr: NewTestIntegerLiteral(42),
			want: "(42)",
		},
		{
			name: "binary expression",
			expr: NewTestBinaryExpression(
				NewTestIntegerLiteral(3),
				"+",
				NewTestIntegerLiteral(5),
			),
			want: "((3 + 5))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestGroupedExpression(tt.expr)

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestExpressionStatement tests the ExpressionStatement node.
func TestExpressionStatement(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
		want string
	}{
		{
			name: "integer literal",
			expr: NewTestIntegerLiteral(42),
			want: "42",
		},
		{
			name: "identifier",
			expr: NewTestIdentifier("x"),
			want: "x",
		},
		{
			name: "nil expression",
			expr: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ExpressionStatement{
				BaseNode:   NewTestBaseNode(lexer.INT, "42"),
				Expression: tt.expr,
			}

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestBlockStatement tests the BlockStatement node.
func TestBlockStatement(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		stmts []Statement
	}{
		{
			name:  "empty block",
			stmts: []Statement{},
			want:  "begin\nend",
		},
		{
			name: "single statement",
			stmts: []Statement{
				&ExpressionStatement{
					BaseNode:   NewTestBaseNode(lexer.INT, "42"),
					Expression: NewTestIntegerLiteral(42),
				},
			},
			want: "begin\n  42\nend",
		},
		{
			name: "multiple statements",
			stmts: []Statement{
				&ExpressionStatement{
					BaseNode:   NewTestBaseNode(lexer.INT, "1"),
					Expression: NewTestIntegerLiteral(1),
				},
				&ExpressionStatement{
					BaseNode:   NewTestBaseNode(lexer.INT, "2"),
					Expression: NewTestIntegerLiteral(2),
				},
			},
			want: "begin\n  1\n  2\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestBlockStatement(tt.stmts)

			if node.String() != tt.want {
				t.Errorf("String() =\n%q\nwant:\n%q", node.String(), tt.want)
			}
		})
	}
}

// TestVarDeclStatement tests the VarDeclStatement node.
func TestVarDeclStatement(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		varType *TypeAnnotation
		value   Expression
		want    string
	}{
		{
			name:    "declaration without initialization",
			varName: "x",
			varType: &TypeAnnotation{Name: "Integer"},
			value:   nil,
			want:    "var x: Integer",
		},
		{
			name:    "declaration with integer initialization",
			varName: "x",
			varType: NewTestTypeAnnotation("Integer"),
			value:   NewTestIntegerLiteral(42),
			want:    "var x: Integer := 42",
		},
		{
			name:    "declaration with string initialization",
			varName: "s",
			varType: NewTestTypeAnnotation("String"),
			value:   NewTestStringLiteral("hello", "'hello'"),
			want:    "var s: String := \"hello\"",
		},
		{
			name:    "declaration without type",
			varName: "x",
			varType: nil,
			value:   NewTestIntegerLiteral(5),
			want:    "var x = 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &VarDeclStatement{
				BaseNode: NewTestBaseNode(lexer.VAR, "var"),
				Names: []*Identifier{
					NewTestIdentifier(tt.varName),
				},
				Type:     tt.varType,
				Value:    tt.value,
				Inferred: tt.varType == nil && tt.value != nil,
			}

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
			if node.TokenLiteral() != "var" {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), "var")
			}
		})
	}
}

// TestAssignmentStatement tests the AssignmentStatement node.
func TestAssignmentStatement(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		value   Expression
		want    string
	}{
		{
			name:    "simple integer assignment",
			varName: "x",
			value:   NewTestIntegerLiteral(10),
			want:    "x := 10",
		},
		{
			name:    "expression assignment",
			varName: "y",
			value: NewTestBinaryExpression(
				NewTestIdentifier("x"),
				"+",
				NewTestIntegerLiteral(1),
			),
			want: "y := (x + 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &AssignmentStatement{
				BaseNode: NewTestBaseNode(lexer.ASSIGN, ":="),
				Target:   NewTestIdentifier(tt.varName),
				Value:    tt.value,
			}

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
			if node.TokenLiteral() != ":=" {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), ":=")
			}
		})
	}
}

// TestCallExpression tests the CallExpression node.
func TestCallExpression(t *testing.T) {
	tests := []struct {
		function  Expression
		name      string
		want      string
		arguments []Expression
	}{
		{
			name:      "no arguments",
			function:  NewTestIdentifier("PrintLn"),
			arguments: []Expression{},
			want:      "PrintLn()",
		},
		{
			name:     "single string argument",
			function: NewTestIdentifier("PrintLn"),
			arguments: []Expression{
				NewTestStringLiteral("hello", "'hello'"),
			},
			want: "PrintLn(\"hello\")",
		},
		{
			name:     "multiple arguments",
			function: NewTestIdentifier("Add"),
			arguments: []Expression{
				NewTestIntegerLiteral(3),
				NewTestIntegerLiteral(5),
			},
			want: "Add(3, 5)",
		},
		{
			name:     "expression arguments",
			function: NewTestIdentifier("PrintLn"),
			arguments: []Expression{
				NewTestBinaryExpression(
					NewTestIntegerLiteral(2),
					"+",
					NewTestIntegerLiteral(3),
				),
			},
			want: "PrintLn((2 + 3))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewTestCallExpression(tt.function, tt.arguments)

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestForInStatement tests the ForInStatement node.
func TestForInStatement(t *testing.T) {
	tests := []struct {
		collection Expression
		body       Statement
		name       string
		variable   string
		want       string
		inlineVar  bool
	}{
		{
			name:       "basic for-in",
			variable:   "e",
			collection: NewTestIdentifier("mySet"),
			body: &ExpressionStatement{
				BaseNode: NewTestBaseNode(lexer.IDENT, "PrintLn"),
				Expression: NewTestCallExpression(
					NewTestIdentifier("PrintLn"),
					[]Expression{
						NewTestIdentifier("e"),
					},
				),
			},
			inlineVar: false,
			want:      "for e in mySet do PrintLn(e)",
		},
		{
			name:       "for-in with inline var",
			variable:   "item",
			collection: NewTestIdentifier("myArray"),
			body: &BlockStatement{
				BaseNode: NewTestBaseNode(lexer.BEGIN, "begin"),
				Statements: []Statement{
					&ExpressionStatement{
						BaseNode: NewTestBaseNode(lexer.IDENT, "Process"),
						Expression: NewTestCallExpression(
							NewTestIdentifier("Process"),
							[]Expression{
								NewTestIdentifier("item"),
							},
						),
					},
				},
			},
			inlineVar: true,
			want:      "for var item in myArray do begin\n  Process(item)\nend",
		},
		{
			name:       "for-in with string literal",
			variable:   "ch",
			collection: NewTestStringLiteral("hello", "'hello'"),
			body: &ExpressionStatement{
				BaseNode: NewTestBaseNode(lexer.IDENT, "Print"),
				Expression: NewTestCallExpression(
					NewTestIdentifier("Print"),
					[]Expression{
						NewTestIdentifier("ch"),
					},
				),
			},
			inlineVar: false,
			want:      "for ch in \"hello\" do Print(ch)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ForInStatement{
				BaseNode:   NewTestBaseNode(lexer.FOR, "for"),
				Variable:   NewTestIdentifier(tt.variable),
				Collection: tt.collection,
				Body:       tt.body,
				InlineVar:  tt.inlineVar,
			}

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
			if node.TokenLiteral() != "for" {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), "for")
			}

			// Test position tracking
			if node.Pos().Line != 0 {
				t.Errorf("Pos().Line = %d, want 0", node.Pos().Line)
			}
		})
	}
}

// TestInterfaceImplementation verifies that all node types implement their respective interfaces.
func TestInterfaceImplementation(_ *testing.T) {
	// Test that expression nodes implement Expression interface
	var _ Expression = &Identifier{}
	var _ Expression = &IntegerLiteral{}
	var _ Expression = &FloatLiteral{}
	var _ Expression = &StringLiteral{}
	var _ Expression = &BooleanLiteral{}
	var _ Expression = &CharLiteral{}
	var _ Expression = &NilLiteral{}
	var _ Expression = &BinaryExpression{}
	var _ Expression = &UnaryExpression{}
	var _ Expression = &GroupedExpression{}
	var _ Expression = &CallExpression{}

	// Test that statement nodes implement Statement interface
	var _ Statement = &ExpressionStatement{}
	var _ Statement = &BlockStatement{}
	var _ Statement = &VarDeclStatement{}
	var _ Statement = &AssignmentStatement{}
	var _ Statement = &ForInStatement{}

	// Test that all nodes implement Node interface
	var _ Node = &Program{}
	var _ Node = &Identifier{}
	var _ Node = &IntegerLiteral{}
	var _ Node = &FloatLiteral{}
	var _ Node = &StringLiteral{}
	var _ Node = &BooleanLiteral{}
	var _ Node = &CharLiteral{}
	var _ Node = &NilLiteral{}
	var _ Node = &BinaryExpression{}
	var _ Node = &UnaryExpression{}
	var _ Node = &GroupedExpression{}
	var _ Node = &ExpressionStatement{}
	var _ Node = &BlockStatement{}
	var _ Node = &VarDeclStatement{}
	var _ Node = &AssignmentStatement{}
	var _ Node = &CallExpression{}
}
