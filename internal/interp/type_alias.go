package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Type Alias Support
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

// GetAliasedType returns the underlying aliased type.
func (tv *TypeAliasValue) GetAliasedType() types.Type {
	return tv.AliasedType
}

// ============================================================================
// Subrange Type Support
// ============================================================================

// A type alias is provided in value.go for backward compatibility.
// The types.SubrangeType struct provides all needed metadata directly.

// evalTypeDeclaration evaluates a type declaration
// Handles type aliases: type TUserID = Integer;
// Handles subrange types: type TDigit = 0..9;
func (i *Interpreter) evalTypeDeclaration(decl *ast.TypeDeclaration) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil type declaration"}
	}

	// Handle subrange types
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

		// Register in TypeSystem (replaces environment-based storage)
		i.typeSystem.RegisterSubrangeType(decl.Name.Value, subrangeType)

		return &NilValue{}
	}

	// Handle function pointer type declarations
	if decl.IsFunctionPointer {
		if decl.FunctionPointerType == nil {
			return &ErrorValue{Message: "function pointer type declaration has no type information"}
		}

		// Extract parameter types
		paramTypes := make([]types.Type, len(decl.FunctionPointerType.Parameters))
		for idx, param := range decl.FunctionPointerType.Parameters {
			if param.Type != nil {
				paramTypes[idx] = i.getTypeByName(param.Type.String())
			} else {
				paramTypes[idx] = &types.IntegerType{} // Default
			}
		}

		// Get return type
		var returnType types.Type
		if decl.FunctionPointerType.ReturnType != nil {
			returnType = i.getTypeByName(decl.FunctionPointerType.ReturnType.String())
		}

		// Create the function pointer type (for potential future use)
		// Currently we just register the type name as existing
		_ = paramTypes // Mark as used
		_ = returnType // Mark as used

		// Store the type name mapping for type resolution
		// We just need to register that this type name exists
		// The actual type checking is done by the semantic analyzer
		typeKey := "__funcptr_type_" + decl.Name.Value
		// Store a simple marker that this is a function pointer type
		i.env.Define(typeKey, &StringValue{Value: "function_pointer_type"})

		return &NilValue{}
	}

	// Handle type aliases
	if decl.IsAlias {
		// Check for inline/complex type expressions
		// For inline types, we don't need to resolve them at runtime because:
		// 1. The semantic analyzer already validated the types during analysis phase
		// 2. Inline type aliases are purely semantic constructs (no runtime storage needed)
		// 3. The interpreter's resolveType() doesn't support complex inline syntax
		switch decl.AliasedType.(type) {
		case *ast.ClassOfTypeNode:
			// Metaclass types (class of TBase) - semantic analyzer handles them
			return &NilValue{}
		case *ast.SetTypeNode:
			// Set types (set of TEnum) - semantic analyzer handles them
			return &NilValue{}
		case *ast.ArrayTypeNode:
			// Inline array types (array of Integer, array[1..10] of String)
			// Note: These could potentially need runtime storage, but semantic analyzer
			// already validated and stored them. The interpreter can resolve simple
			// "array of X" syntax via parseInlineArrayType if needed.
			return &NilValue{}
		case *ast.FunctionPointerTypeNode:
			// Function pointer types - already handled earlier in this function
			return &NilValue{}
		}

		// For TypeAnnotation with InlineType, also skip runtime resolution
		if typeAnnot, ok := decl.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
			// TypeAnnotation wrapping an inline type expression
			return &NilValue{}
		}

		// Resolve the aliased type by name (handles simple named types only)
		aliasedType, err := i.resolveType(decl.AliasedType.String())
		if err != nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' in type alias", decl.AliasedType.String())}
		}

		// Create TypeAliasValue and register it
		typeAlias := &TypeAliasValue{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		// If this alias targets an enum, register it with the type system so
		// scoped accesses like Alias.Value work the same as the original type.
		if enumType, ok := aliasedType.(*types.EnumType); ok {
			if i.typeSystem != nil {
				i.typeSystem.RegisterEnumType(decl.Name.Value, &EnumTypeValue{EnumType: enumType})
			}
		}

		// Store in environment with special prefix
		typeKey := "__type_alias_" + strings.ToLower(decl.Name.Value)
		i.env.Define(typeKey, typeAlias)

		// Also expose the alias name as a type meta value so it can be used
		// in expressions (e.g., scoped enum access via the alias name).
		i.env.Define(decl.Name.Value, NewTypeMetaValue(aliasedType, decl.Name.Value))

		return &NilValue{}
	}

	// Non-alias type declarations (future)
	return &ErrorValue{Message: "non-alias type declarations not yet supported"}
}
