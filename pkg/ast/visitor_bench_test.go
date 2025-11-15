package ast_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// Benchmark simple program traversal
func BenchmarkVisitorGenerated_SimpleProgram(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 42;
		var y: String := 'hello';
	`)

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

// Benchmark complex program with multiple node types
func BenchmarkVisitorGenerated_ComplexProgram(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
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

		var arr: array of Integer;
		var result: Integer;

		begin
			arr := [1, 2, 3, 4, 5];
			result := ProcessArray(arr);

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
		end
	`)

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

// Benchmark deeply nested control structures
func BenchmarkVisitorGenerated_DeepNesting(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		begin
			if x > 0 then begin
				if y > 0 then begin
					if z > 0 then begin
						if a > 0 then begin
							if b > 0 then begin
								PrintLn('deep');
							end;
						end;
					end;
				end;
			end;
		end
	`)

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

// Benchmark wide tree (many siblings)
func BenchmarkVisitorGenerated_WideTree(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		begin
			var a := 1;
			var b := 2;
			var c := 3;
			var d := 4;
			var e := 5;
			var f := 6;
			var g := 7;
			var h := 8;
			var i := 9;
			var j := 10;
			var k := 11;
			var l := 12;
			var m := 13;
			var n := 14;
			var o := 15;
			var p := 16;
			var q := 17;
			var r := 18;
			var s := 19;
			var t := 20;
		end
	`)

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

// Benchmark function declarations
func BenchmarkVisitorGenerated_FunctionDeclarations(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		function Subtract(a, b: Integer): Integer;
		begin
			Result := a - b;
		end;

		function Multiply(a, b: Integer): Integer;
		begin
			Result := a * b;
		end;

		function Divide(a, b: Integer): Float;
		begin
			Result := a / b;
		end;
	`)

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

// Benchmark with custom visitor (more realistic use case)
func BenchmarkVisitorGenerated_CustomVisitor(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		function Test(x: Integer): Integer;
		var
			result: Integer;
			i: Integer;
		begin
			result := 0;
			for i := 0 to x do
				result := result + i;
			Result := result;
		end;

		begin
			PrintLn(Test(100));
		end
	`)

	tree := program

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate a typical visitor use case: counting specific node types
		funcCount := 0
		varCount := 0
		loopCount := 0

		ast.Inspect(tree, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.FunctionDecl:
				funcCount++
			case *ast.VarDeclStatement:
				varCount++
			case *ast.ForStatement:
				loopCount++
			}
			return true
		})
	}
}

// Benchmark visitor with early termination (skip subtrees)
func BenchmarkVisitorGenerated_EarlyTermination(b *testing.B) {
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

		function Another(): Integer;
		begin
			Result := 0;
		end;
	`)

	tree := program

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Find only top-level functions (don't descend into bodies)
		count := 0
		ast.Inspect(tree, func(n ast.Node) bool {
			if _, ok := n.(*ast.FunctionDecl); ok {
				count++
				return false // Don't traverse into function body
			}
			return true
		})
	}
}

// Benchmark comparison: Inspect vs raw Walk
func BenchmarkVisitorGenerated_InspectVsWalk(b *testing.B) {
	engine, _ := dwscript.New()
	program, _ := engine.Parse(`
		var x: Integer := 42;
		var y: String := 'hello';

		function Test(): Integer;
		begin
			Result := x + 10;
		end;
	`)

	tree := program

	b.Run("Inspect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			count := 0
			ast.Inspect(tree, func(n ast.Node) bool {
				if n != nil {
					count++
				}
				return true
			})
		}
	})

	b.Run("Walk", func(b *testing.B) {
		visitor := &nodeCounterVisitor{count: new(int)}
		for i := 0; i < b.N; i++ {
			*visitor.count = 0
			ast.Walk(visitor, tree)
		}
	})
}

// Helper visitor for counting nodes
type nodeCounterVisitor struct {
	count *int
}

func (v *nodeCounterVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		*v.count++
	}
	return v
}
