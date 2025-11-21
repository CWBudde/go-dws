package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func TestParseHelperDeclaration(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedName    string
		expectedFor     string
		isRecordHelper  bool
		methodCount     int
		propertyCount   int
		classVarCount   int
		classConstCount int
	}{
		{
			name: "simple record helper",
			input: `type TStringHelper = record helper for String
				function ToUpper: String;
			end;`,
			expectedName:   "TStringHelper",
			expectedFor:    "String",
			isRecordHelper: true,
			methodCount:    1,
		},
		{
			name: "simple helper without record keyword",
			input: `type TIntHelper = helper for Integer
				function IsEven: Boolean;
			end;`,
			expectedName:   "TIntHelper",
			expectedFor:    "Integer",
			isRecordHelper: false,
			methodCount:    1,
		},
		{
			name: "helper with multiple methods",
			input: `type THelper = helper for String
				function ToUpper: String;
				function ToLower: String;
				procedure Clear;
			end;`,
			expectedName:   "THelper",
			expectedFor:    "String",
			isRecordHelper: false,
			methodCount:    3,
		},
		{
			name: "helper with property",
			input: `type TArrayHelper = helper for TIntArray
				property Count: Integer read GetCount;
			end;`,
			expectedName:   "TArrayHelper",
			expectedFor:    "TIntArray",
			isRecordHelper: false,
			propertyCount:  1,
		},
		{
			name: "helper with class var",
			input: `type THelper = record helper for String
				class var DefaultEncoding: String;
			end;`,
			expectedName:   "THelper",
			expectedFor:    "String",
			isRecordHelper: true,
			classVarCount:  1,
		},
		{
			name: "helper with class const",
			input: `type TMathHelper = helper for Float
				class const PI = 3.14159;
			end;`,
			expectedName:    "TMathHelper",
			expectedFor:     "Float",
			isRecordHelper:  false,
			classConstCount: 1,
		},
		{
			name: "helper with private and public sections",
			input: `type TComplexHelper = record helper for String
				private
					function InternalMethod: Integer;
				public
					function ToUpper: String;
					property Length: Integer read GetLength;
			end;`,
			expectedName:   "TComplexHelper",
			expectedFor:    "String",
			isRecordHelper: true,
			methodCount:    2,
			propertyCount:  1,
		},
		{
			name: "helper with all member types",
			input: `type TFullHelper = helper for Integer
				class const MAX_VALUE = 2147483647;
				class var DefaultValue: Integer;
				function IsPositive: Boolean;
				property AsString: String read ToString;
			end;`,
			expectedName:    "TFullHelper",
			expectedFor:     "Integer",
			isRecordHelper:  false,
			methodCount:     1,
			propertyCount:   1,
			classVarCount:   1,
			classConstCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Check that we got a program
			if program == nil {
				t.Fatal("ParseProgram() returned nil")
			}

			// Should have exactly one statement
			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			// Should be a helper declaration
			helperDecl, ok := program.Statements[0].(*ast.HelperDecl)
			if !ok {
				t.Fatalf("Expected *ast.HelperDecl, got %T", program.Statements[0])
			}

			// Check helper name
			if helperDecl.Name.Value != tt.expectedName {
				t.Errorf("Expected helper name %q, got %q", tt.expectedName, helperDecl.Name.Value)
			}

			// Check target type
			if helperDecl.ForType.String() != tt.expectedFor {
				t.Errorf("Expected ForType %q, got %q", tt.expectedFor, helperDecl.ForType.String())
			}

			// Check if it's a record helper
			if helperDecl.IsRecordHelper != tt.isRecordHelper {
				t.Errorf("Expected IsRecordHelper=%v, got %v", tt.isRecordHelper, helperDecl.IsRecordHelper)
			}

			// Check method count
			if len(helperDecl.Methods) != tt.methodCount {
				t.Errorf("Expected %d methods, got %d", tt.methodCount, len(helperDecl.Methods))
			}

			// Check property count
			if len(helperDecl.Properties) != tt.propertyCount {
				t.Errorf("Expected %d properties, got %d", tt.propertyCount, len(helperDecl.Properties))
			}

			// Check class var count
			if len(helperDecl.ClassVars) != tt.classVarCount {
				t.Errorf("Expected %d class vars, got %d", tt.classVarCount, len(helperDecl.ClassVars))
			}

			// Check class const count
			if len(helperDecl.ClassConsts) != tt.classConstCount {
				t.Errorf("Expected %d class consts, got %d", tt.classConstCount, len(helperDecl.ClassConsts))
			}
		})
	}
}

func TestParseHelperErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // expected error substring
	}{
		{
			name:     "missing for keyword",
			input:    `type THelper = helper String end;`,
			expected: "expected next token to be FOR",
		},
		{
			name:     "missing target type",
			input:    `type THelper = helper for end;`,
			expected: "expected type name",
		},
		{
			name:     "missing end keyword",
			input:    `type THelper = helper for String function Test: Boolean;`,
			expected: "expected 'end'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			// Should have errors
			if len(p.Errors()) == 0 {
				t.Fatal("Expected parser errors, got none")
			}

			// Check if the expected error substring is present
			found := false
			for _, err := range p.Errors() {
				if containsSubstring(err, tt.expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error containing %q, got errors: %v", tt.expected, p.Errors())
			}
		})
	}
}

func TestParseRecordVsHelper(t *testing.T) {
	// Test that we can correctly distinguish between record and helper declarations
	tests := []struct {
		name     string
		input    string
		isHelper bool
		isRecord bool
	}{
		{
			name: "regular record",
			input: `type TPoint = record
				X: Integer;
				Y: Integer;
			end;`,
			isRecord: true,
			isHelper: false,
		},
		{
			name: "record helper",
			input: `type TPointHelper = record helper for TPoint
				function Distance: Float;
			end;`,
			isRecord: false,
			isHelper: true,
		},
		{
			name: "simple helper",
			input: `type THelper = helper for String
				function Test: Boolean;
			end;`,
			isRecord: false,
			isHelper: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			stmt := program.Statements[0]

			if tt.isHelper {
				if _, ok := stmt.(*ast.HelperDecl); !ok {
					t.Errorf("Expected *ast.HelperDecl, got %T", stmt)
				}
			}

			if tt.isRecord {
				if _, ok := stmt.(*ast.RecordDecl); !ok {
					t.Errorf("Expected *ast.RecordDecl, got %T", stmt)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsSubstring(v interface{}, substr string) bool {
	var s string
	switch val := v.(type) {
	case *ParserError:
		if val == nil {
			return false
		}
		s = val.Message
	case string:
		s = val
	default:
		return false
	}
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0:len(substr)] == substr || containsSubstring(s[1:], substr))))
}

// TestParseHelperInheritance tests parsing of helper inheritance
func TestParseHelperInheritance(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedName       string
		expectedParent     string
		expectedFor        string
		isRecordHelper     bool
		expectParentHelper bool
	}{
		{
			name: "helper with parent",
			input: `type TChildHelper = helper(TParentHelper) for String
				function ToLower: String;
			end;`,
			expectedName:       "TChildHelper",
			expectedParent:     "TParentHelper",
			expectedFor:        "String",
			isRecordHelper:     false,
			expectParentHelper: true,
		},
		{
			name: "record helper with parent",
			input: `type TChildHelper = record helper(TParentHelper) for Integer
				function Double: Integer;
			end;`,
			expectedName:       "TChildHelper",
			expectedParent:     "TParentHelper",
			expectedFor:        "Integer",
			isRecordHelper:     true,
			expectParentHelper: true,
		},
		{
			name: "helper without parent",
			input: `type TSimpleHelper = helper for String
				function Test: Boolean;
			end;`,
			expectedName:       "TSimpleHelper",
			expectedFor:        "String",
			isRecordHelper:     false,
			expectParentHelper: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()

			// Check for parser errors
			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			// Check that we got a program
			if program == nil {
				t.Fatal("ParseProgram() returned nil")
			}

			// Should have exactly one statement
			if len(program.Statements) != 1 {
				t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
			}

			// Should be a helper declaration
			helperDecl, ok := program.Statements[0].(*ast.HelperDecl)
			if !ok {
				t.Fatalf("Expected *ast.HelperDecl, got %T", program.Statements[0])
			}

			// Check helper name
			if helperDecl.Name.Value != tt.expectedName {
				t.Errorf("Expected helper name %q, got %q", tt.expectedName, helperDecl.Name.Value)
			}

			// Check parent helper
			if tt.expectParentHelper {
				if helperDecl.ParentHelper == nil {
					t.Fatal("Expected parent helper, got nil")
				}
				if helperDecl.ParentHelper.Value != tt.expectedParent {
					t.Errorf("Expected parent helper %q, got %q", tt.expectedParent, helperDecl.ParentHelper.Value)
				}
			} else {
				if helperDecl.ParentHelper != nil {
					t.Errorf("Expected no parent helper, got %q", helperDecl.ParentHelper.Value)
				}
			}

			// Check target type
			if helperDecl.ForType.String() != tt.expectedFor {
				t.Errorf("Expected ForType %q, got %q", tt.expectedFor, helperDecl.ForType.String())
			}

			// Check if it's a record helper
			if helperDecl.IsRecordHelper != tt.isRecordHelper {
				t.Errorf("Expected IsRecordHelper=%v, got %v", tt.isRecordHelper, helperDecl.IsRecordHelper)
			}
		})
	}
}

// TestParseHelperInheritanceErrors tests error handling for helper inheritance
func TestParseHelperInheritanceErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // expected error substring
	}{
		{
			name:     "missing parent helper name",
			input:    `type THelper = helper() for String end;`,
			expected: "expected parent helper name",
		},
		{
			name:     "missing closing paren",
			input:    `type THelper = helper(TParent for String end;`,
			expected: "expected ')'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			// Should have errors
			if len(p.Errors()) == 0 {
				t.Fatal("Expected parser errors, got none")
			}

			// Check if the expected error substring is present
			found := false
			for _, err := range p.Errors() {
				if containsSubstring(err, tt.expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error containing %q, got errors: %v", tt.expected, p.Errors())
			}
		})
	}
}
