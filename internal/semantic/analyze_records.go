package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Record Type Analysis (Task 8.68-8.71)
// ============================================================================

// analyzeRecordDecl analyzes a record type declaration.
// Task 8.68: Register record types in symbol table
// Task 8.69: Validate record field declarations
func (a *Analyzer) analyzeRecordDecl(decl *ast.RecordDecl) {
	if decl == nil {
		return
	}

	recordName := decl.Name.Value

	// Task 8.68: Check if record is already declared
	if _, exists := a.records[recordName]; exists {
		a.addError("record type '%s' already declared at %s", recordName, decl.Token.Pos.String())
		return
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, make(map[string]types.Type))

	// Task 8.69: Validate field declarations
	// Track field names to detect duplicates
	fieldNames := make(map[string]bool)

	for _, field := range decl.Fields {
		fieldName := field.Name.Value

		// Check for duplicate field names
		if fieldNames[fieldName] {
			a.addError("duplicate field '%s' in record '%s' at %s", fieldName, recordName, field.Token.Pos.String())
			continue
		}
		fieldNames[fieldName] = true

		// Resolve field type
		typeName := getTypeExpressionName(field.Type)
		fieldType, err := a.resolveType(typeName)
		if err != nil {
			a.addError("unknown type '%s' for field '%s' in record '%s' at %s",
				typeName, fieldName, recordName, field.Token.Pos.String())
			continue
		}

		// Add field to record type
		recordType.Fields[fieldName] = fieldType
	}

	// Process methods if any (Task 8.61c)
	for _, method := range decl.Methods {
		methodName := method.Name.Value

		// Create function type for the method
		var paramTypes []types.Type
		for _, param := range method.Parameters {
			paramType, err := a.resolveType(param.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for parameter '%s' in method '%s' at %s",
					param.Type.Name, param.Name.Value, methodName, param.Token.Pos.String())
				continue
			}
			paramTypes = append(paramTypes, paramType)
		}

		var returnType types.Type
		if method.ReturnType != nil {
			rt, err := a.resolveType(method.ReturnType.Name)
			if err != nil {
				a.addError("unknown return type '%s' for method '%s' at %s",
					method.ReturnType.Name, methodName, method.Token.Pos.String())
			} else {
				returnType = rt
			}
		}

		var funcType *types.FunctionType
		if returnType != nil {
			funcType = types.NewFunctionType(paramTypes, returnType)
		} else {
			funcType = types.NewProcedureType(paramTypes)
		}

		recordType.Methods[methodName] = funcType
	}

	// Process properties if any (Task 8.61d)
	for _, prop := range decl.Properties {
		propName := prop.Name.Value

		// Resolve property type
		propType, err := a.resolveType(prop.Type.Name)
		if err != nil {
			a.addError("unknown type '%s' for property '%s' in record '%s' at %s",
				prop.Type.Name, propName, recordName, prop.Token.Pos.String())
			continue
		}

		// Create record property info
		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
		}

		recordType.Properties[propName] = propInfo
	}

	// Register the record type
	a.records[recordName] = recordType

	// Also register in symbol table as a type
	a.symbols.Define(recordName, recordType)
}

// analyzeRecordLiteral analyzes a record literal expression.
// Task 8.70: Type-check record literals (field names/types match, positional vs named)
func (a *Analyzer) analyzeRecordLiteral(lit *ast.RecordLiteral, expectedType types.Type) types.Type {
	if lit == nil {
		return nil
	}

	// For record literals, we need to know the expected type
	// This comes from the variable declaration or assignment context
	recordType, ok := expectedType.(*types.RecordType)
	if !ok {
		a.addError("record literal requires a record type, got %s", expectedType.String())
		return nil
	}

	// Track which fields have been initialized
	initializedFields := make(map[string]bool)

	// Validate each field in the literal
	for _, field := range lit.Fields {
		fieldName := field.Name

		// Check for duplicate field initialization
		if initializedFields[fieldName] {
			a.addError("duplicate field '%s' in record literal", fieldName)
			continue
		}
		initializedFields[fieldName] = true

		// Check if field exists in record type
		expectedFieldType, exists := recordType.Fields[fieldName]
		if !exists {
			a.addError("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
			continue
		}

		// Type-check the field value
		actualType := a.analyzeExpression(field.Value)
		if actualType == nil {
			continue
		}

		// Check type compatibility
		if !a.canAssign(actualType, expectedFieldType) {
			a.addError("cannot assign %s to %s in field '%s'",
				actualType.String(), expectedFieldType.String(), fieldName)
		}
	}

	// Check for missing required fields
	for fieldName := range recordType.Fields {
		if !initializedFields[fieldName] {
			a.addError("missing required field '%s' in record literal", fieldName)
		}
	}

	return recordType
}

// analyzeRecordFieldAccess analyzes access to a record field.
// Task 8.71: Type-check record field access (field exists, visibility rules)
func (a *Analyzer) analyzeRecordFieldAccess(obj ast.Expression, fieldName string) types.Type {
	// Get the type of the object
	objType := a.analyzeExpression(obj)
	if objType == nil {
		return nil
	}

	// Check if the type is a record type
	recordType, ok := objType.(*types.RecordType)
	if !ok {
		a.addError("%s has no fields", objType.String())
		return nil
	}

	// Check if the field exists
	fieldType, exists := recordType.Fields[fieldName]
	if !exists {
		// Task 9.83: Check if a helper provides this member
		_, helperMethod := a.hasHelperMethod(objType, fieldName)
		if helperMethod != nil {
			return helperMethod
		}

		_, helperProp := a.hasHelperProperty(objType, fieldName)
		if helperProp != nil {
			return helperProp.Type
		}

		a.addError("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
		return nil
	}

	// TODO: Check visibility rules if needed
	// For now, we allow all field access

	return fieldType
}
