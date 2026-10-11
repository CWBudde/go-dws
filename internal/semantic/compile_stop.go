package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Compiler stops.
//
// Upstream reports a compiler stop with AddCompilerStop, which raises
// ECompileError. It is caught only at the top of TdwsCompiler.Compile, so the
// compile is abandoned at the stop: nothing after it is read or reported, and
// the end-of-program checks (unimplemented forwards, incomplete classes) never
// run.
//
// The analyzer mirrors that with a sentinel panic. addCompilerStop records the
// stop diagnostic and raises compileStopSignal, which unwinds the statement
// under analysis. It is recovered at the units of work the analyzer runs in
// source order: the top-level statements of a program or unit section, and the
// deferred routine and method bodies. Work positioned after the stop is then
// skipped. Deferred bodies declared before it still run, because upstream
// compiles them before it reaches the stop.

// compileStopSignal is the panic value addCompilerStop raises.
type compileStopSignal struct{}

// addCompilerStop reports err as a compiler stop and abandons the analysis of
// the current statement. It does not return.
func (a *Analyzer) addCompilerStop(err *SemanticError) {
	err.Stop = true
	a.addStructuredError(err)
	a.raiseCompileStop(err)
}

// raiseCompileStop marks an already reported diagnostic as the compiler stop
// and abandons the analysis of the current statement. It does not return.
func (a *Analyzer) raiseCompileStop(err *SemanticError) {
	err.Stop = true
	a.stopped = true
	panic(compileStopSignal{})
}

// compileStopped reports whether the compile was abandoned at a stop, so the
// end-of-program checks must not run.
func (a *Analyzer) compileStopped() bool {
	return a.stopped || a.skipEndOfProgramChecks
}

// analyzeUntilStop runs analyze and recovers a compiler stop raised inside it,
// restoring the analysis context the unwinding skipped. It reports whether a
// stop was raised. Any other panic propagates.
func (a *Analyzer) analyzeUntilStop(analyze func()) (stopped bool) {
	saved := a.saveStopContext()
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(compileStopSignal); !ok {
				panic(recovered)
			}
			saved.restore(a)
			stopped = true
		}
	}()
	analyze()
	return false
}

// analyzeTopLevelStatement analyzes one top-level statement and reports
// whether it raised a compiler stop.
func (a *Analyzer) analyzeTopLevelStatement(stmt ast.Statement) bool {
	return a.analyzeUntilStop(func() { a.analyzeStatement(stmt) })
}

// insertionPrecedesStop reports whether deferred work registered at point lies
// before the earliest compiler stop in source order, so upstream compiled it
// before abandoning the compile. Without a stop, everything does.
func (a *Analyzer) insertionPrecedesStop(point *diagnosticInsertion) bool {
	if !a.stopped {
		return true
	}
	for i, err := range a.structuredErrors {
		if err != nil && err.Stop {
			return point.structuredAt <= i
		}
	}
	return true
}

// stopContext is the analysis context a compiler stop may leave half set up
// when it unwinds past the code that would have restored it.
type stopContext struct {
	symbols                  *SymbolTable
	currentFunction          *ast.FunctionDecl
	currentClass             *types.ClassType
	currentRecord            *types.RecordType
	currentHelperType        *types.HelperType
	currentSelfType          types.Type
	mainStatement            ast.Statement
	indexedWriteTargetMember ast.Expression
	currentProperty          string
	deferredBody             deferredBodyErrorBounds
	loopPosStack             int
	loopExitabilityStack     int
	loopDepth                int
	inLoop                   bool
	inLambda                 bool
	inClassMethod            bool
	inStaticHelperMethod     bool
	inPropertyExpr           bool
	inIndexBase              bool
	inArrayHelperCallback    bool
	inArrayAssignment        bool
	inFinallyBlock           bool
	inExceptionHandler       bool
	reservedRoutineResult    bool
}

func (a *Analyzer) saveStopContext() stopContext {
	return stopContext{
		symbols:                  a.symbols,
		currentFunction:          a.currentFunction,
		currentClass:             a.currentClass,
		currentRecord:            a.currentRecord,
		currentHelperType:        a.currentHelperType,
		currentSelfType:          a.currentSelfType,
		currentProperty:          a.currentProperty,
		mainStatement:            a.mainStatement,
		deferredBody:             a.deferredBody,
		loopPosStack:             len(a.loopPosStack),
		loopExitabilityStack:     len(a.loopExitabilityStack),
		loopDepth:                a.loopDepth,
		inLoop:                   a.inLoop,
		inLambda:                 a.inLambda,
		inClassMethod:            a.inClassMethod,
		inStaticHelperMethod:     a.inStaticHelperMethod,
		inPropertyExpr:           a.inPropertyExpr,
		inIndexBase:              a.inIndexBase,
		inArrayHelperCallback:    a.inArrayHelperCallback,
		inArrayAssignment:        a.inArrayAssignment,
		inFinallyBlock:           a.inFinallyBlock,
		inExceptionHandler:       a.inExceptionHandler,
		reservedRoutineResult:    a.reservedRoutineResult,
		indexedWriteTargetMember: a.indexedWriteTargetMember,
	}
}

func (c stopContext) restore(a *Analyzer) {
	a.symbols = c.symbols
	a.currentFunction = c.currentFunction
	a.currentClass = c.currentClass
	a.currentRecord = c.currentRecord
	a.currentHelperType = c.currentHelperType
	a.currentSelfType = c.currentSelfType
	a.currentProperty = c.currentProperty
	a.mainStatement = c.mainStatement
	a.deferredBody = c.deferredBody
	a.loopPosStack = a.loopPosStack[:min(c.loopPosStack, len(a.loopPosStack))]
	a.loopExitabilityStack = a.loopExitabilityStack[:min(c.loopExitabilityStack, len(a.loopExitabilityStack))]
	a.loopDepth = c.loopDepth
	a.inLoop = c.inLoop
	a.inLambda = c.inLambda
	a.inClassMethod = c.inClassMethod
	a.inStaticHelperMethod = c.inStaticHelperMethod
	a.inPropertyExpr = c.inPropertyExpr
	a.inIndexBase = c.inIndexBase
	a.inArrayHelperCallback = c.inArrayHelperCallback
	a.inArrayAssignment = c.inArrayAssignment
	a.inFinallyBlock = c.inFinallyBlock
	a.inExceptionHandler = c.inExceptionHandler
	a.reservedRoutineResult = c.reservedRoutineResult
	a.indexedWriteTargetMember = c.indexedWriteTargetMember
}

// recoverCompileStop ends an analysis entry point at a compiler stop that no
// inner unit of work recovered, reporting the diagnostics collected so far.
// It must be deferred directly.
func (a *Analyzer) recoverCompileStop(err *error) {
	recovered := recover()
	if recovered == nil {
		return
	}
	if _, ok := recovered.(compileStopSignal); !ok {
		panic(recovered)
	}
	*err = &AnalysisError{Errors: a.errors}
}
