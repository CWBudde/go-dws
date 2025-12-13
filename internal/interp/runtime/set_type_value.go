package runtime

import "github.com/cwbudde/go-dws/internal/types"

// SetTypeValue stores set type metadata in the environment.
// It is a runtime-level "type meta value" to support resolving named set types
// (e.g., for typed empty set literals like `s := []`).
//
// Stored under: __set_type_<name> (normalized).
type SetTypeValue struct {
	Name    string
	SetType *types.SetType
}

func (s *SetTypeValue) Type() string { return "SET_TYPE" }

func (s *SetTypeValue) String() string { return s.Name }

func (s *SetTypeValue) GetSetType() *types.SetType { return s.SetType }
