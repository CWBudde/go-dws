package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/jsonvalue"
)

func TestValueToJSONValue_UndefinedJSONNode(t *testing.T) {
	e := &Evaluator{}
	got := e.valueToJSONValue(runtime.NewJSONValue(nil), nil, nil)
	if got == nil || got.kind != jsonvalue.KindNull {
		t.Fatalf("valueToJSONValue(undefined JSON) = %v, want null", got)
	}
}
