package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Type Resolution, Conversion, and Default Value Helpers
// ============================================================================

// getDefaultValue returns the default/zero value for a given type.
// This is used for Result variable initialization in functions.
func (i *Interpreter) getDefaultValue(typ types.Type) Value {
	if typ == nil {
		return &NilValue{}
	}

	switch typ.TypeKind() {
	case "STRING":
		return &StringValue{Value: ""}
	case "INTEGER":
		return &IntegerValue{Value: 0}
	case "FLOAT":
		return &FloatValue{Value: 0.0}
	case "BOOLEAN":
		return &BooleanValue{Value: false}
	case "CLASS", "INTERFACE", "FUNCTION_POINTER", "METHOD_POINTER":
		return &NilValue{}
	case "ARRAY":
		// Arrays should default to an empty array value of the correct element type.
		// If we can resolve the array type, create an empty array; otherwise fall back to nil.
		if arrType, ok := typ.(*types.ArrayType); ok {
			return NewArrayValue(arrType)
		}
		return &NilValue{}
	case "RECORD":
		// Records should be initialized with default field values
		// For now, return NIL (will be enhanced in future tasks if needed)
		return &NilValue{}
	default:
		// Unknown types default to NIL
		return &NilValue{}
	}
}

// resolveTypeFromExpression resolves a type from any TypeExpression.
func (i *Interpreter) resolveTypeFromExpression(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	// For simple type annotations, delegate to existing function
	if typeAnnot, ok := typeExpr.(*ast.TypeAnnotation); ok {
		return i.resolveTypeFromAnnotation(typeAnnot)
	}

	// For array types, resolve the element type and construct an array type
	if arrayType, ok := typeExpr.(*ast.ArrayTypeNode); ok {
		elementType := i.resolveTypeFromExpression(arrayType.ElementType)
		if elementType == nil {
			return nil
		}

		// Evaluate bound expressions if this is a static array
		if arrayType.IsDynamic() {
			return types.NewDynamicArrayType(elementType)
		}

		// Evaluate low bound
		lowBoundVal := i.Eval(arrayType.LowBound)
		if isError(lowBoundVal) {
			return nil
		}
		lowBound, ok := lowBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		// Evaluate high bound
		highBoundVal := i.Eval(arrayType.HighBound)
		if isError(highBoundVal) {
			return nil
		}
		highBound, ok := highBoundVal.(*IntegerValue)
		if !ok {
			return nil
		}

		return types.NewStaticArrayType(elementType, int(lowBound.Value), int(highBound.Value))
	}

	// For function pointer types, we need full type information
	// For now, return a generic function type placeholder
	if fpt, ok := typeExpr.(*ast.FunctionPointerTypeNode); ok {
		paramTypes := make([]types.Type, len(fpt.Parameters))
		for idx, p := range fpt.Parameters {
			paramTypes[idx] = i.resolveTypeFromExpression(p.Type)
			if paramTypes[idx] == nil {
				return nil
			}
		}

		var returnType types.Type
		if fpt.ReturnType != nil {
			returnType = i.resolveTypeFromExpression(fpt.ReturnType)
			if returnType == nil {
				return nil
			}
		}

		return types.NewFunctionType(paramTypes, returnType)
	}

	return nil
}

// resolveTypeFromAnnotation resolves a type from an AST TypeAnnotation
func (i *Interpreter) resolveTypeFromAnnotation(typeExpr ast.TypeExpression) types.Type {
	if typeExpr == nil {
		return nil
	}

	typeName := typeExpr.String()

	// Normalize type name for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := ident.Normalize(typeName)

	// Check basic types (case-insensitive)
	switch lowerTypeName {
	case "integer":
		return types.INTEGER
	case "float":
		return types.FLOAT
	case "string":
		return types.STRING
	case "boolean":
		return types.BOOLEAN
	case "const":
		// Migrate Const to Variant for proper dynamic typing
		// "Const" was a temporary workaround, now redirects to VARIANT
		return types.VARIANT
	case "variant":
		// Support Variant type for dynamic values
		return types.VARIANT
	}

	// Check for class types (supports nested classes via current class context)
	if classInfo := i.resolveClassInfoByName(typeName); classInfo != nil {
		return types.NewClassType(classInfo.Name, nil)
	}

	// Check for interface types
	if interfaceInfo := i.lookupInterfaceInfo(typeName); interfaceInfo != nil {
		return types.NewInterfaceType(interfaceInfo.Name)
	}

	// Check for record types (stored with special prefix in environment)
	recordTypeKey := "__record_type_" + ident.Normalize(typeName)
	if typeVal, ok := i.Env().Get(recordTypeKey); ok {
		if recordTypeVal, ok := typeVal.(*RecordTypeValue); ok {
			return recordTypeVal.RecordType
		}
	}

	// Check for type aliases (e.g., TPointArray = array of TPoint)
	typeAliasKey := "__type_alias_" + lowerTypeName
	if typeAliasVal, ok := i.Env().Get(typeAliasKey); ok {
		if typeAlias, ok := typeAliasVal.(*TypeAliasValue); ok {
			return typeAlias.AliasedType
		}
		if typeAlias, ok := typeAliasVal.(*runtime.TypeAliasValue); ok {
			return typeAlias.AliasedType
		}
	}

	// Type not found
	return nil
}

// resolveClassInfoByName looks up a class by name, handling both global and nested classes.
func (i *Interpreter) resolveClassInfoByName(name string) *ClassInfo {
	if current := i.currentClassContext(); current != nil {
		if nested := current.lookupNestedClass(name); nested != nil {
			return nested
		}
	}

	return i.lookupRegisteredClassInfo(name)
}

// currentClassContext inspects the execution environment to find the current class scope.
func (i *Interpreter) currentClassContext() *ClassInfo {
	if val, ok := i.Env().Get("__CurrentClass__"); ok {
		if classVal, ok := val.(*ClassInfoValue); ok {
			return classVal.ClassInfo
		}
	}
	if val, ok := i.Env().Get("Self"); ok {
		if classVal, ok := val.(*ClassInfoValue); ok {
			return classVal.ClassInfo
		}
		if obj, ok := AsObject(val); ok {
			concreteClass, ok := obj.Class.(*ClassInfo)
			if ok {
				return concreteClass
			}
		}
	}
	return nil
}

// extractSimpleTypeName extracts the simple type name from a full type string
// Examples:
//   - "array of Integer" -> "Integer"
//   - "array[0..10] of String" -> "String"
//   - "MyClass(ParentClass)" -> "MyClass"
//   - "class of MyClass" -> "MyClass"
//   - "Integer" -> "Integer"
func extractSimpleTypeName(typeName string) string {
	// Handle array types: "array of Integer" or "array[0..10] of String"
	if strings.HasPrefix(typeName, "array") {
		if idx := strings.Index(typeName, " of "); idx != -1 {
			return typeName[idx+4:] // Extract everything after " of "
		}
	}

	// Handle metaclass types: "class of MyClass"
	if strings.HasPrefix(typeName, "class of ") {
		return typeName[9:] // Extract everything after "class of "
	}

	// Handle class types with parent: "MyClass(ParentClass)"
	if idx := strings.Index(typeName, "("); idx != -1 {
		return typeName[:idx] // Extract everything before "("
	}

	// Already a simple type name
	return typeName
}
