package semantic

import (
	"testing"
)

// ============================================================================
// Task 7.139: External Class Semantic Tests
// ============================================================================

func TestExternalClassSemantics(t *testing.T) {
	t.Run("external class without parent is valid", func(t *testing.T) {
		input := `
type TExternal = class external
end;
`
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err)
		}
	})

	t.Run("external class with external parent is valid", func(t *testing.T) {
		input := `
type TExternalParent = class external
end;

type TExternalChild = class(TExternalParent) external
end;
`
		_, err := analyzeSource(t, input)
		if err != nil {
			t.Errorf("Expected no errors, got: %v", err)
		}
	})

	t.Run("external class cannot inherit from non-external class", func(t *testing.T) {
		input := `
type TRegular = class
end;

type TExternal = class(TRegular) external
end;
`
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for external class inheriting from non-external class")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if containsString(errMsg, "external") && containsString(errMsg, "inherit") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error message about external class inheritance, got: %v", errors)
		}
	})

	t.Run("non-external class can inherit from external class", func(t *testing.T) {
		input := `
type TExternal = class external
end;

type TRegular = class(TExternal)
end;
`
		analyzer, err := analyzeSource(t, input)
		if err == nil {
			t.Error("Expected error for non-external class inheriting from external class")
			return
		}

		errors := analyzer.Errors()
		found := false
		for _, errMsg := range errors {
			if containsString(errMsg, "external") && containsString(errMsg, "inherit") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error message about non-external class inheriting from external, got: %v", errors)
		}
	})
}
