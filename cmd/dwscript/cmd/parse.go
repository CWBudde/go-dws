package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/lexer"
	"github.com/cwbudde/go-dws/parser"
	"github.com/spf13/cobra"
)

var (
	parseExpression bool
	parseDumpAST    bool
)

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse DWScript source code and display the AST",
	Long: `Parse DWScript source code and display the Abstract Syntax Tree (AST).

If no file is provided, reads from stdin.
Use -e to parse a single expression from the command line.
Use --dump-ast to show the full AST structure.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().BoolVarP(&parseExpression, "expression", "e", false, "parse an expression from the command line")
	parseCmd.Flags().BoolVar(&parseDumpAST, "dump-ast", false, "dump the full AST structure")
}

func runParse(cmd *cobra.Command, args []string) error {
	var input string

	// Determine input source
	if parseExpression {
		if len(args) == 0 {
			return fmt.Errorf("no expression provided")
		}
		input = args[0]
	} else if len(args) > 0 {
		// Read from file
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		input = string(data)
	} else {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading stdin: %w", err)
		}
		input = string(data)
	}

	// Parse the input
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for errors
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors:\n")
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", msg)
		}
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Display output
	if parseDumpAST {
		fmt.Println("Abstract Syntax Tree:")
		fmt.Println("=====================")
		dumpASTNode(program, 0)
	} else {
		fmt.Println(program.String())
	}

	return nil
}

func dumpASTNode(node any, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram (%d statements)\n", indentStr, len(n.Statements))
		for _, stmt := range n.Statements {
			dumpASTNode(stmt, indent+1)
		}
	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement\n", indentStr)
		dumpASTNode(n.Expression, indent+1)
	case *ast.BlockStatement:
		fmt.Printf("%sBlockStatement (%d statements)\n", indentStr, len(n.Statements))
		for _, stmt := range n.Statements {
			dumpASTNode(stmt, indent+1)
		}
	case *ast.BinaryExpression:
		fmt.Printf("%sBinaryExpression (%s)\n", indentStr, n.Operator)
		fmt.Printf("%s  Left:\n", indentStr)
		dumpASTNode(n.Left, indent+2)
		fmt.Printf("%s  Right:\n", indentStr)
		dumpASTNode(n.Right, indent+2)
	case *ast.UnaryExpression:
		fmt.Printf("%sUnaryExpression (%s)\n", indentStr, n.Operator)
		dumpASTNode(n.Right, indent+1)
	case *ast.IntegerLiteral:
		fmt.Printf("%sIntegerLiteral: %d\n", indentStr, n.Value)
	case *ast.FloatLiteral:
		fmt.Printf("%sFloatLiteral: %g\n", indentStr, n.Value)
	case *ast.StringLiteral:
		fmt.Printf("%sStringLiteral: %q\n", indentStr, n.Value)
	case *ast.BooleanLiteral:
		fmt.Printf("%sBooleanLiteral: %v\n", indentStr, n.Value)
	case *ast.Identifier:
		fmt.Printf("%sIdentifier: %s\n", indentStr, n.Value)
	case *ast.NilLiteral:
		fmt.Printf("%sNilLiteral\n", indentStr)
	default:
		// For other nodes, just print their string representation
		fmt.Printf("%s%T: %v\n", indentStr, node, node)
	}
}
