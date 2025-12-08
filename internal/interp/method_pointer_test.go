package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// testMethodPointerCase is a helper for testing method pointer cases
func testMethodPointerCase(input string, t *testing.T) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
		return nil, ""
	}

	analyzer := semantic.NewAnalyzer()
	_ = analyzer.Analyze(program)

	var buf bytes.Buffer
	interp := New(&buf)
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}
	val := interp.Eval(program)
	return val, buf.String()
}

// TestMethodPointer_BasicValueReturn tests basic method pointer returning a field value
func TestMethodPointer_BasicValueReturn(t *testing.T) {
	input := `
		type TTest = class
			Value: Integer;
			function GetValue: Integer;
		end;
		function TTest.GetValue: Integer;
		begin
			Result := Value;
		end;
		var obj := TTest.Create;
		obj.Value := 42;
		var fp := @obj.GetValue;
		PrintLn(fp());
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "42") {
		t.Errorf("Expected output containing 42, got %q", output)
	}
}

// TestMethodPointer_SelfBindingPersists tests that Self is correctly bound when the method pointer is called
func TestMethodPointer_SelfBindingPersists(t *testing.T) {
	input := `
		type TCounter = class
			Count: Integer;
			function GetCount: Integer;
			procedure Increment;
		end;
		function TCounter.GetCount: Integer;
		begin
			Result := Count;
		end;
		procedure TCounter.Increment;
		begin
			Count := Count + 1;
		end;
		var obj := TCounter.Create;
		obj.Count := 10;
		var getCountPtr := @obj.GetCount;
		var incrPtr := @obj.Increment;
		PrintLn(getCountPtr());  // Should print 10
		incrPtr();
		PrintLn(getCountPtr());  // Should print 11
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "10\n11") {
		t.Errorf("Expected output containing 10 then 11, got %q", output)
	}
}

// TestMethodPointer_WithParameters tests method pointers with parameters
func TestMethodPointer_WithParameters(t *testing.T) {
	input := `
		type TCalc = class
			Base: Integer;
			function Add(x: Integer): Integer;
		end;
		function TCalc.Add(x: Integer): Integer;
		begin
			Result := Base + x;
		end;
		var calc := TCalc.Create;
		calc.Base := 100;
		var addPtr := @calc.Add;
		PrintLn(addPtr(50));  // Should print 150
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "150") {
		t.Errorf("Expected output containing 150, got %q", output)
	}
}

// TestMethodPointer_MultipleParameters tests method pointers with multiple parameters
func TestMethodPointer_MultipleParameters(t *testing.T) {
	input := `
		type TMath = class
			Multiplier: Integer;
			function Calculate(a, b: Integer): Integer;
		end;
		function TMath.Calculate(a, b: Integer): Integer;
		begin
			Result := (a + b) * Multiplier;
		end;
		var math := TMath.Create;
		math.Multiplier := 3;
		var calcPtr := @math.Calculate;
		PrintLn(calcPtr(5, 7));  // Should print (5+7)*3 = 36
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "36") {
		t.Errorf("Expected output containing 36, got %q", output)
	}
}

// TestMethodPointer_StringReturn tests method pointers returning strings
func TestMethodPointer_StringReturn(t *testing.T) {
	input := `
		type TPerson = class
			Name: String;
			Age: Integer;
			function GetInfo: String;
		end;
		function TPerson.GetInfo: String;
		begin
			Result := Name + ' is ' + IntToStr(Age) + ' years old';
		end;
		var person := TPerson.Create;
		person.Name := 'Alice';
		person.Age := 30;
		var infoPtr := @person.GetInfo;
		PrintLn(infoPtr());
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "Alice is 30 years old") {
		t.Errorf("Expected output containing 'Alice is 30 years old', got %q", output)
	}
}

// TestMethodPointer_FloatReturn tests method pointers returning floats
func TestMethodPointer_FloatReturn(t *testing.T) {
	input := `
		type TCircle = class
			Radius: Float;
			function Area: Float;
		end;
		function TCircle.Area: Float;
		begin
			Result := 3.14159 * Radius * Radius;
		end;
		var circle := TCircle.Create;
		circle.Radius := 2.0;
		var areaPtr := @circle.Area;
		PrintLn(areaPtr());  // Should print ~12.56636
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "12.5663") {
		t.Errorf("Expected output containing 12.5663..., got %q", output)
	}
}

// TestMethodPointer_BooleanReturn tests method pointers returning booleans
func TestMethodPointer_BooleanReturn(t *testing.T) {
	input := `
		type TValidator = class
			MinValue: Integer;
			function IsValid(x: Integer): Boolean;
		end;
		function TValidator.IsValid(x: Integer): Boolean;
		begin
			Result := x >= MinValue;
		end;
		var validator := TValidator.Create;
		validator.MinValue := 10;
		var validPtr := @validator.IsValid;
		PrintLn(validPtr(5));   // False
		PrintLn(validPtr(15));  // True
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "False\nTrue") {
		t.Errorf("Expected output containing False then True, got %q", output)
	}
}

// TestMethodPointer_DifferentObjectsSamePointer tests that method pointers capture the object instance
func TestMethodPointer_DifferentObjectsSameMethod(t *testing.T) {
	input := `
		type TContainer = class
			Value: Integer;
			function GetValue: Integer;
		end;
		function TContainer.GetValue: Integer;
		begin
			Result := Value;
		end;
		var obj1 := TContainer.Create;
		var obj2 := TContainer.Create;
		obj1.Value := 100;
		obj2.Value := 200;
		var fp1 := @obj1.GetValue;
		var fp2 := @obj2.GetValue;
		PrintLn(fp1());  // Should print 100
		PrintLn(fp2());  // Should print 200
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "100\n200") {
		t.Errorf("Expected output containing 100 then 200, got %q", output)
	}
}

// TestMethodPointer_ModifyObjectAfterPointerCreation tests that method pointer sees object changes
func TestMethodPointer_ModifyObjectAfterPointerCreation(t *testing.T) {
	input := `
		type THolder = class
			Data: Integer;
			function GetData: Integer;
		end;
		function THolder.GetData: Integer;
		begin
			Result := Data;
		end;
		var holder := THolder.Create;
		holder.Data := 1;
		var fp := @holder.GetData;
		PrintLn(fp());  // Should print 1
		holder.Data := 999;
		PrintLn(fp());  // Should print 999 (sees the change)
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "1\n999") {
		t.Errorf("Expected output containing 1 then 999, got %q", output)
	}
}

// TestMethodPointer_PassedAsParameter tests method pointers passed to functions
func TestMethodPointer_PassedAsParameter(t *testing.T) {
	input := `
		type TProvider = class
			Value: Integer;
			function GetValue: Integer;
		end;
		function TProvider.GetValue: Integer;
		begin
			Result := Value;
		end;
		type TIntFunc = function: Integer;
		function ApplyAndDouble(fn: TIntFunc): Integer;
		begin
			Result := fn() * 2;
		end;
		var provider := TProvider.Create;
		provider.Value := 25;
		var getPtr := @provider.GetValue;
		PrintLn(ApplyAndDouble(getPtr));  // Should print 50
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "50") {
		t.Errorf("Expected output containing 50, got %q", output)
	}
}

// TestMethodPointer_InheritedMethod tests method pointers with inherited methods
func TestMethodPointer_InheritedMethod(t *testing.T) {
	input := `
		type TBase = class
			Value: Integer;
			function GetValue: Integer;
		end;
		function TBase.GetValue: Integer;
		begin
			Result := Value;
		end;
		type TChild = class(TBase)
			Extra: Integer;
		end;
		var child := TChild.Create;
		child.Value := 77;
		child.Extra := 88;
		var fp := @child.GetValue;  // Inherited method
		PrintLn(fp());  // Should print 77
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "77") {
		t.Errorf("Expected output containing 77, got %q", output)
	}
}

// TestMethodPointer_VirtualMethod tests method pointers with virtual methods
func TestMethodPointer_VirtualMethod(t *testing.T) {
	input := `
		type TBase = class
			function Describe: String; virtual;
		end;
		function TBase.Describe: String;
		begin
			Result := 'Base';
		end;
		type TChild = class(TBase)
			function Describe: String; override;
		end;
		function TChild.Describe: String;
		begin
			Result := 'Child';
		end;
		var obj: TBase := TChild.Create;
		var fp := @obj.Describe;
		PrintLn(fp());  // Should print 'Child' due to virtual dispatch
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "Child") {
		t.Errorf("Expected output containing 'Child', got %q", output)
	}
}

// TestMethodPointer_ProcedurePointer tests procedure (no return value) method pointers
func TestMethodPointer_ProcedurePointer(t *testing.T) {
	input := `
		type TAccumulator = class
			Total: Integer;
			procedure Add(x: Integer);
		end;
		procedure TAccumulator.Add(x: Integer);
		begin
			Total := Total + x;
		end;
		var acc := TAccumulator.Create;
		acc.Total := 0;
		var addPtr := @acc.Add;
		addPtr(10);
		addPtr(20);
		addPtr(5);
		PrintLn(acc.Total);  // Should print 35
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "35") {
		t.Errorf("Expected output containing 35, got %q", output)
	}
}

// TestMethodPointer_StoredInArray tests method pointers stored in arrays
// NOTE: This test is skipped because the semantic analyzer does not yet support
// method pointers in array contexts (error: "method pointers (@TClass.Method) not yet implemented").
// The runtime correctly handles method pointers, but the semantic analyzer blocks this usage pattern.
func TestMethodPointer_StoredInArray(t *testing.T) {
	t.Skip("Semantic analyzer does not yet support method pointers in array contexts")
	input := `
		type TValue = class
			Val: Integer;
			function Get: Integer;
		end;
		function TValue.Get: Integer;
		begin
			Result := Val;
		end;
		type TIntFunc = function: Integer;
		var obj1 := TValue.Create;
		var obj2 := TValue.Create;
		var obj3 := TValue.Create;
		obj1.Val := 10;
		obj2.Val := 20;
		obj3.Val := 30;
		var funcs: array of TIntFunc;
		funcs.Add(@obj1.Get);
		funcs.Add(@obj2.Get);
		funcs.Add(@obj3.Get);
		var sum := 0;
		for var i := 0 to High(funcs) do
			sum := sum + funcs[i]();
		PrintLn(sum);  // Should print 60
	`
	_, output := testMethodPointerCase(input, t)
	if !strings.Contains(output, "60") {
		t.Errorf("Expected output containing 60, got %q", output)
	}
}

// TestMethodPointer_ReturnedFromFunction tests method pointers returned from functions
func TestMethodPointer_ReturnedFromFunction(t *testing.T) {
	input := `
		type TGetter = class
			Value: Integer;
			function Get: Integer;
		end;
		function TGetter.Get: Integer;
		begin
			Result := Value;
		end;
		type TIntFunc = function: Integer;
		function CreateGetter(val: Integer): TIntFunc;
		var obj: TGetter;
		begin
			obj := TGetter.Create;
			obj.Value := val;
			Result := @obj.Get;
		end;
		PrintLn('Before CreateGetter');
		var fp := CreateGetter(123);
		PrintLn('After CreateGetter');
		PrintLn(fp());  // Should print 123
	`
	val, output := testMethodPointerCase(input, t)
	t.Logf("Return value: %v", val)
	t.Logf("Output: %q", output)
	if !strings.Contains(output, "123") {
		t.Errorf("Expected output containing 123, got %q", output)
	}
}

// TestSimpleLambdaReturn tests that returning a lambda from a function works
func TestSimpleLambdaReturn(t *testing.T) {
	input := `
		type TIntFunc = function: Integer;
		function MakeLambda(val: Integer): TIntFunc;
		begin
			Result := lambda(): Integer begin Result := val; end;
		end;
		var fp := MakeLambda(42);
		PrintLn(fp());  // Should print 42
	`
	val, output := testMethodPointerCase(input, t)
	t.Logf("Return value: %v", val)
	t.Logf("Output: %q", output)
	if !strings.Contains(output, "42") {
		t.Errorf("Expected output containing 42, got %q", output)
	}
}
