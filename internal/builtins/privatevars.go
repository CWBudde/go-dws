package builtins

import (
	"sync"
	"time"

	"github.com/cwbudde/go-dws/pkg/ident"
)

// PrivateVarStore holds process-wide Variant values partitioned by declaring
// unit. Unit identifiers are case-insensitive; variable names are case-sensitive.
// Each partition uses the global-variable store's expiration and mask behavior,
// but is separate from the public global variables and queues.
//
// The zero value is not usable; call NewPrivateVarStore. Script bindings reject
// main-module access before calling this store.
type PrivateVarStore struct {
	now   func() time.Time
	units map[string]*GlobalVarStore
	mu    sync.RWMutex
}

// NewPrivateVarStore creates an empty private store using the wall clock.
func NewPrivateVarStore() *PrivateVarStore {
	return &PrivateVarStore{
		now:   time.Now,
		units: make(map[string]*GlobalVarStore),
	}
}

// DefaultPrivateVars is the process-wide private store used by script builtins.
// Tests should create isolated stores instead of replacing this instance.
var DefaultPrivateVars = NewPrivateVarStore()

// SetClock replaces the clock for both existing and subsequently created units.
// A nil clock restores time.Now. This supports deterministic expiration tests.
func (s *PrivateVarStore) SetClock(clock func() time.Time) {
	if clock == nil {
		clock = time.Now
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = clock
	for _, unit := range s.units {
		unit.SetClock(clock)
	}
}

// unitStore returns the unit's partition, optionally creating it for a write.
func (s *PrivateVarStore) unitStore(unit string, create bool) *GlobalVarStore {
	key := ident.Normalize(unit)
	s.mu.RLock()
	store := s.units[key]
	s.mu.RUnlock()
	if store != nil || !create {
		return store
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if store = s.units[key]; store == nil {
		store = NewGlobalVarStore()
		store.SetClock(s.now)
		s.units[key] = store
	}
	return store
}

// Write stores value and reports whether the unit's entry was absent or expired.
// A zero or negative expiration keeps the value indefinitely.
func (s *PrivateVarStore) Write(unit, name string, value GlobalVarValue, expireSeconds float64) bool {
	return s.unitStore(unit, true).WriteWithResult(name, value, expireSeconds)
}

// Read returns a unit's value and whether it exists and has not expired.
func (s *PrivateVarStore) Read(unit, name string) (GlobalVarValue, bool) {
	if store := s.unitStore(unit, false); store != nil {
		return store.Read(name)
	}
	return GlobalVarValue{Kind: GlobalVarUnassigned}, false
}

// Names returns sorted live names matching mask in the unit. An empty mask
// means all names, and wildcard matching is case-insensitive.
func (s *PrivateVarStore) Names(unit, mask string) []string {
	if store := s.unitStore(unit, false); store != nil {
		return store.Names(mask)
	}
	return []string{}
}

// Cleanup removes names matching mask from the unit. Unlike global cleanup,
// an empty mask removes only the empty name. Callers supply "*" to remove all.
func (s *PrivateVarStore) Cleanup(unit, mask string) {
	if store := s.unitStore(unit, false); store != nil {
		if mask == "" {
			store.Delete("")
			return
		}
		store.Cleanup(mask)
	}
}
