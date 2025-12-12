package runtime

import "sort"

// ArrayHelperCopy creates a shallow copy of an array's elements.
func ArrayHelperCopy(arr *ArrayValue) Value {
	newArray := &ArrayValue{
		ArrayType: arr.ArrayType,
		Elements:  make([]Value, len(arr.Elements)),
	}
	copy(newArray.Elements, arr.Elements)
	return newArray
}

// ArrayHelperIndexOf searches an array for a value starting from startIndex.
// Returns the 0-based index (>= 0) or -1 if not found.
func ArrayHelperIndexOf(arr *ArrayValue, value Value, startIndex int) Value {
	if startIndex < 0 || startIndex >= len(arr.Elements) {
		return &IntegerValue{Value: -1}
	}
	for idx := startIndex; idx < len(arr.Elements); idx++ {
		if ValuesEqual(arr.Elements[idx], value) {
			return &IntegerValue{Value: int64(idx)}
		}
	}
	return &IntegerValue{Value: -1}
}

// ArrayHelperReverse reverses an array in place.
func ArrayHelperReverse(arr *ArrayValue) Value {
	elements := arr.Elements
	n := len(elements)
	for left := 0; left < n/2; left++ {
		right := n - 1 - left
		elements[left], elements[right] = elements[right], elements[left]
	}
	return &NilValue{}
}

// ArrayHelperSort sorts an array in place.
func ArrayHelperSort(arr *ArrayValue) Value {
	elements := arr.Elements
	if len(elements) <= 1 {
		return &NilValue{}
	}

	switch elements[0].(type) {
	case *IntegerValue:
		sort.Slice(elements, func(i, j int) bool {
			li, lok := elements[i].(*IntegerValue)
			rj, rok := elements[j].(*IntegerValue)
			if !lok || !rok {
				return false
			}
			return li.Value < rj.Value
		})
	case *FloatValue:
		sort.Slice(elements, func(i, j int) bool {
			li, lok := elements[i].(*FloatValue)
			rj, rok := elements[j].(*FloatValue)
			if !lok || !rok {
				return false
			}
			return li.Value < rj.Value
		})
	case *StringValue:
		sort.Slice(elements, func(i, j int) bool {
			li, lok := elements[i].(*StringValue)
			rj, rok := elements[j].(*StringValue)
			if !lok || !rok {
				return false
			}
			return li.Value < rj.Value
		})
	case *BooleanValue:
		sort.Slice(elements, func(i, j int) bool {
			li, lok := elements[i].(*BooleanValue)
			rj, rok := elements[j].(*BooleanValue)
			if !lok || !rok {
				return false
			}
			return !li.Value && rj.Value
		})
	default:
		return &NilValue{}
	}

	return &NilValue{}
}

// ArrayHelperSlice extracts a slice from an array.
// Indices are adjusted relative to the array's low bound.
func ArrayHelperSlice(arr *ArrayValue, startIdx, endIdx int64) Value {
	lowBound := int64(0)
	if arr.ArrayType != nil && arr.ArrayType.LowBound != nil {
		lowBound = int64(*arr.ArrayType.LowBound)
	}

	start := int(startIdx - lowBound)
	end := int(endIdx - lowBound)

	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if end > len(arr.Elements) {
		end = len(arr.Elements)
	}
	if start > end {
		start = end
	}

	resultElements := make([]Value, end-start)
	copy(resultElements, arr.Elements[start:end])

	return &ArrayValue{Elements: resultElements, ArrayType: arr.ArrayType}
}
