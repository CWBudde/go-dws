package bytecode

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestVM_RunArithmeticProgram(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}

	program := &ast.Program{Statements: []ast.Statement{
		&ast.VarDeclStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
			},
			Names: []*ast.Identifier{{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
					},
					Type: intType,
				},
				Value: "x",
			}},
			Type: intType,
			Value: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 10)},
					},
					Type: intType,
				},
				Value: 2,
			},
		},
		&ast.VarDeclStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
			},
			Names: []*ast.Identifier{{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(2, 5)},
					},
					Type: intType,
				},
				Value: "y",
			}},
			Type: intType,
			Value: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(2, 10)},
					},
					Type: intType,
				},
				Value: 3,
			},
		},
		&ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
				},
			ReturnValue: &ast.BinaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(3, 14)},
					},
					Type: intType,
				},
				Operator: "+",
				Left: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.ASTERISK, Literal: "*", Pos: pos(3, 9)},
						},
						Type: intType,
					},
					Operator: "*",
					Left: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 8)},
							},
							Type: intType,
						},
						Value: "x",
					},
					Right: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(3, 12)},
							},
							Type: intType,
						},
						Value: "y",
					},
				},
				Right: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "4", Pos: pos(3, 18)},
						},
						Type: intType,
					},
					Value: 4,
				},
			},
		},
	}}

	runVMAndCompare(t, program)
}

func TestVM_RunWhileLoop(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	xIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)},
			},
			Type: intType,
		},
		Value: "x",
	}
	yIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(2, 5)},
			},
			Type: intType,
		},
		Value: "y",
	}

	program := &ast.Program{Statements: []ast.Statement{
		&ast.VarDeclStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
			},
			Names: []*ast.Identifier{xIdent},
			Type:  intType,
			Value: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 10)},
					},
					Type: intType,
				},
				Value: 2,
			},
		},
		&ast.VarDeclStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
			},
			Names: []*ast.Identifier{yIdent},
			Type:  intType,
			Value: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(2, 10)},
					},
					Type: intType,
				},
				Value: 3,
			},
		},
		&ast.WhileStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.WHILE, Literal: "while", Pos: pos(3, 1)},
			},
			Condition: &ast.BinaryExpression{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.LESS, Literal: "<", Pos: pos(3, 10)},
					},
					Type: boolType,
				},
				Operator: "<",
				Left: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 8)},
						},
						Type: intType,
					},
					Value: "x",
				},
				Right: &ast.IntegerLiteral{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.INT, Literal: "10", Pos: pos(3, 12)},
						},
						Type: intType,
					},
					Value: 10,
				},
			},
			Body: &ast.AssignmentStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 3)},
				},
				Operator: lexer.ASSIGN,
				Target: &ast.Identifier{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 3)},
						},
						Type: intType,
					},
					Value: "x",
				},
				Value: &ast.BinaryExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{
							Token: lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(4, 8)},
						},
						Type: intType,
					},
					Operator: "+",
					Left: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 7)},
							},
							Type: intType,
						},
						Value: "x",
					},
					Right: &ast.Identifier{
						TypedExpressionBase: ast.TypedExpressionBase{
							BaseNode: ast.BaseNode{
								Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(4, 11)},
							},
							Type: intType,
						},
						Value: "y",
					},
				},
			},
		},
		&ast.ReturnStatement{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
				},
			ReturnValue: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(5, 9)},
					},
					Type: intType,
				},
				Value: "x",
			},
		},
	}}

	runVMAndCompare(t, program)
}

func TestVM_FunctionCall(t *testing.T) {
	fnChunk := NewChunk("add_one")
	fnChunk.LocalCount = 1
	constOne := fnChunk.AddConstant(IntValue(1))
	fnChunk.Write(OpLoadLocal, 0, 0, 1)
	fnChunk.Write(OpLoadConst, 0, uint16(constOne), 1)
	fnChunk.WriteSimple(OpAddInt, 1)
	fnChunk.Write(OpReturn, 1, 0, 1)

	function := NewFunctionObject("add_one", fnChunk, 1)
	fnValue := FunctionValue(function)

	mainChunk := NewChunk("main")
	fnIndex := mainChunk.AddConstant(fnValue)
	argIndex := mainChunk.AddConstant(IntValue(41))
	mainChunk.Write(OpLoadConst, 0, uint16(fnIndex), 1)
	mainChunk.Write(OpLoadConst, 0, uint16(argIndex), 1)
	mainChunk.Write(OpCallIndirect, 1, 0, 1)
	mainChunk.Write(OpReturn, 1, 0, 1)

	vm := NewVM()
	result, err := vm.Run(mainChunk)
	if err != nil {
		t.Fatalf("VM Run error = %v", err)
	}

	expected := IntValue(42)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}
}

func TestVM_GlobalLoadStore(t *testing.T) {
	chunk := NewChunk("globals")
	valIdx := chunk.AddConstant(IntValue(99))
	chunk.Write(OpLoadConst, 0, uint16(valIdx), 1)
	chunk.Write(OpStoreGlobal, 0, 0, 1)
	chunk.Write(OpLoadGlobal, 0, 0, 2)
	chunk.Write(OpReturn, 1, 0, 2)

	vm := NewVM()
	result, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM Run error = %v", err)
	}

	expected := IntValue(99)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}

	if !valueEqual(vm.GetGlobal(0), expected) {
		t.Fatalf("global[0] = %v, want %v", vm.GetGlobal(0), expected)
	}
}

func TestVM_RuntimeErrorIncludesStackTrace(t *testing.T) {
	divChunk := NewChunk("div")
	tenIdx := divChunk.AddConstant(IntValue(10))
	zeroIdx := divChunk.AddConstant(IntValue(0))
	divChunk.Write(OpLoadConst, 0, uint16(tenIdx), 1)
	divChunk.Write(OpLoadConst, 0, uint16(zeroIdx), 1)
	divChunk.WriteSimple(OpDivInt, 1)
	divChunk.Write(OpReturn, 1, 0, 1)

	divFn := NewFunctionObject("DivideWithZero", divChunk, 0)

	callChunk := NewChunk("callDiv")
	fnIdx := callChunk.AddConstant(FunctionValue(divFn))
	callChunk.Write(OpLoadConst, 0, uint16(fnIdx), 1)
	callChunk.Write(OpCallIndirect, 0, 0, 1)
	callChunk.Write(OpReturn, 1, 0, 1)

	callFn := NewFunctionObject("CallDivide", callChunk, 0)

	mainChunk := NewChunk("main")
	callFnIdx := mainChunk.AddConstant(FunctionValue(callFn))
	mainChunk.Write(OpLoadConst, 0, uint16(callFnIdx), 1)
	mainChunk.Write(OpCallIndirect, 0, 0, 1)
	mainChunk.Write(OpReturn, 1, 0, 1)

	vm := NewVM()
	_, err := vm.Run(mainChunk)
	if err == nil {
		t.Fatalf("expected runtime error, got nil")
	}

	runtimeErr, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected *RuntimeError, got %T", err)
	}

	if runtimeErr.Trace.Depth() < 2 {
		t.Fatalf("expected stack trace depth >= 2, got %d", runtimeErr.Trace.Depth())
	}

	traceStr := runtimeErr.Trace.String()
	if !strings.Contains(traceStr, "DivideWithZero") || !strings.Contains(traceStr, "CallDivide") {
		t.Fatalf("stack trace missing function names: %s", traceStr)
	}

	if !strings.Contains(runtimeErr.Message, "integer division by zero") {
		t.Fatalf("error message missing detail: %s", runtimeErr.Message)
	}
}

func TestVM_IncDecInt(t *testing.T) {
	chunk := NewChunk("inc_dec")
	startIdx := chunk.AddConstant(IntValue(7))
	chunk.Write(OpLoadConst, 0, uint16(startIdx), 1)
	chunk.WriteSimple(OpIncInt, 1)
	chunk.WriteSimple(OpDecInt, 1)
	chunk.WriteSimple(OpIncInt, 1)
	chunk.Write(OpReturn, 1, 0, 1)

	result := runChunk(t, chunk)
	expected := IntValue(8)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}
}

func TestVM_IntegerBitwiseInstructions(t *testing.T) {
	tests := []struct {
		want  Value
		build func(chunk *Chunk)
		name  string
	}{
		{
			name: "BitAnd",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(IntValue(12))  // 1100
				rightIdx := chunk.AddConstant(IntValue(10)) // 1010
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpBitAnd, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(8), // 1000
		},
		{
			name: "BitOr",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(IntValue(9))  // 1001
				rightIdx := chunk.AddConstant(IntValue(6)) // 0110
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpBitOr, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(15), // 1111
		},
		{
			name: "BitXor",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(IntValue(5))  // 0101
				rightIdx := chunk.AddConstant(IntValue(3)) // 0011
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpBitXor, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(6), // 0110
		},
		{
			name: "BitNot",
			build: func(chunk *Chunk) {
				valueIdx := chunk.AddConstant(IntValue(42))
				chunk.Write(OpLoadConst, 0, uint16(valueIdx), 1)
				chunk.WriteSimple(OpBitNot, 1)
				chunk.WriteSimple(OpBitNot, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk(tt.name)
			tt.build(chunk)
			result := runChunk(t, chunk)
			if !valueEqual(result, tt.want) {
				t.Fatalf("VM value = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestVM_LogicalXor(t *testing.T) {
	tests := []struct {
		name  string
		left  bool
		right bool
		want  bool
	}{
		{"TrueFalse", true, false, true},
		{"FalseTrue", false, true, true},
		{"TrueTrue", true, true, false},
		{"FalseFalse", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk("xor")
			if tt.left {
				chunk.WriteSimple(OpLoadTrue, 1)
			} else {
				chunk.WriteSimple(OpLoadFalse, 1)
			}
			if tt.right {
				chunk.WriteSimple(OpLoadTrue, 1)
			} else {
				chunk.WriteSimple(OpLoadFalse, 1)
			}
			chunk.WriteSimple(OpXor, 1)
			chunk.Write(OpReturn, 1, 0, 1)

			result := runChunk(t, chunk)
			expected := BoolValue(tt.want)
			if !valueEqual(result, expected) {
				t.Fatalf("VM value = %v, want %v", result, expected)
			}
		})
	}
}

func TestVM_TypeConversions(t *testing.T) {
	t.Run("IntToFloat", func(t *testing.T) {
		chunk := NewChunk("int_to_float")
		valueIdx := chunk.AddConstant(IntValue(42))
		chunk.Write(OpLoadConst, 0, uint16(valueIdx), 1)
		chunk.WriteSimple(OpIntToFloat, 1)
		chunk.Write(OpReturn, 1, 0, 1)

		result := runChunk(t, chunk)
		expected := FloatValue(42)
		if !valueEqual(result, expected) {
			t.Fatalf("VM value = %v, want %v", result, expected)
		}
	})

	t.Run("FloatToInt", func(t *testing.T) {
		chunk := NewChunk("float_to_int")
		valueIdx := chunk.AddConstant(FloatValue(3.75))
		chunk.Write(OpLoadConst, 0, uint16(valueIdx), 1)
		chunk.WriteSimple(OpFloatToInt, 1)
		chunk.Write(OpReturn, 1, 0, 1)

		result := runChunk(t, chunk)
		expected := IntValue(3)
		if !valueEqual(result, expected) {
			t.Fatalf("VM value = %v, want %v", result, expected)
		}
	})
}

func TestVM_CompareInstructions(t *testing.T) {
	tests := []struct {
		want  Value
		build func(chunk *Chunk)
		name  string
	}{
		{
			name: "CompareIntLess",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(IntValue(3))
				rightIdx := chunk.AddConstant(IntValue(5))
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpCompareInt, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(-1),
		},
		{
			name: "CompareIntEqual",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(IntValue(4))
				rightIdx := chunk.AddConstant(IntValue(4))
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpCompareInt, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(0),
		},
		{
			name: "CompareFloatGreater",
			build: func(chunk *Chunk) {
				leftIdx := chunk.AddConstant(FloatValue(4.5))
				rightIdx := chunk.AddConstant(FloatValue(2.25))
				chunk.Write(OpLoadConst, 0, uint16(leftIdx), 1)
				chunk.Write(OpLoadConst, 0, uint16(rightIdx), 1)
				chunk.WriteSimple(OpCompareFloat, 1)
				chunk.Write(OpReturn, 1, 0, 1)
			},
			want: IntValue(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := NewChunk(tt.name)
			tt.build(chunk)
			result := runChunk(t, chunk)
			if !valueEqual(result, tt.want) {
				t.Fatalf("VM value = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestVM_PropertyInstructions(t *testing.T) {
	chunk := NewChunk("properties")
	chunk.LocalCount = 1

	obj := NewObjectInstance("Point")
	objIdx := chunk.AddConstant(ObjectValue(obj))
	propIdx := chunk.AddConstant(StringValue("x"))
	valIdx := chunk.AddConstant(IntValue(42))

	chunk.Write(OpLoadConst, 0, uint16(objIdx), 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(valIdx), 1)
	chunk.Write(OpSetProperty, 0, uint16(propIdx), 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpGetProperty, 0, uint16(propIdx), 1)
	chunk.Write(OpReturn, 1, 0, 1)

	result := runChunk(t, chunk)
	expected := IntValue(42)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}
}

func TestVM_FieldInstructions(t *testing.T) {
	obj := NewObjectInstance("Point")
	obj.SetField("x", IntValue(7))

	t.Run("GetField", func(t *testing.T) {
		chunk := NewChunk("field_get")
		objIdx := chunk.AddConstant(ObjectValue(obj))
		fieldIdx := chunk.AddConstant(StringValue("x"))

		chunk.Write(OpLoadConst, 0, uint16(objIdx), 1)
		chunk.Write(OpGetField, 0, uint16(fieldIdx), 1)
		chunk.Write(OpReturn, 1, 0, 1)

		result := runChunk(t, chunk)
		if !valueEqual(result, IntValue(7)) {
			t.Fatalf("VM value = %v, want 7", result)
		}
	})

	t.Run("SetField", func(t *testing.T) {
		chunk := NewChunk("field_set")
		chunk.LocalCount = 1
		objIdx := chunk.AddConstant(ObjectValue(obj))
		fieldIdx := chunk.AddConstant(StringValue("x"))
		valIdx := chunk.AddConstant(IntValue(11))

		chunk.Write(OpLoadConst, 0, uint16(objIdx), 1)
		chunk.Write(OpStoreLocal, 0, 0, 1)
		chunk.Write(OpLoadLocal, 0, 0, 1)
		chunk.Write(OpLoadConst, 0, uint16(valIdx), 1)
		chunk.Write(OpSetField, 0, uint16(fieldIdx), 1)
		chunk.Write(OpLoadLocal, 0, 0, 1)
		chunk.Write(OpGetField, 0, uint16(fieldIdx), 1)
		chunk.Write(OpReturn, 1, 0, 1)

		result := runChunk(t, chunk)
		if !valueEqual(result, IntValue(11)) {
			t.Fatalf("VM value = %v, want 11", result)
		}
	})
}

func TestVM_ArrayInstructions(t *testing.T) {
	chunk := NewChunk("array_ops")
	chunk.LocalCount = 2
	valTenIdx := chunk.AddConstant(IntValue(10))
	valTwentyIdx := chunk.AddConstant(IntValue(20))
	valNinetyNineIdx := chunk.AddConstant(IntValue(99))
	indexOneIdx := chunk.AddConstant(IntValue(1))

	// Build array [10, 20]
	chunk.Write(OpLoadConst, 0, uint16(valTenIdx), 1)
	chunk.Write(OpLoadConst, 0, uint16(valTwentyIdx), 1)
	chunk.Write(OpNewArray, 0, 2, 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)

	// Capture length
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.WriteSimple(OpArrayLength, 1)
	chunk.Write(OpStoreLocal, 0, 1, 1)

	// arr[1] := 99
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(indexOneIdx), 1)
	chunk.Write(OpLoadConst, 0, uint16(valNinetyNineIdx), 1)
	chunk.WriteSimple(OpArraySet, 1)

	// Result := arr[1] + length
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(indexOneIdx), 1)
	chunk.WriteSimple(OpArrayGet, 1)
	chunk.Write(OpLoadLocal, 0, 1, 1)
	chunk.WriteSimple(OpAddInt, 1)
	chunk.Write(OpReturn, 1, 0, 1)

	result := runChunk(t, chunk)
	if !valueEqual(result, IntValue(101)) {
		t.Fatalf("VM value = %v, want 101", result)
	}
}

func TestVM_NewObjectAndInvoke(t *testing.T) {
	methodChunk := NewChunk("Compute")
	constIdx := methodChunk.AddConstant(IntValue(77))
	methodChunk.Write(OpLoadConst, 0, uint16(constIdx), 1)
	methodChunk.Write(OpReturn, 1, 0, 1)
	methodFn := NewFunctionObject("Compute", methodChunk, 0)

	chunk := NewChunk("invoke")
	chunk.LocalCount = 1
	classIdx := chunk.AddConstant(StringValue("TWidget"))
	methodNameIdx := chunk.AddConstant(StringValue("compute"))
	methodFnIdx := chunk.AddConstant(FunctionValue(methodFn))

	chunk.Write(OpNewObject, 0, uint16(classIdx), 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(methodFnIdx), 1)
	chunk.Write(OpSetField, 0, uint16(methodNameIdx), 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpInvoke, 0, uint16(methodNameIdx), 1)
	chunk.Write(OpReturn, 1, 0, 1)

	result := runChunk(t, chunk)
	if !valueEqual(result, IntValue(77)) {
		t.Fatalf("VM value = %v, want 77", result)
	}
}

func TestVM_TryFinallyExecutes(t *testing.T) {
	chunk := NewChunk("try_finally")
	chunk.LocalCount = 1
	oneIdx := chunk.AddConstant(IntValue(1))
	twoIdx := chunk.AddConstant(IntValue(2))

	tryIdx := chunk.Write(OpTry, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(oneIdx), 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	jumpToFinally := chunk.EmitJump(OpJump, 1)

	finallyStart := len(chunk.Code)
	chunk.Write(OpFinally, 0, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(twoIdx), 1)
	chunk.WriteSimple(OpAddInt, 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	chunk.Write(OpFinally, 1, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpReturn, 1, 0, 1)

	setInstructionTarget(t, chunk, jumpToFinally, finallyStart)
	setInstructionTarget(t, chunk, tryIdx, finallyStart)
	chunk.SetTryInfo(tryIdx, TryInfo{
		CatchTarget:   -1,
		FinallyTarget: finallyStart,
		HasCatch:      false,
		HasFinally:    true,
	})

	result := runChunk(t, chunk)
	if !valueEqual(result, IntValue(3)) {
		t.Fatalf("VM value = %v, want 3", result)
	}
}

func TestVM_TryExceptHandlesException(t *testing.T) {
	chunk := NewChunk("try_except")
	chunk.LocalCount = 1
	excIdx := chunk.AddConstant(ObjectValue(NewObjectInstance("Exception")))
	fortyTwoIdx := chunk.AddConstant(IntValue(42))

	tryIdx := chunk.Write(OpTry, 0, 0, 1)
	chunk.Write(OpLoadConst, 0, uint16(excIdx), 1)
	chunk.WriteSimple(OpThrow, 1)
	jumpAfterTry := chunk.EmitJump(OpJump, 1)

	catchStart := len(chunk.Code)
	chunk.Write(OpCatch, 0, 0, 1)
	chunk.WriteSimple(OpPop, 1)
	chunk.Write(OpLoadConst, 0, uint16(fortyTwoIdx), 1)
	chunk.Write(OpStoreLocal, 0, 0, 1)
	afterCatchJump := chunk.EmitJump(OpJump, 1)

	finallyStart := len(chunk.Code)
	chunk.Write(OpFinally, 0, 0, 1)
	chunk.Write(OpFinally, 1, 0, 1)
	chunk.Write(OpLoadLocal, 0, 0, 1)
	chunk.Write(OpReturn, 1, 0, 1)

	setInstructionTarget(t, chunk, jumpAfterTry, finallyStart)
	setInstructionTarget(t, chunk, afterCatchJump, finallyStart)
	setInstructionTarget(t, chunk, catchStart, finallyStart)
	setInstructionTarget(t, chunk, tryIdx, catchStart)
	chunk.SetTryInfo(tryIdx, TryInfo{
		CatchTarget:   catchStart,
		FinallyTarget: finallyStart,
		HasCatch:      true,
		HasFinally:    true,
	})

	result := runChunk(t, chunk)
	if !valueEqual(result, IntValue(42)) {
		t.Fatalf("VM value = %v, want 42", result)
	}
}

func TestVM_ThrowWithoutHandlerFails(t *testing.T) {
	chunk := NewChunk("throw_only")
	excIdx := chunk.AddConstant(ObjectValue(NewObjectInstance("Exception")))
	chunk.Write(OpLoadConst, 0, uint16(excIdx), 1)
	chunk.WriteSimple(OpThrow, 1)
	chunk.WriteSimple(OpReturn, 1)

	err := runChunkExpectError(t, chunk)
	if err == nil {
		t.Fatalf("expected runtime error")
	}
}

func TestVM_TypedExceptHandlerMatches(t *testing.T) {
	program := buildTypedExceptionProgram("MyError", "MyError", 5, nil)
	result := runCompiledProgram(t, "typed_match", program)
	if !valueEqual(result, IntValue(5)) {
		t.Fatalf("VM value = %v, want 5", result)
	}
}

func TestVM_TypedExceptElseExecutes(t *testing.T) {
	elseValue := 9
	program := buildTypedExceptionProgram("OtherError", "MyError", 5, &elseValue)
	result := runCompiledProgram(t, "typed_else", program)
	if !valueEqual(result, IntValue(9)) {
		t.Fatalf("VM value = %v, want 9", result)
	}
}

func TestVM_TypedExceptRethrowsWhenNoMatch(t *testing.T) {
	program := buildTypedExceptionProgram("OtherError", "MyError", 5, nil)
	if err := runCompiledProgramExpectError(t, "typed_rethrow", program); err == nil {
		t.Fatalf("expected rethrow when no handler or else matches")
	}
}

func TestVM_CallMethodUsesSelf(t *testing.T) {
	obj := NewObjectInstance("Counter")
	obj.SetField("value", IntValue(99))

	methodChunk := NewChunk("GetValue")
	fieldNameIdx := methodChunk.AddConstant(StringValue("value"))
	methodChunk.WriteSimple(OpGetSelf, 1)
	methodChunk.Write(OpGetField, 0, uint16(fieldNameIdx), 1)
	methodChunk.Write(OpReturn, 1, 0, 1)

	methodFn := NewFunctionObject("GetValue", methodChunk, 0)
	obj.SetField("getvalue", FunctionValue(methodFn))

	mainChunk := NewChunk("main_method_call")
	objIdx := mainChunk.AddConstant(ObjectValue(obj))
	methodNameIdx := mainChunk.AddConstant(StringValue("getvalue"))
	mainChunk.Write(OpLoadConst, 0, uint16(objIdx), 1)
	mainChunk.Write(OpCallMethod, 0, uint16(methodNameIdx), 1)
	mainChunk.Write(OpReturn, 1, 0, 1)

	result := runChunk(t, mainChunk)
	expected := IntValue(99)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}
}

func TestVM_ClosureCapturesUpvalue(t *testing.T) {
	adderChunk := NewChunk("adder")
	adderChunk.LocalCount = 1
	adderChunk.Write(OpLoadUpvalue, 0, 0, 1)
	adderChunk.Write(OpLoadLocal, 0, 0, 1)
	adderChunk.WriteSimple(OpAddInt, 1)
	adderChunk.Write(OpReturn, 1, 0, 1)

	adderFn := NewFunctionObject("adder", adderChunk, 1)
	adderFn.UpvalueDefs = []UpvalueDef{
		{IsLocal: true, Index: 0},
	}

	makeAdderChunk := NewChunk("makeAdder")
	makeAdderChunk.LocalCount = 1
	fnConstIdx := makeAdderChunk.AddConstant(FunctionValue(adderFn))
	makeAdderChunk.Write(OpClosure, 1, uint16(fnConstIdx), 1)
	makeAdderChunk.Write(OpReturn, 1, 0, 1)

	makeAdderFn := NewFunctionObject("makeAdder", makeAdderChunk, 1)

	mainChunk := NewChunk("main")
	fnIdx := mainChunk.AddConstant(FunctionValue(makeAdderFn))
	fiveIdx := mainChunk.AddConstant(IntValue(5))
	tenIdx := mainChunk.AddConstant(IntValue(10))

	mainChunk.Write(OpLoadConst, 0, uint16(fnIdx), 1)
	mainChunk.Write(OpLoadConst, 0, uint16(fiveIdx), 1)
	mainChunk.Write(OpCallIndirect, 1, 0, 1)
	mainChunk.Write(OpLoadConst, 0, uint16(tenIdx), 2)
	mainChunk.Write(OpCallIndirect, 1, 0, 2)
	mainChunk.Write(OpReturn, 1, 0, 2)

	vm := NewVM()
	result, err := vm.Run(mainChunk)
	if err != nil {
		t.Fatalf("VM Run error = %v", err)
	}

	expected := IntValue(15)
	if !valueEqual(result, expected) {
		t.Fatalf("VM value = %v, want %v", result, expected)
	}
}

func runChunk(t *testing.T, chunk *Chunk) Value {
	t.Helper()

	vm := NewVM()
	result, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM Run error = %v", err)
	}
	return result
}

func runVMAndCompare(t *testing.T, program *ast.Program) {
	t.Helper()

	compiler := NewCompiler("vm_test")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	vm := NewVM()
	vmValue, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("VM Run error = %v", err)
	}

	interpreter := interp.New(io.Discard)
	interpValue := interpreter.Eval(program)
	expected, err := convertInterpreterValue(interpValue)
	if err != nil {
		t.Fatalf("convert interpreter value: %v", err)
	}

	if !valueEqual(vmValue, expected) {
		t.Fatalf("VM value = %v, want %v", vmValue, expected)
	}
}

func convertInterpreterValue(v interp.Value) (Value, error) {
	switch val := v.(type) {
	case *interp.IntegerValue:
		return IntValue(val.Value), nil
	case *interp.FloatValue:
		return FloatValue(val.Value), nil
	case *interp.StringValue:
		return StringValue(val.Value), nil
	case *interp.BooleanValue:
		return BoolValue(val.Value), nil
	case *interp.NilValue:
		return NilValue(), nil
	case nil:
		return NilValue(), nil
	default:
		return Value{}, fmt.Errorf("unsupported interpreter value %T", v)
	}
}

func compileProgramChunk(t *testing.T, name string, program *ast.Program) *Chunk {
	t.Helper()
	compiler := NewCompiler(name)
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return chunk
}

func runCompiledProgram(t *testing.T, name string, program *ast.Program) Value {
	t.Helper()
	chunk := compileProgramChunk(t, name, program)
	return runChunk(t, chunk)
}

func runCompiledProgramExpectError(t *testing.T, name string, program *ast.Program) error {
	t.Helper()
	chunk := compileProgramChunk(t, name, program)
	return runChunkExpectError(t, chunk)
}

func buildTypedExceptionProgram(thrownClass, handlerClass string, handlerValue int, elseValue *int) *ast.Program {
	accIdent := &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(1, 5)},
			},
		},
		Value: "acc",
	}
	varDecl := &ast.VarDeclStatement{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
		},
		Names: []*ast.Identifier{accIdent},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: "0", Pos: pos(1, 10)},
				},
			},
			Value: 0,
		},
	}

	raiseStmt := &ast.RaiseStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.RAISE, Literal: "raise", Pos: pos(2, 3)},
			},
		Exception: &ast.NewExpression{
			Token: lexer.Token{Type: lexer.NEW, Literal: "new", Pos: pos(2, 9)},
			ClassName: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: thrownClass, Pos: pos(2, 13)},
					},
				},
				Value: thrownClass,
			},
		},
	}
	tryBlock := &ast.BlockStatement{Statements: []ast.Statement{raiseStmt}}

	handlerAssign := &ast.AssignmentStatement{
		BaseNode: ast.BaseNode{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(3, 3)},
		},
		Operator: lexer.ASSIGN,
		Target: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(3, 3)},
				},
			},
			Value: "acc",
		},
		Value: &ast.IntegerLiteral{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.INT, Literal: fmt.Sprintf("%d", handlerValue), Pos: pos(3, 10)},
				},
			},
			Value: int64(handlerValue),
		},
	}

	handler := &ast.ExceptionHandler{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.ON, Literal: "on", Pos: pos(3, 1)},
			},
		Variable: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "E", Pos: pos(3, 4)},
				},
			},
			Value: "E",
		},
		ExceptionType: &ast.TypeAnnotation{Name: handlerClass},
		Statement:     &ast.BlockStatement{Statements: []ast.Statement{handlerAssign}},
	}

	clause := &ast.ExceptClause{
	BaseNode: ast.BaseNode{
		Token:    lexer.Token{Type: lexer.EXCEPT, Literal: "except", Pos: pos(3, 1)},
	},
		Handlers: []*ast.ExceptionHandler{handler},
	}
	if elseValue != nil {
		elseAssign := &ast.AssignmentStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(4, 3)},
			},
			Operator: lexer.ASSIGN,
			Target: &ast.Identifier{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(4, 3)},
					},
				},
				Value: "acc",
			},
			Value: &ast.IntegerLiteral{
				TypedExpressionBase: ast.TypedExpressionBase{
					BaseNode: ast.BaseNode{
						Token: lexer.Token{Type: lexer.INT, Literal: fmt.Sprintf("%d", *elseValue), Pos: pos(4, 10)},
					},
				},
				Value: int64(*elseValue),
			},
		}
		clause.ElseBlock = &ast.BlockStatement{Statements: []ast.Statement{elseAssign}}
	}

	tryStmt := &ast.TryStatement{
	BaseNode: ast.BaseNode{
		Token:        lexer.Token{Type: lexer.TRY, Literal: "try", Pos: pos(2, 1)},
	},
		TryBlock:     tryBlock,
		ExceptClause: clause,
	}

	returnStmt := &ast.ReturnStatement{
			BaseNode: ast.BaseNode{
				Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
			},
		ReturnValue: &ast.Identifier{
			TypedExpressionBase: ast.TypedExpressionBase{
				BaseNode: ast.BaseNode{
					Token: lexer.Token{Type: lexer.IDENT, Literal: "acc", Pos: pos(5, 9)},
				},
			},
			Value: "acc",
		},
	}

	return &ast.Program{Statements: []ast.Statement{varDecl, tryStmt, returnStmt}}
}

func runChunkExpectError(t *testing.T, chunk *Chunk) error {
	t.Helper()
	vm := NewVM()
	_, err := vm.Run(chunk)
	return err
}

func setInstructionTarget(t *testing.T, chunk *Chunk, instIndex, target int) {
	t.Helper()
	offset := target - instIndex - 1
	if offset > 32767 || offset < -32768 {
		t.Fatalf("offset out of range: %d", offset)
	}
	inst := chunk.Code[instIndex]
	chunk.Code[instIndex] = MakeInstruction(inst.OpCode(), inst.A(), uint16(offset))
}

// TestVariantToBool tests the variantToBool helper function
func TestVariantToBool(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		// Boolean values
		{"bool true", BoolValue(true), true},
		{"bool false", BoolValue(false), false},

		// Integer values
		{"int zero", IntValue(0), false},
		{"int non-zero positive", IntValue(42), true},
		{"int non-zero negative", IntValue(-5), true},

		// Float values
		{"float zero", FloatValue(0.0), false},
		{"float non-zero positive", FloatValue(3.14), true},
		{"float non-zero negative", FloatValue(-2.5), true},

		// String values
		{"string empty", StringValue(""), false},
		{"string non-empty", StringValue("hello"), true},

		// Nil value
		{"nil", NilValue(), false},

		// Nested Variant values
		{"variant wrapping true", VariantValue(BoolValue(true)), true},
		{"variant wrapping false", VariantValue(BoolValue(false)), false},
		{"variant wrapping zero", VariantValue(IntValue(0)), false},
		{"variant wrapping non-zero", VariantValue(IntValue(1)), true},
		{"variant wrapping nil", VariantValue(NilValue()), false},
		{"double-nested variant true", VariantValue(VariantValue(BoolValue(true))), true},
		{"double-nested variant false", VariantValue(VariantValue(IntValue(0))), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := variantToBool(tt.value)
			if result != tt.expected {
				t.Errorf("variantToBool(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestIsTruthy tests the isTruthy helper function
func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		// Boolean values
		{"bool true", BoolValue(true), true},
		{"bool false", BoolValue(false), false},

		// Variant values
		{"variant wrapping true", VariantValue(BoolValue(true)), true},
		{"variant wrapping false", VariantValue(BoolValue(false)), false},
		{"variant wrapping zero", VariantValue(IntValue(0)), false},
		{"variant wrapping non-zero", VariantValue(IntValue(42)), true},
		{"variant wrapping empty string", VariantValue(StringValue("")), false},
		{"variant wrapping non-empty string", VariantValue(StringValue("test")), true},
		{"variant wrapping nil", VariantValue(NilValue()), false},

		// Non-boolean, non-variant values (should return false)
		{"int value", IntValue(42), false},
		{"string value", StringValue("hello"), false},
		{"nil value", NilValue(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTruthy(tt.value)
			if result != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}
