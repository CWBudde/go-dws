package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// analyzeExpression analyzes an expression and returns its type.
// Returns nil if the expression is invalid.
func (a *Analyzer) analyzeExpression(expr ast.Expression) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return types.INTEGER
	case *ast.FloatLiteral:
		return types.FLOAT
	case *ast.StringLiteral:
		return types.STRING
	case *ast.BooleanLiteral:
		return types.BOOLEAN
	case *ast.CharLiteral:
		// Character literals are treated as single-character strings in DWScript
		return types.STRING
	case *ast.NilLiteral:
		return types.NIL
	case *ast.Identifier:
		return a.analyzeIdentifier(e)
	case *ast.BinaryExpression:
		return a.analyzeBinaryExpression(e)
	case *ast.UnaryExpression:
		return a.analyzeUnaryExpression(e)
	case *ast.GroupedExpression:
		return a.analyzeExpression(e.Expression)
	case *ast.CallExpression:
		return a.analyzeCallExpression(e)
	case *ast.NewExpression:
		return a.analyzeNewExpression(e)
	case *ast.NewArrayExpression:
		return a.analyzeNewArrayExpression(e)
	case *ast.MemberAccessExpression:
		return a.analyzeMemberAccessExpression(e)
	case *ast.MethodCallExpression:
		return a.analyzeMethodCallExpression(e)
	case *ast.ArrayLiteralExpression:
		return a.analyzeArrayLiteral(e, nil)
	case *ast.RecordLiteralExpression:
		// Typed record literals can be analyzed standalone
		if e.TypeName != nil {
			return a.analyzeRecordLiteral(e, nil)
		}
		// Anonymous record literals need context from variable declaration or assignment
		a.addError("anonymous record literal requires type context (use explicit type annotation)")
		return nil
	case *ast.SetLiteral:
		// SetLiteral needs context to know the expected type
		// This will be handled in analyzeVarDecl or analyzeAssignment
		return a.analyzeSetLiteralWithContext(e, nil)
	case *ast.IndexExpression:
		return a.analyzeIndexExpression(e)
	case *ast.AddressOfExpression:
		return a.analyzeAddressOfExpression(e)
	case *ast.LambdaExpression:
		return a.analyzeLambdaExpression(e)
	case *ast.OldExpression:
		return a.analyzeOldExpression(e)
	case *ast.InheritedExpression:
		return a.analyzeInheritedExpression(e)
	case *ast.SelfExpression:
		return a.analyzeSelfExpression(e)
	case *ast.IsExpression:
		return a.analyzeIsExpression(e)
	case *ast.AsExpression:
		return a.analyzeAsExpression(e)
	case *ast.ImplementsExpression:
		return a.analyzeImplementsExpression(e)
	case *ast.IfExpression:
		return a.analyzeIfExpression(e)
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// isBooleanCompatible checks if a type can be implicitly converted to Boolean.
// Includes Boolean itself and Variant (supports implicit boolean conversion).
func isBooleanCompatible(t types.Type) bool {
	if t == nil {
		return false
	}
	if t.Equals(types.BOOLEAN) || t.Equals(types.VARIANT) {
		return true
	}
	return false
}

// analyzeExpressionWithExpectedType analyzes an expression with optional type context.
// Enables context-sensitive type inference for:
// - RecordLiteral, SetLiteral, ArrayLiteral (expected type determines type conversions)
// - Lambda (parameter types from function pointer), Nil (class/interface type)
// - Integer (float when context expects Float), Call (overload resolution)
// For other types, falls back to analyzeExpression() without context.
func (a *Analyzer) analyzeExpressionWithExpectedType(expr ast.Expression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.RecordLiteralExpression:
		return a.analyzeRecordLiteral(e, expectedType)
	case *ast.SetLiteral:
		// Convert SetLiteral to ArrayLiteral when expected type is array
		if expectedType != nil {
			if _, ok := types.GetUnderlyingType(expectedType).(*types.ArrayType); ok {
				arrayLit := &ast.ArrayLiteralExpression{
					TypedExpressionBase: ast.TypedExpressionBase{
						BaseNode: ast.BaseNode{Token: e.Token},
					},
					Elements: e.Elements,
				}
				resultType := a.analyzeArrayLiteral(arrayLit, expectedType)
				if resultType != nil {
					a.semanticInfo.SetType(e, &ast.TypeAnnotation{
						Token: e.Token,
						Name:  resultType.String(),
					})
				}
				return resultType
			}
		}
		return a.analyzeSetLiteralWithContext(e, expectedType)
	case *ast.ArrayLiteralExpression:
		if expectedType != nil {
			if _, ok := types.GetUnderlyingType(expectedType).(*types.SetType); ok {
				convertible := len(e.Elements) == 0
				if !convertible {
					convertible = true
					for _, elem := range e.Elements {
						switch elem.(type) {
						case *ast.Identifier, *ast.RangeExpression, *ast.IntegerLiteral:
						case *ast.CharLiteral, *ast.StringLiteral, *ast.BooleanLiteral:
						default:
							convertible = false
						}
					}
				}
				if convertible {
					setLit := &ast.SetLiteral{
						TypedExpressionBase: e.TypedExpressionBase,
						Elements:            e.Elements,
					}

					resultType := a.analyzeSetLiteralWithContext(setLit, expectedType)
					if resultType != nil && a.semanticInfo != nil {
						a.semanticInfo.SetType(e, &ast.TypeAnnotation{
							Token: e.Token,
							Name:  resultType.String(),
						})
					}
					return resultType
				}
			}
		}
		return a.analyzeArrayLiteral(e, expectedType)
	case *ast.LambdaExpression:
		// Infer parameter types from function pointer context
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			if funcPtrType, ok := underlyingType.(*types.FunctionPointerType); ok {
				return a.analyzeLambdaExpressionWithContext(e, funcPtrType)
			}
		}
		return a.analyzeLambdaExpression(e)
	case *ast.NilLiteral:
		// Infer type from context (class, interface, or function pointer)
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			typeKind := underlyingType.TypeKind()
			if typeKind == "CLASS" || typeKind == "INTERFACE" || typeKind == "FUNCTION_POINTER" {
				return expectedType
			}
		}
		return types.NIL
	case *ast.IntegerLiteral:
		// Treat as float if context expects Float
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			if underlyingType.TypeKind() == "FLOAT" {
				return types.FLOAT
			}
		}
		return types.INTEGER
	case *ast.FloatLiteral:
		// Float literals are always FLOAT type regardless of context
		return types.FLOAT
	case *ast.CallExpression:
		// Pass expected type for overload resolution
		return a.analyzeCallExpressionWithContext(e, expectedType)
	default:
		return a.analyzeExpression(expr)
	}
}

// analyzeIsExpression analyzes the 'is' operator.
// Example: obj is TMyClass -> Boolean (type check)
// Example: boolExpr is True -> Boolean (boolean comparison)
// Returns Boolean type.
func (a *Analyzer) analyzeIsExpression(expr *ast.IsExpression) types.Type {
	// Analyze the left expression
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// Check if this is a boolean value comparison (expr.Right is set)
	// or a type check (expr.TargetType is set)
	if expr.Right != nil {
		// Boolean value comparison: left is right
		// Analyze the right expression
		rightType := a.analyzeExpression(expr.Right)
		if rightType == nil {
			return nil
		}

		// Both sides are convertible to boolean via implicit coercion
		// The interpreter and bytecode VM handle conversion using isTruthy/variantToBool

		// The 'is' operator always returns Boolean
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Token: expr.Token,
			Name:  "Boolean",
		})
		return types.BOOLEAN
	}

	// Type checking mode
	// Resolve the target type (should be a class type)
	targetType, err := a.resolveTypeExpression(expr.TargetType)
	if err != nil || targetType == nil {
		a.addError("cannot resolve target type in 'is' expression at %s: %v", expr.Token.Pos.String(), err)
		return nil
	}

	// Validate that left operand is a class (or nil)
	if leftType != types.NIL {
		leftUnderlying := types.GetUnderlyingType(leftType)
		if _, isClass := leftUnderlying.(*types.ClassType); !isClass {
			a.addError("'is' operator requires class instance, got %s at %s",
				leftType.String(), expr.Token.Pos.String())
			return nil
		}
	}

	// Validate that target type is a class type
	targetUnderlying := types.GetUnderlyingType(targetType)
	if _, isClass := targetUnderlying.(*types.ClassType); !isClass {
		a.addError("'is' operator requires class type, got %s at %s",
			targetType.String(), expr.Token.Pos.String())
		return nil
	}

	// The 'is' operator always returns Boolean
	a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
		Token: expr.Token,
		Name:  "Boolean",
	})
	return types.BOOLEAN
}

// analyzeAsExpression analyzes the 'as' type casting operator.
// Example: obj as IMyInterface or child as TParent
// Supports both interface casting and class-to-class casting (up/down casting).
// Returns the target type.
func (a *Analyzer) analyzeAsExpression(expr *ast.AsExpression) types.Type {
	// Analyze the left expression (the object being cast)
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// Resolve the target type (should be an interface or class type)
	targetType, err := a.resolveTypeExpression(expr.TargetType)
	if err != nil || targetType == nil {
		a.addError("cannot resolve target type in 'as' expression at %s: %v", expr.Token.Pos.String(), err)
		return nil
	}

	// Allow Variant casts to primitive types (and other runtime-resolved targets)
	// DWScript permits using "as" with Variant to force a runtime conversion,
	// e.g. VariantValue as Integer/String/Float/Boolean.
	if leftType.Equals(types.VARIANT) {
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Token: expr.Token,
			Name:  targetType.String(),
		})
		return targetType
	}

	// Target type can be either interface OR class
	targetUnderlying := types.GetUnderlyingType(targetType)
	interfaceType, isInterface := targetUnderlying.(*types.InterfaceType)
	classTargetType, isClassTarget := targetUnderlying.(*types.ClassType)

	if !isInterface && !isClassTarget {
		a.addError("'as' operator requires class or interface type, got %s at %s",
			targetType.String(), expr.Token.Pos.String())
		return targetType
	}

	// Validate that left type is a class or object
	leftUnderlying := types.GetUnderlyingType(leftType)
	classType, isClass := leftUnderlying.(*types.ClassType)

	// Also allow NIL to be cast to any interface or class
	if leftType == types.NIL {
		// Set the expression type and return
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Token: expr.Token,
			Name:  targetType.String(), // Use the actual target type name
		})
		return targetType
	}

	if !isClass {
		a.addError("'as' operator requires class instance, got %s at %s",
			leftType.String(), expr.Token.Pos.String())
		return nil
	}

	// Handle class-to-class casting
	if isClassTarget {
		// For class-to-class casting, we check inheritance relationship
		// Both upcast (child to parent) and downcast (parent to child) are allowed
		// Downcast safety is checked at runtime
		if !types.IsClassRelated(classType, classTargetType) {
			a.addError("cannot cast '%s' to unrelated class '%s' at %s",
				classType.Name, classTargetType.Name, expr.Token.Pos.String())
			return nil
		}

		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Token: expr.Token,
			Name:  classTargetType.Name,
		})
		return classTargetType
	}

	// Handle class-to-interface casting
	// Validate that the class implements the interface
	if !types.ImplementsInterface(classType, interfaceType) {
		a.addError("class '%s' does not implement interface '%s' at %s",
			classType.Name, interfaceType.Name, expr.Token.Pos.String())
		return nil
	}

	// Set the expression type annotation
	a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
		Token: expr.Token,
		Name:  interfaceType.Name,
	})

	return interfaceType
}

// analyzeImplementsExpression analyzes the 'implements' operator.
// Example: obj implements IMyInterface -> Boolean
// Checks whether the object's class implements the target interface.
// Always returns Boolean type.
func (a *Analyzer) analyzeImplementsExpression(expr *ast.ImplementsExpression) types.Type {
	// Analyze the left expression (the object or class being checked)
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		return nil
	}

	// Resolve the target type (should be an interface type)
	targetType, err := a.resolveTypeExpression(expr.TargetType)
	if err != nil || targetType == nil {
		a.addError("cannot resolve target type in 'implements' expression at %s: %v", expr.Token.Pos.String(), err)
		return nil
	}

	// Validate that target type is an interface
	_, ok := types.GetUnderlyingType(targetType).(*types.InterfaceType)
	if !ok {
		a.addError("'implements' operator requires interface type, got %s at %s",
			targetType.String(), expr.Token.Pos.String())
		return nil
	}

	// Validate that left operand is a class (or nil)
	if leftType != types.NIL {
		leftUnderlying := types.GetUnderlyingType(leftType)
		if _, isClass := leftUnderlying.(*types.ClassType); !isClass {
			a.addError("'implements' operator requires class instance, got %s at %s",
				leftType.String(), expr.Token.Pos.String())
			return nil
		}
	}

	// Set the expression type annotation to Boolean
	a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
		Token: expr.Token,
		Name:  "Boolean",
	})

	return types.BOOLEAN
}

// analyzeIfExpression analyzes an inline if-then-else conditional expression.
// Syntax: if <condition> then <expression> [else <expression>]
// Returns the common type of the consequence and alternative branches.
func (a *Analyzer) analyzeIfExpression(expr *ast.IfExpression) types.Type {
	// Check that condition is boolean or Variant
	condType := a.analyzeExpression(expr.Condition)
	if condType != nil && !isBooleanCompatible(condType) {
		a.addError("if expression condition must be boolean, got %s at %s",
			condType.String(), expr.Token.Pos.String())
	}

	// Analyze consequence expression
	consequenceType := a.analyzeExpression(expr.Consequence)
	if consequenceType == nil {
		a.addError("invalid consequence expression in if-then-else at %s", expr.Token.Pos.String())
		return nil
	}

	var resultType types.Type

	// Analyze alternative if present
	if expr.Alternative != nil {
		alternativeType := a.analyzeExpression(expr.Alternative)
		if alternativeType == nil {
			a.addError("invalid alternative expression in if-then-else at %s", expr.Token.Pos.String())
			return nil
		}

		// Find common type between consequence and alternative
		resultType = a.findCommonType(consequenceType, alternativeType)
		if resultType == nil {
			a.addError("incompatible types in if-then-else: %s and %s at %s",
				consequenceType.String(), alternativeType.String(), expr.Token.Pos.String())
			return nil
		}
	} else {
		// No else clause - result type is the consequence type
		// When condition is false, default value of the type is returned
		resultType = consequenceType
	}

	// Set type annotation on the expression
	if resultType != nil {
		a.semanticInfo.SetType(expr, &ast.TypeAnnotation{
			Token: expr.Token,
			Name:  resultType.String(),
		})
	}

	return resultType
}

// findCommonType finds a common type between two types.
// This handles type compatibility for if-then-else expressions:
// - Same types return that type
// - Integer and Float return Float (wider type)
// - For class types, return common base class
// - For nil and class, return the class type
func (a *Analyzer) findCommonType(t1, t2 types.Type) types.Type {
	if t1 == nil || t2 == nil {
		return nil
	}

	// Same types
	if t1.Equals(t2) {
		return t1
	}

	// Integer + Float = Float (wider type)
	if (t1.Equals(types.INTEGER) && t2.Equals(types.FLOAT)) ||
		(t1.Equals(types.FLOAT) && t2.Equals(types.INTEGER)) {
		return types.FLOAT
	}

	// Handle nil compatibility with class types
	if t1.Equals(types.NIL) {
		if _, ok := types.GetUnderlyingType(t2).(*types.ClassType); ok {
			return t2
		}
	}
	if t2.Equals(types.NIL) {
		if _, ok := types.GetUnderlyingType(t1).(*types.ClassType); ok {
			return t1
		}
	}

	// For class types, find common base class
	class1, ok1 := types.GetUnderlyingType(t1).(*types.ClassType)
	class2, ok2 := types.GetUnderlyingType(t2).(*types.ClassType)
	if ok1 && ok2 {
		// Find common ancestor
		return a.findCommonBaseClass(class1, class2)
	}

	// For "class of" types (metaclasses), find common base metaclass
	classOf1, ok1 := types.GetUnderlyingType(t1).(*types.ClassOfType)
	classOf2, ok2 := types.GetUnderlyingType(t2).(*types.ClassOfType)
	if ok1 && ok2 {
		// Get the underlying class types
		if classOf1.ClassType != nil && classOf2.ClassType != nil {
			// Find common base class
			commonBase := a.findCommonBaseClass(classOf1.ClassType, classOf2.ClassType)
			if commonBase != nil {
				// Wrap the common base in a ClassOfType
				if baseClass, ok := commonBase.(*types.ClassType); ok {
					return types.NewClassOfType(baseClass)
				}
			}
		}
	}

	// No common type found
	return nil
}

// findCommonBaseClass finds the common base class between two class types.
func (a *Analyzer) findCommonBaseClass(c1, c2 *types.ClassType) types.Type {
	// Build ancestor chain for c1
	ancestors1 := make(map[string]bool)
	current := c1
	for current != nil {
		ancestors1[current.Name] = true
		current = current.Parent
	}

	// Walk c2's ancestor chain and find first match
	current = c2
	for current != nil {
		if ancestors1[current.Name] {
			return current
		}
		current = current.Parent
	}

	// No common base found - should at least be TObject in a well-formed hierarchy
	return nil
}

// analyzeIdentifier analyzes an identifier and returns its type
