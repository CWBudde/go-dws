package evaluator

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Value to Type Conversion
// ============================================================================

// GetValueType converts a runtime Value to its corresponding types.Type.
// Returns nil for values that don't have a corresponding semantic type (e.g., nil, unassigned).
//
// This is used for array literal type inference where we need to determine
// the element type from the runtime values.
//
// Mapping:
//   - IntegerValue → types.INTEGER
//   - FloatValue → types.FLOAT
//   - StringValue → types.STRING
//   - BooleanValue → types.BOOLEAN
//   - NilValue → nil (context-dependent)
//   - ArrayValue → types.ArrayType (with element type)
//   - RecordValue → types.RecordType
//   - ObjectInstance → types.Class (the object's class)
//   - VariantValue → unwrap to underlying type
//   - EnumValue → types.EnumType
//   - SetValue → types.SetType
//   - NullValue/UnassignedValue → nil
func GetValueType(val Value) types.Type {
	switch runtime.KindOf(val) {
	case runtime.KindNil, runtime.KindNull, runtime.KindUnassigned:
		return nil
	case runtime.KindVariant:
		return unwrapVariantType(val)
	default:
		return runtime.LanguageType(val)
	}
}

// unwrapVariantType unwraps a Variant value to get its actual type.
// This uses a type-safe interface check for VariantWrapper.
func unwrapVariantType(val Value) types.Type {
	// Check if the value implements the VariantWrapper interface
	// This interface is defined in runtime package to avoid circular imports
	type VariantWrapper interface {
		UnwrapVariant() Value
	}

	if wrapper, ok := val.(VariantWrapper); ok {
		// Unwrap and recursively get the type
		unwrapped := wrapper.UnwrapVariant()
		return GetValueType(unwrapped)
	}

	// Variant with unknown content - fallback to Variant type
	return types.VARIANT
}

// ============================================================================
// Function Pointer Creation Helpers
// ============================================================================

// createFunctionPointerFromDecl creates a FunctionPointerValue from a function declaration.
// This is a simple wrapper that creates a function pointer without type information.
func createFunctionPointerFromDecl(fn *ast.FunctionDecl, closure any) Value {
	return &runtime.FunctionPointerValue{
		Function: fn,
		Closure:  closure,
	}
}

// buildFunctionPointerType resolves a function's declared signature without parsing display text.
func (e *Evaluator) buildFunctionPointerType(fn *ast.FunctionDecl, ctx *ExecutionContext) *types.FunctionPointerType {
	paramTypes := make([]types.Type, len(fn.Parameters))
	for i, parameter := range fn.Parameters {
		resolved, err := e.ResolveTypeFromAnnotation(parameter.Type, ctx)
		if err != nil || resolved == nil {
			return nil
		}
		paramTypes[i] = resolved
	}
	var returnType types.Type
	if fn.ReturnType != nil {
		resolved, err := e.ResolveTypeFromAnnotation(fn.ReturnType, ctx)
		if err != nil || resolved == nil {
			return nil
		}
		returnType = resolved
	}
	return types.NewFunctionPointerType(paramTypes, returnType)
}

// createFunctionPointerFromDecl creates a FunctionPointerValue from a method declaration.
// The methodDecl parameter is any to support both ClassMetaValue and ObjectValue callback signatures.
func (e *Evaluator) createFunctionPointerFromDecl(methodDecl any, selfObject Value, ctx *ExecutionContext) Value {
	var fn *ast.FunctionDecl
	var callable *runtime.MethodMetadata
	switch method := methodDecl.(type) {
	case *runtime.MethodMetadata:
		callable = method
		fn = runtime.MethodDeclaration(method)
	case *ast.FunctionDecl:
		fn = method
	}
	if fn == nil {
		return e.newError(nil, "internal error: expected callable declaration, got %T", methodDecl)
	}

	pointerType := e.buildFunctionPointerType(fn, ctx)
	if callable != nil {
		fn = nil
	}
	return &runtime.FunctionPointerValue{
		Callable:    callable,
		Function:    fn,
		Closure:     ctx.Env(),
		SelfObject:  selfObject,
		PointerType: pointerType,
	}
}
