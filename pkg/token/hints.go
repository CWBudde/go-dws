package token

// HintLevel is the source-level {$HINTS} setting captured at a token's position.
// The zero value uses the compiler configuration, including after {$HINTS ON}.
type HintLevel uint8

const (
	HintLevelDefault HintLevel = iota
	HintLevelDisabled
	HintLevelNormal
	HintLevelStrict
	HintLevelPedantic
)
