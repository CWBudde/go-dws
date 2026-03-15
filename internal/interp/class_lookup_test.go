package interp

import "testing"

func mustLookupTestClass(t *testing.T, interp *Interpreter, name string) *ClassInfo {
	t.Helper()

	classInfo := interp.lookupRegisteredClassInfo(name)
	if classInfo == nil {
		t.Fatalf("expected class %s to be registered", name)
	}

	return classInfo
}
