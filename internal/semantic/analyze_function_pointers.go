package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Function Pointer Type Declaration Analysis
// ============================================================================

// analyzeFunctionPointerTypeDeclaration analyzes a function pointer type declaration.
// Examples: type TComparator = function(a, b: Integer): Integer;
func (a *Analyzer) analyzeFunctionPointerTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.FunctionPointerType == nil {
		return
	}

	fpType := decl.FunctionPointerType

	// Check for duplicate parameter names (skip shorthand syntax with nil names)
	paramNames := make(map[string]bool)
	for _, param := range fpType.Parameters {
		if param.Name == nil {
			continue
		}
		if paramNames[param.Name.Value] {
			a.addError("duplicate parameter name '%s' in function pointer type at %s",
				param.Name.Value, param.Name.Token.Pos.String())
			return
		}
		paramNames[param.Name.Value] = true
	}

	// Validate all parameter types exist
	paramTypes := make([]types.Type, 0, len(fpType.Parameters))
	for _, param := range fpType.Parameters {
		paramType, err := a.resolveType(getTypeExpressionName(param.Type))
		if err != nil {
			a.addError("unknown parameter type '%s' in function pointer type at %s",
				getTypeExpressionName(param.Type), param.Type.Pos().String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Validate return type (for functions)
	var returnType types.Type
	if fpType.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(getTypeExpressionName(fpType.ReturnType))
		if err != nil {
			a.addError("unknown return type '%s' in function pointer type at %s",
				getTypeExpressionName(fpType.ReturnType), fpType.ReturnType.Pos().String())
			return
		}
	}

	// Create the function pointer type
	var funcPtrType types.Type
	if fpType.OfObject {
		funcPtrType = types.NewMethodPointerType(paramTypes, returnType)
	} else {
		funcPtrType = types.NewFunctionPointerType(paramTypes, returnType)
	}

	// Register in the function pointers map
	if a.functionPointers == nil {
		a.functionPointers = make(map[string]*types.FunctionPointerType)
	}
	if methodPtr, ok := funcPtrType.(*types.MethodPointerType); ok {
		a.functionPointers[decl.Name.Value] = &methodPtr.FunctionPointerType
	} else if funcPtr, ok := funcPtrType.(*types.FunctionPointerType); ok {
		a.functionPointers[decl.Name.Value] = funcPtr
	}

	// Register as a type alias so resolveType can find it
	typeAlias := &types.TypeAlias{
		Name:        decl.Name.Value,
		AliasedType: funcPtrType,
	}
	a.registerTypeWithPos(decl.Name.Value, typeAlias, decl.Token.Pos)
}

// ============================================================================
// Address-of Expression Analysis
// ============================================================================

// analyzeAddressOfExpression analyzes an address-of expression (@FunctionName).
// Examples: @Ascending, @MyCallback, @TMyClass.MyMethod
func (a *Analyzer) analyzeAddressOfExpression(expr *ast.AddressOfExpression) types.Type {
	if expr == nil || expr.Operator == nil {
		return nil
	}

	// The operator should be an identifier or member access expression
	switch target := expr.Operator.(type) {
	case *ast.Identifier:
		// Simple function/procedure reference: @FunctionName
		return a.analyzeAddressOfFunction(target.Value, expr)

	case *ast.MemberAccessExpression:
		// Method reference: @TMyClass.MyMethod
		// This would require analyzing the class and method
		a.addError("method pointers (@TClass.Method) not yet implemented at %s",
			expr.Token.Pos.String())
		return nil

	default:
		a.addError("address-of operator (@) requires a function or procedure name at %s",
			expr.Token.Pos.String())
		return nil
	}
}

// analyzeAddressOfFunction resolves a function name and creates a function pointer type.
// Queries both symbol table and builtin registry.
func (a *Analyzer) analyzeAddressOfFunction(funcName string, expr *ast.AddressOfExpression) types.Type {
	sym, ok := a.symbols.Resolve(funcName)
	if !ok {
		// Query builtin registry for function signatures
		if sig, found := builtins.DefaultRegistry.GetSignature(funcName); found {
			if sig.IsVariadic {
				a.addError("cannot take address of variadic built-in function '%s' at %s",
					funcName, expr.Token.Pos.String())
				return nil
			}
			// Builtins with optional parameters allowed; validated at call time
			return a.buildFunctionPointerTypeFromBuiltin(funcName, sig, expr)
		}

		// Builtin without signature metadata - return generic function pointer type
		if builtins.DefaultRegistry.Has(funcName) {
			funcPtrType := types.NewFunctionPointerType(nil, types.VARIANT)
			typeAnnotation := &ast.TypeAnnotation{
				Name: fmt.Sprintf("function pointer to %s", funcName),
			}
			a.semanticInfo.SetType(expr, typeAnnotation)
			return funcPtrType
		}

		a.addError("%s", errors.FormatUnknownName(funcName, expr.Token.Pos.Line, expr.Token.Pos.Column))
		return nil
	}

	// The symbol must be a function type
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("'%s' is not a function or procedure (got %s) at %s",
			funcName, sym.Type.String(), expr.Token.Pos.String())
		return nil
	}

	return a.buildFunctionPointerType(funcName, funcType, expr)
}

// buildFunctionPointerTypeFromBuiltin creates a FunctionPointerType from a builtin signature.
func (a *Analyzer) buildFunctionPointerTypeFromBuiltin(funcName string, sig *builtins.FunctionSignature, expr *ast.AddressOfExpression) types.Type {
	var returnType types.Type
	if sig.ReturnType != nil && sig.ReturnType != types.VOID {
		returnType = sig.ReturnType
	}

	funcPtrType := types.NewFunctionPointerType(sig.ParamTypes, returnType)
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("function pointer to %s", funcName),
	}
	a.semanticInfo.SetType(expr, typeAnnotation)

	return funcPtrType
}

// buildFunctionPointerType creates a FunctionPointerType from a function signature.
func (a *Analyzer) buildFunctionPointerType(funcName string, funcType *types.FunctionType, expr *ast.AddressOfExpression) types.Type {
	var returnType types.Type
	if funcType.ReturnType != nil && funcType.ReturnType != types.VOID {
		returnType = funcType.ReturnType
	}

	funcPtrType := types.NewFunctionPointerType(funcType.Parameters, returnType)
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("function pointer to %s", funcName),
	}
	a.semanticInfo.SetType(expr, typeAnnotation)

	return funcPtrType
}

// ============================================================================
// Function Pointer Assignment Validation
// ============================================================================

// NOTE: Function pointer assignment validation is handled by the IsCompatibleWith
// method in the types package, which is called from analyzer.canAssign().
// This provides automatic compatibility checking for all assignment types.

// ============================================================================
// Function Pointer Call Validation
// ============================================================================

// analyzeFunctionPointerCall analyzes a call to a function pointer variable.
// Validates argument types and infers the return type from the function pointer.
func (a *Analyzer) analyzeFunctionPointerCall(callExpr *ast.CallExpression, calleeType types.Type) types.Type {
	underlyingType := types.GetUnderlyingType(calleeType)

	// Extract function pointer type (could be FunctionPointerType or MethodPointerType)
	var funcPtr *types.FunctionPointerType
	if fp, ok := underlyingType.(*types.FunctionPointerType); ok {
		funcPtr = fp
	} else if mp, ok := underlyingType.(*types.MethodPointerType); ok {
		funcPtr = &mp.FunctionPointerType
	} else {
		return nil
	}

	// Validate argument count
	if len(callExpr.Arguments) != len(funcPtr.Parameters) {
		a.addError("function pointer call argument count mismatch at %s: expected %d arguments, got %d",
			callExpr.Token.Pos.String(), len(funcPtr.Parameters), len(callExpr.Arguments))
		return nil
	}

	// Validate each argument type
	for i, arg := range callExpr.Arguments {
		argType := a.analyzeExpression(arg)
		if argType == nil {
			continue
		}
		expectedType := funcPtr.Parameters[i]
		if !a.canAssign(argType, expectedType) {
			a.addError("function pointer call argument %d type mismatch at %s: expected %s, got %s",
				i+1, callExpr.Token.Pos.String(), expectedType.String(), argType.String())
		}
	}

	if funcPtr.ReturnType != nil {
		return funcPtr.ReturnType
	}
	return types.VOID
}
