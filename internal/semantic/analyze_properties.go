package semantic

import (
	"github.com/cwbudde/go-dws/ast"
	"github.com/cwbudde/go-dws/types"
)

// ============================================================================
// Property Semantic Analysis (Tasks 8.46-8.51)
// ============================================================================

// analyzePropertyDecl validates a property declaration and registers it in the class metadata.
// This function implements Tasks 8.46-8.51.
//
// Task 8.46: Register properties in class metadata
// Task 8.47: Validate getter (read specifier)
// Task 8.48: Validate setter (write specifier)
// Task 8.49: Validate indexed properties
// Task 8.50: Check for duplicate property names
// Task 8.51: Validate default property restrictions
func (a *Analyzer) analyzePropertyDecl(prop *ast.PropertyDecl, classType *types.ClassType) {
	propName := prop.Name.Value

	// Task 8.50: Check for duplicate property names within class
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

	// Task 8.49a: Validate indexed property parameters have valid types
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
		Name:      propName,
		Type:      propType,
		IsIndexed: isIndexed,
		IsDefault: prop.IsDefault,
	}

	// Task 8.47: Validate read specifier
	if prop.ReadSpec != nil {
		a.validateReadSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Task 8.48: Validate write specifier
	if prop.WriteSpec != nil {
		a.validateWriteSpec(prop, classType, propInfo, indexParamTypes)
	}

	// Task 8.51: Validate default property restrictions
	if prop.IsDefault {
		// Default properties must be indexed
		if !isIndexed {
			a.addError("default property '%s' must be an indexed property at %s",
				propName, prop.Token.Pos.String())
			return
		}

		// Only one default property per class
		for existingPropName, existingProp := range classType.Properties {
			if existingProp.IsDefault {
				a.addError("class '%s' already has default property '%s'; cannot declare another default property '%s' at %s",
					classType.Name, existingPropName, propName, prop.Token.Pos.String())
				return
			}
		}
	}

	// Task 8.46: Register property in class metadata
	classType.Properties[propName] = propInfo
}

// validateReadSpec validates the read specifier of a property (Task 8.47).
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

		// Task 8.47a: If field, verify field exists and has matching type
		if fieldType, found := classType.GetField(readSpecName); found {
			if !propType.Equals(fieldType) {
				a.addError("property '%s' read field '%s' has type %s, expected %s at %s",
					propName, readSpecName, fieldType.String(), propType.String(),
					prop.Token.Pos.String())
				return
			}
			propInfo.ReadKind = types.PropAccessField
			propInfo.ReadSpec = readSpecName
			return
		}

		// Task 8.47b: If method, verify method exists with correct signature
		if methodType, found := classType.GetMethod(readSpecName); found {
			// Getter signature: for indexed properties, method must accept index parameters
			// and return property type. For non-indexed, method must take no parameters
			// and return property type.

			expectedParamCount := len(indexParamTypes)
			if len(methodType.Parameters) != expectedParamCount {
				a.addError("property '%s' getter method '%s' has %d parameters, expected %d at %s",
					propName, readSpecName, len(methodType.Parameters), expectedParamCount,
					prop.Token.Pos.String())
				return
			}

			// Task 8.49b: Verify getter signature includes index parameters
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

	// Task 8.47c: If expression, validate expression type matches property type
	// Note: For now, we store the expression as a string. Full expression validation
	// would require analyzing the expression in the context of the class.
	// This is a simplified implementation that just marks it as expression-based.
	propInfo.ReadKind = types.PropAccessExpression
	propInfo.ReadSpec = prop.ReadSpec.String()
}

// validateWriteSpec validates the write specifier of a property (Task 8.48).
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

	// Task 8.48a: If field, verify field exists and has matching type
	if fieldType, found := classType.GetField(writeSpecName); found {
		if !propType.Equals(fieldType) {
			a.addError("property '%s' write field '%s' has type %s, expected %s at %s",
				propName, writeSpecName, fieldType.String(), propType.String(),
				prop.Token.Pos.String())
			return
		}
		propInfo.WriteKind = types.PropAccessField
		propInfo.WriteSpec = writeSpecName
		return
	}

	// Task 8.48b: If method, verify method exists with correct signature
	if methodType, found := classType.GetMethod(writeSpecName); found {
		// Setter signature: for indexed properties, method must accept index parameters
		// plus the property value. For non-indexed, method must take only the value parameter.
		// Setter must return void.

		expectedParamCount := len(indexParamTypes) + 1 // index params + value param
		if len(methodType.Parameters) != expectedParamCount {
			a.addError("property '%s' setter method '%s' has %d parameters, expected %d at %s",
				propName, writeSpecName, len(methodType.Parameters), expectedParamCount,
				prop.Token.Pos.String())
			return
		}

		// Task 8.49b: Verify setter signature includes index parameters
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
