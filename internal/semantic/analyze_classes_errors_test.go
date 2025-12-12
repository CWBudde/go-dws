package semantic

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// TestErrorMessageCasing tests that error messages preserve the original
// user-provided casing instead of showing normalized (lowercased) identifiers.
func TestErrorMessageCasing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		errorContains string
		wantErr       bool
	}{
		{
			name: "non-existent member with MixedCase",
			input: `type
	TMyClass = class
	public
		Value: Integer;
	end;

var obj: TMyClass;
begin
	obj.MyField := 42;
end;`,
			wantErr:       true,
			errorContains: "MyField", // Should show 'MyField', not 'myfield'
		},
		{
			name: "non-existent member with lowercase",
			input: `type
	TMyClass = class
	public
		Value: Integer;
	end;

var obj: TMyClass;
begin
	obj.myfield := 42;
end;`,
			wantErr:       true,
			errorContains: "myfield", // Should show 'myfield', not something else
		},
		{
			name: "non-existent member with UPPERCASE",
			input: `type
	TMyClass = class
	public
		Value: Integer;
	end;

var obj: TMyClass;
begin
	obj.MYFIELD := 42;
end;`,
			wantErr:       true,
			errorContains: "MYFIELD", // Should show 'MYFIELD', not 'myfield'
		},
		{
			name: "non-existent field access with camelCase",
			input: `type
	TTest = class
	private
		FData: String;
	public
		function GetData(): String;
	end;

function TTest.GetData(): String;
begin
	Result := FData;
end;

var t: TTest;
var s: String;
begin
	s := t.someMethod();
end;`,
			wantErr:       true,
			errorContains: "someMethod", // Should show 'someMethod', not 'somemethod'
		},
		{
			name: "private field access with PascalCase",
			input: `type
	TMyClass = class
	private
		PrivateField: Integer;
	end;

var obj: TMyClass;
begin
	obj.PrivateField := 42;
end;`,
			wantErr:       true,
			errorContains: "PrivateField", // Should show 'PrivateField' in error
		},
		{
			name: "private method access with MixedCase",
			input: `type
	TMyClass = class
	private
		function PrivateMethod(): Integer;
	end;

function TMyClass.PrivateMethod(): Integer;
begin
	Result := 42;
end;

var obj: TMyClass;
var x: Integer;
begin
	x := obj.PrivateMethod();
end;`,
			wantErr:       true,
			errorContains: "PrivateMethod", // Should show 'PrivateMethod' in error
		},
		{
			name: "non-existent class variable with snake_case style",
			input: `type
	TMyClass = class
	public
		class var Value: Integer;
	end;

var x: Integer;
begin
	x := TMyClass.Non_Existent;
end;`,
			wantErr:       true,
			errorContains: "Non_Existent", // Should show 'Non_Existent', not 'non_existent'
		},
		{
			name: "helper member access with unusual casing",
			input: `var s: String;
var x: Integer;
begin
	x := s.WeIrDcAsE;
end;`,
			wantErr:       true,
			errorContains: "WeIrDcAsE", // Should show 'WeIrDcAsE', not 'weirdcase'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			analyzer := NewAnalyzer()
			err := analyzer.Analyze(program)

			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantErr && err != nil {
				errMsg := err.Error()
				if !strings.Contains(errMsg, tt.errorContains) {
					t.Errorf("Error message should contain original casing '%s'.\nGot error: %s",
						tt.errorContains, errMsg)
				}
			}
		})
	}
}

// TestInterfaceMethodErrorCasing tests that interface method access errors
// preserve the original casing (this was already correct, but we verify it).
func TestInterfaceMethodErrorCasing(t *testing.T) {
	input := `type
	IMyInterface = interface
		function GetValue(): Integer;
	end;

var obj: IMyInterface;
var x: Integer;
begin
	x := obj.MyMethod();
end;`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	err := analyzer.Analyze(program)

	if err == nil {
		t.Fatal("expected error for non-existent interface method")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "MyMethod") {
		t.Errorf("Error message should contain original casing 'MyMethod'.\nGot error: %s", errMsg)
	}
}
