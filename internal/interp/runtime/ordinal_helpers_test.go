package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestNewEnumValue(t *testing.T) {
	enumType := types.NewEnumType("TEven", map[string]int{
		"Zero": 0,
		"Two":  2,
		"Four": 4,
	}, []string{"Zero", "Two", "Four"})

	t.Run("declared ordinal", func(t *testing.T) {
		val := NewEnumValue("TEven", enumType, 2)
		if val.TypeName != "TEven" || val.ValueName != "Two" || val.OrdinalValue != 2 {
			t.Fatalf("unexpected enum value: %+v", val)
		}
	})

	t.Run("undeclared ordinal uses placeholder", func(t *testing.T) {
		val := NewEnumValue("TEven", enumType, 3)
		if val.ValueName != "$3" || val.OrdinalValue != 3 {
			t.Fatalf("unexpected enum placeholder value: %+v", val)
		}
	})
}

func TestEnumValueIndex(t *testing.T) {
	enumType := types.NewEnumType("TEven", map[string]int{
		"Zero": 0,
		"Two":  2,
		"Four": 4,
	}, []string{"Zero", "Two", "Four"})

	t.Run("canonical name", func(t *testing.T) {
		idx, err := EnumValueIndex(&EnumValue{
			TypeName:     "TEven",
			ValueName:    "Two",
			OrdinalValue: 2,
		}, enumType)
		if err != nil {
			t.Fatalf("EnumValueIndex returned error: %v", err)
		}
		if idx != 1 {
			t.Fatalf("index = %d, want 1", idx)
		}
	})

	t.Run("fallback to ordinal", func(t *testing.T) {
		idx, err := EnumValueIndex(&EnumValue{
			TypeName:     "TEven",
			ValueName:    "$2",
			OrdinalValue: 2,
		}, enumType)
		if err != nil {
			t.Fatalf("EnumValueIndex returned error: %v", err)
		}
		if idx != 1 {
			t.Fatalf("index = %d, want 1", idx)
		}
	})
}

func TestRebuildOrdinalValue(t *testing.T) {
	enumType := types.NewEnumType("TEven", map[string]int{
		"Zero": 0,
		"Two":  2,
		"Four": 4,
	}, []string{"Zero", "Two", "Four"})

	resolver := func(typeName string) (*types.EnumType, error) {
		if typeName != "TEven" {
			t.Fatalf("unexpected enum type lookup: %s", typeName)
		}
		return enumType, nil
	}

	tests := []struct {
		name     string
		template Value
		ordinal  int
		wantType string
		wantText string
	}{
		{
			name:     "integer",
			template: &IntegerValue{Value: 1},
			ordinal:  42,
			wantType: "INTEGER",
			wantText: "42",
		},
		{
			name:     "enum",
			template: &EnumValue{TypeName: "TEven", ValueName: "Zero", OrdinalValue: 0},
			ordinal:  2,
			wantType: "ENUM",
			wantText: "2",
		},
		{
			name:     "string",
			template: &StringValue{Value: "a"},
			ordinal:  int('z'),
			wantType: "STRING",
			wantText: "z",
		},
		{
			name:     "boolean",
			template: &BooleanValue{Value: false},
			ordinal:  1,
			wantType: "BOOLEAN",
			wantText: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := RebuildOrdinalValue(tt.template, tt.ordinal, resolver)
			if err != nil {
				t.Fatalf("RebuildOrdinalValue returned error: %v", err)
			}
			if val.Type() != tt.wantType {
				t.Fatalf("type = %s, want %s", val.Type(), tt.wantType)
			}
			if val.String() != tt.wantText {
				t.Fatalf("string = %q, want %q", val.String(), tt.wantText)
			}
		})
	}
}
