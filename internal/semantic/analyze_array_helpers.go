package semantic

import (
	dwserrors "github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

func arrayHelperCanonicalName(methodName string) string {
	switch ident.Normalize(methodName) {
	case "low":
		return "Low"
	case "high":
		return "High"
	case "length":
		return "Length"
	case "count":
		return "Count"
	case "setlength":
		return "SetLength"
	case "indexof":
		return "IndexOf"
	case "delete":
		return "Delete"
	case "remove":
		return "Remove"
	case "insert":
		return "Insert"
	case "move":
		return "Move"
	case "swap":
		return "Swap"
	case "copy":
		return "Copy"
	case "foreach":
		return "ForEach"
	case "sort":
		return "Sort"
	}
	return methodName
}

func semanticTypeNameForDiagnostic(t types.Type) string {
	if t == nil {
		return "nil"
	}
	if t == types.ARRAY_OF_CONST {
		return "array of const"
	}
	if t.Equals(types.NIL) {
		return "nil"
	}
	if t.Equals(types.VOID) {
		return "void"
	}
	if fn, ok := types.GetUnderlyingType(t).(*types.FunctionPointerType); ok {
		return semanticFunctionPointerName(fn)
	}
	return semanticDiagnosticTypeName(dwserrors.SimplifyTypeName(t.String()))
}

func semanticDeclaredTypeName(typeExpr ast.TypeExpression, resolved types.Type) string {
	if resolved == types.ARRAY_OF_CONST {
		return "array of const"
	}
	if typeExpr != nil {
		if name := getTypeExpressionName(typeExpr); name != "" {
			return name
		}
	}
	return semanticTypeNameForDiagnostic(resolved)
}

func semanticFunctionParamTypeName(fn *types.FunctionType, index int, fallback types.Type) string {
	if fn != nil && index >= 0 && index < len(fn.ParamTypeNames) {
		if name := fn.ParamTypeNames[index]; name != "" {
			return name
		}
	}
	return semanticTypeNameForDiagnostic(fallback)
}

func isArrayOfConstType(t types.Type) bool {
	return types.GetUnderlyingType(t) == types.ARRAY_OF_CONST
}

func semanticDiagnosticTypeName(typeName string) string {
	if typeName == "Void" {
		return "void"
	}
	if len(typeName) >= len("array[") && typeName[:len("array[")] == "array[" {
		return "array " + typeName[len("array"):]
	}
	return typeName
}

func semanticFunctionPointerName(fn *types.FunctionPointerType) string {
	if fn == nil {
		return "nil"
	}
	kind := "function"
	if fn.IsProcedure() {
		kind = "procedure"
	}
	name := kind + " ("
	for i, param := range fn.Parameters {
		if i > 0 {
			name += ", "
		}
		name += semanticTypeNameForDiagnostic(param)
	}
	name += ")"
	if fn.IsFunction() && fn.ReturnType != nil {
		name += ": " + semanticTypeNameForDiagnostic(fn.ReturnType)
	}
	return name
}

func semanticNamedFunctionPointerName(name string, fn *types.FunctionPointerType) string {
	if fn == nil {
		return name
	}
	kind := "function"
	if fn.IsProcedure() {
		kind = "procedure"
	}
	result := kind + " " + name + "("
	for i, param := range fn.Parameters {
		if i > 0 {
			result += ", "
		}
		result += semanticTypeNameForDiagnostic(param)
	}
	result += ")"
	if fn.IsFunction() && fn.ReturnType != nil {
		result += ": " + semanticTypeNameForDiagnostic(fn.ReturnType)
	}
	return result
}

func semanticNamedFunctionSignature(name string, fn *types.FunctionType) string {
	if fn == nil {
		return name
	}
	kind := "function"
	if fn.IsProcedure() {
		kind = "procedure"
	}
	result := kind + " " + name + "("
	for i, param := range fn.Parameters {
		if i > 0 {
			result += ", "
		}
		if i < len(fn.ConstParams) && fn.ConstParams[i] {
			result += "const "
		} else if i < len(fn.VarParams) && fn.VarParams[i] {
			result += "var "
		}
		result += semanticTypeNameForDiagnostic(param)
	}
	result += ")"
	if fn.IsFunction() && fn.ReturnType != nil {
		result += ": " + semanticTypeNameForDiagnostic(fn.ReturnType)
	}
	return result
}

func previousColumn(pos token.Position) token.Position {
	if pos.Column > 1 {
		pos.Column--
	}
	if pos.Offset > 0 {
		pos.Offset--
	}
	return pos
}

func arrayHelperCallDiagnosticPos(node ast.Node) token.Position {
	if node == nil {
		return token.Position{}
	}
	switch node.(type) {
	case *ast.MemberAccessExpression:
		return node.End()
	}
	return previousColumn(node.End())
}

func (a *Analyzer) addArrayHelperCaseHint(method *ast.Identifier) {
	if method == nil {
		return
	}
	canonical := arrayHelperCanonicalName(method.Value)
	if canonical != method.Value && ident.Equal(canonical, method.Value) {
		a.addCaseMismatchHint(method.Value, canonical, method.Token.Pos)
	}
}

func (a *Analyzer) addArrayHelperError(pos token.Position, message string) {
	a.addError("%s at %s", message, pos.String())
}

func (a *Analyzer) addArrayHelperTooFewArgs(expr ast.Node) {
	a.addArrayHelperError(arrayHelperCallDiagnosticPos(expr), "More arguments expected")
}

func (a *Analyzer) addArrayHelperTooManyArgs(expr ast.Node) {
	a.addArrayHelperError(arrayHelperCallDiagnosticPos(expr), "Too many arguments")
}

func (a *Analyzer) addArrayHelperNoArgs(expr ast.Node) {
	a.addArrayHelperError(arrayHelperCallDiagnosticPos(expr), "No arguments expected")
}

func (a *Analyzer) addArrayHelperIntegerExpectedAt(pos token.Position) {
	a.addArrayHelperError(pos, "Integer expression expected")
}

func (a *Analyzer) addArrayHelperParamTypeExpectedText(pos token.Position, expected string, got string) {
	a.addArrayHelperError(pos,
		`Incompatible parameter types - "`+expected+`" expected (instead of "`+got+`")`)
}

func (a *Analyzer) addArrayHelperParamTypeExpectedAt(pos token.Position, expected types.Type, got types.Type) {
	a.addArrayHelperError(pos,
		`Incompatible parameter types - "`+semanticTypeNameForDiagnostic(expected)+`" expected (instead of "`+semanticTypeNameForDiagnostic(got)+`")`)
}

func functionPointerFromFunctionType(fn *types.FunctionType) *types.FunctionPointerType {
	if fn == nil {
		return nil
	}
	var returnType types.Type
	if fn.ReturnType != nil && !fn.ReturnType.Equals(types.VOID) {
		returnType = fn.ReturnType
	}
	return types.NewFunctionPointerType(fn.Parameters, returnType)
}

func (a *Analyzer) resolveNamedFunctionPointerType(name string) *types.FunctionPointerType {
	if name == "" {
		return nil
	}
	if sym, ok := a.symbols.Resolve(name); ok {
		if fn, ok := sym.Type.(*types.FunctionType); ok {
			return functionPointerFromFunctionType(fn)
		}
	}
	if fp := a.getBuiltinFunctionPointerType(name); fp != nil {
		return fp
	}
	if a.builtinRegistry != nil {
		if sig, ok := a.builtinRegistry.GetSignature(name); ok {
			var returnType types.Type
			if sig.ReturnType != nil && !sig.ReturnType.Equals(types.VOID) {
				returnType = sig.ReturnType
			}
			return types.NewFunctionPointerType(sig.ParamTypes, returnType)
		}
	}
	return nil
}

func (a *Analyzer) namedArrayHelperCallable(expr ast.Expression) (string, *types.FunctionPointerType, token.Position, bool) {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Value, a.resolveNamedFunctionPointerType(e.Value), e.Token.Pos, false
	case *ast.AddressOfExpression:
		if identExpr, ok := e.Operator.(*ast.Identifier); ok {
			return identExpr.Value, a.resolveNamedFunctionPointerType(identExpr.Value), identExpr.Token.Pos, true
		}
	}
	return "", nil, token.Position{}, false
}

func callbackResultTypeName(fn *types.FunctionPointerType) string {
	if fn == nil || fn.ReturnType == nil {
		return "void"
	}
	return semanticTypeNameForDiagnostic(fn.ReturnType)
}

func (a *Analyzer) validateArrayIntegerArgAt(arg ast.Expression, pos token.Position) types.Type {
	argType := a.analyzeExpressionWithExpectedType(arg, types.INTEGER)
	if argType != nil && !a.canAssign(argType, types.INTEGER) {
		a.addArrayHelperIntegerExpectedAt(pos)
	}
	return argType
}

func isArrayNaturallySortable(elementType types.Type) bool {
	elementType = types.GetUnderlyingType(elementType)
	return types.IsOrderedType(elementType)
}

func (a *Analyzer) validateArrayIntegerArg(arg ast.Expression) types.Type {
	return a.validateArrayIntegerArgAt(arg, arg.Pos())
}

func (a *Analyzer) analyzeArrayMemberAccess(expr *ast.MemberAccessExpression, arrayType *types.ArrayType) types.Type {
	if expr == nil || expr.Member == nil {
		return nil
	}
	memberNameLower := ident.Normalize(expr.Member.Value)
	a.addArrayHelperCaseHint(expr.Member)

	switch memberNameLower {
	case "length", "count", "high", "low":
		return types.INTEGER
	case "delete", "remove", "insert", "move", "swap", "copy", "foreach":
		a.addArrayHelperTooFewArgs(expr)
		if memberNameLower == "remove" {
			return types.INTEGER
		}
		if memberNameLower == "copy" {
			return arrayType
		}
		return types.VOID
	case "sort":
		if !isArrayNaturallySortable(arrayType.ElementType) {
			a.addArrayHelperError(expr.Member.Token.Pos, "Array does not have a natural sort order")
		}
		return arrayType
	}

	return nil
}

func (a *Analyzer) analyzeArrayMethodCall(expr *ast.MethodCallExpression, arrayType *types.ArrayType) types.Type {
	if expr == nil || expr.Method == nil {
		return nil
	}
	methodNameLower := ident.Normalize(expr.Method.Value)
	a.addArrayHelperCaseHint(expr.Method)

	switch methodNameLower {
	case "length", "count", "high", "low":
		if len(expr.Arguments) != 0 {
			a.addArrayHelperNoArgs(expr)
		}
		return types.INTEGER
	case "add", "push":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 1 {
			a.addArrayHelperTooManyArgs(expr)
		}
		elementType := arrayType.ElementType
		arg := expr.Arguments[0]
		argType := a.analyzeExpression(arg)
		if argType == nil {
			return types.VOID
		}
		if argArrayType, isArray := types.GetUnderlyingType(argType).(*types.ArrayType); isArray {
			if !a.canAssign(argType, arrayType) {
				a.addArrayHelperParamTypeExpectedAt(arg.Pos(), elementType, argType)
			} else if argArrayType.ElementType != nil && elementType != nil && !a.canAssign(argArrayType.ElementType, elementType) {
				a.addArrayHelperParamTypeExpectedAt(arg.Pos(), elementType, argType)
			}
			return types.VOID
		}
		if elementType != nil && !a.canAssign(argType, elementType) {
			a.addArrayHelperParamTypeExpectedAt(arg.Pos(), elementType, argType)
		}
		return types.VOID
	case "setlength":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 1 {
			a.addArrayHelperTooManyArgs(expr)
			a.addArrayHelperError(expr.End(), "Expression expected")
		}
		a.validateArrayIntegerArg(expr.Arguments[0])
		return types.VOID
	case "delete":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		a.validateArrayIntegerArg(expr.Arguments[0])
		if len(expr.Arguments) > 1 {
			a.validateArrayIntegerArg(expr.Arguments[1])
		}
		return types.VOID
	case "remove":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.INTEGER
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		if argType := a.analyzeExpressionWithExpectedType(expr.Arguments[0], arrayType.ElementType); argType != nil && !a.canAssign(argType, arrayType.ElementType) {
			a.addArrayHelperParamTypeExpectedAt(expr.Arguments[0].Pos(), arrayType.ElementType, argType)
		}
		if len(expr.Arguments) > 1 {
			a.validateArrayIntegerArgAt(expr.Arguments[1], expr.Arguments[0].Pos())
		}
		return types.INTEGER
	case "indexof":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.INTEGER
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		if argType := a.analyzeExpressionWithExpectedType(expr.Arguments[0], arrayType.ElementType); argType != nil && !a.canAssign(argType, arrayType.ElementType) {
			a.addArrayHelperParamTypeExpectedAt(expr.Arguments[0].Pos(), arrayType.ElementType, argType)
		}
		if len(expr.Arguments) > 1 {
			a.validateArrayIntegerArgAt(expr.Arguments[1], expr.Arguments[0].Pos())
		}
		return types.INTEGER
	case "insert":
		if len(expr.Arguments) < 2 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		a.validateArrayIntegerArg(expr.Arguments[0])
		if argType := a.analyzeExpressionWithExpectedType(expr.Arguments[1], arrayType.ElementType); argType != nil && !a.canAssign(argType, arrayType.ElementType) {
			a.addArrayHelperParamTypeExpectedAt(expr.Arguments[1].Pos(), arrayType.ElementType, argType)
		}
		return types.VOID
	case "move", "swap":
		if len(expr.Arguments) < 2 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		a.validateArrayIntegerArg(expr.Arguments[0])
		a.validateArrayIntegerArg(expr.Arguments[1])
		return types.VOID
	case "copy":
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return arrayType
		}
		if len(expr.Arguments) > 2 {
			a.addArrayHelperTooManyArgs(expr)
		}
		a.validateArrayIntegerArg(expr.Arguments[0])
		if len(expr.Arguments) > 1 {
			a.validateArrayIntegerArg(expr.Arguments[1])
		}
		return arrayType
	case "foreach":
		callbackType := types.NewProcedurePointerType([]types.Type{arrayType.ElementType})
		if len(expr.Arguments) == 0 {
			a.addArrayHelperTooFewArgs(expr)
			return types.VOID
		}
		if len(expr.Arguments) > 1 {
			a.addArrayHelperTooManyArgs(expr)
		}
		arg := expr.Arguments[0]
		argType := a.analyzeExpressionWithExpectedType(arg, callbackType)
		if argType != nil && !a.canAssign(argType, callbackType) {
			if name, fn, namePos, _ := a.namedArrayHelperCallable(arg); fn != nil {
				if len(fn.Parameters) != 1 || !a.canAssign(arrayType.ElementType, fn.Parameters[0]) {
					a.addStructuredError(NewNoOverloadMatchError(namePos, name))
					a.addArrayHelperParamTypeExpectedText(arg.Pos(), semanticFunctionPointerName(callbackType), callbackResultTypeName(fn))
					return types.VOID
				}
				a.addArrayHelperParamTypeExpectedText(arg.Pos(), semanticFunctionPointerName(callbackType), semanticNamedFunctionPointerName(name, fn))
				return types.VOID
			}
		}
		if argType != nil && !a.canAssign(argType, callbackType) {
			a.addArrayHelperParamTypeExpectedAt(arg.Pos(), callbackType, argType)
		}
		return types.VOID
	case "map":
		if len(expr.Arguments) != 1 {
			if len(expr.Arguments) < 1 {
				a.addArrayHelperTooFewArgs(expr)
			} else {
				a.addArrayHelperTooManyArgs(expr)
			}
			return types.NewDynamicArrayType(arrayType.ElementType)
		}
		arg := expr.Arguments[0]
		argType := a.analyzeExpression(arg)
		if argType == nil {
			return types.NewDynamicArrayType(arrayType.ElementType)
		}
		expectedType := types.NewFunctionPointerType([]types.Type{arrayType.ElementType}, types.VARIANT)
		if identExpr, ok := arg.(*ast.Identifier); ok {
			if sym, exists := a.symbols.Resolve(identExpr.Value); exists {
				if fnType, ok := sym.Type.(*types.FunctionType); ok && len(fnType.Parameters) == 1 {
					if fnType.ConstParams != nil && len(fnType.ConstParams) > 0 && fnType.ConstParams[0] {
						a.addArrayHelperError(identExpr.Token.Pos, `More arguments expected`)
						a.addArrayHelperParamTypeExpectedText(arg.Pos(),
							"function ("+semanticTypeNameForDiagnostic(arrayType.ElementType)+"): Any Type",
							semanticNamedFunctionSignature(identExpr.Value, fnType))
						return types.NewDynamicArrayType(arrayType.ElementType)
					}
				}
			}
		}
		if !a.canAssign(argType, expectedType) {
			a.addArrayHelperParamTypeExpectedAt(arg.Pos(), expectedType, argType)
		}
		return types.NewDynamicArrayType(arrayType.ElementType)
	case "sort":
		if len(expr.Arguments) > 1 {
			a.addArrayHelperTooManyArgs(expr)
			return arrayType
		}
		if len(expr.Arguments) == 0 {
			if !isArrayNaturallySortable(arrayType.ElementType) {
				a.addArrayHelperError(expr.Method.Token.Pos, "Array does not have a natural sort order")
			}
			return arrayType
		}
		comparatorType := types.NewFunctionPointerType([]types.Type{arrayType.ElementType, arrayType.ElementType}, types.INTEGER)
		arg := expr.Arguments[0]
		argType := a.analyzeExpressionWithExpectedType(arg, comparatorType)
		if argType != nil && !a.canAssign(argType, comparatorType) {
			if name, fn, namePos, isAddressOf := a.namedArrayHelperCallable(arg); fn != nil {
				if isAddressOf {
					a.addArrayHelperError(namePos, "More arguments expected")
					a.addArrayHelperError(arg.Pos(),
						`Incompatible types: "`+semanticTypeNameForDiagnostic(comparatorType)+`" and "`+semanticNamedFunctionPointerName(name, fn)+`"`)
					a.addArrayHelperParamTypeExpectedText(arg.Pos(), semanticFunctionPointerName(comparatorType), "nil")
					return arrayType
				}
				a.addArrayHelperError(namePos, "More arguments expected")
				a.addArrayHelperParamTypeExpectedText(arg.Pos(), semanticFunctionPointerName(comparatorType), callbackResultTypeName(fn))
				return arrayType
			}
			a.addArrayHelperError(arg.Pos(),
				`Incompatible types: "`+semanticTypeNameForDiagnostic(comparatorType)+`" and "`+semanticTypeNameForDiagnostic(argType)+`"`)
			a.addArrayHelperParamTypeExpectedAt(arg.Pos(), comparatorType, argType)
		}
		return arrayType
	}

	return nil
}
