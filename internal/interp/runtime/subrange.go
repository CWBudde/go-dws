// Package runtime provides runtime value types for the DWScript interpreter.
// This file contains SubrangeValue, the runtime representation of subrange types.
package runtime

import (
	"fmt"

	"github.com/cwbudde/go-dws/internal/types"
)

// SubrangeValue wraps an integer value with subrange bounds checking.
// Subrange types are integer types constrained to a specific range.
//
// Task 3.5.19: Moved from internal/interp/type_alias.go to runtime package
// to enable evaluator package to work with subrange values directly without
// circular imports.
//
// Example DWScript:
//
//	type TDigit = 0..9;
//	var d: TDigit;
//	d := 5;  // OK
//	d := 10; // Runtime error: value out of range
type SubrangeValue struct {
	SubrangeType *types.SubrangeType
	Value        int
}

// Type returns the subrange type name.
func (sv *SubrangeValue) Type() string {
	return sv.SubrangeType.Name
}

// String returns the integer value as a string.
func (sv *SubrangeValue) String() string {
	return fmt.Sprintf("%d", sv.Value)
}

// ValidateAndSet checks if a value is within bounds and updates the subrange value.
// Returns an error if the value is out of range.
func (sv *SubrangeValue) ValidateAndSet(value int) error {
	if err := types.ValidateRange(value, sv.SubrangeType); err != nil {
		return err
	}
	sv.Value = value
	return nil
}

// GetValue returns the current integer value.
// Task 3.5.19: Added to satisfy SubrangeValueAccessor interface in evaluator.
func (sv *SubrangeValue) GetValue() int {
	return sv.Value
}

// GetTypeName returns the subrange type name.
// Task 3.5.19: Added to satisfy SubrangeValueAccessor interface in evaluator.
func (sv *SubrangeValue) GetTypeName() string {
	return sv.SubrangeType.Name
}

// NewSubrangeValue creates a new SubrangeValue with the given type and initial value.
// Returns an error if the initial value is out of range.
//
// Task 3.5.19: Constructor added to enable direct creation in evaluator.
func NewSubrangeValue(subrangeType *types.SubrangeType, initialValue int) (*SubrangeValue, error) {
	if err := types.ValidateRange(initialValue, subrangeType); err != nil {
		return nil, err
	}
	return &SubrangeValue{
		SubrangeType: subrangeType,
		Value:        initialValue,
	}, nil
}

// NewSubrangeValueZero creates a new SubrangeValue initialized to the low bound.
// This is the default initialization for subrange variables.
//
// Task 3.5.19: Constructor for default initialization.
func NewSubrangeValueZero(subrangeType *types.SubrangeType) *SubrangeValue {
	return &SubrangeValue{
		SubrangeType: subrangeType,
		Value:        subrangeType.LowBound,
	}
}
