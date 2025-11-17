package parser

import (
	"errors"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

func TestNewParseContext(t *testing.T) {
	ctx := NewParseContext()

	if ctx == nil {
		t.Fatal("NewParseContext() returned nil")
	}

	if ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be false by default")
	}

	if ctx.BlockDepth() != 0 {
		t.Errorf("expected BlockDepth to be 0, got %d", ctx.BlockDepth())
	}

	if ctx.CurrentBlock() != nil {
		t.Error("expected CurrentBlock to be nil for empty stack")
	}
}

func TestNewParseContextWithFlags(t *testing.T) {
	flags := ContextFlags{
		ParsingPostCondition: true,
	}

	ctx := NewParseContextWithFlags(flags)

	if !ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be true")
	}
}

func TestContextFlags(t *testing.T) {
	ctx := NewParseContext()

	// Test ParsingPostCondition
	if ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be false initially")
	}

	ctx.SetParsingPostCondition(true)
	if !ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be true after setting")
	}

	ctx.SetParsingPostCondition(false)
	if ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be false after unsetting")
	}
}

func TestContextFlagsGetSet(t *testing.T) {
	ctx := NewParseContext()

	// Set flags via SetFlags
	newFlags := ContextFlags{
		ParsingPostCondition: true,
	}
	ctx.SetFlags(newFlags)

	// Get flags via Flags()
	flags := ctx.Flags()
	if !flags.ParsingPostCondition {
		t.Error("expected ParsingPostCondition to be true")
	}

	// Modify returned flags should not affect context (it's a copy)
	flags.ParsingPostCondition = false
	if !ctx.ParsingPostCondition() {
		t.Error("modifying returned flags should not affect context")
	}
}

func TestBlockStack(t *testing.T) {
	ctx := NewParseContext()

	// Initially empty
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected initial depth 0, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock() != nil {
		t.Error("expected CurrentBlock to be nil initially")
	}

	// Push first block
	pos1 := lexer.Position{Line: 1, Column: 5, Offset: 4}
	ctx.PushBlock("begin", pos1)

	if ctx.BlockDepth() != 1 {
		t.Errorf("expected depth 1 after first push, got %d", ctx.BlockDepth())
	}

	block := ctx.CurrentBlock()
	if block == nil {
		t.Fatal("expected CurrentBlock to be non-nil")
	}
	if block.BlockType != "begin" {
		t.Errorf("expected block type 'begin', got %q", block.BlockType)
	}
	if block.StartLine != 1 {
		t.Errorf("expected start line 1, got %d", block.StartLine)
	}

	// Push second block
	pos2 := lexer.Position{Line: 3, Column: 10, Offset: 25}
	ctx.PushBlock("if", pos2)

	if ctx.BlockDepth() != 2 {
		t.Errorf("expected depth 2 after second push, got %d", ctx.BlockDepth())
	}

	block = ctx.CurrentBlock()
	if block == nil {
		t.Fatal("expected CurrentBlock to be non-nil")
	}
	if block.BlockType != "if" {
		t.Errorf("expected block type 'if', got %q", block.BlockType)
	}
	if block.StartLine != 3 {
		t.Errorf("expected start line 3, got %d", block.StartLine)
	}

	// Push third block
	pos3 := lexer.Position{Line: 5, Column: 8, Offset: 50}
	ctx.PushBlock("while", pos3)

	if ctx.BlockDepth() != 3 {
		t.Errorf("expected depth 3 after third push, got %d", ctx.BlockDepth())
	}

	// Pop blocks
	ctx.PopBlock()
	if ctx.BlockDepth() != 2 {
		t.Errorf("expected depth 2 after first pop, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock().BlockType != "if" {
		t.Errorf("expected current block to be 'if', got %q", ctx.CurrentBlock().BlockType)
	}

	ctx.PopBlock()
	if ctx.BlockDepth() != 1 {
		t.Errorf("expected depth 1 after second pop, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock().BlockType != "begin" {
		t.Errorf("expected current block to be 'begin', got %q", ctx.CurrentBlock().BlockType)
	}

	ctx.PopBlock()
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected depth 0 after third pop, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock() != nil {
		t.Error("expected CurrentBlock to be nil after popping all blocks")
	}

	// Pop when empty should not panic
	ctx.PopBlock()
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected depth to remain 0, got %d", ctx.BlockDepth())
	}
}

func TestBlockStackCopy(t *testing.T) {
	ctx := NewParseContext()

	pos1 := lexer.Position{Line: 1, Column: 1, Offset: 0}
	ctx.PushBlock("begin", pos1)

	pos2 := lexer.Position{Line: 2, Column: 1, Offset: 10}
	ctx.PushBlock("if", pos2)

	// Get a copy of the block stack
	stack := ctx.BlockStack()
	if len(stack) != 2 {
		t.Errorf("expected stack length 2, got %d", len(stack))
	}

	// Modify the copy
	stack[0].BlockType = "modified"

	// Original should be unchanged
	if ctx.CurrentBlock().BlockType == "modified" {
		t.Error("modifying returned stack should not affect context")
	}
}

func TestSnapshot(t *testing.T) {
	ctx := NewParseContext()
	ctx.SetParsingPostCondition(true)

	pos1 := lexer.Position{Line: 1, Column: 1, Offset: 0}
	ctx.PushBlock("begin", pos1)

	pos2 := lexer.Position{Line: 2, Column: 1, Offset: 10}
	ctx.PushBlock("if", pos2)

	// Create snapshot
	snapshot := ctx.Snapshot()

	// Verify snapshot is independent
	if snapshot == ctx {
		t.Error("snapshot should be a different instance")
	}

	// Verify snapshot has same state
	if !snapshot.ParsingPostCondition() {
		t.Error("snapshot should have ParsingPostCondition true")
	}
	if snapshot.BlockDepth() != 2 {
		t.Errorf("snapshot should have depth 2, got %d", snapshot.BlockDepth())
	}

	// Modify original
	ctx.SetParsingPostCondition(false)
	ctx.PopBlock()

	// Snapshot should be unchanged
	if !snapshot.ParsingPostCondition() {
		t.Error("snapshot should still have ParsingPostCondition true")
	}
	if snapshot.BlockDepth() != 2 {
		t.Errorf("snapshot should still have depth 2, got %d", snapshot.BlockDepth())
	}

	// Original should be modified
	if ctx.ParsingPostCondition() {
		t.Error("original should have ParsingPostCondition false")
	}
	if ctx.BlockDepth() != 1 {
		t.Errorf("original should have depth 1, got %d", ctx.BlockDepth())
	}
}

func TestRestore(t *testing.T) {
	ctx := NewParseContext()

	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	ctx.PushBlock("begin", pos)

	// Create snapshot
	snapshot := ctx.Snapshot()

	// Modify context
	ctx.SetParsingPostCondition(true)
	ctx.PushBlock("if", lexer.Position{Line: 2, Column: 1, Offset: 10})
	ctx.PushBlock("while", lexer.Position{Line: 3, Column: 1, Offset: 20})

	// Verify modifications
	if !ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be true")
	}
	if ctx.BlockDepth() != 3 {
		t.Errorf("expected depth 3, got %d", ctx.BlockDepth())
	}

	// Restore from snapshot
	ctx.Restore(snapshot)

	// Verify restoration
	if ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be restored to false")
	}
	if ctx.BlockDepth() != 1 {
		t.Errorf("expected depth to be restored to 1, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock().BlockType != "begin" {
		t.Errorf("expected current block to be 'begin', got %q", ctx.CurrentBlock().BlockType)
	}
}

func TestClone(t *testing.T) {
	ctx := NewParseContext()
	ctx.SetParsingPostCondition(true)

	pos := lexer.Position{Line: 5, Column: 10, Offset: 50}
	ctx.PushBlock("for", pos)

	// Clone should work the same as Snapshot
	clone := ctx.Clone()

	if clone == ctx {
		t.Error("clone should be a different instance")
	}

	if !clone.ParsingPostCondition() {
		t.Error("clone should have ParsingPostCondition true")
	}

	if clone.BlockDepth() != 1 {
		t.Errorf("clone should have depth 1, got %d", clone.BlockDepth())
	}

	// Modify original
	ctx.PopBlock()

	// Clone should be unchanged
	if clone.BlockDepth() != 1 {
		t.Errorf("clone should still have depth 1, got %d", clone.BlockDepth())
	}
}

func TestReset(t *testing.T) {
	ctx := NewParseContext()
	ctx.SetParsingPostCondition(true)

	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}
	ctx.PushBlock("begin", pos)
	ctx.PushBlock("if", lexer.Position{Line: 2, Column: 1, Offset: 10})

	// Reset
	ctx.Reset()

	// Verify reset state
	if ctx.ParsingPostCondition() {
		t.Error("expected ParsingPostCondition to be false after reset")
	}
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected depth 0 after reset, got %d", ctx.BlockDepth())
	}
	if ctx.CurrentBlock() != nil {
		t.Error("expected CurrentBlock to be nil after reset")
	}
}

func TestWithBlock(t *testing.T) {
	ctx := NewParseContext()

	// Track whether function was called
	called := false
	returnedError := false

	pos := lexer.Position{Line: 10, Column: 5, Offset: 100}

	err := ctx.WithBlock("case", pos, func() error {
		called = true

		// Should be inside the block
		if ctx.BlockDepth() != 1 {
			t.Errorf("expected depth 1 inside WithBlock, got %d", ctx.BlockDepth())
		}
		if ctx.CurrentBlock() == nil {
			t.Error("expected CurrentBlock to be non-nil inside WithBlock")
		} else if ctx.CurrentBlock().BlockType != "case" {
			t.Errorf("expected block type 'case', got %q", ctx.CurrentBlock().BlockType)
		}

		if returnedError {
			return errors.New("test error")
		}
		return nil
	})

	if !called {
		t.Error("WithBlock should call the function")
	}

	if err != nil {
		t.Errorf("WithBlock returned unexpected error: %v", err)
	}

	// After WithBlock, block should be popped
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected depth 0 after WithBlock, got %d", ctx.BlockDepth())
	}

	// Test with error return
	returnedError = true
	called = false

	err = ctx.WithBlock("try", pos, func() error {
		called = true
		return errors.New("test error")
	})

	if !called {
		t.Error("WithBlock should call the function even when it returns error")
	}

	if err == nil {
		t.Error("WithBlock should return the error from the function")
	}

	// Block should still be popped even on error
	if ctx.BlockDepth() != 0 {
		t.Errorf("expected depth 0 after WithBlock with error, got %d", ctx.BlockDepth())
	}
}

func TestWithBlockPanic(t *testing.T) {
	ctx := NewParseContext()

	pos := lexer.Position{Line: 1, Column: 1, Offset: 0}

	// Test that panic still causes defer to run
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}

		// Block should be popped even on panic
		if ctx.BlockDepth() != 0 {
			t.Errorf("expected depth 0 after panic in WithBlock, got %d", ctx.BlockDepth())
		}
	}()

	_ = ctx.WithBlock("repeat", pos, func() error {
		panic("test panic")
	})
}
