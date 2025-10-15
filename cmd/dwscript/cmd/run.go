package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	evalExpr string
	dumpAST  bool
	trace    bool
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

	// TODO: Implement the actual compilation and execution pipeline
	// For now, just echo the input as a placeholder
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Processing: %s\n", filename)
		fmt.Printf("Input length: %d bytes\n", len(input))
	}

	fmt.Println("⚠️  Compiler not yet implemented - Stage 0 in progress")
	fmt.Printf("\nInput received:\n%s\n", input)

	if dumpAST {
		fmt.Println("\n[AST dump not yet available]")
	}

	if trace {
		fmt.Println("\n[Execution trace not yet available]")
	}

	return nil
}
