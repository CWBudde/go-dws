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
			Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
			Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x", Type: intType}},
			Type:  intType,
			Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 10)}, Value: 2, Type: intType},
		},
		&ast.VarDeclStatement{
			Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
			Names: []*ast.Identifier{{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(2, 5)}, Value: "y", Type: intType}},
			Type:  intType,
			Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(2, 10)}, Value: 3, Type: intType},
		},
		&ast.ReturnStatement{
			Token: lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(3, 1)},
			ReturnValue: &ast.BinaryExpression{
				Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(3, 14)},
				Operator: "+",
				Left: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.ASTERISK, Literal: "*", Pos: pos(3, 9)},
					Operator: "*",
					Left:     &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 8)}, Value: "x", Type: intType},
					Right:    &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(3, 12)}, Value: "y", Type: intType},
					Type:     intType,
				},
				Right: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "4", Pos: pos(3, 18)}, Value: 4, Type: intType},
				Type:  intType,
			},
		},
	}}

	runVMAndCompare(t, program)
}

func TestVM_RunWhileLoop(t *testing.T) {
	intType := &ast.TypeAnnotation{Name: "Integer"}
	boolType := &ast.TypeAnnotation{Name: "Boolean"}

	xIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(1, 5)}, Value: "x", Type: intType}
	yIdent := &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(2, 5)}, Value: "y", Type: intType}

	program := &ast.Program{Statements: []ast.Statement{
		&ast.VarDeclStatement{
			Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(1, 1)},
			Names: []*ast.Identifier{xIdent},
			Type:  intType,
			Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "2", Pos: pos(1, 10)}, Value: 2, Type: intType},
		},
		&ast.VarDeclStatement{
			Token: lexer.Token{Type: lexer.VAR, Literal: "var", Pos: pos(2, 1)},
			Names: []*ast.Identifier{yIdent},
			Type:  intType,
			Value: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "3", Pos: pos(2, 10)}, Value: 3, Type: intType},
		},
		&ast.WhileStatement{
			Token:     lexer.Token{Type: lexer.WHILE, Literal: "while", Pos: pos(3, 1)},
			Condition: &ast.BinaryExpression{Token: lexer.Token{Type: lexer.LESS, Literal: "<", Pos: pos(3, 10)}, Operator: "<", Left: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(3, 8)}, Value: "x", Type: intType}, Right: &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "10", Pos: pos(3, 12)}, Value: 10, Type: intType}, Type: boolType},
			Body: &ast.AssignmentStatement{
				Token:    lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 3)},
				Operator: lexer.ASSIGN,
				Target:   &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 3)}, Value: "x", Type: intType},
				Value: &ast.BinaryExpression{
					Token:    lexer.Token{Type: lexer.PLUS, Literal: "+", Pos: pos(4, 8)},
					Operator: "+",
					Left:     &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(4, 7)}, Value: "x", Type: intType},
					Right:    &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "y", Pos: pos(4, 11)}, Value: "y", Type: intType},
					Type:     intType,
				},
			},
		},
		&ast.ReturnStatement{
			Token:       lexer.Token{Type: lexer.IDENT, Literal: "Result", Pos: pos(5, 1)},
			ReturnValue: &ast.Identifier{Token: lexer.Token{Type: lexer.IDENT, Literal: "x", Pos: pos(5, 9)}, Value: "x", Type: intType},
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
		name  string
		build func(chunk *Chunk)
		want  Value
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
		name  string
		build func(chunk *Chunk)
		want  Value
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
