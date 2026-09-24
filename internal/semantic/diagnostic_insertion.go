package semantic

// diagnosticInsertion remembers a deferred body's declaration in both diagnostic
// streams. Its ordinal disambiguates bodies declared before any diagnostics have
// been emitted, so equal offsets still retain declaration order.
type diagnosticInsertion struct {
	errorsAt     int
	structuredAt int
	ordinal      int
}

func (a *Analyzer) newDiagnosticInsertion() *diagnosticInsertion {
	point := &diagnosticInsertion{
		errorsAt:     len(a.errors),
		structuredAt: len(a.structuredErrors),
		ordinal:      len(a.diagnosticInsertions),
	}
	a.diagnosticInsertions = append(a.diagnosticInsertions, point)
	return point
}

func (a *Analyzer) analyzeAtDiagnosticInsertion(point *diagnosticInsertion, analyze func()) {
	errorsBefore, structuredBefore := len(a.errors), len(a.structuredErrors)
	previousBounds := a.deferredBody
	a.deferredBody = deferredBodyErrorBounds{
		beforeDecl: point.errorsAt,
		bodyStart:  errorsBefore,
		active:     true,
	}
	analyze()
	a.deferredBody = previousBounds
	errorsAdded := len(a.errors) - errorsBefore
	structuredAdded := len(a.structuredErrors) - structuredBefore
	a.errors = spliceBack(a.errors, errorsBefore, point.errorsAt)
	a.structuredErrors = spliceBack(a.structuredErrors, structuredBefore, point.structuredAt)
	for _, later := range a.diagnosticInsertions {
		if later == point {
			continue
		}
		if later.errorsAt > point.errorsAt || (later.errorsAt == point.errorsAt && later.ordinal > point.ordinal) {
			later.errorsAt += errorsAdded
		}
		if later.structuredAt > point.structuredAt || (later.structuredAt == point.structuredAt && later.ordinal > point.ordinal) {
			later.structuredAt += structuredAdded
		}
	}
}
