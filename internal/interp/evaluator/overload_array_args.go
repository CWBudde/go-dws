package evaluator

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// capturedArrayArgument retains the storage and index used while determining
// an overload. Binding is deferred until the selected parameter mode is known.
type capturedArrayArgument struct {
	array         *runtime.ArrayValue
	node          *ast.IndexExpression
	bindContainer func() Value
	reference     Value
	index         int
	// bindReference replaces the dense-array binding for an associative
	// element, whose key was evaluated once during capture.
	bindReference func() (Value, error)
}

func argumentEvaluationError(value Value, ctx *ExecutionContext) error {
	if isError(value) {
		return fmt.Errorf("%s", value.String())
	}
	if ctx.Exception() != nil {
		return fmt.Errorf("argument evaluation raised an exception")
	}
	return nil
}

// captureOverloadArrayArgument uses read semantics until overload selection.
// In particular, probing a missing associative slot must not insert it when
// the selected overload takes its argument by value.
func (e *Evaluator) captureOverloadArrayArgument(arg ast.Expression, ctx *ExecutionContext) (Value, *capturedArrayArgument, bool, error) {
	node, ok := arg.(*ast.IndexExpression)
	if !ok {
		return nil, nil, false, nil
	}
	var typ types.Type
	if e.engineState != nil {
		typ = e.resolvedExpressionType(node.Left, ctx)
	}
	if typ != nil {
		switch types.GetUnderlyingType(typ).(type) {
		case *types.ArrayType, *types.AssociativeArrayType:
		default:
			return nil, nil, false, nil
		}
	} else if base, _ := CollectIndices(node); isMemberRootedBase(base) {
		// An unchecked indexed property needs the normal property dispatcher.
		return nil, nil, false, nil
	}
	container, bind, err := e.captureOverloadArrayContainer(node.Left, ctx)
	if err != nil {
		return nil, nil, true, err
	}
	if assoc, ok := derefAssociative(container); ok {
		return e.captureOverloadAssociativeArgument(assoc, node, bind, ctx)
	}
	arr, ok := container.(*runtime.ArrayValue)
	if !ok || arr.ArrayType == nil {
		value := e.indexResolvedValue(container, node, ctx)
		return value, nil, true, argumentEvaluationError(value, ctx)
	}
	index, err := e.evaluateOverloadArrayIndex(node.Index, ctx)
	if err != nil {
		return nil, nil, true, err
	}
	captured := &capturedArrayArgument{array: arr, index: index, node: node, bindContainer: bind}
	physical, boundsErr := arrayElementPhysicalIndex(arr, index)
	if boundsErr != nil {
		_, err := e.bindArrayElementReference(arr, index, node, ctx)
		return nil, nil, true, err
	}
	// Checking a native array reference does not mutate it. Capture the live
	// reference now so later argument resizes are observed at access time,
	// rather than introducing a second bind-time check after those effects.
	captured.reference, err = e.bindArrayElementReference(arr, index, node, ctx)
	if err != nil {
		return nil, nil, true, err
	}
	if arr.Elements[physical] == nil {
		return e.getZeroValueForType(arr.ArrayType.ElementType, ctx), captured, true, nil
	}
	return arr.Elements[physical], captured, true, nil
}

// captureOverloadAssociativeArgument evaluates the key once and reads the slot
// without inserting it. A selected var overload vivifies that same key rather
// than evaluating the index expression again.
func (e *Evaluator) captureOverloadAssociativeArgument(assoc *runtime.AssociativeArrayValue, node *ast.IndexExpression, bindParent func() Value, ctx *ExecutionContext) (Value, *capturedArrayArgument, bool, error) {
	indexValue := e.Eval(node.Index, ctx)
	if err := argumentEvaluationError(indexValue, ctx); err != nil {
		return nil, nil, true, err
	}
	key, errValue := e.coerceAssociativeKey(assoc, indexValue, ctx)
	if errValue != nil {
		return nil, nil, true, argumentEvaluationError(errValue, ctx)
	}
	value, present := assoc.Get(key)
	if !present {
		value = e.getZeroValueForType(assoc.ElementType(), ctx)
	}
	captured := &capturedArrayArgument{node: node}
	captured.bindReference = func() (Value, error) {
		container := bindParent()
		if err := argumentEvaluationError(container, ctx); err != nil {
			return nil, err
		}
		parent, ok := derefAssociative(container)
		if !ok {
			return nil, fmt.Errorf("associative array container changed during argument evaluation")
		}
		current := e.vivifyAssociativeSlot(parent, key, ctx)
		if err := argumentEvaluationError(current, ctx); err != nil {
			return nil, err
		}
		return newAssignedReference(node.String(), current, func(value Value) error {
			parent.Set(key, cloneIfCopyable(value))
			return nil
		}), nil
	}
	return value, captured, true, argumentEvaluationError(value, ctx)
}

// captureOverloadArrayContainer captures nested array/associative index paths
// without vivification. The returned callback materializes only missing slots
// along that same captured path if a var parameter is selected later.
func (e *Evaluator) captureOverloadArrayContainer(expr ast.Expression, ctx *ExecutionContext) (Value, func() Value, error) {
	node, indexed := expr.(*ast.IndexExpression)
	if !indexed {
		value := e.Eval(expr, ctx)
		if err := argumentEvaluationError(value, ctx); err != nil {
			return nil, nil, err
		}
		return value, func() Value { return value }, nil
	}
	container, bindParent, err := e.captureOverloadArrayContainer(node.Left, ctx)
	if err != nil {
		return nil, nil, err
	}
	assoc, associative := derefAssociative(container)
	if !associative {
		if array, ok := container.(*runtime.ArrayValue); ok {
			index, err := e.evaluateOverloadArrayIndex(node.Index, ctx)
			if err != nil {
				return nil, nil, err
			}
			value := e.IndexArray(array, index, node, ctx)
			bind := func() Value {
				parent := bindParent()
				if argumentEvaluationError(parent, ctx) != nil {
					return parent
				}
				if parent == array {
					return value
				}
				if resolved, ok := parent.(*runtime.ArrayValue); ok {
					return e.IndexArray(resolved, index, node, ctx)
				}
				return e.newError(node, "array container changed during argument evaluation")
			}
			return value, bind, argumentEvaluationError(value, ctx)
		}
		value := e.indexResolvedValue(container, node, ctx)
		return value, func() Value { return value }, argumentEvaluationError(value, ctx)
	}
	indexValue := e.Eval(node.Index, ctx)
	if err := argumentEvaluationError(indexValue, ctx); err != nil {
		return nil, nil, err
	}
	key, errValue := e.coerceAssociativeKey(assoc, indexValue, ctx)
	if errValue != nil {
		return nil, nil, argumentEvaluationError(errValue, ctx)
	}
	value, present := assoc.Get(key)
	if present {
		// Existing storage is already captured. A later argument replacing the
		// associative slot must not redirect this argument to the new array.
		return value, func() Value { return value }, nil
	}
	value = e.getZeroValueForType(assoc.ElementType(), ctx)
	bind := func() Value {
		parent, ok := derefAssociative(bindParent())
		if !ok {
			return e.newError(node, "associative array container changed during argument evaluation")
		}
		if stored, exists := parent.Get(key); exists {
			return stored
		}
		parent.Set(key, value)
		stored, _ := parent.Get(key)
		return stored
	}
	return value, bind, argumentEvaluationError(value, ctx)
}

func (e *Evaluator) evaluateOverloadArrayIndex(expr ast.Expression, ctx *ExecutionContext) (int, error) {
	value := e.Eval(expr, ctx)
	if err := argumentEvaluationError(value, ctx); err != nil {
		return 0, err
	}
	index, ok := e.ExtractIndexWithVariantCast(value, ctx)
	if !ok {
		if err := argumentEvaluationError(value, ctx); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("index must be an ordinal value, got %s", value.Type())
	}
	return index, nil
}
