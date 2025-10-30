package interp

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
)

// evalProgram evaluates all statements in the program.
func (i *Interpreter) evalProgram(program *ast.Program) Value {
	var result Value

	for _, stmt := range program.Statements {
		result = i.Eval(stmt)

		// If we hit an error, stop execution
		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if i.exception != nil {
			break
		}

		// Task 8.235n: Check if exit was called at program level
		if i.exitSignal {
			i.exitSignal = false // Clear signal
			break                // Exit the program
		}
	}

	// If there's an uncaught exception, convert it to an error
	if i.exception != nil {
		exc := i.exception
		return newError("uncaught exception: %s", exc.Inspect())
	}

	return result
}

// evalVarDeclStatement evaluates a variable declaration statement.
// It defines a new variable in the current environment.
// External variables are marked with a special ExternalVarValue.
func (i *Interpreter) evalVarDeclStatement(stmt *ast.VarDeclStatement) Value {
	var value Value

	// Handle multi-identifier declarations
	// All names share the same type, but each needs to be defined separately
	// Note: Parser already validates that multi-name declarations cannot have initializers

	// Handle external variables
	if stmt.IsExternal {
		// External variables only apply to single declarations
		if len(stmt.Names) != 1 {
			return newError("external keyword cannot be used with multiple variable names")
		}
		// Store a special marker for external variables
		externalName := stmt.ExternalName
		if externalName == "" {
			externalName = stmt.Names[0].Value
		}
		value = &ExternalVarValue{
			Name:         stmt.Names[0].Value,
			ExternalName: externalName,
		}
		i.env.Define(stmt.Names[0].Value, value)
		return value
	}

	if stmt.Value != nil {
		// Task 9.177: Special handling for anonymous record literals - they need type context
		if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
			// Anonymous record literal needs explicit type
			if stmt.Type == nil {
				return newError("anonymous record literal requires explicit type annotation")
			}
			typeName := stmt.Type.Name
			recordTypeKey := "__record_type_" + typeName
			if typeVal, ok := i.env.Get(recordTypeKey); ok {
				if rtv, ok := typeVal.(*RecordTypeValue); ok {
					// Temporarily set the type name for evaluation
					recordLit.TypeName = &ast.Identifier{Value: rtv.RecordType.Name}
					value = i.Eval(recordLit)
					recordLit.TypeName = nil
				} else {
					return newError("type '%s' is not a record type", typeName)
				}
			} else {
				return newError("unknown type '%s'", typeName)
			}
		} else {
			value = i.Eval(stmt.Value)
		}
		if isError(value) {
			return value
		}

		// If declaring a subrange variable with an initializer, wrap and validate
		if stmt.Type != nil {
			typeName := stmt.Type.Name
			subrangeTypeKey := "__subrange_type_" + typeName
			if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
				if stv, ok := typeVal.(*SubrangeTypeValue); ok {
					// Extract integer value from initializer
					var intValue int
					if intVal, ok := value.(*IntegerValue); ok {
						intValue = int(intVal.Value)
					} else if srcSubrange, ok := value.(*SubrangeValue); ok {
						intValue = srcSubrange.Value
					} else {
						return newError("cannot initialize subrange type %s with %s", typeName, value.Type())
					}

					// Create subrange value and validate
					subrangeVal := &SubrangeValue{
						Value:        0, // Will be set by ValidateAndSet
						SubrangeType: stv.SubrangeType,
					}
					if err := subrangeVal.ValidateAndSet(intValue); err != nil {
						return &ErrorValue{Message: err.Error()}
					}
					value = subrangeVal
				}
			}
		}
	} else {
		// No initializer - check if we need to initialize based on type
		if stmt.Type != nil {
			typeName := stmt.Type.Name

			// Check for inline array types first
			// Inline array types have signatures like "array of Integer" or "array[1..10] of Integer"
			if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
				// Parse inline array type and create array value
				arrayType := i.parseInlineArrayType(typeName)
				if arrayType != nil {
					value = NewArrayValue(arrayType)
				} else {
					value = &NilValue{}
				}
			} else if typeVal, ok := i.env.Get("__record_type_" + typeName); ok {
				// Check if this is a record type
				if rtv, ok := typeVal.(*RecordTypeValue); ok {
					// Initialize with empty record value
					value = NewRecordValue(rtv.RecordType)
				} else {
					value = &NilValue{}
				}
			} else {
				// Check if this is an array type
				arrayTypeKey := "__array_type_" + typeName
				if typeVal, ok := i.env.Get(arrayTypeKey); ok {
					if atv, ok := typeVal.(*ArrayTypeValue); ok {
						// Initialize with empty array value
						value = NewArrayValue(atv.ArrayType)
					} else {
						value = &NilValue{}
					}
				} else {
					// Check if this is a subrange type
					subrangeTypeKey := "__subrange_type_" + typeName
					if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
						if stv, ok := typeVal.(*SubrangeTypeValue); ok {
							// Initialize with zero value (will be validated if assigned)
							value = &SubrangeValue{
								Value:        0,
								SubrangeType: stv.SubrangeType,
							}
						} else {
							value = &NilValue{}
						}
					} else {
						// Initialize basic types with their zero values
						// Proper initialization allows implicit conversions to work with target type
						switch strings.ToLower(typeName) {
						case "integer":
							value = &IntegerValue{Value: 0}
						case "float":
							value = &FloatValue{Value: 0.0}
						case "string":
							value = &StringValue{Value: ""}
						case "boolean":
							value = &BooleanValue{Value: false}
						default:
							value = &NilValue{}
						}
					}
				}
			}
		} else {
			value = &NilValue{}
		}
	}

	// Define all names with the same value/type
	// For multi-identifier declarations without initializers, each gets its own zero value
	var lastValue Value = value
	for _, name := range stmt.Names {
		// If there's an initializer, all names share the same value (but parser prevents this for multi-names)
		// If no initializer, need to create separate zero values for each variable
		var nameValue Value
		if stmt.Value != nil {
			// Single name with initializer - use the computed value
			nameValue = value
		} else {
			// No initializer - create a new zero value for each name
			// Must create separate instances to avoid aliasing
			nameValue = i.createZeroValue(stmt.Type)
		}
		i.env.Define(name.Value, nameValue)
		lastValue = nameValue
	}

	return lastValue
}

// createZeroValue creates a zero value for the given type
// This is used for multi-identifier declarations where each variable needs its own instance
func (i *Interpreter) createZeroValue(typeAnnotation *ast.TypeAnnotation) Value {
	if typeAnnotation == nil {
		return &NilValue{}
	}

	typeName := typeAnnotation.Name

	// Check for inline array types first
	if strings.HasPrefix(typeName, "array of ") || strings.HasPrefix(typeName, "array[") {
		arrayType := i.parseInlineArrayType(typeName)
		if arrayType != nil {
			return NewArrayValue(arrayType)
		}
		return &NilValue{}
	}

	// Check if this is a record type
	if typeVal, ok := i.env.Get("__record_type_" + typeName); ok {
		if rtv, ok := typeVal.(*RecordTypeValue); ok {
			return NewRecordValue(rtv.RecordType)
		}
	}

	// Check if this is an array type
	arrayTypeKey := "__array_type_" + typeName
	if typeVal, ok := i.env.Get(arrayTypeKey); ok {
		if atv, ok := typeVal.(*ArrayTypeValue); ok {
			return NewArrayValue(atv.ArrayType)
		}
	}

	// Check if this is a subrange type
	subrangeTypeKey := "__subrange_type_" + typeName
	if typeVal, ok := i.env.Get(subrangeTypeKey); ok {
		if stv, ok := typeVal.(*SubrangeTypeValue); ok {
			return &SubrangeValue{
				Value:        0,
				SubrangeType: stv.SubrangeType,
			}
		}
	}

	// Initialize basic types with their zero values
	switch strings.ToLower(typeName) {
	case "integer":
		return &IntegerValue{Value: 0}
	case "float":
		return &FloatValue{Value: 0.0}
	case "string":
		return &StringValue{Value: ""}
	case "boolean":
		return &BooleanValue{Value: false}
	default:
		return &NilValue{}
	}
}

// evalConstDecl evaluates a const declaration statement
// Constants are immutable values stored in the environment.
// Immutability is enforced at semantic analysis time, so at runtime
// we simply evaluate the value and store it like a variable.
func (i *Interpreter) evalConstDecl(stmt *ast.ConstDecl) Value {
	// Constants must have a value
	if stmt.Value == nil {
		return newError("constant '%s' must have a value", stmt.Name.Value)
	}

	// Evaluate the constant value
	var value Value

	// Task 9.177: Special handling for anonymous record literals - they need type context
	if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
		// Anonymous record literal needs explicit type
		if stmt.Type == nil {
			return newError("anonymous record literal requires explicit type annotation")
		}
		typeName := stmt.Type.Name
		recordTypeKey := "__record_type_" + typeName
		if typeVal, ok := i.env.Get(recordTypeKey); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				// Temporarily set the type name for evaluation
				recordLit.TypeName = &ast.Identifier{Value: rtv.RecordType.Name}
				value = i.Eval(recordLit)
				recordLit.TypeName = nil
			} else {
				return newError("type '%s' is not a record type", typeName)
			}
		} else {
			return newError("unknown type '%s'", typeName)
		}
	} else {
		value = i.Eval(stmt.Value)
	}
	if isError(value) {
		return value
	}

	// Store the constant in the environment
	// Note: Immutability is enforced by semantic analysis, not at runtime
	i.env.Define(stmt.Name.Value, value)
	return value
}

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
		// Regular assignment - evaluate the value to assign
		// Special handling for record literals without type names
		if recordLit, ok := stmt.Value.(*ast.RecordLiteralExpression); ok && recordLit.TypeName == nil {
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

		// Task 8.77: Records have value semantics - copy when assigning
		if recordVal, ok := value.(*RecordValue); ok {
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
		// += works with Integer, Float, String
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				return &IntegerValue{Value: l.Value + r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Integer", right.Type())
		case *FloatValue:
			if r, ok := right.(*FloatValue); ok {
				return &FloatValue{Value: l.Value + r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot add %s to Float", right.Type())
		case *StringValue:
			if r, ok := right.(*StringValue); ok {
				return &StringValue{Value: l.Value + r.Value}
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
			if r, ok := right.(*FloatValue); ok {
				return &FloatValue{Value: l.Value - r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot subtract %s from Float", right.Type())
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
			if r, ok := right.(*FloatValue); ok {
				return &FloatValue{Value: l.Value * r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot multiply Float by %s", right.Type())
		default:
			return i.newErrorWithLocation(stmt, "operator *= not supported for type %s", left.Type())
		}

	case lexer.DIVIDE_ASSIGN:
		// /= works with Integer, Float
		switch l := left.(type) {
		case *IntegerValue:
			if r, ok := right.(*IntegerValue); ok {
				if r.Value == 0 {
					return i.newErrorWithLocation(stmt, "division by zero")
				}
				return &IntegerValue{Value: l.Value / r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Integer by %s", right.Type())
		case *FloatValue:
			if r, ok := right.(*FloatValue); ok {
				if r.Value == 0.0 {
					return i.newErrorWithLocation(stmt, "division by zero")
				}
				return &FloatValue{Value: l.Value / r.Value}
			}
			return i.newErrorWithLocation(stmt, "type mismatch: cannot divide Float by %s", right.Type())
		default:
			return i.newErrorWithLocation(stmt, "operator /= not supported for type %s", left.Type())
		}

	default:
		return i.newErrorWithLocation(stmt, "unknown compound operator: %v", op)
	}
}

// evalSimpleAssignment handles simple variable assignment: x := value
func (i *Interpreter) evalSimpleAssignment(target *ast.Identifier, value Value, _ *ast.AssignmentStatement) Value {
	// Check if trying to assign to an external variable
	// Apply implicit conversion if types don't match
	// Validate subrange assignments
	if existingVal, ok := i.env.Get(target.Value); ok {
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
	// Check if the left side is a class identifier (for static assignment: TClass.Variable := value)
	if ident, ok := target.Object.(*ast.Identifier); ok {
		// Check if this identifier refers to a class
		if classInfo, exists := i.classes[ident.Value]; exists {
			// This is a class variable assignment
			fieldName := target.Member.Value
			if _, exists := classInfo.ClassVars[fieldName]; !exists {
				return i.newErrorWithLocation(stmt, "class variable '%s' not found in class '%s'", fieldName, ident.Value)
			}
			// Assign to the class variable
			classInfo.ClassVars[fieldName] = value
			return value
		}
	}

	// Not static access - evaluate the object expression for instance access
	objVal := i.Eval(target.Object)
	if isError(objVal) {
		return objVal
	}

	// Task 8.76: Check if it's a record value
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
							Token:  stmt.Token,
							Target: indexExpr,
							Value:  &ast.Identifier{Value: "__temp__"},
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

	// Task 8.54: Check if this is a property assignment (properties take precedence over fields)
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
	// Evaluate the array expression
	arrayVal := i.Eval(target.Left)
	if isError(arrayVal) {
		return arrayVal
	}

	// Evaluate the index expression
	indexVal := i.Eval(target.Index)
	if isError(indexVal) {
		return indexVal
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

// evalBlockStatement evaluates a block of statements.
func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement) Value {
	if block == nil {
		return &NilValue{}
	}

	var result Value

	for _, stmt := range block.Statements {
		result = i.Eval(stmt)

		if isError(result) {
			return result
		}

		// Check if exception is active - if so, unwind the stack
		if i.exception != nil {
			return nil
		}

		// Task 8.235o: Check for control flow signals and propagate them upward
		// These signals should propagate up to the appropriate control structure
		if i.breakSignal || i.continueSignal || i.exitSignal {
			return nil // Propagate signal upward by returning early
		}
	}

	return result
}

func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) Value {
	// Evaluate the condition
	condition := i.Eval(stmt.Condition)
	if isError(condition) {
		return condition
	}

	// Convert condition to boolean
	if isTruthy(condition) {
		return i.Eval(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return i.Eval(stmt.Alternative)
	}

	// No alternative and condition was false - return nil
	return &NilValue{}
}

// isTruthy determines if a value is considered "true" for conditional logic.
// In DWScript, only boolean true is truthy. Everything else requires explicit conversion.
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *BooleanValue:
		return v.Value
	default:
		// In DWScript, only booleans can be used in conditions
		// Non-boolean values in conditionals would be a type error
		// For now, treat non-booleans as false
		return false
	}
}

// evalWhileStatement evaluates a while loop.
// It repeatedly evaluates the condition and executes the body while the condition is true.
func (i *Interpreter) evalWhileStatement(stmt *ast.WhileStatement) Value {
	var result Value = &NilValue{}

	for {
		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true
		if !isTruthy(condition) {
			break
		}

		// Execute the body
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Task 8.235m: Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			continue
		}
		// Task 8.235m: Handle exit signal (exit from function while in loop)
		if i.exitSignal {
			// Don't clear the signal - let the function handle it
			break
		}
	}

	return result
}

// evalRepeatStatement evaluates a repeat-until loop.
// The body executes at least once, then repeats until the condition becomes true.
// This differs from while loops: the body always executes at least once,
// and the loop continues UNTIL the condition is true (not WHILE it's true).
func (i *Interpreter) evalRepeatStatement(stmt *ast.RepeatStatement) Value {
	var result Value = &NilValue{}

	for {
		// Execute the body first (repeat-until always executes at least once)
		result = i.Eval(stmt.Body)
		if isError(result) {
			return result
		}

		// Task 8.235m: Handle break/continue signals
		if i.breakSignal {
			i.breakSignal = false // Clear signal
			break
		}
		if i.continueSignal {
			i.continueSignal = false // Clear signal
			// Continue to condition check
		}
		// Task 8.235m: Handle exit signal (exit from function while in loop)
		if i.exitSignal {
			// Don't clear the signal - let the function handle it
			break
		}

		// Evaluate the condition
		condition := i.Eval(stmt.Condition)
		if isError(condition) {
			return condition
		}

		// Check if condition is true - if so, exit the loop
		// Note: repeat UNTIL condition, so we break when condition is TRUE
		if isTruthy(condition) {
			break
		}
	}

	return result
}

// evalForStatement evaluates a for loop.
// DWScript for loops iterate from start to end (or downto), with the loop variable
// scoped to the loop body. The loop variable is automatically created and managed.
func (i *Interpreter) evalForStatement(stmt *ast.ForStatement) Value {
	var result Value = &NilValue{}

	// Evaluate start value
	startVal := i.Eval(stmt.Start)
	if isError(startVal) {
		return startVal
	}

	// Evaluate end value
	endVal := i.Eval(stmt.End)
	if isError(endVal) {
		return endVal
	}

	// Both start and end must be integers for for loops
	startInt, ok := startVal.(*IntegerValue)
	if !ok {
		return newError("for loop start value must be integer, got %s", startVal.Type())
	}

	endInt, ok := endVal.(*IntegerValue)
	if !ok {
		return newError("for loop end value must be integer, got %s", endVal.Type())
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	loopEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = loopEnv

	// Define the loop variable in the loop environment
	loopVarName := stmt.Variable.Value

	// Execute the loop based on direction
	if stmt.Direction == ast.ForTo {
		// Ascending loop: for i := start to end
		for current := startInt.Value; current <= endInt.Value; current++ {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Task 8.235m: Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Task 8.235m: Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	} else {
		// Descending loop: for i := start downto end
		for current := startInt.Value; current >= endInt.Value; current-- {
			// Set the loop variable to the current value
			i.env.Define(loopVarName, &IntegerValue{Value: current})

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Task 8.235m: Handle break/continue signals
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			// Task 8.235m: Handle exit signal (exit from function while in loop)
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}
	}

	// Restore the original environment after the loop
	i.env = savedEnv

	return result
}

// evalForInStatement evaluates a for-in loop statement.
// It iterates over the elements of a collection (array, set, or string).
// The loop variable is assigned each element in turn, and the body is executed.
func (i *Interpreter) evalForInStatement(stmt *ast.ForInStatement) Value {
	var result Value = &NilValue{}

	// Evaluate the collection expression
	collectionVal := i.Eval(stmt.Collection)
	if isError(collectionVal) {
		return collectionVal
	}

	// Create a new enclosed environment for the loop variable
	// This ensures the loop variable is scoped to the loop body
	loopEnv := NewEnclosedEnvironment(i.env)
	savedEnv := i.env
	i.env = loopEnv

	loopVarName := stmt.Variable.Value

	// Type-switch on the collection type to determine iteration strategy
	switch col := collectionVal.(type) {
	case *ArrayValue:
		// Iterate over array elements
		for _, element := range col.Elements {
			// Assign the current element to the loop variable
			i.env.Define(loopVarName, element)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle control flow signals (break, continue, exit)
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	case *SetValue:
		// Iterate over set elements
		// Sets contain enum values; we iterate through the enum's ordered names
		// and check which ones are present in the set
		if col.SetType == nil || col.SetType.ElementType == nil {
			i.env = savedEnv
			return newError("invalid set type for iteration")
		}

		enumType := col.SetType.ElementType
		// Iterate through enum values in their defined order
		for _, name := range enumType.OrderedNames {
			ordinal := enumType.Values[name]
			// Check if this enum value is in the set
			if col.HasElement(ordinal) {
				// Create an enum value for this element
				enumVal := &EnumValue{
					TypeName:     enumType.Name,
					ValueName:    name,
					OrdinalValue: ordinal,
				}

				// Assign the enum value to the loop variable
				i.env.Define(loopVarName, enumVal)

				// Execute the body
				result = i.Eval(stmt.Body)
				if isError(result) {
					i.env = savedEnv // Restore environment before returning
					return result
				}

				// Handle control flow signals (break, continue, exit)
				if i.breakSignal {
					i.breakSignal = false // Clear signal
					break
				}
				if i.continueSignal {
					i.continueSignal = false // Clear signal
					continue
				}
				if i.exitSignal {
					// Don't clear the signal - let the function handle it
					break
				}
			}
		}

	case *StringValue:
		// Iterate over string characters
		// Each character becomes a single-character string
		for idx := 0; idx < len(col.Value); idx++ {
			// Create a single-character string for this iteration
			charVal := &StringValue{Value: string(col.Value[idx])}

			// Assign the character to the loop variable
			i.env.Define(loopVarName, charVal)

			// Execute the body
			result = i.Eval(stmt.Body)
			if isError(result) {
				i.env = savedEnv // Restore environment before returning
				return result
			}

			// Handle control flow signals (break, continue, exit)
			if i.breakSignal {
				i.breakSignal = false // Clear signal
				break
			}
			if i.continueSignal {
				i.continueSignal = false // Clear signal
				continue
			}
			if i.exitSignal {
				// Don't clear the signal - let the function handle it
				break
			}
		}

	default:
		// If we reach here, the semantic analyzer missed something
		// This is defensive programming
		i.env = savedEnv
		return newError("for-in loop: cannot iterate over %s", collectionVal.Type())
	}

	// Restore the original environment after the loop
	i.env = savedEnv

	return result
}

// evalCaseStatement evaluates a case statement.
// It evaluates the case expression, then checks each branch in order.
// The first branch with a matching value has its statement executed.
// If no branch matches and there's an else clause, it's executed.
func (i *Interpreter) evalCaseStatement(stmt *ast.CaseStatement) Value {
	// Evaluate the case expression
	caseValue := i.Eval(stmt.Expression)
	if isError(caseValue) {
		return caseValue
	}

	// Check each case branch in order
	for _, branch := range stmt.Cases {
		// Check each value in this branch
		for _, branchVal := range branch.Values {
			// Evaluate the branch value
			branchValue := i.Eval(branchVal)
			if isError(branchValue) {
				return branchValue
			}

			// Check if values match
			if i.valuesEqual(caseValue, branchValue) {
				// Execute this branch's statement
				return i.Eval(branch.Statement)
			}
		}
	}

	// No branch matched - execute else clause if present
	if stmt.Else != nil {
		return i.Eval(stmt.Else)
	}

	// No match and no else clause - return nil
	return &NilValue{}
}

// evalBreakStatement evaluates a break statement (Task 8.235j).
// Sets the break signal to exit the innermost loop.
func (i *Interpreter) evalBreakStatement(_ *ast.BreakStatement) Value {
	i.breakSignal = true
	return &NilValue{}
}

// evalContinueStatement evaluates a continue statement (Task 8.235k).
// Sets the continue signal to skip to the next iteration of the innermost loop.
func (i *Interpreter) evalContinueStatement(_ *ast.ContinueStatement) Value {
	i.continueSignal = true
	return &NilValue{}
}

// evalExitStatement evaluates an exit statement (Task 8.235l).
// Sets the exit signal to exit the current function.
// If at program level, sets exit signal to terminate the program.
// evalReturnStatement handles return statements in lambda expressions.
// Task 9.222: Return statements are used in shorthand lambda syntax.
//
// In shorthand lambda syntax, the parser creates a return statement:
//
//	lambda(x) => x * 2
//
// becomes:
//
//	lambda(x) begin return x * 2; end
//
// The return value is assigned to the Result variable if it exists.
func (i *Interpreter) evalReturnStatement(stmt *ast.ReturnStatement) Value {
	// Evaluate the return value
	var returnVal Value
	if stmt.ReturnValue != nil {
		returnVal = i.Eval(stmt.ReturnValue)
		if isError(returnVal) {
			return returnVal
		}
		if returnVal == nil {
			return i.newErrorWithLocation(stmt, "return expression evaluated to nil")
		}
	} else {
		returnVal = &NilValue{}
	}

	// Assign to Result variable if it exists (for functions)
	// This allows the function to return the value
	if _, exists := i.env.Get("Result"); exists {
		i.env.Set("Result", returnVal)
	}

	// Set exit signal to indicate early return
	i.exitSignal = true

	return returnVal
}

func (i *Interpreter) evalExitStatement(_ *ast.ExitStatement) Value {
	i.exitSignal = true
	// Exit doesn't return a value - the function's default return value is used
	// (or the program exits if at top level)
	return &NilValue{}
}

// valuesEqual compares two values for equality.
// This is used by case statements to match values.
func (i *Interpreter) valuesEqual(left, right Value) bool {
	// Handle same type comparisons
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *IntegerValue:
		r, ok := right.(*IntegerValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *FloatValue:
		r, ok := right.(*FloatValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *StringValue:
		r, ok := right.(*StringValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *BooleanValue:
		r, ok := right.(*BooleanValue)
		if !ok {
			return false
		}
		return l.Value == r.Value
	case *NilValue:
		return true // nil == nil
	case *RecordValue:
		r, ok := right.(*RecordValue)
		if !ok {
			return false
		}
		return i.recordsEqual(l, r)
	default:
		// For other types, use string comparison as fallback
		return left.String() == right.String()
	}
}
