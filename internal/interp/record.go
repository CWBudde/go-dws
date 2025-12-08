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

// RecordTypeValue is defined in the evaluator package.
type RecordTypeValue = evaluator.RecordTypeValue

// ============================================================================
// Helper Functions
// ============================================================================

// initializeInterfaceField creates an InterfaceInstance for interface-typed fields.
func (i *Interpreter) initializeInterfaceField(fieldType types.Type) Value {
	if interfaceType, ok := fieldType.(*types.InterfaceType); ok {
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

// createRecordValue creates a new RecordValue with proper initialization.
// Supports nested records, field initializers, and interface-typed fields.
func (i *Interpreter) createRecordValue(recordType *types.RecordType) Value {
	// Look up the record type value for field declarations
	recordTypeKey := "__record_type_" + ident.Normalize(recordType.Name)
	var rtv *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		rtv, _ = typeVal.(*RecordTypeValue)
	}

	// Extract metadata if available
	var metadata *runtime.RecordMetadata
	if rtv != nil {
		metadata = rtv.Metadata
	}

	// Metadata lookup callback for nested records
	metadataLookup := func(rt *types.RecordType) *runtime.RecordMetadata {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if nestedRtv, ok := typeVal.(*RecordTypeValue); ok {
				return nestedRtv.Metadata
			}
		}
		return nil
	}

	// Create the record value with metadata lookup
	rv := newRecordValueInternalWithMetadataLookup(recordType, metadata, metadataLookup)

	// Initialize fields with initializers or default values
	if rtv != nil {
		for fieldName, fieldType := range recordType.Fields {
			var fieldValue Value

			// Evaluate field initializer if present
			if fieldDecl, hasDecl := rtv.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				fieldValue = i.Eval(fieldDecl.InitValue)
				if isError(fieldValue) {
					return fieldValue
				}
			} else {
				// Handle nested records recursively
				if nestedRecordType, ok := fieldType.(*types.RecordType); ok {
					fieldValue = i.createRecordValue(nestedRecordType)
				} else {
					fieldValue = getZeroValueForType(fieldType, nil)
				}

				// Initialize interface fields specially for nil checking
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

	var recordType *types.RecordType

	// Resolve the record type from the literal's type name
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
		return &ErrorValue{Message: "record literal requires explicit type name or type context"}
	}

	// Look up the record type value for field declarations
	recordTypeKey := "__record_type_" + ident.Normalize(literal.TypeName.Value)
	var recordTypeValue *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		recordTypeValue, _ = typeVal.(*RecordTypeValue)
	}

	// Create the record value
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
		Metadata:   nil,
	}

	// Set metadata if available
	if recordTypeValue != nil {
		recordValue.Metadata = recordTypeValue.Metadata
	}

	// Evaluate and assign field values from literal
	for _, field := range literal.Fields {
		if field.Name == nil {
			return &ErrorValue{Message: "positional record field initialization not yet supported"}
		}

		fieldName := field.Name.Value
		fieldNameLower := ident.Normalize(fieldName)

		// Verify field exists in record type
		if _, exists := recordType.Fields[fieldNameLower]; !exists {
			return &ErrorValue{Message: fmt.Sprintf("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)}
		}

		// Evaluate the field value expression
		fieldValue := i.Eval(field.Value)
		if isError(fieldValue) {
			return fieldValue
		}

		recordValue.Fields[fieldNameLower] = fieldValue
	}

	// Initialize remaining fields with initializers or defaults
	for fieldName, fieldType := range recordType.Fields {
		if _, exists := recordValue.Fields[fieldName]; !exists {
			var fieldValue Value

			// Evaluate field initializer if present
			if recordTypeValue != nil {
				if fieldDecl, hasDecl := recordTypeValue.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
					fieldValue = i.Eval(fieldDecl.InitValue)
					if isError(fieldValue) {
						return fieldValue
					}
				}
			}

			// Use default value if no initializer
			if fieldValue == nil {
				fieldValue = getZeroValueForType(fieldType, nil)

				// Initialize interface fields for nil checking
				if intfValue := i.initializeInterfaceField(fieldType); intfValue != nil {
					fieldValue = intfValue
				}
			}

			recordValue.Fields[fieldName] = fieldValue
		}
	}

	return recordValue
}

// resolveType resolves a type name to a types.Type.
// Handles built-in types, inline arrays, and custom types (enums, records, classes, etc.).
func (i *Interpreter) resolveType(typeName string) (types.Type, error) {
	// Check for inline array types first
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType := i.parseInlineArrayType(typeName)
		if arrayType != nil {
			return arrayType, nil
		}
		return nil, fmt.Errorf("invalid inline array type: %s", typeName)
	}

	// Inline function/method pointer types
	if lowerOrig := ident.Normalize(typeName); strings.HasPrefix(lowerOrig, "function(") || strings.HasPrefix(lowerOrig, "procedure(") {
		if funcPtrType, err := i.resolveInlineFunctionPointerType(typeName); err == nil {
			return funcPtrType, nil
		}
	}

	// Strip parent qualification from class type strings like "TSub(TBase)"
	cleanTypeName := typeName
	if idx := strings.Index(cleanTypeName, "("); idx != -1 {
		cleanTypeName = strings.TrimSpace(cleanTypeName[:idx])
	}

	// Normalize for case-insensitive comparison
	lowerTypeName := ident.Normalize(cleanTypeName)

	// Check built-in types
	switch lowerTypeName {
	case "integer":
		return types.INTEGER, nil
	case "float":
		return types.FLOAT, nil
	case "string":
		return types.STRING, nil
	case "boolean":
		return types.BOOLEAN, nil
	case "const", "variant":
		return types.VARIANT, nil
	}

	// Built-in metaclass type (class of TObject)
	if ident.Equal(lowerTypeName, "tclass") {
		if objClass := i.typeSystem.LookupClass("TObject"); objClass != nil {
			if ct, ok := objClass.(*types.ClassType); ok {
				return types.NewClassOfType(ct), nil
			}
		}
		return types.NewClassOfType(types.NewClassType("TObject", nil)), nil
	}

	// Check custom types via TypeSystem
	if enumMetadata := i.typeSystem.LookupEnumMetadata(typeName); enumMetadata != nil {
		if etv, ok := enumMetadata.(*EnumTypeValue); ok {
			return etv.EnumType, nil
		}
	}

	if recordTypeValueAny := i.typeSystem.LookupRecord(typeName); recordTypeValueAny != nil {
		if rtv, ok := recordTypeValueAny.(*RecordTypeValue); ok {
			return rtv.RecordType, nil
		}
	}

	if arrayType := i.typeSystem.LookupArrayType(typeName); arrayType != nil {
		return arrayType, nil
	}

	// Check type aliases
	if typeAliasVal, ok := i.env.Get("__type_alias_" + lowerTypeName); ok {
		if tav, ok := typeAliasVal.(*TypeAliasValue); ok {
			return tav.AliasedType, nil
		}
	}

	if subrangeType := i.typeSystem.LookupSubrangeType(typeName); subrangeType != nil {
		return subrangeType, nil
	}

	// Check class types
	if i.typeSystem != nil && i.typeSystem.HasClass(cleanTypeName) {
		return types.NewClassType(cleanTypeName, nil), nil
	}

	// Check function pointer types
	if funcPtrType := i.typeSystem.LookupFunctionPointerType(cleanTypeName); funcPtrType != nil {
		return funcPtrType, nil
	}

	return nil, fmt.Errorf("unknown type: %s", typeName)
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
