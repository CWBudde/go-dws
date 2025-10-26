package cmd

import (
	"fmt"
	"os"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/spf13/cobra"
)

var (
	evalExpr  string
	dumpAST   bool
	trace     bool
	typeCheck bool
)

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run a DWScript file or expression",
	Long: `Execute a DWScript program from a file or inline expression.

Examples:
  # Run a script file
  dwscript run script.dws

  # Evaluate an inline expression
  dwscript run -e "PrintLn('Hello, World!');"

  # Run with AST dump (for debugging)
  dwscript run --dump-ast script.dws

  # Run with execution trace
  dwscript run --trace script.dws`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScript,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&evalExpr, "eval", "e", "", "evaluate inline code instead of reading from file")
	runCmd.Flags().BoolVar(&dumpAST, "dump-ast", false, "dump the parsed AST (for debugging)")
	runCmd.Flags().BoolVar(&trace, "trace", false, "trace execution (for debugging)")
	runCmd.Flags().BoolVar(&typeCheck, "type-check", true, "perform semantic type checking before execution (default: true)")
}

func runScript(cmd *cobra.Command, args []string) error {
	var input string
	var filename string

	// Determine input source
	if evalExpr != "" {
		// Inline expression provided
		input = evalExpr
		filename = "<eval>"
	} else if len(args) == 1 {
		// File path provided
		filename = args[0]
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		input = string(content)
	} else {
		return fmt.Errorf("either provide a file path or use -e flag for inline code")
	}

	// Lexer: tokenize the input
	l := lexer.New(input)

	// Parser: build the AST
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		// Convert string errors to CompilerError format with pretty output
		compilerErrors := errors.FromStringErrors(p.Errors(), input, filename)
		fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
		fmt.Fprintln(os.Stderr) // Add newline
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Run semantic analysis if type checking is enabled
	if typeCheck {
		analyzer := semantic.NewAnalyzer()
		if err := analyzer.Analyze(program); err != nil {
			// Set errors on parser for compatibility
			p.SetSemanticErrors(analyzer.Errors())

			// Convert string errors to CompilerError format with pretty output
			compilerErrors := errors.FromStringErrors(analyzer.Errors(), input, filename)
			fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
			fmt.Fprintln(os.Stderr) // Add newline
			return fmt.Errorf("semantic analysis failed with %d error(s)", len(analyzer.Errors()))
		}
	}

	// Dump AST if requested
	if dumpAST {
		fmt.Println("AST:")
		fmt.Println(program.String())
		fmt.Println()
	}

	// Interpreter: execute the program
	interpreter := interp.New(os.Stdout)

	if trace {
		fmt.Fprintf(os.Stderr, "[Trace mode enabled - executing %s]\n", filename)
	}

	result := interpreter.Eval(program)

	// Check for runtime errors
	if result != nil && result.Type() == "ERROR" {
		fmt.Fprintf(os.Stderr, "Runtime error: %s\n", result.String())
		return fmt.Errorf("execution failed")
	}

	return nil
}
