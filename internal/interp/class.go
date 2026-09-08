package interp

import "github.com/cwbudde/go-dws/internal/interp/runtime"

// ClassInfo stores runtime class metadata.
type ClassInfo = runtime.ClassInfo

// ClassValue represents a metaclass reference.
type ClassValue = runtime.ClassValue

// VirtualMethodEntry stores a class virtual method implementation.
type VirtualMethodEntry = runtime.ClassVirtualMethodEntry

// NewClassInfo constructs initialized runtime class metadata.
func NewClassInfo(name string) *ClassInfo { return runtime.NewClassInfo(name) }

// AsClassValue returns the metaclass reference when v is a class value.
func AsClassValue(v Value) (*ClassValue, bool) { return runtime.AsClassValue(v) }
