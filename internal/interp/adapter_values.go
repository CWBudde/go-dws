package interp

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Value Creation Adapter Methods =====
// Task 3.5.130d: CreateExternalVar removed - evaluator now constructs runtime.ExternalVarValue directly

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

// Task 3.5.128f: CreateRecordZeroValue removed - evaluator now handles record zero-value creation directly

// Task 3.5.129: CreateArrayZeroValue, CreateSetZeroValue, CreateSubrangeZeroValue, CreateInterfaceZeroValue, CreateClassZeroValue removed

// ===== Task 3.5.129: Bridge Adapter Methods for Zero Value Creation =====

// CreateSubrangeValueDirect creates a subrange value from subrange type metadata.
// Task 3.5.129c: Bridge constructor - SubrangeValue cannot be constructed in evaluator (circular import).
func (i *Interpreter) CreateSubrangeValueDirect(subrangeTypeAny any) evaluator.Value {
	// Type assert to extract SubrangeType
	type subrangeTypeProvider interface {
		GetSubrangeType() *types.SubrangeType
	}

	stProvider, ok := subrangeTypeAny.(subrangeTypeProvider)
	if !ok {
		return &NilValue{}
	}

	st := stProvider.GetSubrangeType()
	if st == nil {
		return &NilValue{}
	}

	// Construct SubrangeValue with low bound as zero value
	return &SubrangeValue{
		Value:        st.LowBound,
		SubrangeType: st,
	}
}

// CreateInterfaceInstanceDirect creates a nil interface instance from metadata.
// Task 3.5.129d: Bridge constructor - InterfaceInstance cannot be constructed in evaluator.
func (i *Interpreter) CreateInterfaceInstanceDirect(interfaceInfoAny any) evaluator.Value {
	ifaceInfo, ok := interfaceInfoAny.(*InterfaceInfo)
	if !ok {
		return &NilValue{}
	}

	// Create nil interface instance (Object field is nil)
	return &InterfaceInstance{
		Interface: ifaceInfo,
		Object:    nil,
	}
}

// CreateTypedNilValue creates a typed nil value for a class.
// Task 3.5.129e: Bridge constructor - NilValue.ClassType cannot be set in evaluator.
func (i *Interpreter) CreateTypedNilValue(className string) evaluator.Value {
	return &NilValue{ClassType: className}
}

// ===== Task 3.5.40: Record Literal Adapter Method Implementations =====
// Task 3.5.128e: CreateRecordValue removed - evaluator now handles record creation directly

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

// Task 3.5.135: MatchesExceptionType removed - migrated to evaluator.matchesExceptionType()
// Uses TypeSystem.IsClassDescendantOf for class hierarchy checking.

// GetExceptionInstance returns the ObjectInstance from an exception.
func (i *Interpreter) GetExceptionInstance(exc interface{}) evaluator.Value {
	excVal, ok := exc.(*ExceptionValue)
	if !ok {
		return nil
	}
	return excVal.Instance
}

// Task 3.5.134: CreateExceptionFromObject removed - migrated to evaluator.createExceptionFromObject()
// The evaluator now handles nil objects and uses WrapObjectInException bridge constructor.

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
