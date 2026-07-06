package generics

import (
	"reflect"
	"strings"

	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// typeParamsOf returns the generic type-parameter names declared by stmt, or nil
// if stmt is not a generic template.
func typeParamsOf(stmt ast.Statement) []string {
	switch d := stmt.(type) {
	case *ast.ClassDecl:
		return d.TypeParams
	case *ast.RecordDecl:
		return d.TypeParams
	case *ast.TypeDeclaration:
		return d.TypeParams
	default:
		return nil
	}
}

// isTemplateDecl reports whether stmt is a generic template declaration.
func isTemplateDecl(stmt ast.Statement) bool {
	return len(typeParamsOf(stmt)) > 0
}

// declName returns the declared type name for a type declaration statement.
func declName(stmt ast.Statement) string {
	switch d := stmt.(type) {
	case *ast.ClassDecl:
		if d.Name != nil {
			return d.Name.Value
		}
	case *ast.RecordDecl:
		if d.Name != nil {
			return d.Name.Value
		}
	case *ast.TypeDeclaration:
		if d.Name != nil {
			return d.Name.Value
		}
	}
	return ""
}

// specializeDecl finalizes a cloned template as a concrete declaration: it sets
// the declaration's name to the mangled specialized name and clears its type
// parameters so downstream phases treat it as an ordinary type.
func specializeDecl(stmt ast.Statement, mangled string) {
	switch d := stmt.(type) {
	case *ast.ClassDecl:
		setIdentValue(d.Name, mangled)
		d.TypeParams = nil
	case *ast.RecordDecl:
		setIdentValue(d.Name, mangled)
		d.TypeParams = nil
	case *ast.TypeDeclaration:
		setIdentValue(d.Name, mangled)
		d.TypeParams = nil
	}
}

func setIdentValue(id *ast.Identifier, value string) {
	if id != nil {
		id.Value = value
	}
}

// cloneNode performs a deep copy of an AST value, applying generic type-parameter
// substitution: any TypeAnnotation or Identifier whose name matches a key in
// subst is replaced by the corresponding concrete type argument. Pass a nil subst
// for a plain deep copy.
func cloneNode(v reflect.Value, subst map[string]ast.TypeExpression) reflect.Value {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return v
		}
		cp := reflect.New(v.Elem().Type())
		cp.Elem().Set(cloneNode(v.Elem(), subst))
		if len(subst) > 0 {
			applySubst(cp.Interface(), subst)
		}
		return cp

	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		inner := cloneNode(v.Elem(), subst)
		out := reflect.New(v.Type()).Elem()
		out.Set(inner)
		return out

	case reflect.Slice:
		if v.IsNil() {
			return v
		}
		cp := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := 0; i < v.Len(); i++ {
			cp.Index(i).Set(cloneNode(v.Index(i), subst))
		}
		return cp

	case reflect.Map:
		if v.IsNil() {
			return v
		}
		cp := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			cp.SetMapIndex(iter.Key(), cloneNode(iter.Value(), subst))
		}
		return cp

	case reflect.Struct:
		cp := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanInterface() {
				continue // skip unexported fields (not part of the decl tree)
			}
			cp.Field(i).Set(cloneNode(f, subst))
		}
		return cp

	default:
		return v
	}
}

// applySubst rewrites a freshly-cloned TypeAnnotation or Identifier node in place
// when its name matches a generic type parameter.
func applySubst(node any, subst map[string]ast.TypeExpression) {
	switch n := node.(type) {
	case *ast.TypeAnnotation:
		if arg, ok := subst[ident.Normalize(n.Name)]; ok {
			applyTypeArg(n, arg)
		}
	case *ast.Identifier:
		if arg, ok := subst[ident.Normalize(n.Value)]; ok {
			n.Value = renderTypeName(arg)
		}
	}
}

// applyTypeArg replaces the type named by a type parameter with the concrete
// type argument, preserving any nested generic arguments so they can be
// specialized in turn.
func applyTypeArg(node *ast.TypeAnnotation, arg ast.TypeExpression) {
	cloned := cloneNode(reflect.ValueOf(arg), nil).Interface()
	if ta, ok := cloned.(*ast.TypeAnnotation); ok {
		node.Name = ta.Name
		node.TypeArgs = ta.TypeArgs
		node.InlineType = ta.InlineType
		return
	}
	if te, ok := cloned.(ast.TypeExpression); ok {
		node.Name = ""
		node.TypeArgs = nil
		node.InlineType = te
	}
}

// renderTypeName renders a type argument to its canonical name for substitution
// into identifier positions such as Default(T) or a type cast T(x).
func renderTypeName(arg ast.TypeExpression) string {
	if arg == nil {
		return ""
	}
	return strings.TrimSpace(arg.String())
}
