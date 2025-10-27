package semantic

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Type Declaration Analysis
// ============================================================================

// evaluateConstantInt evaluates a compile-time constant integer expression.
// Returns the integer value and an error if the expression is not a constant.
// Task 9.98: Used for subrange bound evaluation.
func (a *Analyzer) evaluateConstantInt(expr ast.Expression) (int, error) {
	if expr == nil {
		return 0, fmt.Errorf("nil expression")
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		// Direct integer literal
		return int(e.Value), nil

	case *ast.UnaryExpression:
		// Handle negative numbers: -40
		if e.Operator == "-" {
			value, err := a.evaluateConstantInt(e.Right)
			if err != nil {
				return 0, err
			}
			return -value, nil
		}
		if e.Operator == "+" {
			// Unary plus: +5
			return a.evaluateConstantInt(e.Right)
		}
		return 0, fmt.Errorf("non-constant unary expression with operator %s", e.Operator)

	default:
		return 0, fmt.Errorf("expression is not a compile-time constant integer")
	}
}

// analyzeTypeDeclaration analyzes a type declaration statement
// Handles type aliases: type TUserID = Integer;
// Handles subrange types: type TDigit = 0..9; (Task 9.98)
func (a *Analyzer) analyzeTypeDeclaration(decl *ast.TypeDeclaration) {
	if decl == nil {
		return
	}

	// Check if type name already exists
	if _, err := a.resolveType(decl.Name.Value); err == nil {
		a.addError("type '%s' already declared at %s", decl.Name.Value, decl.Token.Pos.String())
		return
	}

	// Task 9.98: Handle subrange types
	if decl.IsSubrange {
		// Evaluate low bound (must be compile-time constant)
		lowBound, err := a.evaluateConstantInt(decl.LowBound)
		if err != nil {
			a.addError("subrange low bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}

		// Evaluate high bound (must be compile-time constant)
		highBound, err := a.evaluateConstantInt(decl.HighBound)
		if err != nil {
			a.addError("subrange high bound must be a compile-time constant integer at %s: %v",
				decl.Token.Pos.String(), err)
			return
		}

		// Task 9.98: Validate low <= high
		if lowBound > highBound {
			a.addError("subrange low bound (%d) cannot be greater than high bound (%d) at %s",
				lowBound, highBound, decl.Token.Pos.String())
			return
		}

		// Task 9.98: Create SubrangeType and register in type environment
		subrangeType := &types.SubrangeType{
			BaseType:  types.INTEGER, // Subranges are currently based on Integer
			Name:      decl.Name.Value,
			LowBound:  lowBound,
			HighBound: highBound,
		}

		a.subranges[decl.Name.Value] = subrangeType
		return
	}

	// Handle type aliases
	if decl.IsAlias {
		// Resolve the aliased type
		aliasedType, err := a.resolveType(decl.AliasedType.Name)
		if err != nil {
			a.addError("unknown type '%s' in type alias at %s", decl.AliasedType.Name, decl.Token.Pos.String())
			return
		}

		// Create TypeAlias and register it
		typeAlias := &types.TypeAlias{
			Name:        decl.Name.Value,
			AliasedType: aliasedType,
		}

		a.typeAliases[decl.Name.Value] = typeAlias
	}
}
