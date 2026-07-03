// Package generics implements monomorphization of DWScript generic types.
//
// DWScript generics are specialized (monomorphized): a declaration such as
//
//	type TList<T> = class ... end;
//
// combined with a use such as `new TList<Integer>` produces a concrete class
// named "TList<Integer>" with every occurrence of the type parameter T replaced
// by Integer. This package performs that transformation on the parsed AST,
// before semantic analysis and evaluation, so that the rest of the pipeline only
// ever sees ordinary concrete classes and records.
//
// The transformation:
//   - Collects generic templates (class/record/alias declarations that carry
//     type parameters) and removes them from the program.
//   - Rewrites every generic type reference (`TList<Integer>` in a type
//     annotation, a `new` expression, or expression position) to the mangled
//     concrete name, generating the corresponding specialized declaration the
//     first time it is seen.
//   - Inserts each specialized declaration immediately before its first use, so
//     that its type arguments (which DWScript requires to be declared earlier)
//     are already registered.
package generics

import (
	"reflect"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Monomorphize rewrites generic type usage in prog into concrete specializations.
// If the program declares no generic templates it is left unchanged.
func Monomorphize(prog *ast.Program) {
	if prog == nil {
		return
	}
	m := &monomorphizer{
		templates: make(map[string]templateInfo),
		emitted:   make(map[string]bool),
	}
	m.collectTemplates(prog.Statements)
	if len(m.templates) == 0 {
		return
	}
	m.run(prog)
}

type templateInfo struct {
	decl   ast.Statement
	params []string
}

type monomorphizer struct {
	templates map[string]templateInfo
	emitted   map[string]bool
	// out is the statement list being built for the current pass; specialized
	// declarations are appended here just before the statement that uses them.
	out []ast.Statement
}

// collectTemplates records every generic template declaration by its base name.
func (m *monomorphizer) collectTemplates(stmts []ast.Statement) {
	for _, stmt := range stmts {
		if params := typeParamsOf(stmt); len(params) > 0 {
			m.templates[ident.Normalize(declName(stmt))] = templateInfo{decl: stmt, params: params}
		}
	}
}

func (m *monomorphizer) run(prog *ast.Program) {
	m.out = make([]ast.Statement, 0, len(prog.Statements))
	for _, stmt := range prog.Statements {
		if isTemplateDecl(stmt) {
			continue // templates are replaced by their specializations
		}
		m.rewrite(reflect.ValueOf(stmt))
		m.out = append(m.out, stmt)
	}
	prog.Statements = m.out
}

// isTemplate reports whether name refers to a collected generic template.
func (m *monomorphizer) isTemplate(name string) bool {
	_, ok := m.templates[ident.Normalize(name)]
	return ok
}

// ensureSpecialized generates (once) the concrete declaration for base<args> and
// returns the mangled specialized name. Newly generated declarations — and any
// nested specializations they require — are appended to m.out in dependency
// order, before the statement currently being rewritten.
func (m *monomorphizer) ensureSpecialized(base string, args []ast.TypeExpression) string {
	tpl, ok := m.templates[ident.Normalize(base)]
	if !ok {
		// Not a generic template: leave the reference untouched.
		return ast.MangleGenericName(base, args)
	}

	mangled := ast.MangleGenericName(base, args)
	key := ident.Normalize(mangled)
	if m.emitted[key] {
		return mangled
	}
	// Mark before recursing so a self-referential generic terminates.
	m.emitted[key] = true

	subst := make(map[string]ast.TypeExpression, len(tpl.params))
	for i, p := range tpl.params {
		if i < len(args) {
			subst[ident.Normalize(p)] = args[i]
		}
	}

	clone, ok := cloneNode(reflect.ValueOf(tpl.decl), subst).Interface().(ast.Statement)
	if !ok {
		return mangled
	}
	specializeDecl(clone, mangled)

	// Rewrite any generic references inside the freshly specialized declaration,
	// emitting their dependencies before it.
	m.rewrite(reflect.ValueOf(clone))
	m.out = append(m.out, clone)
	return mangled
}

// rewrite walks the tree at v, rewriting generic type references in place and
// replacing GenericTypeRef expression nodes with plain identifiers carrying the
// mangled name.
func (m *monomorphizer) rewrite(v reflect.Value) {
	switch v.Kind() {
	case reflect.Interface:
		m.rewriteInterface(v)
	case reflect.Ptr:
		m.rewritePtr(v)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanInterface() {
				continue // unexported field (e.g. SemanticInfo internals)
			}
			m.rewrite(f)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			m.rewrite(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			m.rewrite(v.MapIndex(key))
		}
	}
}

// rewriteInterface handles an interface-typed slot, replacing a GenericTypeRef
// with a plain identifier and otherwise recursing into the concrete value.
func (m *monomorphizer) rewriteInterface(v reflect.Value) {
	if v.IsNil() {
		return
	}
	concrete := v.Elem()
	if gtr, ok := concrete.Interface().(*ast.GenericTypeRef); ok {
		mangled := m.ensureSpecialized(gtr.Base.Value, gtr.TypeArgs)
		if v.CanSet() {
			v.Set(reflect.ValueOf(identForMangled(gtr, mangled)))
		}
		return
	}
	m.rewrite(concrete)
}

// rewritePtr handles a pointer-typed slot, rewriting generic type references on
// TypeAnnotation and NewExpression nodes in place after recursing into children.
func (m *monomorphizer) rewritePtr(v reflect.Value) {
	if v.IsNil() {
		return
	}
	switch node := v.Interface().(type) {
	case *ast.TypeAnnotation:
		m.rewrite(v.Elem()) // rewrite children (InlineType, nested TypeArgs) first
		if len(node.TypeArgs) > 0 && m.isTemplate(node.Name) {
			node.Name = m.ensureSpecialized(node.Name, node.TypeArgs)
			node.TypeArgs = nil
		}
	case *ast.NewExpression:
		m.rewrite(v.Elem())
		if len(node.TypeArgs) > 0 && node.ClassName != nil && m.isTemplate(node.ClassName.Value) {
			node.ClassName.Value = m.ensureSpecialized(node.ClassName.Value, node.TypeArgs)
			node.TypeArgs = nil
		}
	default:
		m.rewrite(v.Elem())
	}
}

// identForMangled builds an identifier that stands in for a GenericTypeRef after
// specialization, carrying the mangled concrete name.
func identForMangled(gtr *ast.GenericTypeRef, mangled string) *ast.Identifier {
	tok := gtr.Token
	if gtr.Base != nil {
		tok = gtr.Base.Token
	}
	return &ast.Identifier{
		TypedExpressionBase: ast.TypedExpressionBase{BaseNode: ast.BaseNode{Token: tok}},
		Value:               mangled,
	}
}
