package interp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
)

// ============================================================================
// Property Interpreter Tests
// ============================================================================

// TestPropertyFieldBacked tests field-backed properties
func TestPropertyFieldBacked(t *testing.T) {
	input := `
type TTest = class
	FName: String;
	property Name: String read FName write FName;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Name := 'Hello';
PrintLn(obj.Name);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "Hello\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyMethodBacked tests method-backed properties
func TestPropertyMethodBacked(t *testing.T) {
	input := `
type TTest = class
	FCount: Integer;

	function GetCount: Integer;
	begin
		Result := FCount;
	end;

	procedure SetCount(value: Integer);
	begin
		FCount := value;
	end;

	property Count: Integer read GetCount write SetCount;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Count := 42;
PrintLn(obj.Count);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyReadOnly tests read-only properties
func TestPropertyReadOnly(t *testing.T) {
	input := `
type TTest = class
	FSize: Integer;
	property Size: Integer read FSize;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.FSize := 100;
PrintLn(obj.Size);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "100\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyReadOnlyAssignmentError tests that writing to a read-only property fails
func TestPropertyReadOnlyAssignmentError(t *testing.T) {
	input := `
type TTest = class
	FSize: Integer;
	property Size: Integer read FSize;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Size := 100;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for read-only property assignment, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message contains the expected text (it may have location info)
	expectedError := "property 'Size' is read-only"
	if !strings.Contains(errorMsg, expectedError) {
		t.Errorf("expected error containing '%s', got '%s'", expectedError, errorMsg)
	}
}

// TestPropertyInheritance tests that properties are inherited
func TestPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FValue: Integer;
	property Value: Integer read FValue write FValue;

	constructor Create; begin end;
end;

type TDerived = class(TBase)
	FName: String;
	property Name: String read FName write FName;
end;

var obj := TDerived.Create();
obj.Value := 42;
obj.Name := 'Test';
PrintLn(obj.Value);
PrintLn(obj.Name);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\nTest\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestPropertyAutoProperty tests auto-properties (with auto-generated backing fields)
func TestPropertyAutoProperty(t *testing.T) {
	input := `
type TTest = class
	FValue: Integer;
	property Value: Integer;

	constructor Create; begin end;
end;

var obj := TTest.Create();
obj.Value := 99;
PrintLn(obj.Value);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "99\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Indexed Property Tests (Task 9.1c)
// ============================================================================

// TestIndexedPropertyRead tests basic indexed property read functionality
func TestIndexedPropertyRead(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['first', 'second', 'third']; end;
end;

var arr := TArray.Create();
PrintLn(arr.Items[0]);
PrintLn(arr.Items[1]);
PrintLn(arr.Items[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "first\nsecond\nthird\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWithVariableIndex tests indexed property with variable index
func TestIndexedPropertyWithVariableIndex(t *testing.T) {
	input := `
type TArray = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	property Values[i: Integer]: Integer read GetValue;
	constructor Create; begin FData := [0, 10, 20, 30, 40]; end;
end;

var arr := TArray.Create();
var idx: Integer;
idx := 2;
PrintLn(arr.Values[idx]);
idx := 4;
PrintLn(arr.Values[idx]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "20\n40\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyInheritance tests that indexed properties are inherited
func TestIndexedPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['base0', 'base1']; end;
end;

type TDerived = class(TBase)
end;

var obj := TDerived.Create();
PrintLn(obj.Items[0]);
PrintLn(obj.Items[1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "base0\nbase1\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWithoutIndexError tests error when indexed property accessed without index
func TestIndexedPropertyWithoutIndexError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin end;
end;

var arr := TArray.Create();
PrintLn(arr.Items);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for indexed property without index, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates the property requires an index
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "requires index") {
		t.Errorf("expected error about indexed property requiring index, got '%s'", errorMsg)
	}
}

// TestIndexedPropertyFieldBackedError tests that indexed properties cannot be field-backed
func TestIndexedPropertyFieldBackedError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;

	property Items[Index: Integer]: String read FData;

	constructor Create; begin end;
end;

var arr := TArray.Create();
PrintLn(arr.Items[0]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for field-backed indexed property, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates indexed properties require a method
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "getter method") {
		t.Errorf("expected error about indexed property requiring getter method, got '%s'", errorMsg)
	}
}

// ============================================================================
// Indexed Property Write Tests (Task 9.2b/c)
// ============================================================================

// TestIndexedPropertyWrite tests basic indexed property write functionality
func TestIndexedPropertyWrite(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem;
	constructor Create; begin FData := ['a', 'b', 'c']; end;
end;

var arr := TArray.Create();
arr.Items[0] := 'first';
arr.Items[1] := 'second';
arr.Items[2] := 'third';
PrintLn(arr.Items[0]);
PrintLn(arr.Items[1]);
PrintLn(arr.Items[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "first\nsecond\nthird\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWriteWithVariableIndex tests indexed property write with variable index
func TestIndexedPropertyWriteWithVariableIndex(t *testing.T) {
	input := `
type TArray = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	procedure SetValue(i: Integer; v: Integer); begin FData[i] := v; end;
	property Values[i: Integer]: Integer read GetValue write SetValue;
	constructor Create; begin FData := [0, 0, 0, 0, 0]; end;
end;

var arr := TArray.Create();
var idx: Integer;
idx := 2;
arr.Values[idx] := 42;
idx := 4;
arr.Values[idx] := 99;
PrintLn(arr.Values[2]);
PrintLn(arr.Values[4]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "42\n99\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyWriteInheritance tests that indexed property writes work through inheritance
func TestIndexedPropertyWriteInheritance(t *testing.T) {
	input := `
type TBase = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem;
	constructor Create; begin FData := ['', '']; end;
end;

type TDerived = class(TBase)
end;

var obj := TDerived.Create();
obj.Items[0] := 'derived0';
obj.Items[1] := 'derived1';
PrintLn(obj.Items[0]);
PrintLn(obj.Items[1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "derived0\nderived1\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestIndexedPropertyReadOnlyError tests error when writing to read-only indexed property
func TestIndexedPropertyReadOnlyError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	property Items[i: Integer]: String read GetItem;
	constructor Create; begin FData := ['a', 'b']; end;
end;

var arr := TArray.Create();
arr.Items[0] := 'test';
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for read-only indexed property write, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates the property is read-only
	if !strings.Contains(errorMsg, "read-only") {
		t.Errorf("expected error about read-only property, got '%s'", errorMsg)
	}
}

// TestIndexedPropertyWriteFieldBackedError tests that indexed property writes cannot use fields
func TestIndexedPropertyWriteFieldBackedError(t *testing.T) {
	input := `
type TArray = class
	FData: array of String;
	property Items[i: Integer]: String read FData write FData;
	constructor Create; begin end;
end;

var arr := TArray.Create();
arr.Items[0] := 'test';
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	interp := New(nil)
	result := interp.Eval(program)

	// Should get an error
	if !isError(result) {
		t.Fatalf("expected error for field-backed indexed property write, got %v", result)
	}

	errorMsg := result.(*ErrorValue).Message
	// Check that the error message indicates indexed properties require a setter method
	if !strings.Contains(errorMsg, "indexed property") && !strings.Contains(errorMsg, "setter method") {
		t.Errorf("expected error about indexed property requiring setter method, got '%s'", errorMsg)
	}
}

// ============================================================================
// Multi-Index Property Tests (Task 9.2d)
// ============================================================================

// TestMultiIndexPropertyRead tests 2D indexed property access (matrix)
func TestMultiIndexPropertyRead(t *testing.T) {
	input := `
type TMatrix = class
	FData: array of array of Integer;
	function GetCell(x, y: Integer): Integer;
	begin
		Result := FData[x][y];
	end;
	property Cells[x, y: Integer]: Integer read GetCell;
	constructor Create;
	begin
		FData := [[1, 2, 3], [4, 5, 6], [7, 8, 9]];
	end;
end;

var m := TMatrix.Create();
PrintLn(m.Cells[0, 0]);
PrintLn(m.Cells[1, 2]);
PrintLn(m.Cells[2, 1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "1\n6\n8\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestMultiIndexPropertyWrite tests 2D indexed property mutation
func TestMultiIndexPropertyWrite(t *testing.T) {
	input := `
type TMatrix = class
	FData: array of array of Integer;
	function GetCell(x, y: Integer): Integer; begin Result := FData[x][y]; end;
	procedure SetCell(x, y: Integer; value: Integer); begin FData[x][y] := value; end;
	property Cells[x, y: Integer]: Integer read GetCell write SetCell;
	constructor Create; begin FData := [[0, 0], [0, 0]]; end;
end;

var m := TMatrix.Create();
m.Cells[0, 1] := 99;
m.Cells[1, 0] := 42;
PrintLn(m.Cells[0, 1]);
PrintLn(m.Cells[1, 0]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "99\n42\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestThreeDimensionalProperty tests 3D indexed property (voxel grid)
func TestThreeDimensionalProperty(t *testing.T) {
	input := `
type TVoxelGrid = class
	FData: array of array of array of Integer;
	function GetVoxel(x, y, z: Integer): Integer;
	begin
		Result := FData[x][y][z];
	end;
	procedure SetVoxel(x, y, z: Integer; value: Integer);
	begin
		FData[x][y][z] := value;
	end;
	property Voxels[x, y, z: Integer]: Integer read GetVoxel write SetVoxel;
	constructor Create;
	begin
		FData := [[[1, 2], [3, 4]], [[5, 6], [7, 8]]];
	end;
end;

var grid := TVoxelGrid.Create();
PrintLn(grid.Voxels[0, 0, 0]);
PrintLn(grid.Voxels[1, 1, 1]);
grid.Voxels[0, 1, 0] := 100;
PrintLn(grid.Voxels[0, 1, 0]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "1\n8\n100\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestMultiIndexPropertyInheritance tests that multi-index properties work through inheritance
func TestMultiIndexPropertyInheritance(t *testing.T) {
	input := `
type TBaseMatrix = class
	FData: array of array of String;
	function GetItem(row, col: Integer): String; begin Result := FData[row][col]; end;
	procedure SetItem(row, col: Integer; value: String); begin FData[row][col] := value; end;
	property Items[row, col: Integer]: String read GetItem write SetItem;
	constructor Create; begin FData := [['a', 'b'], ['c', 'd']]; end;
end;

type TDerivedMatrix = class(TBaseMatrix)
end;

var m := TDerivedMatrix.Create();
PrintLn(m.Items[0, 0]);
PrintLn(m.Items[1, 1]);
m.Items[0, 1] := 'X';
PrintLn(m.Items[0, 1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "a\nd\nX\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Expression-Based Property Tests (Task 9.3c)
// ============================================================================

// TestExpressionBasedPropertySimple tests basic expression-based property getters
func TestExpressionBasedPropertySimple(t *testing.T) {
	input := `
type TRectangle = class
	FWidth: Integer;
	FHeight: Integer;
	property Area: Integer read (FWidth * FHeight);

	constructor Create(w, h: Integer);
	begin
		FWidth := w;
		FHeight := h;
	end;
end;

var rect := TRectangle.Create(10, 5);
PrintLn(IntToStr(rect.Area));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Area should be 10 * 5 = 50
	expectedOutput := "50\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyComplex tests complex expression-based properties
func TestExpressionBasedPropertyComplex(t *testing.T) {
	input := `
type TRectangle = class
	FWidth: Integer;
	FHeight: Integer;
	property Perimeter: Integer read (2 * (FWidth + FHeight));

	constructor Create(w, h: Integer);
	begin
		FWidth := w;
		FHeight := h;
	end;
end;

var rect := TRectangle.Create(10, 5);
PrintLn(IntToStr(rect.Perimeter));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Perimeter should be 2 * (10 + 5) = 30
	expectedOutput := "30\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyWithFunctionCall tests expression with function calls
func TestExpressionBasedPropertyWithFunctionCall(t *testing.T) {
	input := `
type TDataHolder = class
	FCount: Integer;
	property Count: Integer read (FCount * 2);

	constructor Create;
	begin
		FCount := 5;
	end;
end;

var data := TDataHolder.Create();
PrintLn(IntToStr(data.Count));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Count should be FCount * 2 = 10
	expectedOutput := "10\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyRecalculates tests that expressions recalculate on each access
func TestExpressionBasedPropertyRecalculates(t *testing.T) {
	input := `
type TCounter = class
	FValue: Integer;
	property Doubled: Integer read (FValue * 2);

	constructor Create;
	begin
		FValue := 5;
	end;
end;

var counter := TCounter.Create();
PrintLn(IntToStr(counter.Doubled));
counter.FValue := 10;
PrintLn(IntToStr(counter.Doubled));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - First should be 10, then 20 after FValue changes
	expectedOutput := "10\n20\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestExpressionBasedPropertyWithDivision tests expression with integer division
func TestExpressionBasedPropertyWithDivision(t *testing.T) {
	input := `
type TTest = class
	FValue: Integer;
	property Half: Integer read (FValue div 2);

	constructor Create(v: Integer);
	begin
		FValue := v;
	end;
end;

var obj := TTest.Create(20);
PrintLn(IntToStr(obj.Half));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - Half should be 20 div 2 = 10
	expectedOutput := "10\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Class Property Tests (Task 9.13)
// ============================================================================

// TestClassPropertyFieldBacked tests class property backed by class variable
func TestClassPropertyFieldBacked(t *testing.T) {
	input := `
type TCounter = class
	class var FCount: Integer;
	class property Count: Integer read FCount write FCount;
end;

TCounter.Count := 10;  // Write via property
PrintLn(IntToStr(TCounter.Count));  // Read via property
TCounter.Count := 15;  // Write via property
PrintLn(IntToStr(TCounter.Count));  // Read via property
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - property reads should return class var values
	expectedOutput := "10\n15\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyMethodBacked tests class property backed by class methods
func TestClassPropertyMethodBacked(t *testing.T) {
	input := `
type TConfig = class
	class var FVersion: String;

	class function GetVersion: String;
	begin
		if FVersion = '' then
			Result := '1.0.0'
		else
			Result := FVersion;
	end;

	class procedure SetVersion(value: String);
	begin
		FVersion := value;
	end;

	class property Version: String read GetVersion write SetVersion;
end;

PrintLn(TConfig.Version);  // Should print '1.0.0' (default via getter)
TConfig.Version := '2.0.0';  // Write via property (calls setter)
PrintLn(TConfig.Version);  // Should print '2.0.0' (via getter)
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "1.0.0\n2.0.0\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyReadOnly tests read-only class property
func TestClassPropertyReadOnly(t *testing.T) {
	input := `
type TApp = class
	class var FName: String;
	class property AppName: String read FName;
end;

TApp.FName := 'MyApp';
PrintLn(TApp.AppName);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "MyApp\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyWithInstances tests class property is shared across instances
func TestClassPropertyWithInstances(t *testing.T) {
	input := `
type TCounter = class
	class var FCount: Integer;
	class property Count: Integer read FCount write FCount;
	FName: String;

	constructor Create(name: String);
	begin
		FName := name;
		FCount := FCount + 1;
	end;
end;

var obj1 := TCounter.Create('First');
PrintLn(IntToStr(TCounter.Count));  // Should be 1

var obj2 := TCounter.Create('Second');
PrintLn(IntToStr(TCounter.Count));  // Should be 2

var obj3 := TCounter.Create('Third');
PrintLn(IntToStr(TCounter.Count));  // Should be 3
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output - instance counter should increment
	expectedOutput := "1\n2\n3\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestMixedClassAndInstanceProperties tests mixing class and instance properties
func TestMixedClassAndInstanceProperties(t *testing.T) {
	input := `
type TMixed = class
	class var FCounter: Integer;
	FName: String;

	class property Counter: Integer read FCounter write FCounter;
	property Name: String read FName write FName;

	constructor Create(name: String);
	begin
		FName := name;
		FCounter := FCounter + 1;  // Increment class var directly
	end;
end;

var obj1 := TMixed.Create('Object1');
PrintLn(obj1.Name);  // Read via instance property
PrintLn(IntToStr(TMixed.Counter));  // Read via class property

var obj2 := TMixed.Create('Object2');
PrintLn(obj2.Name);  // Read via instance property
PrintLn(IntToStr(TMixed.Counter));  // Read via class property
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	// Check output
	expectedOutput := "Object1\n1\nObject2\n2\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Class Property Write Tests (Task 9.14)
// ============================================================================

// TestClassPropertyWriteFieldBacked tests writing to field-backed class properties
func TestClassPropertyWriteFieldBacked(t *testing.T) {
	input := `
type TSettings = class
	class var FAppName: String;
	class var FVersion: Integer;

	class property AppName: String read FAppName write FAppName;
	class property Version: Integer read FVersion write FVersion;
end;

TSettings.AppName := 'MyApp';
TSettings.Version := 100;
PrintLn(TSettings.AppName);
PrintLn(IntToStr(TSettings.Version));

TSettings.AppName := 'UpdatedApp';
TSettings.Version := 200;
PrintLn(TSettings.AppName);
PrintLn(IntToStr(TSettings.Version));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "MyApp\n100\nUpdatedApp\n200\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyWriteMethodBacked tests writing to method-backed class properties
func TestClassPropertyWriteMethodBacked(t *testing.T) {
	input := `
type TValidator = class
	class var FMinValue: Integer;
	class var FMaxValue: Integer;

	class function GetMin: Integer;
	begin
		Result := FMinValue;
	end;

	class procedure SetMin(value: Integer);
	begin
		if value >= 0 then
			FMinValue := value
		else
			FMinValue := 0;  // Clamp to 0
	end;

	class function GetMax: Integer;
	begin
		Result := FMaxValue;
	end;

	class procedure SetMax(value: Integer);
	begin
		if value <= 100 then
			FMaxValue := value
		else
			FMaxValue := 100;  // Clamp to 100
	end;

	class property MinValue: Integer read GetMin write SetMin;
	class property MaxValue: Integer read GetMax write SetMax;
end;

TValidator.MinValue := -5;  // Should be clamped to 0
PrintLn(IntToStr(TValidator.MinValue));

TValidator.MinValue := 10;  // Should be set to 10
PrintLn(IntToStr(TValidator.MinValue));

TValidator.MaxValue := 150;  // Should be clamped to 100
PrintLn(IntToStr(TValidator.MaxValue));

TValidator.MaxValue := 75;  // Should be set to 75
PrintLn(IntToStr(TValidator.MaxValue));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "0\n10\n100\n75\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyWriteReadOnly tests error when writing to read-only class property
func TestClassPropertyWriteReadOnly(t *testing.T) {
	input := `
type TApp = class
	class var FName: String;
	class property AppName: String read FName;
end;

TApp.AppName := 'test';  // Should error - read-only property
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	// Should get an error about read-only property
	if !isError(result) {
		t.Fatalf("expected error for writing to read-only property, got: %v", result)
	}

	errMsg := result.(*ErrorValue).Message
	if !contains(errMsg, "read-only") && !contains(errMsg, "cannot write") {
		t.Errorf("expected read-only error, got: %s", errMsg)
	}
}

// TestClassPropertyWriteSharedState tests class property writes are shared across instances
func TestClassPropertyWriteSharedState(t *testing.T) {
	input := `
type TShared = class
	class var FCounter: Integer;
	FInstanceName: String;

	class property Counter: Integer read FCounter write FCounter;
	property Name: String read FInstanceName write FInstanceName;

	constructor Create(name: String);
	begin
		FInstanceName := name;
	end;
end;

var obj1 := TShared.Create('Object1');
var obj2 := TShared.Create('Object2');

TShared.Counter := 100;
PrintLn(obj1.Name + ': ' + IntToStr(TShared.Counter));
PrintLn(obj2.Name + ': ' + IntToStr(TShared.Counter));

TShared.Counter := 200;
PrintLn(obj1.Name + ': ' + IntToStr(TShared.Counter));
PrintLn(obj2.Name + ': ' + IntToStr(TShared.Counter));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "Object1: 100\nObject2: 100\nObject1: 200\nObject2: 200\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestClassPropertyWritePersistence tests class property writes persist across multiple accesses
func TestClassPropertyWritePersistence(t *testing.T) {
	input := `
type TState = class
	class var FFlag: Boolean;
	class var FMessage: String;

	class property Flag: Boolean read FFlag write FFlag;
	class property Message: String read FMessage write FMessage;
end;

TState.Flag := true;
TState.Message := 'initialized';

PrintLn(BoolToStr(TState.Flag));
PrintLn(TState.Message);

TState.Flag := false;
TState.Message := 'updated';

PrintLn(BoolToStr(TState.Flag));
PrintLn(TState.Message);

TState.Flag := true;
TState.Message := 'final';

PrintLn(BoolToStr(TState.Flag));
PrintLn(TState.Message);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "True\ninitialized\nFalse\nupdated\nTrue\nfinal\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// ============================================================================
// Default Property Tests (Task 9.16)
// ============================================================================

// TestDefaultPropertyRead tests reading from default property using obj[index] syntax
func TestDefaultPropertyRead(t *testing.T) {
	input := `
type TStringList = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem; default;
	constructor Create; begin FData := ['first', 'second', 'third']; end;
end;

var list := TStringList.Create();
PrintLn(list[0]);  // Using default property syntax
PrintLn(list[1]);
PrintLn(list[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "first\nsecond\nthird\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestDefaultPropertyWrite tests writing to default property using obj[index] := value syntax
func TestDefaultPropertyWrite(t *testing.T) {
	input := `
type TStringList = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem; default;
	constructor Create; begin FData := ['', '', '']; end;
end;

var list := TStringList.Create();
list[0] := 'alpha';  // Using default property write syntax
list[1] := 'beta';
list[2] := 'gamma';
PrintLn(list[0]);
PrintLn(list[1]);
PrintLn(list[2]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "alpha\nbeta\ngamma\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestDefaultPropertyEquivalence tests that obj[i] is equivalent to obj.Items[i]
func TestDefaultPropertyEquivalence(t *testing.T) {
	input := `
type TArray = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	procedure SetValue(i: Integer; v: Integer); begin FData[i] := v; end;
	property Values[i: Integer]: Integer read GetValue write SetValue; default;
	constructor Create; begin FData := [10, 20, 30]; end;
end;

var arr := TArray.Create();
PrintLn(IntToStr(arr[1]));  // Default property syntax
PrintLn(IntToStr(arr.Values[1]));  // Explicit property syntax
arr[1] := 99;  // Default property write
PrintLn(IntToStr(arr.Values[1]));  // Read via explicit syntax
arr.Values[2] := 88;  // Explicit property write
PrintLn(IntToStr(arr[2]));  // Read via default syntax
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "20\n20\n99\n88\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestDefaultPropertyInheritance tests default property inherited from parent class
func TestDefaultPropertyInheritance(t *testing.T) {
	input := `
type TBase = class
	FData: array of String;
	function GetItem(i: Integer): String; begin Result := FData[i]; end;
	procedure SetItem(i: Integer; v: String); begin FData[i] := v; end;
	property Items[i: Integer]: String read GetItem write SetItem; default;
	constructor Create; begin FData := ['a', 'b', 'c']; end;
end;

type TDerived = class(TBase)
	// Inherits default property from TBase
end;

var obj := TDerived.Create();
PrintLn(obj[0]);  // Should work via inherited default property
obj[1] := 'modified';
PrintLn(obj[1]);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "a\nmodified\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestDefaultPropertyWithVariableIndex tests default property with variable index
func TestDefaultPropertyWithVariableIndex(t *testing.T) {
	input := `
type TCollection = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	procedure SetValue(i: Integer; v: Integer); begin FData[i] := v; end;
	property Values[i: Integer]: Integer read GetValue write SetValue; default;
	constructor Create; begin FData := [100, 200, 300, 400, 500]; end;
end;

var coll := TCollection.Create();
var idx: Integer;
idx := 0;
PrintLn(IntToStr(coll[idx]));
idx := 2;
coll[idx] := 999;
PrintLn(IntToStr(coll[2]));
idx := 4;
PrintLn(IntToStr(coll[idx]));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "100\n999\n500\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}

// TestDefaultPropertyInLoop tests default property used in loop
func TestDefaultPropertyInLoop(t *testing.T) {
	input := `
type TNumbers = class
	FData: array of Integer;
	function GetValue(i: Integer): Integer; begin Result := FData[i]; end;
	procedure SetValue(i: Integer; v: Integer); begin FData[i] := v; end;
	property Values[i: Integer]: Integer read GetValue write SetValue; default;
	constructor Create; begin FData := [1, 2, 3, 4, 5]; end;
end;

var nums := TNumbers.Create();
var i: Integer;
for i := 0 to 4 do
begin
	nums[i] := nums[i] * 10;  // Using default property for both read and write
end;

for i := 0 to 4 do
	PrintLn(IntToStr(nums[i]));
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	var buf bytes.Buffer
	interp := New(&buf)
	result := interp.Eval(program)

	if isError(result) {
		t.Fatalf("eval error: %v", result)
	}

	expectedOutput := "10\n20\n30\n40\n50\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, buf.String())
	}
}
