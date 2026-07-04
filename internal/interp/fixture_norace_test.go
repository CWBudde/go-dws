//go:build !race

package interp

// fixtureTimeoutScale is 1 in normal builds; race-detector builds stretch the
// per-fixture timeout (see fixture_race_test.go).
const fixtureTimeoutScale = 1
