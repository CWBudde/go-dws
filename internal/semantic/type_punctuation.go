package semantic

import (
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/token"
)

func identifierLookaheadPos(identifier *ast.Identifier) token.Position {
	if identifier.NextTokenPos.IsValid() {
		return identifier.NextTokenPos
	}
	return identifier.End()
}

func (a *Analyzer) addPunctuationStop(pos token.Position, message string) {
	diagnostic := NewGenericError(pos, message)
	a.addCompilerStop(diagnostic)
}

func (a *Analyzer) addUnexpectedAddressOf(pos token.Position) {
	diagnostic := NewGenericError(pos, `unexpected "@"`)
	diagnostic.AfterChildren = true
	a.addStructuredError(diagnostic)
}
