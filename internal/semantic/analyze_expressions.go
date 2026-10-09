package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// analyzeExpression analyzes an expression and returns its type.
// Returns nil if the expression is invalid.
func (a *Analyzer) analyzeExpression(expr ast.Expression) (resolvedType types.Type) {
	defer func() { a.semanticInfo.SetResolvedType(expr, resolvedType) }()
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.InvalidExpression:
		return nil
	case *ast.DebugBreakExpression:
		return a.analyzeDebugBreak(e, false)
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
		if result, handled := a.analyzeBareBoundHelper(e, nil); handled {
			return result
		}
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
	case *ast.AnonymousRecordExpression:
		// Structurally typed (record a := 1; end): needs no context, unlike
		// an anonymous RecordLiteralExpression.
		return a.analyzeAnonymousRecordExpression(e)
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
	if t.Equals(types.BOOLEAN) || t.Equals(types.VARIANT) || types.IsJSONVariant(t) {
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
func (a *Analyzer) analyzeExpressionWithExpectedType(expr ast.Expression, expectedType types.Type) (resolvedType types.Type) {
	defer func() { a.semanticInfo.SetResolvedType(expr, resolvedType) }()
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
					BaseNode:         e.BaseNode,
					Elements:         e.Elements,
					ElementPositions: e.ElementPositions,
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
			// A bracket literal in a set-typed context is a set constructor,
			// whatever shape its elements have. The parser's heuristic
			// (shouldParseAsSetLiteral) only sees syntax, so a typecast element
			// such as `[TMyEnum(3)]` arrives here as an array literal; the
			// expected type is the authority. analyzeSetLiteralWithContext owns
			// the per-element ordinal, element-type and bounds diagnostics.
			if _, ok := types.GetUnderlyingType(expectedType).(*types.SetType); ok {
				setLit := &ast.SetLiteral{
					BaseNode: e.BaseNode,
					Elements: e.Elements,
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
	case *ast.MemberAccessExpression:
		if a.isDefaultNamespace(e.Object) {
			return a.analyzeDefaultNamespaceMember(e, expectedType, false)
		}
		// In a function-pointer target context, a method reference `obj.Method`
		// becomes a bound method pointer rather than an auto-invoked call.
		if expectedType != nil {
			expectedKind := types.GetUnderlyingType(expectedType).TypeKind()
			if expectedKind == "FUNCTION_POINTER" || expectedKind == "METHOD_POINTER" {
				if ptrType, ok := a.analyzeMethodReferenceInPointerContext(e, expectedType); ok {
					return ptrType
				}
			}
		}
		return a.analyzeMemberAccessWithExpectedType(e, expectedType)
	case *ast.AddressOfExpression:
		// `@Test` is a routine reference like a bare name, and upstream subjects
		// it to the same rule: where the reference does not fit the expected
		// pointer type, the call reading stands and a routine with required
		// parameters is short of arguments. The expected type reaches no other
		// part of the address-of analysis, which resolves the operand on its own.
		if expectedType != nil && !types.IsPointerType(expectedType) && !expectedType.Equals(types.VARIANT) {
			a.addError("unexpected \"@\" at %s", e.Token.Pos.String())
			resultType := a.analyzeExpression(e)
			if resultType != nil {
				err := NewIncompatibleTypesPairError(e.Token.Pos,
					semanticTypeNameForDiagnostic(expectedType), semanticTypeNameForDiagnostic(resultType))
				err.AfterChildren = true
				a.addStructuredError(err)
			}
			// ReadAt replaces the invalid operand with a nil constant so the
			// enclosing argument checker reports its own mismatch and continues.
			return types.NIL
		}
		resultType := a.analyzeExpression(e)
		a.checkPointerContextArity(e, resultType, expectedType)
		return resultType
	case *ast.CallExpression:
		// The expected type deliberately plays no part here. Return type is not
		// part of overload identity in DWScript (see types.SignaturesEqual), so
		// it could only serve as a last-resort tie-break between candidates that
		// already score equally on argument distance. Resolving such a call in
		// the analyzer would admit programs the AST evaluator then executes with
		// a different overload: the evaluator resolves overloads independently
		// at run time (see evaluator.ResolveOverloadMultiple) and has no channel
		// for a call site's expected type. Wiring that channel — or recording the
		// analyzer's chosen overload in ast.SemanticInfo for the evaluator to
		// reuse — has to come first.
		return a.analyzeCallExpression(e)
	case *ast.GroupedExpression:
		result := a.analyzeExpression(e.Expression)
		// Grouped helper reads lose the outer expected type in DWScript. Other
		// routine references retain the behavior from before grouping was preserved.
		if result != nil && expectedType != nil && types.IsPointerType(expectedType) && !a.isGroupedHelperValue(e) {
			return a.analyzeExpressionWithExpectedType(e.Expression, expectedType)
		}
		return result
	case *ast.Identifier:
		if result, handled := a.analyzeBareBoundHelper(e, expectedType); handled {
			return result
		}
		// In contexts like `x := GetValue;`, DWScript auto-invokes a
		// parameterless function when the expected type matches its return type.
		// Keep function-pointer assignment behavior when the expected type is
		// itself a function/method pointer.
		if expectedType != nil {
			expectedUnderlying := types.GetUnderlyingType(expectedType)
			expectedKind := expectedUnderlying.TypeKind()
			if expectedKind != "FUNCTION_POINTER" && expectedKind != "METHOD_POINTER" {
				if implicitType := a.getImplicitCallType(e); implicitType != nil && a.canAssign(implicitType, expectedType) {
					return implicitType
				}
				// The context wants a value, and a routine reference is not one.
				// Upstream reads the name as a call here too, so a routine with
				// required parameters is short of them — `Test([Test])` against
				// `array of Integer` reports it for the inner name
				// (const_procedure_array).
				resultType := a.analyzeIdentifier(e)
				a.checkPointerContextArity(e, resultType, expectedType)
				return resultType
			} else {
				// Function-pointer context: a bare routine name resolves to a
				// pointer type in analyzeIdentifier, but the node is only annotated
				// for builtins there. Annotate user-routine references too, so the
				// evaluator produces a FunctionPointerValue instead of auto-invoking
				// a parameterless routine (func_ptr1, func_ptr_var_param).
				resultType := a.analyzeIdentifier(e)
				a.checkPointerContextArity(e, resultType, expectedType)
				if a.semanticInfo != nil && resultType != nil && a.semanticInfo.GetType(e) == nil {
					resultUnderlying := types.GetUnderlyingType(resultType)
					if rk := resultUnderlying.TypeKind(); rk == "FUNCTION_POINTER" || rk == "METHOD_POINTER" {
						a.semanticInfo.SetType(e, &ast.TypeAnnotation{
							Token: e.Token,
							Name:  resultUnderlying.String(),
						})
					}
				}
				return resultType
			}
		}
		return a.analyzeIdentifier(e)
	default:
		return a.analyzeExpression(expr)
	}
}

// analyzeCastTarget distinguishes a lexical value from a type name. Class names
// are class references, while variables keep their declared type.
func (a *Analyzer) analyzeCastTarget(target ast.TypeExpression, right ast.Expression) (result types.Type) {
	firstDiagnostic := len(a.structuredErrors)
	defer func() {
		// Unknown names while reading a cast/check target stop upstream's term
		// reader. Preserve that stop in both analyzer state and the emitted stream.
		if target != nil {
			a.semanticInfo.SetResolvedType(target, result)
		}
		if right != nil {
			a.semanticInfo.SetResolvedType(right, result)
		}
		for _, diagnostic := range a.structuredErrors[firstDiagnostic:] {
			if diagnostic.Type == ErrorGeneric && strings.HasPrefix(diagnostic.Message, "Unknown name ") {
				a.raiseCompileStop(diagnostic)
			}
		}
	}()
	if right != nil {
		if target != nil && !a.hasLexicalValueReceiver(right) {
			if resolved := a.resolveCastTargetType(target); resolved != nil {
				return resolved
			}
		}
		return a.analyzeCastTargetValue(right)
	}
	if annotation, ok := target.(*ast.TypeAnnotation); ok {
		id := &ast.Identifier{BaseNode: ast.BaseNode{Token: annotation.Token}, Value: annotation.Name}
		if sym, found := a.symbols.resolveIdentity(annotation.Name, true); found && !sym.lookupOnly {
			return a.analyzeCastTargetValue(id)
		}
		if resolved := a.resolveCastTargetType(target); resolved != nil {
			return resolved
		}
		return a.analyzeIdentifier(id)
	}
	resolved, err := a.resolveTypeExpression(target)
	if err != nil {
		return nil
	}
	return resolved
}

// resolveCastTargetType preserves type-name hints and represents class names as
// class references. An unresolved name falls back to value analysis at its caller.
func (a *Analyzer) resolveCastTargetType(target ast.TypeExpression) types.Type {
	resolved, err := a.resolveTypeExpression(target)
	if err != nil || resolved == nil {
		return nil
	}
	if annotation, ok := target.(*ast.TypeAnnotation); ok {
		if symbol, found := a.symbols.resolveIdentity(annotation.Name, true); found && symbol.lookupOnly {
			a.addIdentifierCaseHint(&ast.Identifier{BaseNode: ast.BaseNode{Token: annotation.Token}, Value: annotation.Name}, symbol.Name)
		}
	}
	if class, ok := types.GetUnderlyingType(resolved).(*types.ClassType); ok {
		return types.NewClassOfType(class)
	}
	if meta := types.NewRecordMetaType(resolved); meta != nil {
		if annotation, ok := target.(*ast.TypeAnnotation); ok && !strings.Contains(annotation.Name, ".") {
			if _, found := a.symbols.resolveIdentity(annotation.Name, true); !found {
				return nil
			}
		}
		return meta
	}
	return resolved
}

// hasLexicalValueReceiver prevents qualified values from being read as
// namespace/type names, including locals shadowing System and imported units.
func (a *Analyzer) hasLexicalValueReceiver(expr ast.Expression) bool {
	for {
		switch node := expr.(type) {
		case *ast.MemberAccessExpression:
			expr = node.Object
		case *ast.GroupedExpression:
			expr = node.Expression
		case *ast.Identifier:
			symbol, found := a.symbols.resolveIdentity(node.Value, true)
			return found && !symbol.lookupOnly && !symbol.IsEnumTypeName
		default:
			return true
		}
	}
}

// analyzeCastTargetValue uses the shared one-call reading for bare factories.
// The evaluator discards AS targets, so this metadata never executes the call.
func (a *Analyzer) analyzeCastTargetValue(expr ast.Expression) types.Type {
	if grouped, ok := expr.(*ast.GroupedExpression); ok {
		result := a.analyzeCastTargetValue(grouped.Expression)
		a.semanticInfo.SetResolvedType(expr, result)
		return result
	}
	if identifier, ok := expr.(*ast.Identifier); ok {
		if result, called := a.analyzeAssignmentIdentifierCall(identifier, types.NewClassOfType(a.getClassType("TObject")), true); called {
			return result
		}
	}
	return a.analyzeExpression(expr)
}

// analyzeIsExpression validates the target category while retaining Boolean recovery.
func (a *Analyzer) analyzeIsExpression(expr *ast.IsExpression) types.Type {
	left := a.analyzeExpression(expr.Left)
	// Boolean value targets execute at runtime. Retain the actual analyzed node
	// so implicit-call intent, bindings, and usage metadata keep their identity.
	if expr.Right == nil {
		if annotation, ok := expr.TargetType.(*ast.TypeAnnotation); ok {
			expr.Right = &ast.Identifier{BaseNode: ast.BaseNode{Token: annotation.Token, EndPos: annotation.EndPos}, Value: annotation.Name}
		}
	}
	right := a.analyzeCastTarget(expr.TargetType, expr.Right)
	if left == nil || right == nil {
		return nil
	}
	if (types.GetUnderlyingType(left).Equals(types.BOOLEAN) || types.GetUnderlyingType(left).Equals(types.VARIANT)) && types.GetUnderlyingType(right).Equals(types.BOOLEAN) {
		return a.annotateCastResult(expr, types.BOOLEAN)
	}
	switch types.GetUnderlyingType(left).(type) {
	case *types.ClassType, *types.InterfaceType:
	default:
		if left != types.NIL {
			a.addError("Object expected at %s", expr.Token.Pos.String())
			return a.annotateCastResult(expr, types.BOOLEAN)
		}
	}
	switch types.GetUnderlyingType(right).(type) {
	case *types.ClassOfType, *types.InterfaceType:
	default:
		a.addError("Class reference expected at %s", expr.Token.Pos.String())
	}
	return a.annotateCastResult(expr, types.BOOLEAN)
}

func (a *Analyzer) annotateCastResult(expr ast.Expression, result types.Type) types.Type {
	name := result.String()
	switch target := result.(type) {
	case *types.ClassType:
		name = target.Name
	case *types.InterfaceType:
		name = target.Name
	}
	a.semanticInfo.SetType(expr, &ast.TypeAnnotation{Name: name})
	a.semanticInfo.SetResolvedType(expr, result)
	return result
}

// analyzeAsExpression selects the caster from the source and declared RHS types.
// Invalid object targets recover TObject; invalid metaclass targets stop compilation.
func (a *Analyzer) analyzeAsExpression(expr *ast.AsExpression) types.Type {
	left := a.analyzeExpression(expr.Left)
	target := a.analyzeCastTarget(expr.TargetType, expr.Right)
	if left == nil || target == nil {
		return nil
	}
	meta, isMeta := types.GetUnderlyingType(target).(*types.ClassOfType)
	_, isInterface := types.GetUnderlyingType(target).(*types.InterfaceType)
	result := target
	switch source := types.GetUnderlyingType(left).(type) {
	case *types.ClassType, *types.InterfaceType:
		if isMeta {
			result = meta.ClassType
			if class, ok := source.(*types.ClassType); ok && !types.IsClassRelated(class, meta.ClassType) {
				a.addStructuredError(NewIncompatibleTypesPairError(expr.Token.Pos, class.Name, meta.ClassType.Name))
			}
		} else if !isInterface {
			a.addError("Class reference expected at %s", expr.Token.Pos.String())
			result = a.getClassType("TObject")
		}
	case *types.ClassOfType:
		if !isMeta {
			a.addCompilerStop(&SemanticError{Type: ErrorInvalidOperation, Message: "Class reference expected", Pos: expr.Token.Pos, Severity: SeverityError})
			return nil
		}
		if !types.IsClassRelated(source.ClassType, meta.ClassType) {
			a.addStructuredError(NewIncompatibleTypesPairError(expr.Token.Pos, source.ClassType.Name, meta.ClassType.Name))
		}
	default:
		if left == types.NIL {
			if isMeta {
				result = meta.ClassType
			}
		} else if !left.Equals(types.VARIANT) {
			a.addError("Cannot cast %q as %q at %s", left.String(), target.String(), expr.Token.Pos.String())
		}
	}
	return a.annotateCastResult(expr, result)
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

	// Instances, class names and metaclass variables all support implementation
	// checks. The evaluator resolves their runtime class metadata.
	if leftType != types.NIL {
		leftUnderlying := types.GetUnderlyingType(leftType)
		switch leftUnderlying.(type) {
		case *types.ClassType, *types.ClassOfType:
		default:
			a.addError("'implements' operator requires class instance or class reference, got %s at %s",
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
		a.addBooleanExpected(expr.Token.Pos)
	}

	// Analyze consequence expression
	consequenceType := a.analyzeExpression(expr.Consequence)
	if consequenceType == nil {
		if containsParserRecovery(expr.Consequence) {
			// The parser already stopped with "Expression expected" here.
			return nil
		}
		a.addError("invalid consequence expression in if-then-else at %s", expr.Token.Pos.String())
		return nil
	}

	var resultType types.Type

	// Analyze alternative if present
	if expr.Alternative != nil {
		// A bracket literal branch adopts the other branch's type
		// (e.g. `if c then [two] else []` or `if c then [] else s` yields a
		// set, not an array), whichever side the literal is on.
		var alternativeType types.Type
		if isBracketLiteral(expr.Alternative) {
			alternativeType = a.analyzeExpressionWithExpectedType(expr.Alternative, consequenceType)
		} else {
			alternativeType = a.analyzeExpression(expr.Alternative)
			if alternativeType != nil && isBracketLiteral(expr.Consequence) {
				if reTyped := a.analyzeExpressionWithExpectedType(expr.Consequence, alternativeType); reTyped != nil {
					consequenceType = reTyped
				}
			}
		}
		if alternativeType == nil {
			if containsParserRecovery(expr.Alternative) {
				// The parser already stopped with "Expression expected" here.
				return nil
			}
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
