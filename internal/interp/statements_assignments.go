package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// This file contains assignment statement evaluation (simple, member, index, compound).

// evalAssignmentStatement evaluates an assignment statement.
// It updates an existing variable's value or sets an object/array element.
// Supports: x := value, obj.field := value, arr[i] := value
// Also supports compound assignments: x += value, x -= value, x *= value, x /= value
func (i *Interpreter) evalAssignmentStatement(stmt *ast.AssignmentStatement) Value {
	// Check if this is a compound assignment
	isCompound := stmt.Operator != lexer.ASSIGN && stmt.Operator != lexer.TokenType(0)

	var value Value

	if isCompound {
		// For compound assignments, we need to:
		// 1. Read the current value
		// 2. Evaluate the RHS
		// 3. Apply the operation
		// 4. Store the result

		// Read current value
		currentValue := i.Eval(stmt.Target)
		if isError(currentValue) {
			return currentValue
		}

		// Evaluate the RHS
		rhsValue := i.Eval(stmt.Value)
		if isError(rhsValue) {
			return rhsValue
		}

		// Apply the compound operation
		value = i.applyCompoundOperation(stmt.Operator, currentValue, rhsValue, stmt)
		if isError(value) {
			return value
		}
	} else {
		handledLiteral := false
		// Regular assignment - evaluate the value to assign with potential context
		if arrayLit, ok := stmt.Value.(*ast.ArrayLiteralExpression); ok {
			handledLiteral = true
			var expected *types.ArrayType
			if targetIdent, ok := stmt.Target.(*ast.Identifier); ok {
				if existingVal, exists := i.env.Get(targetIdent.Value); exists {
					if arrVal, ok := existingVal.(*ArrayValue); ok {
						expected = arrVal.ArrayType
					}
				}
			}
			value = i.evalArrayLiteralWithExpected(arrayLit, expected)
			if isError(value) {
				return value
			}
		} else if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			// This is an untyped record literal - get type from target variable if it's a simple identifier
			if targetIdent, ok := stmt.Target.(*ast.Identifier); ok {
				targetVar, exists := i.env.Get(targetIdent.Value)
				if exists {
					if recVal, ok := targetVar.(*RecordValue); ok {
						// Set the type name in the literal temporarily
						recordLit.TypeName = &ast.Identifier{Value: recVal.RecordType.Name}
						value = i.Eval(recordLit)
						recordLit.TypeName = nil
					} else {
						value = i.Eval(stmt.Value)
					}
				} else {
					value = i.Eval(stmt.Value)
				}
			} else {
				value = i.Eval(stmt.Value)
			}
		} else {
			value = i.Eval(stmt.Value)
		}

		if isError(value) {
			return value
		}

		// Records have value semantics - copy when assigning
		if recordVal, ok := value.(*RecordValue); ok && !handledLiteral {
			value = recordVal.Copy()
		}
	}

	// Handle different target types
	switch target := stmt.Target.(type) {
	case *ast.Identifier:
		// Simple variable assignment: x := value or x += value
		return i.evalSimpleAssignment(target, value, stmt)

	case *ast.MemberAccessExpression:
		// Member assignment: obj.field := value or obj.field += value
		return i.evalMemberAssignment(target, value, stmt)

	case *ast.IndexExpression:
		// Array index assignment: arr[i] := value or arr[i] += value
		return i.evalIndexAssignment(target, value, stmt)

	default:
		return i.newErrorWithLocation(stmt, "invalid assignment target type: %T", target)
	}
}

// applyCompoundOperation applies a compound assignment operation (+=, -=, *=, /=).
func (i *Interpreter) applyCompoundOperation(op lexer.TokenType, left, right Value, stmt *ast.AssignmentStatement) Value {
	switch op {
	case lexer.PLUS_ASSIGN:
		// Task 9.14: Check for class operator overrides first
		if objInst, ok := left.(*ObjectInstance); ok {
			// Check if the class has an operator override for +=
			result := i.tryCallClassOperator(objInst, "+=", []Value{right}, stmt)
			if result != nil {
				// Operator was found and called (either successfully or with error)
				return result
			}
			// No operator override - fall through to default error
		}

		// += works with Integer, Float, String
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value + r.Value}
			}
			// Allow implicit Float to Integer conversion would lose precision, not allowed
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Integer", right.Type())
		case *FloatValue:
			// Support Float + Float and Float + Integer (with implicit conversion)
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value + r.Value}
			case *IntegerValue:
				// Implicit Integer to Float conversion
				return &FloatValue{Value: l.Value + float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Float", right.Type())
			}
		case *StringValue:
			if r, ok := right.(*StringValue); ok {
				return &StringValue{Value: l.Value + r.Value}
			}
			// Task 9.24.2: Handle Variant-to-String conversion for array of const elements
			if variantVal, ok := right.(*VariantValue); ok {
				// Unwrap the variant and convert to string
				innerVal, ok := unboxVariant(variantVal)
				if !ok {
					return i.newErrorWithLocation(stmt, "failed to unbox variant")
				}
				strVal := i.convertToString(innerVal)
				return &StringValue{Value: l.Value + strVal}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to String", right.Type())
		default:
			return i.newErrorWithLocation(stmt, "operator += not supported for type %s", left.Type())
		}

	case lexer.MINUS_ASSIGN:
		// -= works with Integer, Float
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value - r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot subtract %s from Integer", right.Type())
		case *FloatValue:
			// Support Float - Float and Float - Integer (with implicit conversion)
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value - r.Value}
			case *IntegerValue:
				// Implicit Integer to Float conversion
				return &FloatValue{Value: l.Value - float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot subtract %s from Float", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator -= not supported for type %s", left.Type())
		}

	case lexer.TIMES_ASSIGN:
		// *= works with Integer, Float
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value * r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot multiply Integer by %s", right.Type())
		case *FloatValue:
			// Support Float * Float and Float * Integer (with implicit conversion)
			switch r := right.(type) {
			case *FloatValue:
				return &FloatValue{Value: l.Value * r.Value}
			case *IntegerValue:
				// Implicit Integer to Float conversion
				return &FloatValue{Value: l.Value * float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot multiply Float by %s", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator *= not supported for type %s", left.Type())
		}

	case lexer.DIVIDE_ASSIGN:
		// /= works with Integer, Float
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				if r.Value == 0 {
					// Task 9.111: Enhanced error with operand values
					return i.NewRuntimeError(
						stmt,
						"division_by_zero",
						fmt.Sprintf("Division by zero: %d /= %d", l.Value, r.Value),
						map[string]string{
							"left":  fmt.Sprintf("%d", l.Value),
							"right": fmt.Sprintf("%d", r.Value),
						},
					)
				}
				return &IntegerValue{Value: l.Value / r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Integer by %s", right.Type())
		case *FloatValue:
			// Support Float / Float and Float / Integer (with implicit conversion)
			switch r := right.(type) {
			case *FloatValue:
				if r.Value == 0.0 {
					// Task 9.111: Enhanced error with operand values
					return i.NewRuntimeError(
						stmt,
						"division_by_zero",
						fmt.Sprintf("Division by zero: %v /= %v", l.Value, r.Value),
						map[string]string{
							"left":  fmt.Sprintf("%v", l.Value),
							"right": fmt.Sprintf("%v", r.Value),
						},
					)
				}
				return &FloatValue{Value: l.Value / r.Value}
			case *IntegerValue:
				// Implicit Integer to Float conversion
				if r.Value == 0 {
					// Task 9.111: Enhanced error with operand values
					return i.NewRuntimeError(
						stmt,
						"division_by_zero",
						fmt.Sprintf("Division by zero: %v /= %d", l.Value, r.Value),
						map[string]string{
							"left":  fmt.Sprintf("%v", l.Value),
							"right": fmt.Sprintf("%d", r.Value),
						},
					)
				}
				return &FloatValue{Value: l.Value / float64(r.Value)}
			default:
				return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Float by %s", right.Type())
			}
		default:
			return i.newErrorWithLocation(stmt, "operator /= not supported for type %s", left.Type())
		}

	default:
		return i.newErrorWithLocation(stmt, "unknown compound operator: %v", op)
	}
}

// evalSimpleAssignment handles simple variable assignment: x := value
func (i *Interpreter) evalSimpleAssignment(target *ast.Identifier, value Value, _ *ast.AssignmentStatement) Value {
	// Task 9.35: Check if target is a var parameter (ReferenceValue)
	if existingVal, ok := i.env.Get(target.Value); ok {
		if refVal, isRef := existingVal.(*ReferenceValue); isRef {
			// This is a var parameter - write through the reference
			// First get the current value to check type compatibility
			currentVal, err := refVal.Dereference()
			if err != nil {
				return &ErrorValue{Message: err.Error()}
			}

			// Try implicit conversion if types don't match
			targetType := currentVal.Type()
			sourceType := value.Type()
			if targetType != sourceType {
				if converted, ok := i.tryImplicitConversion(value, targetType); ok {
					value = converted
				}
			}

			// Box value if target is a Variant
			if targetType == "VARIANT" && sourceType != "VARIANT" {
				value = boxVariant(value)
			}

			// Task 9.1.5: Handle interface reference counting when assigning through var parameters
			// Release the old reference if the target currently holds an interface
			if oldIntf, isOldIntf := currentVal.(*InterfaceInstance); isOldIntf {
				i.ReleaseInterfaceReference(oldIntf)
			}

			// If assigning an interface, increment ref count for the new reference
			if intfInst, isIntf := value.(*InterfaceInstance); isIntf {
				// Increment ref count because the target variable gets a new reference
				if intfInst.Object != nil {
					intfInst.Object.RefCount++
				}
			}

			// Write through the reference
			if err := refVal.Assign(value); err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			return value
		}

		if extVar, isExternal := existingVal.(*ExternalVarValue); isExternal {
			return newError("Unsupported external variable assignment: %s", extVar.Name)
		}

		// Check if assigning to a subrange variable
		if subrangeVal, isSubrange := existingVal.(*SubrangeValue); isSubrange {
			// Extract integer value from source
			var intValue int
			if intVal, ok := value.(*IntegerValue); ok {
				intValue = int(intVal.Value)
			} else if srcSubrange, ok := value.(*SubrangeValue); ok {
				// Assigning from another subrange - extract the value
				intValue = srcSubrange.Value
			} else {
				return newError("cannot assign %s to subrange type %s", value.Type(), subrangeVal.SubrangeType.Name)
			}

			// Validate the value is in range
			if err := subrangeVal.ValidateAndSet(intValue); err != nil {
				return &ErrorValue{Message: err.Error()}
			}
			return subrangeVal
		}

		// Try implicit conversion if types don't match
		targetType := existingVal.Type()
		sourceType := value.Type()
		if targetType != sourceType {
			if converted, ok := i.tryImplicitConversion(value, targetType); ok {
				value = converted
			}
		}

		// Task 9.227: Box value if target is a Variant
		if targetType == "VARIANT" && sourceType != "VARIANT" {
			value = boxVariant(value)
		}

		// Task 9.1.5: Handle object variable assignment - manage ref count
		if objInst, isObj := existingVal.(*ObjectInstance); isObj {
			// Variable currently holds an object
			if _, isNil := value.(*NilValue); isNil {
				// Setting object variable to nil - decrement ref count
				objInst.RefCount--
				if objInst.RefCount <= 0 {
					// Call destructor if ref count reaches 0
					destructor := objInst.Class.lookupMethod("Destroy")
					if destructor != nil {
						destructorEnv := NewEnclosedEnvironment(i.env)
						destructorEnv.Define("Self", objInst)
						prevEnv := i.env
						i.env = destructorEnv
						i.Eval(destructor.Body)
						i.env = prevEnv
					}
				}
			} else if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
				// Replacing old object with new object
				// Decrement old object's ref count
				objInst.RefCount--
				if objInst.RefCount <= 0 {
					destructor := objInst.Class.lookupMethod("Destroy")
					if destructor != nil {
						destructorEnv := NewEnclosedEnvironment(i.env)
						destructorEnv.Define("Self", objInst)
						prevEnv := i.env
						i.env = destructorEnv
						i.Eval(destructor.Body)
						i.env = prevEnv
					}
				}
				// Increment new object's ref count
				newObj.RefCount++
			}
		} else {
			// Variable doesn't currently hold an object (could be nil, new variable, etc.)
			// If we're assigning an object, increment its ref count
			// BUT: Don't increment if the target is an interface - NewInterfaceInstance will do it
			if newObj, isNewObj := value.(*ObjectInstance); isNewObj {
				if _, isIface := existingVal.(*InterfaceInstance); !isIface {
					// Not an interface variable, so increment ref count
					newObj.RefCount++
				}
			}
		}

		// Task 9.16.2: Wrap object instances in InterfaceInstance when assigning to interface variables
		if ifaceInst, isIface := existingVal.(*InterfaceInstance); isIface {
			// Task 9.1.5: Release the old interface reference before assigning new value
			// This decrements ref count and calls destructor if ref count reaches 0
			i.ReleaseInterfaceReference(ifaceInst)

			// Target is an interface variable - wrap the value if it's an object
			if objInst, ok := value.(*ObjectInstance); ok {
				// Assigning an object to an interface variable - wrap it
				value = NewInterfaceInstance(ifaceInst.Interface, objInst)
			} else if _, isNil := value.(*NilValue); isNil {
				// Assigning nil to interface - create interface instance with nil object
				// No need to increment ref count since object is nil
				value = &InterfaceInstance{
					Interface: ifaceInst.Interface,
					Object:    nil,
				}
			} else if srcIface, isSrcIface := value.(*InterfaceInstance); isSrcIface {
				// Assigning interface to interface
				// Task 9.1.5: Increment ref count on the underlying object (if not nil)
				// This implements copy semantics - both variables will hold references
				if srcIface.Object != nil {
					srcIface.Object.RefCount++
				}
				// Use the underlying object but with the target interface type
				value = &InterfaceInstance{
					Interface: ifaceInst.Interface,
					Object:    srcIface.Object,
				}
			}
		}
	}

	// First try to set in current environment
	err := i.env.Set(target.Value, value)
	if err == nil {
		return value
	}

	// Not in environment - check if we're in a method context and this is a field/class variable
	// Check if Self is bound (instance method)
	selfVal, selfOk := i.env.Get("Self")
	if selfOk {
		if obj, ok := AsObject(selfVal); ok {
			// Check if it's an instance field
			if _, exists := obj.Class.Fields[target.Value]; exists {
				obj.SetField(target.Value, value)
				return value
			}
			// Check if it's a class variable
			if _, exists := obj.Class.ClassVars[target.Value]; exists {
				obj.Class.ClassVars[target.Value] = value
				return value
			}
			// Task 9.32b: Check if it's a property (properties can be assigned without Self.)
			if propInfo := obj.Class.lookupProperty(target.Value); propInfo != nil {
				// For field-backed properties, write the field directly to avoid recursion
				if propInfo.WriteKind == types.PropAccessField {
					// Check if WriteSpec is actually a field (not a method)
					if _, isField := obj.Class.Fields[propInfo.WriteSpec]; isField {
						obj.SetField(propInfo.WriteSpec, value)
						return value
					}
				}
				// For method-backed properties, use evalPropertyWrite
				return i.evalPropertyWrite(obj, propInfo, value, target)
			}
		}
	}

	// Check if __CurrentClass__ is bound (class method)
	currentClassVal, hasCurrentClass := i.env.Get("__CurrentClass__")
	if hasCurrentClass {
		if classInfo, ok := currentClassVal.(*ClassInfoValue); ok {
			// Check if it's a class variable
			if _, exists := classInfo.ClassInfo.ClassVars[target.Value]; exists {
				classInfo.ClassInfo.ClassVars[target.Value] = value
				return value
			}
		}
	}

	// Still not found - return original error
	return newError("undefined variable: %s", target.Value)
}

// evalMemberAssignment handles member assignment: obj.field := value or TClass.Variable := value
func (i *Interpreter) evalMemberAssignment(target *ast.MemberAccessExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Check if the left side is a class identifier (for static assignment: TClass.Variable := value or TClass.Property := value)
	if ident, ok := target.Object.(*ast.Identifier); ok {
		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			memberName := target.Member.Value

			// Check if this is a class property assignment (properties take precedence)
			if propInfo := classInfo.lookupProperty(memberName); propInfo != nil && propInfo.IsClassProperty {
				return i.evalClassPropertyWrite(classInfo, propInfo, value, stmt)
			}

			// Otherwise, try class variable assignment
			if _, exists := classInfo.ClassVars[memberName]; !exists {
				return i.newErrorWithLocation(stmt, "class variable '%s' not found in class '%s'", memberName, ident.Value)
			}
			// Assign to the class variable
			classInfo.ClassVars[memberName] = value
			return value
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(target.Object)
	if isError(objVal) {
		return objVal
	}

	// Check if it's a record value
	if recordVal, ok := objVal.(*RecordValue); ok {
		fieldName := target.Member.Value
		// Verify field exists in the record type
		if _, exists := recordVal.RecordType.Fields[fieldName]; !exists {
			return i.newErrorWithLocation(stmt, "field '%s' not found in record '%s'", fieldName, recordVal.RecordType.Name)
		}

		// Set the field value
		recordVal.Fields[fieldName] = value
		return value
	}

	// Special case: If objVal is NilValue and target.Object is an IndexExpression,
	// we might be trying to assign to an uninitialized record array element.
	// Auto-initialize the record and retry.
	if _, isNil := objVal.(*NilValue); isNil {
		if indexExpr, ok := target.Object.(*ast.IndexExpression); ok {
			// Evaluate the array
			arrayVal := i.Eval(indexExpr.Left)
			if isError(arrayVal) {
				return arrayVal
			}

			if arrVal, ok := arrayVal.(*ArrayValue); ok {
				// Check if the element type is a record
				if arrVal.ArrayType != nil && arrVal.ArrayType.ElementType != nil {
					if recordType, ok := arrVal.ArrayType.ElementType.(*types.RecordType); ok {
						// Auto-initialize a new record
						newRecord := &RecordValue{
							RecordType: recordType,
							Fields:     make(map[string]Value),
						}

						// Assign it to the array element using evalIndexAssignment
						assignStmt := &ast.AssignmentStatement{
							BaseNode: ast.BaseNode{Token: stmt.Token},
							Target:   indexExpr,
							Value:    &ast.Identifier{Value: "__temp__"},
						}

						// Temporarily store the record
						tempResult := i.evalIndexAssignment(indexExpr, newRecord, assignStmt)
						if isError(tempResult) {
							return tempResult
						}

						// Now retry the member assignment with the initialized record
						objVal = newRecord
					}
				}
			}
		}
	}

	// Re-check if it's a record value after potential initialization
	if recordVal, ok := objVal.(*RecordValue); ok {
		fieldName := target.Member.Value
		// Verify field exists in the record type
		if _, exists := recordVal.RecordType.Fields[fieldName]; !exists {
			return i.newErrorWithLocation(stmt, "field '%s' not found in record '%s'", fieldName, recordVal.RecordType.Name)
		}

		// Set the field value
		recordVal.Fields[fieldName] = value
		return value
	}

	// Check if it's an object instance
	obj, ok := AsObject(objVal)
	if !ok {
		return i.newErrorWithLocation(stmt, "cannot assign to field of non-object type '%s'", objVal.Type())
	}

	memberName := target.Member.Value

	// Check if this is a property assignment (properties take precedence over fields)
	if propInfo := obj.Class.lookupProperty(memberName); propInfo != nil {
		return i.evalPropertyWrite(obj, propInfo, value, stmt)
	}

	// Not a property - try direct field assignment
	// Verify field exists in the class
	if _, exists := obj.Class.Fields[memberName]; !exists {
		return i.newErrorWithLocation(stmt, "field '%s' not found in class '%s'", memberName, obj.Class.Name)
	}

	// Set the field value
	obj.SetField(memberName, value)
	return value
}

// evalIndexAssignment handles array index assignment: arr[i] := value
func (i *Interpreter) evalIndexAssignment(target *ast.IndexExpression, value Value, stmt *ast.AssignmentStatement) Value {
	// Task 9.2d: Check if this might be a multi-index property write
	// We only flatten indices if the base is a MemberAccessExpression (property access)
	// For regular array writes like arr[i][j] := value, we process each level separately
	base, indices := collectIndices(target)

	// Task 9.2b + 9.2d: Check if this is indexed property write: obj.Property[index1, index2, ...] := value
	// Only flatten indices for property access, not for regular arrays
	if memberAccess, ok := base.(*ast.MemberAccessExpression); ok {
		// Evaluate the object being accessed
		objVal := i.Eval(memberAccess.Object)
		if isError(objVal) {
			return objVal
		}

		// Check if it's a class instance with an indexed property
		if obj, ok := AsObject(objVal); ok {
			propInfo := obj.Class.lookupProperty(memberAccess.Member.Value)
			if propInfo != nil && propInfo.IsIndexed {
				// This is a multi-index property write: flatten and evaluate ALL indices
				indexVals := make([]Value, len(indices))
				for idx, indexExpr := range indices {
					indexVals[idx] = i.Eval(indexExpr)
					if isError(indexVals[idx]) {
						return indexVals[idx]
					}
				}

				// Call indexed property write with all indices
				return i.evalIndexedPropertyWrite(obj, propInfo, indexVals, value, stmt)
			}
		}
	}

	// Not a property access - this is regular array indexing
	// Process ONLY the outermost index, not all nested indices
	// This allows FData[x][y] := value to work as: (FData[x])[y] := value
	arrayVal := i.Eval(target.Left)
	if isError(arrayVal) {
		return arrayVal
	}

	// Evaluate the index for this level only
	indexVal := i.Eval(target.Index)
	if isError(indexVal) {
		return indexVal
	}

	// Task 9.16: Check if left side is an object with a default property
	// This allows obj[index] := value to be equivalent to obj.DefaultProperty[index] := value
	if obj, ok := AsObject(arrayVal); ok {
		defaultProp := obj.Class.getDefaultProperty()
		if defaultProp != nil {
			// Route to the default indexed property write
			return i.evalIndexedPropertyWrite(obj, defaultProp, []Value{indexVal}, value, stmt)
		}
	}

	// Index must be an integer
	indexInt, ok := indexVal.(*IntegerValue)
	if !ok {
		return i.newErrorWithLocation(stmt, "array index must be an integer, got %s", indexVal.Type())
	}
	index := int(indexInt.Value)

	// Check if left side is an array
	arrayValue, ok := arrayVal.(*ArrayValue)
	if !ok {
		return i.newErrorWithLocation(stmt, "cannot index type %s", arrayVal.Type())
	}

	// Perform bounds checking and get physical index
	if arrayValue.ArrayType == nil {
		return i.newErrorWithLocation(stmt, "array has no type information")
	}

	var physicalIndex int
	if arrayValue.ArrayType.IsStatic() {
		// Static array: check bounds and adjust for low bound
		lowBound := *arrayValue.ArrayType.LowBound
		highBound := *arrayValue.ArrayType.HighBound

		if index < lowBound || index > highBound {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (bounds are %d..%d)", index, lowBound, highBound)
		}

		physicalIndex = index - lowBound
	} else {
		// Dynamic array: zero-based indexing
		if index < 0 || index >= len(arrayValue.Elements) {
			return i.newErrorWithLocation(stmt, "array index out of bounds: %d (array length is %d)", index, len(arrayValue.Elements))
		}

		physicalIndex = index
	}

	// Check physical bounds
	if physicalIndex < 0 || physicalIndex >= len(arrayValue.Elements) {
		return i.newErrorWithLocation(stmt, "array index out of bounds: physical index %d, length %d", physicalIndex, len(arrayValue.Elements))
	}

	// Update the array element
	arrayValue.Elements[physicalIndex] = value

	return value
}
