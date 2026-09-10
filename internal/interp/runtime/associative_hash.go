package runtime

import (
	"math"
	"math/bits"
	"sort"
	"sync/atomic"
	"unicode/utf16"
)

// This file ports the key-hashing half of DWScript's associative-array
// implementation (`TScriptAssociativeArray` in `Source/dwsAssociativeArrays.pas`,
// hashing in `Source/dwsDataContext.pas` and `Source/dwsUtils.pas`).
//
// The hash is observable: `.Keys` walks the bucket array in index order, so the
// order in which an associative array yields its keys is a deterministic
// function of these hash values. Fixtures such as
// `testdata/fixtures/AssociativePass/records.pas` pin that order, which is why
// the exact upstream hash — not merely *a* hash — has to be reproduced.

// FNV-1a mixing constants used by DWScript's DWSHashCode.
const (
	dwsFNVBasis = uint32(2166136261)
	dwsFNVPrime = uint32(16777619)
)

// xxHash32 primes.
const (
	xxPrime1 = uint32(2654435761)
	xxPrime2 = uint32(2246822519)
	xxPrime3 = uint32(3266489917)
	xxPrime4 = uint32(668265263)
	xxPrime5 = uint32(374761393)
)

// xxHash32 computes the 32-bit xxHash of data with the given seed.
//
// DWScript hashes strings with `xxHash32.Full(p, byteLength)` (seed 0) over the
// raw UTF-16 code units, so this must match the reference xxHash32 exactly.
func xxHash32(data []byte, seed uint32) uint32 {
	var h uint32
	n := len(data)
	i := 0

	if n >= 16 {
		v1 := seed + xxPrime1 + xxPrime2
		v2 := seed + xxPrime2
		v3 := seed
		v4 := seed - xxPrime1
		for ; i+16 <= n; i += 16 {
			v1 = xxRound(v1, leUint32(data[i:]))
			v2 = xxRound(v2, leUint32(data[i+4:]))
			v3 = xxRound(v3, leUint32(data[i+8:]))
			v4 = xxRound(v4, leUint32(data[i+12:]))
		}
		h = bits.RotateLeft32(v1, 1) + bits.RotateLeft32(v2, 7) +
			bits.RotateLeft32(v3, 12) + bits.RotateLeft32(v4, 18)
	} else {
		h = seed + xxPrime5
	}

	h += uint32(n)

	for ; i+4 <= n; i += 4 {
		h += leUint32(data[i:]) * xxPrime3
		h = bits.RotateLeft32(h, 17) * xxPrime4
	}
	for ; i < n; i++ {
		h += uint32(data[i]) * xxPrime5
		h = bits.RotateLeft32(h, 11) * xxPrime1
	}

	h ^= h >> 15
	h *= xxPrime2
	h ^= h >> 13
	h *= xxPrime3
	h ^= h >> 16
	return h
}

func xxRound(acc, input uint32) uint32 {
	acc += input * xxPrime2
	acc = bits.RotateLeft32(acc, 13)
	acc *= xxPrime1
	return acc
}

func leUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// utf16LEBytes encodes s as little-endian UTF-16 code units, the in-memory
// layout of a Delphi UnicodeString that DWScript feeds to xxHash32.
func utf16LEBytes(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, 0, len(units)*2)
	for _, u := range units {
		out = append(out, byte(u), byte(u>>8))
	}
	return out
}

// simpleStringHash ports dwsUtils.SimpleStringHash(const s : UnicodeString).
func simpleStringHash(s string) uint32 {
	return xxHash32(utf16LEBytes(s), 0)
}

// simpleIntegerHash ports dwsUtils.SimpleIntegerHash (a simplified MurmurHash3).
func simpleIntegerHash(x uint32) uint32 {
	r := x * 0xcc9e2d51
	r = bits.RotateLeft32(r, 15)
	r = r*0x1b873593 + 0xe6546b64
	if r == 0 {
		r = 1
	}
	return r
}

// simpleInt64Hash ports dwsUtils.SimpleInt64Hash (a simplified MurmurHash3).
func simpleInt64Hash(x int64) uint32 {
	r := uint32(x) * 0xcc9e2d51 //nolint:gosec // intentional 32-bit truncation, matches Delphi
	r = bits.RotateLeft32(r, 15)
	r = r*0x1b873593 + 0xe6546b64

	k := uint32(x>>32) * 0xcc9e2d51 //nolint:gosec // intentional 32-bit truncation
	k = bits.RotateLeft32(k, 15)
	r = k*0x1b873593 ^ r

	if r == 0 {
		r = 1
	}
	return r
}

// nextObjectIdentityHash supplies the per-instance hash used for object keys.
//
// DWScript stores an object key as a `varUnknown` variant and hashes the low 32
// bits of its interface pointer, so upstream's bucket order for object keys is
// heap-address dependent and not reproducible even between two runs of
// DWScript itself. We substitute a stable monotonic identity so that lookups
// are correct and the order is at least deterministic within a run.
var objectIdentityCounter atomic.Uint32

// mixValueHash applies one DWSHashCode accumulation step.
func mixValueHash(partial, elem uint32) uint32 {
	return (partial ^ elem) * dwsFNVPrime
}

// elementHash ports the per-variant branch of dwsDataContext.DWSHashCode.
func elementHash(v Value) uint32 {
	switch val := v.(type) {
	case *IntegerValue:
		// Script integers are varInt64.
		return simpleInt64Hash(val.Value)
	case *FloatValue:
		// varDouble hashes the raw 64-bit pattern.
		return simpleInt64Hash(int64(math.Float64bits(val.Value))) //nolint:gosec // reinterpreting bits, matches Delphi
	case *StringValue:
		return simpleStringHash(val.Value)
	case *BooleanValue:
		// varBoolean is a Delphi WordBool (True = $FFFF) read through VByte,
		// so True hashes as 255 and False as 0.
		if val.Value {
			return simpleIntegerHash(255)
		}
		return simpleIntegerHash(0)
	case *EnumValue:
		// Enums are ordinals at runtime.
		return simpleInt64Hash(int64(val.OrdinalValue))
	case *VariantValue:
		if val.Value == nil {
			return simpleIntegerHash(0)
		}
		return elementHash(val.Value)
	case *ObjectInstance:
		return simpleIntegerHash(val.associativeIdentity())
	case *NilValue, nil:
		return simpleIntegerHash(0)
	}
	// Records and static arrays are multi-slot keys: DWScript chains the hash
	// over their data slots. Anything else falls back to its textual form,
	// which is stable within a run.
	if flat, ok := flattenKeyElements(v); ok {
		h := dwsFNVBasis
		for _, e := range flat {
			h = mixValueHash(h, elementHash(e))
		}
		return h
	}
	return simpleStringHash(v.String())
}

// associativeHashCode ports dwsDataContext.DWSHashCode for a whole key.
//
// A hash code of 0 marks an empty bucket, so upstream substitutes the FNV basis
// whenever the computed hash lands on 0.
func associativeHashCode(key Value) uint32 {
	var h uint32
	if flat, ok := flattenKeyElements(key); ok {
		h = dwsFNVBasis
		for _, e := range flat {
			h = mixValueHash(h, elementHash(e))
		}
	} else {
		h = mixValueHash(dwsFNVBasis, elementHash(key))
	}
	if h == 0 {
		h = dwsFNVBasis
	}
	return h
}

// flattenKeyElements returns the data slots of a multi-slot key (record or
// static array) in a deterministic order, or ok=false for single-slot keys.
//
// Upstream walks the record's fields in declaration order. Our RecordType keeps
// fields in a map with no declaration order, so record keys are flattened in
// sorted field order instead: deterministic, but a divergence from upstream's
// bucket order for record-keyed associative arrays. No fixture pins that order.
func flattenKeyElements(v Value) ([]Value, bool) {
	switch val := v.(type) {
	case *RecordValue:
		names := make([]string, 0, len(val.Fields))
		for n := range val.Fields {
			names = append(names, n)
		}
		sort.Strings(names)
		out := make([]Value, 0, len(names))
		for _, n := range names {
			out = append(out, val.Fields[n])
		}
		return out, true
	case *ArrayValue:
		if val.ArrayType == nil || val.ArrayType.IsDynamic() {
			return nil, false
		}
		return val.Elements, true
	}
	return nil, false
}

// associativeIdentity returns this instance's stable identity hash, assigning
// one on first use. Object keys are matched by pointer identity, so the hash
// only has to be stable per instance.
func (o *ObjectInstance) associativeIdentity() uint32 {
	if o.assocIdentity == 0 {
		o.assocIdentity = objectIdentityCounter.Add(1)
	}
	return o.assocIdentity
}
