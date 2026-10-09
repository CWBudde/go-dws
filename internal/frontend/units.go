package frontend

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/cwbudde/go-dws/internal/generics"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/units"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

type deferredUnitForwardCheck struct {
	analyzer        *semantic.Analyzer
	diagnosticCount int
}

func (r *Result) deferredIndexStops() []*ast.IndexExpression {
	var indices []*ast.IndexExpression
	for _, diag := range r.Diagnostics {
		if diag.Stop && diag.deferredIndex != nil {
			indices = append(indices, diag.deferredIndex)
		}
	}
	return indices
}

// completeDeferredUnitIndexForwards applies the resolved main-file stop state
// only to the held unit end checks. Their diagnostics retain unit source ownership.
func completeDeferredUnitIndexForwards(result *Result, mainDiagnostics []Diagnostic) []Diagnostic {
	mainStopped := false
	for _, diag := range mainDiagnostics {
		mainStopped = mainStopped || diag.Stop
	}
	var diags []Diagnostic
	for _, pending := range result.deferredUnitForwards {
		pending.analyzer.CompleteDeferredIndexForwardChecks(mainStopped)
		updated := semanticDiagnostics(pending.analyzer)
		diags = append(diags, updated[pending.diagnosticCount:]...)
	}
	result.deferredUnitForwards = nil
	return diags
}

func safeAnalyzeWithUnits(analyzer *semantic.Analyzer, result *Result, opts Options) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("semantic analysis panic: %v", recovered)
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Message: "internal semantic analysis panic", Rendered: fmt.Sprintf("%v\n%s", err, strings.TrimSpace(string(debug.Stack()))), Code: "E_SEMANTIC_PANIC", Phase: PhaseSemantic, Severity: SeverityError, Fatal: true})
		}
	}()
	if err = analyzeUnits(analyzer, result, opts); err != nil {
		return err
	}
	return safeAnalyze(analyzer, result)
}

func analyzeUnits(analyzer *semantic.Analyzer, result *Result, opts Options) error {
	imports := programUnitImports(result.Program)
	if len(imports) == 0 {
		return nil
	}
	paths := opts.UnitSearchPaths
	if len(paths) == 0 {
		if dir := includeDirFor(opts.Filename); dir != "" {
			paths = []string{dir}
		}
	}
	registry := units.NewUnitRegistry(paths)
	registry.SetSourceFile(opts.Filename)
	registry.SetDefines(opts.Defines)
	result.UnitRegistry = registry
	fail := func(err error) error {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{Message: err.Error(), Phase: PhaseSemantic, Severity: SeverityError, Fatal: true})
		return err
	}
	for _, name := range imports {
		if _, err := registry.LoadUnit(name.Value, nil); err != nil {
			return fail(err)
		}
	}
	order, err := registry.ComputeInitializationOrder()
	if err != nil {
		return fail(err)
	}
	available := make(map[string]*semantic.SymbolTable)
	for _, name := range order {
		unit, _ := registry.GetUnit(name)
		directives := filterSourceHints(lexerDiagnostics(unit.DirectiveDiagnostics, false), opts.HintsLevel)
		result.Diagnostics = append(result.Diagnostics, directives...)
		for _, diagnostic := range directives {
			if diagnostic.Fatal {
				return fmt.Errorf("compiler directive error in unit %q", name)
			}
		}
		// Monomorphization must happen on the same nodes later passed to execution.
		if err := generics.Monomorphize(&ast.Program{Statements: []ast.Statement{unit.Declaration}}); err != nil {
			return fail(err)
		}
		unitAnalyzer := semantic.NewAnalyzer()
		unitAnalyzer.SetSemanticInfo(analyzer.GetSemanticInfo())
		unitAnalyzer.SetHintsLevel(opts.HintsLevel)
		unitAnalyzer.SetSymbolDictionaryDiagnostics(!opts.DisableSymbolDictionaryDiagnostics)
		// Upstream initializes the unit symbol tables in the same guarded block as
		// the program's, so a stop in the main source skips the end-of-compilation
		// checks of the units too.
		unitAnalyzer.SetCompileStopped(result.HasParserStop())
		indexStops := result.deferredIndexStops()
		unitAnalyzer.SetDeferredIndexStops(indexStops)
		unitAnalyzer.SetSource(unit.Source, unit.FilePath)
		err := unitAnalyzer.AnalyzeUnitWithDependencies(unit.Declaration, available)
		diagnostics := semanticDiagnostics(unitAnalyzer)
		if len(indexStops) != 0 {
			result.deferredUnitForwards = append(result.deferredUnitForwards, deferredUnitForwardCheck{analyzer: unitAnalyzer, diagnosticCount: len(diagnostics)})
		}
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		if err != nil {
			fatal := false
			for _, diagnostic := range diagnostics {
				fatal = fatal || diagnostic.Fatal
			}
			if !fatal {
				return fail(err)
			}
			return err
		}
		unit.Symbols = unitAnalyzer.GetUnitSymbols(unit.Name)
		available[ident.Normalize(unit.Name)] = unit.Symbols
	}
	for _, name := range imports {
		if err := analyzer.ProvideUnitSymbols(name.Value, available[ident.Normalize(name.Value)]); err != nil {
			return fail(err)
		}
	}
	return nil
}

// programUnitImports preserves the declaration order of program-level uses clauses.
func programUnitImports(program *ast.Program) []*ast.Identifier {
	var imports []*ast.Identifier
	for _, stmt := range program.Statements {
		if uses, ok := stmt.(*ast.UsesClause); ok {
			imports = append(imports, uses.Units...)
		}
	}
	return imports
}
