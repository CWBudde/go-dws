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

// buildRecordMetadata builds RecordMetadata from AST declarations.
// Task 3.5.42: Helper to create AST-free metadata for records.
// Deprecated: Task 3.5.10 - Moved to evaluator.buildRecordMetadata().
// Kept for backward compatibility during Phase 3.5 migration.
func (i *Interpreter) buildRecordMetadata(
	recordName string,
	recordType *types.RecordType,
	methods map[string]*ast.FunctionDecl,
	staticMethods map[string]*ast.FunctionDecl,
	constants map[string]Value,
	classVars map[string]Value,
) *runtime.RecordMetadata {
	metadata := runtime.NewRecordMetadata(recordName, recordType)

	// Convert instance methods to MethodMetadata
	for methodName, methodDecl := range methods {
		methodMeta := i.buildMethodMetadata(methodDecl)
		metadata.Methods[methodName] = methodMeta
		metadata.MethodOverloads[methodName] = []*runtime.MethodMetadata{methodMeta}
	}

	// Convert static methods to MethodMetadata
	for methodName, methodDecl := range staticMethods {
		methodMeta := i.buildMethodMetadata(methodDecl)
		methodMeta.IsClassMethod = true
		metadata.StaticMethods[methodName] = methodMeta
		metadata.StaticMethodOverloads[methodName] = []*runtime.MethodMetadata{methodMeta}
	}

	// Copy constants and class vars
	for k, v := range constants {
		metadata.Constants[k] = v
	}
	for k, v := range classVars {
		metadata.ClassVars[k] = v
	}

	return metadata
}

// buildMethodMetadata converts an AST FunctionDecl to MethodMetadata.
// Task 3.5.42: Helper to extract metadata from AST method declarations.
// Deprecated: Task 3.5.10 - Moved to evaluator.buildMethodMetadata().
// Kept for backward compatibility during Phase 3.5 migration.
func (i *Interpreter) buildMethodMetadata(decl *ast.FunctionDecl) *runtime.MethodMetadata {
	// Build parameter metadata
	params := make([]runtime.ParameterMetadata, len(decl.Parameters))
	for idx, param := range decl.Parameters {
		typeName := ""
		if param.Type != nil {
			typeName = param.Type.String()
		}
		params[idx] = runtime.ParameterMetadata{
			Name:         param.Name.Value,
			TypeName:     typeName,
			Type:         nil, // Will be resolved later if needed
			ByRef:        param.ByRef,
			DefaultValue: param.DefaultValue,
		}
	}

	// Determine return type
	returnTypeName := ""
	if decl.ReturnType != nil {
		returnTypeName = decl.ReturnType.String()
	}

	return &runtime.MethodMetadata{
		Name:           decl.Name.Value,
		Parameters:     params,
		ReturnTypeName: returnTypeName,
		ReturnType:     nil, // Will be resolved later if needed
		Body:           decl.Body,
		IsClassMethod:  decl.IsClassMethod,
		IsConstructor:  decl.IsConstructor,
		IsDestructor:   decl.IsDestructor,
	}
}

// ============================================================================
// Record Declaration Evaluation
// ============================================================================

// evalRecordDeclaration evaluates a record type declaration.
// It registers the record type in the interpreter's symbol table.
// evalRecordDeclaration evaluates a record declaration.
// Deprecated: Task 3.5.10 - Full implementation moved to evaluator.VisitRecordDecl().
// Kept for backward compatibility during Phase 3.5 migration.
// This function is no longer called by the evaluator.
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

		// Task 9.12.1: Handle type inference for fields
		var fieldType types.Type
		if field.Type != nil {
			// Explicit type
			// Task 9.170.1: Updated to support inline array types
			fieldType = i.resolveTypeFromExpression(field.Type)
			if fieldType == nil {
				return &ErrorValue{Message: fmt.Sprintf("unknown or invalid type for field '%s' in record '%s'", fieldName, recordName)}
			}
		} else if field.InitValue != nil {
			// Type inference from initializer
			initValue := i.Eval(field.InitValue)
			if isError(initValue) {
				return initValue
			}
			fieldType = i.getValueType(initValue)
			if fieldType == nil {
				return &ErrorValue{Message: fmt.Sprintf("cannot infer type for field '%s' in record '%s'", fieldName, recordName)}
			}
		} else {
			return &ErrorValue{Message: fmt.Sprintf("field '%s' in record '%s' must have either a type or initializer", fieldName, recordName)}
		}

		// Use lowercase key for case-insensitive access
		fieldNameLower := ident.Normalize(fieldName)
		fields[fieldNameLower] = fieldType
		// Task 9.5: Store field declaration (use lowercase key)
		fieldDecls[fieldNameLower] = field
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, fields)

	// Task 9.7: Store method AST nodes for runtime invocation
	// Build maps for instance methods and static methods (class function/procedure)
	// Task 9.7f: Separate static methods from instance methods
	// Note: Use lowercase keys for case-insensitive lookup
	methods := make(map[string]*ast.FunctionDecl)
	staticMethods := make(map[string]*ast.FunctionDecl)
	for _, method := range decl.Methods {
		methodKey := ident.Normalize(method.Name.Value)
		if method.IsClassMethod {
			staticMethods[methodKey] = method
		} else {
			methods[methodKey] = method
		}
	}

	// Task 9.12.2: Evaluate record constants
	constants := make(map[string]Value)
	for _, constant := range decl.Constants {
		constName := constant.Name.Value
		constValue := i.Eval(constant.Value)
		if isError(constValue) {
			return constValue
		}
		// Normalize to lowercase for case-insensitive access
		constants[ident.Normalize(constName)] = constValue
	}

	// Task 9.12.2: Initialize class variables
	classVars := make(map[string]Value)
	for _, classVar := range decl.ClassVars {
		varName := classVar.Name.Value
		var varValue Value

		if classVar.InitValue != nil {
			// Evaluate the initializer
			varValue = i.Eval(classVar.InitValue)
			if isError(varValue) {
				return varValue
			}
		} else {
			// Use type to determine zero value
			var varType types.Type
			if classVar.Type != nil {
				varType = i.resolveTypeFromExpression(classVar.Type)
				if varType == nil {
					return &ErrorValue{Message: fmt.Sprintf("unknown type for class variable '%s' in record '%s'", varName, recordName)}
				}
			}
			varValue = getZeroValueForType(varType, nil)
		}

		// Normalize to lowercase for case-insensitive access
		classVars[ident.Normalize(varName)] = varValue
	}

	// Process properties
	for _, prop := range decl.Properties {
		propName := prop.Name.Value
		propNameLower := ident.Normalize(propName)

		// Resolve property type
		propType := i.resolveTypeFromExpression(prop.Type)
		if propType == nil {
			return &ErrorValue{Message: fmt.Sprintf("unknown type for property '%s' in record '%s'", propName, recordName)}
		}

		// Create property info
		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
		}

		// Store in recordType.Properties (case-insensitive)
		recordType.Properties[propNameLower] = propInfo
	}

	// Task 3.5.42: Build RecordMetadata from AST declarations
	metadata := i.buildRecordMetadata(recordName, recordType, methods, staticMethods, constants, classVars)

	// Store record type metadata in environment with special key
	// This allows variable declarations to resolve the type
	recordTypeKey := "__record_type_" + ident.Normalize(recordName)
	recordTypeValue := &RecordTypeValue{
		RecordType:           recordType,
		FieldDecls:           fieldDecls, // Task 9.5: Include field declarations
		Metadata:             metadata,   // Task 3.5.42: AST-free metadata
		Methods:              methods,
		StaticMethods:        staticMethods,
		ClassMethods:         make(map[string]*ast.FunctionDecl),
		ClassMethodOverloads: make(map[string][]*ast.FunctionDecl),
		MethodOverloads:      make(map[string][]*ast.FunctionDecl),
		Constants:            constants, // Task 9.12.2: Record constants
		ClassVars:            classVars, // Task 9.12.2: Class variables
	}
	// Initialize ClassMethods with StaticMethods for compatibility
	for k, v := range staticMethods {
		recordTypeValue.ClassMethods[k] = v
	}
	i.env.Define(recordTypeKey, recordTypeValue)

	// Also store in records map for easier access during method implementation
	// Register record in TypeSystem
	i.typeSystem.RegisterRecord(recordName, recordTypeValue)
	// Also maintain legacy map for backward compatibility during migration
	i.records[ident.Normalize(recordName)] = recordTypeValue

	// Initialize overload lists from method declarations
	// Note: methodName is already lowercase from the maps above
	for methodName, methodDecl := range methods {
		recordTypeValue.MethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}
	for methodName, methodDecl := range staticMethods {
		recordTypeValue.ClassMethodOverloads[methodName] = []*ast.FunctionDecl{methodDecl}
	}

	return &NilValue{}
}

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
