package types

import "testing"

func TestResolveOverload_CandidateIndex(t *testing.T) {
	integer := NewFunctionType([]Type{INTEGER}, INTEGER)
	float := NewFunctionType([]Type{FLOAT}, FLOAT)
	tests := []struct {
		name       string
		candidates []Type
		want       int
		wantError  string
	}{
		{name: "preserve index when skipping non-functions", candidates: []Type{STRING, float, integer}, want: 2},
		{name: "single non-function", candidates: []Type{STRING}, want: -1, wantError: "candidate is not a function type"},
		{name: "empty", want: -1, wantError: "no overload candidates provided"},
		{name: "ambiguous", candidates: []Type{integer, integer}, want: -1, wantError: "ambiguous overload call: 2 candidates with equal distance 0 for argument types: (Integer)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveOverload(tt.candidates, []Type{INTEGER})
			if got != tt.want {
				t.Errorf("candidate index = %d, want %d", got, tt.want)
			}
			if tt.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Error() != tt.wantError {
				t.Errorf("error = %v, want %q", err, tt.wantError)
			}
		})
	}
}

// TestSignaturesEqual tests signature comparison for overload detection
func TestSignaturesEqual(t *testing.T) {
	tests := []struct {
		sig1     *FunctionType
		sig2     *FunctionType
		name     string
		expected bool
	}{
		{
			name: "identical signatures",
			sig1: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: FLOAT,
			},
			sig2: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: FLOAT,
			},
			expected: true,
		},
		{
			name: "same parameters, different return type",
			sig1: &FunctionType{
				Parameters: []Type{INTEGER},
				ReturnType: FLOAT,
			},
			sig2: &FunctionType{
				Parameters: []Type{INTEGER},
				ReturnType: STRING,
			},
			expected: true, // Return type doesn't matter for signature equality
		},
		{
			name: "different parameter count",
			sig1: &FunctionType{
				Parameters: []Type{INTEGER},
				ReturnType: VOID,
			},
			sig2: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: VOID,
			},
			expected: false,
		},
		{
			name: "different parameter types",
			sig1: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: VOID,
			},
			sig2: &FunctionType{
				Parameters: []Type{INTEGER, FLOAT},
				ReturnType: VOID,
			},
			expected: false,
		},
		{
			name: "same types, different var modifier",
			sig1: &FunctionType{
				Parameters: []Type{INTEGER},
				VarParams:  []bool{false},
				ReturnType: VOID,
			},
			sig2: &FunctionType{
				Parameters: []Type{INTEGER},
				VarParams:  []bool{true}, // var parameter
				ReturnType: VOID,
			},
			expected: false,
		},
		{
			name: "same types, different const modifier",
			sig1: &FunctionType{
				Parameters:  []Type{STRING},
				ConstParams: []bool{false},
				ReturnType:  VOID,
			},
			sig2: &FunctionType{
				Parameters:  []Type{STRING},
				ConstParams: []bool{true}, // const parameter
				ReturnType:  VOID,
			},
			expected: false,
		},
		{
			name: "same types, different lazy modifier",
			sig1: &FunctionType{
				Parameters: []Type{BOOLEAN},
				LazyParams: []bool{false},
				ReturnType: VOID,
			},
			sig2: &FunctionType{
				Parameters: []Type{BOOLEAN},
				LazyParams: []bool{true}, // lazy parameter
				ReturnType: VOID,
			},
			expected: false,
		},
		{
			name: "both variadic with same element type",
			sig1: &FunctionType{
				Parameters:   []Type{NewDynamicArrayType(INTEGER)},
				IsVariadic:   true,
				VariadicType: INTEGER,
				ReturnType:   VOID,
			},
			sig2: &FunctionType{
				Parameters:   []Type{NewDynamicArrayType(INTEGER)},
				IsVariadic:   true,
				VariadicType: INTEGER,
				ReturnType:   VOID,
			},
			expected: true,
		},
		{
			name: "variadic vs non-variadic",
			sig1: &FunctionType{
				Parameters:   []Type{NewDynamicArrayType(INTEGER)},
				IsVariadic:   true,
				VariadicType: INTEGER,
				ReturnType:   VOID,
			},
			sig2: &FunctionType{
				Parameters: []Type{NewDynamicArrayType(INTEGER)},
				IsVariadic: false,
				ReturnType: VOID,
			},
			expected: false,
		},
		{
			name: "empty parameter lists",
			sig1: &FunctionType{
				Parameters: []Type{},
				ReturnType: INTEGER,
			},
			sig2: &FunctionType{
				Parameters: []Type{},
				ReturnType: STRING,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignaturesEqual(tt.sig1, tt.sig2)
			if result != tt.expected {
				t.Errorf("SignaturesEqual() = %v, want %v", result, tt.expected)
				t.Errorf("  sig1: %s", tt.sig1.String())
				t.Errorf("  sig2: %s", tt.sig2.String())
			}
		})
	}
}

// TestSignatureDistance tests distance calculation for overload resolution
func TestSignatureDistance(t *testing.T) {
	tests := []struct {
		signature *FunctionType
		name      string
		argTypes  []Type
		expected  int
	}{
		{
			name:     "exact match - single parameter",
			argTypes: []Type{INTEGER},
			signature: &FunctionType{
				Parameters: []Type{INTEGER},
				ReturnType: VOID,
			},
			expected: 0,
		},
		{
			name:     "exact match - multiple parameters",
			argTypes: []Type{INTEGER, STRING, FLOAT},
			signature: &FunctionType{
				Parameters: []Type{INTEGER, STRING, FLOAT},
				ReturnType: VOID,
			},
			expected: 0,
		},
		{
			name:     "implicit conversion - Integer to Float",
			argTypes: []Type{INTEGER},
			signature: &FunctionType{
				Parameters: []Type{FLOAT},
				ReturnType: VOID,
			},
			expected: 1,
		},
		{
			name:     "multiple conversions",
			argTypes: []Type{INTEGER, INTEGER},
			signature: &FunctionType{
				Parameters: []Type{FLOAT, FLOAT},
				ReturnType: VOID,
			},
			expected: 2, // Two Integer->Float conversions
		},
		{
			name:     "mixed exact and conversion",
			argTypes: []Type{INTEGER, STRING},
			signature: &FunctionType{
				Parameters: []Type{FLOAT, STRING},
				ReturnType: VOID,
			},
			expected: 1, // One conversion (Integer->Float), one exact (String)
		},
		{
			name:     "incompatible types",
			argTypes: []Type{STRING},
			signature: &FunctionType{
				Parameters: []Type{INTEGER},
				ReturnType: VOID,
			},
			expected: -1,
		},
		{
			name:     "too few arguments",
			argTypes: []Type{INTEGER},
			signature: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: VOID,
			},
			expected: -1,
		},
		{
			name:     "too many arguments",
			argTypes: []Type{INTEGER, STRING, FLOAT},
			signature: &FunctionType{
				Parameters: []Type{INTEGER, STRING},
				ReturnType: VOID,
			},
			expected: -1,
		},
		{
			name:     "empty arguments and parameters",
			argTypes: []Type{},
			signature: &FunctionType{
				Parameters: []Type{},
				ReturnType: VOID,
			},
			expected: 0,
		},
		{
			name:     "variadic with exact match",
			argTypes: []Type{INTEGER, INTEGER, INTEGER},
			signature: &FunctionType{
				Parameters:   []Type{INTEGER, NewDynamicArrayType(INTEGER)},
				IsVariadic:   true,
				VariadicType: INTEGER,
				ReturnType:   VOID,
			},
			expected: 0, // First arg exact, rest match variadic type
		},
		{
			name:     "variadic with no variadic args",
			argTypes: []Type{INTEGER},
			signature: &FunctionType{
				Parameters:   []Type{INTEGER, NewDynamicArrayType(INTEGER)},
				IsVariadic:   true,
				VariadicType: INTEGER,
				ReturnType:   VOID,
			},
			expected: 0, // Just the required parameter
		},
		{
			name:     "variadic with type conversion",
			argTypes: []Type{STRING, INTEGER, INTEGER},
			signature: &FunctionType{
				Parameters:   []Type{STRING, NewDynamicArrayType(FLOAT)},
				IsVariadic:   true,
				VariadicType: FLOAT,
				ReturnType:   VOID,
			},
			expected: 2, // Two Integer->Float conversions for variadic args
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SignatureDistance(tt.argTypes, tt.signature)
			if result != tt.expected {
				t.Errorf("SignatureDistance() = %d, want %d", result, tt.expected)
				t.Errorf("  args: %v", formatArgTypes(tt.argTypes))
				t.Errorf("  signature: %s", tt.signature.String())
			}
		})
	}
}

// TestTypeDistance tests individual type distance calculations
func TestTypeDistance(t *testing.T) {
	tests := []struct {
		from     Type
		to       Type
		name     string
		expected int
	}{
		{
			name:     "exact match - Integer",
			from:     INTEGER,
			to:       INTEGER,
			expected: 0,
		},
		{
			name:     "exact match - String",
			from:     STRING,
			to:       STRING,
			expected: 0,
		},
		{
			name:     "Integer to Float",
			from:     INTEGER,
			to:       FLOAT,
			expected: 1,
		},
		{
			name:     "Float to Integer (not allowed)",
			from:     FLOAT,
			to:       INTEGER,
			expected: -1,
		},
		{
			name:     "String to Integer (not allowed)",
			from:     STRING,
			to:       INTEGER,
			expected: -1,
		},
		{
			name:     "Integer to String (not allowed)",
			from:     INTEGER,
			to:       STRING,
			expected: -1,
		},
		{
			name:     "Boolean to String (not allowed)",
			from:     BOOLEAN,
			to:       STRING,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeDistance(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("typeDistance(%s, %s) = %d, want %d",
					tt.from.String(), tt.to.String(), result, tt.expected)
			}
		})
	}
}

// TestTypeDistanceReferenceConversions covers nil-literal, metaclass, and
// Variant-boxing distances used by cross-scope overload resolution (P4).
func TestTypeDistanceReferenceConversions(t *testing.T) {
	base := NewClassType("TBase", nil)
	sub := NewClassType("TSub", base)
	dynBase := NewDynamicArrayType(base)
	dynVariant := NewDynamicArrayType(VARIANT)

	tests := []struct {
		from     Type
		to       Type
		name     string
		expected int
	}{
		{name: "nil to class", from: NIL, to: base, expected: 1},
		{name: "nil to metaclass", from: NIL, to: NewClassOfType(base), expected: 1},
		{name: "nil to dynamic array (worse than class)", from: NIL, to: dynBase, expected: 2},
		{name: "subclass to base class", from: sub, to: base, expected: 1},
		{name: "metaclass of sub to metaclass of base", from: NewClassOfType(sub), to: NewClassOfType(base), expected: 1},
		{name: "variant boxing is worst-ranked", from: INTEGER, to: VARIANT, expected: 4},
		{name: "array elem variant conversion capped below boxing", from: NewDynamicArrayType(INTEGER), to: dynVariant, expected: 2},
		{name: "array of sub to array of base", from: NewDynamicArrayType(sub), to: dynBase, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeDistance(tt.from, tt.to); got != tt.expected {
				t.Errorf("typeDistance(%s, %s) = %d, want %d",
					tt.from.String(), tt.to.String(), got, tt.expected)
			}
		})
	}
}

// TestSignatureDistanceDefaultParams verifies parameters with default values
// are optional during overload resolution.
func TestSignatureDistanceDefaultParams(t *testing.T) {
	sig := NewFunctionTypeWithMetadata(
		[]Type{STRING, STRING},
		[]string{"s1", "s2"},
		[]interface{}{nil, struct{}{}}, // s2 has a default value
		[]bool{false, false}, []bool{false, false}, []bool{false, false},
		VOID,
	)

	tests := []struct {
		name     string
		args     []Type
		expected int
	}{
		{name: "all args provided", args: []Type{STRING, STRING}, expected: 0},
		{name: "optional arg omitted", args: []Type{STRING}, expected: 0},
		{name: "required arg missing", args: []Type{}, expected: -1},
		{name: "too many args", args: []Type{STRING, STRING, STRING}, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignatureDistance(tt.args, sig); got != tt.expected {
				t.Errorf("SignatureDistance(%v) = %d, want %d", tt.args, got, tt.expected)
			}
		})
	}
}
