package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
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
		// Task 9.176: Typed record literals can be analyzed standalone
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
		// Task 9.160: Handle address-of expressions (@FunctionName)
		return a.analyzeAddressOfExpression(e)
	case *ast.LambdaExpression:
		// Task 9.216: Handle lambda expressions
		return a.analyzeLambdaExpression(e)
	case *ast.OldExpression:
		// Task 9.143: Handle 'old' expressions in postconditions
		return a.analyzeOldExpression(e)
	case *ast.InheritedExpression:
		// Task 9.161: Handle 'inherited' expressions
		return a.analyzeInheritedExpression(e)
	case *ast.IsExpression:
		// Task 9.40: Handle 'is' type checking operator
		return a.analyzeIsExpression(e)
	case *ast.AsExpression:
		// Task 9.48: Handle 'as' type casting operator
		return a.analyzeAsExpression(e)
	case *ast.ImplementsExpression:
		// Task 9.48: Handle 'implements' interface checking operator
		return a.analyzeImplementsExpression(e)
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// analyzeExpressionWithExpectedType analyzes an expression with optional expected type context.
// This enables context-sensitive type inference for expressions that benefit from knowing
// the expected type (e.g., lambda parameters, set/array literals, record literals).
//
// Task 9.19: Context-aware expression analysis infrastructure.
//
// Currently supported expression types with context inference:
//
//   - RecordLiteralExpression: Validates record literal fields against expected record type
//   - SetLiteral: Converts to ArrayLiteral when expected type is array (e.g., array of const)
//   - ArrayLiteralExpression: Converts to SetLiteral when expected type is set
//   - LambdaExpression: Infers parameter types from expected function pointer type (Task 9.19)
//   - NilLiteral: Returns the expected class/interface type instead of generic NIL (Task 9.19.5)
//   - IntegerLiteral: Returns FLOAT type when expected type is Float (Task 9.19.2)
//   - CallExpression: Passes expected type for future overload resolution (Task 9.19.2)
//
// For all other expression types, falls back to analyzeExpression() without context.
//
// Context-aware analysis is used in:
//   - Variable declarations: var x: T := <expr>  (expected type = T)
//   - Assignments: x := <expr>                    (expected type = type of x)
//   - Function arguments: f(<expr>)               (expected type = parameter type)
//   - Return statements: return <expr>            (expected type = function return type)
//   - Array elements: arr[i] := <expr>            (expected type = array element type)
//
// Parameters:
//   - expr: The expression to analyze
//   - expectedType: The expected type from context (may be nil if no context available)
//
// Returns:
//   - The actual type of the expression, or nil if analysis failed
func (a *Analyzer) analyzeExpressionWithExpectedType(expr ast.Expression, expectedType types.Type) types.Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.RecordLiteralExpression:
		return a.analyzeRecordLiteral(e, expectedType)
	case *ast.SetLiteral:
		// Task 9.156: Convert SetLiteral to ArrayLiteral when expected type is array of const
		// This fixes Format('%s', [varName]) where parser creates SetLiteral for [varName]
		if expectedType != nil {
			if _, ok := types.GetUnderlyingType(expectedType).(*types.ArrayType); ok {
				// If expected type is array (especially array of const), treat as array literal
				arrayLit := &ast.ArrayLiteralExpression{
					Token:    e.Token,
					Elements: e.Elements,
				}
				resultType := a.analyzeArrayLiteral(arrayLit, expectedType)

				// Set type annotation on the SetLiteral so interpreter knows to treat it as array
				if resultType != nil {
					e.Type = &ast.TypeAnnotation{
						Token: e.Token,
						Name:  resultType.String(),
					}
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
						case *ast.Identifier:
							// Identifiers are valid (enum values)
						case *ast.RangeExpression:
							// Ranges are valid (e.g., 1..10, 'a'..'z')
						case *ast.IntegerLiteral:
							// Task 9.226: Integer literals are valid
						case *ast.CharLiteral, *ast.StringLiteral:
							// Task 9.226: Character/string literals are valid
						case *ast.BooleanLiteral:
							// Task 9.226: Boolean literals are valid
						default:
							convertible = false
						}
					}
				}
				if convertible {
					setLit := &ast.SetLiteral{
						Token:    e.Token,
						Elements: e.Elements,
					}
					return a.analyzeSetLiteralWithContext(setLit, expectedType)
				}
			}
		}
		return a.analyzeArrayLiteral(e, expectedType)
	case *ast.LambdaExpression:
		// Task 9.19: Lambda parameter type inference from context
		// If expected type is a function pointer type, use it to infer parameter types
		if expectedType != nil {
			// Get underlying type to handle type aliases (e.g., type TFunc = function...)
			underlyingType := types.GetUnderlyingType(expectedType)
			if funcPtrType, ok := underlyingType.(*types.FunctionPointerType); ok {
				return a.analyzeLambdaExpressionWithContext(e, funcPtrType)
			}
		}
		// No expected type or not a function pointer - use regular analysis
		return a.analyzeLambdaExpression(e)
	case *ast.NilLiteral:
		// Task 9.19.5: Nil literal type inference from context
		// If expected type is a class or interface type, return that type instead of NIL
		// This makes nil more specific and helps with type checking
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			typeKind := underlyingType.TypeKind()
			if typeKind == "CLASS" || typeKind == "INTERFACE" {
				return expectedType
			}
		}
		// No expected type or not a class/interface - return generic NIL type
		return types.NIL
	case *ast.IntegerLiteral:
		// Task 9.19.2: Integer literal type inference from context
		// If expected type is Float, treat integer literal as float for better type compatibility
		if expectedType != nil {
			underlyingType := types.GetUnderlyingType(expectedType)
			if underlyingType.TypeKind() == "FLOAT" {
				return types.FLOAT
			}
		}
		// Default to INTEGER type
		return types.INTEGER
	case *ast.FloatLiteral:
		// Float literals are always FLOAT type regardless of context
		return types.FLOAT
	case *ast.CallExpression:
		// Task 9.19.2: Call expression with context for overload resolution
		// Pass expected type to help with overload resolution
		return a.analyzeCallExpressionWithContext(e, expectedType)
	default:
		return a.analyzeExpression(expr)
	}
}

// analyzeIsExpression analyzes the 'is' type checking operator (Task 9.40, 9.16.5.2).
// Example: obj is TMyClass -> Boolean
// Returns Boolean type.
func (a *Analyzer) analyzeIsExpression(expr *ast.IsExpression) types.Type {
	// Analyze the left expression (the object being checked)
	leftType := a.analyzeExpression(expr.Left)
	if leftType == nil {
		return nil
	}

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
	expr.SetType(&ast.TypeAnnotation{
		Token: expr.Token,
		Name:  "Boolean",
	})
	return types.BOOLEAN
}

// analyzeAsExpression analyzes the 'as' type casting operator (Task 9.48).
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

	// Target type can be either interface OR class
	targetUnderlying := types.GetUnderlyingType(targetType)
	interfaceType, isInterface := targetUnderlying.(*types.InterfaceType)
	classTargetType, isClassTarget := targetUnderlying.(*types.ClassType)

	if !isInterface && !isClassTarget {
		a.addError("'as' operator requires interface or class type, got %s at %s",
			targetType.String(), expr.Token.Pos.String())
		return nil
	}

	// Validate that left type is a class or object
	leftUnderlying := types.GetUnderlyingType(leftType)
	classType, isClass := leftUnderlying.(*types.ClassType)

	// Also allow NIL to be cast to any interface or class
	if leftType == types.NIL {
		// Set the expression type and return
		expr.SetType(&ast.TypeAnnotation{
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

		expr.SetType(&ast.TypeAnnotation{
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
	expr.SetType(&ast.TypeAnnotation{
		Token: expr.Token,
		Name:  interfaceType.Name,
	})

	return interfaceType
}

// analyzeImplementsExpression analyzes the 'implements' operator (Task 9.48).
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
	expr.SetType(&ast.TypeAnnotation{
		Token: expr.Token,
		Name:  "Boolean",
	})

	return types.BOOLEAN
}

// analyzeIdentifier analyzes an identifier and returns its type
