package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/spf13/cobra"
)

var (
	evalExpr     string
	dumpAST      bool
	trace        bool
	typeCheck    bool
	showUnits    bool
	maxRecursion int
	bytecodeMode bool
)

// simpleOptions implements interp.Options for the CLI.
type simpleOptions struct {
	MaxRecursionDepth int
}

func (o *simpleOptions) GetExternalFunctions() *interp.ExternalFunctionRegistry {
	return nil // CLI doesn't use external functions
}

func (o *simpleOptions) GetMaxRecursionDepth() int {
	return o.MaxRecursionDepth
}

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
  dwscript run --trace script.dws

  # Run with custom recursion limit
  dwscript run --max-recursion 2048 script.dws`,
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
	runCmd.Flags().IntVar(&maxRecursion, "max-recursion", 1024, "maximum recursion depth (default: 1024)")
	runCmd.Flags().BoolVar(&bytecodeMode, "bytecode", false, "execute via bytecode VM instead of AST interpreter (experimental)")
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

		// Check if this is a precompiled bytecode file
		if filepath.Ext(filename) == ".dwc" {
			return runBytecodeFile(filename)
		}

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
		// Convert ParserError to CompilerError format with pretty output
		compilerErrors := make([]*errors.CompilerError, 0, len(p.Errors()))
		for _, perr := range p.Errors() {
			compilerErrors = append(compilerErrors, errors.NewCompilerError(
				perr.Pos,
				perr.Message,
				input,
				filename,
			))
		}
		fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
		fmt.Fprintln(os.Stderr) // Add newline
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Check if the program uses any units
	usedUnits := extractUsedUnits(program)
	hasUnits := len(usedUnits) > 0

	// Prepare unit search paths (shared by interpreter + bytecode modes)
	searchPaths := append([]string{}, unitSearchPaths...)
	if len(searchPaths) == 0 && filename != "<eval>" {
		searchPaths = append(searchPaths, filepath.Dir(filename))
	}

	var unitRegistry *units.UnitRegistry
	compiledProgram := program
	if bytecodeMode {
		var err error
		compiledProgram, unitRegistry, err = buildBytecodeProgram(program, usedUnits, searchPaths)
		if err != nil {
			return fmt.Errorf("failed to prepare bytecode program: %w", err)
		}
	}

	// Run semantic analysis if type checking is enabled
	// Skip type checking if units are used, since symbols from units
	// aren't available until runtime
	var semanticInfo *ast.SemanticInfo
	if typeCheck && !hasUnits {
		analyzer := semantic.NewAnalyzer()
		// Set source code for rich error messages
		analyzer.SetSource(input, filename)

		if err := analyzer.Analyze(program); err != nil {
			// Use structured errors if available, fall back to string errors
			var compilerErrors []*errors.CompilerError
			if len(analyzer.StructuredErrors()) > 0 {
				// Convert structured errors directly to CompilerError
				for _, semErr := range analyzer.StructuredErrors() {
					compilerErrors = append(compilerErrors, semErr.ToCompilerError(input, filename))
				}
			} else {
				// Fall back to string error conversion for backward compatibility
				compilerErrors = errors.FromStringErrors(analyzer.Errors(), input, filename)
			}

			fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
			fmt.Fprintln(os.Stderr) // Add newline
			return fmt.Errorf("semantic analysis failed with %d error(s)", len(analyzer.Errors()))
		}
		// Capture semantic info to pass to interpreter
		semanticInfo = analyzer.GetSemanticInfo()
	} else if verbose && hasUnits {
		fmt.Fprintf(os.Stderr, "Type checking disabled (program uses units)\n")
	}

	// Dump AST if requested
	if dumpAST {
		fmt.Println("AST:")
		fmt.Println(compiledProgram.String())
		fmt.Println()
	}

	if bytecodeMode {
		if showUnits && unitRegistry != nil && len(usedUnits) > 0 {
			displayUnitDependencyTree(unitRegistry, usedUnits)
		}
		return executeBytecodeProgram(compiledProgram, bytecodeExecOptions{
			filename: filename,
			trace:    trace,
		})
	}

	// Interpreter: execute the program
	// Create a simple options struct for passing maxRecursionDepth
	opts := &simpleOptions{
		MaxRecursionDepth: maxRecursion,
	}
	interpreter := interp.NewWithOptions(os.Stdout, opts)

	// Set source code for enhanced runtime error messages
	interpreter.SetSource(input, filename)

	// Pass semantic info to interpreter if available (enables type inference for empty arrays)
	if semanticInfo != nil {
		interpreter.SetSemanticInfo(semanticInfo)
	}

	// Set up unit registry if search paths are provided or if we're running from a file
	if len(searchPaths) > 0 {
		registry := units.NewUnitRegistry(searchPaths)
		interpreter.SetUnitRegistry(registry)

		// Check if the program uses any units and load them
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

	// Check for unhandled exceptions
	if exc := interpreter.GetException(); exc != nil {
		// Format and print unhandled exception with position (if available) and stack trace
		// DWScript format: "Runtime Error: <Message> [line: N, column: M]"
		if exc.Position != nil {
			fmt.Fprintf(os.Stderr, "Runtime Error: %s: %s [line: %d, column: %d]\n",
				exc.ClassInfo.Name, exc.Message, exc.Position.Line, exc.Position.Column)
		} else {
			// If no position (e.g., internal errors), use simple format
			fmt.Fprintf(os.Stderr, "Runtime Error: %s: %s\n", exc.ClassInfo.Name, exc.Message)
		}

		// Print stack trace if available
		// The StackTrace.String() method formats each frame with position info
		if len(exc.CallStack) > 0 {
			fmt.Fprint(os.Stderr, exc.CallStack.String())
			fmt.Fprintln(os.Stderr) // Add final newline
		}

		return fmt.Errorf("unhandled exception: %s", exc.Message)
	}

	// Check for runtime errors
	if result != nil && result.Type() == "ERROR" {
		// Check if it's a structured RuntimeError with rich formatting
		if runtimeErr, ok := result.(*interp.RuntimeError); ok {
			if compilerErr := runtimeErr.ToCompilerError(); compilerErr != nil {
				// Use rich error formatting with source snippet
				fmt.Fprint(os.Stderr, compilerErr.Format(true))
				fmt.Fprintln(os.Stderr)
				return fmt.Errorf("execution failed")
			}
		}
		// Fall back to simple error display for non-structured errors
		fmt.Fprintf(os.Stderr, "Runtime error: %s\n", result.String())
		return fmt.Errorf("execution failed")
	}

	return nil
}

// runBytecodeFile loads and executes a precompiled bytecode file (.dwc)
func runBytecodeFile(filename string) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Loading precompiled bytecode from %s...\n", filename)
	}

	// Read the bytecode file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read bytecode file %s: %w", filename, err)
	}

	// Deserialize the bytecode
	serializer := bytecode.NewSerializer()
	chunk, err := serializer.DeserializeChunk(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize bytecode: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Loaded bytecode:\n")
		fmt.Fprintf(os.Stderr, "  Name: %s\n", chunk.Name)
		fmt.Fprintf(os.Stderr, "  Instructions: %d\n", len(chunk.Code))
		fmt.Fprintf(os.Stderr, "  Constants: %d\n", len(chunk.Constants))
		fmt.Fprintf(os.Stderr, "  Locals: %d\n", chunk.LocalCount)
	}

	// Show disassembly if trace is enabled
	if trace {
		fmt.Fprintf(os.Stderr, "\n== Bytecode Trace (%s) ==\n", chunk.Name)
		bytecode.NewDisassembler(chunk, os.Stderr).Disassemble()
	}

	// Execute the bytecode
	vm := bytecode.NewVMWithOutput(os.Stdout)
	result, err := vm.Run(chunk)
	if err != nil {
		if runtimeErr, ok := err.(*bytecode.RuntimeError); ok {
			fmt.Fprintf(os.Stderr, "Bytecode runtime error: %s\n", runtimeErr.Message)
			if len(runtimeErr.Trace) > 0 {
				fmt.Fprint(os.Stderr, runtimeErr.Trace.String())
				fmt.Fprintln(os.Stderr)
			}
			return fmt.Errorf("bytecode execution failed: %w", runtimeErr)
		}
		return fmt.Errorf("bytecode execution failed: %w", err)
	}

	if verbose && !result.IsNil() {
		fmt.Fprintf(os.Stderr, "Bytecode result: %s\n", result.String())
	}

	return nil
}

type bytecodeExecOptions struct {
	filename string
	trace    bool
}

func executeBytecodeProgram(program *ast.Program, opts bytecodeExecOptions) error {
	if program == nil {
		return fmt.Errorf("bytecode execution: nil program")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Bytecode mode enabled - executing %s]\n", opts.filename)
	}

	compiler := bytecode.NewCompiler(opts.filename)
	chunk, err := compiler.Compile(program)
	if err != nil {
		return fmt.Errorf("bytecode compilation failed: %w", err)
	}

	if opts.trace {
		fmt.Fprintf(os.Stderr, "\n== Bytecode Trace (%s) ==\n", chunk.Name)
		bytecode.NewDisassembler(chunk, os.Stderr).Disassemble()
	}

	vm := bytecode.NewVMWithOutput(os.Stdout)
	result, err := vm.Run(chunk)
	if err != nil {
		if runtimeErr, ok := err.(*bytecode.RuntimeError); ok {
			fmt.Fprintf(os.Stderr, "Bytecode runtime error: %s\n", runtimeErr.Message)
			if len(runtimeErr.Trace) > 0 {
				fmt.Fprint(os.Stderr, runtimeErr.Trace.String())
				fmt.Fprintln(os.Stderr)
			}
			return fmt.Errorf("bytecode execution failed: %w", runtimeErr)
		}
		return fmt.Errorf("bytecode execution failed: %w", err)
	}

	if verbose && !result.IsNil() {
		fmt.Fprintf(os.Stderr, "Bytecode result: %s\n", result.String())
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

func buildBytecodeProgram(program *ast.Program, usedUnits []string, searchPaths []string) (*ast.Program, *units.UnitRegistry, error) {
	if program == nil {
		return nil, nil, fmt.Errorf("bytecode: nil program")
	}

	filteredMain := filterOutUses(program.Statements)
	if len(usedUnits) == 0 {
		if len(filteredMain) == len(program.Statements) {
			return program, nil, nil
		}
		return &ast.Program{Statements: filteredMain}, nil, nil
	}

	registry := units.NewUnitRegistry(searchPaths)
	for _, unitName := range usedUnits {
		if _, err := registry.LoadUnit(unitName, searchPaths); err != nil {
			return nil, nil, fmt.Errorf("failed to load unit '%s': %w", unitName, err)
		}
	}

	order, err := registry.ComputeInitializationOrder()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compute unit initialization order: %w", err)
	}

	combined := &ast.Program{}
	combined.Statements = append(combined.Statements, collectUnitImplementation(order, registry)...)
	combined.Statements = append(combined.Statements, collectUnitInitialization(order, registry)...)
	combined.Statements = append(combined.Statements, filteredMain...)
	combined.Statements = append(combined.Statements, collectUnitFinalization(order, registry)...)

	return combined, registry, nil
}

func collectUnitImplementation(order []string, registry *units.UnitRegistry) []ast.Statement {
	var stmts []ast.Statement
	for _, name := range order {
		unit, ok := registry.GetUnit(name)
		if !ok || unit == nil {
			continue
		}
		stmts = append(stmts, blockStatements(unit.ImplementationSection)...)
	}
	return stmts
}

func collectUnitInitialization(order []string, registry *units.UnitRegistry) []ast.Statement {
	var stmts []ast.Statement
	for _, name := range order {
		unit, ok := registry.GetUnit(name)
		if !ok || unit == nil {
			continue
		}
		stmts = append(stmts, blockStatements(unit.InitializationSection)...)
	}
	return stmts
}

func collectUnitFinalization(order []string, registry *units.UnitRegistry) []ast.Statement {
	var stmts []ast.Statement
	for i := len(order) - 1; i >= 0; i-- {
		unit, ok := registry.GetUnit(order[i])
		if !ok || unit == nil {
			continue
		}
		stmts = append(stmts, blockStatements(unit.FinalizationSection)...)
	}
	return stmts
}

func blockStatements(block *ast.BlockStatement) []ast.Statement {
	if block == nil {
		return nil
	}
	return filterOutUses(block.Statements)
}

func filterOutUses(stmts []ast.Statement) []ast.Statement {
	if len(stmts) == 0 {
		return nil
	}
	filtered := make([]ast.Statement, 0, len(stmts))
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		if _, ok := stmt.(*ast.UsesClause); ok {
			continue
		}
		filtered = append(filtered, stmt)
	}
	return filtered
}
