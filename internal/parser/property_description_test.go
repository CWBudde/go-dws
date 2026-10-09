package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParse_PropertyDescriptionMetadata(t *testing.T) {
	tests := []struct {
		name, directive, value string
		present                bool
	}{
		{"absent", "", "", false},
		{"literal", " description 'text'", "text", true},
		{"empty", " description ''", "", true},
		{"escaped", " description 'it''s \"quoted\"'", "it's \"quoted\"", true},
		{"comments", " DeScRiPtIoN {comment}\n 'text'", "text", true},
		{"double quoted", " description \"a\"\"quote\"", "a\"quote", true},
		{"multiline", " description \"first\nsecond\"", "first\nsecond", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New("type T = class F: Integer; property P: Integer read F write F" + tt.directive + " reintroduce; end;"))
			program := p.ParseProgram()
			if errors := p.Errors(); len(errors) != 0 {
				t.Fatalf("parse: %v", errors)
			}
			class, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok || len(class.Properties) != 1 {
				t.Fatalf("missing class property: %v", program)
			}
			prop := class.Properties[0]
			if prop.Description != tt.value || prop.HasDescription != tt.present {
				t.Fatalf("description %q, present %v; want %q, %v", prop.Description, prop.HasDescription, tt.value, tt.present)
			}
			if prop.ReadSpec.String() != "F" || prop.WriteSpec.String() != "F" || !prop.IsReintroduce {
				t.Fatalf("lost directives: %s", prop)
			}
		})
	}
}

func TestParse_PropertyDescriptionMissingSemicolonRetainsAST(t *testing.T) {
	p := New(lexer.New("type T = class F: Integer; property P: Integer read F description 'text' end;"))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 1 || errors[0].Stop {
		t.Fatalf("want one ordinary semicolon error, got %v", errors)
	}
	class, ok := program.Statements[0].(*ast.ClassDecl)
	if !ok || len(class.Properties) != 1 {
		t.Fatalf("lost property after recoverable error: %v", program)
	}
	prop := class.Properties[0]
	if prop.Name.Value != "P" || prop.Description != "text" || !prop.HasDescription || prop.ReadSpec.String() != "F" {
		t.Fatalf("lost reached property metadata: %s", prop)
	}
}

func TestParse_PropertyDescriptionRootEndAfterStop(t *testing.T) {
	p := New(lexer.New("var X := ; end;"))
	_ = p.ParseProgram()
	errors := p.Errors()
	if len(errors) != 1 || errors[0].Message != "Expression expected" || !errors[0].Stop {
		t.Fatalf("want only the earlier expression stop, got %v", errors)
	}
}

// Recovery belongs to the ordinary class and root readers. A number stops
// ReadNameList; a semicolon lets the next member survive; root END is ordinary.
func TestParse_PropertyDescriptionRecoveryBoundaries(t *testing.T) {
	const prefix = "type T = class F: Integer; property P: Integer read F description"
	tests := []struct {
		name, source string
		want         []string
		stopped      bool
		properties   int
	}{
		{"semicolon preserves later property", prefix + "; property Q: Integer read F; end;", []string{"String expected"}, false, 2},
		{"number stops before later property", prefix + " 42; property Q: Integer read F; end;", []string{"String expected", `";" expected`, "Name expected"}, true, 1},
		{"end preserves property", prefix + " end; end;", []string{"String expected", `";" expected`, "Unexpected END"}, false, 1},
		{"EOF hot position", prefix, []string{"String expected", `";" expected`, "Name expected"}, true, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(lexer.New(tt.source))
			program := p.ParseProgram()
			errors := p.Errors()
			if len(errors) != len(tt.want) {
				t.Fatalf("errors %v; want %v", errors, tt.want)
			}
			for i, want := range tt.want {
				if errors[i].Message != want {
					t.Fatalf("error %d: %q; want %q", i, errors[i].Message, want)
				}
			}
			if errors[len(errors)-1].Stop != tt.stopped {
				t.Fatalf("stop=%v; want %v", errors[len(errors)-1].Stop, tt.stopped)
			}
			class, ok := program.Statements[0].(*ast.ClassDecl)
			if !ok || len(class.Properties) != tt.properties {
				t.Fatalf("property recovery: %v", program)
			}
		})
	}
}

func TestParse_PropertyDescriptionAfterReadonlyField(t *testing.T) {
	p := New(lexer.New("type T = class F: Integer := 1; readonly; G: Integer; property P: Integer read F description 'text'; end;"))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse: %v", errors)
	}
	class := program.Statements[0].(*ast.ClassDecl)
	if len(class.Fields) != 2 || class.Fields[0].Name.Value != "F" || class.Fields[1].Name.Value != "G" {
		t.Fatalf("lost fields at readonly boundary: %v", class.Fields)
	}
	if len(class.Properties) != 1 || class.Properties[0].Description != "text" || !class.Properties[0].HasDescription {
		t.Fatalf("lost property at readonly boundary: %v", class.Properties)
	}
}

func TestParse_PropertyDescriptionReadonlyHelperBoundary(t *testing.T) {
	for _, qualifier := range []string{"readonly;", `external "field";`} {
		t.Run(qualifier, func(t *testing.T) {
			p := New(lexer.New("type H = record helper for Integer class var F: Integer; " + qualifier + " end;"))
			_ = p.ParseProgram()
			errors := p.Errors()
			if len(errors) != 1 || errors[0].Message != "END expected" || !errors[0].Stop {
				t.Fatalf("helper class variable cannot own class instance qualifier: %v", errors)
			}
		})
	}
}

func TestParse_PropertyDescriptionRecordMethodCompilerStop(t *testing.T) {
	p := New(lexer.New("type R = record X: Integer; procedure M PrintLn(X); end; end; {$ERROR 'later'}"))
	program := p.ParseProgram()
	errors := p.Errors()
	if len(errors) != 2 || errors[0].Message != `";" expected` || errors[0].Stop || errors[1].Message != "Record fields must be declared before record methods" || !errors[1].Stop {
		t.Fatalf("want retained header error and actual stop, got %v", errors)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("unreached statements appended: %v", program.Statements)
	}
	record := program.Statements[0].(*ast.RecordDecl)
	if len(record.Fields) != 1 || len(record.Methods) != 1 {
		t.Fatalf("lost reached record members: %v", record)
	}
}

func TestParse_PropertyDescriptionInvalidFieldDoesNotOwnQualifier(t *testing.T) {
	for _, qualifier := range []string{"readonly;", `external "field";`} {
		t.Run(qualifier, func(t *testing.T) {
			p := New(lexer.New("type T = class F: Integer " + qualifier + " end;"))
			_ = p.ParseProgram()
			errors := p.Errors()
			if len(errors) != 2 || errors[0].Message != `";" expected` || errors[0].Stop || errors[1].Message != "Name expected" || !errors[1].Stop {
				t.Fatalf("invalid field must leave qualifier for normal class recovery: %v", errors)
			}
		})
	}
}

func TestParse_PropertyDescriptionExternalFieldBoundary(t *testing.T) {
	p := New(lexer.New(`type T = class F: Integer; external "field"; readonly; G: Integer; property P: Integer read F description "text"; end;`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 0 {
		t.Fatalf("parse: %v", errors)
	}
	class := program.Statements[0].(*ast.ClassDecl)
	if len(class.Fields) != 2 || len(class.Properties) != 1 || class.Properties[0].Description != "text" {
		t.Fatalf("lost field/property ownership: %v", class)
	}
}
