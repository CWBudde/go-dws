package semantic

import (
	"fmt"
	"strconv"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Property Semantic Analysis
// ============================================================================

// pluralizeParam returns "parameter" or "parameters" based on count.
func pluralizeParam(count int) string {
	if count == 1 {
		return "parameter"
	}
	return "parameters"
}

func formatInt(v int) string {
	return strconv.Itoa(v)
}

// analyzePropertyDecl validates a property declaration and registers it in the class metadata.
func (a *Analyzer) analyzePropertyDecl(prop *ast.PropertyDecl, classType *types.ClassType) {
	propName := prop.Name.Value

	//  Check for duplicate property names within class
	for existingName := range classType.Properties {
		if pkgident.Equal(existingName, propName) {
			a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
				"duplicate property '"+propName+"' in class '"+classType.Name+"'"))
			return
		}
	}

	// Register properties with their declared casing
	if classType.Properties == nil {
		classType.Properties = make(map[string]*types.PropertyInfo)
	}
	if _, exists := classType.Properties[propName]; exists {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"duplicate property '"+propName+"' in class '"+classType.Name+"'"))
		return
	}

	// Resolve property type
	if prop.Type == nil {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"property '"+propName+"' missing type annotation in class '"+classType.Name+"'"))
		return
	}

	propType, err := a.resolveType(getTypeExpressionName(prop.Type))
	if err != nil {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"unknown type '"+getTypeExpressionName(prop.Type)+"' for property '"+propName+"' in class '"+classType.Name+"'"))
		return
	}
	a.warnDeprecatedResolvedType(prop.Type.Pos(), propType)

	// Validate indexed property parameters have valid types
	isIndexed := prop.IndexParams != nil && len(prop.IndexParams) > 0
	var indexParamTypes []types.Type
	if isIndexed {
		for _, param := range prop.IndexParams {
			if param.Type == nil {
				a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
					"index parameter '"+param.Name.Value+"' missing type annotation in property '"+propName+"'"))
				return
			}
			paramType, err := a.resolveType(getTypeExpressionName(param.Type))
			if err != nil {
				a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
					"unknown type '"+getTypeExpressionName(param.Type)+"' for index parameter '"+param.Name.Value+"' in property '"+propName+"'"))
				return
			}
			a.warnDeprecatedResolvedType(param.Type.Pos(), paramType)
			indexParamTypes = append(indexParamTypes, paramType)
		}
	}

	// Index directive (implicit index argument) - cannot be combined with explicit index params
	var implicitIndexTypes []types.Type
	var indexLiteralValue int64
	if prop.IndexValue != nil {
		if isIndexed {
			a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
				"property '"+propName+"' cannot combine index parameters with an index directive"))
			return
		}

		// Analyze the index expression type in a property context
		savedClass := a.currentClass
		savedInClassMethod := a.inClassMethod
		a.currentClass = classType
		a.inClassMethod = false
		idxType := a.analyzeExpression(prop.IndexValue)
		a.currentClass = savedClass
		a.inClassMethod = savedInClassMethod

		if idxType == nil {
			// Errors already recorded
			return
		}

		// Currently support integer-typed index directives
		if !idxType.Equals(types.INTEGER) {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
				"property '"+propName+"' index directive must be an integer literal, got "+idxType.String()))
			return
		}

		val, ok := ast.ExtractIntegerLiteral(prop.IndexValue)
		if !ok {
			a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
				"property '"+propName+"' index directive must be an integer literal"))
			return
		}

		indexLiteralValue = val
		implicitIndexTypes = append(implicitIndexTypes, idxType)
	}

	// Combine implicit index directive types with explicit index parameters for signature validation
	totalIndexParamTypes := append(implicitIndexTypes, indexParamTypes...)

	// Create PropertyInfo to store in class metadata
	propInfo := &types.PropertyInfo{
		Name:            propName,
		Type:            propType,
		IsIndexed:       isIndexed,
		IsDefault:       prop.IsDefault,
		IsClassProperty: prop.IsClassProperty,
	}
	if prop.IndexValue != nil {
		propInfo.HasIndexValue = true
		propInfo.IndexValue = indexLiteralValue
		propInfo.IndexValueType = types.INTEGER
	}

	// Register property stub before validating specs (allows circular reference detection)
	classType.Properties[propName] = propInfo

	// Validate read specifier
	if prop.ReadSpec != nil {
		a.validateReadSpec(prop, classType, propInfo, totalIndexParamTypes)
	}

	// Validate write specifier
	if prop.WriteSpec != nil {
		a.validateWriteSpec(prop, classType, propInfo, totalIndexParamTypes)
	}

	// Validate default property restrictions
	if prop.IsDefault {
		// Default properties must be indexed
		if !isIndexed {
			a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
				"default property '"+propName+"' must be an indexed property"))
			return
		}

		// Only one default property per class
		for existingPropName, existingProp := range classType.Properties {
			if existingProp.IsDefault && existingPropName != propName {
				a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
					"class '"+classType.Name+"' already has default property '"+existingPropName+"'; cannot declare another default property '"+propName+"'"))
				return
			}
		}
	}

	// Property already registered at the beginning (line 68) for circular reference detection
}

// validateReadSpec validates the read specifier of a property.
// The read specifier can be:
//   - Field: A field name (identifier) - the field must exist and have matching type
//   - Constant: A constant name (identifier) - the constant must exist and have matching type
//   - Method: A method name (identifier) - the method must exist with correct signature
//   - Expression: An inline expression - the expression type must match property type
func (a *Analyzer) validateReadSpec(prop *ast.PropertyDecl, classType *types.ClassType, propInfo *types.PropertyInfo, indexParamTypes []types.Type) {
	propName := prop.Name.Value
	propType := propInfo.Type

	// Check if read spec is an identifier (field, constant, or method name)
	if ident, ok := prop.ReadSpec.(*ast.Identifier); ok {
		readSpecName := ident.Value

		// Check class-level members first: class vars, then constants, then instance fields

		// 1. Check if it's a class variable (only for class properties)
		if fieldType, found := classType.ClassVars[pkgident.Normalize(readSpecName)]; found {
			fieldOwner := a.getClassVarOwner(classType, readSpecName)
			if fieldOwner != nil {
				visibility, hasVisibility := fieldOwner.ClassVarVisibility[pkgident.Normalize(readSpecName)]
				if hasVisibility && !a.checkVisibility(fieldOwner, visibility, readSpecName, "class variable") {
					a.addStructuredError(NewPropertyDeclarationError(ident.Token.Pos,
						fmt.Sprintf(`Field/method "%s" not found`, readSpecName)))
					return
				}
			}
			// Only class properties can read from class variables
			if propInfo.IsClassProperty {
				if !propType.Equals(fieldType) {
					a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
						"property '"+propName+"' read class variable '"+readSpecName+"' has type "+fieldType.String()+", expected "+propType.String()))
					return
				}
				propInfo.ReadKind = types.PropAccessField
				propInfo.ReadSpec = readSpecName
				return
			}
			// Instance property cannot read from class variable - skip this match
		}

		// 2. Check if it's a constant
		if constantType, constantFound := a.getConstantType(classType, readSpecName); constantFound {
			if !propType.Equals(constantType) {
				a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
					"property '"+propName+"' read constant '"+readSpecName+"' has type "+constantType.String()+", expected "+propType.String()))
				return
			}
			propInfo.ReadKind = types.PropAccessField // Constants are treated like fields
			propInfo.ReadSpec = readSpecName
			return
		}

		// 3. Check if it's an instance field (only for instance properties)
		if !propInfo.IsClassProperty {
			// Instance property can use instance field
			if fieldType, found := classType.GetField(pkgident.Normalize(readSpecName)); found {
				fieldOwner := a.getFieldOwner(classType, readSpecName)
				if fieldOwner != nil {
					visibility, hasVisibility := fieldOwner.FieldVisibility[pkgident.Normalize(readSpecName)]
					if hasVisibility && !a.checkVisibility(fieldOwner, visibility, readSpecName, "field") {
						a.addStructuredError(NewPropertyDeclarationError(ident.Token.Pos,
							fmt.Sprintf(`Field/method "%s" not found`, readSpecName)))
						return
					}
				}
				if !propType.Equals(fieldType) {
					a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
						"property '"+propName+"' read field '"+readSpecName+"' has type "+fieldType.String()+", expected "+propType.String()))
					return
				}
				propInfo.ReadKind = types.PropAccessField
				propInfo.ReadSpec = readSpecName
				return
			}
		}

		if propInfo.IsClassProperty {
			if fieldType, found := classType.GetField(pkgident.Normalize(readSpecName)); found {
				fieldOwner := a.getFieldOwner(classType, readSpecName)
				if fieldOwner != nil {
					visibility, hasVisibility := fieldOwner.FieldVisibility[pkgident.Normalize(readSpecName)]
					if hasVisibility && !a.checkVisibility(fieldOwner, visibility, readSpecName, "field") {
						a.addStructuredError(NewPropertyDeclarationError(ident.Token.Pos,
							fmt.Sprintf(`Field/method "%s" not found`, readSpecName)))
						return
					}
				}
				_ = fieldType
				a.addStructuredError(NewClassMemberExpectedError(ident.Token.Pos))
				return
			}
		}

		// If method, verify method exists with correct signature
		if methodType, found := classType.GetMethod(pkgident.Normalize(readSpecName)); found {
			// For class properties, verify the method is a class method
			if propInfo.IsClassProperty {
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(readSpecName)]
				if !isClassMethod {
					a.addStructuredError(NewPropertyReadShouldBeStaticMethodError(ident.Token.Pos))
					return
				}
			} else {
				// For instance properties, verify the method is NOT a class method
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(readSpecName)]
				if isClassMethod {
					a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
						"instance property '"+propName+"' read method '"+readSpecName+"' cannot be a class method"))
					return
				}
			}

			// Getter signature: for indexed properties, method must accept index parameters
			// and return property type. For non-indexed, method must take no parameters
			// and return property type.

			expectedParamCount := len(indexParamTypes)
			if len(methodType.Parameters) != expectedParamCount {
				a.addStructuredError(NewPropertyDeclarationArgumentCountError(prop.Token.Pos,
					"property '"+propName+"' getter method '"+readSpecName+"' has "+
						formatInt(len(methodType.Parameters))+" "+pluralizeParam(len(methodType.Parameters))+
						", expected "+formatInt(expectedParamCount)+" "+pluralizeParam(expectedParamCount)))
				return
			}

			// Verify getter signature includes index parameters
			for i, paramType := range indexParamTypes {
				if !methodType.Parameters[i].Equals(paramType) {
					a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
						"property '"+propName+"' getter method '"+readSpecName+"' parameter "+
							formatInt(i+1)+" has type "+methodType.Parameters[i].String()+", expected "+paramType.String()))
					return
				}
			}

			// Verify return type matches property type
			if !methodType.ReturnType.Equals(propType) {
				a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
					"property '"+propName+"' getter method '"+readSpecName+"' returns "+
						methodType.ReturnType.String()+", expected "+propType.String()))
				return
			}

			propInfo.ReadKind = types.PropAccessMethod
			propInfo.ReadSpec = readSpecName
			return
		}

		// Neither field nor method found
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"property '"+propName+"' read specifier '"+readSpecName+"' not found in class '"+classType.Name+"'"))
		return
	}

	// If expression, validate expression type matches property type
	// Set up class context for expression analysis to enable implicit self access
	savedClass := a.currentClass
	savedInClassMethod := a.inClassMethod
	savedInPropertyExpr := a.inPropertyExpr
	savedCurrentProperty := a.currentProperty

	a.currentClass = classType
	a.inClassMethod = false      // Property expressions are in instance context
	a.inPropertyExpr = true      // Flag to enable special property expression handling
	a.currentProperty = propName // Track current property for circular reference detection

	defer func() {
		a.currentClass = savedClass
		a.inClassMethod = savedInClassMethod
		a.inPropertyExpr = savedInPropertyExpr
		a.currentProperty = savedCurrentProperty
	}()

	// Analyze the expression with implicit self context
	exprType := a.analyzeExpression(prop.ReadSpec)
	if exprType == nil {
		// Error already reported by analyzeExpression
		return
	}

	// Validate expression type matches property type
	if !exprType.Equals(propType) {
		a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
			"property '"+propName+"' read expression has type "+exprType.String()+", expected "+propType.String()))
		return
	}

	// Store the expression for runtime evaluation
	propInfo.ReadKind = types.PropAccessExpression
	propInfo.ReadSpec = prop.ReadSpec.String()
	propInfo.ReadExpr = prop.ReadSpec // Store AST node for interpreter
}

// validateWriteSpec validates the write specifier of a property.
// The write specifier can be:
//   - Field: A field name (identifier) - the field must exist and have matching type
//   - Method: A method name (identifier) - the method must exist with correct signature
func (a *Analyzer) validateWriteSpec(prop *ast.PropertyDecl, classType *types.ClassType, propInfo *types.PropertyInfo, indexParamTypes []types.Type) {
	propName := prop.Name.Value
	propType := propInfo.Type

	// Write spec must be an identifier (field or method name)
	ident, ok := prop.WriteSpec.(*ast.Identifier)
	if !ok {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"property '"+propName+"' write specifier must be a field or method name"))
		return
	}

	writeSpecName := ident.Value

	// Check if it's a field (instance or class field)
	// For class properties, look in ClassVars; for instance properties, look in Fields
	var fieldType types.Type
	var found bool

	if propInfo.IsClassProperty {
		// Class property must use class variable
		fieldType, found = classType.ClassVars[pkgident.Normalize(writeSpecName)]
		if found {
			fieldOwner := a.getClassVarOwner(classType, writeSpecName)
			if fieldOwner != nil {
				visibility, hasVisibility := fieldOwner.ClassVarVisibility[pkgident.Normalize(writeSpecName)]
				if hasVisibility && !a.checkVisibility(fieldOwner, visibility, writeSpecName, "class variable") {
					a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
						fmt.Sprintf(`Field/method "%s" not found`, writeSpecName)))
					return
				}
			}
		}
		if found && !propType.Equals(fieldType) {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
				"class property '"+propName+"' write field '"+writeSpecName+"' has type "+fieldType.String()+", expected "+propType.String()))
			return
		}
	} else {
		// Instance property can only use instance field
		fieldType, found = classType.GetField(pkgident.Normalize(writeSpecName))
		if found {
			fieldOwner := a.getFieldOwner(classType, writeSpecName)
			if fieldOwner != nil {
				visibility, hasVisibility := fieldOwner.FieldVisibility[pkgident.Normalize(writeSpecName)]
				if hasVisibility && !a.checkVisibility(fieldOwner, visibility, writeSpecName, "field") {
					a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
						fmt.Sprintf(`Field/method "%s" not found`, writeSpecName)))
					return
				}
			}
		}
		if found && !propType.Equals(fieldType) {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
				"property '"+propName+"' write field '"+writeSpecName+"' has type "+fieldType.String()+", expected "+propType.String()))
			return
		}
	}

	if found {
		propInfo.WriteKind = types.PropAccessField
		propInfo.WriteSpec = writeSpecName
		return
	}

	if propInfo.IsClassProperty {
		if _, found := classType.GetField(pkgident.Normalize(writeSpecName)); found {
			a.addStructuredError(NewClassMemberExpectedError(ident.Token.Pos))
			return
		}
	}

	// Check if it's a constant (constants are read-only, so error if used as write spec)
	if _, constantFound := a.getConstantType(classType, writeSpecName); constantFound {
		a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
			"property '"+propName+"' write specifier '"+writeSpecName+"' is a constant and cannot be written to"))
		return
	}

	// If method, verify method exists with correct signature
	if methodType, found := classType.GetMethod(pkgident.Normalize(writeSpecName)); found {
		// For class properties, verify the method is a class method
		if propInfo.IsClassProperty {
			isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[pkgident.Normalize(writeSpecName)]
			if !isClassMethod {
				a.addStructuredError(NewPropertyWriteShouldBeStaticMethodError(ident.Token.Pos))
				return
			}
		}

		// Setter signature: for indexed properties, method must accept index parameters
		// plus the property value. For non-indexed, method must take only the value parameter.
		// Setter must return void.

		expectedParamCount := len(indexParamTypes) + 1 // index params + value param
		if len(methodType.Parameters) != expectedParamCount {
			a.addStructuredError(NewPropertyDeclarationArgumentCountError(prop.Token.Pos,
				"property '"+propName+"' setter method '"+writeSpecName+"' has "+
					formatInt(len(methodType.Parameters))+" "+pluralizeParam(len(methodType.Parameters))+
					", expected "+formatInt(expectedParamCount)+" "+pluralizeParam(expectedParamCount)))
			return
		}

		// Verify setter signature includes index parameters
		for i, paramType := range indexParamTypes {
			if !methodType.Parameters[i].Equals(paramType) {
				a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
					"property '"+propName+"' setter method '"+writeSpecName+"' parameter "+
						formatInt(i+1)+" has type "+methodType.Parameters[i].String()+", expected "+paramType.String()))
				return
			}
		}

		// Verify last parameter is the property value with matching type
		valueParamIndex := len(indexParamTypes)
		if !methodType.Parameters[valueParamIndex].Equals(propType) {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
				"property '"+propName+"' setter method '"+writeSpecName+"' value parameter has type "+
					methodType.Parameters[valueParamIndex].String()+", expected "+propType.String()))
			return
		}

		// Verify return type is void
		if methodType.ReturnType != types.VOID {
			a.addStructuredError(NewPropertyDeclarationTypeMismatchError(prop.Token.Pos,
				"property '"+propName+"' setter method '"+writeSpecName+"' must return void, not "+methodType.ReturnType.String()))
			return
		}

		propInfo.WriteKind = types.PropAccessMethod
		propInfo.WriteSpec = writeSpecName
		return
	}

	// Neither field nor method found
	a.addStructuredError(NewPropertyDeclarationError(prop.Token.Pos,
		"property '"+propName+"' write specifier '"+writeSpecName+"' not found in class '"+classType.Name+"'"))
}

// getConstantType looks up a constant in a class or its ancestors.
func (a *Analyzer) getConstantType(classType *types.ClassType, constantName string) (types.Type, bool) {
	if classType == nil {
		return nil, false
	}

	// Search for the constant in this class and ancestors
	current := classType
	for current != nil {
		// Case-insensitive constant lookup
		for existingName, constType := range current.ConstantTypes {
			if pkgident.Equal(existingName, constantName) {
				return constType, true
			}
		}
		current = current.Parent
	}

	return nil, false
}
