package interp

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Registrations are constructed from internal/types' shared helper catalog. These
// tests pin the registry adapters and execution handlers to that catalog so an
// operation cannot silently disappear during a refactor.

// evaluatorHelperFiles lists the evaluator files whose eval*Helper switches
// dispatch builtin helper specs.
var evaluatorHelperFiles = []string{
	"evaluator/array_helpers.go",
	"evaluator/string_helpers.go",
	"evaluator/integer_helpers.go",
	"evaluator/float_helpers.go",
	"evaluator/boolean_helpers.go",
	"evaluator/enum_helpers.go",
	"evaluator/helper_methods.go",
}

// specSet is a set of helper spec strings.
type specSet map[string]struct{}

func (s specSet) add(spec string) { s[spec] = struct{}{} }

func (s specSet) sorted() []string {
	out := make([]string, 0, len(s))
	for spec := range s {
		out = append(out, spec)
	}
	sort.Strings(out)
	return out
}

// filter returns the subset of s for which keep returns true.
func (s specSet) filter(keep func(string) bool) specSet {
	out := specSet{}
	for spec := range s {
		if keep(spec) {
			out.add(spec)
		}
	}
	return out
}

// minus returns the entries of s that are absent from other.
func (s specSet) minus(other specSet) []string {
	out := specSet{}
	for spec := range s {
		if _, ok := other[spec]; !ok {
			out.add(spec)
		}
	}
	return out.sorted()
}

func isMagicSpec(spec string) bool { return strings.HasPrefix(spec, "__") }

// collectPropertySpecs adds every builtin-backed read/write spec of props to out.
func collectPropertySpecs(out specSet, props map[string]*types.PropertyInfo) {
	for _, prop := range props {
		if prop == nil {
			continue
		}
		if prop.ReadKind == types.PropAccessBuiltin && prop.ReadSpec != "" {
			out.add(prop.ReadSpec)
		}
		if prop.WriteKind == types.PropAccessBuiltin && prop.WriteSpec != "" {
			out.add(prop.WriteSpec)
		}
	}
}

// semanticHelperSpecs returns every builtin spec referenced by the semantic
// analyzer's eagerly built helper table.
func semanticHelperSpecs(t *testing.T) specSet {
	t.Helper()
	out := specSet{}
	for _, helpers := range semantic.NewAnalyzer().GetHelpers() {
		for _, helper := range helpers {
			if helper == nil {
				continue
			}
			for _, spec := range helper.BuiltinMethods {
				out.add(spec)
			}
			collectPropertySpecs(out, helper.Properties)
		}
	}
	if len(out) == 0 {
		t.Fatal("semantic helper table is empty; expected builtin helper specs")
	}
	return out
}

// runtimeHelperTable returns the interpreter's registered helper table keyed by
// normalized target type name.
func runtimeHelperTable(t *testing.T) map[string][]*HelperInfo {
	t.Helper()
	interp := New(io.Discard)
	if interp.typeSystem == nil {
		t.Fatal("interpreter has no type system")
	}
	out := make(map[string][]*HelperInfo)
	for typeName, helpers := range interp.typeSystem.AllHelpers() {
		out[typeName] = append(out[typeName], helpers...)
	}
	return out
}

// runtimeHelperSpecs returns every builtin spec referenced by the interpreter's
// runtime helper table.
func runtimeHelperSpecs(t *testing.T) specSet {
	t.Helper()
	out := specSet{}
	for _, helpers := range runtimeHelperTable(t) {
		for _, helper := range helpers {
			for _, spec := range helper.BuiltinMethods {
				out.add(spec)
			}
			collectPropertySpecs(out, helper.Properties)
		}
	}
	if len(out) == 0 {
		t.Fatal("runtime helper table is empty; expected builtin helper specs")
	}
	return out
}

// switchCaseStrings parses path and returns the string literals that appear as
// direct elements of a case clause inside any function whose name satisfies
// matchFunc.
func switchCaseStrings(t *testing.T, path string, matchFunc func(string) bool) specSet {
	t.Helper()
	operationConstants := helperOperationConstants(t)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	out := specSet{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*goast.FuncDecl)
		if !ok || fn.Body == nil || !matchFunc(fn.Name.Name) {
			continue
		}
		goast.Inspect(fn.Body, func(n goast.Node) bool {
			clause, ok := n.(*goast.CaseClause)
			if !ok {
				return true
			}
			for _, expr := range clause.List {
				if selector, ok := expr.(*goast.SelectorExpr); ok {
					if pkg, ok := selector.X.(*goast.Ident); ok && pkg.Name == "types" {
						if spec, ok := operationConstants[selector.Sel.Name]; ok {
							out.add(spec)
						}
					}
					continue
				}
				lit, ok := expr.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatalf("%s: unquote %s: %v", fset.Position(lit.Pos()), lit.Value, err)
				}
				out.add(value)
			}
			return true
		})
	}
	return out
}

// evaluatorHandledSpecs returns every spec string that has a case arm in one of
// the evaluator's eval*Helper dispatch switches or in evalBuiltinHelperProperty.
func evaluatorHandledSpecs(t *testing.T) specSet {
	t.Helper()
	isDispatcher := func(name string) bool {
		return name == "evalBuiltinHelperProperty" ||
			(strings.HasPrefix(name, "eval") && strings.HasSuffix(name, "Helper"))
	}
	out := specSet{}
	for _, rel := range evaluatorHelperFiles {
		for spec := range switchCaseStrings(t, filepath.FromSlash(rel), isDispatcher) {
			out.add(spec)
		}
	}
	if len(out) == 0 {
		t.Fatal("no evaluator helper case strings found; parser matcher is probably broken")
	}
	return out
}

// registryHandles reports whether spec is a plain builtin function name
// registered in builtins.DefaultRegistry.
func registryHandles(spec string) bool {
	_, ok := builtins.DefaultRegistry.Lookup(spec)
	return ok
}

// dispatcherFallsBackToRegistry reports whether the named evaluator dispatcher
// in evaluator/helper_methods.go contains a builtins.DefaultRegistry.Lookup
// call. A spec that is only a plain builtin name counts as handled only when
// the dispatcher actually routes unknown specs to the registry.
func dispatcherFallsBackToRegistry(t *testing.T, funcName string) bool {
	t.Helper()
	fset := token.NewFileSet()
	path := filepath.FromSlash("evaluator/helper_methods.go")
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*goast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name.Name != funcName {
			continue
		}
		goast.Inspect(fn.Body, func(n goast.Node) bool {
			sel, ok := n.(*goast.SelectorExpr)
			if !ok || sel.Sel.Name != "Lookup" {
				return true
			}
			inner, ok := sel.X.(*goast.SelectorExpr)
			if !ok || inner.Sel.Name != "DefaultRegistry" {
				return true
			}
			if pkg, ok := inner.X.(*goast.Ident); ok && pkg.Name == "builtins" {
				found = true
			}
			return !found
		})
	}
	return found
}

func TestHelperSpecParity(t *testing.T) {
	semanticSpecs := semanticHelperSpecs(t)
	runtimeSpecs := runtimeHelperSpecs(t)
	evaluatorSpecs := evaluatorHandledSpecs(t)
	registryRouted := dispatcherFallsBackToRegistry(t, "CallBuiltinHelperMethod") &&
		dispatcherFallsBackToRegistry(t, "CallBuiltinHelperProperty")

	t.Run("semantic and runtime tables declare the same magic specs", func(t *testing.T) {
		semanticMagic := semanticSpecs.filter(isMagicSpec)
		runtimeMagic := runtimeSpecs.filter(isMagicSpec)
		if missing := runtimeMagic.minus(semanticMagic); len(missing) > 0 {
			t.Errorf("%d spec(s) registered at runtime but missing from the semantic table:\n  %s",
				len(missing), strings.Join(missing, "\n  "))
		}
		if missing := semanticMagic.minus(runtimeMagic); len(missing) > 0 {
			t.Errorf("%d spec(s) registered in the semantic table but missing at runtime:\n  %s",
				len(missing), strings.Join(missing, "\n  "))
		}
	})

	t.Run("every registered spec has an evaluator or registry handler", func(t *testing.T) {
		union := specSet{}
		for spec := range semanticSpecs {
			union.add(spec)
		}
		for spec := range runtimeSpecs {
			union.add(spec)
		}
		var unhandled []string
		for _, spec := range union.sorted() {
			if _, ok := evaluatorSpecs[spec]; ok {
				continue
			}
			if registryRouted && registryHandles(spec) {
				continue
			}
			unhandled = append(unhandled, spec)
		}
		if len(unhandled) > 0 {
			t.Errorf("%d spec(s) registered in a helper table but handled by neither an evaluator case nor a builtins.DefaultRegistry fallback in CallBuiltinHelperMethod/CallBuiltinHelperProperty:\n  %s",
				len(unhandled), strings.Join(unhandled, "\n  "))
		}
	})

	t.Run("every evaluator case is registered in a helper table", func(t *testing.T) {
		var orphans []string
		for _, spec := range evaluatorSpecs.sorted() {
			_, inSemantic := semanticSpecs[spec]
			_, inRuntime := runtimeSpecs[spec]
			if !inSemantic && !inRuntime {
				orphans = append(orphans, spec)
			}
		}
		if len(orphans) > 0 {
			t.Errorf("%d evaluator case string(s) are not registered in any helper table (orphan handlers):\n  %s",
				len(orphans), strings.Join(orphans, "\n  "))
		}
	})
}

// TestArrayHelperCatalogMatchesRuntimeTable checks member names as well as
// operations: aliases and property call forms must survive runtime adaptation.
func TestArrayHelperCatalogMatchesRuntimeTable(t *testing.T) {
	arrayHelpers := runtimeHelperTable(t)[ident.Normalize("array")]
	if len(arrayHelpers) != 1 {
		t.Fatalf("expected one array helper, got %d", len(arrayHelpers))
	}
	runtimeHelper := arrayHelpers[0]
	semanticHelper := semantic.NewAnalyzer().GetHelpers()["array"][0]
	for name, spec := range runtimeHelper.BuiltinMethods {
		member, ok := types.LookupBuiltinHelper("array", name)
		if !ok || !member.Method || string(member.Operation) != spec {
			t.Errorf("array.%s missing from catalog", name)
		}
		if semanticHelper.BuiltinMethods[name] != spec {
			t.Errorf("array.%s differs between runtime and analyzer", name)
		}
	}
	if len(runtimeHelper.BuiltinMethods) != 24 {
		t.Errorf("expected 24 array helpers, got %d", len(runtimeHelper.BuiltinMethods))
	}
}

func helperOperationConstants(t *testing.T) map[string]string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "types", "helper_specs.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]string)
	goast.Inspect(file, func(node goast.Node) bool {
		spec, ok := node.(*goast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		lit, ok := spec.Values[0].(*goast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatal(err)
		}
		result[spec.Names[0].Name] = value
		return true
	})
	return result
}
