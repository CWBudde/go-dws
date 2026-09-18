package token

// HintLevel is the source-level {$HINTS} setting captured at a token's position.
// The zero value uses the compiler configuration, including after {$HINTS ON}.
type HintLevel uint8

// Source-level hint settings, in the order {$HINTS} can select them.
const (
	// HintLevelDefault defers to the compiler configuration.
	HintLevelDefault HintLevel = iota
	// HintLevelDisabled suppresses all hints ({$HINTS OFF}).
	HintLevelDisabled
	// HintLevelNormal reports normal hints ({$HINTS NORMAL}).
	HintLevelNormal
	// HintLevelStrict adds strict hints ({$HINTS STRICT}).
	HintLevelStrict
	// HintLevelPedantic adds pedantic hints such as case mismatches ({$HINTS PEDANTIC}).
	HintLevelPedantic
)
