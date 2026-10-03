package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestGlobalOperatorOverload(t *testing.T) {
	input := `
		function StrPlusInt(s: String; i: Integer): String;
		begin
			Result := s;
		end;

		operator + (String, Integer) : String uses StrPlusInt;

		var result := 'abc' + 1;
	`
	expectNoErrors(t, input)
}

func TestImplicitConversionOperator(t *testing.T) {
	input := `
	function IntToStr(i: Integer): String;
	begin
		Result := 'value';
	end;

		operator implicit (Integer) : String uses IntToStr;

		var s: String;
		s := 123;
	`
	expectNoErrors(t, input)
}

func TestDuplicateGlobalOperator(t *testing.T) {
	input := `
		function StrPlusInt(s: String; i: Integer): String;
		begin
			Result := s;
		end;

		operator + (String, Integer) : String uses StrPlusInt;
		operator + (String, Integer) : String uses StrPlusInt;
	`
	expectError(t, input, "An overload already exists for this operator and types")
}

// Test class operator overload validation
func TestClassOperatorOverload(t *testing.T) {
	input := `
		type TTest = class
			Field: String;
			constructor Create;
			function AddString(str: String): TTest;
			class operator + (TTest, String) : TTest uses AddString;
		end;

		constructor TTest.Create;
		begin
			Field := '';
		end;

		function TTest.AddString(str: String): TTest;
		begin
			Field := Field + str;
			Result := Self;
		end;

		var t: TTest;
		begin
			t := TTest.Create();
			t := t + 'test';
		end
	`
	expectNoErrors(t, input)
}

// Test operator signature mismatch errors
func TestOperatorSignatureMismatch(t *testing.T) {
	// Test 1: Wrong number of parameters
	t.Run("wrong parameter count", func(t *testing.T) {
		input := `
			function WrongParams(s: String): String;
			begin
				Result := s;
			end;

			operator + (String, Integer) : String uses WrongParams;
		`
		expectError(t, input, "Expected 2 parameters (instead of 1)")
	})

	// Test 2: Wrong parameter types
	t.Run("wrong parameter types", func(t *testing.T) {
		input := `
			function WrongTypes(i: Integer; s: String): String;
			begin
				Result := s;
			end;

			operator + (String, Integer) : String uses WrongTypes;
		`
		expectError(t, input, `Parameter 0 - Type "String" expected (instead of "Integer")`)
	})

	// Test 3: Wrong return type (for class operators)
	t.Run("wrong return type class operator", func(t *testing.T) {
		input := `
			type TTest = class
				Field: String;
				function WrongReturn(str: String): Integer;
				class operator + (TTest, String) : TTest uses WrongReturn;
			end;

			function TTest.WrongReturn(str: String): Integer;
			begin
				Result := 0;
			end;
		`
		expectError(t, input, "return type")
	})
}

// Test invalid binding function (not found, wrong signature)
func TestInvalidBindingFunction(t *testing.T) {
	// Test 1: Binding function not found
	t.Run("binding not found", func(t *testing.T) {
		input := `
			operator + (String, Integer) : String uses NonExistentFunction;
		`
		expectError(t, input, "not found")
	})

	// Test 2: Binding is not a function (bound to a variable)
	t.Run("binding not a function", func(t *testing.T) {
		input := `
			var NotAFunction: String;

			operator + (String, Integer) : String uses NotAFunction;
		`
		expectError(t, input, "not a function")
	})

	// Test 3: Class operator binding method not found
	t.Run("class operator method not found", func(t *testing.T) {
		input := `
			type TTest = class
				Field: String;
				class operator + (TTest, String) : TTest uses NonExistentMethod;
			end;
		`
		expectError(t, input, "not found")
	})
}

// Invalid bindings remain declaration entries for duplicate checks, but never
// become executable overloads.
func TestGlobalOperatorRejectedBindingRecovery(t *testing.T) {
	source := "function Bad: String; begin Result := ''; end;\n" +
		"function Good(a, b: Integer): String; begin Result := ''; end;\n" +
		"operator + (Integer, Integer): String uses Bad;\n" +
		"operator + (Integer, Integer): String uses Good;"
	analyzer, err := analyzeSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "An overload already exists for this operator and types") {
		t.Fatalf("errors = %v", err)
	}
	if sig, found := analyzer.globalOperators.Lookup("+", []types.Type{types.INTEGER, types.INTEGER}); found && sig.Binding == "Bad" {
		t.Fatal("rejected binding is executable")
	}
}

func TestGlobalOperatorWrongOperandCountDoesNotReserveSignature(t *testing.T) {
	source := "function F(a, b: Integer): Integer; begin Result := a+b; end;\n" +
		"operator + (Integer): Integer uses F;\n" +
		"operator + (Integer, Integer): Integer uses F;"
	analyzer, err := analyzeSource(t, source)
	if err == nil || strings.Contains(err.Error(), "An overload already exists") {
		t.Fatalf("errors = %v", err)
	}
	if sig, found := analyzer.globalOperators.Lookup("+", []types.Type{types.INTEGER, types.INTEGER}); !found || sig.Binding != "F" {
		t.Fatal("valid later declaration was lost")
	}
}

func TestGlobalOperatorRejectedBindingIsLocalToScope(t *testing.T) {
	analyzer, err := analyzeSource(t, "function Bad: String; begin Result := ''; end; function Good(a, b: Integer): String; begin Result := ''; end;")
	if err != nil {
		t.Fatal(err)
	}
	program := parseProgram(t, "operator + (Integer, Integer): String uses Bad; operator + (Integer, Integer): String uses Good;")
	root := analyzer.symbols
	analyzer.symbols = NewEnclosedSymbolTable(root)
	analyzer.analyzeOperatorDecl(program.Statements[0].(*ast.OperatorDecl))
	count := len(analyzer.Errors())
	if count != 1 {
		t.Fatalf("rejected binding errors = %v", analyzer.Errors())
	}
	analyzer.symbols = NewEnclosedSymbolTable(root)
	analyzer.analyzeOperatorDecl(program.Statements[1].(*ast.OperatorDecl))
	if len(analyzer.Errors()) != count {
		t.Fatalf("unrelated scope inherited duplicate: %v", analyzer.Errors())
	}
	if _, found := analyzer.globalOperators.Lookup("+", []types.Type{types.INTEGER, types.INTEGER}); !found {
		t.Fatal("valid binding was not registered")
	}
}

func TestGlobalOperatorWrongOperandCountDoesNotRegisterHelper(t *testing.T) {
	source := "type H = helper for Integer function F: Integer; begin Result := Self; end; end; operator + (Integer): Integer uses H.F;"
	analyzer, err := analyzeSource(t, source)
	if err == nil {
		t.Fatal("invalid operand count accepted")
	}
	if _, found := analyzer.globalOperators.Lookup("+", []types.Type{types.INTEGER}); found {
		t.Fatal("invalid helper binding became executable")
	}
}

func TestGlobalOperatorRejectedBindingConflictsWithHelper(t *testing.T) {
	source := "type H = helper for Integer function F(v: Integer): String; begin Result := ''; end; end;\n" +
		"function Bad: String; begin Result := ''; end;\n" +
		"operator + (Integer, Integer): String uses Bad;\n" +
		"operator + (Integer, Integer): String uses H.F;"
	analyzer, err := analyzeSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "An overload already exists for this operator and types") {
		t.Fatalf("errors = %v", err)
	}
	if _, found := analyzer.globalOperators.Lookup("+", []types.Type{types.INTEGER, types.INTEGER}); found {
		t.Fatal("duplicate helper was registered")
	}
}

func TestInvalidBuiltinConversionDoesNotRegister(t *testing.T) {
	source := "operator implicit (Integer, Integer): String uses IntToStr; operator implicit (Integer): String uses IntToStr;"
	analyzer, err := analyzeSource(t, source)
	if err == nil || strings.Contains(err.Error(), "already defined") {
		t.Fatalf("errors = %v", err)
	}
	if _, found := analyzer.conversionRegistry.FindImplicit(types.INTEGER, types.STRING); !found {
		t.Fatal("valid later conversion lost")
	}
	if len(analyzer.Errors()) != 1 {
		t.Fatalf("errors = %v", analyzer.Errors())
	}
}

// Upstream TDynamicArraySymbol.DoIsOfType compares element types in reverse
// (typSym.Typ.DoIsOfType(Typ)), so a binding returning an ancestor element
// array satisfies a descendant element array result, but not vice versa.
func TestGlobalOperatorArrayResultMirrorsUpstreamElementDirection(t *testing.T) {
	decls := "type TBase = class end; type TChild = class(TBase) end;\n" +
		"type TBaseArr = array of TBase; type TChildArr = array of TChild;\n" +
		"function MakeBase(a, b: Integer): TBaseArr; begin end;\n" +
		"function MakeChild(a, b: Integer): TChildArr; begin end;\n"

	if _, err := analyzeSource(t, decls+"operator + (Integer, Integer): TChildArr uses MakeBase;"); err != nil {
		t.Fatalf("ancestor element result rejected: %v", err)
	}
	if _, err := analyzeSource(t, decls+"operator + (Integer, Integer): TBaseArr uses MakeChild;"); err == nil ||
		!strings.Contains(err.Error(), "Result type should be") {
		t.Fatalf("descendant element result accepted or wrong error: %v", err)
	}
}
