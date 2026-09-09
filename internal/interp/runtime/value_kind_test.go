package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestValueKindSeparatesRepresentationAndLanguageType(t *testing.T) {
	first := &RecordValue{RecordType: &types.RecordType{Name: "First"}}
	second := &RecordValue{RecordType: &types.RecordType{Name: "Second"}}
	if KindOf(first) != KindRecord || KindOf(second) != KindRecord {
		t.Fatal("records must have record representation")
	}
	if types.OperatorTypesEqual(LanguageType(first), LanguageType(second)) {
		t.Fatal("distinct records must retain language identity")
	}
	if KindOf(&FunctionPointerValue{SelfObject: &NilValue{}}) != KindMethodPointer {
		t.Fatal("bound method representation lost")
	}
	if KindOf(&NullValue{}) == KindOf(&NilValue{}) {
		t.Fatal("null and nil representations must remain distinct")
	}
}
