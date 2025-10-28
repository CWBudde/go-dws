package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/spf13/cobra"
)

var (
	evalExpr  string
	dumpAST   bool
	trace     bool
	typeCheck bool
	showUnits bool
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
	runCmd.Flags().BoolVar(&showUnits, "show-units", false, "display unit dependency tree")
}

func runScript(_ *cobra.Command, args []string) error {
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

	// Check if the program uses any units
	usedUnits := extractUsedUnits(program)
	hasUnits := len(usedUnits) > 0

	// Run semantic analysis if type checking is enabled
	// Skip type checking if units are used, since symbols from units
	// aren't available until runtime
	if typeCheck && !hasUnits {
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
	} else if verbose && hasUnits {
		fmt.Fprintf(os.Stderr, "Type checking disabled (program uses units)\n")
	}

	// Dump AST if requested
	if dumpAST {
		fmt.Println("AST:")
		fmt.Println(program.String())
		fmt.Println()
	}

	// Interpreter: execute the program
	interpreter := interp.New(os.Stdout)

	// Set up unit registry if search paths are provided or if we're running from a file
	// Task 9.139: Add unit search path support
	searchPaths := unitSearchPaths
	if len(searchPaths) == 0 && filename != "<eval>" {
		// Add the directory of the script file as a default search path
		dir := filepath.Dir(filename)
		searchPaths = append(searchPaths, dir)
	}

	if len(searchPaths) > 0 {
		registry := units.NewUnitRegistry(searchPaths)
		interpreter.SetUnitRegistry(registry)

		// Check if the program uses any units and load them
		usedUnits := extractUsedUnits(program)
		if len(usedUnits) > 0 {
			if verbose {
				fmt.Fprintf(os.Stderr, "Loading %d unit(s)...\n", len(usedUnits))
			}

			for _, unitName := range usedUnits {
				unit, err := interpreter.LoadUnit(unitName, nil)
				if err != nil {
					return fmt.Errorf("failed to load unit '%s': %w", unitName, err)
				}

				// Import unit symbols
				if err := interpreter.ImportUnitSymbols(unit); err != nil {
					return fmt.Errorf("failed to import symbols from unit '%s': %w", unitName, err)
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "  ✓ Loaded unit: %s\n", unitName)
				}
			}

			// Initialize all loaded units
			if err := interpreter.InitializeUnits(); err != nil {
				return fmt.Errorf("failed to initialize units: %w", err)
			}

			// Display dependency order if verbose
			if verbose {
				loadedUnits := interpreter.ListLoadedUnits()
				if len(loadedUnits) > 0 {
					fmt.Fprintf(os.Stderr, "Unit initialization order: %v\n", loadedUnits)
				}
			}

			// Display unit dependency tree if requested
			if showUnits {
				displayUnitDependencyTree(interpreter.GetUnitRegistry(), usedUnits)
			}

			// Ensure units are finalized on exit
			defer func() {
				if err := interpreter.FinalizeUnits(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: error during unit finalization: %v\n", err)
				}
			}()
		}
	}

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

// extractUsedUnits extracts unit names from uses clauses in the program
func extractUsedUnits(program *ast.Program) []string {
	var usedUnits []string
	seen := make(map[string]bool)

	for _, stmt := range program.Statements {
		if usesClause, ok := stmt.(*ast.UsesClause); ok {
			for _, unitIdent := range usesClause.Units {
				if !seen[unitIdent.Value] {
					usedUnits = append(usedUnits, unitIdent.Value)
					seen[unitIdent.Value] = true
				}
			}
		}
	}

	return usedUnits
}

// displayUnitDependencyTree displays a tree view of loaded units and their dependencies.
// Shows which units are loaded and what other units they depend on (via 'uses' clauses).
func displayUnitDependencyTree(registry *units.UnitRegistry, rootUnits []string) {
	if registry == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "\n=== Unit Dependency Tree ===\n")

	// Track which units have been displayed to avoid duplicates
	displayed := make(map[string]bool)

	// Display each root unit and its dependencies
	for _, unitName := range rootUnits {
		displayUnitAndDependencies(registry, unitName, "", displayed, true)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

// displayUnitAndDependencies recursively displays a unit and its dependencies.
func displayUnitAndDependencies(registry *units.UnitRegistry, unitName string, prefix string, displayed map[string]bool, isLast bool) {
	// Get the unit
	unit, ok := registry.GetUnit(unitName)
	if !ok {
		return
	}

	// Determine the tree characters
	var connector, nextPrefix string
	if prefix == "" {
		// Root level
		connector = ""
		nextPrefix = "  "
	} else if isLast {
		connector = "└─ "
		nextPrefix = prefix + "   "
	} else {
		connector = "├─ "
		nextPrefix = prefix + "│  "
	}

	// Display the unit name
	if displayed[unitName] {
		// Already displayed, just show reference
		fmt.Fprintf(os.Stderr, "%s%s%s (see above)\n", prefix, connector, unitName)
		return
	}

	fmt.Fprintf(os.Stderr, "%s%s%s\n", prefix, connector, unitName)
	displayed[unitName] = true

	// Display dependencies
	if len(unit.Uses) > 0 {
		for i, depName := range unit.Uses {
			isLastDep := i == len(unit.Uses)-1
			displayUnitAndDependencies(registry, depName, nextPrefix, displayed, isLastDep)
		}
	}
}
