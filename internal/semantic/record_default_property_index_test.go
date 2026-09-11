package semantic

import "testing"

// recordDefaultPropertyPrelude declares a record whose default property is
// indexed by an Integer and backed by a getter method.
const recordDefaultPropertyPrelude = `
type TRec = record
   FData : array [0..3] of String;
   function GetIdx(i : Integer) : String;
   begin
      Result := FData[i];
   end;
   property Idx[i : Integer] : String read GetIdx; default;
end;
`

// TestRecordDefaultPropertyIndexType checks that indexing a record through its
// default property validates the index expression against the property's
// declared index parameter type, matching the class behaviour.
func TestRecordDefaultPropertyIndexType(t *testing.T) {
	// Any index-type mismatch shares this prefix; want holds the full message.
	const indexErrorPrefix = `Array index expected "Integer" but got `

	tests := []struct {
		name   string
		source string
		// want is the expected index-type diagnostic, or empty when the index
		// is valid.
		want string
	}{
		{
			name:   "string index on integer-indexed default property",
			source: recordDefaultPropertyPrelude + "var r : TRec;\nPrintLn(r['oops']);\n",
			want:   indexErrorPrefix + `"String"`,
		},
		{
			name:   "float index on integer-indexed default property",
			source: recordDefaultPropertyPrelude + "var r : TRec;\nPrintLn(r[1.5]);\n",
			want:   indexErrorPrefix + `"Float"`,
		},
		{
			name:   "integer index is accepted",
			source: recordDefaultPropertyPrelude + "var r : TRec;\nPrintLn(r[1]);\n",
		},
		{
			name:   "integer expression index is accepted",
			source: recordDefaultPropertyPrelude + "var r : TRec;\nvar i := 2;\nPrintLn(r[i + 1]);\n",
		},
		{
			name: "field-backed default property validates the declared index type",
			source: `
type TRec = record
   FData : array [0..3] of String;
   property Idx[i : Integer] : String read FData; default;
end;
var r : TRec;
PrintLn(r['oops']);
`,
			want: indexErrorPrefix + `"String"`,
		},
		{
			name: "non-indexed default property still accepts any index",
			source: `
type TRec = record
   FData : array [0..3] of String;
   property Data : array [0..3] of String read FData; default;
end;
var r : TRec;
PrintLn(r[1]);
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := analyzeRecordSource(t, tt.source)
			if tt.want != "" {
				if !hasDiagnostic(diags, tt.want) {
					t.Errorf("missing diagnostic %q (got: %v)", tt.want, diags)
				}
				return
			}
			if hasDiagnostic(diags, indexErrorPrefix) {
				t.Errorf("unexpected index-type diagnostic (got: %v)", diags)
			}
		})
	}
}
