//go:build !race

package interp

// fixtureTimeoutScale gives normal package-parallel test runs enough room for
// compute-heavy fixtures while keeping hung fixtures bounded.
const fixtureTimeoutScale = 4
