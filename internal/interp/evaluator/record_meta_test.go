package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestRecordMeta_OverloadClassifiers(t *testing.T) {
	e := &Evaluator{}
	record := types.NewRecordType("R", nil)
	value := &runtime.RecordTypeValue{RecordType: record}
	for _, tt := range []struct {
		classify func(Value) types.Type
		name     string
	}{
		{e.getValueType, "function"}, {e.runtimeValueType, "method"}, {runtime.LanguageType, "language"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			meta := tt.classify(value)
			if meta == nil || meta.TypeKind() != "RECORD_META" || meta.Equals(record) {
				t.Fatalf("type = %v; want distinct record meta", meta)
			}
			if got := tt.classify(&runtime.RecordValue{RecordType: record}); got != record {
				t.Fatalf("instance type = %v", got)
			}
		})
	}
}
