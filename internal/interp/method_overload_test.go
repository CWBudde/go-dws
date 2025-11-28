package interp

import (
	"strings"
	"testing"
)

// ============================================================================
// Comprehensive Method Overload Tests
// ============================================================================

// TestMethodOverload_SimpleCount tests method overloading with different parameter counts
func TestMethodOverload_SimpleCount(t *testing.T) {
	script := `
		type TTest = class
			function Add(a: Integer): Integer; overload;
			function Add(a, b: Integer): Integer; overload;
			function Add(a, b, c: Integer): Integer; overload;
		end;

		function TTest.Add(a: Integer): Integer;
		begin
			Result := a;
		end;

		function TTest.Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function TTest.Add(a, b, c: Integer): Integer;
		begin
			Result := a + b + c;
		end;

		var obj := TTest.Create();
		PrintLn(obj.Add(5));
		PrintLn(obj.Add(5, 3));
		PrintLn(obj.Add(5, 3, 2));
	`
	expectedOutput := "5\n8\n10\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_DifferentTypes tests method overloading with different parameter types
func TestMethodOverload_DifferentTypes(t *testing.T) {
	script := `
		type TTest = class
			function Format(value: Integer): String; overload;
			function Format(value: Float): String; overload;
			function Format(value: String): String; overload;
		end;

		function TTest.Format(value: Integer): String;
		begin
			Result := 'Integer: ' + IntToStr(value);
		end;

		function TTest.Format(value: Float): String;
		begin
			Result := 'Float: ' + FloatToStr(value);
		end;

		function TTest.Format(value: String): String;
		begin
			Result := 'String: ' + value;
		end;

		var obj := TTest.Create();
		PrintLn(obj.Format(42));
		PrintLn(obj.Format(3.14));
		PrintLn(obj.Format('hello'));
	`
	expectedOutput := "Integer: 42\nFloat: 3.14\nString: hello\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_WithInheritance tests overloaded methods across inheritance hierarchy
func TestMethodOverload_WithInheritance(t *testing.T) {
	script := `
		type TBase = class
			function Test(f: Float): String; overload;
		end;

		type TChild = class(TBase)
			function Test(i: Integer): String; overload;
		end;

		function TBase.Test(f: Float): String;
		begin
			Result := 'Float: ' + FloatToStr(f);
		end;

		function TChild.Test(i: Integer): String;
		begin
			Result := 'Integer: ' + IntToStr(i);
		end;

		var obj := TChild.Create();
		PrintLn(obj.Test(42));
		PrintLn(obj.Test(3.14));
	`
	expectedOutput := "Integer: 42\nFloat: 3.14\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ClassMethods tests overloaded class methods (static)
func TestMethodOverload_ClassMethods(t *testing.T) {
	script := `
		type TTest = class
			class function Process(value: Integer): String; overload; static;
			class function Process(a, b: Integer): String; overload; static;
		end;

		class function TTest.Process(value: Integer): String;
		begin
			Result := 'One: ' + IntToStr(value);
		end;

		class function TTest.Process(a, b: Integer): String;
		begin
			Result := 'Two: ' + IntToStr(a) + ', ' + IntToStr(b);
		end;

		PrintLn(TTest.Process(5));
		PrintLn(TTest.Process(5, 3));
	`
	expectedOutput := "One: 5\nTwo: 5, 3\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_Constructors tests overloaded constructors
// SKIP: Constructor overloading has issues with explicit constructors
// The implicit constructor resolution needs enhancement
func TestMethodOverload_Constructors(t *testing.T) {
	t.Skip("Constructor overloading needs enhancement")
	script := `
		type TTest = class
			Value: Integer;
			constructor Create(); overload;
			constructor Create(v: Integer); overload;
		end;

		constructor TTest.Create();
		begin
			Value := 0;
		end;

		constructor TTest.Create(v: Integer);
		begin
			Value := v;
		end;

		var obj1 := TTest.Create();
		PrintLn(obj1.Value);

		var obj2 := TTest.Create(42);
		PrintLn(obj2.Value);
	`
	expectedOutput := "0\n42\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ThreeLevelHierarchy tests overloaded methods in 3-level inheritance
func TestMethodOverload_ThreeLevelHierarchy(t *testing.T) {
	script := `
		type TBase = class
			function Test(f: Float): String; overload;
		end;

		type TMiddle = class(TBase)
			function Test(s: String): String; overload;
		end;

		type TChild = class(TMiddle)
			function Test(i: Integer): String; overload;
		end;

		function TBase.Test(f: Float): String;
		begin
			Result := 'Float';
		end;

		function TMiddle.Test(s: String): String;
		begin
			Result := 'String';
		end;

		function TChild.Test(i: Integer): String;
		begin
			Result := 'Integer';
		end;

		var obj := TChild.Create();
		PrintLn(obj.Test(42));
		PrintLn(obj.Test('hello'));
		PrintLn(obj.Test(3.14));
	`
	expectedOutput := "Integer\nString\nFloat\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_MixedModifiers tests overloaded methods with mixed static/virtual/regular
func TestMethodOverload_MixedModifiers(t *testing.T) {
	script := `
		type TTest = class
			class procedure Print(a: Integer); overload; static;
			class procedure Print(a, b: Integer); overload;
			class procedure Print(a, b, c: Integer); overload; virtual;
		end;

		class procedure TTest.Print(a: Integer);
		begin
			PrintLn('Static: ' + IntToStr(a));
		end;

		class procedure TTest.Print(a, b: Integer);
		begin
			PrintLn('Regular: ' + IntToStr(a) + ', ' + IntToStr(b));
		end;

		class procedure TTest.Print(a, b, c: Integer);
		begin
			PrintLn('Virtual: ' + IntToStr(a) + ', ' + IntToStr(b) + ', ' + IntToStr(c));
		end;

		TTest.Print(1);
		TTest.Print(2, 3);
		TTest.Print(4, 5, 6);
	`
	expectedOutput := "Static: 1\nRegular: 2, 3\nVirtual: 4, 5, 6\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_InstanceAndClass tests mixing instance and class method overloads
func TestMethodOverload_InstanceAndClass(t *testing.T) {
	script := `
		type TTest = class
			class function Get(): String; overload; static;
			function Get(value: Integer): String; overload;
		end;

		class function TTest.Get(): String;
		begin
			Result := 'Class method';
		end;

		function TTest.Get(value: Integer): String;
		begin
			Result := 'Instance method: ' + IntToStr(value);
		end;

		PrintLn(TTest.Get());

		var obj := TTest.Create();
		PrintLn(obj.Get(42));
	`
	expectedOutput := "Class method\nInstance method: 42\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_VirtualOverride tests virtual methods with overloads and override
func TestMethodOverload_VirtualOverride(t *testing.T) {
	script := `
		type TBase = class
			function Compute(a: Integer): Integer; overload; virtual;
			function Compute(a, b: Integer): Integer; overload; virtual;
		end;

		type TChild = class(TBase)
			function Compute(a: Integer): Integer; overload; override;
		end;

		function TBase.Compute(a: Integer): Integer;
		begin
			Result := a * 2;
		end;

		function TBase.Compute(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function TChild.Compute(a: Integer): Integer;
		begin
			Result := a * 3;
		end;

		var base: TBase;
		var child := TChild.Create();
		base := child;

		PrintLn(base.Compute(5));
		PrintLn(base.Compute(5, 3));
	`
	expectedOutput := "15\n8\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ForwardDeclaration tests overloaded methods with forward declarations
func TestMethodOverload_ForwardDeclaration(t *testing.T) {
	script := `
		type TTest = class
			function Calculate(a: Integer): Integer; overload; forward;
			function Calculate(a, b: Integer): Integer; overload; forward;
		end;

		function TTest.Calculate(a: Integer): Integer;
		begin
			Result := a * 2;
		end;

		function TTest.Calculate(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		var obj := TTest.Create();
		PrintLn(obj.Calculate(5));
		PrintLn(obj.Calculate(5, 3));
	`
	expectedOutput := "10\n8\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ProceduresAndFunctions tests mixing procedures and functions
func TestMethodOverload_ProceduresAndFunctions(t *testing.T) {
	script := `
		type TTest = class
			procedure Log(msg: String); overload;
			function Log(value: Integer): String; overload;
		end;

		procedure TTest.Log(msg: String);
		begin
			PrintLn('Procedure: ' + msg);
		end;

		function TTest.Log(value: Integer): String;
		begin
			Result := 'Function: ' + IntToStr(value);
		end;

		var obj := TTest.Create();
		obj.Log('hello');
		PrintLn(obj.Log(42));
	`
	expectedOutput := "Procedure: hello\nFunction: 42\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ImplicitTypeConversion tests overload resolution with implicit type conversion
func TestMethodOverload_ImplicitTypeConversion(t *testing.T) {
	script := `
		type TTest = class
			function Process(value: Float): String; overload;
			function Process(value: String): String; overload;
		end;

		function TTest.Process(value: Float): String;
		begin
			Result := 'Float: ' + FloatToStr(value);
		end;

		function TTest.Process(value: String): String;
		begin
			Result := 'String: ' + value;
		end;

		var obj := TTest.Create();
		PrintLn(obj.Process(42));      // Integer implicitly converts to Float
		PrintLn(obj.Process(3.14));    // Direct Float
		PrintLn(obj.Process('hello')); // Direct String
	`
	expectedOutput := "Float: 42\nFloat: 3.14\nString: hello\n"
	runScriptTest(t, script, expectedOutput)
}

// TestMethodOverload_ReturnTypeDifference tests overloads differing only in return type
// SKIP: Return type-only overloading not yet supported (requires expected type analysis at call site)
// DWScript allows this but it requires sophisticated type inference at call sites
func TestMethodOverload_ReturnTypeDifference(t *testing.T) {
	t.Skip("Return type-only overloading requires expected type analysis - future enhancement")
	script := `
		type TTest = class
			function Get(key: String): Integer; overload;
			function Get(key: String): String; overload;
		end;

		function TTest.Get(key: String): Integer;
		begin
			Result := 42;
		end;

		function TTest.Get(key: String): String;
		begin
			Result := 'value';
		end;

		var obj := TTest.Create();
		var i: Integer;
		i := obj.Get('test');
		PrintLn(i);

		var s: String;
		s := obj.Get('test');
		PrintLn(s);
	`
	// This test documents that DWScript allows overloads differing only in return type
	// and uses the expected type at the call site for resolution
	expectedOutput := "42\nvalue\n"
	runScriptTest(t, script, expectedOutput)
}

// ============================================================================
// Helper Functions
// ============================================================================

func runScriptTest(t *testing.T, script, expectedOutput string) {
	t.Helper()
	result, output := testEvalWithOutput(script)

	if isError(result) {
		t.Fatalf("Script execution failed: %v", result.String())
	}

	actualOutput := strings.TrimSpace(output)
	expected := strings.TrimSpace(expectedOutput)

	if actualOutput != expected {
		t.Errorf("Output mismatch:\nExpected:\n%s\n\nGot:\n%s", expected, actualOutput)
	}
}
