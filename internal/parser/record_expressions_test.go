package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Parser Tests for Anonymous Record Constructor Expressions
//
// Syntax: record name := value; ... end
// Distinct from the parenthesised record literal (x: 10; y: 20) covered by
// record_literals_test.go: this form is structurally typed and stands alone as
// an expression.
// ============================================================================

// parseSingleRecordExpr parses `var r := <input>;` and returns the record
// expression bound to r.
func parseSingleRecordExpr(t *testing.T, input string) *ast.AnonymousRecordExpression {
	t.Helper()

	program := testParse(t, "var r := "+input+";")
	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.VarDeclStatement)
	if !ok {
		t.Fatalf("statement is not *ast.VarDeclStatement, got %T", program.Statements[0])
	}

	recordExpr, ok := stmt.Value.(*ast.AnonymousRecordExpression)
	if !ok {
		t.Fatalf("value is not *ast.AnonymousRecordExpression, got %T", stmt.Value)
	}

	return recordExpr
}

func TestParseAnonymousRecordExpression_FieldNames(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFields []string
	}{
		{
			name:       "single field",
			input:      "record a := 1; end",
			wantFields: []string{"a"},
		},
		{
			name:       "multiple fields",
			input:      "record a := 1; b := 'x'; c := True; end",
			wantFields: []string{"a", "b", "c"},
		},
		{
			name:       "no trailing semicolon",
			input:      "record Field := 123 end",
			wantFields: []string{"Field"},
		},
		{
			name:       "quoted field names are kept verbatim",
			input:      `record "i*i" := 1; "2i" := 2; end`,
			wantFields: []string{"i*i", "2i"},
		},
		{
			name:       "empty record",
			input:      "record end",
			wantFields: []string{},
		},
		{
			name:       "field names preserve casing",
			input:      "record SomeField := 1; end",
			wantFields: []string{"SomeField"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordExpr := parseSingleRecordExpr(t, tt.input)

			if len(recordExpr.Fields) != len(tt.wantFields) {
				t.Fatalf("expected %d fields, got %d", len(tt.wantFields), len(recordExpr.Fields))
			}

			for i, want := range tt.wantFields {
				field := recordExpr.Fields[i]
				if field.Name == nil {
					t.Fatalf("field %d has no name", i)
				}
				if field.Name.Value != want {
					t.Errorf("field %d: expected name %q, got %q", i, want, field.Name.Value)
				}
				if field.Value == nil {
					t.Errorf("field %d (%q) has no value", i, want)
				}
			}
		})
	}
}

func TestParseAnonymousRecordExpression_FieldValuesAreExpressions(t *testing.T) {
	recordExpr := parseSingleRecordExpr(t, "record a := 1 + 2; b := f(3); end")

	if len(recordExpr.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(recordExpr.Fields))
	}

	if _, ok := recordExpr.Fields[0].Value.(*ast.BinaryExpression); !ok {
		t.Errorf("field 'a' value should be *ast.BinaryExpression, got %T", recordExpr.Fields[0].Value)
	}
	if _, ok := recordExpr.Fields[1].Value.(*ast.CallExpression); !ok {
		t.Errorf("field 'b' value should be *ast.CallExpression, got %T", recordExpr.Fields[1].Value)
	}
}

func TestParseAnonymousRecordExpression_Nested(t *testing.T) {
	recordExpr := parseSingleRecordExpr(t, "record outer := record inner := 1; end; end")

	if len(recordExpr.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(recordExpr.Fields))
	}

	inner, ok := recordExpr.Fields[0].Value.(*ast.AnonymousRecordExpression)
	if !ok {
		t.Fatalf("nested value should be *ast.AnonymousRecordExpression, got %T", recordExpr.Fields[0].Value)
	}
	if len(inner.Fields) != 1 || inner.Fields[0].Name.Value != "inner" {
		t.Errorf("unexpected nested record fields: %v", inner.Fields)
	}
}

// A record expression must also parse in argument position, which is how the
// JSONConnector fixtures use it: JSON.Stringify(record a := 1; end).
func TestParseAnonymousRecordExpression_AsArgument(t *testing.T) {
	program := testParse(t, "PrintLn(JSON.Stringify(record a := 1; end));")

	if len(program.Statements) != 1 {
		t.Fatalf("program should have 1 statement, got %d", len(program.Statements))
	}
}

func TestParseAnonymousRecordExpression_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing end", input: "var r := record a := 1;"},
		{name: "missing assign", input: "var r := record a 1; end;"},
		{name: "missing field name", input: "var r := record := 1; end;"},
		{name: "colon instead of assign", input: "var r := record a: 1; end;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := testParseWithErrors(t, tt.input)
			if len(errors) == 0 {
				t.Errorf("expected parser errors for %q, got none", tt.input)
			}
		})
	}
}

// Registering a prefix parse function for RECORD must not disturb inline
// anonymous record *types*, which are parsed in type position.
func TestParseAnonymousRecordExpression_DoesNotBreakInlineRecordTypes(t *testing.T) {
	inputs := []string{
		"const a : array [0..1] of record x : Integer; end = [];",
		"type TR = record x : Integer; end;",
		"var v : record x : Integer; end;",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, errors := testParseWithErrors(t, input); len(errors) > 0 {
				t.Errorf("unexpected parser errors for %q: %v", input, errors)
			}
		})
	}
}
