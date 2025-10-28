package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Function Pointer Type Declaration Analysis (Task 9.159)
// ============================================================================

// analyzeFunctionPointerTypeDeclaration analyzes a function pointer type declaration.
// It validates the function pointer signature and registers it in the type environment.
//
// Examples:
//   - type TComparator = function(a, b: Integer): Integer;
//   - type TCallback = procedure(msg: String);
//   - type TNotifyEvent = procedure(Sender: TObject) of object;
//
// Task 9.159: Validation checks:
//   - No duplicate parameter names
//   - All parameter types exist and are valid
//   - Return type exists and is valid (for functions)
func (a *Analyzer) analyzeFunctionPointerTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil || decl.FunctionPointerType == nil {
		return
	}

	fpType := decl.FunctionPointerType

	// Task 9.159: Check for duplicate parameter names
	paramNames := make(map[string]bool)
	for _, param := range fpType.Parameters {
		if paramNames[param.Name.Value] {
			a.addError("duplicate parameter name '%s' in function pointer type at %s",
				param.Name.Value, param.Name.Token.Pos.String())
			return
		}
		paramNames[param.Name.Value] = true
	}

	// Task 9.159: Validate all parameter types exist
	paramTypes := make([]types.Type, 0, len(fpType.Parameters))
	for _, param := range fpType.Parameters {
		paramType, err := a.resolveType(param.Type.Name)
		if err != nil {
			a.addError("unknown parameter type '%s' in function pointer type at %s",
				param.Type.Name, param.Type.Token.Pos.String())
			return
		}
		paramTypes = append(paramTypes, paramType)
	}

	// Task 9.159: Validate return type (for functions)
	var returnType types.Type
	if fpType.ReturnType != nil {
		var err error
		returnType, err = a.resolveType(fpType.ReturnType.Name)
		if err != nil {
			a.addError("unknown return type '%s' in function pointer type at %s",
				fpType.ReturnType.Name, fpType.ReturnType.Token.Pos.String())
			return
		}
	}

	// Task 9.159: Create and register the function pointer type
	var funcPtrType types.Type
	if fpType.OfObject {
		// Method pointer type (procedure/function of object)
		funcPtrType = types.NewMethodPointerType(paramTypes, returnType)
	} else {
		// Regular function pointer type
		funcPtrType = types.NewFunctionPointerType(paramTypes, returnType)
	}

	// Register in the function pointers map
	if a.functionPointers == nil {
		a.functionPointers = make(map[string]*types.FunctionPointerType)
	}

	// Store the underlying function pointer type (even for method pointers)
	if methodPtr, ok := funcPtrType.(*types.MethodPointerType); ok {
		a.functionPointers[decl.Name.Value] = &methodPtr.FunctionPointerType
	} else if funcPtr, ok := funcPtrType.(*types.FunctionPointerType); ok {
		a.functionPointers[decl.Name.Value] = funcPtr
	}

	// Also register as a type alias so resolveType can find it
	typeAlias := &types.TypeAlias{
		Name:        decl.Name.Value,
		AliasedType: funcPtrType,
	}
	a.typeAliases[decl.Name.Value] = typeAlias
}

// ============================================================================
// Address-of Expression Analysis (Task 9.160)
// ============================================================================

// analyzeAddressOfExpression analyzes an address-of expression (@FunctionName).
// It resolves the target function/procedure and creates a function pointer value.
//
// Examples:
//   - @Ascending (function pointer)
//   - @MyCallback (procedure pointer)
//   - @TMyClass.MyMethod (method pointer)
//
// Task 9.160: Analysis steps:
//   - Resolve the target function/procedure in the symbol table
//   - Verify it's actually a function or procedure declaration
//   - Extract parameter types and return type from the function
//   - Create a FunctionPointerType or MethodPointerType
//   - Set the expression type for use in assignments/calls
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
func (a *Analyzer) analyzeAddressOfFunction(funcName string, expr *ast.AddressOfExpression) types.Type {
	// Look up the function in the symbol table
	sym, ok := a.symbols.Resolve(funcName)
	if !ok {
		a.addError("undefined function '%s' in address-of expression at %s",
			funcName, expr.Token.Pos.String())
		return nil
	}

	// The symbol must be a function type
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("'%s' is not a function or procedure (got %s) at %s",
			funcName, sym.Type.String(), expr.Token.Pos.String())
		return nil
	}

	// Create a function pointer type from the function's signature
	// Task 9.160: Extract parameter types and return type
	// For procedures, the return type should be nil (not VOID)
	var returnType types.Type
	if funcType.ReturnType != nil && funcType.ReturnType != types.VOID {
		returnType = funcType.ReturnType
	}
	funcPtrType := types.NewFunctionPointerType(funcType.Parameters, returnType)

	// Task 9.160: Set the type on the expression node
	// Note: We need to create a TypeAnnotation for the AST node
	typeAnnotation := &ast.TypeAnnotation{
		Name: fmt.Sprintf("function pointer to %s", funcName),
	}
	expr.SetType(typeAnnotation)

	return funcPtrType
}

// ============================================================================
// Function Pointer Assignment Validation (Task 9.161)
// ============================================================================

// NOTE: Function pointer assignment validation is handled by the IsCompatibleWith
// method in the types package, which is called from analyzer.canAssign().
// This provides automatic compatibility checking for all assignment types.

// ============================================================================
// Function Pointer Call Validation (Task 9.162)
// ============================================================================

// analyzeFunctionPointerCall analyzes a call to a function pointer variable.
// It validates argument types and infers the return type from the function pointer.
//
// Task 9.162: Analysis steps:
//   - Detect when callee is a variable with function pointer type
//   - Validate argument types against pointer's parameter types
//   - Infer return type from function pointer type (not function declaration)
//   - Handle both function and method pointers
func (a *Analyzer) analyzeFunctionPointerCall(callExpr *ast.CallExpression, calleeType types.Type) types.Type {
	// Resolve type aliases to get the underlying type
	underlyingType := types.GetUnderlyingType(calleeType)

	// Extract function pointer type (could be FunctionPointerType or MethodPointerType)
	var funcPtr *types.FunctionPointerType
	if fp, ok := underlyingType.(*types.FunctionPointerType); ok {
		funcPtr = fp
	} else if mp, ok := underlyingType.(*types.MethodPointerType); ok {
		funcPtr = &mp.FunctionPointerType
	} else {
		// Not a function pointer call
		return nil
	}

	// Task 9.162: Validate argument count
	if len(callExpr.Arguments) != len(funcPtr.Parameters) {
		a.addError("function pointer call argument count mismatch at %s: expected %d arguments, got %d",
			callExpr.Token.Pos.String(), len(funcPtr.Parameters), len(callExpr.Arguments))
		return nil
	}

	// Task 9.162: Validate each argument type
	for i, arg := range callExpr.Arguments {
		argType := a.analyzeExpression(arg)
		if argType == nil {
			// Error already reported
			continue
		}

		expectedType := funcPtr.Parameters[i]
		if !a.canAssign(argType, expectedType) {
			a.addError("function pointer call argument %d type mismatch at %s: expected %s, got %s",
				i+1, callExpr.Token.Pos.String(), expectedType.String(), argType.String())
		}
	}

	// Task 9.162: Return the function pointer's return type
	if funcPtr.ReturnType != nil {
		return funcPtr.ReturnType
	}
	return types.VOID
}
