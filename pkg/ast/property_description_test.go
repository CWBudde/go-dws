package ast

import "testing"

func TestPropertyDecl_StringDescription(t *testing.T) {
	tests := []struct {
		name, value, want string
		present           bool
	}{
		{"absent", "", "property P: Integer reintroduce;", false},
		{"empty", "", `property P: Integer description "" reintroduce;`, true},
		{"multiline", "first\nsecond", "property P: Integer description \"first\nsecond\" reintroduce;", true},
		{"escaped", "it's \"quoted\"", `property P: Integer description "it's ""quoted""" reintroduce;`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := &PropertyDecl{Name: &Identifier{Value: "P"}, Type: &TypeAnnotation{Name: "Integer"}, Description: tt.value, HasDescription: tt.present, IsReintroduce: true}
			if got := prop.String(); got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
