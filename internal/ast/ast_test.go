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
				Token: lexer.Token{Type: lexer.INT, Literal: "42"},
				Expression: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "42"},
					Value: 42,
				},
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
	ident := &Identifier{
		Token: lexer.Token{Type: lexer.IDENT, Literal: "myVar"},
		Value: "myVar",
	}

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
			node := &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: tt.literal},
				Value: tt.value,
			}

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
			node := &FloatLiteral{
				Token: lexer.Token{Type: lexer.FLOAT, Literal: tt.literal},
				Value: tt.value,
			}

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
			node := &StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: tt.literal},
				Value: tt.value,
			}

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
			node := &BooleanLiteral{
				Token: lexer.Token{Type: lexer.TRUE, Literal: tt.literal},
				Value: tt.value,
			}

			if node.TokenLiteral() != tt.literal {
				t.Errorf("TokenLiteral() = %q, want %q", node.TokenLiteral(), tt.literal)
			}
			if node.String() != tt.literal {
				t.Errorf("String() = %q, want %q", node.String(), tt.literal)
			}
		})
	}
}

// TestNilLiteral tests the NilLiteral node.
func TestNilLiteral(t *testing.T) {
	node := &NilLiteral{
		Token: lexer.Token{Type: lexer.NIL, Literal: "nil"},
	}

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
			name: "simple addition",
			left: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "1"},
				Value: 1,
			},
			operator: "+",
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "2"},
				Value: 2,
			},
			want: "(1 + 2)",
		},
		{
			name: "multiplication",
			left: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "3"},
				Value: 3,
			},
			operator: "*",
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "4"},
				Value: 4,
			},
			want: "(3 * 4)",
		},
		{
			name: "comparison",
			left: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Value: "x",
			},
			operator: "<",
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "10"},
				Value: 10,
			},
			want: "(x < 10)",
		},
		{
			name: "nested expression",
			left: &BinaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Left:     &IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "1"}, Value: 1},
				Operator: "+",
				Right:    &IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2"}, Value: 2},
			},
			operator: "*",
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "3"},
				Value: 3,
			},
			want: "((1 + 2) * 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &BinaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: tt.operator},
				Left:     tt.left,
				Operator: tt.operator,
				Right:    tt.right,
			}

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
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "5"},
				Value: 5,
			},
			want: "(-5)",
		},
		{
			name:     "not",
			operator: "not",
			right: &BooleanLiteral{
				Token: lexer.Token{Type: lexer.TRUE, Literal: "true"},
				Value: true,
			},
			want: "(not true)",
		},
		{
			name:     "unary plus",
			operator: "+",
			right: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "3"},
				Value: 3,
			},
			want: "(+3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &UnaryExpression{
				Token:    lexer.Token{Type: lexer.MINUS, Literal: tt.operator},
				Operator: tt.operator,
				Right:    tt.right,
			}

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
			expr: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42"},
				Value: 42,
			},
			want: "(42)",
		},
		{
			name: "binary expression",
			expr: &BinaryExpression{
				Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Left: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "3"},
					Value: 3,
				},
				Operator: "+",
				Right: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "5"},
					Value: 5,
				},
			},
			want: "((3 + 5))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &GroupedExpression{
				Token:      lexer.Token{Type: lexer.LPAREN, Literal: "("},
				Expression: tt.expr,
			}

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
			expr: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42"},
				Value: 42,
			},
			want: "42",
		},
		{
			name: "identifier",
			expr: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
				Value: "x",
			},
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
				Token:      lexer.Token{Type: lexer.INT, Literal: "42"},
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
					Token: lexer.Token{Type: lexer.INT, Literal: "42"},
					Expression: &IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "42"},
						Value: 42,
					},
				},
			},
			want: "begin\n  42\nend",
		},
		{
			name: "multiple statements",
			stmts: []Statement{
				&ExpressionStatement{
					Token: lexer.Token{Type: lexer.INT, Literal: "1"},
					Expression: &IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "1"},
						Value: 1,
					},
				},
				&ExpressionStatement{
					Token: lexer.Token{Type: lexer.INT, Literal: "2"},
					Expression: &IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "2"},
						Value: 2,
					},
				},
			},
			want: "begin\n  1\n  2\nend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &BlockStatement{
				Token:      lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
				Statements: tt.stmts,
			}

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
			varType: &TypeAnnotation{Name: "Integer"},
			value: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42"},
				Value: 42,
			},
			want: "var x: Integer := 42",
		},
		{
			name:    "declaration with string initialization",
			varName: "s",
			varType: &TypeAnnotation{Name: "String"},
			value: &StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "'hello'"},
				Value: "hello",
			},
			want: "var s: String := \"hello\"",
		},
		{
			name:    "declaration without type",
			varName: "x",
			varType: nil,
			value: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "5"},
				Value: 5,
			},
			want: "var x := 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &VarDeclStatement{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var"},
				Names: []*Identifier{
					{
						Token: lexer.Token{Type: lexer.IDENT, Literal: tt.varName},
						Value: tt.varName,
					},
				},
				Type:  tt.varType,
				Value: tt.value,
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
			value: &IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "10"},
				Value: 10,
			},
			want: "x := 10",
		},
		{
			name:    "expression assignment",
			varName: "y",
			value: &BinaryExpression{
				Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
				Left: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x"},
					Value: "x",
				},
				Operator: "+",
				Right: &IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "1"},
					Value: 1,
				},
			},
			want: "y := (x + 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &AssignmentStatement{
				Token: lexer.Token{Type: lexer.ASSIGN, Literal: ":="},
				Target: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: tt.varName},
					Value: tt.varName,
				},
				Value: tt.value,
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
			name: "no arguments",
			function: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
				Value: "PrintLn",
			},
			arguments: []Expression{},
			want:      "PrintLn()",
		},
		{
			name: "single string argument",
			function: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
				Value: "PrintLn",
			},
			arguments: []Expression{
				&StringLiteral{
					Token: lexer.Token{Type: lexer.STRING, Literal: "'hello'"},
					Value: "hello",
				},
			},
			want: "PrintLn(\"hello\")",
		},
		{
			name: "multiple arguments",
			function: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Add"},
				Value: "Add",
			},
			arguments: []Expression{
				&IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "3"},
					Value: 3,
				},
				&IntegerLiteral{
					Token: lexer.Token{Type: lexer.INT, Literal: "5"},
					Value: 5,
				},
			},
			want: "Add(3, 5)",
		},
		{
			name: "expression arguments",
			function: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
				Value: "PrintLn",
			},
			arguments: []Expression{
				&BinaryExpression{
					Token: lexer.Token{Type: lexer.PLUS, Literal: "+"},
					Left: &IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "2"},
						Value: 2,
					},
					Operator: "+",
					Right: &IntegerLiteral{
						Token: lexer.Token{Type: lexer.INT, Literal: "3"},
						Value: 3,
					},
				},
			},
			want: "PrintLn((2 + 3))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &CallExpression{
				Token:     lexer.Token{Type: lexer.LPAREN, Literal: "("},
				Function:  tt.function,
				Arguments: tt.arguments,
			}

			if node.String() != tt.want {
				t.Errorf("String() = %q, want %q", node.String(), tt.want)
			}
		})
	}
}

// TestForInStatement tests the ForInStatement node.
func TestForInStatement(t *testing.T) {
	tests := []struct {
		name       string
		variable   string
		collection Expression
		body       Statement
		inlineVar  bool
		want       string
	}{
		{
			name:     "basic for-in",
			variable: "e",
			collection: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "mySet"},
				Value: "mySet",
			},
			body: &ExpressionStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
				Expression: &CallExpression{
					Token: lexer.Token{Type: lexer.LPAREN, Literal: "("},
					Function: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "PrintLn"},
						Value: "PrintLn",
					},
					Arguments: []Expression{
						&Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "e"},
							Value: "e",
						},
					},
				},
			},
			inlineVar: false,
			want:      "for e in mySet do PrintLn(e)",
		},
		{
			name:     "for-in with inline var",
			variable: "item",
			collection: &Identifier{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "myArray"},
				Value: "myArray",
			},
			body: &BlockStatement{
				Token: lexer.Token{Type: lexer.BEGIN, Literal: "begin"},
				Statements: []Statement{
					&ExpressionStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Process"},
						Expression: &CallExpression{
							Token: lexer.Token{Type: lexer.LPAREN, Literal: "("},
							Function: &Identifier{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "Process"},
								Value: "Process",
							},
							Arguments: []Expression{
								&Identifier{
									Token: lexer.Token{Type: lexer.IDENT, Literal: "item"},
									Value: "item",
								},
							},
						},
					},
				},
			},
			inlineVar: true,
			want:      "for var item in myArray do begin\n  Process(item)\nend",
		},
		{
			name:     "for-in with string literal",
			variable: "ch",
			collection: &StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "'hello'"},
				Value: "hello",
			},
			body: &ExpressionStatement{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Print"},
				Expression: &CallExpression{
					Token: lexer.Token{Type: lexer.LPAREN, Literal: "("},
					Function: &Identifier{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Print"},
						Value: "Print",
					},
					Arguments: []Expression{
						&Identifier{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "ch"},
							Value: "ch",
						},
					},
				},
			},
			inlineVar: false,
			want:      "for ch in \"hello\" do Print(ch)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &ForInStatement{
				Token: lexer.Token{Type: lexer.FOR, Literal: "for"},
				Variable: &Identifier{
					Token: lexer.Token{Type: lexer.IDENT, Literal: tt.variable},
					Value: tt.variable,
				},
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
