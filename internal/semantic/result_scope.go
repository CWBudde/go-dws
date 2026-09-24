package semantic

import (
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// enterResultScope records the nearest routine's implicit result. Upstream's
// TdwsProcedure.FindLocal checks InternalParams even inside nested block tables;
// a nested procedure, however, has its own internal parameters.
func (a *Analyzer) enterResultScope(hasResult bool) func() {
	previous := a.reservedRoutineResult
	a.reservedRoutineResult = hasResult
	return func() { a.reservedRoutineResult = previous }
}

func (a *Analyzer) rejectResultDeclaration(name *ast.Identifier) bool {
	if !a.reservedRoutineResult || name == nil || !ident.Equal(name.Value, "Result") {
		return false
	}
	pos := name.Token.Pos
	a.addError("%s", errors.FormatNameAlreadyExists(name.Value, pos.Line, pos.Column))
	return true
}

// lambdaBodyReturnsValue identifies the return syntax used by lambda inference
// before analyzing its expressions, which may themselves contain name errors.
func lambdaBodyReturnsValue(body *ast.BlockStatement) bool {
	if body == nil {
		return false
	}
	returnsValue := false
	ast.Inspect(body, func(node ast.Node) bool {
		if returnsValue {
			return false
		}
		switch n := node.(type) {
		case *ast.FunctionDecl, *ast.LambdaExpression:
			return false
		case *ast.ReturnStatement:
			returnsValue = n.ReturnValue != nil
		case *ast.AssignmentStatement:
			if target, ok := n.Target.(*ast.Identifier); ok {
				returnsValue = ident.Equal(target.Value, "Result")
			}
		}
		return !returnsValue
	})
	return returnsValue
}

// checkInferredLambdaResultDeclarations validates declarations skipped by the
// return-inference walk, once syntax or context establishes that Result exists.
// Other routine bodies own their own Result reservation.
func (a *Analyzer) checkInferredLambdaResultDeclarations(body *ast.BlockStatement) {
	if body == nil || !a.reservedRoutineResult {
		return
	}
	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FunctionDecl, *ast.LambdaExpression:
			return false
		case *ast.VarDeclStatement:
			for _, name := range n.Names {
				if a.rejectResultDeclaration(name) {
					break
				}
			}
		case *ast.ForStatement:
			if n.InlineVar {
				a.rejectResultDeclaration(n.Variable)
			}
		case *ast.ForInStatement:
			if n.InlineVar {
				a.rejectResultDeclaration(n.Variable)
			}
		}
		return true
	})
}
