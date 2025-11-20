package bytecode

import (
	"testing"
)

// TestFoldFloatOp tests constant folding for floating-point operations
func TestFoldFloatOp(t *testing.T) {
	tests := []struct {
		left      Value
		right     Value
		wantValue Value
		name      string
		op        OpCode
		wantOk    bool
	}{
		{
			name:      "add floats",
			op:        OpAddFloat,
			left:      FloatValue(3.5),
			right:     FloatValue(2.5),
			wantValue: FloatValue(6.0),
			wantOk:    true,
		},
		{
			name:      "subtract floats",
			op:        OpSubFloat,
			left:      FloatValue(10.5),
			right:     FloatValue(3.5),
			wantValue: FloatValue(7.0),
			wantOk:    true,
		},
		{
			name:      "multiply floats",
			op:        OpMulFloat,
			left:      FloatValue(2.5),
			right:     FloatValue(4.0),
			wantValue: FloatValue(10.0),
			wantOk:    true,
		},
		{
			name:      "divide floats",
			op:        OpDivFloat,
			left:      FloatValue(10.0),
			right:     FloatValue(2.5),
			wantValue: FloatValue(4.0),
			wantOk:    true,
		},
		{
			name:      "divide by zero",
			op:        OpDivFloat,
			left:      FloatValue(10.0),
			right:     FloatValue(0.0),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "add int and float",
			op:        OpAddFloat,
			left:      IntValue(3),
			right:     FloatValue(2.5),
			wantValue: FloatValue(5.5),
			wantOk:    true,
		},
		{
			name:      "add float and int",
			op:        OpAddFloat,
			left:      FloatValue(3.5),
			right:     IntValue(2),
			wantValue: FloatValue(5.5),
			wantOk:    true,
		},
		{
			name:      "subtract mixed types",
			op:        OpSubFloat,
			left:      IntValue(10),
			right:     FloatValue(3.5),
			wantValue: FloatValue(6.5),
			wantOk:    true,
		},
		{
			name:      "multiply mixed types",
			op:        OpMulFloat,
			left:      FloatValue(2.5),
			right:     IntValue(4),
			wantValue: FloatValue(10.0),
			wantOk:    true,
		},
		{
			name:      "non-number left operand",
			op:        OpAddFloat,
			left:      StringValue("hello"),
			right:     FloatValue(2.5),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "non-number right operand",
			op:        OpAddFloat,
			left:      FloatValue(3.5),
			right:     StringValue("world"),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "negative floats",
			op:        OpAddFloat,
			left:      FloatValue(-3.5),
			right:     FloatValue(-2.5),
			wantValue: FloatValue(-6.0),
			wantOk:    true,
		},
		{
			name:      "unknown opcode",
			op:        OpAddInt,
			left:      FloatValue(3.5),
			right:     FloatValue(2.5),
			wantValue: Value{},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := foldFloatOp(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("foldFloatOp() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk && gotValue.Type != tt.wantValue.Type {
				t.Errorf("foldFloatOp() type = %v, want %v", gotValue.Type, tt.wantValue.Type)
			}
			if tt.wantOk && gotValue.IsFloat() && gotValue.AsFloat() != tt.wantValue.AsFloat() {
				t.Errorf("foldFloatOp() value = %v, want %v", gotValue.AsFloat(), tt.wantValue.AsFloat())
			}
		})
	}
}

// TestFoldEqualityOp tests constant folding for equality operations
func TestFoldEqualityOp(t *testing.T) {
	tests := []struct {
		left      Value
		right     Value
		wantValue Value
		name      string
		op        OpCode
		wantOk    bool
	}{
		{
			name:      "equal ints",
			op:        OpEqual,
			left:      IntValue(42),
			right:     IntValue(42),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "not equal ints",
			op:        OpEqual,
			left:      IntValue(42),
			right:     IntValue(10),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "equal floats",
			op:        OpEqual,
			left:      FloatValue(3.14),
			right:     FloatValue(3.14),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "not equal floats",
			op:        OpEqual,
			left:      FloatValue(3.14),
			right:     FloatValue(2.71),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "equal strings",
			op:        OpEqual,
			left:      StringValue("hello"),
			right:     StringValue("hello"),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "not equal strings",
			op:        OpEqual,
			left:      StringValue("hello"),
			right:     StringValue("world"),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "equal bools",
			op:        OpEqual,
			left:      BoolValue(true),
			right:     BoolValue(true),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "not equal bools",
			op:        OpEqual,
			left:      BoolValue(true),
			right:     BoolValue(false),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "equal nils",
			op:        OpEqual,
			left:      NilValue(),
			right:     NilValue(),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "int equals float with same value",
			op:        OpEqual,
			left:      IntValue(42),
			right:     FloatValue(42.0),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "int not equals float with different value",
			op:        OpEqual,
			left:      IntValue(42),
			right:     FloatValue(3.14),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "not equal different types",
			op:        OpEqual,
			left:      IntValue(42),
			right:     StringValue("42"),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "not equal operator ints",
			op:        OpNotEqual,
			left:      IntValue(42),
			right:     IntValue(10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "not equal operator equal values",
			op:        OpNotEqual,
			left:      IntValue(42),
			right:     IntValue(42),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "not equal operator strings",
			op:        OpNotEqual,
			left:      StringValue("hello"),
			right:     StringValue("world"),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "unknown opcode",
			op:        OpAddInt,
			left:      IntValue(42),
			right:     IntValue(10),
			wantValue: Value{},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := foldEqualityOp(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("foldEqualityOp() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk && gotValue.Type != tt.wantValue.Type {
				t.Errorf("foldEqualityOp() type = %v, want %v", gotValue.Type, tt.wantValue.Type)
			}
			if tt.wantOk && gotValue.IsBool() && gotValue.AsBool() != tt.wantValue.AsBool() {
				t.Errorf("foldEqualityOp() value = %v, want %v", gotValue.AsBool(), tt.wantValue.AsBool())
			}
		})
	}
}

// TestFoldComparisonOp tests constant folding for comparison operations
func TestFoldComparisonOp(t *testing.T) {
	tests := []struct {
		left      Value
		right     Value
		wantValue Value
		name      string
		op        OpCode
		wantOk    bool
	}{
		{
			name:      "greater than true",
			op:        OpGreater,
			left:      IntValue(10),
			right:     IntValue(5),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "greater than false",
			op:        OpGreater,
			left:      IntValue(5),
			right:     IntValue(10),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "greater than equal values",
			op:        OpGreater,
			left:      IntValue(10),
			right:     IntValue(10),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "greater equal true",
			op:        OpGreaterEqual,
			left:      IntValue(10),
			right:     IntValue(10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "greater equal false",
			op:        OpGreaterEqual,
			left:      IntValue(5),
			right:     IntValue(10),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "less than true",
			op:        OpLess,
			left:      IntValue(5),
			right:     IntValue(10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "less than false",
			op:        OpLess,
			left:      IntValue(10),
			right:     IntValue(5),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "less equal true",
			op:        OpLessEqual,
			left:      IntValue(10),
			right:     IntValue(10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "less equal false",
			op:        OpLessEqual,
			left:      IntValue(10),
			right:     IntValue(5),
			wantValue: BoolValue(false),
			wantOk:    true,
		},
		{
			name:      "compare floats greater",
			op:        OpGreater,
			left:      FloatValue(3.14),
			right:     FloatValue(2.71),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "compare floats less",
			op:        OpLess,
			left:      FloatValue(2.71),
			right:     FloatValue(3.14),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "compare mixed int float",
			op:        OpGreater,
			left:      IntValue(10),
			right:     FloatValue(5.5),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "compare mixed float int",
			op:        OpLess,
			left:      FloatValue(3.5),
			right:     IntValue(10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "non-number left operand",
			op:        OpGreater,
			left:      StringValue("hello"),
			right:     IntValue(10),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "non-number right operand",
			op:        OpGreater,
			left:      IntValue(10),
			right:     StringValue("world"),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "negative numbers",
			op:        OpGreater,
			left:      IntValue(-5),
			right:     IntValue(-10),
			wantValue: BoolValue(true),
			wantOk:    true,
		},
		{
			name:      "unknown opcode",
			op:        OpAddInt,
			left:      IntValue(10),
			right:     IntValue(5),
			wantValue: Value{},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := foldComparisonOp(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("foldComparisonOp() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk && gotValue.Type != tt.wantValue.Type {
				t.Errorf("foldComparisonOp() type = %v, want %v", gotValue.Type, tt.wantValue.Type)
			}
			if tt.wantOk && gotValue.IsBool() && gotValue.AsBool() != tt.wantValue.AsBool() {
				t.Errorf("foldComparisonOp() value = %v, want %v", gotValue.AsBool(), tt.wantValue.AsBool())
			}
		})
	}
}

// TestFoldIntegerOp tests constant folding for integer operations
func TestFoldIntegerOp(t *testing.T) {
	tests := []struct {
		left      Value
		right     Value
		wantValue Value
		name      string
		op        OpCode
		wantOk    bool
	}{
		{
			name:      "add integers",
			op:        OpAddInt,
			left:      IntValue(3),
			right:     IntValue(5),
			wantValue: IntValue(8),
			wantOk:    true,
		},
		{
			name:      "subtract integers",
			op:        OpSubInt,
			left:      IntValue(10),
			right:     IntValue(3),
			wantValue: IntValue(7),
			wantOk:    true,
		},
		{
			name:      "multiply integers",
			op:        OpMulInt,
			left:      IntValue(4),
			right:     IntValue(5),
			wantValue: IntValue(20),
			wantOk:    true,
		},
		{
			name:      "divide integers",
			op:        OpDivInt,
			left:      IntValue(20),
			right:     IntValue(4),
			wantValue: IntValue(5),
			wantOk:    true,
		},
		{
			name:      "modulo integers",
			op:        OpModInt,
			left:      IntValue(10),
			right:     IntValue(3),
			wantValue: IntValue(1),
			wantOk:    true,
		},
		{
			name:      "divide by zero",
			op:        OpDivInt,
			left:      IntValue(10),
			right:     IntValue(0),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "modulo by zero",
			op:        OpModInt,
			left:      IntValue(10),
			right:     IntValue(0),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "non-int left operand",
			op:        OpAddInt,
			left:      FloatValue(3.5),
			right:     IntValue(5),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "non-int right operand",
			op:        OpAddInt,
			left:      IntValue(3),
			right:     FloatValue(5.5),
			wantValue: Value{},
			wantOk:    false,
		},
		{
			name:      "negative integers",
			op:        OpAddInt,
			left:      IntValue(-3),
			right:     IntValue(-5),
			wantValue: IntValue(-8),
			wantOk:    true,
		},
		{
			name:      "unknown opcode",
			op:        OpAddFloat,
			left:      IntValue(3),
			right:     IntValue(5),
			wantValue: Value{},
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := foldIntegerOp(tt.op, tt.left, tt.right)
			if gotOk != tt.wantOk {
				t.Errorf("foldIntegerOp() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk && gotValue.Type != tt.wantValue.Type {
				t.Errorf("foldIntegerOp() type = %v, want %v", gotValue.Type, tt.wantValue.Type)
			}
			if tt.wantOk && gotValue.IsInt() && gotValue.AsInt() != tt.wantValue.AsInt() {
				t.Errorf("foldIntegerOp() value = %v, want %v", gotValue.AsInt(), tt.wantValue.AsInt())
			}
		})
	}
}

// TestFoldBinaryOp tests the main foldBinaryOp dispatcher
func TestFoldBinaryOp(t *testing.T) {
	t.Run("unknown values", func(t *testing.T) {
		left := valueState{known: false, value: IntValue(42)}
		right := valueState{known: true, value: IntValue(10)}
		_, ok := foldBinaryOp(OpAddInt, left, right)
		if ok {
			t.Errorf("foldBinaryOp() with unknown left should return false")
		}
	})

	t.Run("unknown right", func(t *testing.T) {
		left := valueState{known: true, value: IntValue(42)}
		right := valueState{known: false, value: IntValue(10)}
		_, ok := foldBinaryOp(OpAddInt, left, right)
		if ok {
			t.Errorf("foldBinaryOp() with unknown right should return false")
		}
	})

	t.Run("known integer addition", func(t *testing.T) {
		left := valueState{known: true, value: IntValue(42)}
		right := valueState{known: true, value: IntValue(10)}
		result, ok := foldBinaryOp(OpAddInt, left, right)
		if !ok {
			t.Errorf("foldBinaryOp() should succeed with known values")
		}
		if result.AsInt() != 52 {
			t.Errorf("foldBinaryOp() = %v, want 52", result.AsInt())
		}
	})

	t.Run("known float multiplication", func(t *testing.T) {
		left := valueState{known: true, value: FloatValue(3.5)}
		right := valueState{known: true, value: FloatValue(2.0)}
		result, ok := foldBinaryOp(OpMulFloat, left, right)
		if !ok {
			t.Errorf("foldBinaryOp() should succeed with known floats")
		}
		if result.AsFloat() != 7.0 {
			t.Errorf("foldBinaryOp() = %v, want 7.0", result.AsFloat())
		}
	})

	t.Run("known equality check", func(t *testing.T) {
		left := valueState{known: true, value: IntValue(42)}
		right := valueState{known: true, value: IntValue(42)}
		result, ok := foldBinaryOp(OpEqual, left, right)
		if !ok {
			t.Errorf("foldBinaryOp() should succeed with equality check")
		}
		if !result.AsBool() {
			t.Errorf("foldBinaryOp() = %v, want true", result.AsBool())
		}
	})

	t.Run("known comparison", func(t *testing.T) {
		left := valueState{known: true, value: IntValue(10)}
		right := valueState{known: true, value: IntValue(5)}
		result, ok := foldBinaryOp(OpGreater, left, right)
		if !ok {
			t.Errorf("foldBinaryOp() should succeed with comparison")
		}
		if !result.AsBool() {
			t.Errorf("foldBinaryOp() = %v, want true", result.AsBool())
		}
	})

	t.Run("unsupported opcode", func(t *testing.T) {
		left := valueState{known: true, value: IntValue(42)}
		right := valueState{known: true, value: IntValue(10)}
		_, ok := foldBinaryOp(OpPop, left, right)
		if ok {
			t.Errorf("foldBinaryOp() with unsupported opcode should return false")
		}
	})
}

// TestValuesEqual tests the valuesEqual helper function
func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    Value
		b    Value
		want bool
	}{
		{"nil equals nil", NilValue(), NilValue(), true},
		{"bool true equals true", BoolValue(true), BoolValue(true), true},
		{"bool false equals false", BoolValue(false), BoolValue(false), true},
		{"bool true not equals false", BoolValue(true), BoolValue(false), false},
		{"int equals int", IntValue(42), IntValue(42), true},
		{"int not equals int", IntValue(42), IntValue(10), false},
		{"float equals float", FloatValue(3.14), FloatValue(3.14), true},
		{"float not equals float", FloatValue(3.14), FloatValue(2.71), false},
		{"string equals string", StringValue("hello"), StringValue("hello"), true},
		{"string not equals string", StringValue("hello"), StringValue("world"), false},
		{"int equals float same value", IntValue(42), FloatValue(42.0), true},
		{"int not equals float different value", IntValue(42), FloatValue(3.14), false},
		{"different types not equal", IntValue(42), StringValue("42"), false},
		{"array not equal", ArrayValue(NewArrayInstance([]Value{})), ArrayValue(NewArrayInstance([]Value{})), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
