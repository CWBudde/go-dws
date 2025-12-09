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

// buildClassTypeHierarchy constructs a types.ClassType (with parent chain)
// from the runtime ClassInfo registered in the interpreter. This lets us
// expose "class of" aliases as real types that downstream alias resolution
// can see.
func (i *Interpreter) buildClassTypeHierarchy(info *ClassInfo) *types.ClassType {
	if info == nil {
		return nil
	}

	var parentType *types.ClassType
	if info.Parent != nil {
		parentType = i.buildClassTypeHierarchy(info.Parent)
	}

	return types.NewClassType(info.Name, parentType)
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

		// Create function or method pointer type and register in TypeSystem
		var funcPtrType types.Type
		if decl.FunctionPointerType.OfObject {
			funcPtrType = types.NewMethodPointerType(paramTypes, returnType)
		} else {
			funcPtrType = types.NewFunctionPointerType(paramTypes, returnType)
		}
		i.typeSystem.RegisterFunctionPointerType(decl.Name.Value, funcPtrType)

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
		var aliasedType types.Type
		var resolveErr error

		// Check for inline/complex type expressions that need special handling
		switch t := decl.AliasedType.(type) {
		case *ast.ClassOfTypeNode:
			// Metaclass types (class of TBase)
			baseClassName := ""
			if t.ClassType != nil {
				baseClassName = t.ClassType.String()
			}
			classInfo := i.resolveClassInfoByName(baseClassName)
			classType := i.buildClassTypeHierarchy(classInfo)
			if classType == nil && i.typeSystem != nil && i.typeSystem.HasClass(baseClassName) {
				// Fall back to nominal class type if hierarchy info isn't available
				classType = types.NewClassType(baseClassName, nil)
			}
			if classType == nil {
				return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' in type alias", baseClassName)}
			}
			aliasedType = types.NewClassOfType(classType)
		case *ast.SetTypeNode:
			// Set types (set of TEnum) - semantic analyzer handles them
			return &NilValue{}
		case *ast.ArrayTypeNode:
			// Inline array types (array of Integer, array[1..10] of String)
			// Need to resolve and store these so helpers can target array type aliases
			aliasedType = i.resolveArrayTypeNode(t)
			if aliasedType == nil {
				return &ErrorValue{Message: fmt.Sprintf("cannot resolve array type in alias '%s'", decl.Name.Value)}
			}
		case *ast.FunctionPointerTypeNode:
			// Function pointer types - already handled earlier in this function
			return &NilValue{}
		default:
			// For TypeAnnotation with InlineType, handle specially
			if typeAnnot, ok := decl.AliasedType.(*ast.TypeAnnotation); ok && typeAnnot.InlineType != nil {
				// TypeAnnotation wrapping an inline type expression
				return &NilValue{}
			}

			// Resolve the aliased type by name (handles simple named types only)
			aliasedType, resolveErr = i.resolveType(decl.AliasedType.String())
			if resolveErr != nil {
				return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' in type alias", decl.AliasedType.String())}
			}
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
