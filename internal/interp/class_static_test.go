package interp

import (
	"bytes"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// Helper function to evaluate class static tests with output capture
func testEvalClassStatic(input string) (Value, string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		panic("Parser errors: " + joinParserErrorsNewline(p.Errors()))
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)
	return result, buf.String()
}

// ============================================================================
// Class Variables (Static Fields) Tests
// ============================================================================

func TestClassVariable(t *testing.T) {
	input := `
	type TCounter = class
		class var Count: Integer;
	end;

	begin
		TCounter.Count := 5;
		PrintLn(TCounter.Count);
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "5\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestClassVariableSharedAcrossInstances(t *testing.T) {
	input := `
	type TCounter = class
		class var SharedCount: Integer;
		InstanceID: Integer;

		function Create(id: Integer): TCounter;
		begin
			InstanceID := id;
			SharedCount := SharedCount + 1;
			Result := Self;
		end;
	end;

	var c1: TCounter;
	var c2: TCounter;
	begin
		TCounter.SharedCount := 0;
		c1 := TCounter.Create(1);
		c2 := TCounter.Create(2);
		PrintLn(TCounter.SharedCount);
		PrintLn(c1.InstanceID);
		PrintLn(c2.InstanceID);
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "2\n1\n2\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestClassVariableAccessFromInstance(t *testing.T) {
	input := `
	type TExample = class
		class var ClassLevel: Integer;
		InstanceLevel: Integer;

		procedure ShowValues;
		begin
			PrintLn(ClassLevel);
			PrintLn(InstanceLevel);
		end;
	end;

	var obj: TExample;
	begin
		TExample.ClassLevel := 42;
		obj := TExample.Create();
		obj.InstanceLevel := 100;
		obj.ShowValues();
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n100\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestMultipleClassVariables(t *testing.T) {
	input := `
	type TConfig = class
		class var ServerName: String;
		class var Port: Integer;
		class var Enabled: Boolean;
	end;

	begin
		TConfig.ServerName := 'localhost';
		TConfig.Port := 8080;
		TConfig.Enabled := true;

		PrintLn(TConfig.ServerName);
		PrintLn(TConfig.Port);
		PrintLn(TConfig.Enabled);
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "localhost\n8080\ntrue\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

// ============================================================================
// Class Methods (Static Methods) Tests
// ============================================================================

func TestClassMethod(t *testing.T) {
	input := `
	type TMath = class
		class function Add(a, b: Integer): Integer; static;
		begin
			Result := a + b;
		end;
	end;

	begin
		PrintLn(TMath.Add(3, 5));
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "8\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestClassMethodAccessingClassVariable(t *testing.T) {
	input := `
	type TCounter = class
		class var Count: Integer;

		class procedure Increment; static;
		begin
			Count := Count + 1;
		end;

		class function GetCount: Integer; static;
		begin
			Result := Count;
		end;
	end;

	begin
		TCounter.Count := 0;
		TCounter.Increment();
		TCounter.Increment();
		TCounter.Increment();
		PrintLn(TCounter.GetCount());
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "3\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestClassMethodWithoutStatic(t *testing.T) {
	input := `
	type TUtils = class
		class function Double(x: Integer): Integer;
		begin
			Result := x * 2;
		end;
	end;

	begin
		PrintLn(TUtils.Double(21));
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "42\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestClassProcedureWithSideEffects(t *testing.T) {
	input := `
	type TLogger = class
		class var LogCount: Integer;

		class procedure Log(msg: String); static;
		begin
			PrintLn(msg);
			LogCount := LogCount + 1;
		end;
	end;

	begin
		TLogger.LogCount := 0;
		TLogger.Log('First message');
		TLogger.Log('Second message');
		PrintLn(TLogger.LogCount);
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "First message\nSecond message\n2\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}

func TestMixedClassAndInstanceMembers(t *testing.T) {
	input := `
	type TExample = class
		class var InstanceCount: Integer;
		Name: String;

		class function GetInstanceCount: Integer; static;
		begin
			Result := InstanceCount;
		end;

		function Create(n: String): TExample;
		begin
			Name := n;
			InstanceCount := InstanceCount + 1;
			Result := Self;
		end;

		function GetName: String;
		begin
			Result := Name;
		end;
	end;

	var obj1: TExample;
	var obj2: TExample;
	begin
		TExample.InstanceCount := 0;
		obj1 := TExample.Create('Alice');
		obj2 := TExample.Create('Bob');

		PrintLn(obj1.GetName());
		PrintLn(obj2.GetName());
		PrintLn(TExample.GetInstanceCount());
	end;
	`

	result, output := testEvalClassStatic(input)
	if isError(result) {
		t.Fatalf("interpreter error: %s", result.String())
	}

	expected := "Alice\nBob\n2\n"
	if output != expected {
		t.Errorf("wrong output. expected=%q, got=%q", expected, output)
	}
}
