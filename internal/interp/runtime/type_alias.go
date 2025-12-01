package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
)

// TypeAliasValue stores a type alias definition.
// Task 3.5.16: Migrated from internal/interp/type_alias.go to runtime package
// to enable evaluator package to work with type aliases directly.
//
// Type aliases are defined with syntax like:
//
//	type TUserID = Integer;
//	type TMyString = String;
//
// The alias stores the underlying type it represents, allowing the interpreter
// to resolve the actual type when the alias name is used in expressions.
type TypeAliasValue struct {
	AliasedType types.Type
	Name        string
}

// Type returns "TYPE_ALIAS".
func (tv *TypeAliasValue) Type() string {
	return "TYPE_ALIAS"
}

// String returns a string representation of the type alias.
func (tv *TypeAliasValue) String() string {
	return fmt.Sprintf("type %s = %s", tv.Name, tv.AliasedType.String())
}

// GetAliasedType returns the underlying aliased type.
// This method is used by the type resolution system to look up the actual type.
func (tv *TypeAliasValue) GetAliasedType() types.Type {
	return tv.AliasedType
}
