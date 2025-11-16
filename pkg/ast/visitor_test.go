package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// TestWalk_VisitsAllNodes tests that Walk visits all nodes in the tree
func TestWalk_VisitsAllNodes(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 42;
		var y: String := 'hello';

		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;
	`)

	tree := program

	// Count all visited nodes
	nodeCount := 0
	visitor := &countingVisitor{count: &nodeCount}
	ast.Walk(visitor, tree)

	// Should visit many nodes (program, statements, expressions, etc.)
	if nodeCount == 0 {
		t.Error("Walk() did not visit any nodes")
	}

	// We should visit at least: program, 3 statements (2 vars + 1 func),
	// plus all their children
	if nodeCount < 10 {
		t.Errorf("Expected at least 10 nodes visited, got %d", nodeCount)
	}
}

// TestWalk_VisitorReturnsNil tests that returning nil from Visit stops traversal
func TestWalk_VisitorReturnsNil(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 42;
		function Test(): Integer;
		begin
			Result := 1;
		end;
	`)

	tree := program

	// Count only first-level nodes (stop after visiting root)
	nodeCount := 0
	visitor := &stopAfterFirstVisitor{count: &nodeCount}
	ast.Walk(visitor, tree)

	// Should only visit the root Program node
	if nodeCount != 1 {
		t.Errorf("Expected 1 node visited (root only), got %d", nodeCount)
	}
}

// TestInspect_FindsFunctions tests using Inspect to find function declarations
func TestInspect_FindsFunctions(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Multiply(a, b: Integer): Integer;
		begin
			Result := a * b;
		end;
	`)

	tree := program

	// Find all function declarations
	functionNames := []string{}
	ast.Inspect(tree, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FunctionDecl); ok {
			functionNames = append(functionNames, funcDecl.Name.Value)
		}
		return true // continue traversal
	})

	if len(functionNames) != 2 {
		t.Errorf("Expected 2 functions, found %d", len(functionNames))
	}

	expected := map[string]bool{"Add": true, "Multiply": true}
	for _, name := range functionNames {
		if !expected[name] {
			t.Errorf("Unexpected function name: %s", name)
		}
	}
}

// TestInspect_FindsVariables tests finding all variable declarations
func TestInspect_FindsVariables(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 1;
		var y: String := 'test';
		var z: Float := 3.14;
	`)

	tree := program

	varCount := 0
	ast.Inspect(tree, func(n ast.Node) bool {
		if _, ok := n.(*ast.VarDeclStatement); ok {
			varCount++
		}
		return true
	})

	if varCount != 3 {
		t.Errorf("Expected 3 variables, found %d", varCount)
	}
}

// TestInspect_StopsTraversal tests that returning false stops traversal
func TestInspect_StopsTraversal(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 42;
		var y: String := 'hello';
	`)

	tree := program

	visitCount := 0
	ast.Inspect(tree, func(n ast.Node) bool {
		visitCount++
		// Stop after first node
		return false
	})

	// Should only visit the root node
	if visitCount != 1 {
		t.Errorf("Expected 1 visit (stopped traversal), got %d", visitCount)
	}
}

// TestInspect_NestedStructures tests traversing nested structures
func TestInspect_NestedStructures(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Test(): Integer;
		begin
			if x > 0 then
				Result := 1
			else
				Result := 0;
		end;
	`)

	tree := program

	// Find if statement
	foundIf := false
	ast.Inspect(tree, func(n ast.Node) bool {
		if _, ok := n.(*ast.IfStatement); ok {
			foundIf = true
		}
		return true
	})

	if !foundIf {
		t.Error("Did not find IfStatement in nested structure")
	}
}

// TestWalk_AllNodeTypes tests that Walk handles all major node types
func TestWalk_AllNodeTypes(t *testing.T) {
	engine, _ := dwscript.New()

	// Complex program with many node types
	program, _ := engine.Parse(`
		type TMyEnum = (Value1, Value2, Value3);

		type TMyRecord = record
			Field1: Integer;
			Field2: String;
		end;

		class TMyClass
			FValue: Integer;
			function GetValue(): Integer;
		end;

		function TMyClass.GetValue(): Integer;
		begin
			Result := FValue;
		end;

		var arr: array of Integer;
		var e: TMyEnum;

		begin
			arr := [1, 2, 3];
			e := TMyEnum.Value1;

			for var i := 0 to 10 do
				PrintLn(i);

			while x > 0 do
				x := x - 1;

			case e of
				Value1: PrintLn('one');
				Value2: PrintLn('two');
			end;

			try
				DoSomething();
			except
				on E: Exception do
					PrintLn('error');
			end;
		end
	`)

	tree := program

	// Collect all node type names
	nodeTypes := make(map[string]bool)
	ast.Inspect(tree, func(n ast.Node) bool {
		if n != nil {
			nodeTypes[fmt.Sprintf("%T", n)] = true
		}
		return true
	})

	// Verify we encountered various node types
	expectedTypes := []string{
		"*ast.Program",
		"*ast.EnumDecl",
		"*ast.RecordDecl",
		// Note: ClassDecl not included as parser doesn't fully support it yet
		"*ast.FunctionDecl",
		"*ast.VarDeclStatement",
		"*ast.BlockStatement",
		"*ast.ForStatement",
		"*ast.WhileStatement",
		"*ast.CaseStatement",
		"*ast.TryStatement",
		"*ast.ArrayLiteralExpression",
		"*ast.CallExpression",
	}

	for _, expectedType := range expectedTypes {
		if !nodeTypes[expectedType] {
			t.Errorf("Did not encounter node type: %s", expectedType)
		}
	}
}

// countingVisitor counts all visited nodes
type countingVisitor struct {
	count *int
}

func (v *countingVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		*v.count++
	}
	return v
}

// stopAfterFirstVisitor stops traversal after first node
type stopAfterFirstVisitor struct {
	count *int
}

func (v *stopAfterFirstVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		*v.count++
		return nil // stop traversal
	}
	return v
}

// Example_walk demonstrates using the Walk function with a custom visitor
func Example_walk() {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;
	`)

	tree := program

	// Create a visitor that prints function names
	visitor := &functionPrinter{}
	ast.Walk(visitor, tree)

	// Output:
	// Function: Add
}

type functionPrinter struct{}

func (v *functionPrinter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	if funcDecl, ok := node.(*ast.FunctionDecl); ok {
		fmt.Printf("Function: %s\n", funcDecl.Name.Value)
	}

	return v
}

// Example_inspect demonstrates using Inspect to find specific nodes
func Example_inspect() {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 10;
		var y: Integer := 20;
		var z: String := 'hello';
	`)

	tree := program

	// Find all variable declarations
	varNames := []string{}
	ast.Inspect(tree, func(n ast.Node) bool {
		if varDecl, ok := n.(*ast.VarDeclStatement); ok {
			for _, name := range varDecl.Names {
				varNames = append(varNames, name.Value)
			}
		}
		return true
	})

	fmt.Println("Variables:", strings.Join(varNames, ", "))

	// Output:
	// Variables: x, y, z
}

// Example_inspectStopTraversal shows how to stop traversal at specific nodes
func Example_inspectStopTraversal() {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Outer(): Integer;
		begin
			function Inner(): Integer;
			begin
				Result := 42;
			end;
			Result := Inner();
		end;
	`)

	tree := program

	// Find only top-level functions (don't descend into function bodies)
	topLevelFunctions := []string{}
	ast.Inspect(tree, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FunctionDecl); ok {
			topLevelFunctions = append(topLevelFunctions, funcDecl.Name.Value)
			// Don't traverse into the function body (would find nested functions)
			return false
		}
		return true
	})

	// Should only find 'Outer', not 'Inner'
	fmt.Println("Top-level functions:", strings.Join(topLevelFunctions, ", "))

	// Output:
	// Top-level functions: Outer
}

// TestInspect_ComplexExpressions tests traversing complex expressions
func TestInspect_ComplexExpressions(t *testing.T) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var result: Integer := (a + b) * (c - d) / e;
	`)

	tree := program

	// Count binary expressions
	binaryCount := 0
	ast.Inspect(tree, func(n ast.Node) bool {
		if _, ok := n.(*ast.BinaryExpression); ok {
			binaryCount++
		}
		return true
	})

	// Should have: +, *, -, / = 4 binary expressions
	if binaryCount != 4 {
		t.Errorf("Expected 4 binary expressions, found %d", binaryCount)
	}
}

// TestWalk_WithNilNodes tests that Walk handles nil nodes gracefully
func TestWalk_WithNilNodes(t *testing.T) {
	// Create a simple program with proper Names field (slice, not single)
	program := &ast.Program{
		Statements: []ast.Statement{
			&ast.VarDeclStatement{
				Names: []*ast.Identifier{{Value: "x"}},
				Value: nil, // Explicitly nil
			},
		},
	}

	// Should not panic with nil child nodes
	nodeCount := 0
	ast.Inspect(program, func(n ast.Node) bool {
		if n != nil {
			nodeCount++
		}
		return true
	})

	if nodeCount == 0 {
		t.Error("Expected to visit some nodes")
	}
}

// TestWalk_HelperTypesAsNodes tests that helper types now properly implement Node
// and are visited during traversal (Task 9.20)
func TestWalk_HelperTypesAsNodes(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		expectedType string
		description  string
	}{
		{
			name: "Parameter in function",
			code: `
				function Add(a: Integer; b: Integer): Integer;
				begin
					Result := a + b;
				end;
			`,
			expectedType: "*ast.Parameter",
			description:  "Parameters should be visited as Nodes",
		},
		{
			name: "CaseBranch in case statement",
			code: `
				var x: Integer := 1;
				case x of
					1: PrintLn('one');
					2, 3: PrintLn('two or three');
				end;
			`,
			expectedType: "*ast.CaseBranch",
			description:  "CaseBranch should be visited as a Node",
		},
		{
			name: "ExceptionHandler in try statement",
			code: `
				try
					raise Exception.Create('error');
				except
					on E: Exception do
						PrintLn(E.Message);
				end;
			`,
			expectedType: "*ast.ExceptionHandler",
			description:  "ExceptionHandler should be visited as a Node",
		},
		{
			name: "ExceptClause in try statement",
			code: `
				try
					DoSomething();
				except
					on E: Exception do
						PrintLn('error');
				end;
			`,
			expectedType: "*ast.ExceptClause",
			description:  "ExceptClause should be visited as a Node",
		},
		{
			name: "FinallyClause in try statement",
			code: `
				try
					DoSomething();
				finally
					Cleanup();
				end;
			`,
			expectedType: "*ast.FinallyClause",
			description:  "FinallyClause should be visited as a Node",
		},
		{
			name: "FieldInitializer in record literal",
			code: `
				type TPoint = record
					x: Integer;
					y: Integer;
				end;
				var p := TPoint(x: 10; y: 20);
			`,
			expectedType: "*ast.FieldInitializer",
			description:  "FieldInitializer should be visited as a Node",
		},
		{
			name: "InterfaceMethodDecl in interface",
			code: `
				type IMyInterface = interface
					procedure DoSomething;
					function GetValue: Integer;
				end;
			`,
			expectedType: "*ast.InterfaceMethodDecl",
			description:  "InterfaceMethodDecl should be visited as a Node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, _ := dwscript.New()
			program, err := engine.Parse(tt.code)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Track if we found the expected type
			found := false
			ast.Inspect(program, func(n ast.Node) bool {
				if n == nil {
					return true
				}
				typeName := fmt.Sprintf("%T", n)
				if typeName == tt.expectedType {
					found = true

					// Verify the node implements the required methods
					// All nodes must have these methods
					_ = n.TokenLiteral()
					_ = n.String()
					_ = n.Pos()
					_ = n.End()

					return false // Found it, can stop
				}
				return true
			})

			if !found {
				t.Errorf("%s: Expected to find %s node, but it was not visited",
					tt.description, tt.expectedType)
			}
		})
	}
}

