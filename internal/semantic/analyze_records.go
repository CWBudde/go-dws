package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Record Type Analysis
// ============================================================================

// analyzeRecordDecl analyzes a record type declaration.
func (a *Analyzer) analyzeRecordDecl(decl *ast.RecordDecl) {
	if decl == nil {
		return
	}

	recordName := decl.Name.Value

	// Check if record is already declared
	// Use lowercase for case-insensitive duplicate check
	if _, exists := a.records[strings.ToLower(recordName)]; exists {
		a.addError("record type '%s' already declared at %s", recordName, decl.Token.Pos.String())
		return
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, make(map[string]types.Type))

	// Validate field declarations
	// Track field names to detect duplicates
	fieldNames := make(map[string]bool)

	for _, field := range decl.Fields {
		fieldName := field.Name.Value
		// Normalize to lowercase for case-insensitive field access
		lowerFieldName := strings.ToLower(fieldName)

		// Check for duplicate field names (case-insensitive)
		if fieldNames[lowerFieldName] {
			a.addError("duplicate field '%s' in record '%s' at %s", fieldName, recordName, field.Token.Pos.String())
			continue
		}
		fieldNames[lowerFieldName] = true

		var fieldType types.Type
		var err error

		// Check if type is provided or needs inference
		if field.Type != nil {
			// Explicit type
			typeName := getTypeExpressionName(field.Type)
			fieldType, err = a.resolveType(typeName)
			if err != nil {
				a.addError("unknown type '%s' for field '%s' in record '%s' at %s",
					typeName, fieldName, recordName, field.Token.Pos.String())
				continue
			}

			// Validate field initializer if present
			a.validateFieldInitializer(field, fieldName, fieldType)
		} else if field.InitValue != nil {
			// Type inference from initializer
			initType := a.analyzeExpression(field.InitValue)
			if initType == nil {
				a.addError("cannot infer type for field '%s' in record '%s' at %s",
					fieldName, recordName, field.Token.Pos.String())
				continue
			}
			fieldType = initType
		} else {
			a.addError("field '%s' in record '%s' must have either a type or initializer at %s",
				fieldName, recordName, field.Token.Pos.String())
			continue
		}

		// Add field to record type (using lowercase key for case-insensitive lookup)
		recordType.Fields[lowerFieldName] = fieldType
	}

	// Register the record type early so it's visible in method signatures
	// Use lowercase key for case-insensitive lookup
	a.records[strings.ToLower(recordName)] = recordType
	// Also register in symbol table as a type
	a.symbols.Define(recordName, recordType)

	// Process constants
	for _, constant := range decl.Constants {
		constName := constant.Name.Value
		lowerConstName := strings.ToLower(constName)

		// Analyze the constant value
		constValueType := a.analyzeExpression(constant.Value)
		if constValueType == nil {
			a.addError("cannot evaluate constant '%s' in record '%s' at %s",
				constName, recordName, constant.Token.Pos.String())
			continue
		}

		// If type is specified, check compatibility
		var constType types.Type
		if constant.Type != nil {
			ct, err := a.resolveType(constant.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for constant '%s' in record '%s' at %s",
					constant.Type.Name, constName, recordName, constant.Token.Pos.String())
				continue
			}
			constType = ct

			// Check if value type is compatible with declared type
			if !types.IsAssignableFrom(constType, constValueType) {
				a.addError("constant '%s' type mismatch: expected %s, got %s at %s",
					constName, constType.String(), constValueType.String(), constant.Token.Pos.String())
				continue
			}
		} else {
			constType = constValueType
		}

		// Create constant info (value evaluation will be done at runtime)
		constInfo := &types.ConstantInfo{
			Name:         constName,
			Type:         constType,
			Value:        nil, // Will be set by interpreter
			IsClassConst: constant.IsClassConst,
		}

		recordType.Constants[lowerConstName] = constInfo
	}

	// Process class variables
	for _, classVar := range decl.ClassVars {
		varName := classVar.Name.Value
		lowerVarName := strings.ToLower(varName)

		var varType types.Type
		var err error

		// Resolve class variable type
		if classVar.Type != nil {
			typeName := getTypeExpressionName(classVar.Type)
			varType, err = a.resolveType(typeName)
			if err != nil {
				a.addError("unknown type '%s' for class variable '%s' in record '%s' at %s",
					typeName, varName, recordName, classVar.Token.Pos.String())
				continue
			}
		} else if classVar.InitValue != nil {
			// Type inference from initializer
			initType := a.analyzeExpression(classVar.InitValue)
			if initType == nil {
				a.addError("cannot infer type for class variable '%s' in record '%s' at %s",
					varName, recordName, classVar.Token.Pos.String())
				continue
			}
			varType = initType
		} else {
			a.addError("class variable '%s' in record '%s' must have either a type or initializer at %s",
				varName, recordName, classVar.Token.Pos.String())
			continue
		}

		// Validate initializer if present
		if classVar.InitValue != nil && classVar.Type != nil {
			initType := a.analyzeExpression(classVar.InitValue)
			if initType != nil && !types.IsAssignableFrom(varType, initType) {
				a.addError("class variable '%s' initializer type mismatch: expected %s, got %s at %s",
					varName, varType.String(), initType.String(), classVar.Token.Pos.String())
			}
		}

		recordType.ClassVars[lowerVarName] = varType
	}

	// Process methods if any
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

		// Create MethodInfo for overload tracking
		methodInfo := &types.MethodInfo{
			Signature:            funcType,
			IsClassMethod:        method.IsClassMethod,
			HasOverloadDirective: method.IsOverload,
			Visibility:           int(method.Visibility),
		}

		lowerMethodName := strings.ToLower(methodName)

		// Store in appropriate maps based on whether it's a class method (static)
		if method.IsClassMethod {
			// Store primary signature and add to overloads
			if recordType.ClassMethods[lowerMethodName] == nil {
				recordType.ClassMethods[lowerMethodName] = funcType
			}
			recordType.ClassMethodOverloads[lowerMethodName] = append(
				recordType.ClassMethodOverloads[lowerMethodName], methodInfo)
		} else {
			// Store primary signature and add to overloads
			if recordType.Methods[lowerMethodName] == nil {
				recordType.Methods[lowerMethodName] = funcType
			}
			recordType.MethodOverloads[lowerMethodName] = append(
				recordType.MethodOverloads[lowerMethodName], methodInfo)
		}
	}

	// Process properties if any
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

	// Record type already registered above (after fields, before methods)
}

// analyzeRecordFieldAccess analyzes access to a record field.
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
	if exists {
		// TODO: Check visibility rules if needed
		// For now, we allow all field access
		return fieldType
	}

	// Check if it's a constant
	constInfo, constExists := recordType.Constants[fieldName]
	if constExists {
		return constInfo.Type
	}

	// Check if it's a class variable
	classVarType, classVarExists := recordType.ClassVars[fieldName]
	if classVarExists {
		return classVarType
	}

	// Check if it's an instance method of the record
	methodType, methodExists := recordType.Methods[fieldName]
	if methodExists {
		// If method is parameterless, it will be auto-invoked by the interpreter
		// Return the method's return type, not the method type itself
		if len(methodType.Parameters) == 0 {
			if methodType.ReturnType != nil {
				return methodType.ReturnType
			}
			return types.VOID
		}
		// Method has parameters - return function type for deferred invocation
		return methodType
	}

	// Note: Class methods are not accessible via instance.method syntax
	// They must be called via RecordType.method (handled in analyzeMemberAccessExpression)

	// Check if it's a property of the record
	if recordType.Properties != nil {
		propInfo, propExists := recordType.Properties[fieldName]
		if propExists {
			return propInfo.Type
		}
	}

	// Check if a helper provides this member
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
