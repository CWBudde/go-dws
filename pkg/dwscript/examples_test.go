package dwscript_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/cwbudde/go-dws/pkg/dwscript"
)

// Example shows basic usage of the DWScript engine.
func Example() {
	engine, err := dwscript.New()
	if err != nil {
		log.Fatal(err)
	}

	result, err := engine.Eval(`PrintLn('Hello, World!');`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(result.Output)
	// Output: Hello, World!
}

// Example_compile demonstrates compiling once and running multiple times.
func Example_compile() {
	engine, err := dwscript.New()
	if err != nil {
		log.Fatal(err)
	}

	// Compile once
	program, err := engine.Compile(`
		var greeting: String := 'Hello!';
		PrintLn(greeting);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Run multiple times
	result1, _ := engine.Run(program)
	fmt.Print(result1.Output)

	result2, _ := engine.Run(program)
	fmt.Print(result2.Output)

	// Output:
	// Hello!
	// Hello!
}

// Example_withOutput shows how to capture program output to a custom writer.
func Example_withOutput() {
	var buf bytes.Buffer

	engine, err := dwscript.New(dwscript.WithOutput(&buf))
	if err != nil {
		log.Fatal(err)
	}

	_, err = engine.Eval(`PrintLn('Captured!');`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(buf.String())
	// Output: Captured!
}

// Example_typeChecking demonstrates type checking behavior.
func Example_typeChecking() {
	engine, err := dwscript.New(dwscript.WithTypeCheck(true))
	if err != nil {
		log.Fatal(err)
	}

	_, err = engine.Eval(`
		var x: Integer := 42;
		var y: String := x;  // Type error!
	`)
	if err != nil {
		fmt.Println("Compilation failed:", err != nil)
	}

	// Output:
	// Compilation failed: true
}

// Example_arithmetic shows evaluating arithmetic expressions.
func Example_arithmetic() {
	engine, err := dwscript.New()
	if err != nil {
		log.Fatal(err)
	}

	result, err := engine.Eval(`
		var a: Integer := 10;
		var b: Integer := 32;
		PrintLn(IntToStr(a + b));
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(result.Output)
	// Output: 42
}

// Example_functions demonstrates defining and calling functions.
func Example_functions() {
	engine, err := dwscript.New()
	if err != nil {
		log.Fatal(err)
	}

	result, err := engine.Eval(`
		function Add(a, b: Integer): Integer;
		begin
			Result := a + b;
		end;

		PrintLn(IntToStr(Add(20, 22)));
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(result.Output)
	// Output: 42
}
