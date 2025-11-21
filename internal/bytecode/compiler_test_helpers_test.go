package bytecode

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

const (
	numBuiltinGlobals = 50 // ExceptObject + 21 string/IO builtins + 4 type casts + 2 char ops + 6 math + 9 extended math + 7 advanced math (Pi is a constant, not a global)
	userGlobal0       = numBuiltinGlobals
	userGlobal1       = numBuiltinGlobals + 1
)

var testCompileOptimizeOptions = []OptimizeOption{
	WithOptimizationPass(PassLiteralDiscard, false),
	WithOptimizationPass(PassStackShuffle, false),
	WithOptimizationPass(PassConstPropagation, false),
	WithOptimizationPass(PassDeadCode, false),
}

type expectedInstruction struct {
	Op OpCode
	A  byte
	B  uint16
}

func newTestCompiler(name string) *Compiler {
	return NewCompiler(name, WithCompilerOptimizeOptions(testCompileOptimizeOptions...))
}

func firstFunctionValue(constants []Value) *FunctionObject {
	for _, val := range constants {
		if val.IsFunction() {
			return val.AsFunction()
		}
	}
	return nil
}

func expectInstructions(t *testing.T, chunk *Chunk, expected []expectedInstruction) {
	t.Helper()

	if len(chunk.Code) != len(expected) {
		t.Fatalf("instruction count = %d, want %d", len(chunk.Code), len(expected))
	}

	for i, exp := range expected {
		inst := chunk.Code[i]
		if inst.OpCode() != exp.Op {
			t.Fatalf("instruction %d opcode = %v, want %v", i, inst.OpCode(), exp.Op)
		}
		if inst.A() != exp.A {
			t.Fatalf("instruction %d operand A = %d, want %d", i, inst.A(), exp.A)
		}
		if inst.B() != exp.B {
			t.Fatalf("instruction %d operand B = %d, want %d", i, inst.B(), exp.B)
		}
	}
}

func pos(line, column int) lexer.Position {
	return lexer.Position{Line: line, Column: column}
}

func compileProgram(t *testing.T, program *ast.Program) *Chunk {
	t.Helper()
	compiler := newTestCompiler("test_program")
	chunk, err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return chunk
}

func executeChunk(t testing.TB, chunk *Chunk) Value {
	t.Helper()
	stack := make([]Value, 0, 16)
	globals := make(map[int]Value)
	push := func(v Value) {
		stack = append(stack, v)
	}
	pop := func() Value {
		if len(stack) == 0 {
			t.Fatalf("stack underflow during execution")
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v
	}

	localCount := chunk.LocalCount
	if localCount < 0 {
		t.Fatalf("invalid local count %d", localCount)
	}
	locals := make([]Value, localCount)
	ip := 0

	for ip < len(chunk.Code) {
		inst := chunk.Code[ip]
		ip++

		switch inst.OpCode() {
		case OpLoadConst:
			idx := int(inst.B())
			if idx >= len(chunk.Constants) {
				t.Fatalf("constant index %d out of range", idx)
			}
			push(chunk.Constants[idx])
		case OpLoadConst0:
			if len(chunk.Constants) == 0 {
				t.Fatalf("LOAD_CONST_0 with empty constant pool")
			}
			push(chunk.Constants[0])
		case OpLoadConst1:
			if len(chunk.Constants) < 2 {
				t.Fatalf("LOAD_CONST_1 requires at least two constants")
			}
			push(chunk.Constants[1])
		case OpLoadNil:
			push(NilValue())
		case OpLoadTrue:
			push(BoolValue(true))
		case OpLoadFalse:
			push(BoolValue(false))
		case OpStoreLocal:
			idx := int(inst.B())
			if idx >= len(locals) {
				t.Fatalf("local index %d out of range", idx)
			}
			locals[idx] = pop()
		case OpLoadGlobal:
			idx := int(inst.B())
			if val, ok := globals[idx]; ok {
				push(val)
			} else {
				push(NilValue())
			}
		case OpStoreGlobal:
			idx := int(inst.B())
			globals[idx] = pop()
		case OpLoadLocal:
			idx := int(inst.B())
			if idx >= len(locals) {
				t.Fatalf("local index %d out of range", idx)
			}
			push(locals[idx])
		case OpAddInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() + b.AsInt()))
		case OpSubInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() - b.AsInt()))
		case OpMulInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() * b.AsInt()))
		case OpDivInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() / b.AsInt()))
		case OpModInt:
			b := pop()
			a := pop()
			push(IntValue(a.AsInt() % b.AsInt()))
		case OpAddFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() + b.AsFloat()))
		case OpSubFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() - b.AsFloat()))
		case OpMulFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() * b.AsFloat()))
		case OpDivFloat:
			b := pop()
			a := pop()
			push(FloatValue(a.AsFloat() / b.AsFloat()))
		case OpNegateInt:
			push(IntValue(-pop().AsInt()))
		case OpNegateFloat:
			push(FloatValue(-pop().AsFloat()))
		case OpStringConcat:
			b := pop()
			a := pop()
			push(StringValue(a.AsString() + b.AsString()))
		case OpEqual:
			b := pop()
			a := pop()
			eq, ok := valuesEqualForFold(a, b)
			if !ok {
				eq = valueEqual(a, b)
			}
			push(BoolValue(eq))
		case OpNotEqual:
			b := pop()
			a := pop()
			eq, ok := valuesEqualForFold(a, b)
			if !ok {
				eq = valueEqual(a, b)
			}
			push(BoolValue(!eq))
		case OpGreater:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() > b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() > b.AsInt()))
			}
		case OpGreaterEqual:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() >= b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() >= b.AsInt()))
			}
		case OpLess:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() < b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() < b.AsInt()))
			}
		case OpLessEqual:
			b := pop()
			a := pop()
			if a.Type == ValueFloat || b.Type == ValueFloat {
				push(BoolValue(a.AsFloat() <= b.AsFloat()))
			} else {
				push(BoolValue(a.AsInt() <= b.AsInt()))
			}
		case OpAnd:
			b := pop()
			a := pop()
			push(BoolValue(a.AsBool() && b.AsBool()))
		case OpOr:
			b := pop()
			a := pop()
			push(BoolValue(a.AsBool() || b.AsBool()))
		case OpNot:
			push(BoolValue(!pop().AsBool()))
		case OpPop:
			_ = pop()
		case OpJump:
			ip += int(inst.SignedB())
		case OpJumpIfFalse:
			cond := pop()
			if !cond.AsBool() {
				ip += int(inst.SignedB())
			}
		case OpJumpIfTrue:
			cond := pop()
			if cond.AsBool() {
				ip += int(inst.SignedB())
			}
		case OpLoop:
			ip += int(inst.SignedB())
		case OpReturn:
			if inst.A() != 0 {
				return pop()
			}
			return NilValue()
		case OpHalt:
			if len(stack) > 0 {
				return stack[len(stack)-1]
			}
			return NilValue()
		case OpCallIndirect:
			t.Fatalf("CALL_INDIRECT not supported in test VM")
		default:
			t.Fatalf("unsupported opcode %v", inst.OpCode())
		}
	}

	return NilValue()
}

func valueEqual(a, b Value) bool {
	if a.Type != b.Type {
		if a.IsNumber() && b.IsNumber() {
			return a.AsFloat() == b.AsFloat()
		}
		return false
	}

	switch a.Type {
	case ValueNil:
		return true
	case ValueBool:
		return a.AsBool() == b.AsBool()
	case ValueInt:
		return a.AsInt() == b.AsInt()
	case ValueFloat:
		return a.AsFloat() == b.AsFloat()
	case ValueString:
		return a.AsString() == b.AsString()
	case ValueFunction:
		return a.AsFunction() == b.AsFunction()
	case ValueObject:
		return a.AsObject() == b.AsObject()
	default:
		return false
	}
}
