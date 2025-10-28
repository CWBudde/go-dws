package parser

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestParseFunctionPointerTypeDeclarations tests parsing of function pointer type declarations.
// Task 9.157: Add parser tests for function pointer types
func TestParseFunctionPointerTypeDeclarations(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedName    string
		isFunction      bool
		paramCount      int
		hasReturnType   bool
		returnTypeName  string
		ofObject        bool
		firstParamName  string
		firstParamType  string
		firstParamByRef bool
	}{
		{
			name:            "simple function pointer with one parameter",
			input:           "type TFunc = function(x: Integer): Boolean;",
			expectedName:    "TFunc",
			isFunction:      true,
			paramCount:      1,
			hasReturnType:   true,
			returnTypeName:  "Boolean",
			ofObject:        false,
			firstParamName:  "x",
			firstParamType:  "Integer",
			firstParamByRef: false,
		},
		{
			name:            "simple procedure pointer with one parameter",
			input:           "type TProc = procedure(msg: String);",
			expectedName:    "TProc",
			isFunction:      false,
			paramCount:      1,
			hasReturnType:   false,
			ofObject:        false,
			firstParamName:  "msg",
			firstParamType:  "String",
			firstParamByRef: false,
		},
		{
			name:           "function pointer with no parameters",
			input:          "type TCallback = function(): Integer;",
			expectedName:   "TCallback",
			isFunction:     true,
			paramCount:     0,
			hasReturnType:  true,
			returnTypeName: "Integer",
			ofObject:       false,
		},
		{
			name:          "procedure pointer with no parameters",
			input:         "type TSimpleProc = procedure();",
			expectedName:  "TSimpleProc",
			isFunction:    false,
			paramCount:    0,
			hasReturnType: false,
			ofObject:      false,
		},
		{
			name:            "function pointer with multiple parameters",
			input:           "type TCompare = function(a, b: Integer): Integer;",
			expectedName:    "TCompare",
			isFunction:      true,
			paramCount:      2,
			hasReturnType:   true,
			returnTypeName:  "Integer",
			ofObject:        false,
			firstParamName:  "a",
			firstParamType:  "Integer",
			firstParamByRef: false,
		},
		{
			name:            "procedure pointer with by-ref parameter",
			input:           "type TModifier = procedure(var x: Integer);",
			expectedName:    "TModifier",
			isFunction:      false,
			paramCount:      1,
			hasReturnType:   false,
			ofObject:        false,
			firstParamName:  "x",
			firstParamType:  "Integer",
			firstParamByRef: true,
		},
		{
			name:            "method pointer (of object) - procedure",
			input:           "type TEvent = procedure(Sender: TObject) of object;",
			expectedName:    "TEvent",
			isFunction:      false,
			paramCount:      1,
			hasReturnType:   false,
			ofObject:        true,
			firstParamName:  "Sender",
			firstParamType:  "TObject",
			firstParamByRef: false,
		},
		{
			name:            "method pointer (of object) - function",
			input:           "type TValidate = function(x: Integer): Boolean of object;",
			expectedName:    "TValidate",
			isFunction:      true,
			paramCount:      1,
			hasReturnType:   true,
			returnTypeName:  "Boolean",
			ofObject:        true,
			firstParamName:  "x",
			firstParamType:  "Integer",
			firstParamByRef: false,
		},
		{
			name:          "method pointer with no parameters",
			input:         "type TNotify = procedure() of object;",
			expectedName:  "TNotify",
			isFunction:    false,
			paramCount:    0,
			hasReturnType: false,
			ofObject:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program.Statements does not contain 1 statement. got=%d",
					len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.TypeDeclaration)
			if !ok {
				t.Fatalf("program.Statements[0] is not *ast.TypeDeclaration. got=%T",
					program.Statements[0])
			}

			if !stmt.IsFunctionPointer {
				t.Fatalf("expected IsFunctionPointer to be true, got false")
			}

			if stmt.Name.Value != tt.expectedName {
				t.Errorf("expected type name %q, got %q", tt.expectedName, stmt.Name.Value)
			}

			if stmt.FunctionPointerType == nil {
				t.Fatalf("FunctionPointerType is nil")
			}

			funcPtrType := stmt.FunctionPointerType

			// Check if it's a function or procedure
			isFunc := funcPtrType.ReturnType != nil
			if isFunc != tt.isFunction {
				t.Errorf("expected isFunction=%v, got=%v", tt.isFunction, isFunc)
			}

			// Check parameter count
			if len(funcPtrType.Parameters) != tt.paramCount {
				t.Errorf("expected %d parameters, got %d", tt.paramCount, len(funcPtrType.Parameters))
			}

			// Check return type
			if tt.hasReturnType {
				if funcPtrType.ReturnType == nil {
					t.Errorf("expected return type, got nil")
				} else if funcPtrType.ReturnType.Name != tt.returnTypeName {
					t.Errorf("expected return type %q, got %q", tt.returnTypeName, funcPtrType.ReturnType.Name)
				}
			} else {
				if funcPtrType.ReturnType != nil {
					t.Errorf("expected no return type, got %v", funcPtrType.ReturnType)
				}
			}

			// Check of object clause
			if funcPtrType.OfObject != tt.ofObject {
				t.Errorf("expected OfObject=%v, got=%v", tt.ofObject, funcPtrType.OfObject)
			}

			// Check first parameter if present
			if tt.paramCount > 0 && tt.firstParamName != "" {
				firstParam := funcPtrType.Parameters[0]
				if firstParam.Name.Value != tt.firstParamName {
					t.Errorf("expected first param name %q, got %q", tt.firstParamName, firstParam.Name.Value)
				}
				if firstParam.Type.Name != tt.firstParamType {
					t.Errorf("expected first param type %q, got %q", tt.firstParamType, firstParam.Type.Name)
				}
				if firstParam.ByRef != tt.firstParamByRef {
					t.Errorf("expected first param ByRef=%v, got=%v", tt.firstParamByRef, firstParam.ByRef)
				}
			}
		})
	}
}

// TestParseFunctionPointerTypeDeclarationString tests the String() method output.
func TestParseFunctionPointerTypeDeclarationString(t *testing.T) {
	input := `type TComparator = function(a, b: Integer): Integer;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d",
			len(program.Statements))
	}

	stmt := program.Statements[0]
	expected := "type TComparator = function(a: Integer; b: Integer): Integer"
	if stmt.String() != expected {
		t.Errorf("expected String() to be %q, got %q", expected, stmt.String())
	}
}

// TestParseAddressOfExpression tests parsing of address-of (@) expressions.
// Task 9.158: Add parser tests for address-of expressions
func TestParseAddressOfExpression(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTarget string
	}{
		{
			name:           "simple address-of identifier",
			input:          "var x := @MyFunction;",
			expectedTarget: "MyFunction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got=%d", len(program.Statements))
			}

			varStmt, ok := program.Statements[0].(*ast.VarDeclStatement)
			if !ok {
				t.Fatalf("statement is not *ast.VarDeclStatement. got=%T", program.Statements[0])
			}

			decl := varStmt
			addrOf, ok := decl.Value.(*ast.AddressOfExpression)
			if !ok {
				t.Fatalf("value is not *ast.AddressOfExpression. got=%T", decl.Value)
			}

			if addrOf.Operator == nil {
				t.Fatalf("Operator is nil")
			}

			ident, ok := addrOf.Operator.(*ast.Identifier)
			if !ok {
				t.Fatalf("operator is not *ast.Identifier. got=%T", addrOf.Operator)
			}

			if ident.Value != tt.expectedTarget {
				t.Errorf("expected target to be %q, got %q", tt.expectedTarget, ident.Value)
			}
		})
	}
}

// TestParseAddressOfInAssignment tests address-of in variable assignments.
func TestParseAddressOfInAssignment(t *testing.T) {
	input := `
var f: TFunc;
begin
  f := @MyFunction;
end.
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) < 2 {
		t.Fatalf("expected at least 2 statements, got=%d", len(program.Statements))
	}

	// Check the begin block
	beginBlock, ok := program.Statements[1].(*ast.BlockStatement)
	if !ok {
		t.Fatalf("statement is not *ast.BlockStatement. got=%T", program.Statements[1])
	}

	if len(beginBlock.Statements) < 1 {
		t.Fatalf("expected at least 1 statement in begin block, got=%d", len(beginBlock.Statements))
	}

	// Check the assignment
	assignStmt, ok := beginBlock.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is not *ast.AssignmentStatement. got=%T", beginBlock.Statements[0])
	}

	// Check the right side is an address-of expression
	addrOf, ok := assignStmt.Value.(*ast.AddressOfExpression)
	if !ok {
		t.Fatalf("assignment value is not *ast.AddressOfExpression. got=%T", assignStmt.Value)
	}

	// Check the target
	ident, ok := addrOf.Operator.(*ast.Identifier)
	if !ok {
		t.Fatalf("address-of operator is not *ast.Identifier. got=%T", addrOf.Operator)
	}

	if ident.Value != "MyFunction" {
		t.Errorf("expected target to be 'MyFunction', got %q", ident.Value)
	}
}

// TestParseAddressOfInFunctionCall tests address-of as function argument.
func TestParseAddressOfInFunctionCall(t *testing.T) {
	input := `Sort(arr, @Ascending);`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got=%d", len(program.Statements))
	}

	exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	callExpr, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is not *ast.CallExpression. got=%T", exprStmt.Expression)
	}

	if len(callExpr.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got=%d", len(callExpr.Arguments))
	}

	// Check second argument is address-of
	addrOf, ok := callExpr.Arguments[1].(*ast.AddressOfExpression)
	if !ok {
		t.Fatalf("second argument is not *ast.AddressOfExpression. got=%T", callExpr.Arguments[1])
	}

	ident, ok := addrOf.Operator.(*ast.Identifier)
	if !ok {
		t.Fatalf("address-of operator is not *ast.Identifier. got=%T", addrOf.Operator)
	}

	if ident.Value != "Ascending" {
		t.Errorf("expected target to be 'Ascending', got %q", ident.Value)
	}
}
