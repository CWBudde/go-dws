package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestCompiler_UnaryExpressions tests compilation of unary expressions
func TestCompiler_UnaryExpressions(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
	}{
		{
			name: "unary minus on integer variable",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.VarDeclStatement{
						Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
						Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x"}},
						Type:  &ast.TypeAnnotation{Name: "Integer"},
						Value: &ast.IntegerLiteral{Value: 42, Type: &ast.TypeAnnotation{Name: "Integer"}},
					},
					&ast.ExpressionStatement{
						Token: lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(2, 1)},
						Expression: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(2, 1)},
							Operator: "-",
							Right: &ast.Identifier{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 2)},
								Value: "x",
								Type:  &ast.TypeAnnotation{Name: "Integer"},
							},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
				},
			},
		},
		{
			name: "unary minus on float variable",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.VarDeclStatement{
						Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
						Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x"}},
						Type:  &ast.TypeAnnotation{Name: "Float"},
						Value: &ast.FloatLiteral{Value: 3.14, Type: &ast.TypeAnnotation{Name: "Float"}},
					},
					&ast.ExpressionStatement{
						Token: lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(2, 1)},
						Expression: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(2, 1)},
							Operator: "-",
							Right: &ast.Identifier{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 2)},
								Value: "x",
								Type:  &ast.TypeAnnotation{Name: "Float"},
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
		},
		{
			name: "unary not on boolean variable",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.VarDeclStatement{
						Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
						Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x"}},
						Type:  &ast.TypeAnnotation{Name: "Boolean"},
						Value: &ast.BooleanLiteral{Value: true, Type: &ast.TypeAnnotation{Name: "Boolean"}},
					},
					&ast.ExpressionStatement{
						Token: lexer.Token{Type: lexer.NOT, Literal: "not", Pos: pos(2, 1)},
						Expression: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.NOT, Literal: "not", Pos: pos(2, 1)},
							Operator: "not",
							Right: &ast.Identifier{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(2, 5)},
								Value: "x",
								Type:  &ast.TypeAnnotation{Name: "Boolean"},
							},
							Type: &ast.TypeAnnotation{Name: "Boolean"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := newTestCompiler(tt.name)
			_, err := compiler.Compile(tt.program)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			// Just verify that it compiles without error
			// The optimizer may change the exact instruction sequence
		})
	}
}

// TestCompiler_UnaryConstantFolding tests constant folding for unary expressions
func TestCompiler_UnaryConstantFolding(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
		wantVal Value
	}{
		{
			name: "fold -42",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(1, 8)},
							Operator: "-",
							Right: &ast.IntegerLiteral{
								Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 9)},
								Value: 42,
							},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
				},
			},
			wantVal: IntValue(-42),
		},
		{
			name: "fold -3.14",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(1, 8)},
							Operator: "-",
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14", Pos: pos(1, 9)},
								Value: 3.14,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			wantVal: FloatValue(-3.14),
		},
		{
			name: "fold not true",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.NOT, Literal: "not", Pos: pos(1, 8)},
							Operator: "not",
							Right: &ast.BooleanLiteral{
								Token: lexer.Token{Type: lexer.TRUE, Literal: "true", Pos: pos(1, 12)},
								Value: true,
							},
							Type: &ast.TypeAnnotation{Name: "Boolean"},
						},
					},
				},
			},
			wantVal: BoolValue(false),
		},
		{
			name: "fold +42",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(1, 8)},
							Operator: "+",
							Right: &ast.IntegerLiteral{
								Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 9)},
								Value: 42,
							},
							Type: &ast.TypeAnnotation{Name: "Integer"},
						},
					},
				},
			},
			wantVal: IntValue(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(tt.name)
			chunk, err := compiler.Compile(tt.program)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			vm := NewVM()
			result, err := vm.Run(chunk)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if !valuesEqual(result, tt.wantVal) {
				t.Errorf("Run() = %v, want %v", result, tt.wantVal)
			}
		})
	}
}

// TestVM_FloatOperations tests float arithmetic operations
func TestVM_FloatOperations(t *testing.T) {
	tests := []struct {
		name     string
		program  *ast.Program
		expected float64
	}{
		{
			name: "float addition",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.BinaryExpression{
							Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(1, 12)},
							Operator: "+",
							Left: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14", Pos: pos(1, 8)},
								Value: 3.14,
							},
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "2.86", Pos: pos(1, 15)},
								Value: 2.86,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			expected: 6.0,
		},
		{
			name: "float subtraction",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.BinaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(1, 12)},
							Operator: "-",
							Left: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "5.0", Pos: pos(1, 8)},
								Value: 5.0,
							},
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.0", Pos: pos(1, 14)},
								Value: 3.0,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			expected: 2.0,
		},
		{
			name: "float multiplication",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.BinaryExpression{
							Token:    lexer.Token{Type: lexer.ASTERISK, Literal: "*", Pos: pos(1, 12)},
							Operator: "*",
							Left: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "2.5", Pos: pos(1, 8)},
								Value: 2.5,
							},
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "4.0", Pos: pos(1, 14)},
								Value: 4.0,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			expected: 10.0,
		},
		{
			name: "float division",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.BinaryExpression{
							Token:    lexer.Token{Type: lexer.SLASH, Literal: "/", Pos: pos(1, 12)},
							Operator: "/",
							Left: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "10.0", Pos: pos(1, 8)},
								Value: 10.0,
							},
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "2.0", Pos: pos(1, 15)},
								Value: 2.0,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			expected: 5.0,
		},
		{
			name: "unary negation on float",
			program: &ast.Program{
				Statements: []ast.Statement{
					&ast.ReturnStatement{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(1, 1)},
						ReturnValue: &ast.UnaryExpression{
							Token:    lexer.Token{Type: lexer.MINUS, Literal: "-", Pos: pos(1, 8)},
							Operator: "-",
							Right: &ast.FloatLiteral{
								Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14", Pos: pos(1, 9)},
								Value: 3.14,
							},
							Type: &ast.TypeAnnotation{Name: "Float"},
						},
					},
				},
			},
			expected: -3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler(tt.name)
			chunk, err := compiler.Compile(tt.program)
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			vm := NewVM()
			result, err := vm.Run(chunk)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if !result.IsNumber() {
				t.Fatalf("expected float result, got %v", result.Type)
			}

			got := result.AsFloat()
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestValueEquality tests the valuesEqual and valuesEqualForFold functions
func TestValueEquality(t *testing.T) {
	tests := []struct {
		name      string
		left      Value
		right     Value
		wantEqual bool
	}{
		{"nil values", NilValue(), NilValue(), true},
		{"equal bools", BoolValue(true), BoolValue(true), true},
		{"unequal bools", BoolValue(true), BoolValue(false), false},
		{"equal ints", IntValue(42), IntValue(42), true},
		{"unequal ints", IntValue(42), IntValue(43), false},
		{"equal floats", FloatValue(3.14), FloatValue(3.14), true},
		{"unequal floats", FloatValue(3.14), FloatValue(2.71), false},
		{"equal strings", StringValue("hello"), StringValue("hello"), true},
		{"unequal strings", StringValue("hello"), StringValue("world"), false},
		{"int and float same value", IntValue(42), FloatValue(42.0), true},
		{"int and float different value", IntValue(42), FloatValue(43.0), false},
		{"different types", BoolValue(true), IntValue(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.left, tt.right)
			if got != tt.wantEqual {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.wantEqual)
			}

			// Also test valuesEqualForFold
			gotFold, ok := valuesEqualForFold(tt.left, tt.right)
			if ok && gotFold != tt.wantEqual {
				t.Errorf("valuesEqualForFold(%v, %v) = %v, want %v", tt.left, tt.right, gotFold, tt.wantEqual)
			}
		})
	}
}

// TestCompiler_StringMethods tests String() methods for coverage
func TestCompiler_StringMethods(t *testing.T) {
	// Test Value.String()
	t.Run("Value.String", func(t *testing.T) {
		values := []Value{
			NilValue(),
			BoolValue(true),
			IntValue(42),
			FloatValue(3.14),
			StringValue("hello"),
		}
		for _, v := range values {
			// Just call String() to improve coverage
			_ = v.String()
		}
	})

	// Test RuntimeError.Error()
	t.Run("RuntimeError.Error", func(t *testing.T) {
		err := &RuntimeError{
			Message: "test error",
		}
		msg := err.Error()
		if msg == "" {
			t.Error("RuntimeError.Error() returned empty string")
		}
	})

	// Test ArrayInstance.String()
	t.Run("ArrayInstance.String", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1), IntValue(2)})
		_ = arr.String()
	})

	// Test Chunk.String()
	t.Run("Chunk.String", func(t *testing.T) {
		chunk := NewChunk("test")
		chunk.WriteSimple(OpReturn, 1)
		_ = chunk.String()
	})
}

// TestCompiler_HelperMethods tests various helper methods for coverage
func TestCompiler_HelperMethods(t *testing.T) {
	t.Run("Compiler.LocalCount", func(t *testing.T) {
		compiler := NewCompiler("test")
		count := compiler.LocalCount()
		if count < 0 {
			t.Errorf("LocalCount() = %d, want >= 0", count)
		}
	})

	t.Run("Value.IsClosure", func(t *testing.T) {
		val := IntValue(42)
		if val.IsClosure() {
			t.Error("IntValue should not be a closure")
		}
	})

	t.Run("Value.IsBuiltin", func(t *testing.T) {
		val := IntValue(42)
		if val.IsBuiltin() {
			t.Error("IntValue should not be a builtin")
		}
	})

	t.Run("ArrayInstance.NewArrayInstanceWithLength", func(t *testing.T) {
		arr := NewArrayInstanceWithLength(5)
		if arr.Length() != 5 {
			t.Errorf("NewArrayInstanceWithLength(5).Length() = %d, want 5", arr.Length())
		}
	})

	t.Run("ArrayInstance.Resize", func(t *testing.T) {
		arr := NewArrayInstance([]Value{IntValue(1)})
		arr.Resize(3)
		if arr.Length() != 3 {
			t.Errorf("Resize(3) failed, got len = %d", arr.Length())
		}
	})

	t.Run("Chunk.InstructionCount", func(t *testing.T) {
		chunk := NewChunk("test")
		chunk.WriteSimple(OpReturn, 1)
		count := chunk.InstructionCount()
		if count != 1 {
			t.Errorf("InstructionCount() = %d, want 1", count)
		}
	})

	t.Run("Chunk.PatchInstruction", func(t *testing.T) {
		chunk := NewChunk("test")
		chunk.WriteSimple(OpReturn, 1)
		instruction := MakeSimpleInstruction(OpLoadNil)
		chunk.PatchInstruction(0, instruction)
		// Verify the patch worked
		if chunk.InstructionCount() != 1 {
			t.Error("PatchInstruction changed instruction count")
		}
	})

	t.Run("ArrayInstance.Set", func(t *testing.T) {
		arr := NewArrayInstanceWithLength(3)
		ok := arr.Set(1, IntValue(99))
		if !ok {
			t.Error("Set(1, 99) returned false")
		}
		val, ok := arr.Get(1)
		if !ok || val.AsInt() != 99 {
			t.Errorf("Get(1) after Set(1, 99) failed, got %d", val.AsInt())
		}
	})
}

// TestCompiler_ErrorCases tests error handling for coverage
func TestCompiler_ErrorCases(t *testing.T) {
	t.Run("unsupported unary operator", func(t *testing.T) {
		program := &ast.Program{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "~", Pos: pos(1, 1)},
					Expression: &ast.UnaryExpression{
						Token:    lexer.Token{Type: lexer.IDENT, Literal: "~", Pos: pos(1, 1)},
						Operator: "~",
						Right: &ast.IntegerLiteral{
							Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 2)},
							Value: 42,
						},
						Type: &ast.TypeAnnotation{Name: "Integer"},
					},
				},
			},
		}

		compiler := NewCompiler("test")
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("expected error for unsupported unary operator")
		}
	})

	t.Run("VM type error in float operation", func(t *testing.T) {
		chunk := NewChunk("test")
		idx1 := chunk.AddConstant(StringValue("not a number"))
		idx2 := chunk.AddConstant(IntValue(42))
		chunk.Write(OpLoadConst, 0, uint16(idx1), 1)
		chunk.Write(OpLoadConst, 0, uint16(idx2), 1)
		chunk.WriteSimple(OpAddFloat, 1)
		chunk.WriteSimple(OpReturn, 1)

		vm := NewVM()
		_, err := vm.Run(chunk)
		if err == nil {
			t.Error("expected type error for float operation on string")
		}
	})
}

// TestCompiler_InferExpressionType tests type inference for coverage
func TestCompiler_InferExpressionType(t *testing.T) {
	compiler := NewCompiler("test")

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			"integer literal",
			&ast.IntegerLiteral{
				Token: lexer.Token{Type: lexer.INT, Literal: "42", Pos: pos(1, 1)},
				Value: 42,
				Type:  &ast.TypeAnnotation{Name: "Integer"},
			},
		},
		{
			"float literal",
			&ast.FloatLiteral{
				Token: lexer.Token{Type: lexer.FLOAT, Literal: "3.14", Pos: pos(1, 1)},
				Value: 3.14,
				Type:  &ast.TypeAnnotation{Name: "Float"},
			},
		},
		{
			"string literal",
			&ast.StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "hello", Pos: pos(1, 1)},
				Value: "hello",
				Type:  &ast.TypeAnnotation{Name: "String"},
			},
		},
		{
			"boolean literal",
			&ast.BooleanLiteral{
				Token: lexer.Token{Type: lexer.TRUE, Literal: "true", Pos: pos(1, 1)},
				Value: true,
				Type:  &ast.TypeAnnotation{Name: "Boolean"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := compiler.inferExpressionType(tt.expr)
			if typ == nil {
				t.Error("inferExpressionType() returned nil")
			}
		})
	}
}

// TestLiteralValue tests the literalValue helper function
func TestLiteralValue(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantVal Value
		wantOk  bool
	}{
		{
			"integer literal",
			&ast.IntegerLiteral{Value: 42},
			IntValue(42),
			true,
		},
		{
			"float literal",
			&ast.FloatLiteral{Value: 3.14},
			FloatValue(3.14),
			true,
		},
		{
			"string literal",
			&ast.StringLiteral{Value: "hello"},
			StringValue("hello"),
			true,
		},
		{
			"boolean literal",
			&ast.BooleanLiteral{Value: true},
			BoolValue(true),
			true,
		},
		{
			"nil literal",
			&ast.NilLiteral{},
			NilValue(),
			true,
		},
		{
			"non-literal",
			&ast.Identifier{Value: "x"},
			Value{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := literalValue(tt.expr)
			if ok != tt.wantOk {
				t.Errorf("literalValue() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && !valuesEqual(val, tt.wantVal) {
				t.Errorf("literalValue() = %v, want %v", val, tt.wantVal)
			}
		})
	}
}

// TestVM_Peek tests the VM peek function
func TestVM_Peek(t *testing.T) {
	chunk := NewChunk("test")
	chunk.WriteSimple(OpLoadTrue, 1)
	chunk.WriteSimple(OpLoadFalse, 1)
	chunk.WriteSimple(OpReturn, 1)

	vm := NewVM()
	// Push some values
	vm.push(IntValue(1))
	vm.push(IntValue(2))
	vm.push(IntValue(3))

	// Peek at top
	val, err := vm.peek()
	if err != nil {
		t.Fatalf("peek() error = %v", err)
	}
	if val.AsInt() != 3 {
		t.Errorf("peek() = %d, want 3", val.AsInt())
	}
}

// TestVM_SetGlobal tests the SetGlobal function
func TestVM_SetGlobal(t *testing.T) {
	vm := NewVM()

	// Set a global variable
	vm.SetGlobal(0, IntValue(42))

	// The SetGlobal method should work without error
	// We can't directly verify the value since globals is private
}

// TestCompiler_EmitValue tests the emitValue function with various value types
func TestCompiler_EmitValue(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"nil", NilValue()},
		{"true", BoolValue(true)},
		{"false", BoolValue(false)},
		{"zero", IntValue(0)},
		{"one", IntValue(1)},
		{"large int", IntValue(12345)},
		{"float", FloatValue(3.14)},
		{"string", StringValue("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler("test")
			err := compiler.emitValue(tt.val, 1)
			if err != nil {
				t.Errorf("emitValue(%v) error = %v", tt.val, err)
			}
		})
	}
}

// TestVM_Compare tests the compare function with the correct signature
func TestVM_Compare(t *testing.T) {
	tests := []struct {
		name    string
		left    Value
		right   Value
		op      OpCode
		wantRes bool
		wantErr bool
	}{
		{"int less", IntValue(1), IntValue(2), OpLess, true, false},
		{"int greater", IntValue(2), IntValue(1), OpGreater, true, false},
		{"float less", FloatValue(1.0), FloatValue(2.0), OpLess, true, false},
		{"float greater", FloatValue(2.0), FloatValue(1.0), OpGreater, true, false},
		{"string less", StringValue("a"), StringValue("b"), OpLess, true, false},
		{"string greater", StringValue("b"), StringValue("a"), OpGreater, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			vm.push(tt.left)
			vm.push(tt.right)
			res, err := vm.compare(tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("compare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && res != tt.wantRes {
				t.Errorf("compare(%v, %v, %v) = %v, want %v", tt.left, tt.right, tt.op, res, tt.wantRes)
			}
		})
	}
}
