package interp

import (
	"fmt"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Record Declaration Evaluation
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
	// Task 9.5: Store field declarations for initializer access
	fieldDecls := make(map[string]*ast.FieldDecl)

	for _, field := range decl.Fields {
		fieldName := field.Name.Value

		// Resolve field type using type expression
		// Task 9.170.1: Updated to support inline array types
		fieldType := i.resolveTypeFromExpression(field.Type)
		if fieldType == nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown or invalid type for field '%s' in record '%s'", fieldName, recordName)}
		}

		fields[fieldName] = fieldType
		// Task 9.5: Store field declaration
		fieldDecls[fieldName] = field
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, fields)

	// Task 9.7: Store method AST nodes for runtime invocation
	// Build maps for instance methods and static methods (class function/procedure)
	// Task 9.7f: Separate static methods from instance methods
	methods := make(map[string]*ast.FunctionDecl)
	staticMethods := make(map[string]*ast.FunctionDecl)
	for _, method := range decl.Methods {
		if method.IsClassMethod {
			staticMethods[method.Name.Value] = method
		} else {
			methods[method.Name.Value] = method
		}
	}

	// TODO: Handle properties if needed

	// Store record type metadata in environment with special key
	// This allows variable declarations to resolve the type
	// Task 9.225: Normalize to lowercase for case-insensitive lookups
	recordTypeKey := "__record_type_" + strings.ToLower(recordName)
	recordTypeValue := &RecordTypeValue{
		RecordType:           recordType,
		FieldDecls:           fieldDecls, // Task 9.5: Include field declarations
		Methods:              methods,
		StaticMethods:        staticMethods,
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
	}
	// Initialize ClassMethods with StaticMethods for compatibility
	for k, v := range staticMethods {
		recordTypeValue.ClassMethods[k] = v
	}
	i.env.Define(recordTypeKey, recordTypeValue)

	// Also store in records map for easier access during method implementation
	i.records[recordName] = recordTypeValue

	// Initialize overload lists from method declarations
	for methodName, methodDecl := range methods {
		lowerName := strings.ToLower(methodName)
		recordTypeValue.MethodOverloads[lowerName] = []*ast.FunctionDecl{methodDecl}
	}
	for methodName, methodDecl := range staticMethods {
		lowerName := strings.ToLower(methodName)
		recordTypeValue.ClassMethodOverloads[lowerName] = []*ast.FunctionDecl{methodDecl}
	}

	return &NilValue{}
}

// RecordTypeValue is an internal value type used to store record type metadata
// in the interpreter's environment.
// Task 9.7: Extended to include method AST nodes for runtime execution.
type RecordTypeValue struct {
	RecordType           *types.RecordType
	FieldDecls           map[string]*ast.FieldDecl      // Field declarations (for initializers) - Task 9.5
	Methods              map[string]*ast.FunctionDecl   // Instance methods: Method name -> AST declaration
	StaticMethods        map[string]*ast.FunctionDecl   // Static methods: Method name -> AST declaration (class function/procedure)
	ClassMethods         map[string]*ast.FunctionDecl   // Alias for StaticMethods (for compatibility)
	MethodOverloads      map[string][]*ast.FunctionDecl // Instance method overloads
	ClassMethodOverloads map[string][]*ast.FunctionDecl // Static method overloads
}

// Type returns "RECORD_TYPE".
func (r *RecordTypeValue) Type() string {
	return "RECORD_TYPE"
}

// String returns the record type name.
func (r *RecordTypeValue) String() string {
	return r.RecordType.Name
}

// createRecordValue creates a new RecordValue with proper method lookup for nested records.
// Task 9.7e1: Helper to create records with method resolution for nested record fields.
// Task 9.5: Initialize fields with field initializers.
func (i *Interpreter) createRecordValue(recordType *types.RecordType, methods map[string]*ast.FunctionDecl) Value {
	// Create a method lookup callback that can resolve methods for nested records
	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		// Look up the record type in the environment
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		key := "__record_type_" + strings.ToLower(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	// Task 9.5: Look up the record type value to get field declarations before creating the value
	recordTypeKey := "__record_type_" + strings.ToLower(recordType.Name)
	var rtv *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		rtv, _ = typeVal.(*RecordTypeValue)
	}

	// Create the record value
	rv := newRecordValueInternal(recordType, methods, methodsLookup)

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
				// Use getZeroValueForType to properly initialize nested records
				// This ensures nested record fields are initialized as RecordValue instances
				// rather than NilValue, fixing the bug where Outer.Inner.X would fail
				fieldValue = getZeroValueForType(fieldType, methodsLookup)
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
		// Task 9.225: Normalize to lowercase for case-insensitive lookups
		recordTypeKey := "__record_type_" + strings.ToLower(typeName)
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
	recordTypeKey := "__record_type_" + strings.ToLower(literal.TypeName.Value)
	var recordTypeValue *RecordTypeValue
	if typeVal, ok := i.env.Get(recordTypeKey); ok {
		recordTypeValue, _ = typeVal.(*RecordTypeValue)
	}

	// Create the record value
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
	}

	// Evaluate and assign field values from literal
	for _, field := range literal.Fields {
		// Skip positional fields (not yet implemented)
		if field.Name == nil {
			return &ErrorValue{Message: "positional record field initialization not yet supported"}
		}

		fieldName := field.Name.Value

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

	// Task 9.5: Initialize remaining fields with field initializers or default values
	// Create a method lookup callback for nested records
	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + strings.ToLower(rt.Name)
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

	// Normalize type name to lowercase for case-insensitive comparison
	// DWScript (like Pascal) is case-insensitive for all identifiers including type names
	lowerTypeName := strings.ToLower(typeName)

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
		// Try enum type
		if enumTypeVal, ok := i.env.Get("__enum_type_" + lowerTypeName); ok {
			if etv, ok := enumTypeVal.(*EnumTypeValue); ok {
				return etv.EnumType, nil
			}
		}
		// Try record type
		if recordTypeVal, ok := i.env.Get("__record_type_" + lowerTypeName); ok {
			if rtv, ok := recordTypeVal.(*RecordTypeValue); ok {
				return rtv.RecordType, nil
			}
		}
		// Try array type
		if arrayTypeVal, ok := i.env.Get("__array_type_" + lowerTypeName); ok {
			if atv, ok := arrayTypeVal.(*ArrayTypeValue); ok {
				return atv.ArrayType, nil
			}
		}
		// Try type alias
		if typeAliasVal, ok := i.env.Get("__type_alias_" + lowerTypeName); ok {
			if tav, ok := typeAliasVal.(*TypeAliasValue); ok {
				// Return the underlying type (type aliases are transparent at runtime)
				return tav.AliasedType, nil
			}
		}
		// Try subrange type
		if subrangeTypeVal, ok := i.env.Get("__subrange_type_" + lowerTypeName); ok {
			if stv, ok := subrangeTypeVal.(*SubrangeTypeValue); ok {
				return stv.SubrangeType, nil
			}
		}
		// Unknown type
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// ============================================================================
// Record Comparison
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
