package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Two-phase class construction
// ============================================================================
//
// Class declarations are constructed in two phases so that class analysis does
// not depend on source order:
//
//	Phase 1 (identity)    — every top-level class name gets a single
//	                        `*types.ClassType` shell registered up front. This is
//	                        the pre-existing predeclaration mechanism; there is
//	                        still exactly one type representation per class.
//	Phase 2 (inheritance) — the parent link and the class-level shape flags of
//	                        every predeclared shell are resolved before any
//	                        member or body checking runs.
//
// Phase 2 deliberately stays silent: unknown parents, conflicting forward/partial
// parents and inheritance cycles are left unlinked so that `analyzeClassDecl`
// reports them at the existing positions with the existing messages. What phase 2
// guarantees is that by the time a class declaration is analyzed, its ancestors'
// identities and shape flags are already known, regardless of where they were
// declared.

// classDeclGroup collects every top-level declaration that contributes to one
// class name (an explicit `forward` declaration plus its implementation, or the
// parts of a partial class).
type classDeclGroup struct {
	name  string
	decls []*ast.ClassDecl
}

// classInheritanceResolver performs phase 2 of class construction.
type classInheritanceResolver struct {
	analyzer *Analyzer
	groups   map[string]*classDeclGroup
	state    map[string]int
	order    []string
}

const (
	classResolveUnvisited = iota
	classResolveInProgress
	classResolveDone
)

// predeclareTopLevelClassTypes runs both phases of class construction over the
// top-level statements of a program.
func (a *Analyzer) predeclareTopLevelClassTypes(program *ast.Program) {
	for _, stmt := range program.Statements {
		a.predeclareClassTypesInStatement(stmt)
	}
	a.resolveTopLevelClassInheritance(program)
}

// predeclareClassTypesInStatement is phase 1: it registers the class identity.
func (a *Analyzer) predeclareClassTypesInStatement(stmt ast.Statement) {
	switch n := stmt.(type) {
	case *ast.BlockStatement:
		for _, inner := range n.Statements {
			a.predeclareClassTypesInStatement(inner)
		}
	case *ast.ClassDecl:
		className := classFullName(n)
		if className == "" || n.EnclosingClass != nil || a.hasType(className) {
			return
		}
		classType := types.NewClassType(className, nil)
		classType.IsForward = true
		a.registerTypeWithPos(className, classType, n.Token.Pos)
		a.predeclaredClassTypes[ident.Normalize(className)] = true
	}
}

// resolveTopLevelClassInheritance is phase 2: it propagates class-level shape
// flags onto the predeclared shells and links parents, so member and body
// checking sees a complete inheritance graph independent of declaration order.
func (a *Analyzer) resolveTopLevelClassInheritance(program *ast.Program) {
	r := &classInheritanceResolver{
		analyzer: a,
		groups:   make(map[string]*classDeclGroup),
		state:    make(map[string]int),
	}
	for _, stmt := range program.Statements {
		r.collect(stmt)
	}

	// Shape flags first: parent linking checks nothing that depends on them, but
	// the checks in analyzeClassDecl do.
	for _, key := range r.order {
		r.applyShapeFlags(r.groups[key])
	}
	for _, key := range r.order {
		r.resolve(key)
	}
}

// collect groups the top-level class declarations by normalized full name.
func (r *classInheritanceResolver) collect(stmt ast.Statement) {
	switch n := stmt.(type) {
	case *ast.BlockStatement:
		for _, inner := range n.Statements {
			r.collect(inner)
		}
	case *ast.ClassDecl:
		className := classFullName(n)
		if className == "" || n.EnclosingClass != nil {
			return
		}
		key := ident.Normalize(className)
		if !r.analyzer.isPredeclaredClassType(className) {
			// Not a shell we own (builtin name clash, or already fully defined).
			return
		}
		group, ok := r.groups[key]
		if !ok {
			group = &classDeclGroup{name: className}
			r.groups[key] = group
			r.order = append(r.order, key)
		}
		group.decls = append(group.decls, n)
	}
}

// applyShapeFlags copies the declaration-level class flags onto the shell so
// that inheritance validation of a *child* can consult them even when the
// parent is declared later in the source.
//
// `IsForward` and `IsPartial` are intentionally not touched here: they are
// lifecycle flags owned by analyzeClassDecl.
func (r *classInheritanceResolver) applyShapeFlags(group *classDeclGroup) {
	if group == nil {
		return
	}
	classType := r.analyzer.getClassType(group.name)
	if classType == nil {
		return
	}
	for _, decl := range group.decls {
		if decl.IsAbstract {
			classType.IsAbstract = true
		}
		if decl.IsExternal {
			classType.IsExternal = true
		}
		if decl.IsStaticClass {
			classType.IsStatic = true
		}
		if decl.ExternalName != "" {
			classType.ExternalName = decl.ExternalName
		}
		if decl.IsDeprecated {
			classType.IsDeprecated = true
			if decl.DeprecatedMessage != "" {
				classType.DeprecatedMessage = decl.DeprecatedMessage
			}
		}
	}
}

// declaredParent returns the parent declaration shared by every declaration in
// the group, or nil when the group declares no parent or the declarations
// disagree. Disagreement is left to analyzeClassDecl, which reports the specific
// forward/partial mismatch diagnostic.
func (r *classInheritanceResolver) declaredParent(group *classDeclGroup) *ast.Identifier {
	var parent *ast.Identifier
	for _, decl := range group.decls {
		candidate := decl.Parent
		if candidate == nil {
			// A class declaration may name its ancestor in the interface list;
			// mirror resolveParentClass and treat a leading class name as the
			// parent, rewriting the AST the same way it does.
			candidate = r.interfaceListParent(decl)
		}
		if candidate == nil {
			continue
		}
		if parent == nil {
			parent = candidate
			continue
		}
		if !ident.Equal(parent.Value, candidate.Value) {
			return nil
		}
	}
	return parent
}

// interfaceListParent applies the `class(TParent, IFoo)` disambiguation: if the
// first entry of the interface list names a class, it is really the parent.
func (r *classInheritanceResolver) interfaceListParent(decl *ast.ClassDecl) *ast.Identifier {
	if decl.Parent != nil || len(decl.Interfaces) == 0 {
		return nil
	}
	candidate := decl.Interfaces[0]
	if r.analyzer.getClassType(candidate.Value) == nil {
		return nil
	}
	decl.Parent = candidate
	decl.Interfaces = decl.Interfaces[1:]
	return candidate
}

// resolve links the parent of one class group, recursing into the parent first
// so that ancestors are linked before descendants. Cycles and unknown parents
// leave the link unset.
func (r *classInheritanceResolver) resolve(key string) {
	switch r.state[key] {
	case classResolveDone:
		return
	case classResolveInProgress:
		// Inheritance cycle: leave every class on the cycle unlinked here and let
		// analyzeClassDecl emit the existing "circular inheritance" diagnostic.
		return
	}
	r.state[key] = classResolveInProgress
	defer func() { r.state[key] = classResolveDone }()

	group := r.groups[key]
	if group == nil {
		return
	}
	classType := r.analyzer.getClassType(group.name)
	if classType == nil || classType.Parent != nil {
		return
	}

	parentIdent := r.declaredParent(group)
	if parentIdent == nil {
		// No explicit parent. The implicit TObject link stays with
		// analyzeClassDecl, which owns the "TObject missing" diagnostic.
		return
	}

	parentKey := ident.Normalize(parentIdent.Value)
	if parentKey == key {
		// Direct self-inheritance; reported by analyzeClassDecl.
		return
	}
	if _, ok := r.groups[parentKey]; ok {
		r.resolve(parentKey)
		if r.state[parentKey] != classResolveDone {
			return
		}
	}

	parentClass := r.analyzer.getClassType(parentIdent.Value)
	if parentClass == nil {
		// Unknown parent; reported by analyzeClassDecl.
		return
	}
	if parentClass == classType || r.reaches(parentClass, classType) {
		return
	}
	classType.Parent = parentClass
}

// reaches reports whether `ancestor` is reachable from `start` by walking the
// already-linked parent chain. The walk is length-bounded so that a pre-existing
// cycle in the graph can never spin forever.
func (r *classInheritanceResolver) reaches(start, target *types.ClassType) bool {
	seen := make(map[*types.ClassType]bool)
	for current := start; current != nil; current = current.Parent {
		if current == target {
			return true
		}
		if seen[current] {
			return true
		}
		seen[current] = true
	}
	return false
}
