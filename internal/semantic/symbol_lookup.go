package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

// declarationTable follows the pinned DWScript TSymbolTable algorithm. Equal
// names retain distinct identities; both insertion and lookup stop at midpoint.
// Parameter tables instead preserve declaration order.
type declarationTable struct {
	entries []*Symbol
	sorted  bool
	ordered bool
}

func (t *declarationTable) add(sym *Symbol) {
	i := len(t.entries)
	if t.sorted && !t.ordered {
		lo, hi := 0, len(t.entries)-1
	insertionSearch:
		for lo <= hi {
			mid := (lo + hi) / 2
			cmp := compareSymbolNames(t.entries[mid].Name, sym.Name)
			switch {
			case cmp < 0:
				lo = mid + 1
			case cmp > 0:
				hi = mid - 1
			default:
				lo = mid
				break insertionSearch
			}
		}
		i = lo
	}
	t.entries = append(t.entries, nil)
	copy(t.entries[i+1:], t.entries[i:])
	t.entries[i] = sym
}
func compareSymbolNames(a, b string) int {
	return strings.Compare(ident.Normalize(a), ident.Normalize(b))
}

func (t *declarationTable) find(name string) (*Symbol, bool) {
	if t.ordered {
		for _, sym := range t.entries {
			if ident.Equal(sym.Name, name) {
				return sym, true
			}
		}
		return nil, false
	}
	if len(t.entries) == 0 {
		return nil, false
	}
	if !t.sorted {
		t.sort(0, len(t.entries)-1)
		t.sorted = true
	}
	lo, hi := 0, len(t.entries)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		sym := t.entries[mid]
		cmp := compareSymbolNames(sym.Name, name)
		switch {
		case cmp < 0:
			lo = mid + 1
		case cmp > 0:
			hi = mid - 1
		default:
			return sym, true
		}
	}
	return nil, false
}

// sort deliberately uses upstream's unstable midpoint pivot, including pivot
// tracking after swaps. A stable sort changes which duplicate gets selected.
func (t *declarationTable) sort(lo, hi int) {
	if hi <= lo {
		return
	}
	for {
		i, j, p := lo, hi, (lo+hi)/2
		for {
			pivot := t.entries[p]
			for compareSymbolNames(t.entries[i].Name, pivot.Name) < 0 {
				i++
			}
			for compareSymbolNames(t.entries[j].Name, pivot.Name) > 0 {
				j--
			}
			if i <= j {
				t.entries[i], t.entries[j] = t.entries[j], t.entries[i]
				switch p {
				case i:
					p = j
				case j:
					p = i
				}
				i++
				j--
			}
			if i > j {
				break
			}
		}
		if lo < j {
			t.sort(lo, j)
		}
		lo = i
		if i >= hi {
			break
		}
	}
}

func (t declarationTable) snapshot() declarationTable {
	t.entries = append([]*Symbol(nil), t.entries...)
	return t
}

func (st *SymbolTable) put(name string, sym *Symbol) {
	st.symbols.Set(name, sym)
	if sym.DeclPosition.Line == 0 {
		return
	}
	// Type registration and its value binding denote one source declaration.
	for i, entry := range st.declarations.entries {
		if entry.lookupOnly && ident.Equal(entry.Name, name) {
			st.declarations.entries[i] = sym
			return
		}
	}
	st.declarations.add(sym)
}

func (st *SymbolTable) registerTypeEntry(name string, typ types.Type, pos token.Position) {
	if pos.Line == 0 {
		return
	}
	if _, ok := st.declarations.find(name); ok {
		return
	}
	st.declarations.add(&Symbol{Name: name, Type: typ, DeclPosition: pos, lookupOnly: true})
}

func (st *SymbolTable) registerUnitEntry(name string) {
	if prefix, _, dotted := strings.Cut(name, "."); dotted {
		st.registerUnitEntry(prefix)
	}
	if _, ok := st.declarations.find(name); !ok {
		st.declarations.add(&Symbol{Name: name, lookupOnly: true, isUnitName: true})
	}
}

func (st *SymbolTable) findLocal(name string) (*Symbol, bool) {
	sym, ok := st.findLocalIdentity(name)
	if ok && sym.lookupOnly {
		return nil, false
	}
	return sym, ok
}

func (st *SymbolTable) findLocalIdentity(name string) (*Symbol, bool) {
	table := &st.declarations
	if st.sourceSnapshot != nil {
		table = st.sourceSnapshot
	}
	if sym, ok := table.find(name); ok {
		if !sym.lookupOnly {
			if set, ok := st.symbols.Get(name); ok && set.IsOverloadSet {
				return set, true
			}
			return sym, true
		}
		return sym, true
	}
	if sym, ok := st.parameters.find(name); ok {
		return sym, true
	}
	if sym, ok := st.internalParameters.find(name); ok {
		return sym, true
	}
	sym, ok := st.symbols.Get(name)
	// Source constants absent from a deferred body's snapshot are not visible.
	if ok && st.sourceSnapshot != nil && (sym.EnumElement != nil || sym.IsRecordTypeName) {
		return nil, false
	}
	return sym, ok
}

// sourceScopeSnapshot retains declaration vectors, sharing immutable symbol
// identities. Activation restores the captured source root and parent order
// while preserving deliberate later routine visibility in each scope.
type sourceScopeState struct {
	scope   *SymbolTable
	parents []*SymbolTable
	table   declarationTable
}
type sourceScopeSnapshot struct {
	root     *SymbolTable
	analyzer *Analyzer
	units    map[string]*SymbolTable
	imports  map[string]bool
	scopes   []sourceScopeState
}

func (a *Analyzer) captureSourceScope(scope *SymbolTable) sourceScopeSnapshot {
	result := sourceScopeSnapshot{root: scope, analyzer: a, units: make(map[string]*SymbolTable, len(a.unitSymbols)), imports: make(map[string]bool, len(a.typeRegistry.activeImports))}
	for name, unit := range a.unitSymbols {
		result.units[name] = unit
	}
	for name, visible := range a.typeRegistry.activeImports {
		result.imports[name] = visible
	}
	for ; scope != nil; scope = scope.outer {
		table := scope.declarations
		if scope.sourceSnapshot != nil {
			table = *scope.sourceSnapshot
		}
		result.scopes = append(result.scopes, sourceScopeState{scope: scope, table: table.snapshot(), parents: append([]*SymbolTable(nil), scope.importedParents...)})
	}
	return result
}
func (snapshot sourceScopeSnapshot) activate() func() {
	previous := make([]*declarationTable, len(snapshot.scopes))
	parents := make([][]*SymbolTable, len(snapshot.scopes))
	for i := range snapshot.scopes {
		state := &snapshot.scopes[i]
		previous[i] = state.scope.sourceSnapshot
		parents[i] = state.scope.importedParents
		state.scope.sourceSnapshot = &state.table
		state.scope.importedParents = state.parents
	}
	a := snapshot.analyzer
	previousRoot := a.symbols
	a.symbols = snapshot.root
	units, imports := a.unitSymbols, a.typeRegistry.activeImports
	a.unitSymbols = snapshot.units
	a.typeRegistry.activeImports = snapshot.imports
	// A deferred local declaration may reuse a name imported only later. Keep
	// that later descriptor so leaving the body restores the outer type view.
	hiddenTypes := make(map[string]*TypeDescriptor)
	a.typeRegistry.types.Range(func(name string, descriptor *TypeDescriptor) bool {
		if !a.typeRegistry.visible(descriptor) {
			hiddenTypes[name] = descriptor
		}
		return true
	})
	a.typeRegistry.kindIndex = make(map[string][]string)
	return func() {
		for i, state := range snapshot.scopes {
			state.scope.sourceSnapshot = previous[i]
			state.scope.importedParents = parents[i]
		}
		a.symbols = previousRoot
		a.unitSymbols = units
		a.typeRegistry.activeImports = imports
		for name, descriptor := range hiddenTypes {
			a.typeRegistry.types.Set(name, descriptor)
		}
		a.typeRegistry.kindIndex = make(map[string][]string)
	}
}

// defineInternal records Result, Self, or a compatibility alias without adding
// a source local. DWScript stores Result/Self in ordered internal parameters.
func (st *SymbolTable) defineInternal(name string, typ types.Type, pos token.Position) {
	sym := &Symbol{Name: name, Type: typ, DeclPosition: pos}
	st.symbols.Set(name, sym)
	st.internalParameters.add(sym)
}
