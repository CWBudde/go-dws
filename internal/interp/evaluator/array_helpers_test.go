package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

func TestEvalArrayPush_CopiesRecordValues(t *testing.T) {
	e := &Evaluator{}

	original := &runtime.RecordValue{
		Fields: map[string]runtime.Value{
			"Start":     &runtime.IntegerValue{Value: 1},
			"Direction": &runtime.IntegerValue{Value: 2},
		},
	}

	arr := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
	}

	if result := e.evalArrayPush(arr, []Value{original}, nil); isError(result) {
		t.Fatalf("evalArrayPush returned error: %v", result)
	}

	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 element after push, got %d", len(arr.Elements))
	}

	pushed, ok := arr.Elements[0].(*runtime.RecordValue)
	if !ok {
		t.Fatalf("pushed element has type %T, want *runtime.RecordValue", arr.Elements[0])
	}

	if pushed == original {
		t.Fatal("expected array push to store a copy, not the original record pointer")
	}

	original.Fields["Start"] = &runtime.IntegerValue{Value: 99}

	start, ok := pushed.Fields["Start"].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("pushed record field has type %T, want *runtime.IntegerValue", pushed.Fields["Start"])
	}

	if start.Value != 1 {
		t.Fatalf("expected copied record to preserve original value 1, got %d", start.Value)
	}
}

func TestArrayHelperCopy_CopiesRecordValues(t *testing.T) {
	original := &runtime.RecordValue{
		Fields: map[string]runtime.Value{
			"Start":     &runtime.IntegerValue{Value: 1},
			"Direction": &runtime.IntegerValue{Value: 2},
		},
	}

	arr := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.INTEGER),
		Elements:  []runtime.Value{original},
	}

	result := ArrayHelperCopy(arr)
	copied, ok := result.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("ArrayHelperCopy returned %T, want *runtime.ArrayValue", result)
	}

	if len(copied.Elements) != 1 {
		t.Fatalf("expected 1 element after copy, got %d", len(copied.Elements))
	}

	copiedRecord, ok := copied.Elements[0].(*runtime.RecordValue)
	if !ok {
		t.Fatalf("copied element has type %T, want *runtime.RecordValue", copied.Elements[0])
	}

	if copiedRecord == original {
		t.Fatal("expected ArrayHelperCopy to store a copy, not the original record pointer")
	}

	original.Fields["Start"] = &runtime.IntegerValue{Value: 99}

	start, ok := copiedRecord.Fields["Start"].(*runtime.IntegerValue)
	if !ok {
		t.Fatalf("copied record field has type %T, want *runtime.IntegerValue", copiedRecord.Fields["Start"])
	}

	if start.Value != 1 {
		t.Fatalf("expected copied record to preserve original value 1, got %d", start.Value)
	}
}
