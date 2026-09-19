package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
)

func TestObjectFieldDeclared_UninitializedField(t *testing.T) {
	class := runtime.NewClassInfo("TTarget")
	runtime.AddFieldToClass(class.Metadata, &runtime.FieldMetadata{Name: "Value"})
	obj := runtime.NewObjectInstance(class)

	if obj.GetField("Value") != nil {
		t.Fatalf("expected the field to start without a stored value")
	}
	if !objectFieldDeclared(obj, "vALUE") {
		t.Fatalf("expected a declared but unset field to count as declared")
	}
	if objectFieldDeclared(obj, "Missing") {
		t.Fatalf("expected an undeclared field not to count as declared")
	}
}
