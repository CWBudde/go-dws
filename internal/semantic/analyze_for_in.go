package semantic

import (
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// analyzeForIn analyzes a for-in loop statement.
//
// Moved here unchanged when analyze_statements.go outgrew the file-length limit.
//
//nolint:gocyclo // pre-existing shape: one branch per enumerable kind, each owning its own diagnostic
func (a *Analyzer) analyzeForIn(stmt *ast.ForInStatement) {
	if stmt == nil {
		return
	}
	// Create a new scope for the loop variable
	oldSymbols := a.symbols
	a.symbols = NewEnclosedSymbolTable(oldSymbols)
	defer func() { a.symbols = oldSymbols }()
	defer a.emitUnusedWarningsForCurrentScope()

	var existingLoopVarType types.Type
	if !stmt.InlineVar {
		if sym, ok := oldSymbols.Resolve(stmt.Variable.Value); ok {
			existingLoopVarType = sym.Type
		}
	}

	// Analyze collection expression
	collectionType := a.analyzeExpression(stmt.Collection)
	if implicitType := a.getImplicitCallType(stmt.Collection); implicitType != nil {
		collectionType = implicitType
	}

	// Every for-in diagnostic is anchored at the `in`, the way `until` anchors a
	// repeat's condition.
	inPos := stmt.InPos
	if inPos.Line == 0 {
		inPos = stmt.Token.Pos
	}

	// Determine the element type and validate the collection is enumerable.
	// enumerable stays false when the collection cannot drive a loop at all, in
	// which case upstream abandons the loop and reports nothing further about it.
	var elementType types.Type
	enumerable := true
	iteratesString := false

	if collectionType != nil {
		switch ct := collectionType.(type) {
		case *types.ArrayType:
			// Arrays are enumerable, element type is the array's element type
			elementType = ct.ElementType

		case *types.SetType:
			// Sets are enumerable, element type is the set's element type
			elementType = ct.ElementType

		case *types.StringType:
			// Strings are enumerable. Existing Integer loop variables receive
			// character ordinals; inline/string loop variables receive characters.
			iteratesString = true
			if types.GetUnderlyingType(existingLoopVarType) == types.INTEGER {
				elementType = types.INTEGER
			} else {
				elementType = types.STRING
			}

		case *types.EnumType:
			// When iterating over an enum type directly (e.g., for var e in TColor do),
			// we iterate over all values of the enum type
			// The element type is the enum type itself
			elementType = ct

		case *types.TypeAlias:
			// Unwrap type alias and check the underlying type
			underlyingType := ct.AliasedType
			// Re-check with underlying type
			switch ut := underlyingType.(type) {
			case *types.ArrayType:
				elementType = ut.ElementType
			case *types.SetType:
				elementType = ut.ElementType
			case *types.StringType:
				iteratesString = true
				if types.GetUnderlyingType(existingLoopVarType) == types.INTEGER {
					elementType = types.INTEGER
				} else {
					elementType = types.STRING
				}
			case *types.EnumType:
				elementType = ut
			default:
				a.reportNotEnumerable(stmt.Collection, underlyingType, inPos, &enumerable)
				elementType = types.VOID
			}

		default:
			a.reportNotEnumerable(stmt.Collection, collectionType, inPos, &enumerable)
			elementType = types.VOID
		}
	} else {
		// Collection type could not be determined, use VOID
		elementType = types.VOID
	}

	// Define loop variable with the element type
	if !stmt.InlineVar {
		a.symbols.RecordUsage(stmt.Variable.Value, stmt.Variable.Token.Pos)
		// A collection whose element type could not be determined says nothing
		// about the loop variable; reporting on it would only cascade.
		if existingLoopVarType != nil && elementType != nil && !elementType.Equals(types.VOID) &&
			!a.forInAccepts(elementType, existingLoopVarType) {
			a.addStructuredError(NewIncompatibleTypesPairError(inPos,
				semanticTypeNameForDiagnostic(existingLoopVarType),
				semanticTypeNameForDiagnostic(elementType)))
		}
	}
	a.symbols.DefineLoopVariable(stmt.Variable.Value, elementType, stmt.Variable.Token.Pos)

	if stmt.Step != nil {
		stepType := a.analyzeExpression(stmt.Step)
		if implicitType := a.getImplicitCallType(stmt.Step); implicitType != nil {
			stepType = implicitType
		}
		if stepType != nil && types.GetUnderlyingType(stepType) != types.INTEGER {
			a.addError("for-in loop step must be Integer, got %s at %s",
				stepType.String(), stmt.Token.Pos.String())
		}
		if stepLiteral, ok := stmt.Step.(*ast.IntegerLiteral); ok && stepLiteral.Value <= 0 {
			a.addError("for-in loop step must be strictly positive, got %d at %s",
				stepLiteral.Value, stmt.Token.Pos.String())
		}
	}

	// Set loop context before analyzing body
	oldInLoop := a.inLoop
	a.inLoop = true
	a.loopDepth++
	defer func() {
		a.inLoop = oldInLoop
		a.loopDepth--
	}()

	// Analyze body. Iterating a string is a different loop upstream — character
	// by character rather than over a container — and it draws no empty-body
	// hint (`SimpleScripts/for_var_in_string`); neither does a loop that was
	// never built because the collection is not enumerable.
	if empty, ok := stmt.Body.(*ast.EmptyStatement); ok && enumerable && !iteratesString {
		a.addHintAt(empty.Token.Pos, "Empty FOR loop [line: %d, column: %d]",
			empty.Token.Pos.Line, empty.Token.Pos.Column)
	}
	a.analyzeStatement(stmt.Body)
}

// reportNotEnumerable reports a for-in collection that cannot be iterated.
// DWScript splits the case in two: naming a *type* that is not an enumeration
// is `Enumeration expected`, and the loop is still built, so the empty-body hint
// follows it (`for_in2`); an expression that is not a container is
// `Array expected`, and the loop is abandoned (`for_error3`).
func (a *Analyzer) reportNotEnumerable(collection ast.Expression, collectionType types.Type, pos lexer.Position, enumerable *bool) {
	if a.namesType(collection) {
		a.addStructuredError(NewEnumerationExpectedError(pos))
		return
	}
	name := "void"
	if collectionType != nil {
		name = collectionType.String()
	}
	a.addStructuredError(NewCannotIndexTypeError(pos, name))
	*enumerable = false
}

// namesType reports whether expr is a bare reference to a type rather than to a
// value — `for i in Integer do`, where `Integer` is not shadowed by a variable.
func (a *Analyzer) namesType(expr ast.Expression) bool {
	identExpr, ok := expr.(*ast.Identifier)
	if !ok {
		return false
	}
	if _, shadowed := a.symbols.Resolve(identExpr.Value); shadowed {
		return false
	}
	resolved, err := a.resolveType(identExpr.Value)
	return err == nil && resolved != nil
}

// warnForLoopVariableAssignment reports DWScript's warning for writing to a
// `for` control variable. The caller picks the anchor: upstream reports at the
// token its scanner happens to hold, which is the assignment target for an
// ordinary statement and the `:=` for a loop header that reuses the variable.
func (a *Analyzer) warnForLoopVariableAssignment(pos lexer.Position) {
	a.addWarning("Assignment to FOR-Loop variable [line: %d, column: %d]", pos.Line, pos.Column)
}

// forInAccepts reports whether a collection whose elements have type element can
// drive a for-in loop over a variable of type loopVar. It is stricter than
// assignment: the loop header converts nothing, so DWScript rejects the
// Integer-to-Float promotion an assignment would perform (`for_in1`).
func (a *Analyzer) forInAccepts(element, loopVar types.Type) bool {
	if types.GetUnderlyingType(element).TypeKind() == "INTEGER" &&
		types.GetUnderlyingType(loopVar).TypeKind() == "FLOAT" {
		return false
	}
	return a.canAssign(element, loopVar)
}
