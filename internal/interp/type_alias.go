package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Alias Support (Task 9.21)
// ============================================================================

// TypeAliasValue stores a type alias definition
type TypeAliasValue struct {
	AliasedType types.Type
	Name        string
}

func (tv *TypeAliasValue) Type() string {
	return "TYPE_ALIAS"
}

func (tv *TypeAliasValue) String() string {
	return fmt.Sprintf("type %s = %s", tv.Name, tv.AliasedType.String())
}

// ============================================================================
// Subrange Type Support (Task 9.100)
// ============================================================================

// SubrangeTypeValue stores a subrange type definition
type SubrangeTypeValue struct {
	SubrangeType *types.SubrangeType
	Name         string
}

func (sv *SubrangeTypeValue) Type() string {
	return "SUBRANGE_TYPE"
}

func (sv *SubrangeTypeValue) String() string {
	return fmt.Sprintf("type %s = %d..%d", sv.Name, sv.SubrangeType.LowBound, sv.SubrangeType.HighBound)
}

// SubrangeValue wraps an integer value with subrange bounds checking.
// Task 9.100: Runtime subrange validation
type SubrangeValue struct {
	Value        int
	SubrangeType *types.SubrangeType
}

func (sv *SubrangeValue) Type() string {
	return sv.SubrangeType.Name
}

func (sv *SubrangeValue) String() string {
	return fmt.Sprintf("%d", sv.Value)
}

// ValidateAndSet checks if a value is within bounds and updates the subrange value.
// Returns an error if the value is out of range.
func (sv *SubrangeValue) ValidateAndSet(value int) error {
	if err := types.ValidateRange(value, sv.SubrangeType); err != nil {
		return err
	}
	sv.Value = value
	return nil
}

// evalTypeDeclaration evaluates a type declaration (Task 9.21, Task 9.100)
// Handles type aliases: type TUserID = Integer;
// Handles subrange types: type TDigit = 0..9; (Task 9.100)
func (i *Interpreter) evalTypeDeclaration(decl *ast.TypeDeclaration) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil type declaration"}
	}

	// Task 9.100: Handle subrange types
	if decl.IsSubrange {
		// Evaluate low bound
		lowBoundVal := i.Eval(decl.LowBound)
		if isError(lowBoundVal) {
			return lowBoundVal
		}
		lowBoundIntVal, ok := lowBoundVal.(*IntegerValue)
		if !ok {
			return &ErrorValue{Message: "subrange low bound must be an integer"}
		}
		lowBoundInt := int(lowBoundIntVal.Value)

		// Evaluate high bound
		highBoundVal := i.Eval(decl.HighBound)
		if isError(highBoundVal) {
			return highBoundVal
		}
		highBoundIntVal, ok := highBoundVal.(*IntegerValue)
		if !ok {
			return &ErrorValue{Message: "subrange high bound must be an integer"}
		}
		highBoundInt := int(highBoundIntVal.Value)

		// Validate bounds
		if lowBoundInt > highBoundInt {
			return &ErrorValue{Message: fmt.Sprintf("subrange low bound (%d) cannot be greater than high bound (%d)", lowBoundInt, highBoundInt)}
		}

		// Create SubrangeType
		subrangeType := &types.SubrangeType{
			BaseType:  types.INTEGER,
			Name:      decl.Name.Value,
			LowBound:  lowBoundInt,
			HighBound: highBoundInt,
		}

		// Create SubrangeTypeValue and register it
		subrangeTypeValue := &SubrangeTypeValue{
			Name:         decl.Name.Value,
			SubrangeType: subrangeType,
		}

		// Store in environment with special prefix
		typeKey := "__subrange_type_" + decl.Name.Value
		i.env.Define(typeKey, subrangeTypeValue)

		return &NilValue{}
	}

	// Handle type aliases
	if decl.IsAlias {
		// Resolve the aliased type
		aliasedType, err := i.resolveType(decl.AliasedType.Name)
		if err != nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' in type alias", decl.AliasedType.Name)}
		}

		// Create TypeAliasValue and register it
		typeAlias := &TypeAliasValue{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		// Store in environment with special prefix
		typeKey := "__type_alias_" + decl.Name.Value
		i.env.Define(typeKey, typeAlias)

		return &NilValue{}
	}

	// Non-alias type declarations (future)
	return &ErrorValue{Message: "non-alias type declarations not yet supported"}
}
