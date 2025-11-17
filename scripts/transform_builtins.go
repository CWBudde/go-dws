// transform_builtins.go - Helper script to transform builtin methods to functions
//
// This script transforms builtin methods from:
//   func (i *Interpreter) builtinXxx(args []Value) Value
// to:
//   func Xxx(ctx Context, args []Value) Value
//
// And replaces:
//   i.newErrorWithLocation(i.currentNode, ...) → ctx.NewError(...)
//   i.currentNode → ctx.CurrentNode()
//   *StringValue → *runtime.StringValue
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-file> <output-file> <category>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s builtins_math_basic.go math.go math\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]
	category := "builtins"
	if len(os.Args) > 3 {
		category = os.Args[3]
	}

	// Read input file
	input, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer input.Close()

	// Collect all lines
	var lines []string
	scanner := bufio.Scanner(input)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Transform lines
	transformed := transformLines(lines, category)

	// Write output
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	for _, line := range transformed {
		fmt.Fprintln(output, line)
	}

	fmt.Printf("Transformed %s → %s\n", inputFile, outputFile)
}

func transformLines(lines []string, category string) []string {
	var result []string

	// Track if we're in package/import section
	inFile := true
	hasImports := false

	// Regex patterns
	methodRegex := regexp.MustCompile(`^func \(i \*Interpreter\) builtin(\w+)\(args \[\]Value\) Value \{`)
	errorCallRegex := regexp.MustCompile(`i\.newErrorWithLocation\(i\.currentNode,\s*`)
	currentNodeRegex := regexp.MustCompile(`i\.currentNode`)
	builtinCallRegex := regexp.MustCompile(`i\.(builtin\w+)\(`)

	for i, line := range lines {
		// Skip package declaration from input
		if strings.HasPrefix(line, "package interp") {
			if inFile {
				inFile = false
				continue
			}
		}

		// Skip imports from input file (we'll add our own)
		if strings.HasPrefix(line, "import") || (hasImports && (strings.HasPrefix(line, "\t") || line == ")")) {
			hasImports = true
			if line == ")" {
				hasImports = false
			}
			continue
		}

		// Transform method signature: func (i *Interpreter) builtinXxx → func Xxx
		if matches := methodRegex.FindStringSubmatch(line); matches != nil {
			funcName := matches[1]
			result = append(result, fmt.Sprintf("// %s implements the %s() built-in function.", funcName, funcName))
			result = append(result, fmt.Sprintf("func %s(ctx Context, args []Value) Value {", funcName))
			continue
		}

		// Transform error calls: i.newErrorWithLocation(i.currentNode, → ctx.NewError(
		line = errorCallRegex.ReplaceAllString(line, "ctx.NewError(")

		// Transform currentNode references: i.currentNode → ctx.CurrentNode()
		if !strings.Contains(line, "ctx.NewError") { // Don't double-transform
			line = currentNodeRegex.ReplaceAllString(line, "ctx.CurrentNode()")
		}

		// Transform builtin calls: i.builtinXxx( → Xxx(ctx,
		line = builtinCallRegex.ReplaceAllString(line, "$1(ctx, ")

		result = append(result, line)
	}

	return result
}
