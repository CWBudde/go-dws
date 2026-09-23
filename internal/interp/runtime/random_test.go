package runtime

import "testing"

// TestXorShift_UpstreamSequence pins the sequence of FunctionsMath/randseed:
// SetRandSeed(1234) followed by three RandomInt(1000000) calls.
func TestXorShift_UpstreamSequence(t *testing.T) {
	rng := NewXorShift()
	rng.SetState(1234 ^ DefaultRandSeed)
	if got := rng.State() ^ DefaultRandSeed; got != 1234 {
		t.Fatalf("seed round trip: got %d", got)
	}
	for i, want := range []int64{47687, 125347, 903103} {
		if got := int64(rng.Float64() * 1000000); got != want {
			t.Fatalf("draw %d: want %d, got %d", i, want, got)
		}
	}
}

func TestXorShift_ZeroStateSelectsDefault(t *testing.T) {
	rng := NewXorShift()
	rng.SetState(0)
	if rng.State() != DefaultRandSeed {
		t.Fatalf("zero state: got %d", rng.State())
	}
	for range 1000 {
		if f := rng.Float64(); f < 0 || f >= 1 {
			t.Fatalf("out of range: %v", f)
		}
	}
}
