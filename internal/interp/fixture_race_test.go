//go:build race

package interp

// fixtureTimeoutScale stretches the per-fixture timeout under the race
// detector, which slows compute-heavy fixtures well past the normal budget.
const fixtureTimeoutScale = 5
