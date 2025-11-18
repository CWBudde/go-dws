package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// Task 2.7.1.6: Integration testing for cursor migration
//
// This file tests complex programs that combine multiple migrated features:
// - Uses clauses
// - Type declarations (enums, records, classes, arrays, sets)
// - Const declarations
// - Var declarations
// - Operator overloading
// - Program declarations
// - Nested structures
//
// These tests ensure that the dispatcher pattern works correctly when
// features are combined in realistic programs.

// TestMigration_Integration_SimpleProgram tests a simple complete program
func TestMigration_Integration_SimpleProgram(t *testing.T) {
	input := `
program HelloWorld;

const
	GREETING = 'Hello, World!';
	VERSION = 1;

var
	message: String;
	count: Integer;

begin
	message := GREETING;
	count := VERSION;
end.
`

	// Traditional mode
	tradParser := New(lexer.New(input))
	tradProgram := tradParser.ParseProgram()
	if len(tradParser.Errors()) > 0 {
		t.Errorf("Traditional parser errors: %v", tradParser.Errors())
	}

	// Cursor mode
	cursorParser := NewCursorParser(lexer.New(input))
	cursorProgram := cursorParser.ParseProgram()
	if len(cursorParser.Errors()) > 0 {
		t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
	}

	// Both should produce programs
	if tradProgram == nil || cursorProgram == nil {
		t.Fatal("Parser returned nil program")
	}

	// Error counts should match
	tradErrors := len(tradParser.Errors())
	cursorErrors := len(cursorParser.Errors())
	if tradErrors != cursorErrors {
		t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
			tradErrors, cursorErrors)
	}

	// AST strings should match
	if tradProgram.String() != cursorProgram.String() {
		t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
			tradProgram.String(), cursorProgram.String())
	}
}

// TestMigration_Integration_UnitWithTypes tests unit with various type declarations
func TestMigration_Integration_UnitWithTypes(t *testing.T) {
	input := `
unit MyTypes;

interface

type
	TWeekday = (Monday, Tuesday, Wednesday, Thursday, Friday);
	TDays = set of TWeekday;
	TInts = array[1..10] of Integer;
	TMatrix = array[0..9, 0..9] of Float;

	TPoint = record
		X, Y: Integer;
	end;

implementation

end.
`

	// Traditional mode
	tradParser := New(lexer.New(input))
	tradProgram := tradParser.ParseProgram()
	if len(tradParser.Errors()) > 0 {
		t.Errorf("Traditional parser errors: %v", tradParser.Errors())
	}

	// Cursor mode
	cursorParser := NewCursorParser(lexer.New(input))
	cursorProgram := cursorParser.ParseProgram()
	if len(cursorParser.Errors()) > 0 {
		t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
	}

	// Both should produce programs
	if tradProgram == nil || cursorProgram == nil {
		t.Fatal("Parser returned nil program")
	}

	// Error counts should match
	tradErrors := len(tradParser.Errors())
	cursorErrors := len(cursorParser.Errors())
	if tradErrors != cursorErrors {
		t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
			tradErrors, cursorErrors)
	}
}

// TestMigration_Integration_MixedDeclarations tests combinations in single statements
func TestMigration_Integration_MixedDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"uses clause",
			"uses System, SysUtils;",
		},
		{
			"type with set",
			"type TColors = set of TColor;",
		},
		{
			"type with array",
			"type TInts = array[1..10] of Integer;",
		},
		{
			"const declaration",
			"const MAX_SIZE = 100;",
		},
		{
			"var with simple type",
			"var count: Integer;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_Integration_OperatorsAndTypes tests operator declarations
func TestMigration_Integration_OperatorsAndTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"type declaration",
			"type TVector = record X, Y, Z: Float; end;",
		},
		{
			"operator with record type",
			"operator + (const A, B: TVector) : TVector uses VectorAdd;",
		},
		{
			"operator with array type",
			"operator * (const A, B: TMatrix) : TMatrix uses MatrixMul;",
		},
		{
			"type with 2D array",
			"type TMatrix = array[0..2, 0..2] of Float;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_Integration_ComplexArrays tests complex array scenarios
func TestMigration_Integration_ComplexArrays(t *testing.T) {
	input := `
type
	TByteArray = array of Byte;
	TWordArray = array of Word;
	TStaticInts = array[0..99] of Integer;
	T2DArray = array[1..10, 1..20] of Float;
	T3DArray = array[0..4, 0..4, 0..4] of Boolean;
	TNestedArray = array[1..5] of array[1..10] of String;

var
	bytes: TByteArray;
	matrix: T2DArray;
	cube: T3DArray;

const
	SIZES: array[1..3] of Integer = (10, 20, 30);
`

	// Traditional mode
	tradParser := New(lexer.New(input))
	tradProgram := tradParser.ParseProgram()
	if len(tradParser.Errors()) > 0 {
		t.Errorf("Traditional parser errors: %v", tradParser.Errors())
	}

	// Cursor mode
	cursorParser := NewCursorParser(lexer.New(input))
	cursorProgram := cursorParser.ParseProgram()
	if len(cursorParser.Errors()) > 0 {
		t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
	}

	// Both should produce programs
	if tradProgram == nil || cursorProgram == nil {
		t.Fatal("Parser returned nil program")
	}

	// Error counts should match
	tradErrors := len(tradParser.Errors())
	cursorErrors := len(cursorParser.Errors())
	if tradErrors != cursorErrors {
		t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
			tradErrors, cursorErrors)
	}
}

// TestMigration_Integration_SetOperations tests set type declarations
func TestMigration_Integration_SetOperations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"set of enum",
			"type TLetters = set of TLetter;",
		},
		{
			"set of char",
			"type TChars = set of Char;",
		},
		{
			"set of byte",
			"type TBytes = set of Byte;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}

// TestMigration_Integration_ErrorRecovery tests error recovery
func TestMigration_Integration_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"incomplete type",
			`type T1 = ;`,
		},
		{
			"missing var type and init",
			`var X;`,
		},
		{
			"missing set element type",
			`type TSet = set of;`,
		},
		{
			"missing array bounds",
			`type TArr = array[] of Integer;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			_ = tradParser.ParseProgram()
			tradErrors := len(tradParser.Errors())

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			_ = cursorParser.ParseProgram()
			cursorErrors := len(cursorParser.Errors())

			// Both should have errors
			if tradErrors == 0 {
				t.Error("Traditional parser should have errors")
			}
			if cursorErrors == 0 {
				t.Error("Cursor parser should have errors")
			}

			// Log error count differences (not enforced, as recovery may differ)
			if tradErrors != cursorErrors {
				t.Logf("Error count difference: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}
		})
	}
}

// TestMigration_Integration_NestedStructures tests nested type definitions
func TestMigration_Integration_NestedStructures(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"nested arrays",
			"type TMatrix = array[1..3] of array[1..4] of Integer;",
		},
		{
			"array of records",
			"type TData = array[1..5] of TRecord;",
		},
		{
			"complex nesting",
			"type TComplex = array[0..2, 0..3] of array of String;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Traditional mode
			tradParser := New(lexer.New(tt.input))
			tradProgram := tradParser.ParseProgram()
			if len(tradParser.Errors()) > 0 {
				t.Errorf("Traditional parser errors: %v", tradParser.Errors())
			}

			// Cursor mode
			cursorParser := NewCursorParser(lexer.New(tt.input))
			cursorProgram := cursorParser.ParseProgram()
			if len(cursorParser.Errors()) > 0 {
				t.Errorf("Cursor parser errors: %v", cursorParser.Errors())
			}

			// Both should produce programs
			if tradProgram == nil || cursorProgram == nil {
				t.Fatal("Parser returned nil program")
			}

			// Error counts should match
			tradErrors := len(tradParser.Errors())
			cursorErrors := len(cursorParser.Errors())
			if tradErrors != cursorErrors {
				t.Errorf("Error count mismatch: traditional=%d, cursor=%d",
					tradErrors, cursorErrors)
			}

			// AST strings should match
			if tradProgram.String() != cursorProgram.String() {
				t.Errorf("AST mismatch:\nTraditional: %s\nCursor: %s",
					tradProgram.String(), cursorProgram.String())
			}
		})
	}
}
