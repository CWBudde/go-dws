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
		// Monomorphization must happen on the same nodes later passed to execution.
		if err := generics.Monomorphize(&ast.Program{Statements: []ast.Statement{unit.Declaration}}); err != nil {
			return fail(err)
		}
		unitAnalyzer := semantic.NewAnalyzer()
		unitAnalyzer.SetSemanticInfo(analyzer.GetSemanticInfo())
		unitAnalyzer.SetHintsLevel(opts.HintsLevel)
		unitAnalyzer.SetSource(unit.Source, unit.FilePath)
		err := unitAnalyzer.AnalyzeUnitWithDependencies(unit.Declaration, available)
		diagnostics := semanticDiagnostics(unitAnalyzer)
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
		if err := analyzer.ImportUnitSymbols(name.Value, available[ident.Normalize(name.Value)]); err != nil {
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
