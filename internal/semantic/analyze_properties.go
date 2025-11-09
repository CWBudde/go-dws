package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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

// analyzePropertyDecl validates a property declaration and registers it in the class metadata.
func (a *Analyzer) analyzePropertyDecl(prop *ast.PropertyDecl, classType *types.ClassType) {
	propName := prop.Name.Value

	//  Check for duplicate property names within class
	if _, exists := classType.Properties[propName]; exists {
		a.addError("duplicate property '%s' in class '%s' at %s",
			propName, classType.Name, prop.Token.Pos.String())
		return
	}

	// Resolve property type
	if prop.Type == nil {
		a.addError("property '%s' missing type annotation in class '%s' at %s",
			propName, classType.Name, prop.Token.Pos.String())
		return
	}

	propType, err := a.resolveType(prop.Type.Name)
	if err != nil {
		a.addError("unknown type '%s' for property '%s' in class '%s' at %s",
			prop.Type.Name, propName, classType.Name, prop.Token.Pos.String())
		return
	}

	// Validate indexed property parameters have valid types
	isIndexed := prop.IndexParams != nil && len(prop.IndexParams) > 0
	var indexParamTypes []types.Type
	if isIndexed {
		for _, param := range prop.IndexParams {
			if param.Type == nil {
				a.addError("index parameter '%s' missing type annotation in property '%s' at %s",
					param.Name.Value, propName, prop.Token.Pos.String())
				return
			}
			paramType, err := a.resolveType(param.Type.Name)
			if err != nil {
				a.addError("unknown type '%s' for index parameter '%s' in property '%s' at %s",
					param.Type.Name, param.Name.Value, propName, prop.Token.Pos.String())
				return
			}
			indexParamTypes = append(indexParamTypes, paramType)
		}
	}

	// Create PropertyInfo to store in class metadata
	propInfo := &types.PropertyInfo{
		Name:            propName,
		Type:            propType,
		IsIndexed:       isIndexed,
		IsDefault:       prop.IsDefault,
		IsClassProperty: prop.IsClassProperty,
	}

	// Task 9.49: Register property stub before validating specs
	// This allows circular reference detection in property expressions
	classType.Properties[propName] = propInfo

	// Validate read specifier
	if prop.ReadSpec != nil {
		a.validateReadSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Validate write specifier
	if prop.WriteSpec != nil {
		a.validateWriteSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Validate default property restrictions
	if prop.IsDefault {
		// Default properties must be indexed
		if !isIndexed {
			a.addError("default property '%s' must be an indexed property at %s",
				propName, prop.Token.Pos.String())
			return
		}

		// Only one default property per class
		for existingPropName, existingProp := range classType.Properties {
			if existingProp.IsDefault && existingPropName != propName {
				a.addError("class '%s' already has default property '%s'; cannot declare another default property '%s' at %s",
					classType.Name, existingPropName, propName, prop.Token.Pos.String())
				return
			}
		}
	}

	// Property already registered at the beginning (line 68) for circular reference detection
}

// validateReadSpec validates the read specifier of a property.
// The read specifier can be:
//   - Field: A field name (identifier) - the field must exist and have matching type
//   - Method: A method name (identifier) - the method must exist with correct signature
//   - Expression: An inline expression - the expression type must match property type
func (a *Analyzer) validateReadSpec(prop *ast.PropertyDecl, classType *types.ClassType, propInfo *types.PropertyInfo, indexParamTypes []types.Type) {
	propName := prop.Name.Value
	propType := propInfo.Type

	// Check if read spec is an identifier (field or method name)
	if ident, ok := prop.ReadSpec.(*ast.Identifier); ok {
		readSpecName := ident.Value

		// Check if it's a field (instance or class field)
		// For class properties, look in ClassVars; for instance properties, look in Fields
		var fieldType types.Type
		var found bool

		if propInfo.IsClassProperty {
			// Class property must use class variable
			fieldType, found = classType.ClassVars[readSpecName]
			if found && !propType.Equals(fieldType) {
				a.addError("class property '%s' read field '%s' has type %s, expected %s at %s",
					propName, readSpecName, fieldType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}
		} else {
			// Instance property must use instance field
			fieldType, found = classType.GetField(readSpecName)
			if found && !propType.Equals(fieldType) {
				a.addError("property '%s' read field '%s' has type %s, expected %s at %s",
					propName, readSpecName, fieldType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}
		}

		if found {
			propInfo.ReadKind = types.PropAccessField
			propInfo.ReadSpec = readSpecName
			return
		}

		// If method, verify method exists with correct signature
		if methodType, found := classType.GetMethod(readSpecName); found {
			// For class properties, verify the method is a class method
			if propInfo.IsClassProperty {
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[readSpecName]
				if !isClassMethod {
					a.addError("class property '%s' read method '%s' must be a class method at %s",
						propName, readSpecName, prop.Token.Pos.String())
					return
				}
			} else {
				// For instance properties, verify the method is NOT a class method
				isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[readSpecName]
				if isClassMethod {
					a.addError("instance property '%s' read method '%s' cannot be a class method at %s",
						propName, readSpecName, prop.Token.Pos.String())
					return
				}
			}

			// Getter signature: for indexed properties, method must accept index parameters
			// and return property type. For non-indexed, method must take no parameters
			// and return property type.

			expectedParamCount := len(indexParamTypes)
			if len(methodType.Parameters) != expectedParamCount {
				a.addError("property '%s' getter method '%s' has %d %s, expected %d %s at %s",
					propName, readSpecName, len(methodType.Parameters), pluralizeParam(len(methodType.Parameters)),
					expectedParamCount, pluralizeParam(expectedParamCount),
					prop.Token.Pos.String())
				return
			}

			// Verify getter signature includes index parameters
			for i, paramType := range indexParamTypes {
				if !methodType.Parameters[i].Equals(paramType) {
					a.addError("property '%s' getter method '%s' parameter %d has type %s, expected %s at %s",
						propName, readSpecName, i+1,
						methodType.Parameters[i].String(), paramType.String(),
						prop.Token.Pos.String())
					return
				}
			}

			// Verify return type matches property type
			if !methodType.ReturnType.Equals(propType) {
				a.addError("property '%s' getter method '%s' returns %s, expected %s at %s",
					propName, readSpecName,
					methodType.ReturnType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}

			propInfo.ReadKind = types.PropAccessMethod
			propInfo.ReadSpec = readSpecName
			return
		}

		// Neither field nor method found
		a.addError("property '%s' read specifier '%s' not found in class '%s' at %s",
			propName, readSpecName, classType.Name, prop.Token.Pos.String())
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
		a.addError("property '%s' read expression has type %s, expected %s at %s",
			propName, exprType.String(), propType.String(),
			prop.Token.Pos.String())
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
		a.addError("property '%s' write specifier must be a field or method name at %s",
			propName, prop.Token.Pos.String())
		return
	}

	writeSpecName := ident.Value

	// Check if it's a field (instance or class field)
	// For class properties, look in ClassVars; for instance properties, look in Fields
	var fieldType types.Type
	var found bool

	if propInfo.IsClassProperty {
		// Class property must use class variable
		fieldType, found = classType.ClassVars[writeSpecName]
		if found && !propType.Equals(fieldType) {
			a.addError("class property '%s' write field '%s' has type %s, expected %s at %s",
				propName, writeSpecName, fieldType.String(), propType.String(),
				prop.Token.Pos.String())
			return
		}
	} else {
		// Instance property must use instance field
		fieldType, found = classType.GetField(writeSpecName)
		if found && !propType.Equals(fieldType) {
			a.addError("property '%s' write field '%s' has type %s, expected %s at %s",
				propName, writeSpecName, fieldType.String(), propType.String(),
				prop.Token.Pos.String())
			return
		}
	}

	if found {
		propInfo.WriteKind = types.PropAccessField
		propInfo.WriteSpec = writeSpecName
		return
	}

	// If method, verify method exists with correct signature
	if methodType, found := classType.GetMethod(writeSpecName); found {
		// For class properties, verify the method is a class method
		if propInfo.IsClassProperty {
			isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[writeSpecName]
			if !isClassMethod {
				a.addError("class property '%s' write method '%s' must be a class method at %s",
					propName, writeSpecName, prop.Token.Pos.String())
				return
			}
		} else {
			// For instance properties, verify the method is NOT a class method
			isClassMethod := classType.ClassMethodFlags != nil && classType.ClassMethodFlags[writeSpecName]
			if isClassMethod {
				a.addError("instance property '%s' write method '%s' cannot be a class method at %s",
					propName, writeSpecName, prop.Token.Pos.String())
				return
			}
		}

		// Setter signature: for indexed properties, method must accept index parameters
		// plus the property value. For non-indexed, method must take only the value parameter.
		// Setter must return void.

		expectedParamCount := len(indexParamTypes) + 1 // index params + value param
		if len(methodType.Parameters) != expectedParamCount {
			a.addError("property '%s' setter method '%s' has %d %s, expected %d %s at %s",
				propName, writeSpecName, len(methodType.Parameters), pluralizeParam(len(methodType.Parameters)),
				expectedParamCount, pluralizeParam(expectedParamCount),
				prop.Token.Pos.String())
			return
		}

		// Verify setter signature includes index parameters
		for i, paramType := range indexParamTypes {
			if !methodType.Parameters[i].Equals(paramType) {
				a.addError("property '%s' setter method '%s' parameter %d has type %s, expected %s at %s",
					propName, writeSpecName, i+1,
					methodType.Parameters[i].String(), paramType.String(),
					prop.Token.Pos.String())
				return
			}
		}

		// Verify last parameter is the property value with matching type
		valueParamIndex := len(indexParamTypes)
		if !methodType.Parameters[valueParamIndex].Equals(propType) {
			a.addError("property '%s' setter method '%s' value parameter has type %s, expected %s at %s",
				propName, writeSpecName,
				methodType.Parameters[valueParamIndex].String(), propType.String(),
				prop.Token.Pos.String())
			return
		}

		// Verify return type is void
		if methodType.ReturnType != types.VOID {
			a.addError("property '%s' setter method '%s' must return void, not %s at %s",
				propName, writeSpecName, methodType.ReturnType.String(),
				prop.Token.Pos.String())
			return
		}

		propInfo.WriteKind = types.PropAccessMethod
		propInfo.WriteSpec = writeSpecName
		return
	}

	// Neither field nor method found
	a.addError("property '%s' write specifier '%s' not found in class '%s' at %s",
		propName, writeSpecName, classType.Name, prop.Token.Pos.String())
}
