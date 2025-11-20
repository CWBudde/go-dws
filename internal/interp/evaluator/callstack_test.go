package evaluator

import (
	"strings"
	"testing"

	"github.com/cwbudde/go-dws/internal/lexer"
)

// TestNewCallStack tests creating a new call stack.
func TestNewCallStack(t *testing.T) {
	t.Run("default max depth", func(t *testing.T) {
		cs := NewCallStack(0)
		if cs.MaxDepth() != 1024 {
			t.Errorf("NewCallStack(0) max depth = %d, want 1024", cs.MaxDepth())
		}
		if cs.Depth() != 0 {
			t.Errorf("NewCallStack() depth = %d, want 0", cs.Depth())
		}
		if !cs.IsEmpty() {
			t.Errorf("NewCallStack() IsEmpty() = false, want true")
		}
	})

	t.Run("custom max depth", func(t *testing.T) {
		cs := NewCallStack(100)
		if cs.MaxDepth() != 100 {
			t.Errorf("NewCallStack(100) max depth = %d, want 100", cs.MaxDepth())
		}
	})

	t.Run("negative max depth uses default", func(t *testing.T) {
		cs := NewCallStack(-1)
		if cs.MaxDepth() != 1024 {
			t.Errorf("NewCallStack(-1) max depth = %d, want 1024", cs.MaxDepth())
		}
	})
}

// TestCallStack_Push tests pushing frames onto the stack.
func TestCallStack_Push(t *testing.T) {
	cs := NewCallStack(3)
	pos := lexer.Position{Line: 1, Column: 1}

	// Push first frame
	err := cs.Push("func1", "test.dws", &pos)
	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
	if cs.Depth() != 1 {
		t.Errorf("After first push, depth = %d, want 1", cs.Depth())
	}

	// Push second frame
	err = cs.Push("func2", "test.dws", &pos)
	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
	if cs.Depth() != 2 {
		t.Errorf("After second push, depth = %d, want 2", cs.Depth())
	}

	// Push third frame (should succeed - at max depth)
	err = cs.Push("func3", "test.dws", &pos)
	if err != nil {
		t.Errorf("Push() error = %v, want nil", err)
	}
	if cs.Depth() != 3 {
		t.Errorf("After third push, depth = %d, want 3", cs.Depth())
	}

	// Push fourth frame (should fail - exceeds max depth)
	err = cs.Push("func4", "test.dws", &pos)
	if err == nil {
		t.Errorf("Push() at max depth should return error, got nil")
	}
	if !strings.Contains(err.Error(), "stack overflow") {
		t.Errorf("Push() error = %v, want stack overflow error", err)
	}
	if cs.Depth() != 3 {
		t.Errorf("After failed push, depth = %d, want 3", cs.Depth())
	}
}

// TestCallStack_Pop tests popping frames from the stack.
func TestCallStack_Pop(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	// Push some frames
	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)
	cs.Push("func3", "test.dws", &pos)

	if cs.Depth() != 3 {
		t.Fatalf("After pushes, depth = %d, want 3", cs.Depth())
	}

	// Pop first frame
	cs.Pop()
	if cs.Depth() != 2 {
		t.Errorf("After first pop, depth = %d, want 2", cs.Depth())
	}

	// Pop second frame
	cs.Pop()
	if cs.Depth() != 1 {
		t.Errorf("After second pop, depth = %d, want 1", cs.Depth())
	}

	// Pop third frame
	cs.Pop()
	if cs.Depth() != 0 {
		t.Errorf("After third pop, depth = %d, want 0", cs.Depth())
	}
	if !cs.IsEmpty() {
		t.Errorf("After all pops, IsEmpty() = false, want true")
	}

	// Pop from empty stack (should not panic)
	cs.Pop()
	if cs.Depth() != 0 {
		t.Errorf("After popping empty stack, depth = %d, want 0", cs.Depth())
	}
}

// TestCallStack_Current tests getting the current frame.
func TestCallStack_Current(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 10, Column: 5}

	// Current on empty stack should return nil
	if current := cs.Current(); current != nil {
		t.Errorf("Current() on empty stack = %v, want nil", current)
	}

	// Push a frame
	cs.Push("testFunc", "test.dws", &pos)

	// Get current frame
	current := cs.Current()
	if current == nil {
		t.Fatal("Current() = nil, want non-nil")
	}
	if current.FunctionName != "testFunc" {
		t.Errorf("Current().FunctionName = %s, want testFunc", current.FunctionName)
	}

	// Push another frame
	pos2 := lexer.Position{Line: 20, Column: 10}
	cs.Push("anotherFunc", "other.dws", &pos2)

	// Current should be the most recent frame
	current = cs.Current()
	if current == nil {
		t.Fatal("Current() after second push = nil, want non-nil")
	}
	if current.FunctionName != "anotherFunc" {
		t.Errorf("Current().FunctionName = %s, want anotherFunc", current.FunctionName)
	}
}

// TestCallStack_Frames tests getting all frames.
func TestCallStack_Frames(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	// Empty stack
	frames := cs.Frames()
	if len(frames) != 0 {
		t.Errorf("Frames() on empty stack length = %d, want 0", len(frames))
	}

	// Push some frames
	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)
	cs.Push("func3", "test.dws", &pos)

	frames = cs.Frames()
	if len(frames) != 3 {
		t.Errorf("Frames() length = %d, want 3", len(frames))
	}

	// Check order (oldest to newest)
	if frames[0].FunctionName != "func1" {
		t.Errorf("frames[0].FunctionName = %s, want func1", frames[0].FunctionName)
	}
	if frames[1].FunctionName != "func2" {
		t.Errorf("frames[1].FunctionName = %s, want func2", frames[1].FunctionName)
	}
	if frames[2].FunctionName != "func3" {
		t.Errorf("frames[2].FunctionName = %s, want func3", frames[2].FunctionName)
	}

	// Verify it's a copy (modifying returned frames shouldn't affect the stack)
	frames[0].FunctionName = "modified"
	originalFrames := cs.Frames()
	if originalFrames[0].FunctionName != "func1" {
		t.Errorf("Modifying returned frames affected original stack")
	}
}

// TestCallStack_WillOverflow tests overflow detection.
func TestCallStack_WillOverflow(t *testing.T) {
	cs := NewCallStack(3)
	pos := lexer.Position{Line: 1, Column: 1}

	if cs.WillOverflow() {
		t.Errorf("WillOverflow() on empty stack = true, want false")
	}

	cs.Push("func1", "test.dws", &pos)
	if cs.WillOverflow() {
		t.Errorf("WillOverflow() at depth 1 = true, want false")
	}

	cs.Push("func2", "test.dws", &pos)
	if cs.WillOverflow() {
		t.Errorf("WillOverflow() at depth 2 = true, want false")
	}

	cs.Push("func3", "test.dws", &pos)
	if !cs.WillOverflow() {
		t.Errorf("WillOverflow() at max depth = false, want true")
	}

	// After popping, should not overflow
	cs.Pop()
	if cs.WillOverflow() {
		t.Errorf("WillOverflow() after pop = true, want false")
	}
}

// TestCallStack_Clear tests clearing the stack.
func TestCallStack_Clear(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)
	cs.Push("func3", "test.dws", &pos)

	if cs.Depth() != 3 {
		t.Fatalf("Before clear, depth = %d, want 3", cs.Depth())
	}

	cs.Clear()

	if cs.Depth() != 0 {
		t.Errorf("After clear, depth = %d, want 0", cs.Depth())
	}
	if !cs.IsEmpty() {
		t.Errorf("After clear, IsEmpty() = false, want true")
	}

	// Max depth should be preserved
	if cs.MaxDepth() != 10 {
		t.Errorf("After clear, MaxDepth() = %d, want 10", cs.MaxDepth())
	}
}

// TestCallStack_SetMaxDepth tests setting the max depth.
func TestCallStack_SetMaxDepth(t *testing.T) {
	cs := NewCallStack(10)

	cs.SetMaxDepth(50)
	if cs.MaxDepth() != 50 {
		t.Errorf("SetMaxDepth(50) max depth = %d, want 50", cs.MaxDepth())
	}

	cs.SetMaxDepth(0)
	if cs.MaxDepth() != 1024 {
		t.Errorf("SetMaxDepth(0) max depth = %d, want 1024", cs.MaxDepth())
	}

	cs.SetMaxDepth(-10)
	if cs.MaxDepth() != 1024 {
		t.Errorf("SetMaxDepth(-10) max depth = %d, want 1024", cs.MaxDepth())
	}
}

// TestCallStack_Clone tests cloning the stack.
func TestCallStack_Clone(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)

	clone := cs.Clone()

	// Check that clone has same content
	if clone.Depth() != cs.Depth() {
		t.Errorf("clone.Depth() = %d, want %d", clone.Depth(), cs.Depth())
	}
	if clone.MaxDepth() != cs.MaxDepth() {
		t.Errorf("clone.MaxDepth() = %d, want %d", clone.MaxDepth(), cs.MaxDepth())
	}

	// Modifying clone shouldn't affect original
	clone.Push("func3", "test.dws", &pos)
	if cs.Depth() != 2 {
		t.Errorf("After modifying clone, original depth = %d, want 2", cs.Depth())
	}
	if clone.Depth() != 3 {
		t.Errorf("After push to clone, clone depth = %d, want 3", clone.Depth())
	}
}

// TestCallStack_GetFrameAt tests getting frames by index.
func TestCallStack_GetFrameAt(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)
	cs.Push("func3", "test.dws", &pos)

	tests := []struct {
		wantName string
		index    int
		wantNil  bool
	}{
		{0, "func1", false},
		{1, "func2", false},
		{2, "func3", false},
		{3, "", true},  // Out of bounds
		{-1, "", true}, // Negative index
		{10, "", true}, // Way out of bounds
	}

	for _, tt := range tests {
		frame := cs.GetFrameAt(tt.index)
		if tt.wantNil {
			if frame != nil {
				t.Errorf("GetFrameAt(%d) = %v, want nil", tt.index, frame)
			}
		} else {
			if frame == nil {
				t.Errorf("GetFrameAt(%d) = nil, want non-nil", tt.index)
			} else if frame.FunctionName != tt.wantName {
				t.Errorf("GetFrameAt(%d).FunctionName = %s, want %s", tt.index, frame.FunctionName, tt.wantName)
			}
		}
	}
}

// TestCallStack_FindFrame tests finding frames by name.
func TestCallStack_FindFrame(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)
	cs.Push("func3", "test.dws", &pos)

	tests := []struct {
		name      string
		wantIndex int
		wantFound bool
	}{
		{"func1", 0, true},
		{"func2", 1, true},
		{"func3", 2, true},
		{"nonexistent", -1, false},
		{"", -1, false},
	}

	for _, tt := range tests {
		frame, index := cs.FindFrame(tt.name)
		if tt.wantFound {
			if frame == nil {
				t.Errorf("FindFrame(%s) frame = nil, want non-nil", tt.name)
			}
			if index != tt.wantIndex {
				t.Errorf("FindFrame(%s) index = %d, want %d", tt.name, index, tt.wantIndex)
			}
			if frame != nil && frame.FunctionName != tt.name {
				t.Errorf("FindFrame(%s).FunctionName = %s, want %s", tt.name, frame.FunctionName, tt.name)
			}
		} else {
			if frame != nil {
				t.Errorf("FindFrame(%s) frame = %v, want nil", tt.name, frame)
			}
			if index != -1 {
				t.Errorf("FindFrame(%s) index = %d, want -1", tt.name, index)
			}
		}
	}
}

// TestCallStack_ContainsFunction tests checking if a function is in the stack.
func TestCallStack_ContainsFunction(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	cs.Push("func1", "test.dws", &pos)
	cs.Push("func2", "test.dws", &pos)

	tests := []struct {
		name string
		want bool
	}{
		{"func1", true},
		{"func2", true},
		{"func3", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := cs.ContainsFunction(tt.name); got != tt.want {
			t.Errorf("ContainsFunction(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestCallStack_String tests string representation.
func TestCallStack_String(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 10, Column: 5}

	// Empty stack
	if s := cs.String(); s != "" {
		t.Errorf("String() on empty stack = %q, want empty string", s)
	}

	// With frames
	cs.Push("testFunc", "test.dws", &pos)
	s := cs.String()
	if !strings.Contains(s, "testFunc") {
		t.Errorf("String() = %q, should contain 'testFunc'", s)
	}
}

// TestCallStack_FormatError tests error formatting with stack trace.
func TestCallStack_FormatError(t *testing.T) {
	cs := NewCallStack(10)
	pos := lexer.Position{Line: 1, Column: 1}

	// Empty stack
	formatted := cs.FormatError("test error")
	if formatted != "test error" {
		t.Errorf("FormatError() on empty stack = %q, want 'test error'", formatted)
	}

	// With frames
	cs.Push("func1", "test.dws", &pos)
	formatted = cs.FormatError("test error")
	if !strings.Contains(formatted, "test error") {
		t.Errorf("FormatError() should contain error message")
	}
	if !strings.Contains(formatted, "Call stack:") {
		t.Errorf("FormatError() should contain 'Call stack:'")
	}
	if !strings.Contains(formatted, "func1") {
		t.Errorf("FormatError() should contain function name")
	}
}

// TestCallStack_StackOverflow tests stack overflow scenarios.
func TestCallStack_StackOverflow(t *testing.T) {
	cs := NewCallStack(5)
	pos := lexer.Position{Line: 1, Column: 1}

	// Fill the stack
	for i := 0; i < 5; i++ {
		err := cs.Push("func", "test.dws", &pos)
		if err != nil {
			t.Fatalf("Push() at depth %d returned error: %v", i, err)
		}
	}

	// Next push should fail
	err := cs.Push("overflow", "test.dws", &pos)
	if err == nil {
		t.Error("Push() at overflow should return error")
	}
	if !strings.Contains(err.Error(), "stack overflow") {
		t.Errorf("Error should mention stack overflow, got: %v", err)
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("Error should mention function name, got: %v", err)
	}
}
