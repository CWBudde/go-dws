package interp

import (
	"github.com/cwbudde/go-dws/internal/interp/evaluator"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ===== Value Creation Adapter Methods =====
// Task 3.5.130d: CreateExternalVar removed - evaluator now constructs runtime.ExternalVarValue directly
// Task 3.5.139h: ResolveArrayTypeNode removed - evaluator uses resolveArrayTypeNode() directly

// Task 3.5.128f: CreateRecordZeroValue removed - evaluator now handles record zero-value creation directly

// Task 3.5.129: CreateArrayZeroValue, CreateSetZeroValue, CreateSubrangeZeroValue, CreateInterfaceZeroValue, CreateClassZeroValue removed

// ===== Task 3.5.129: Bridge Adapter Methods for Zero Value Creation =====

// Task 3.5.19: CreateSubrangeValueDirect removed - evaluator now uses runtime.NewSubrangeValueZero directly
// Task 3.5.23: CreateInterfaceInstanceDirect removed - evaluator now uses runtime.NewInterfaceInstance directly

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

// Task 3.5.136: GetExceptionInstance removed - migrated to evaluator.getExceptionInstance()
// Uses ExceptionValue.GetInstance() method to extract ObjectInstance without adapter.

// Task 3.5.134: CreateExceptionFromObject removed - migrated to evaluator.createExceptionFromObject()
// The evaluator now handles nil objects and uses WrapObjectInException bridge constructor.

// ===== Statement Evaluation Adapter Methods =====

// Task 3.5.137: EvalBlockStatement and EvalStatement removed - evaluator calls e.Eval() directly.
// Exception handling code now uses direct evaluation instead of adapter delegation.
// The syncFromContext/syncToContext helper methods were also removed as they were only used by these methods.

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

// ===== Type Conversion & Introspection Adapter Methods (Task 3.5.143g) =====
// Note: These methods are already implemented in builtins_context.go as part of
// the builtins.Context interface. They take builtins.Value (runtime.Value) and
// are used by both the Interpreter directly and via the adapter bridge.

// ===== Enum Operation Methods (Task 3.5.143q) =====
// Note: Enum operation methods (GetEnumOrdinal, GetEnumSuccessor, GetEnumPredecessor,
// GetJSONVarType) are implemented in builtins_context.go as part of the builtins.Context
// interface. The Evaluator accesses them by casting its adapter to builtins.Context.
// See context_enums.go for the Evaluator-side implementations.
