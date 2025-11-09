package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/token"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// analyzeCallExpressionWithContext analyzes a call expression with optional expected type context.
// Task 9.19.2: The expected type can help with overload resolution in the future.
// Currently, this is a wrapper that delegates to analyzeCallExpression, but can be extended
// to use the expected type for disambiguating between multiple overloads.
func (a *Analyzer) analyzeCallExpressionWithContext(expr *ast.CallExpression, expectedType types.Type) types.Type {
	// TODO: Use expectedType for overload resolution when overloading is implemented
	// For now, just delegate to the regular analysis
	_ = expectedType // Mark as intentionally unused for now
	return a.analyzeCallExpression(expr)
}

func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access expressions (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		// Task 9.13-9.19: Check if this is a constructor call (TClass.Create(...))
		// First, analyze the object to see if it's a class type
		objectType := a.analyzeExpression(memberAccess.Object)
		if objectType == nil {
			return nil
		}

		// Check if the object is a ClassType (not an instance, but the type itself)
		if classType, isClassType := objectType.(*types.ClassType); isClassType {
			// This is a constructor call like TClass.Create(args)
			// fmt.Printf("DEBUG: Detected constructor call on class %s.%s\n", classType.Name, memberAccess.Member.Value)
			return a.analyzeConstructorCall(expr, classType, memberAccess.Member.Value)
		}

		// Task 9.73.2: Check if the object is a ClassOfType (metaclass variable)
		if metaclassType, isMetaclassType := objectType.(*types.ClassOfType); isMetaclassType {
			// This is a constructor call through a metaclass variable like cls.Create(args)
			// where cls is of type "class of TBase"
			// The constructor creates an instance of the base class (or its descendants at runtime)
			return a.analyzeConstructorCall(expr, metaclassType.ClassType, memberAccess.Member.Value)
		}

		// Not a constructor call - analyze as normal method call
		// Analyze the member access to get the method type
		methodType := a.analyzeMemberAccessExpression(memberAccess)
		if methodType == nil {
			// Error already reported by analyzeMemberAccessExpression
			return nil
		}

		// Verify it's a function type
		funcType, ok := methodType.(*types.FunctionType)
		if !ok {
			a.addError("cannot call non-function type %s at %s",
				methodType.String(), expr.Token.Pos.String())
			return nil
		}

		// Validate argument count
		if len(expr.Arguments) != len(funcType.Parameters) {
			a.addError("method call expects %d argument(s), got %d at %s",
				len(funcType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
		}

		// Validate argument types
		for i, arg := range expr.Arguments {
			if i >= len(funcType.Parameters) {
				break // Already reported count mismatch
			}

			// Task 9.2b: Validate var parameter receives an lvalue
			isVar := len(funcType.VarParams) > i && funcType.VarParams[i]
			if isVar && !a.isLValue(arg) {
				a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
					i+1, arg.String(), arg.Pos().String())
			}

			paramType := funcType.Parameters[i]
			argType := a.analyzeExpressionWithExpectedType(arg, paramType)
			if argType != nil && !a.canAssign(argType, paramType) {
				a.addError("argument %d has type %s, expected %s at %s",
					i+1, argType.String(), paramType.String(),
					expr.Token.Pos.String())
			}
		}

		return funcType.ReturnType
	}

	// Handle regular function calls (identifier-based)
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("function call must use identifier or member access at %s", expr.Token.Pos.String())
		return nil
	}

	// Look up function
	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check if it's a built-in function (using new dispatcher)
		if resultType, isBuiltin := a.analyzeBuiltinFunction(funcIdent.Value, expr.Arguments, expr); isBuiltin {
			return resultType
		}

		// Low built-in function

		// Assert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Assert") {
			// Assert takes 1-2 arguments: Boolean condition and optional String message
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Assert' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// First argument must be Boolean
			condType := a.analyzeExpression(expr.Arguments[0])
			if condType != nil && condType != types.BOOLEAN {
				a.addError("function 'Assert' first argument must be Boolean, got %s at %s",
					condType.String(), expr.Token.Pos.String())
			}
			// If there's a second argument (message), it must be String
			if len(expr.Arguments) == 2 {
				msgType := a.analyzeExpression(expr.Arguments[1])
				if msgType != nil && msgType != types.STRING {
					a.addError("function 'Assert' second argument must be String, got %s at %s",
						msgType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Insert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Insert") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Insert' expects 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			sourceType := a.analyzeExpression(expr.Arguments[0])
			if sourceType != nil && sourceType != types.STRING {
				a.addError("function 'Insert' first argument must be String, got %s at %s",
					sourceType.String(), expr.Token.Pos.String())
			}
			if _, ok := expr.Arguments[1].(*ast.Identifier); !ok {
				a.addError("function 'Insert' second argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				targetType := a.analyzeExpression(expr.Arguments[1])
				if targetType != nil && targetType != types.STRING {
					a.addError("function 'Insert' second argument must be String, got %s at %s",
						targetType.String(), expr.Token.Pos.String())
				}
			}
			posType := a.analyzeExpression(expr.Arguments[2])
			if posType != nil && posType != types.INTEGER {
				a.addError("function 'Insert' third argument must be Integer, got %s at %s",
					posType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Task 9.227: Higher-order functions for working with lambdas
		if strings.EqualFold(funcIdent.Value, "Map") {
			// Map(array, lambda) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Map' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Filter") {
			// Filter(array, predicate) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Filter' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Reduce") {
			// Reduce(array, lambda, initial) -> value
			if len(expr.Arguments) != 3 {
				a.addError("function 'Reduce' expects 3 arguments (array, lambda, initial), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			initialType := a.analyzeExpression(expr.Arguments[2])

			// Return type is the same as initial value type
			return initialType
		}

		if strings.EqualFold(funcIdent.Value, "ForEach") {
			// ForEach(array, lambda) -> void
			if len(expr.Arguments) != 2 {
				a.addError("function 'ForEach' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			return types.VOID
		}

		// Task 9.95-9.97: Current date/time functions

		// Allow calling methods within the current class without explicit Self
		if a.currentClass != nil {
			if methodType, found := a.currentClass.GetMethod(funcIdent.Value); found {
				if len(expr.Arguments) != len(methodType.Parameters) {
					a.addError("method '%s' expects %d arguments, got %d at %s",
						funcIdent.Value, len(methodType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
					return methodType.ReturnType
				}
				for i, arg := range expr.Arguments {
					argType := a.analyzeExpression(arg)
					expectedType := methodType.Parameters[i]
					if argType != nil && !a.canAssign(argType, expectedType) {
						a.addError("argument %d to method '%s' has type %s, expected %s at %s",
							i+1, funcIdent.Value, argType.String(), expectedType.String(), expr.Token.Pos.String())
					}
				}
				return methodType.ReturnType
			}
		}

		// Task 9.232: Variant introspection functions

		// Task 9.114: GetStackTrace() built-in function
		if strings.EqualFold(funcIdent.Value, "GetStackTrace") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetStackTrace' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns String
			return types.STRING
		}

		// Task 9.116: GetCallStack() built-in function
		if strings.EqualFold(funcIdent.Value, "GetCallStack") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetCallStack' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns dynamic array of records
			// Each record has: FunctionName: String, FileName: String, Line: Integer, Column: Integer
			// For simplicity in semantic analysis, we return a generic dynamic array type
			return types.NewDynamicArrayType(types.VARIANT)
		}

		// Task 9.8.1-9.8.2: Check if this is a type cast (TypeName(expression))
		// Type casts look like function calls but the "function" name is actually a type name
		if castType := a.analyzeTypeCast(funcIdent.Value, expr.Arguments, expr); castType != nil {
			return castType
		}

		a.addError("undefined function '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Task 9.65-9.66: Check if this is an overloaded function
	// If so, resolve the overload set to select the best match
	var funcType *types.FunctionType
	if sym.IsOverloadSet {
		// Get all overload candidates
		candidates := a.symbols.GetOverloadSet(funcIdent.Value)
		if candidates == nil || len(candidates) == 0 {
			a.addError("no overload candidates found for '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}

		// Task 9.21.5: Detect overloaded function calls with lambda arguments
		hasLambdas, lambdaIndices := detectOverloadedCallWithLambdas(expr.Arguments)
		if hasLambdas {
			// We have lambda arguments that need type inference
			// This requires special handling (tasks 9.21.6-9.21.7)
			// For now, report that lambdas need explicit types when calling overloaded functions
			a.addError("lambda type inference not yet supported for overloaded function '%s' - please provide explicit parameter types for lambda at argument position(s) %v at %s",
				funcIdent.Value, lambdaIndices, expr.Token.Pos.String())
			return nil
		}

		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return nil // Error already reported
			}
			argTypes[i] = argType
		}

		// Resolve overload based on argument types
		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			// Task 9.63: Provide DWScript-compatible error message for failed overload resolution
			a.addError("Syntax Error: There is no overloaded version of \"%s\" that can be called with these arguments [line: %d, column: %d]",
				funcIdent.Value, expr.Token.Pos.Line, expr.Token.Pos.Column)
			return nil
		}

		// Use the selected overload's function type
		var ok bool
		funcType, ok = selected.Type.(*types.FunctionType)
		if !ok {
			a.addError("selected overload for '%s' is not a function type at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}
	} else {
		// Task 9.162: Check if it's a function pointer type first
		if funcPtrType := a.analyzeFunctionPointerCall(expr, sym.Type); funcPtrType != nil {
			return funcPtrType
		}

		// Check that symbol is a function
		var ok bool
		funcType, ok = sym.Type.(*types.FunctionType)
		if !ok {
			a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
			return nil
		}
	}

	// Task 9.1: Check argument count with optional parameters support
	// Count required parameters (those without defaults)
	requiredParams := 0
	for _, defaultVal := range funcType.DefaultValues {
		if defaultVal == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(expr.Arguments) < requiredParams {
		// Use more precise error message based on whether function has optional parameters
		if requiredParams == len(funcType.Parameters) {
			// All parameters are required
			a.addError("function '%s' expects %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		} else {
			// Function has optional parameters
			a.addError("function '%s' expects at least %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		}
		return nil
	}
	if len(expr.Arguments) > len(funcType.Parameters) {
		a.addError("function '%s' expects at most %d arguments, got %d at %s",
			funcIdent.Value, len(funcType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return nil
	}

	// Check argument types
	// Task 9.137: Handle lazy parameters - validate expression type without evaluating
	// Task 9.2b: Handle var parameters - validate that argument is an lvalue
	for i, arg := range expr.Arguments {
		expectedType := funcType.Parameters[i]

		// Check if this parameter is lazy
		isLazy := len(funcType.LazyParams) > i && funcType.LazyParams[i]

		// Check if this parameter is var (by-reference)
		isVar := len(funcType.VarParams) > i && funcType.VarParams[i]

		// Task 9.2b: Validate var parameter receives an lvalue
		if isVar && !a.isLValue(arg) {
			a.addError("var parameter %d to function '%s' requires a variable (identifier, array element, or field), got %s at %s",
				i+1, funcIdent.Value, arg.String(), arg.Pos().String())
		}

		if isLazy {
			// For lazy parameters, check expression type but don't evaluate
			// The expression will be passed as-is to the interpreter for deferred evaluation
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("lazy argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		} else {
			// Regular parameter: validate type normally
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}

	return funcType.ReturnType
}

// analyzeConstructorCall analyzes a constructor call like TClass.Create(args)
// Implements tasks 9.13-9.16:
//   - Task 9.13: Constructor overload resolution based on argument types
//   - Task 9.14: Constructor parameter type validation
//   - Task 9.15: Constructor parameter count validation
//   - Task 9.16: Constructor visibility enforcement
func (a *Analyzer) analyzeConstructorCall(expr *ast.CallExpression, classType *types.ClassType, constructorName string) types.Type {
	// Task 9.13: Get all constructor overloads for this name (case-insensitive)
	// Use getMethodOverloadsInHierarchy which already handles constructors
	constructorOverloads := a.getMethodOverloadsInHierarchy(constructorName, classType)

	if len(constructorOverloads) == 0 {
		// No constructor with this name found
		a.addError("class '%s' has no constructor named '%s' at %s",
			classType.Name, constructorName, expr.Token.Pos.String())
		return classType
	}

	// Task 9.13: Resolve overload based on argument types
	var selectedConstructor *types.MethodInfo
	var selectedSignature *types.FunctionType

	if len(constructorOverloads) == 1 {
		// Single constructor - no overload resolution needed
		selectedConstructor = constructorOverloads[0]
		selectedSignature = selectedConstructor.Signature
	} else {
		// Multiple constructors - perform overload resolution
		// Analyze argument types first
		argTypes := make([]types.Type, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			argType := a.analyzeExpression(arg)
			if argType == nil {
				return classType // Error already reported
			}
			argTypes[i] = argType
		}

		// Convert MethodInfo to Symbol for ResolveOverload
		candidates := make([]*Symbol, len(constructorOverloads))
		for i, overload := range constructorOverloads {
			candidates[i] = &Symbol{
				Type: overload.Signature,
			}
		}

		// Resolve overload based on argument types
		selected, err := ResolveOverload(candidates, argTypes)
		if err != nil {
			// Task 9.15: Report clear error for wrong argument count/types
			a.addError("there is no overloaded constructor '%s' that can be called with these arguments at %s",
				constructorName, expr.Token.Pos.String())
			return classType
		}

		selectedSignature = selected.Type.(*types.FunctionType)
		// Find the matching MethodInfo
		for _, overload := range constructorOverloads {
			if overload.Signature == selectedSignature {
				selectedConstructor = overload
				break
			}
		}
	}

	// Task 9.16: Check constructor visibility
	// Find the class that owns this constructor
	var ownerClass *types.ClassType
	for class := classType; class != nil; class = class.Parent {
		if class.HasConstructor(constructorName) {
			ownerClass = class
			break
		}
	}
	if ownerClass != nil && selectedConstructor != nil {
		visibility := selectedConstructor.Visibility
		if !a.checkVisibility(ownerClass, visibility, constructorName, "constructor") {
			visibilityStr := ast.Visibility(visibility).String()
			a.addError("cannot access %s constructor '%s' of class '%s' at %s",
				visibilityStr, constructorName, ownerClass.Name, expr.Token.Pos.String())
			return classType
		}
	}

	// Task 9.15: Validate argument count
	if len(expr.Arguments) != len(selectedSignature.Parameters) {
		a.addError("constructor '%s' expects %d arguments, got %d at %s",
			constructorName, len(selectedSignature.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return classType
	}

	// Task 9.14: Validate argument types
	for i, arg := range expr.Arguments {
		if i >= len(selectedSignature.Parameters) {
			break
		}

		paramType := selectedSignature.Parameters[i]
		argType := a.analyzeExpressionWithExpectedType(arg, paramType)
		if argType != nil && !a.canAssign(argType, paramType) {
			a.addError("argument %d to constructor '%s' has type %s, expected %s at %s",
				i+1, constructorName, argType.String(), paramType.String(),
				expr.Token.Pos.String())
		}
	}

	// Constructor calls return an instance of the class
	return classType
}

// analyzeOldExpression analyzes an 'old' expression in a postcondition
// Task 9.143: Return the type of the referenced identifier
func (a *Analyzer) analyzeOldExpression(expr *ast.OldExpression) types.Type {
	if expr.Identifier == nil {
		return nil
	}

	// Look up the identifier in the symbol table
	sym, ok := a.symbols.Resolve(expr.Identifier.Value)
	if !ok {
		// Error already reported in validateOldExpressions
		return nil
	}

	// Return the type of the identifier
	return sym.Type
}

// ============================================================================
// Task 9.21.5: Overload Detection with Lambda Arguments
// ============================================================================

// isLambdaNeedingInference checks if an expression is a lambda that needs type inference.
// Returns true if the expression is a lambda with untyped parameters or untyped return.
func isLambdaNeedingInference(expr ast.Expression) bool {
	lambda, ok := expr.(*ast.LambdaExpression)
	if !ok {
		return false
	}

	// Check if any parameters lack type annotations
	for _, param := range lambda.Parameters {
		if param.Type == nil {
			return true // Parameter needs type inference
		}
	}

	// Check if return type lacks annotation (for statement-body lambdas)
	// Note: shorthand lambdas always need return type inference from body
	if lambda.ReturnType == nil {
		return true
	}

	return false
}

// detectOverloadedCallWithLambdas detects when we're calling an overloaded function
// with lambda arguments that need type inference.
// Returns (hasLambdas bool, lambdaIndices []int)
func detectOverloadedCallWithLambdas(args []ast.Expression) (bool, []int) {
	lambdaIndices := []int{}

	for i, arg := range args {
		if isLambdaNeedingInference(arg) {
			lambdaIndices = append(lambdaIndices, i)
		}
	}

	return len(lambdaIndices) > 0, lambdaIndices
}

// analyzeTypeCast analyzes a type cast expression like Integer(x), Float(y), or TMyClass(obj).
// Returns the target type if this is a valid type cast, or nil if not a type cast.
// Task 9.8.2: Semantic analysis for type casts
func (a *Analyzer) analyzeTypeCast(typeName string, args []ast.Expression, expr *ast.CallExpression) types.Type {
	// Type casts must have exactly one argument
	if len(args) != 1 {
		return nil // Not a type cast
	}

	// Check if typeName is a built-in type
	var targetType types.Type

	// Check for built-in types (case-insensitive)
	switch strings.ToLower(typeName) {
	case "integer":
		targetType = types.INTEGER
	case "float":
		targetType = types.FLOAT
	case "string":
		targetType = types.STRING
	case "boolean":
		targetType = types.BOOLEAN
	case "variant":
		targetType = types.VARIANT
	default:
		// Check if it's a class type (case-insensitive)
		for name, class := range a.classes {
			if strings.EqualFold(name, typeName) {
				targetType = class
				break
			}
		}
		// If not a class, check in symbol table
		if targetType == nil {
			sym, ok := a.symbols.Resolve(typeName)
			if ok {
				targetType = sym.Type
			} else {
				return nil // Not a type name
			}
		}
	}

	// Analyze the argument expression
	argType := a.analyzeExpression(args[0])
	if argType == nil {
		return nil // Error already reported
	}

	// Validate the cast is legal
	if !a.isValidCast(argType, targetType, expr.Token.Pos) {
		return nil
	}

	return targetType
}

// isValidCast checks if casting from sourceType to targetType is valid.
// Task 9.8.2: Type cast validation
func (a *Analyzer) isValidCast(sourceType, targetType types.Type, pos token.Position) bool {
	// Resolve type aliases
	sourceType = types.GetUnderlyingType(sourceType)
	targetType = types.GetUnderlyingType(targetType)

	// Same type is always valid
	if sourceType.Equals(targetType) {
		return true
	}

	// Numeric conversions
	if a.isNumericType(sourceType) && a.isNumericType(targetType) {
		return true // Integer <-> Float allowed
	}

	// Boolean conversions
	if targetType == types.BOOLEAN {
		if sourceType == types.INTEGER || sourceType == types.FLOAT {
			return true // Integer/Float -> Boolean (0 = false, non-zero = true)
		}
	}

	// String conversions
	if targetType == types.STRING {
		// Most types can be converted to string
		return true
	}

	// Variant conversions (Variant can be cast to/from anything)
	if sourceType == types.VARIANT || targetType == types.VARIANT {
		return true
	}

	// Class/Interface casts
	sourceClass, sourceIsClass := sourceType.(*types.ClassType)
	targetClass, targetIsClass := targetType.(*types.ClassType)

	if sourceIsClass && targetIsClass {
		// Class to class cast: check if types are related by inheritance
		if types.IsSubclassOf(sourceClass, targetClass) ||
			types.IsSubclassOf(targetClass, sourceClass) {
			return true // Upcast or downcast allowed
		}
		a.addError("cannot cast %s to %s: types are not related by inheritance at %s",
			sourceType.String(), targetType.String(), pos.String())
		return false
	}

	// Interface casts
	_, sourceIsInterface := sourceType.(*types.InterfaceType)
	_, targetIsInterface := targetType.(*types.InterfaceType)

	if sourceIsInterface || targetIsInterface {
		// Interface casts are allowed (checked at runtime)
		return true
	}

	// Class to interface or interface to class
	if (sourceIsClass && targetIsInterface) || (sourceIsInterface && targetIsClass) {
		return true // Allowed, checked at runtime
	}

	// Enum casts to/from Integer
	_, sourceIsEnum := sourceType.(*types.EnumType)
	_, targetIsEnum := targetType.(*types.EnumType)

	if (sourceIsEnum && targetType == types.INTEGER) ||
		(sourceType == types.INTEGER && targetIsEnum) {
		return true
	}

	// Otherwise, report error
	a.addError("cannot cast %s to %s at %s",
		sourceType.String(), targetType.String(), pos.String())
	return false
}

// isNumericType checks if a type is Integer or Float
func (a *Analyzer) isNumericType(t types.Type) bool {
	t = types.GetUnderlyingType(t)
	return t == types.INTEGER || t == types.FLOAT
}
