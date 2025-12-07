package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Task 3.5.10: RecordTypeValue moved to evaluator package.
// This type alias provides backward compatibility for code in interp package.
type RecordTypeValue = evaluator.RecordTypeValue

// ============================================================================
// Helper Functions
// ============================================================================

// initializeInterfaceField creates an InterfaceInstance for interface-typed fields.
// Task 9.1.4: Helper to reduce code duplication in record field initialization.
// Task 3.5.184: Use TypeSystem lookup instead of i.interfaces map.
func (i *Interpreter) initializeInterfaceField(fieldType types.Type) Value {
	if interfaceType, ok := fieldType.(*types.InterfaceType); ok {
		// Look up the InterfaceInfo from the TypeSystem
		if interfaceInfo := i.lookupInterfaceInfo(interfaceType.Name); interfaceInfo != nil {
			return &InterfaceInstance{
				Interface: interfaceInfo,
				Object:    nil,
			}
		}
	}
	return nil
}

// ============================================================================
// Record Declaration Evaluation
// ============================================================================

// Task 3.5.10: evalRecordDeclaration removed - migrated to evaluator.VisitRecordDecl().
// Task 3.10.3: Delegation completed in Phase 3.10.

// Task 3.5.10: RecordTypeValue struct and its methods moved to internal/interp/evaluator/record_type_value.go
// A type alias is provided at the top of this file for backward compatibility.

// createRecordValue creates a new RecordValue with proper method lookup for nested records.
// Task 9.7e1: Helper to create records with method resolution for nested record fields.
// Task 9.5: Initialize fields with field initializers.
// Task 3.5.128a: Removed deprecated methods parameter - now uses only Metadata.
func (i *Interpreter) createRecordValue(recordType *types.RecordType) Value {
	// Task 9.5: Look up the record type value to get field declarations before creating the value
	recordTypeKey := "__record_type_" + ident.Normalize(recordType.Name)
	var rtv *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		rtv, _ = typeVal.(*RecordTypeValue)
	}

	// Task 3.5.42: Extract metadata from RecordTypeValue if available
	var metadata *runtime.RecordMetadata
	if rtv != nil {
		metadata = rtv.Metadata
	}

	// Create a metadata lookup callback for nested records
	// Task 3.5.128a: This replaces the old methodsLookup callback
	metadataLookup := func(rt *types.RecordType) *runtime.RecordMetadata {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if nestedRtv, ok := typeVal.(*RecordTypeValue); ok {
				return nestedRtv.Metadata
			}
		}
		return nil
	}

	// Create the record value with metadata lookup for nested records
	rv := newRecordValueInternalWithMetadataLookup(recordType, metadata, metadataLookup)

	// Task 9.5: Initialize fields with field initializers or default values
	if rtv != nil {
		for fieldName, fieldType := range recordType.Fields {
			var fieldValue Value

			// Check if field has an initializer expression
			if fieldDecl, hasDecl := rtv.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				// Evaluate the field initializer
				fieldValue = i.Eval(fieldDecl.InitValue)
				if isError(fieldValue) {
					return fieldValue
				}
			} else {
				// Task 3.5.128a: Handle nested record types specially
				// Use createRecordValue recursively to ensure metadata is properly set
				if nestedRecordType, ok := fieldType.(*types.RecordType); ok {
					fieldValue = i.createRecordValue(nestedRecordType)
				} else {
					// Use getZeroValueForType for other types
					fieldValue = getZeroValueForType(fieldType, nil)
				}

				// Task 9.1.4: Handle interface-typed fields specially
				// Interface fields should be initialized as InterfaceInstance with nil object
				// This allows proper "Interface is nil" error messages when accessed
				if intfValue := i.initializeInterfaceField(fieldType); intfValue != nil {
					fieldValue = intfValue
				}
			}

			rv.Fields[fieldName] = fieldValue
		}
	}

	return rv
}

// ============================================================================
// Record Literal Evaluation
// ============================================================================

// evalRecordLiteral evaluates a record literal expression.
// Examples: (X: 10, Y: 20) or TPoint(X: 10, Y: 20)
func (i *Interpreter) evalRecordLiteral(literal *ast.RecordLiteralExpression) Value {
	if literal == nil {
		return &ErrorValue{Message: "nil record literal"}
	}

	// We need to determine the record type from context
	// For now, we'll require explicit type name in the literal or get it from variable declaration
	// This is a simplified implementation - a full implementation would use type inference from context

	var recordType *types.RecordType

	// If the literal has an explicit type name, use it
	if literal.TypeName != nil {
		typeName := literal.TypeName.Value
		recordTypeKey := "__record_type_" + ident.Normalize(typeName)
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				recordType = rtv.RecordType
			}
		}

		if recordType == nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown record type '%s'", typeName)}
		}
	} else {
		// For untyped literals, we need to get the type from context
		// This is handled during assignment - we'll store the type requirement in a temp variable
		// For now, return an error
		return &ErrorValue{Message: "record literal requires explicit type name or type context"}
	}

	// Look up the record type value to get field declarations
	recordTypeKey := "__record_type_" + ident.Normalize(literal.TypeName.Value)
	var recordTypeValue *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		recordTypeValue, _ = typeVal.(*RecordTypeValue)
	}

	// Task 9.12.4: Create the record value with methods
	// Task 3.5.42: Updated to use RecordMetadata
	// Task 3.5.128a: Removed deprecated Methods field
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
		Metadata:   nil, // Will be set below if recordTypeValue is available
	}

	// Set metadata if available
	if recordTypeValue != nil {
		recordValue.Metadata = recordTypeValue.Metadata
	}

	// Evaluate and assign field values from literal
	for _, field := range literal.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			return &ErrorValue{Message: "positional record field initialization not yet supported"}
		}

		fieldName := field.Name.Value
		fieldNameLower := ident.Normalize(fieldName)

		// Check if field exists in record type (use lowercase key)
		if _, exists := recordType.Fields[fieldNameLower]; !exists {
			return &ErrorValue{Message: fmt.Sprintf("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)}
		}

		// Evaluate the field value expression
		fieldValue := i.Eval(field.Value)
		if isError(fieldValue) {
			return fieldValue
		}

		// Store the field value (use lowercase key)
		recordValue.Fields[fieldNameLower] = fieldValue
	}

	// Task 9.5: Initialize remaining fields with field initializers or default values
	// Create a method lookup callback for nested records
	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	for fieldName, fieldType := range recordType.Fields {
		if _, exists := recordValue.Fields[fieldName]; !exists {
			var fieldValue Value

			// Check if field has an initializer expression
			if recordTypeValue != nil {
				if fieldDecl, hasDecl := recordTypeValue.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
					// Evaluate the field initializer
					fieldValue = i.Eval(fieldDecl.InitValue)
					if isError(fieldValue) {
						return fieldValue
					}
				}
			}

			// If no initializer, use getZeroValueForType to properly initialize nested records
			if fieldValue == nil {
				fieldValue = getZeroValueForType(fieldType, methodsLookup)

				// Task 9.1.4: Handle interface-typed fields specially
				// Interface fields should be initialized as InterfaceInstance with nil object
				// This allows proper "Interface is nil" error messages when accessed
				if intfValue := i.initializeInterfaceField(fieldType); intfValue != nil {
					fieldValue = intfValue
				}
			}

			recordValue.Fields[fieldName] = fieldValue
		}
	}

	return recordValue
}

// resolveType resolves a type name to a types.Type
// This is a helper for record field type resolution
func (i *Interpreter) resolveType(typeName string) (types.Type, error) {
	// Task 9.56: Check for inline array types first
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType := i.parseInlineArrayType(typeName)
		if arrayType != nil {
			return arrayType, nil
		}
		return nil, fmt.Errorf("invalid inline array type: %s", typeName)
	}

	// Strip parent qualification from class type strings like "TSub(TBase)"
	// to enable runtime resolution using the declared class name.
	cleanTypeName := typeName
	if idx := strings.Index(cleanTypeName, "("); idx != -1 {
		cleanTypeName = strings.TrimSpace(cleanTypeName[:idx])
	}

	// Normalize type name to lowercase for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := ident.Normalize(cleanTypeName)

	switch lowerTypeName {
	case "integer":
		return types.INTEGER, nil
	case "float":
		return types.FLOAT, nil
	case "string":
		return types.STRING, nil
	case "boolean":
		return types.BOOLEAN, nil
	case "const":
		// Task 9.235: Migrate Const to Variant for proper dynamic typing
		// "Const" was a temporary workaround, now redirects to VARIANT
		return types.VARIANT, nil
	case "variant":
		// Task 9.227: Support Variant type for dynamic values
		return types.VARIANT, nil
	default:
		// Check for custom types (enums, records, arrays, subranges)
		// Task 9.225: Use lowerTypeName for case-insensitive lookups
		// Try enum type via TypeSystem (Task 3.5.143b)
		if enumMetadata := i.typeSystem.LookupEnumMetadata(typeName); enumMetadata != nil {
			if etv, ok := enumMetadata.(*EnumTypeValue); ok {
				return etv.EnumType, nil
			}
		}
		// Try record type via TypeSystem
		// Task 3.5.22b: Use TypeSystem registry instead of i.env.Get()
		// This fixes the issue where i.env is the caller's environment in ExecuteUserFunction
		if recordTypeValueAny := i.typeSystem.LookupRecord(typeName); recordTypeValueAny != nil {
			if rtv, ok := recordTypeValueAny.(*RecordTypeValue); ok {
				return rtv.RecordType, nil
			}
		}
		// Try array type (Task 3.5.69c: use TypeSystem)
		if arrayType := i.typeSystem.LookupArrayType(typeName); arrayType != nil {
			return arrayType, nil
		}
		// Try type alias
		if typeAliasVal, ok := i.env.Get("__type_alias_" + lowerTypeName); ok {
			if tav, ok := typeAliasVal.(*TypeAliasValue); ok {
				// Return the underlying type (type aliases are transparent at runtime)
				return tav.AliasedType, nil
			}
		}
		// Try subrange type
		// Task 3.5.182: Use TypeSystem for subrange type lookup
		if subrangeType := i.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
			return subrangeType, nil
		}
		// Try class type via TypeSystem
		if i.typeSystem != nil && i.typeSystem.HasClass(cleanTypeName) {
			// Use nominal class type for runtime type information
			return types.NewClassType(cleanTypeName, nil), nil
		}

		// Function/method pointer types registered in the TypeSystem
		if funcPtrType := i.typeSystem.LookupFunctionPointerType(cleanTypeName); funcPtrType != nil {
			return funcPtrType, nil
		}

		// Unknown type
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// ============================================================================
// Record Comparison
// ============================================================================

// recordsEqual checks if two records are equal by comparing all fields.
func (i *Interpreter) recordsEqual(left, right *RecordValue) bool {
	// Different types are not equal
	if left.RecordType.Name != right.RecordType.Name {
		return false
	}

	// Check if all fields are equal
	for fieldName := range left.RecordType.Fields {
		leftVal, leftExists := left.Fields[fieldName]
		rightVal, rightExists := right.Fields[fieldName]

		// Both should exist
		if !leftExists || !rightExists {
			return false
		}

		// Compare field values using the existing valuesEqual method
		if !i.valuesEqual(leftVal, rightVal) {
			return false
		}
	}

	return true
}
