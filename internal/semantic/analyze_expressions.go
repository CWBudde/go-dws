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
	default:
		a.addError("unknown expression type: %T", expr)
		return nil
	}
}

// analyzeExpressionWithExpectedType analyzes an expression with optional expected type context.
// Used for literals that require context (records, sets, arrays) to resolve their types.
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
						case *ast.Identifier, *ast.RangeExpression:
							// valid set elements
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
	default:
		return a.analyzeExpression(expr)
	}
}

// analyzeIdentifier analyzes an identifier and returns its type
