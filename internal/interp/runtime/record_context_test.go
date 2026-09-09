package runtime

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/types"
)

func TestExecutionContext_RecordTypeIdentity(t *testing.T) {
	outer := types.NewRecordType("TOuter", map[string]types.Type{"value": types.INTEGER})
	inner := types.NewRecordType("", map[string]types.Type{"text": types.STRING})
	ctx := NewExecutionContext(NewEnvironment())
	ctx.SetRecordTypeContext(outer)
	clone := ctx.Clone()
	if clone.RecordTypeContext() != outer {
		t.Fatal("clone lost the exact expected record type")
	}
	previous := ctx.RecordTypeContext()
	ctx.SetRecordTypeContext(inner)
	ctx.SetRecordTypeContext(previous)
	if ctx.RecordTypeContext() != outer {
		t.Fatal("nested anonymous record context did not restore its parent")
	}
	clone.Reset()
	if clone.RecordTypeContext() != nil || ctx.RecordTypeContext() != outer {
		t.Fatal("reset changed another execution context")
	}
}
