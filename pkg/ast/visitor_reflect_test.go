package ast_test

import (
	"fmt"
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// Note: Test helper types countingVisitor and stopAfterFirstVisitor are
// defined in visitor_test.go and shared across all ast_test package tests.

// TestWalkReflect_VisitsAllNodes tests that WalkReflect visits all nodes in the tree
func TestWalkReflect_VisitsAllNodes(t *testing.T) {
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

	// Count nodes with reflection-based visitor
	reflectCount := 0
	reflectVisitor := &countingVisitor{count: &reflectCount}
	ast.WalkReflect(reflectVisitor, tree)

	// Count nodes with manual visitor for comparison
	manualCount := 0
	manualVisitor := &countingVisitor{count: &manualCount}
	ast.Walk(manualVisitor, tree)

	// Should visit the same number of nodes
	if reflectCount != manualCount {
		t.Errorf("Node count mismatch: WalkReflect=%d, Walk=%d", reflectCount, manualCount)
	}

	// Should visit many nodes (program, statements, expressions, etc.)
	if reflectCount == 0 {
		t.Error("WalkReflect() did not visit any nodes")
	}

	if reflectCount < 10 {
		t.Errorf("Expected at least 10 nodes visited, got %d", reflectCount)
	}
}

// TestWalkReflect_VisitorReturnsNil tests that returning nil from Visit stops traversal
func TestWalkReflect_VisitorReturnsNil(t *testing.T) {
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
	reflectCount := 0
	reflectVisitor := &stopAfterFirstVisitor{count: &reflectCount}
	ast.WalkReflect(reflectVisitor, tree)

	// Should only visit the root Program node
	if reflectCount != 1 {
		t.Errorf("Expected 1 node visited (root only), got %d", reflectCount)
	}
}

// TestInspectReflect_FindsFunctions tests using InspectReflect to find function declarations
func TestInspectReflect_FindsFunctions(t *testing.T) {
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

	// Find all function declarations using reflection
	reflectFunctions := []string{}
	ast.InspectReflect(tree, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FunctionDecl); ok {
			reflectFunctions = append(reflectFunctions, funcDecl.Name.Value)
		}
		return true
	})

	// Find using manual visitor for comparison
	manualFunctions := []string{}
	ast.Inspect(tree, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FunctionDecl); ok {
			manualFunctions = append(manualFunctions, funcDecl.Name.Value)
		}
		return true
	})

	if len(reflectFunctions) != len(manualFunctions) {
		t.Errorf("Function count mismatch: reflect=%d, manual=%d",
			len(reflectFunctions), len(manualFunctions))
	}

	if len(reflectFunctions) != 2 {
		t.Errorf("Expected 2 functions, found %d", len(reflectFunctions))
	}

	expected := map[string]bool{"Add": true, "Multiply": true}
	for _, name := range reflectFunctions {
		if !expected[name] {
			t.Errorf("Unexpected function name: %s", name)
		}
	}
}

// TestWalkReflect_CompareWithManualWalk tests that reflection and manual walkers visit the same nodes
func TestWalkReflect_CompareWithManualWalk(t *testing.T) {
	testCases := []struct {
		name   string
		source string
	}{
		{
			name: "Simple variables",
			source: `
				var x: Integer := 1;
				var y: String := 'test';
			`,
		},
		{
			name: "Binary expressions",
			source: `
				var result: Integer := (a + b) * (c - d) / e;
			`,
		},
		{
			name: "Control flow",
			source: `
				if x > 0 then
					y := 1
				else
					y := 0;

				while x > 0 do
					x := x - 1;

				for var i := 0 to 10 do
					PrintLn(i);
			`,
		},
		{
			name: "Functions",
			source: `
				function Test(a: Integer): Integer;
				begin
					Result := a * 2;
				end;
			`,
		},
		{
			name: "Arrays",
			source: `
				var arr: array of Integer := [1, 2, 3, 4, 5];
			`,
		},
		{
			name: "Case statement",
			source: `
				case x of
					1: y := 'one';
					2: y := 'two';
					3..5: y := 'three to five';
				end;
			`,
		},
		{
			name: "Try-except",
			source: `
				try
					DoSomething();
				except
					on E: Exception do
						PrintLn('error');
				end;
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine, _ := dwscript.New()
			program, err := engine.Parse(tc.source)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			tree := program

			// Collect node types with reflection walker
			reflectTypes := make(map[string]int)
			ast.InspectReflect(tree, func(n ast.Node) bool {
				if n != nil {
					reflectTypes[fmt.Sprintf("%T", n)]++
				}
				return true
			})

			// Collect node types with manual walker
			manualTypes := make(map[string]int)
			ast.Inspect(tree, func(n ast.Node) bool {
				if n != nil {
					manualTypes[fmt.Sprintf("%T", n)]++
				}
				return true
			})

			// Compare counts
			if len(reflectTypes) != len(manualTypes) {
				t.Errorf("Node type count mismatch: reflect=%d, manual=%d",
					len(reflectTypes), len(manualTypes))
				t.Errorf("Reflect types: %v", reflectTypes)
				t.Errorf("Manual types: %v", manualTypes)
			}

			// Compare each type
			for nodeType, reflectCount := range reflectTypes {
				manualCount, exists := manualTypes[nodeType]
				if !exists {
					t.Errorf("Node type %s found by reflection but not manual walker", nodeType)
					continue
				}
				if reflectCount != manualCount {
					t.Errorf("Node type %s: reflect count=%d, manual count=%d",
						nodeType, reflectCount, manualCount)
				}
			}

			// Check for types in manual but not in reflect
			for nodeType := range manualTypes {
				if _, exists := reflectTypes[nodeType]; !exists {
					t.Errorf("Node type %s found by manual walker but not reflection", nodeType)
				}
			}
		})
	}
}

// TestWalkReflect_WithNilNodes tests that WalkReflect handles nil nodes gracefully
func TestWalkReflect_WithNilNodes(t *testing.T) {
	// Create a simple program with nil child nodes
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
	ast.InspectReflect(program, func(n ast.Node) bool {
		if n != nil {
			nodeCount++
		}
		return true
	})

	if nodeCount == 0 {
		t.Error("Expected to visit some nodes")
	}
}

// TestWalkReflect_ComplexProgram tests reflection walker on a complex program
func TestWalkReflect_ComplexProgram(t *testing.T) {
	engine, _ := dwscript.New()

	// Complex program with many node types
	program, _ := engine.Parse(`
		type TMyEnum = (Value1, Value2, Value3);

		type TMyRecord = record
			Field1: Integer;
			Field2: String;
		end;

		var arr: array of Integer;
		var e: TMyEnum;

		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

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

	// Count nodes with both walkers
	reflectCount := 0
	ast.InspectReflect(tree, func(n ast.Node) bool {
		if n != nil {
			reflectCount++
		}
		return true
	})

	manualCount := 0
	ast.Inspect(tree, func(n ast.Node) bool {
		if n != nil {
			manualCount++
		}
		return true
	})

	if reflectCount != manualCount {
		t.Errorf("Node count mismatch in complex program: reflect=%d, manual=%d",
			reflectCount, manualCount)
	}

	// Should visit many nodes
	if reflectCount < 50 {
		t.Errorf("Expected at least 50 nodes in complex program, got %d", reflectCount)
	}
}

// BenchmarkWalk_Manual benchmarks the manual type-switch based walker
func BenchmarkWalk_Manual(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(benchmarkSource)
	tree := program

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		ast.Inspect(tree, func(n ast.Node) bool {
			if n != nil {
				count++
			}
			return true
		})
	}
}

// BenchmarkWalk_Reflect benchmarks the reflection-based walker
func BenchmarkWalk_Reflect(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(benchmarkSource)
	tree := program

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		ast.InspectReflect(tree, func(n ast.Node) bool {
			if n != nil {
				count++
			}
			return true
		})
	}
}

// BenchmarkWalk_SimpleProgram compares performance on a simple program
func BenchmarkWalk_SimpleProgram(b *testing.B) {
	source := `
		var x: Integer := 42;
		var y: String := 'hello';
	`

	engine, _ := dwscript.New()
	program, _ := engine.Parse(source)
	tree := program

	b.Run("Manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			ast.Inspect(tree, func(n ast.Node) bool {
				count++
				return true
			})
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			ast.InspectReflect(tree, func(n ast.Node) bool {
				count++
				return true
			})
		}
	})
}

// BenchmarkWalk_ComplexProgram compares performance on a complex program
func BenchmarkWalk_ComplexProgram(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(benchmarkSource)
	tree := program

	b.Run("Manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			ast.Inspect(tree, func(n ast.Node) bool {
				count++
				return true
			})
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			ast.InspectReflect(tree, func(n ast.Node) bool {
				count++
				return true
			})
		}
	})
}

const benchmarkSource = `
	type TMyEnum = (Value1, Value2, Value3);

	type TMyRecord = record
		Field1: Integer;
		Field2: String;
		Field3: Float;
	end;

	var globalVar: Integer := 100;
	var globalStr: String := 'test';

	function Fibonacci(n: Integer): Integer;
	begin
		if n <= 1 then
			Result := n
		else
			Result := Fibonacci(n - 1) + Fibonacci(n - 2);
	end;

	function ProcessArray(arr: array of Integer): Integer;
	var
		sum: Integer;
		i: Integer;
	begin
		sum := 0;
		for i := 0 to High(arr) do
			sum := sum + arr[i];
		Result := sum;
	end;

	function TestCase(value: Integer): String;
	begin
		case value of
			1: Result := 'one';
			2: Result := 'two';
			3..5: Result := 'three to five';
			6, 7, 8: Result := 'six to eight';
		else
			Result := 'other';
		end;
	end;

	var arr: array of Integer;
	var rec: TMyRecord;
	var e: TMyEnum;
	var result: Integer;

	begin
		arr := [1, 2, 3, 4, 5];
		result := ProcessArray(arr);

		rec.Field1 := 42;
		rec.Field2 := 'hello';
		rec.Field3 := 3.14;

		e := TMyEnum.Value1;

		for var i := 0 to 10 do begin
			PrintLn(i);
			if i mod 2 = 0 then
				PrintLn('even')
			else
				PrintLn('odd');
		end;

		while result > 0 do begin
			result := result - 1;
			if result = 5 then
				break;
		end;

		try
			result := Fibonacci(10);
			PrintLn(result);
		except
			on E: Exception do
				PrintLn('Error: ' + E.Message);
		end;

		result := result + (10 * 20 div 3) mod 7;
	end
`
