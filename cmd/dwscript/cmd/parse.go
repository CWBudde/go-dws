package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/spf13/cobra"
)

var (
	parseExpression bool
	parseDumpAST    bool
	parseFormat     string
)

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse DWScript source code and display the AST",
	Long: `Parse DWScript source code and display the Abstract Syntax Tree (AST).

If no file is provided, reads from stdin.
Use -e to parse a single expression from the command line.

Output formats:
  dwscript  Valid DWScript source code (default)
  tree      Hierarchical AST structure visualization
  json      JSON representation of the AST`,
	Args: cobra.MaximumNArgs(1),
	RunE: runParse,
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().BoolVarP(&parseExpression, "expression", "e", false, "parse an expression from the command line")
	parseCmd.Flags().BoolVar(&parseDumpAST, "dump-ast", false, "dump the full AST structure (deprecated: use --format=tree)")
	parseCmd.Flags().StringVar(&parseFormat, "format", "dwscript", "output format: dwscript, tree, or json")
}

func runParse(_ *cobra.Command, args []string) error {
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
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
		}
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Determine output format
	format := parseFormat
	if parseDumpAST {
		// Backward compatibility: --dump-ast flag uses tree format
		format = "tree"
	}

	// Create printer with appropriate format
	var printerOpts printer.Options
	switch format {
	case "tree":
		printerOpts = printer.Options{
			Format: printer.FormatTree,
			Style:  printer.StyleDetailed,
		}
	case "json":
		printerOpts = printer.Options{
			Format: printer.FormatJSON,
			Style:  printer.StyleDetailed,
		}
	case "dwscript":
		printerOpts = printer.Options{
			Format: printer.FormatDWScript,
			Style:  printer.StyleDetailed,
		}
	default:
		return fmt.Errorf("unknown format: %s (use dwscript, tree, or json)", format)
	}

	// Print the AST
	prnt := printer.New(printerOpts)
	output := prnt.Print(program)
	fmt.Println(output)

	return nil
}
