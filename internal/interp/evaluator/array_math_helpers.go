package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

func arrayMathArity(operation types.BuiltinHelperOperation) (int, bool) {
	switch operation {
	case types.HelperArrayOffset, types.HelperArrayMultiply:
		return 1, true
	case types.HelperArrayMultiplyAdd:
		return 2, true
	case types.HelperArrayReciprocal:
		return 0, true
	default:
		return 0, false
	}
}

func (e *Evaluator) checkArrayMathReceiver(self Value, argumentCount, arity int, node ast.Node) (*runtime.ArrayValue, Value) {
	arr, ok := self.(*runtime.ArrayValue)
	if !ok || arr.ArrayType == nil || !arr.ArrayType.IsDynamic() || types.GetUnderlyingType(arr.ArrayType.ElementType) != types.FLOAT {
		return nil, e.newError(node, "array math helper requires a dynamic array of Float")
	}
	if argumentCount != arity {
		return nil, e.newError(node, "array math helper expects %d arguments, got %d", arity, argumentCount)
	}
	return arr, nil
}

// evalArrayMathCall resolves the actual helper before deciding whether arguments
// are needed. DWScript skips scalar operands for an empty Float array; a user
// helper with the same name must retain ordinary argument evaluation.
func (e *Evaluator) evalArrayMathCall(self Value, node *ast.MethodCallExpression, ctx *ExecutionContext) (Value, bool) {
	if _, ok := self.(*runtime.ArrayValue); !ok {
		return nil, false
	}
	helper := e.FindHelperMethod(self, node.Method.Value)
	if helper == nil {
		return nil, false
	}
	operation := types.BuiltinHelperOperation(helper.BuiltinSpec)
	arity, handled := arrayMathArity(operation)
	if !handled {
		return nil, false
	}
	if ctx.Exception() != nil {
		return &runtime.NilValue{}, true
	}
	arr, err := e.checkArrayMathReceiver(self, len(node.Arguments), arity, node)
	if err != nil {
		return err, true
	}
	if len(arr.Elements) == 0 {
		return arr, true
	}
	args := make([]Value, len(node.Arguments))
	for i, expression := range node.Arguments {
		args[i] = e.Eval(expression, ctx)
		if isError(args[i]) {
			return args[i], true
		}
		if ctx.Exception() != nil {
			return &runtime.NilValue{}, true
		}
	}
	return e.evalArrayMathHelper(operation, arr, args, node), true
}

func (e *Evaluator) evalArrayMathHelper(operation types.BuiltinHelperOperation, self Value, args []Value, node ast.Node) Value {
	arity, _ := arrayMathArity(operation)
	arr, err := e.checkArrayMathReceiver(self, len(args), arity, node)
	if err != nil {
		return err
	}
	operands := [2]float64{}
	for i, arg := range args {
		switch value := arg.(type) {
		case *runtime.FloatValue:
			operands[i] = value.Value
		case *runtime.IntegerValue:
			operands[i] = float64(value.Value)
		default:
			return e.newError(node, "array math helper operand %d must be numeric", i+1)
		}
	}
	// Validate before mutation, including callers without semantic analysis.
	for _, element := range arr.Elements {
		if _, ok := element.(*runtime.FloatValue); !ok {
			return e.newError(node, "array math helper requires Float elements")
		}
	}
	for i, element := range arr.Elements {
		floatValue, ok := element.(*runtime.FloatValue)
		if !ok {
			return e.newError(node, "array math helper requires Float elements")
		}
		value := floatValue.Value
		switch operation {
		case types.HelperArrayOffset:
			value += operands[0]
		case types.HelperArrayMultiply:
			value *= operands[0]
		case types.HelperArrayMultiplyAdd:
			// Explicit rounding prevents fusion into a single multiply-add.
			product := float64(value * operands[0])
			value = product + operands[1]
		case types.HelperArrayReciprocal:
			value = 1 / value
		}
		// Scalars can be shared with other variables; replace their wrappers.
		arr.Elements[i] = &runtime.FloatValue{Value: value}
	}
	return arr
}
