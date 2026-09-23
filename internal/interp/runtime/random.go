package runtime

// DefaultRandSeed is DWScript's cDefaultRandSeed. Every script-visible seed is
// stored xor-ed with it (SetRandSeed/RandSeed), and it replaces a zero state,
// which xorshift can never leave.
const DefaultRandSeed uint64 = 88172645463325252

// XorShift is DWScript's per-execution random generator: Marsaglia's 64-bit
// xorshift (13, 17, 5), as in TdwsExecution.Random. Seeded scripts depend on
// the exact sequence, so Random, RandomInt and RandG all draw from it.
//
// Upstream seeds a new execution from the host RNG; go-dws starts from
// DefaultRandSeed instead (the state SetRandSeed(0) produces) so an unseeded
// run stays reproducible.
type XorShift struct {
	state uint64
}

// NewXorShift returns a generator in its default state.
func NewXorShift() *XorShift {
	return &XorShift{state: DefaultRandSeed}
}

// State returns the raw generator state.
func (x *XorShift) State() uint64 {
	return x.state
}

// SetState replaces the raw generator state; zero selects DefaultRandSeed.
func (x *XorShift) SetState(state uint64) {
	if state == 0 {
		state = DefaultRandSeed
	}
	x.state = state
}

// Float64 advances the generator and returns a value in [0, 1).
func (x *XorShift) Float64() float64 {
	buf := x.state
	if buf == 0 {
		buf = DefaultRandSeed
	} else {
		buf ^= buf << 13
		buf ^= buf >> 17
		buf ^= buf << 5
	}
	x.state = buf
	return float64(buf>>1) * (1.0 / (1 << 63))
}
