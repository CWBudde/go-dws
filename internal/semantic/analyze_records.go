package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
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
	if a.hasType(recordName) {
		a.addError("%s", errors.FormatNameAlreadyExists(recordName, decl.Token.Pos.Line, decl.Token.Pos.Column))
		return
	}

	// Create the record type
	recordType := types.NewRecordType(recordName, make(map[string]types.Type))

	// Register the record before resolving fields so recursive field types like
	// `array of TRecord` can resolve to this in-progress type.
	a.registerTypeWithPos(recordName, recordType, decl.Token.Pos)
	a.symbols.Define(recordName, recordType, decl.Token.Pos)

	// Validate field declarations
	// Track field names to detect duplicates
	fieldNames := make(map[string]bool)

	for _, field := range decl.Fields {
		fieldName := field.Name.Value
		// Normalize to lowercase for case-insensitive field access
		lowerFieldName := ident.Normalize(fieldName)

		// Check for duplicate field names (case-insensitive)
		if fieldNames[lowerFieldName] {
			a.addError("%s", errors.FormatNameAlreadyExists(fieldName, field.Token.Pos.Line, field.Token.Pos.Column))
			continue
		}
		fieldNames[lowerFieldName] = true

		var fieldType types.Type
		var err error

		// Check if type is provided or needs inference
		if field.Type != nil {
			// Explicit type
			typeName := getTypeExpressionName(field.Type)
			fieldType, err = a.resolveTypeExpression(field.Type)
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

		if a.recordFieldContainsRecordByValue(fieldType, recordType, make(map[*types.RecordType]bool)) {
			pos := field.Token.Pos
			if field.Type != nil {
				pos = field.Type.End()
			}
			a.addStructuredError(NewGenericError(pos, fmt.Sprintf(`Record type "%s" is not fully defined`, recordName)))
		}

		recordType.AddField(fieldName, fieldType, field.InitValue != nil)
	}

	// Process constants
	for _, constant := range decl.Constants {
		constName := constant.Name.Value
		lowerConstName := ident.Normalize(constName)

		// Analyze the constant value
		constValueType := a.analyzeExpression(constant.Value)
		if constValueType == nil {
			a.addError("cannot evaluate constant '%s' in record '%s' at %s",
				constName, recordName, constant.Token.Pos.String())
			continue
		}

		// If type is specified, check compatibility
		var constType types.Type
		constTypeName := getTypeExpressionName(constant.Type)
		if constant.Type != nil && constTypeName != "" {
			ct, err := a.resolveType(constTypeName)
			if err != nil {
				a.addError("unknown type '%s' for constant '%s' in record '%s' at %s",
					constTypeName, constName, recordName, constant.Token.Pos.String())
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
		lowerVarName := ident.Normalize(varName)

		var varType types.Type
		var err error

		// Resolve class variable type
		if classVar.Type != nil {
			typeName := getTypeExpressionName(classVar.Type)
			varType, err = a.resolveTypeExpression(classVar.Type)
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
		recordType.ClassVarNames[lowerVarName] = varName
	}

	// Process methods if any
	for _, method := range decl.Methods {
		methodName := method.Name.Value
		lowerMethodName := ident.Normalize(methodName)

		if method.IsStatic && !method.IsClassMethod {
			pos := method.StaticPos
			if pos.Line == 0 {
				pos = method.Token.Pos
			}
			a.addError("Syntax Error: Only non-virtual class methods can be marked as static [line: %d, column: %d]",
				pos.Line, pos.Column)
		}

		// Preserve original casing for later hinting and scope binding
		if method.IsClassMethod {
			recordType.ClassMethodNames[lowerMethodName] = methodName
		} else {
			recordType.MethodNames[lowerMethodName] = methodName
		}

		// Create function type for the method
		var paramTypes []types.Type
		for _, param := range method.Parameters {
			paramType, err := a.resolveType(getTypeExpressionName(param.Type))
			if err != nil {
				a.addError("unknown type '%s' for parameter '%s' in method '%s' at %s",
					getTypeExpressionName(param.Type), param.Name.Value, methodName, param.Token.Pos.String())
				continue
			}
			paramTypes = append(paramTypes, paramType)
		}

		var returnType types.Type
		if method.ReturnType != nil {
			rt, err := a.resolveType(getTypeExpressionName(method.ReturnType))
			if err != nil {
				a.addError("unknown return type '%s' for method '%s' at %s",
					getTypeExpressionName(method.ReturnType), methodName, method.Token.Pos.String())
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

		// Analyze method body if present (inline method)
		if method.Body != nil {
			a.analyzeRecordMethodBody(method, recordType)
		}
	}

	// Process properties if any
	for _, prop := range decl.Properties {
		propName := prop.Name.Value
		lowerPropName := ident.Normalize(propName)

		// Resolve property type
		propType, err := a.resolveTypeExpression(prop.Type)
		if err != nil {
			a.addError("unknown type '%s' for property '%s' in record '%s' at %s",
				getTypeExpressionName(prop.Type), propName, recordName, prop.Token.Pos.String())
			continue
		}

		// Create record property info
		propInfo := &types.RecordPropertyInfo{
			Name:       propName,
			Type:       propType,
			ReadField:  prop.ReadField,
			WriteField: prop.WriteField,
			IsDefault:  prop.IsDefault,
			IsIndexed:  len(prop.IndexParams) > 0,
		}

		// Determine read/write access kinds (field vs. expression).
		switch {
		case prop.ReadField != "":
			propInfo.ReadKind = types.PropAccessField
		case prop.ReadExpr != nil:
			propInfo.ReadKind = types.PropAccessExpression
			propInfo.ReadExpr = prop.ReadExpr
		default:
			propInfo.ReadKind = types.PropAccessNone
		}
		switch {
		case prop.WriteField != "":
			propInfo.WriteKind = types.PropAccessField
		case prop.WriteStmt != nil:
			propInfo.WriteKind = types.PropAccessExpression
			propInfo.WriteExpr = prop.WriteStmt
		default:
			propInfo.WriteKind = types.PropAccessNone
		}

		// Store with lowercase key for case-insensitive lookup
		recordType.Properties[lowerPropName] = propInfo
	}

	// Record type already registered above (after fields, before methods)
}

func (a *Analyzer) recordFieldContainsRecordByValue(fieldType types.Type, target *types.RecordType, seen map[*types.RecordType]bool) bool {
	if fieldType == nil || target == nil {
		return false
	}

	switch t := types.GetUnderlyingType(fieldType).(type) {
	case *types.RecordType:
		if t == target || ident.Equal(t.Name, target.Name) {
			return true
		}
		if seen[t] {
			return false
		}
		seen[t] = true
		for _, nestedFieldType := range t.Fields {
			if a.recordFieldContainsRecordByValue(nestedFieldType, target, seen) {
				return true
			}
		}
	case *types.ArrayType:
		if t.IsDynamic() {
			return false
		}
		return a.recordFieldContainsRecordByValue(t.ElementType, target, seen)
	}

	return false
}

// analyzeRecordFieldAccess analyzes access to a record field.
func (a *Analyzer) analyzeRecordFieldAccess(obj ast.Expression, field *ast.Identifier) types.Type {
	if field == nil {
		return nil
	}

	fieldName := field.Value

	// Get the type of the object
	objType := a.analyzeExpression(obj)
	if objType == nil {
		return nil
	}
	if implicitType := a.getImplicitCallType(obj); implicitType != nil {
		objType = implicitType
	} else if implicitType := implicitCallReturnTypeFromType(objType); implicitType != nil {
		objType = implicitType
	}
	objType = types.GetUnderlyingType(objType)

	// Check if the type is a record type
	recordType, ok := objType.(*types.RecordType)
	if !ok {
		a.addError("%s has no fields", objType.String())
		return nil
	}

	// Normalize field name to lowercase for case-insensitive lookup
	lowerFieldName := ident.Normalize(fieldName)

	// Check if the field exists
	fieldType, exists := recordType.Fields[lowerFieldName]
	if exists {
		if declaredName := recordType.FieldNames[lowerFieldName]; declaredName != "" &&
			declaredName != fieldName && ident.Equal(declaredName, fieldName) {
			a.addCaseMismatchHint(fieldName, declaredName, field.Token.Pos)
		}
		// TODO: Check visibility rules if needed
		// For now, we allow all field access
		return fieldType
	}

	// Check if it's a constant
	constInfo, constExists := recordType.Constants[lowerFieldName]
	if constExists {
		if constInfo.Name != "" && constInfo.Name != fieldName && ident.Equal(constInfo.Name, fieldName) {
			a.addCaseMismatchHint(fieldName, constInfo.Name, field.Token.Pos)
		}
		return constInfo.Type
	}

	// Check if it's a class variable
	classVarType, classVarExists := recordType.ClassVars[lowerFieldName]
	if classVarExists {
		if declaredName := recordType.ClassVarNames[lowerFieldName]; declaredName != "" &&
			declaredName != fieldName && ident.Equal(declaredName, fieldName) {
			a.addCaseMismatchHint(fieldName, declaredName, field.Token.Pos)
		}
		return classVarType
	}

	// Check if it's an instance method of the record
	methodType, methodExists := recordType.Methods[lowerFieldName]
	if methodExists {
		if declaredName := recordType.MethodNames[lowerFieldName]; declaredName != "" &&
			declaredName != fieldName && ident.Equal(declaredName, fieldName) {
			a.addCaseMismatchHint(fieldName, declaredName, field.Token.Pos)
		}
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

	// Check if it's a class method (can be called on instances)
	classMethodType, classMethodExists := recordType.ClassMethods[lowerFieldName]
	if classMethodExists {
		if declaredName := recordType.ClassMethodNames[lowerFieldName]; declaredName != "" &&
			declaredName != fieldName && ident.Equal(declaredName, fieldName) {
			a.addCaseMismatchHint(fieldName, declaredName, field.Token.Pos)
		}
		// Class methods can be accessed on instances
		// If parameterless, will be auto-invoked
		if len(classMethodType.Parameters) == 0 {
			if classMethodType.ReturnType != nil {
				return classMethodType.ReturnType
			}
			return types.VOID
		}
		// Method has parameters - return function type
		return classMethodType
	}

	// Check if it's a property of the record
	if recordType.Properties != nil {
		propInfo, propExists := recordType.Properties[lowerFieldName]
		if propExists {
			if propInfo.Name != "" && propInfo.Name != fieldName && ident.Equal(propInfo.Name, fieldName) {
				a.addCaseMismatchHint(fieldName, propInfo.Name, field.Token.Pos)
			}
			return propInfo.Type
		}
	}

	// Check if a helper provides this member
	helperMethod := a.hasHelperMethod(objType, fieldName)
	if helperMethod != nil {
		// Parameterless helper methods auto-invoke on member access
		if len(helperMethod.Parameters) == 0 {
			if helperMethod.ReturnType != nil {
				return helperMethod.ReturnType
			}
			return types.VOID
		}
		return helperMethod
	}

	helperProp := a.hasHelperProperty(objType, fieldName)
	if helperProp != nil {
		return helperProp.Type
	}

	if _, constType := a.hasHelperClassConst(objType, fieldName); constType != nil {
		if t, ok := constType.(types.Type); ok {
			return t
		}
	}
	if _, varType := a.hasHelperClassVar(objType, fieldName); varType != nil {
		return varType
	}

	a.addStructuredError(NewAccessibleMemberError(field.Token.Pos, fieldName, recordType.Name))
	return nil
}
