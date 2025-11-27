package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Value Creation Adapter Methods =====

// CreateExternalVar creates an external variable reference value.
func (i *Interpreter) CreateExternalVar(varName, externalName string) evaluator.Value {
	return &ExternalVarValue{
		Name:         varName,
		ExternalName: externalName,
	}
}

// ResolveArrayTypeNode resolves an array type from an AST ArrayTypeNode.
func (i *Interpreter) ResolveArrayTypeNode(arrayNode ast.Node) (any, error) {
	arrNode, ok := arrayNode.(*ast.ArrayTypeNode)
	if !ok {
		return nil, fmt.Errorf("expected ArrayTypeNode")
	}

	arrType := i.resolveArrayTypeNode(arrNode)
	if arrType == nil {
		return nil, fmt.Errorf("failed to resolve array type")
	}
	return arrType, nil
}

// CreateRecordZeroValue creates a zero-initialized record value.
func (i *Interpreter) CreateRecordZeroValue(recordTypeName string) (evaluator.Value, error) {
	normalizedName := ident.Normalize(recordTypeName)
	recordTypeKey := "__record_type_" + normalizedName
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil, fmt.Errorf("record type '%s' not found", recordTypeName)
	}

	rtv, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a record type", recordTypeName)
	}

	return i.createRecordValue(rtv.RecordType, rtv.Methods), nil
}

// CreateArrayZeroValue creates a zero-initialized array value.
// Task 3.5.69c: Migrated to use TypeSystem instead of environment lookup.
func (i *Interpreter) CreateArrayZeroValue(arrayTypeName string) (evaluator.Value, error) {
	arrayType := i.typeSystem.LookupArrayType(arrayTypeName)
	if arrayType == nil {
		return nil, fmt.Errorf("array type '%s' not found", arrayTypeName)
	}

	return NewArrayValue(arrayType), nil
}

// CreateSetZeroValue creates an empty set value.
func (i *Interpreter) CreateSetZeroValue(setTypeName string) (evaluator.Value, error) {
	setType := i.parseInlineSetType(setTypeName)
	if setType == nil {
		return nil, fmt.Errorf("invalid set type: %s", setTypeName)
	}
	return NewSetValue(setType), nil
}

// CreateSubrangeZeroValue creates a zero-initialized subrange value.
func (i *Interpreter) CreateSubrangeZeroValue(subrangeTypeName string) (evaluator.Value, error) {
	normalizedName := ident.Normalize(subrangeTypeName)
	subrangeTypeKey := "__subrange_type_" + normalizedName
	typeVal, ok := i.env.Get(subrangeTypeKey)
	if !ok {
		return nil, fmt.Errorf("subrange type '%s' not found", subrangeTypeName)
	}

	stv, ok := typeVal.(*SubrangeTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a subrange type", subrangeTypeName)
	}

	return &SubrangeValue{
		Value:        stv.SubrangeType.LowBound,
		SubrangeType: stv.SubrangeType,
	}, nil
}

// CreateInterfaceZeroValue creates a nil interface instance.
func (i *Interpreter) CreateInterfaceZeroValue(interfaceName string) (evaluator.Value, error) {
	ifaceInfo, exists := i.interfaces[ident.Normalize(interfaceName)]
	if !exists {
		return nil, fmt.Errorf("interface '%s' not found", interfaceName)
	}

	return &InterfaceInstance{
		Interface: ifaceInfo,
		Object:    nil,
	}, nil
}

// CreateClassZeroValue creates a typed nil value for a class.
// Task 3.5.46: Use TypeSystem for class lookup.
func (i *Interpreter) CreateClassZeroValue(className string) (evaluator.Value, error) {
	if !i.typeSystem.HasClass(className) {
		return nil, fmt.Errorf("class '%s' not found", className)
	}

	return &NilValue{ClassType: className}, nil
}

// ===== Task 3.5.40: Record Literal Adapter Method Implementations =====

// CreateRecordValue creates a record value with field initialization.
func (i *Interpreter) CreateRecordValue(recordTypeName string, fieldValues map[string]evaluator.Value) (evaluator.Value, error) {
	normalizedName := ident.Normalize(recordTypeName)
	recordTypeKey := "__record_type_" + normalizedName
	typeVal, ok := i.env.Get(recordTypeKey)
	if !ok {
		return nil, fmt.Errorf("record type '%s' not found", recordTypeName)
	}

	recordTypeValue, ok := typeVal.(*RecordTypeValue)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not a record type", recordTypeName)
	}

	recordType := recordTypeValue.RecordType

	// Create the record value with methods
	// Task 3.5.42: Updated to use RecordMetadata
	recordValue := &RecordValue{
		RecordType: recordType,
		Fields:     make(map[string]Value),
		Metadata:   recordTypeValue.Metadata,
		Methods:    recordTypeValue.Methods, // Deprecated: backward compatibility
	}

	// Copy provided field values (already evaluated)
	for fieldName, fieldValue := range fieldValues {
		fieldNameLower := ident.Normalize(fieldName)
		// Validate field exists
		if _, exists := recordType.Fields[fieldNameLower]; !exists {
			return nil, fmt.Errorf("field '%s' does not exist in record type '%s'", fieldName, recordType.Name)
		}
		// Convert evaluator.Value to internal Value
		recordValue.Fields[fieldNameLower] = fieldValue.(Value)
	}

	// Initialize remaining fields with field initializers or default values
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
			if fieldDecl, hasDecl := recordTypeValue.FieldDecls[fieldName]; hasDecl && fieldDecl.InitValue != nil {
				// Evaluate the field initializer
				fieldValue = i.Eval(fieldDecl.InitValue)
				if isError(fieldValue) {
					return nil, fmt.Errorf("error evaluating field initializer for '%s': %s", fieldName, fieldValue.(*ErrorValue).Message)
				}
			}

			// If no initializer, use getZeroValueForType
			if fieldValue == nil {
				fieldValue = getZeroValueForType(fieldType, methodsLookup)

				// Handle interface-typed fields specially
				if intfValue := i.initializeInterfaceField(fieldType); intfValue != nil {
					fieldValue = intfValue
				}
			}

			recordValue.Fields[fieldName] = fieldValue
		}
	}

	return recordValue, nil
}

// GetZeroValueForType returns the zero/default value for a given type.
func (i *Interpreter) GetZeroValueForType(typeInfo any) evaluator.Value {
	t, ok := typeInfo.(types.Type)
	if !ok {
		return &NilValue{}
	}

	methodsLookup := func(rt *types.RecordType) map[string]*ast.FunctionDecl {
		key := "__record_type_" + ident.Normalize(rt.Name)
		if typeVal, ok := i.env.Get(key); ok {
			if rtv, ok := typeVal.(*RecordTypeValue); ok {
				return rtv.Methods
			}
		}
		return nil
	}

	return getZeroValueForType(t, methodsLookup)
}

// ===== Exception Handling Adapter Methods =====

// MatchesExceptionType checks if an exception matches a given type expression.
func (i *Interpreter) MatchesExceptionType(exc interface{}, typeExpr ast.TypeExpression) bool {
	excVal, ok := exc.(*ExceptionValue)
	if !ok {
		return false
	}
	return i.matchesExceptionType(excVal, typeExpr)
}

// GetExceptionInstance returns the ObjectInstance from an exception.
func (i *Interpreter) GetExceptionInstance(exc interface{}) evaluator.Value {
	excVal, ok := exc.(*ExceptionValue)
	if !ok {
		return nil
	}
	return excVal.Instance
}

// CreateExceptionFromObject creates an ExceptionValue from an object instance.
func (i *Interpreter) CreateExceptionFromObject(obj evaluator.Value, ctx *evaluator.ExecutionContext, pos any) interface{} {
	// Should be an object instance
	objInst, ok := obj.(*ObjectInstance)
	if !ok {
		panic(fmt.Sprintf("runtime error: raise requires exception object, got %s", obj.Type()))
	}

	// Get the class info
	classInfo := objInst.Class

	// Extract message from the object's Message field
	message := ""
	if msgVal, ok := objInst.Fields["Message"]; ok {
		if strVal, ok := msgVal.(*StringValue); ok {
			message = strVal.Value
		}
	}

	// Capture current call stack from context
	callStack := make(errors.StackTrace, len(ctx.CallStack()))
	copy(callStack, ctx.CallStack())

	// Get position
	var excPos *lexer.Position
	if p, ok := pos.(lexer.Position); ok {
		excPos = &p
	} else if p, ok := pos.(*lexer.Position); ok {
		excPos = p
	}

	return &ExceptionValue{
		ClassInfo: classInfo,
		Message:   message,
		Instance:  objInst,
		Position:  excPos,
		CallStack: callStack,
	}
}

// ===== Statement Evaluation Adapter Methods =====

// EvalBlockStatement evaluates a block statement in the given context.
func (i *Interpreter) EvalBlockStatement(block *ast.BlockStatement, ctx *evaluator.ExecutionContext) {
	// Sync context state to interpreter
	i.syncFromContext(ctx)
	defer i.syncToContext(ctx)

	i.evalBlockStatement(block)
}

// EvalStatement evaluates a single statement in the given context.
func (i *Interpreter) EvalStatement(stmt ast.Statement, ctx *evaluator.ExecutionContext) {
	// Sync context state to interpreter
	i.syncFromContext(ctx)
	defer i.syncToContext(ctx)

	i.Eval(stmt)
}

// syncFromContext syncs execution state from context to interpreter.
func (i *Interpreter) syncFromContext(ctx *evaluator.ExecutionContext) {
	// Sync exception state
	if exc := ctx.Exception(); exc != nil {
		if excVal, ok := exc.(*ExceptionValue); ok {
			i.exception = excVal
		}
	} else {
		i.exception = nil
	}

	// Sync handler exception
	if hexc := ctx.HandlerException(); hexc != nil {
		if excVal, ok := hexc.(*ExceptionValue); ok {
			i.handlerException = excVal
		}
	} else {
		i.handlerException = nil
	}
}

// syncToContext syncs execution state from interpreter to context.
func (i *Interpreter) syncToContext(ctx *evaluator.ExecutionContext) {
	// Sync exception state back
	ctx.SetException(i.exception)
	ctx.SetHandlerException(i.handlerException)
}

// ===== Task 3.5.6: Array and Collection Adapter Method Implementations =====

// CreateArray creates an array from a list of elements with a specified element type.
func (i *Interpreter) CreateArray(elementType any, elements []evaluator.Value) evaluator.Value {
	// Convert elementType to types.Type
	var typedElementType types.Type
	if elementType != nil {
		if t, ok := elementType.(types.Type); ok {
			typedElementType = t
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Create array type (dynamic array has nil bounds)
	arrayType := &types.ArrayType{
		ElementType: typedElementType,
		LowBound:    nil,
		HighBound:   nil,
	}

	// Create array value
	arrayVal := NewArrayValue(arrayType)

	// Add elements (append to Elements slice)
	arrayVal.Elements = append(arrayVal.Elements, internalElements...)

	return arrayVal
}

// CreateArrayValue creates an array value with a specific array type.
func (i *Interpreter) CreateArrayValue(arrayType any, elements []evaluator.Value) evaluator.Value {
	// Convert arrayType to *types.ArrayType
	var typedArrayType *types.ArrayType
	if arrayType != nil {
		if at, ok := arrayType.(*types.ArrayType); ok {
			typedArrayType = at
		}
	}

	// Convert evaluator.Value slice to internal Value slice
	internalElements := make([]Value, len(elements))
	for idx, elem := range elements {
		internalElements[idx] = elem.(Value)
	}

	// Create and return the array value
	return &ArrayValue{
		ArrayType: typedArrayType,
		Elements:  internalElements,
	}
}
