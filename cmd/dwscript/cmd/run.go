package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/interp"
	"github.com/cwbudde/go-dws/internal/interp/runner"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
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
	hintsLevel   string

	// diagnosticsMode selects "pretty" (default) or "plain" (DWScript wire format).
	diagnosticsMode string
	// testEnvelope wraps output in the DWScript test-harness framing.
	testEnvelope bool
	// compileOnly stops after compilation and reports its diagnostics.
	compileOnly bool
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
	runCmd.Flags().StringVar(&hintsLevel, "hints", "off", "print compiler hints/warnings to stderr, non-fatal: off|normal|strict|pedantic (pedantic includes case-mismatch hints)")
	runCmd.Flags().BoolVar(&compileOnly, "compile-only", false, "compile (parse, type-check) and report diagnostics without executing; every message is printed, hints included, in the DWScript wire format (implies --diagnostics=plain)")
	runCmd.Flags().BoolVar(&testEnvelope, "test-envelope", false, "wrap output in DWScript's test-harness 'Errors >>>>' / 'Result >>>>' framing when there are messages; buffers all program output until exit (implies --diagnostics=plain)")
	runCmd.Flags().StringVar(&diagnosticsMode, "diagnostics", "pretty", "diagnostic output style: pretty (source excerpt, colors on a terminal) or plain (DWScript wire format, one message per line)")
}

// parseHintsLevel maps the --hints flag to a semantic hint level and whether
// hint/warning output was requested at all. DWScript is case-insensitive, so
// these are purely informational — they never affect whether a program compiles.
func parseHintsLevel(s string) (semantic.HintsLevel, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "off", "none", "disabled":
		return semantic.HintsLevelDisabled, false
	case "normal", "on":
		return semantic.HintsLevelNormal, true
	case "strict":
		return semantic.HintsLevelStrict, true
	case "pedantic", "all":
		return semantic.HintsLevelPedantic, true
	default:
		return semantic.HintsLevelNormal, true
	}
}

// reportCompileFailure prints the compile diagnostics of a failed frontend result
// and returns the error the command should exit with.
func reportCompileFailure(res *frontend.Result, source, filename string) error {
	if diagnosticsMode == "plain" {
		// Wire format, all severities in emission order: exactly what the fixture
		// harness compares for the *Fail suites.
		for _, line := range res.DiagnosticStrings() {
			fmt.Fprintln(os.Stderr, line)
		}
		return ErrSilent
	}
	compilerErrors := prettyCompilerErrors(res, source, filename)
	if len(compilerErrors) == 0 {
		// Every error was filtered out of the rendered set; show what there is.
		for _, line := range res.DiagnosticStrings() {
			fmt.Fprintln(os.Stderr, line)
		}
		return fmt.Errorf("compilation failed")
	}
	fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, colorEnabled()))
	fmt.Fprintln(os.Stderr)
	return fmt.Errorf("compilation failed with %d error(s)", len(compilerErrors))
}

// prettyCompilerErrors builds the source-annotated errors for pretty mode: parse
// diagnostics from the frontend result, semantic errors from the analyzer's structured
// errors (which carry the Expected:/Got: detail lines), falling back to the rendered
// diagnostics when no structured errors exist.
func prettyCompilerErrors(res *frontend.Result, source, filename string) []*errors.CompilerError {
	var out []*errors.CompilerError
	for _, d := range res.Diagnostics {
		if d.Severity == frontend.SeverityError && d.Phase == frontend.PhaseParsing {
			out = append(out, errors.NewCompilerError(
				lexer.Position{Line: d.Line, Column: d.Column}, d.Message, source, filename))
		}
	}
	if res.Analyzer != nil && len(res.Analyzer.StructuredErrors()) > 0 {
		for _, se := range res.Analyzer.StructuredErrors() {
			out = append(out, se.ToCompilerError(source, filename))
		}
		return out
	}
	for _, d := range res.Diagnostics {
		if d.Severity == frontend.SeverityError && d.Phase != frontend.PhaseParsing {
			out = append(out, errors.NewCompilerError(
				lexer.Position{Line: d.Line, Column: d.Column}, d.Message, source, filename))
		}
	}
	return out
}

// colorEnabled reports whether pretty diagnostics may use ANSI colors: never in
// plain mode, never when NO_COLOR is set, and only when stderr is a terminal.
func colorEnabled() bool {
	if diagnosticsMode == "plain" || os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stderr.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func runScript(cmd *cobra.Command, args []string) error {
	if err := normalizeDiagnosticsMode(); err != nil {
		return err
	}
	// Precompiled bytecode files bypass the source pipeline entirely.
	if evalExpr == "" && len(args) == 1 && filepath.Ext(args[0]) == ".dwc" {
		if testEnvelope || compileOnly {
			return fmt.Errorf("--test-envelope and --compile-only are not supported for precompiled .dwc files")
		}
		return runBytecodeFile(args[0])
	}
	input, filename, err := loadRunInput(args)
	if err != nil {
		return err
	}
	if cmd != nil {
		// Argument/flag errors above showed usage; failures from here on are the
		// script's, so do not append the usage block to diagnostics.
		cmd.SilenceUsage = true
	}
	cs, done, err := compileRunInput(input, filename)
	if err != nil || done {
		return err
	}
	return executeScript(cs)
}

// normalizeDiagnosticsMode validates --diagnostics and applies the plain mode that
// the harness flags --test-envelope and --compile-only imply.
func normalizeDiagnosticsMode() error {
	if bytecodeMode && (testEnvelope || compileOnly) {
		return fmt.Errorf("--test-envelope and --compile-only are not supported with --bytecode")
	}
	if testEnvelope || compileOnly {
		diagnosticsMode = "plain"
	}
	switch strings.ToLower(strings.TrimSpace(diagnosticsMode)) {
	case "", "pretty":
		diagnosticsMode = "pretty"
	case "plain":
		diagnosticsMode = "plain"
	default:
		return fmt.Errorf("invalid --diagnostics value %q (want pretty or plain)", diagnosticsMode)
	}
	return nil
}

// loadRunInput returns the script source and the name to report it under: the -e
// expression as "<eval>", or the decoded contents of the single file argument.
func loadRunInput(args []string) (input, filename string, err error) {
	if evalExpr != "" {
		return evalExpr, "<eval>", nil
	}
	if len(args) != 1 {
		return "", "", fmt.Errorf("either provide a file path or use -e flag for inline code")
	}
	filename = args[0]
	content, err := encoding.DecodeFile(filename)
	if err != nil {
		// DecodeFile wraps the underlying cause (read failure or a
		// BOM/UTF-16 decoding error), so it stays visible in the chain.
		return "", "", fmt.Errorf("failed to load script %s: %w", filename, err)
	}
	return content, filename, nil
}

// compiledScript is what compileRunInput hands to executeScript.
type compiledScript struct {
	input, filename string
	result          *frontend.Result
	program         *ast.Program // what the analyzer and the interpreter run
	compiledProgram *ast.Program // what --dump-ast and the bytecode VM see (units spliced in)
	unitRegistry    *units.UnitRegistry
	usedUnits       []string
	searchPaths     []string
	envelopeMsgs    []string
}

// compileRunInput runs the shared frontend (the same pipeline pkg/dwscript and the
// fixture harness use) and reports its diagnostics. done is true when nothing is
// left to execute (--compile-only); a non-nil error means compilation failed.
func compileRunInput(input, filename string) (cs *compiledScript, done bool, err error) {
	// {$INCLUDE} resolves relative to the script's directory; inline -e code has no
	// include root.
	hintLevel, wantHints := parseHintsLevel(hintsLevel)
	compileOpts := frontend.Options{Filename: filename, HintsLevel: hintLevel}
	if evalExpr == "" {
		compileOpts.IncludeDir = filepath.Dir(filename)
	}

	parsed := frontend.ParseWithOptions(input, compileOpts)
	if parsed.HasSemanticBlockingDiagnosticsInPhase(frontend.PhaseParsing) {
		return nil, false, reportCompileFailure(parsed, input, filename)
	}
	cs = &compiledScript{input: input, filename: filename, program: parsed.Program}
	cs.compiledProgram = cs.program
	cs.usedUnits = extractUsedUnits(cs.program)
	hasUnits := len(cs.usedUnits) > 0

	// Unit search paths (shared by interpreter + bytecode modes)
	cs.searchPaths = append([]string{}, unitSearchPaths...)
	if len(cs.searchPaths) == 0 && filename != "<eval>" {
		cs.searchPaths = append(cs.searchPaths, filepath.Dir(filename))
	}
	// Semantic analysis is skipped for unit-using programs: neither the analyzer nor
	// the frontend resolves program-level `uses` yet (PLAN.md §3.2), so unit symbols
	// would all be reported as unknown. This is the only place that bypass lives.
	compileOpts.SkipTypeCheck = !typeCheck || hasUnits
	cs.result = frontend.AnalyzeParsed(parsed, input, compileOpts)
	if cs.result.HasFatalDiagnostics() || (cs.result.SemanticAttempted && !cs.result.SemanticSuccessful) {
		return nil, false, reportCompileFailure(cs.result, input, filename)
	}
	// The bytecode program is assembled from the monomorphized AST.
	if bytecodeMode {
		cs.compiledProgram, cs.unitRegistry, err = buildBytecodeProgram(cs.program, cs.usedUnits, cs.searchPaths)
		if err != nil {
			return nil, false, fmt.Errorf("failed to prepare bytecode program: %w", err)
		}
	}
	if verbose && typeCheck && hasUnits {
		fmt.Fprintf(os.Stderr, "Type checking disabled (program uses units)\n")
	}
	cs.envelopeMsgs, done = reportCompileMessages(cs.result, wantHints)
	return cs, done, nil
}

// reportCompileMessages prints the hints/warnings of a successful compile (or holds
// them back for the test envelope) and handles --compile-only, whose output is the
// compiler's full message list like DWScript's Msgs.AsInfo (compile-only always runs
// in plain mode). done reports that the command is finished.
func reportCompileMessages(compiled *frontend.Result, wantHints bool) (envelopeMsgs []string, done bool) {
	switch {
	case compileOnly:
		for _, line := range compiled.DiagnosticStrings() {
			fmt.Fprintln(os.Stderr, line)
		}
	case wantHints && testEnvelope && !compileOnly:
		envelopeMsgs = compiled.HintStrings()
	case wantHints:
		for _, h := range compiled.HintStrings() {
			fmt.Fprintln(os.Stderr, h)
		}
	}
	return envelopeMsgs, compileOnly
}

// executeScript runs a compiled script on the bytecode VM or the AST interpreter.
func executeScript(cs *compiledScript) error {
	if dumpAST {
		fmt.Println("AST:")
		fmt.Println(cs.compiledProgram.String())
		fmt.Println()
	}
	if bytecodeMode {
		if showUnits && cs.unitRegistry != nil && len(cs.usedUnits) > 0 {
			displayUnitDependencyTree(cs.unitRegistry, cs.usedUnits)
		}
		return executeBytecodeProgram(cs.compiledProgram, bytecodeExecOptions{
			filename: cs.filename,
			trace:    trace,
		})
	}

	// With --test-envelope the program's output is buffered so the framing can be
	// decided once compile messages and the runtime outcome are known.
	var progOut bytes.Buffer
	var stdout io.Writer = os.Stdout
	if testEnvelope {
		stdout = &progOut
	}
	interpreter := runner.NewWithOptions(stdout, &simpleOptions{MaxRecursionDepth: maxRecursion})
	interpreter.SetSource(cs.input, cs.filename)
	if cs.result.Analyzer != nil {
		// Enables type inference for empty arrays and carries helper declarations over.
		interpreter.SetSemanticInfo(cs.result.SemanticInfo)
		interpreter.TransferHelpersFromSemanticAnalysis(cs.result.Analyzer.GetHelpers())
	}

	loaded, err := loadUnits(interpreter, cs)
	if err != nil {
		return err
	}

	if trace {
		fmt.Fprintf(os.Stderr, "[Trace mode enabled - executing %s]\n", cs.filename)
	}
	result := interpreter.Eval(cs.program)
	if loaded {
		// Finalization sections print through the same writer, so they must run
		// before the buffered output is emitted or the outcome is reported.
		if err := interpreter.FinalizeUnits(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error during unit finalization: %v\n", err)
		}
	}
	if testEnvelope {
		return emitTestEnvelope(interpreter, result, cs.envelopeMsgs, &progOut)
	}
	return reportRuntimeOutcome(interpreter, result)
}

// loadUnits sets up the unit registry and loads, imports and initializes every unit
// the program uses. loaded reports whether any unit was initialized (and therefore
// needs finalizing).
func loadUnits(interpreter *interp.Interpreter, cs *compiledScript) (loaded bool, err error) {
	if len(cs.searchPaths) == 0 || len(cs.usedUnits) == 0 {
		return false, nil
	}
	interpreter.SetUnitRegistry(units.NewUnitRegistry(cs.searchPaths))
	if verbose {
		fmt.Fprintf(os.Stderr, "Loading %d unit(s)...\n", len(cs.usedUnits))
	}
	for _, unitName := range cs.usedUnits {
		unit, err := interpreter.LoadUnit(unitName, nil)
		if err != nil {
			return false, fmt.Errorf("failed to load unit '%s': %w", unitName, err)
		}
		if err := interpreter.ImportUnitSymbols(unit); err != nil {
			return false, fmt.Errorf("failed to import symbols from unit '%s': %w", unitName, err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "  ✓ Loaded unit: %s\n", unitName)
		}
	}
	if err := interpreter.InitializeUnits(); err != nil {
		return false, fmt.Errorf("failed to initialize units: %w", err)
	}
	if verbose {
		if loadedUnits := interpreter.ListLoadedUnits(); len(loadedUnits) > 0 {
			fmt.Fprintf(os.Stderr, "Unit initialization order: %v\n", loadedUnits)
		}
	}
	if showUnits {
		displayUnitDependencyTree(interpreter.GetUnitRegistry(), cs.usedUnits)
	}
	return true, nil
}

// emitTestEnvelope writes the program output the way DWScript's test runner does:
// bare when there were no messages, otherwise
//
//	Errors >>>>
//	<compile hints/warnings, then the runtime error>
//	Result >>>>
//	<program output>
//
// Compile failures never get here; they are printed as a flat diagnostic list.
func emitTestEnvelope(interpreter *interp.Interpreter, result interp.Value, msgs []string, progOut *bytes.Buffer) error {
	var runErr error
	switch {
	case result != nil && result.Type() == "ERROR":
		msgs = append(msgs, interp.FormatRuntimeErrorValue(result))
		runErr = ErrSilent
	case interpreter.GetException() != nil:
		msgs = append(msgs, formatUnhandledException(interpreter.GetException()))
		runErr = ErrSilent
	}
	if len(msgs) > 0 {
		fmt.Print("Errors >>>>\n")
		for _, m := range msgs {
			fmt.Println(m)
		}
		fmt.Print("Result >>>>\n")
	}
	if _, err := os.Stdout.Write(progOut.Bytes()); err != nil {
		return err
	}
	return runErr
}

// reportRuntimeOutcome prints an unhandled exception or runtime error, if any, and
// returns the error the command should exit with (nil when the program succeeded).
func reportRuntimeOutcome(interpreter *interp.Interpreter, result interp.Value) error {
	// plain mirrors the fixture harness and reports the ERROR value first (an
	// uncaught exception surfaces there as "User defined exception: ..."); pretty
	// keeps the exception-first presentation with class name and call stack.
	if diagnosticsMode == "plain" {
		if result != nil && result.Type() == "ERROR" {
			fmt.Fprintln(os.Stderr, interp.FormatRuntimeErrorValue(result))
			return ErrSilent
		}
		if exc := interpreter.GetException(); exc != nil {
			fmt.Fprintln(os.Stderr, formatUnhandledException(exc))
			return ErrSilent
		}
		return nil
	}

	// Check for unhandled exceptions
	if exc := interpreter.GetException(); exc != nil {
		fmt.Fprintln(os.Stderr, formatUnhandledException(exc))
		// The StackTrace.String() method formats each frame with position info
		if len(exc.CallStack) > 0 {
			fmt.Fprint(os.Stderr, exc.CallStack.String())
			fmt.Fprintln(os.Stderr)
		}
		return fmt.Errorf("unhandled exception: %s", exc.Message)
	}

	// Check for runtime errors
	if result != nil && result.Type() == "ERROR" {
		// Structured RuntimeError: rich formatting with a source snippet
		if runtimeErr, ok := result.(*interp.RuntimeError); ok {
			if compilerErr := runtimeErr.ToCompilerError(); compilerErr != nil {
				fmt.Fprint(os.Stderr, compilerErr.Format(colorEnabled()))
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

// formatUnhandledException renders an uncaught script exception as
// "Runtime Error: <Class>: <message> [line: N, column: M]".
func formatUnhandledException(exc *runtime.ExceptionValue) string {
	className := "Exception"
	if exc.Metadata != nil {
		className = exc.Metadata.Name
	}
	if exc.Position != nil {
		return fmt.Sprintf("Runtime Error: %s: %s [line: %d, column: %d]",
			className, exc.Message, exc.Position.Line, exc.Position.Column)
	}
	return fmt.Sprintf("Runtime Error: %s: %s", className, exc.Message)
}
