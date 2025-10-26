package semantic

import (
	"fmt"
	"strings"
)

// AnalysisError represents one or more semantic analysis errors
type AnalysisError struct {
	Errors []string
}

// Error returns a formatted error message containing all semantic errors
func (e *AnalysisError) Error() string {
	if len(e.Errors) == 0 {
		return "semantic analysis failed"
	}

	if len(e.Errors) == 1 {
		return fmt.Sprintf("semantic error: %s", e.Errors[0])
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("semantic analysis failed with %d errors:\n", len(e.Errors)))
	for i, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err))
	}

	return sb.String()
}
