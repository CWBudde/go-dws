package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Record Declaration Evaluation (Task 8.73)
// ============================================================================

// evalRecordDeclaration evaluates a record type declaration.
// It registers the record type in the interpreter's symbol table.
func (i *Interpreter) evalRecordDeclaration(decl *ast.RecordDecl) Value {
	if decl == nil {
		return &ErrorValue{Message: "nil record declaration"}
	}

	recordName := decl.Name.Value

	// Build the record type from the declaration
	fields := make(map[string]types.Type)

	for _, field := range decl.Fields {
		fieldName := field.Name.Value
		fieldTypeName := field.Type.Name

		// Resolve field type
		fieldType, err := i.resolveType(fieldTypeName)
		if err != nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown type '%s' for field '%s' in record '%s'", fieldTypeName, fieldName, recordName)}
		}

		fields[fieldName] = fieldType
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, fields)

	// TODO: Handle methods and properties if needed (Task 8.78)

	// Store record type metadata in environment with special key
	// This allows variable declarations to resolve the type
	recordTypeKey := "__record_type_" + recordName
	i.env.Define(recordTypeKey, &RecordTypeValue{RecordType: recordType})

	return &NilValue{}
}

// RecordTypeValue is an internal value type used to store record type metadata
// in the interpreter's environment.
type RecordTypeValue struct {
	RecordType *types.RecordType
}

// Type returns "RECORD_TYPE".
func (r *RecordTypeValue) Type() string {
	return "RECORD_TYPE"
}

// String returns the record type name.
func (r *RecordTypeValue) String() string {
	return r.RecordType.Name
}

// ============================================================================
// Record Literal Evaluation (Task 8.74)
// ============================================================================

// evalRecordLiteral evaluates a record literal expression.
// Examples: (X: 10, Y: 20) or TPoint(X: 10, Y: 20)
func (i *Interpreter) evalRecordLiteral(literal *ast.RecordLiteral) Value {
	if literal == nil {
		return &ErrorValue{Message: "nil record literal"}
	}

	// We need to determine the record type from context
	// For now, we'll require explicit type name in the literal or get it from variable declaration
	// This is a simplified implementation - a full implementation would use type inference from context

	var recordType *types.RecordType

	// If the literal has an explicit type name, use it
	if literal.TypeName != "" {
		recordTypeKey := "__record_type_" + literal.TypeName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				recordType = rtv.RecordType
			}
		}

		if recordType == nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown record type '%s'", literal.TypeName)}
		}
	} else {
		// For untyped literals, we need to get the type from context
		// This is handled during assignment - we'll store the type requirement in a temp variable
		// For now, return an error
		return &ErrorValue{Message: "record literal requires explicit type name or type context"}
	}

	// Create the record value
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
	}

	// Evaluate and assign field values
	for _, field := range literal.Fields {
		fieldName := field.Name

		// Check if field exists in record type
		if _, exists := recordType.Fields[fieldName]; !exists {
			return &ErrorValue{Message: fmt.Sprintf("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)}
		}

		// Evaluate the field value expression
		fieldValue := i.Eval(field.Value)
		if isError(fieldValue) {
			return fieldValue
		}

		// Store the field value
		recordValue.Fields[fieldName] = fieldValue
	}

	// Check that all required fields are initialized
	for fieldName := range recordType.Fields {
		if _, exists := recordValue.Fields[fieldName]; !exists {
			return &ErrorValue{Message: fmt.Sprintf("missing required field '%s' in record literal", fieldName)}
		}
	}

	return recordValue
}

// resolveType resolves a type name to a types.Type
// This is a helper for record field type resolution
func (i *Interpreter) resolveType(typeName string) (types.Type, error) {
	switch typeName {
	case "Integer":
		return types.INTEGER, nil
	case "Float":
		return types.FLOAT, nil
	case "String":
		return types.STRING, nil
	case "Boolean":
		return types.BOOLEAN, nil
	default:
		// Check for custom types (enums, records, arrays)
		// Try enum type
		if enumTypeVal, ok := i.env.Get("__enum_type_" + typeName); ok {
			if etv, ok := enumTypeVal.(*EnumTypeValue); ok {
				return etv.EnumType, nil
			}
		}
		// Try record type
		if recordTypeVal, ok := i.env.Get("__record_type_" + typeName); ok {
			if rtv, ok := recordTypeVal.(*RecordTypeValue); ok {
				return rtv.RecordType, nil
			}
		}
		// Try array type
		if arrayTypeVal, ok := i.env.Get("__array_type_" + typeName); ok {
			if atv, ok := arrayTypeVal.(*ArrayTypeValue); ok {
				return atv.ArrayType, nil
			}
		}
		// Try type alias (Task 9.21)
		if typeAliasVal, ok := i.env.Get("__type_alias_" + typeName); ok {
			if tav, ok := typeAliasVal.(*TypeAliasValue); ok {
				// Return the underlying type (type aliases are transparent at runtime)
				return tav.AliasedType, nil
			}
		}
		// Unknown type
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// ============================================================================
// Record Comparison (Task 8.79)
// ============================================================================

// evalRecordBinaryOp evaluates binary operations on records.
// Currently supports = (equality) and <> (inequality).
func (i *Interpreter) evalRecordBinaryOp(op string, left, right Value) Value {
	leftRecord, ok := left.(*RecordValue)
	if !ok {
		return &ErrorValue{Message: fmt.Sprintf("expected record, got %s", left.Type())}
	}
	rightRecord, ok := right.(*RecordValue)
	if !ok {
		return &ErrorValue{Message: fmt.Sprintf("expected record, got %s", right.Type())}
	}

	switch op {
	case "=":
		return &BooleanValue{Value: i.recordsEqual(leftRecord, rightRecord)}
	case "<>":
		return &BooleanValue{Value: !i.recordsEqual(leftRecord, rightRecord)}
	default:
		return &ErrorValue{Message: fmt.Sprintf("unsupported operator '%s' for records", op)}
	}
}

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
