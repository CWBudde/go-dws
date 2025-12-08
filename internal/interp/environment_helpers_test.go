package interp

import (
	"io"
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/evaluator"
)

// TestSetEnvironment verifies that SetEnvironment updates both i.env and i.ctx.env atomically.
func TestSetEnvironment(t *testing.T) {
	i := New(io.Discard)
	newEnv := NewEnvironment()

	// Set a variable in the new environment to distinguish it
	newEnv.Define("testVar", &IntegerValue{Value: 42})

	// Call SetEnvironment
	i.SetEnvironment(newEnv)

	// Verify i.env is updated
	if i.env != newEnv {
		t.Error("i.env not updated to newEnv")
	}

	// Verify i.ctx.env is synced (extract from adapter)
	ctxEnv := i.ctx.Env()
	if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
		underlying := adapter.Underlying()
		if underlying != newEnv {
			t.Error("i.ctx.env not synced with i.env")
		}

		// Verify the variable is accessible through the adapter
		val, found := ctxEnv.Get("testVar")
		if !found {
			t.Error("testVar not found in ctx.env")
		}
		if intVal, ok := val.(*IntegerValue); ok {
			if intVal.Value != 42 {
				t.Errorf("testVar value mismatch: got %d, want 42", intVal.Value)
			}
		} else {
			t.Error("testVar type mismatch")
		}
	} else {
		t.Error("i.ctx.env is not an EnvironmentAdapter")
	}
}

// TestPushEnvironment verifies that PushEnvironment creates a new enclosed environment
// and atomically sets both i.env and i.ctx.env.
func TestPushEnvironment(t *testing.T) {
	i := New(io.Discard)
	originalEnv := i.env

	// Define a variable in the original environment
	originalEnv.Define("parentVar", &IntegerValue{Value: 100})

	// Push a new environment
	newEnv := i.PushEnvironment(i.env)

	// Verify new environment is returned
	if newEnv == nil {
		t.Fatal("PushEnvironment returned nil")
	}

	// Verify i.env is updated to the new environment
	if i.env != newEnv {
		t.Error("i.env not updated to new environment")
	}

	// Verify the new environment is enclosed by the original
	if newEnv.outer != originalEnv {
		t.Error("new environment not enclosed by original")
	}

	// Verify i.ctx.env is synced
	ctxEnv := i.ctx.Env()
	if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
		underlying := adapter.Underlying()
		if underlying != newEnv {
			t.Error("i.ctx.env not synced with new environment")
		}
	} else {
		t.Error("i.ctx.env is not an EnvironmentAdapter")
	}

	// Verify the new environment can access parent variables
	val, found := newEnv.Get("parentVar")
	if !found {
		t.Error("parentVar not accessible from new environment")
	}
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 100 {
			t.Errorf("parentVar value mismatch: got %d, want 100", intVal.Value)
		}
	}
}

// TestRestoreEnvironment verifies that RestoreEnvironment restores a previously saved
// environment to both i.env and i.ctx.env.
func TestRestoreEnvironment(t *testing.T) {
	i := New(io.Discard)
	originalEnv := i.env

	// Define a variable in the original environment
	originalEnv.Define("original", &IntegerValue{Value: 1})

	// Push a new environment
	tempEnv := i.PushEnvironment(i.env)
	tempEnv.Define("temp", &IntegerValue{Value: 2})

	// Verify we're in the temp environment
	if i.env != tempEnv {
		t.Fatal("not in temp environment after push")
	}

	// Restore the original environment
	i.RestoreEnvironment(originalEnv)

	// Verify i.env is restored
	if i.env != originalEnv {
		t.Error("i.env not restored to original")
	}

	// Verify i.ctx.env is synced
	ctxEnv := i.ctx.Env()
	if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
		underlying := adapter.Underlying()
		if underlying != originalEnv {
			t.Error("i.ctx.env not synced with restored environment")
		}
	} else {
		t.Error("i.ctx.env is not an EnvironmentAdapter")
	}

	// Verify the original variable is accessible
	val, found := i.env.Get("original")
	if !found {
		t.Error("original variable not found after restore")
	}
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 1 {
			t.Errorf("original value mismatch: got %d, want 1", intVal.Value)
		}
	}

	// Verify the temp variable is not accessible (we're out of that scope)
	_, found = i.env.GetLocal("temp")
	if found {
		t.Error("temp variable should not be in local scope after restore")
	}
}

// TestEnvironmentSyncAfterPush verifies that i.env and i.ctx.env stay synchronized
// after multiple push/restore operations.
func TestEnvironmentSyncAfterPush(t *testing.T) {
	i := New(io.Discard)

	// Helper to verify sync
	verifySynced := func(t *testing.T, name string) {
		t.Helper()
		ctxEnv := i.ctx.Env()
		if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
			underlying := adapter.Underlying()
			if underlying != i.env {
				t.Errorf("%s: i.ctx.env not synced with i.env", name)
			}
		} else {
			t.Errorf("%s: i.ctx.env is not an EnvironmentAdapter", name)
		}
	}

	// Initial sync
	verifySynced(t, "initial")

	// Push environment 1
	env1 := i.PushEnvironment(i.env)
	env1.Define("var1", &IntegerValue{Value: 1})
	verifySynced(t, "after push 1")

	// Push environment 2
	env2 := i.PushEnvironment(i.env)
	env2.Define("var2", &IntegerValue{Value: 2})
	verifySynced(t, "after push 2")

	// Restore to env1
	i.RestoreEnvironment(env1)
	verifySynced(t, "after restore to env1")
	if i.env != env1 {
		t.Error("not in env1 after restore")
	}

	// Verify var1 is accessible but var2 is not
	if _, found := i.env.Get("var1"); !found {
		t.Error("var1 not found in env1")
	}
	if _, found := i.env.GetLocal("var2"); found {
		t.Error("var2 should not be in local scope of env1")
	}
}

// TestNestedPushRestore verifies correct behavior with deeply nested scopes.
// This simulates nested function calls or nested loops.
func TestNestedPushRestore(t *testing.T) {
	i := New(io.Discard)
	env0 := i.env
	env0.Define("level", &IntegerValue{Value: 0})

	// Push 3 levels of nesting
	env1 := i.PushEnvironment(i.env)
	env1.Define("level", &IntegerValue{Value: 1})

	env2 := i.PushEnvironment(i.env)
	env2.Define("level", &IntegerValue{Value: 2})

	env3 := i.PushEnvironment(i.env)
	env3.Define("level", &IntegerValue{Value: 3})

	// Verify we're at level 3
	val, found := i.env.GetLocal("level")
	if !found {
		t.Fatal("level not found at level 3")
	}
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 3 {
			t.Errorf("expected level 3, got %d", intVal.Value)
		}
	}

	// Verify i.env and i.ctx.env are synced at deepest level
	ctxEnv := i.ctx.Env()
	if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
		if adapter.Underlying() != i.env {
			t.Error("environments not synced at level 3")
		}
	}

	// Restore to level 2
	i.RestoreEnvironment(env2)
	val, _ = i.env.GetLocal("level")
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 2 {
			t.Errorf("expected level 2 after first restore, got %d", intVal.Value)
		}
	}

	// Restore to level 1
	i.RestoreEnvironment(env1)
	val, _ = i.env.GetLocal("level")
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 1 {
			t.Errorf("expected level 1 after second restore, got %d", intVal.Value)
		}
	}

	// Restore to level 0
	i.RestoreEnvironment(env0)
	val, _ = i.env.GetLocal("level")
	if intVal, ok := val.(*IntegerValue); ok {
		if intVal.Value != 0 {
			t.Errorf("expected level 0 after third restore, got %d", intVal.Value)
		}
	}

	// Final sync verification
	ctxEnv = i.ctx.Env()
	if adapter, ok := ctxEnv.(*evaluator.EnvironmentAdapter); ok {
		if adapter.Underlying() != env0 {
			t.Error("environments not synced after full restore")
		}
	}
}
