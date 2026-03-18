package semantic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
	"github.com/cwbudde/go-dws/pkg/token"
)

type unusedSymbolCandidate struct {
	name string
	sym  *Symbol
	pos  token.Position
}

func blockEndStart(pos token.Position) token.Position {
	if pos.Column > 4 {
		pos.Column -= 4
	}
	if pos.Offset > 4 {
		pos.Offset -= 4
	}
	return pos
}

// recordSymbolUsage marks a symbol as used in the current or outer scope.
func (a *Analyzer) recordSymbolUsage(name string, pos token.Position) {
	if a == nil || a.symbols == nil {
		return
	}
	a.symbols.RecordUsage(name, pos)
}

// emitUnusedWarningsForCurrentScope emits DWScript-style warnings for
// unused locals in the current scope. It intentionally skips parameters,
// constants, read-only bindings, and injected symbols such as Self.
func (a *Analyzer) emitUnusedWarningsForCurrentScope() {
	if a == nil || a.symbols == nil {
		return
	}
	if a.hintsLevel < HintsLevelPedantic {
		return
	}
	if a.currentFunction == nil && !a.inLambda {
		return
	}
	if a.currentFunction != nil && a.currentFunction.Body == nil {
		return
	}

	candidates := make([]unusedSymbolCandidate, 0)
	a.symbols.symbols.Range(func(name string, sym *Symbol) bool {
		if sym == nil {
			return true
		}
		if sym.SuppressUnusedWarning || sym.IsConst || sym.ReadOnly {
			return true
		}
		if sym.Name == "" || sym.DeclPosition.Line == 0 || sym.DeclPosition.Column == 0 {
			return true
		}
		if ident.Equal(sym.Name, "Self") {
			return true
		}
		if a.currentFunction != nil && ident.Equal(sym.Name, a.currentFunction.Name.Value) {
			return true
		}
		if len(sym.Usages) > 0 {
			return true
		}

		candidates = append(candidates, unusedSymbolCandidate{
			name: name,
			sym:  sym,
			pos:  sym.DeclPosition,
		})
		return true
	})

	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i].pos
		right := candidates[j].pos
		if left.Line != right.Line {
			return left.Line > right.Line
		}
		if left.Column != right.Column {
			return left.Column > right.Column
		}
		return candidates[i].name > candidates[j].name
	})

	for _, candidate := range candidates {
		if ident.Equal(candidate.name, "Result") {
			a.addHint("Result is never used [line: %d, column: %d]",
				candidate.pos.Line, candidate.pos.Column)
			continue
		}
		a.addHint("Variable \"%s\" declared but not used [line: %d, column: %d]",
			candidate.sym.Name, candidate.pos.Line, candidate.pos.Column)
	}
}

func (a *Analyzer) recordClassFieldUsage(classType *types.ClassType, name string) {
	if classType == nil {
		return
	}
	classType.MarkFieldUsed(name)
}

func (a *Analyzer) recordClassMethodUsage(classType *types.ClassType, name string) {
	if classType == nil {
		return
	}
	classType.MarkMethodUsed(name)
}

func (a *Analyzer) queueUnusedPrivateClassMembers(classType *types.ClassType) {
	if a == nil || classType == nil {
		return
	}
	if a.pendingClassWarnings == nil {
		a.pendingClassWarnings = make([]*types.ClassType, 0)
	}
	a.pendingClassWarnings = append(a.pendingClassWarnings, classType)
}

func (a *Analyzer) collectUnusedPrivateClassMemberWarnings(classType *types.ClassType) []string {
	if a == nil || classType == nil {
		return nil
	}
	if a.hintsLevel < HintsLevelPedantic {
		return nil
	}
	for _, err := range a.errors {
		if !strings.HasPrefix(err, "Hint:") && !strings.HasPrefix(err, "Warning:") {
			return nil
		}
	}

	type memberWarning struct {
		pos     token.Position
		message string
	}

	warnings := make([]memberWarning, 0)

	for name, vis := range classType.FieldVisibility {
		if vis != int(ast.VisibilityPrivate) {
			continue
		}
		if classType.FieldUsed(name) {
			continue
		}
		pos, ok := classType.FieldDeclPositions[name]
		if !ok || pos.Line == 0 || pos.Column == 0 {
			continue
		}
		declName := classType.FieldDeclNames[name]
		if declName == "" {
			declName = name
		}
		warnings = append(warnings, memberWarning{
			pos:     pos,
			message: "Private field \"" + declName + "\" declared but never used",
		})
	}

	for name, vis := range classType.MethodVisibility {
		if vis != int(ast.VisibilityPrivate) {
			continue
		}
		if classType.MethodUsed(name) {
			continue
		}
		pos, ok := classType.MethodDeclPositions[name]
		if !ok || pos.Line == 0 || pos.Column == 0 {
			continue
		}
		declName := classType.MethodDeclNames[name]
		if declName == "" {
			declName = name
		}
		warnings = append(warnings, memberWarning{
			pos:     pos,
			message: "Private method \"" + declName + "\" declared but never used",
		})
	}

	sort.SliceStable(warnings, func(i, j int) bool {
		left := warnings[i].pos
		right := warnings[j].pos
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		return warnings[i].message < warnings[j].message
	})

	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		out = append(out, fmt.Sprintf("Hint: %s [line: %d, column: %d]",
			warning.message, warning.pos.Line, warning.pos.Column))
	}

	return out
}
