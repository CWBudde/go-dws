package parser

import (
	"github.com/cwbudde/go-dws/internal/lexer"
)

// ParseContext encapsulates the parsing context including flags and block stack.
// This consolidates scattered state variables into a single, well-structured type.
//
// The context tracks:
//   - Parsing phase flags (semantic analysis, postconditions, etc.)
//   - Block nesting stack for better error messages
//   - Future extensibility for additional context (loop depth, function scope, etc.)
//
// Example usage:
//
//	ctx := NewParseContext()
//	ctx.PushBlock("if", token.Pos)
//	defer ctx.PopBlock()
//	// ... parse if statement ...
type ParseContext struct {
	// flags contains structured boolean flags for parsing behavior
	flags ContextFlags

	// blockStack tracks nested block contexts for error reporting
	// Managed via PushBlock/PopBlock methods
	blockStack []BlockContext

	// Future extensibility examples (not planned features, just illustrative):
	// - loopDepth int (for break/continue validation)
	// - functionScope *FunctionContext (for return type checking)
	// - classScope *ClassContext (for member access validation)
}

// ContextFlags holds structured boolean flags that control parsing behavior.
// This replaces scattered boolean fields throughout the Parser struct.
type ContextFlags struct {
	// EnableSemanticAnalysis enables type checking and semantic validation
	// during parsing (vs. deferring to a separate analysis phase)
	EnableSemanticAnalysis bool

	// ParsingPostCondition is true when parsing postconditions in contracts.
	// This allows the 'old' keyword to reference pre-state values.
	ParsingPostCondition bool

	// Future flag examples (illustrative only, not planned features):
	// - InLoopBody bool (for break/continue validation)
	// - InFunctionBody bool (for return statement validation)
	// - InClassDeclaration bool (for member visibility rules)
}

// NewParseContext creates a new ParseContext with default values.
func NewParseContext() *ParseContext {
	return &ParseContext{
		flags:      ContextFlags{},
		blockStack: []BlockContext{},
	}
}

// NewParseContextWithFlags creates a new ParseContext with specific flag values.
func NewParseContextWithFlags(flags ContextFlags) *ParseContext {
	return &ParseContext{
		flags:      flags,
		blockStack: []BlockContext{},
	}
}

// Flags returns a copy of the current context flags.
// This prevents external modification of internal state.
func (ctx *ParseContext) Flags() ContextFlags {
	return ctx.flags
}

// SetFlags updates the context flags.
func (ctx *ParseContext) SetFlags(flags ContextFlags) {
	ctx.flags = flags
}

// EnableSemanticAnalysis returns whether semantic analysis is enabled.
func (ctx *ParseContext) EnableSemanticAnalysis() bool {
	return ctx.flags.EnableSemanticAnalysis
}

// SetEnableSemanticAnalysis sets the semantic analysis flag.
func (ctx *ParseContext) SetEnableSemanticAnalysis(enable bool) {
	ctx.flags.EnableSemanticAnalysis = enable
}

// ParsingPostCondition returns whether we're currently parsing a postcondition.
func (ctx *ParseContext) ParsingPostCondition() bool {
	return ctx.flags.ParsingPostCondition
}

// SetParsingPostCondition sets the postcondition parsing flag.
func (ctx *ParseContext) SetParsingPostCondition(parsing bool) {
	ctx.flags.ParsingPostCondition = parsing
}

// PushBlock adds a new block context to the stack.
// This should be called when entering a block-level construct (begin, if, while, etc.)
//
// Example:
//
//	ctx.PushBlock("if", p.curToken.Pos)
//	defer ctx.PopBlock()
func (ctx *ParseContext) PushBlock(blockType string, startPos lexer.Position) {
	ctx.blockStack = append(ctx.blockStack, BlockContext{
		BlockType: blockType,
		StartPos:  startPos,
		StartLine: startPos.Line,
	})
}

// PopBlock removes the most recent block context from the stack.
// Should be called when exiting a block (typically via defer).
func (ctx *ParseContext) PopBlock() {
	if len(ctx.blockStack) > 0 {
		ctx.blockStack = ctx.blockStack[:len(ctx.blockStack)-1]
	}
}

// CurrentBlock returns the current block context, or nil if not in any block.
func (ctx *ParseContext) CurrentBlock() *BlockContext {
	if len(ctx.blockStack) == 0 {
		return nil
	}
	return &ctx.blockStack[len(ctx.blockStack)-1]
}

// BlockDepth returns the current block nesting depth.
// Useful for validation and debugging.
func (ctx *ParseContext) BlockDepth() int {
	return len(ctx.blockStack)
}

// BlockStack returns a copy of the current block stack.
// This is useful for error reporting and debugging.
func (ctx *ParseContext) BlockStack() []BlockContext {
	// Return a copy to prevent external modification
	stack := make([]BlockContext, len(ctx.blockStack))
	copy(stack, ctx.blockStack)
	return stack
}

// Snapshot creates a deep copy of the current context.
// This is useful for speculative parsing and backtracking.
//
// Example:
//
//	snapshot := ctx.Snapshot()
//	// ... attempt to parse ...
//	if failed {
//	    ctx.Restore(snapshot)
//	}
func (ctx *ParseContext) Snapshot() *ParseContext {
	// Create a copy with the same flags
	snapshot := &ParseContext{
		flags: ctx.flags, // ContextFlags is a value type, so this is a copy
	}

	// Deep copy the block stack
	snapshot.blockStack = make([]BlockContext, len(ctx.blockStack))
	copy(snapshot.blockStack, ctx.blockStack)

	return snapshot
}

// Restore restores the context to a previously saved snapshot.
// This is used for backtracking after speculative parsing fails.
func (ctx *ParseContext) Restore(snapshot *ParseContext) {
	ctx.flags = snapshot.flags

	// Replace block stack with snapshot's stack
	ctx.blockStack = make([]BlockContext, len(snapshot.blockStack))
	copy(ctx.blockStack, snapshot.blockStack)
}

// Clone creates a deep copy of the context.
// Alias for Snapshot() for clarity in some use cases.
func (ctx *ParseContext) Clone() *ParseContext {
	return ctx.Snapshot()
}

// Reset resets the context to its initial state.
// This clears all flags and empties the block stack.
func (ctx *ParseContext) Reset() {
	ctx.flags = ContextFlags{}
	ctx.blockStack = []BlockContext{}
}

// WithBlock executes a function within a block context.
// This is a convenience method that handles push/pop automatically.
//
// Example:
//
//	err := ctx.WithBlock("if", p.curToken.Pos, func() error {
//	    // ... parse if statement ...
//	    return nil
//	})
func (ctx *ParseContext) WithBlock(blockType string, startPos lexer.Position, fn func() error) error {
	ctx.PushBlock(blockType, startPos)
	defer ctx.PopBlock()
	return fn()
}
