package builtins

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// GlobalVars / GlobalQueues Built-in Library
// ============================================================================
//
// This file implements DWScript's dwsGlobalVarsFunctions library: a
// process-wide, thread-safe store of named Variant values (global variables)
// plus a set of named double-ended queues (global queues).
//
// The store is shared by every script executed in the host process, which
// mirrors the original Delphi implementation. All access is guarded by a
// sync.RWMutex.
//
// Name lookup is case-SENSITIVE, matching DWScript: writing 'hello' and
// 'Hello' creates two distinct globals. Wildcard masks, in contrast, are
// matched case-INSENSITIVELY (GlobalVarsNames('h*') returns both 'Hello' and
// 'hello'), so mask comparisons go through pkg/ident.

// GlobalVarKind enumerates the value shapes a global variable or queue entry
// can hold. DWScript only allows simple Variant payloads to be stored.
type GlobalVarKind int

const (
	// GlobalVarUnassigned is the Unassigned (varEmpty) Variant.
	GlobalVarUnassigned GlobalVarKind = iota
	// GlobalVarNull is the Null (varNull) Variant.
	GlobalVarNull
	// GlobalVarInteger holds a 64-bit signed integer.
	GlobalVarInteger
	// GlobalVarFloat holds a double-precision float.
	GlobalVarFloat
	// GlobalVarString holds a Unicode string.
	GlobalVarString
	// GlobalVarBoolean holds a boolean.
	GlobalVarBoolean
)

// GlobalVarValue is the immutable, storable representation of a Variant held
// by the global variable store. Using a plain value type (rather than a
// runtime.Value pointer) keeps the shared store free of aliasing between
// concurrent scripts and makes serialization straightforward.
type GlobalVarValue struct {
	// Str holds the payload for GlobalVarString.
	Str string
	// Int holds the payload for GlobalVarInteger.
	Int int64
	// Float holds the payload for GlobalVarFloat.
	Float float64
	// Kind selects which payload field is meaningful.
	Kind GlobalVarKind
	// Bool holds the payload for GlobalVarBoolean.
	Bool bool
}

// IsUnassigned reports whether the value is the Unassigned Variant.
func (v GlobalVarValue) IsUnassigned() bool { return v.Kind == GlobalVarUnassigned }

// globalVarEntry pairs a stored value with its optional expiration instant.
type globalVarEntry struct {
	// expires is the instant after which the entry is considered absent.
	// The zero time means the entry never expires.
	expires time.Time
	value   GlobalVarValue
}

// GlobalVarStore is a process-wide, mutex-guarded store of named Variant
// globals and named double-ended queues.
//
// The zero value is not usable; call NewGlobalVarStore. A default instance is
// available as DefaultGlobalVars; tests should create their own store to stay
// isolated.
type GlobalVarStore struct {
	// now supplies the current time. It is a field so that tests can inject a
	// deterministic clock instead of time.Now.
	now    func() time.Time
	vars   map[string]*globalVarEntry
	queues map[string][]GlobalVarValue
	mu     sync.RWMutex
}

// NewGlobalVarStore creates an empty store backed by the real wall clock.
func NewGlobalVarStore() *GlobalVarStore {
	return &GlobalVarStore{
		now:    time.Now,
		vars:   make(map[string]*globalVarEntry),
		queues: make(map[string][]GlobalVarValue),
	}
}

// SetClock replaces the store's time source. It exists for deterministic
// testing of expiration; production code leaves the default time.Now in place.
func (s *GlobalVarStore) SetClock(clock func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if clock == nil {
		clock = time.Now
	}
	s.now = clock
}

// DefaultGlobalVars is the process-wide store used by the built-in functions.
var DefaultGlobalVars = NewGlobalVarStore()

// expiryInstant converts a duration in seconds into an absolute instant.
// A non-positive duration means "never expires" and yields the zero time.
func (s *GlobalVarStore) expiryInstant(seconds float64) time.Time {
	if seconds <= 0 || math.IsNaN(seconds) {
		return time.Time{}
	}
	return s.now().Add(time.Duration(seconds * float64(time.Second)))
}

// liveLocked returns the entry for name if it exists and has not expired.
// Expired entries are dropped. The caller must hold the write lock.
func (s *GlobalVarStore) liveLocked(name string) (*globalVarEntry, bool) {
	entry, ok := s.vars[name]
	if !ok {
		return nil, false
	}
	if !entry.expires.IsZero() && !s.now().Before(entry.expires) {
		delete(s.vars, name)
		return nil, false
	}
	return entry, true
}

// Write stores value under name, replacing any previous value.
// expireSeconds sets a lifetime in seconds; zero or negative means the value
// never expires. Writing always replaces the previous expiration.
func (s *GlobalVarStore) Write(name string, value GlobalVarValue, expireSeconds float64) {
	s.WriteWithResult(name, value, expireSeconds)
}

// WriteWithResult stores value and reports whether name was absent or expired.
// Checking the old entry and replacing it happen atomically. The new expiration
// replaces any previous expiration, with zero or negative meaning no expiration.
func (s *GlobalVarStore) WriteWithResult(name string, value GlobalVarValue, expireSeconds float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, existed := s.liveLocked(name)
	s.vars[name] = &globalVarEntry{value: value, expires: s.expiryInstant(expireSeconds)}
	return !existed
}

// Read returns the value stored under name and whether it was present.
// Expired entries report as absent.
func (s *GlobalVarStore) Read(name string) (GlobalVarValue, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.liveLocked(name)
	if !ok {
		return GlobalVarValue{Kind: GlobalVarUnassigned}, false
	}
	return entry.value, true
}

// Delete removes name from the store and reports whether it was present.
func (s *GlobalVarStore) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, existed := s.liveLocked(name)
	delete(s.vars, name)
	return existed
}

// Cleanup removes every global whose name matches mask (case-insensitive
// wildcards, '*' and '?'). An empty mask is treated as '*'.
func (s *GlobalVarStore) Cleanup(mask string) {
	if mask == "" {
		mask = "*"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if mask == "*" {
		s.vars = make(map[string]*globalVarEntry)
		return
	}
	for name := range s.vars {
		if matchesMask(name, mask) {
			delete(s.vars, name)
		}
	}
}

// Names returns the sorted names of all live globals matching mask.
// An empty mask is treated as '*'.
func (s *GlobalVarStore) Names(mask string) []string {
	if mask == "" {
		mask = "*"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.vars))
	for name := range s.vars {
		if _, ok := s.liveLocked(name); !ok {
			continue
		}
		if matchesMask(name, mask) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// Increment atomically adds delta to the integer stored under name and returns
// the new value. A missing, expired or non-integer value is treated as zero.
// The expiration is always reset from expireSeconds, so incrementing without an
// explicit lifetime clears any previous one.
func (s *GlobalVarStore) Increment(name string, delta int64, expireSeconds float64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var base int64
	if entry, ok := s.liveLocked(name); ok {
		switch entry.value.Kind {
		case GlobalVarInteger:
			base = entry.value.Int
		case GlobalVarFloat:
			base = int64(entry.value.Float)
		case GlobalVarBoolean:
			if entry.value.Bool {
				base = 1
			}
		case GlobalVarUnassigned, GlobalVarNull, GlobalVarString:
			base = 0
		}
	}

	result := base + delta
	s.vars[name] = &globalVarEntry{
		value:   GlobalVarValue{Kind: GlobalVarInteger, Int: result},
		expires: s.expiryInstant(expireSeconds),
	}
	return result
}

// CompareExchange atomically replaces the value stored under name with value
// when the current value equals comparand, and returns the previous value.
// A missing entry compares equal to Unassigned.
func (s *GlobalVarStore) CompareExchange(name string, value, comparand GlobalVarValue) GlobalVarValue {
	s.mu.Lock()
	defer s.mu.Unlock()

	previous := GlobalVarValue{Kind: GlobalVarUnassigned}
	if entry, ok := s.liveLocked(name); ok {
		previous = entry.value
	}
	if globalVarEqual(previous, comparand) {
		s.vars[name] = &globalVarEntry{value: value}
	}
	return previous
}

// matchesMask reports whether name matches a DWScript wildcard mask.
// Matching is case-insensitive, as DWScript compares global names without
// regard to case when filtering.
func matchesMask(name, mask string) bool {
	if mask == "*" {
		return true
	}
	return wildcardMatch(ident.Normalize(name), ident.Normalize(mask))
}

// globalVarEqual compares two stored Variants with DWScript's Variant
// equality: numeric kinds compare numerically, everything else compares by
// kind and payload.
func globalVarEqual(a, b GlobalVarValue) bool {
	// Two integers compare exactly. Routing them through float64 would make
	// distinct 64-bit values above 2^53 collapse onto the same float, which
	// would break CompareExchange as a synchronization primitive.
	if a.Kind == GlobalVarInteger && b.Kind == GlobalVarInteger {
		return a.Int == b.Int
	}
	if isNumericKind(a.Kind) && isNumericKind(b.Kind) {
		return numericValue(a) == numericValue(b)
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case GlobalVarString:
		return a.Str == b.Str
	case GlobalVarBoolean:
		return a.Bool == b.Bool
	case GlobalVarUnassigned, GlobalVarNull:
		return true
	case GlobalVarInteger, GlobalVarFloat:
		return numericValue(a) == numericValue(b)
	}
	return false
}

// isNumericKind reports whether the kind carries a numeric payload.
func isNumericKind(k GlobalVarKind) bool {
	return k == GlobalVarInteger || k == GlobalVarFloat
}

// numericValue returns the numeric payload of a numeric kind as a float64.
func numericValue(v GlobalVarValue) float64 {
	if v.Kind == GlobalVarInteger {
		return float64(v.Int)
	}
	return v.Float
}

// ============================================================================
// Global queues
// ============================================================================

// QueuePush appends value to the back of the named queue.
func (s *GlobalVarStore) QueuePush(name string, value GlobalVarValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queues[name] = append(s.queues[name], value)
}

// QueueInsert prepends value to the front of the named queue.
func (s *GlobalVarStore) QueueInsert(name string, value GlobalVarValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := s.queues[name]
	updated := make([]GlobalVarValue, 0, len(queue)+1)
	updated = append(updated, value)
	updated = append(updated, queue...)
	s.queues[name] = updated
}

// QueuePull removes and returns the value at the front of the named queue.
func (s *GlobalVarStore) QueuePull(name string) (GlobalVarValue, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := s.queues[name]
	if len(queue) == 0 {
		return GlobalVarValue{}, false
	}
	value := queue[0]
	// Clear the consumed slot so the backing array does not retain the
	// payload: queues are process-wide and can be long-lived.
	queue[0] = GlobalVarValue{}
	s.queues[name] = queue[1:]
	return value, true
}

// QueuePop removes and returns the value at the back of the named queue.
func (s *GlobalVarStore) QueuePop(name string) (GlobalVarValue, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue := s.queues[name]
	if len(queue) == 0 {
		return GlobalVarValue{}, false
	}
	value := queue[len(queue)-1]
	queue[len(queue)-1] = GlobalVarValue{}
	s.queues[name] = queue[:len(queue)-1]
	return value, true
}

// QueuePeek returns the value at the back of the named queue without removing it.
func (s *GlobalVarStore) QueuePeek(name string) (GlobalVarValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	queue := s.queues[name]
	if len(queue) == 0 {
		return GlobalVarValue{}, false
	}
	return queue[len(queue)-1], true
}

// QueueFirst returns the value at the front of the named queue without removing it.
func (s *GlobalVarStore) QueueFirst(name string) (GlobalVarValue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	queue := s.queues[name]
	if len(queue) == 0 {
		return GlobalVarValue{}, false
	}
	return queue[0], true
}

// QueueLength returns the number of entries in the named queue.
func (s *GlobalVarStore) QueueLength(name string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.queues[name])
}

// QueueSnapshot returns a copy of the named queue's contents, front first.
func (s *GlobalVarStore) QueueSnapshot(name string) []GlobalVarValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	queue := s.queues[name]
	snapshot := make([]GlobalVarValue, len(queue))
	copy(snapshot, queue)
	return snapshot
}

// CleanupQueues removes every queue whose name matches mask (case-insensitive
// wildcards). An empty mask is treated as '*'.
func (s *GlobalVarStore) CleanupQueues(mask string) {
	if mask == "" {
		mask = "*"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if mask == "*" {
		s.queues = make(map[string][]GlobalVarValue)
		return
	}
	for name := range s.queues {
		if matchesMask(name, mask) {
			delete(s.queues, name)
		}
	}
}

// ============================================================================
// Serialization
// ============================================================================

// globalVarsFileTag is the first line of the SaveGlobalVarsToString payload.
// LoadGlobalVarsFromString rejects any input that does not start with it.
const globalVarsFileTag = "DWSGV1"

// serializedGlobalVar is the on-the-wire shape of one persisted global.
type serializedGlobalVar struct {
	Name string `json:"n"`
	Str  string `json:"s,omitempty"`
	// FloatSpecial carries NaN and +/-Infinity, which JSON cannot express as
	// numbers. When set it takes precedence over Float on restore.
	FloatSpecial string `json:"fs,omitempty"`
	// TTL is the remaining lifetime in seconds at save time; 0 means no expiry.
	TTL   float64       `json:"ttl,omitempty"`
	Int   int64         `json:"i,omitempty"`
	Float float64       `json:"f,omitempty"`
	Kind  GlobalVarKind `json:"k"`
	Bool  bool          `json:"b,omitempty"`
}

// encodeSpecialFloat returns the textual tag for a non-finite float, or an
// empty string for a finite one. encoding/json rejects NaN and +/-Infinity,
// and both are reachable from script via NaN() and Infinity().
func encodeSpecialFloat(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	default:
		return ""
	}
}

// decodeSpecialFloat restores the float encoded by encodeSpecialFloat.
func decodeSpecialFloat(tag string) float64 {
	switch tag {
	case "nan":
		return math.NaN()
	case "inf":
		return math.Inf(1)
	case "-inf":
		return math.Inf(-1)
	default:
		return 0
	}
}

// SaveToString serializes every live global into a self-describing string that
// LoadFromString can restore. Queues are not part of the snapshot, matching
// DWScript.
func (s *GlobalVarStore) SaveToString() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.vars))
	for name := range s.vars {
		if _, ok := s.liveLocked(name); ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	records := make([]serializedGlobalVar, 0, len(names))
	now := s.now()
	for _, name := range names {
		entry := s.vars[name]
		record := serializedGlobalVar{
			Name:  name,
			Kind:  entry.value.Kind,
			Int:   entry.value.Int,
			Float: entry.value.Float,
			Str:   entry.value.Str,
			Bool:  entry.value.Bool,
		}
		if special := encodeSpecialFloat(entry.value.Float); special != "" {
			record.Float = 0
			record.FloatSpecial = special
		}
		if !entry.expires.IsZero() {
			record.TTL = entry.expires.Sub(now).Seconds()
		}
		records = append(records, record)
	}

	payload, err := json.Marshal(records)
	if err != nil {
		// The record shape only contains JSON-safe scalars, so this cannot
		// fail in practice; degrade to an empty snapshot rather than panic.
		payload = []byte("[]")
	}
	return globalVarsFileTag + "\n" + string(payload)
}

// LoadFromString replaces the entire set of globals with the snapshot encoded
// in data. An empty string clears the store. A payload that does not start
// with the expected tag returns ErrInvalidGlobalVarsFileTag.
func (s *GlobalVarStore) LoadFromString(data string) error {
	if data == "" {
		s.Cleanup("*")
		return nil
	}

	header, payload, found := strings.Cut(data, "\n")
	if !found || header != globalVarsFileTag {
		return ErrInvalidGlobalVarsFileTag
	}

	var records []serializedGlobalVar
	if err := json.Unmarshal([]byte(payload), &records); err != nil {
		return ErrInvalidGlobalVarsFileTag
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars = make(map[string]*globalVarEntry, len(records))
	for _, record := range records {
		floatValue := record.Float
		if record.FloatSpecial != "" {
			floatValue = decodeSpecialFloat(record.FloatSpecial)
		}
		entry := &globalVarEntry{
			value: GlobalVarValue{
				Kind:  record.Kind,
				Int:   record.Int,
				Float: floatValue,
				Str:   record.Str,
				Bool:  record.Bool,
			},
		}
		if record.TTL > 0 {
			entry.expires = s.now().Add(time.Duration(record.TTL * float64(time.Second)))
		}
		s.vars[record.Name] = entry
	}
	return nil
}

// globalVarsError is the error type reported by the GlobalVars library.
// Its message is surfaced verbatim as a catchable script exception.
type globalVarsError string

// Error implements the error interface.
func (e globalVarsError) Error() string { return string(e) }

// ErrInvalidGlobalVarsFileTag is reported by LoadFromString when the payload
// does not carry the expected header. The message matches DWScript's wording.
const ErrInvalidGlobalVarsFileTag = globalVarsError("Invalid file tag")

// ============================================================================
// runtime.Value conversion
// ============================================================================

// ToRuntimeValue converts a stored Variant into the interpreter's runtime
// representation.
func (v GlobalVarValue) ToRuntimeValue() Value {
	switch v.Kind {
	case GlobalVarInteger:
		return &runtime.IntegerValue{Value: v.Int}
	case GlobalVarFloat:
		return &runtime.FloatValue{Value: v.Float}
	case GlobalVarString:
		return &runtime.StringValue{Value: v.Str}
	case GlobalVarBoolean:
		return &runtime.BooleanValue{Value: v.Bool}
	case GlobalVarNull:
		return &runtime.NullValue{}
	case GlobalVarUnassigned:
		return &runtime.UnassignedValue{}
	}
	return &runtime.UnassignedValue{}
}

// dwsSymbolName maps a runtime value kind onto the Delphi symbol class name
// DWScript reports when a value cannot be stored in a global.
func dwsSymbolName(value Value) string {
	switch runtime.KindOf(value) {
	case runtime.KindInterface:
		return "TInterfaceSymbol"
	case runtime.KindObject, runtime.KindClass, runtime.KindClassInfo:
		return "TClassSymbol"
	case runtime.KindRecord, runtime.KindRecordType:
		return "TRecordSymbol"
	case runtime.KindArray:
		return "TDynamicArraySymbol"
	case runtime.KindAssociativeArray:
		return "TAssociativeArraySymbol"
	case runtime.KindSet, runtime.KindSetType:
		return "TSetOfSymbol"
	case runtime.KindFunctionPointer, runtime.KindMethodPointer, runtime.KindLambda:
		return "TFuncSymbol"
	}
	return value.Type()
}

// FromRuntimeValue converts an interpreter value into a storable Variant.
// It returns an error naming the DWScript symbol class when the value's type
// cannot be held by a global.
func FromRuntimeValue(value Value) (GlobalVarValue, error) {
	if value == nil {
		return GlobalVarValue{Kind: GlobalVarUnassigned}, nil
	}
	if variant, ok := value.(*runtime.VariantValue); ok {
		return FromRuntimeValue(variant.UnwrapVariant())
	}

	switch typed := value.(type) {
	case *runtime.IntegerValue:
		return GlobalVarValue{Kind: GlobalVarInteger, Int: typed.Value}, nil
	case *runtime.FloatValue:
		return GlobalVarValue{Kind: GlobalVarFloat, Float: typed.Value}, nil
	case *runtime.StringValue:
		return GlobalVarValue{Kind: GlobalVarString, Str: typed.Value}, nil
	case *runtime.BooleanValue:
		return GlobalVarValue{Kind: GlobalVarBoolean, Bool: typed.Value}, nil
	case *runtime.NullValue, *runtime.NilValue:
		return GlobalVarValue{Kind: GlobalVarNull}, nil
	case *runtime.UnassignedValue:
		return GlobalVarValue{Kind: GlobalVarUnassigned}, nil
	case *runtime.InterfaceInstance:
		// A nil interface reference degrades to Null, exactly as DWScript's
		// IUnknown-to-Variant conversion does.
		if typed == nil || typed.Object == nil {
			return GlobalVarValue{Kind: GlobalVarNull}, nil
		}
	}

	return GlobalVarValue{}, globalVarsError("Cannot store global of type " + dwsSymbolName(value))
}
