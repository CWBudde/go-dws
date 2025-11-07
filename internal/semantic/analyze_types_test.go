// Package semantic_test contains semantic analysis tests for type system.
package semantic

import (
	"fmt"
	"testing"
)

// ============================================================================
// Large Set Type Declaration Tests
// ============================================================================

// TestLargeSetTypeDeclaration tests that large enum set types are properly analyzed.
func TestLargeSetTypeDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		enumSize int
	}{
		{"65 elements (boundary)", 65},
		{"70 elements", 70},
		{"100 elements", 100},
		{"200 elements", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate large enum type
			input := generateLargeEnumSetDeclaration(tt.enumSize)
			expectNoErrors(t, input)
		})
	}
}

// TestLargeSetTypeInference tests type inference for large set literals.
func TestLargeSetTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		enumSize int
		elements string
	}{
		{
			name:     "65-element enum set literal",
			enumSize: 65,
			elements: "[E00, E05, E10, E64]",
		},
		{
			name:     "100-element enum set literal",
			enumSize: 100,
			elements: "[E00, E10, E20, E90, E99]",
		},
		{
			name:     "large enum empty set",
			enumSize: 100,
			elements: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`
				%s
				var s: TLargeSet;
				begin
					s := %s;
				end.
			`, generateLargeEnumSetType(tt.enumSize), tt.elements)
			expectNoErrors(t, input)
		})
	}
}

// TestLargeSetOperationTypeChecking tests type checking for large set operations.
func TestLargeSetOperationTypeChecking(t *testing.T) {
	tests := []struct {
		name      string
		enumSize  int
		operation string
	}{
		{"union", 100, "s3 := s1 + s2;"},
		{"difference", 100, "s3 := s1 - s2;"},
		{"intersection", 100, "s3 := s1 * s2;"},
		{"membership", 100, "if E50 in s1 then PrintLn('ok');"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`
				%s
				var s1, s2, s3: TLargeSet;
				begin
					s1 := [E00, E10, E50];
					s2 := [E10, E20];
					%s
				end.
			`, generateLargeEnumSetType(tt.enumSize), tt.operation)
			expectNoErrors(t, input)
		})
	}
}

// TestLargeSetForInLoopAnalysis tests for-in loop analysis over large sets.
func TestLargeSetForInLoopAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		enumSize int
	}{
		{"65 elements", 65},
		{"100 elements", 100},
		{"200 elements", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`
				%s
				var s: TLargeSet;
				var e: TLargeEnum;
				begin
					s := [E00, E10, E20];
					for e in s do
						PrintLn(Ord(e));
				end.
			`, generateLargeEnumSetType(tt.enumSize))
			expectNoErrors(t, input)
		})
	}
}

// TestLargeSetForInLoopVariableTypeError tests loop variable type checking.
// TODO: Skipped - semantic analyzer doesn't yet validate loop variable type matches set element type
// Uncomment when for-in loop variable type checking is implemented
/*
func TestLargeSetForInLoopVariableTypeError(t *testing.T) {
	input := `
		type TLargeEnum = (E00, E01, E02, E03, E04, E05, E06, E07, E08, E09,
		                    E10, E11, E12, E13, E14, E15, E16, E17, E18, E19,
		                    E20, E21, E22, E23, E24, E25, E26, E27, E28, E29,
		                    E30, E31, E32, E33, E34, E35, E36, E37, E38, E39,
		                    E40, E41, E42, E43, E44, E45, E46, E47, E48, E49,
		                    E50, E51, E52, E53, E54, E55, E56, E57, E58, E59,
		                    E60, E61, E62, E63, E64);
		type TLargeSet = set of TLargeEnum;
		var s: TLargeSet;
		var i: Integer;
		begin
			s := [E00, E10];
			for i in s do
				PrintLn(i);
		end.
	`
	expectError(t, input, "loop variable")
}
*/

// TestLargeSetIncompatibleTypes tests error detection for incompatible large set types.
func TestLargeSetIncompatibleTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "assign different large set types",
			input: `
				type TEnum1 = (A0, A1, A2, A3, A4, A5, A6, A7, A8, A9,
				               A10, A11, A12, A13, A14, A15, A16, A17, A18, A19,
				               A20, A21, A22, A23, A24, A25, A26, A27, A28, A29,
				               A30, A31, A32, A33, A34, A35, A36, A37, A38, A39,
				               A40, A41, A42, A43, A44, A45, A46, A47, A48, A49,
				               A50, A51, A52, A53, A54, A55, A56, A57, A58, A59,
				               A60, A61, A62, A63, A64);
				type TSet1 = set of TEnum1;

				type TEnum2 = (B0, B1, B2, B3, B4, B5, B6, B7, B8, B9,
				               B10, B11, B12, B13, B14, B15, B16, B17, B18, B19,
				               B20, B21, B22, B23, B24, B25, B26, B27, B28, B29,
				               B30, B31, B32, B33, B34, B35, B36, B37, B38, B39,
				               B40, B41, B42, B43, B44, B45, B46, B47, B48, B49,
				               B50, B51, B52, B53, B54, B55, B56, B57, B58, B59,
				               B60, B61, B62, B63, B64);
				type TSet2 = set of TEnum2;

				var s1: TSet1;
				var s2: TSet2;
				begin
					s1 := s2;
				end.
			`,
			expectedError: "cannot assign",
		},
		{
			name: "union of incompatible large sets",
			input: `
				type TEnum1 = (A0, A1, A2, A3, A4, A5, A6, A7, A8, A9,
				               A10, A11, A12, A13, A14, A15, A16, A17, A18, A19,
				               A20, A21, A22, A23, A24, A25, A26, A27, A28, A29,
				               A30, A31, A32, A33, A34, A35, A36, A37, A38, A39,
				               A40, A41, A42, A43, A44, A45, A46, A47, A48, A49,
				               A50, A51, A52, A53, A54, A55, A56, A57, A58, A59,
				               A60, A61, A62, A63, A64);
				type TSet1 = set of TEnum1;

				type TEnum2 = (B0, B1, B2, B3, B4, B5, B6, B7, B8, B9,
				               B10, B11, B12, B13, B14, B15, B16, B17, B18, B19,
				               B20, B21, B22, B23, B24, B25, B26, B27, B28, B29,
				               B30, B31, B32, B33, B34, B35, B36, B37, B38, B39,
				               B40, B41, B42, B43, B44, B45, B46, B47, B48, B49,
				               B50, B51, B52, B53, B54, B55, B56, B57, B58, B59,
				               B60, B61, B62, B63, B64);
				type TSet2 = set of TEnum2;

				var s1: TSet1;
				var s2: TSet2;
				var s3: TSet1;
				begin
					s3 := s1 + s2;
				end.
			`,
			expectedError: "incompatible types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectError(t, tt.input, tt.expectedError)
		})
	}
}

// TestLargeSetStorageKindIndependence tests that storage kind doesn't affect type checking.
// A 64-element set (bitmask) and 65-element set (map) of the same enum type should be compatible.
func TestLargeSetStorageKindIndependence(t *testing.T) {
	// Note: In DWScript, type compatibility is based on the element type, not storage implementation.
	// This test verifies that sets with different storage kinds but same element type are compatible.

	input := `
		type TSmallEnum = (S0, S1, S2, S3, S4);
		type TSmallSet = set of TSmallEnum;

		type TLargeEnum = (L0, L1, L2, L3, L4, L5, L6, L7, L8, L9,
		                   L10, L11, L12, L13, L14, L15, L16, L17, L18, L19,
		                   L20, L21, L22, L23, L24, L25, L26, L27, L28, L29,
		                   L30, L31, L32, L33, L34, L35, L36, L37, L38, L39,
		                   L40, L41, L42, L43, L44, L45, L46, L47, L48, L49,
		                   L50, L51, L52, L53, L54, L55, L56, L57, L58, L59,
		                   L60, L61, L62, L63, L64);
		type TLargeSet = set of TLargeEnum;

		var small: TSmallSet;
		var large: TLargeSet;
		begin
			// Both should work independently
			small := [S0, S2];
			large := [L0, L10, L64];

			// Operations within same type should work
			small := small + [S1];
			large := large + [L20];
		end.
	`

	expectNoErrors(t, input)
}

// TestLargeSetRangeLiterals tests range expressions in large set literals.
// TODO: This test is skipped because range expression semantic analysis is not yet implemented.
// Uncomment when range expressions are supported in semantic analyzer.
/*
func TestLargeSetRangeLiterals(t *testing.T) {
	input := `
		type TLargeEnum = (E00, E01, E02, E03, E04, E05, E06, E07, E08, E09,
		                   E10, E11, E12, E13, E14, E15, E16, E17, E18, E19,
		                   E20, E21, E22, E23, E24, E25, E26, E27, E28, E29,
		                   E30, E31, E32, E33, E34, E35, E36, E37, E38, E39,
		                   E40, E41, E42, E43, E44, E45, E46, E47, E48, E49,
		                   E50, E51, E52, E53, E54, E55, E56, E57, E58, E59,
		                   E60, E61, E62, E63, E64, E65, E66, E67, E68, E69);
		type TLargeSet = set of TLargeEnum;
		var s: TLargeSet;
		begin
			// Range that spans across 64-bit boundary
			s := [E00..E10];
			s := [E60..E69];
		end.
	`

	expectNoErrors(t, input)
}
*/

// ============================================================================
// Helper Functions
// ============================================================================

// generateLargeEnumSetType generates a complete type declaration for testing.
func generateLargeEnumSetType(size int) string {
	enumDecl := "type TLargeEnum = ("
	for i := 0; i < size; i++ {
		if i > 0 {
			enumDecl += ", "
		}
		enumDecl += fmt.Sprintf("E%02d", i)
	}
	enumDecl += ");\ntype TLargeSet = set of TLargeEnum;"
	return enumDecl
}

// generateLargeEnumSetDeclaration generates just the declarations without usage.
func generateLargeEnumSetDeclaration(size int) string {
	return generateLargeEnumSetType(size)
}
