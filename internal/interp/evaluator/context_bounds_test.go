package evaluator

import (
	"testing"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// TestGetLowBound_StaticArray tests GetLowBound with a static array.
func TestGetLowBound_StaticArray(t *testing.T) {
	e := createTestEvaluator()

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewStaticArrayType(types.INTEGER, 5, 10),
		Elements:  make([]runtime.Value, 6),
	}

	result, err := e.GetLowBound(arrayVal)
	if err != nil {
		t.Errorf("GetLowBound with static array should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetLowBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("GetLowBound expected 5, got %d", intResult.Value)
	}
}

// TestGetHighBound_StaticArray tests GetHighBound with a static array.
func TestGetHighBound_StaticArray(t *testing.T) {
	e := createTestEvaluator()

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewStaticArrayType(types.INTEGER, 5, 10),
		Elements:  make([]runtime.Value, 6),
	}

	result, err := e.GetHighBound(arrayVal)
	if err != nil {
		t.Errorf("GetHighBound with static array should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetHighBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 10 {
		t.Errorf("GetHighBound expected 10, got %d", intResult.Value)
	}
}

// TestGetLowBound_DynamicArray tests GetLowBound with a dynamic array.
func TestGetLowBound_DynamicArray(t *testing.T) {
	e := createTestEvaluator()

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.STRING),
		Elements:  make([]runtime.Value, 5),
	}

	result, err := e.GetLowBound(arrayVal)
	if err != nil {
		t.Errorf("GetLowBound with dynamic array should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetLowBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 0 {
		t.Errorf("GetLowBound for dynamic array expected 0, got %d", intResult.Value)
	}
}

// TestGetHighBound_DynamicArray tests GetHighBound with a dynamic array.
func TestGetHighBound_DynamicArray(t *testing.T) {
	e := createTestEvaluator()

	arrayVal := &runtime.ArrayValue{
		ArrayType: types.NewDynamicArrayType(types.STRING),
		Elements:  make([]runtime.Value, 5),
	}

	result, err := e.GetHighBound(arrayVal)
	if err != nil {
		t.Errorf("GetHighBound with dynamic array should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetHighBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 4 {
		t.Errorf("GetHighBound for dynamic array expected 4, got %d", intResult.Value)
	}
}

// TestGetLowBound_String tests GetLowBound with a string.
func TestGetLowBound_String(t *testing.T) {
	e := createTestEvaluator()

	strVal := &runtime.StringValue{Value: "hello"}

	result, err := e.GetLowBound(strVal)
	if err != nil {
		t.Errorf("GetLowBound with string should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetLowBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 1 {
		t.Errorf("GetLowBound for string expected 1, got %d", intResult.Value)
	}
}

// TestGetHighBound_String tests GetHighBound with a string.
func TestGetHighBound_String(t *testing.T) {
	e := createTestEvaluator()

	strVal := &runtime.StringValue{Value: "hello"}

	result, err := e.GetHighBound(strVal)
	if err != nil {
		t.Errorf("GetHighBound with string should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetHighBound should return IntegerValue, got %T", result)
	}
	if intResult.Value != 5 {
		t.Errorf("GetHighBound for string expected 5, got %d", intResult.Value)
	}
}

// TestGetLowBound_TypeMetaInteger tests GetLowBound with Integer type.
func TestGetLowBound_TypeMetaInteger(t *testing.T) {
	e := createTestEvaluator()

	typeVal := &runtime.TypeMetaValue{TypeInfo: types.INTEGER}

	result, err := e.GetLowBound(typeVal)
	if err != nil {
		t.Errorf("GetLowBound with Integer type should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetLowBound should return IntegerValue, got %T", result)
	}
	// Integer low bound should be MinInt64
	if intResult.Value >= 0 {
		t.Errorf("GetLowBound for Integer should be negative MinInt64")
	}
}

// TestGetHighBound_TypeMetaInteger tests GetHighBound with Integer type.
func TestGetHighBound_TypeMetaInteger(t *testing.T) {
	e := createTestEvaluator()

	typeVal := &runtime.TypeMetaValue{TypeInfo: types.INTEGER}

	result, err := e.GetHighBound(typeVal)
	if err != nil {
		t.Errorf("GetHighBound with Integer type should not return error: %v", err)
	}
	intResult, ok := result.(*runtime.IntegerValue)
	if !ok {
		t.Errorf("GetHighBound should return IntegerValue, got %T", result)
	}
	// Integer high bound should be MaxInt64
	if intResult.Value <= 0 {
		t.Errorf("GetHighBound for Integer should be positive MaxInt64")
	}
}

// TestGetLowBound_TypeMetaBoolean tests GetLowBound with Boolean type.
func TestGetLowBound_TypeMetaBoolean(t *testing.T) {
	e := createTestEvaluator()

	typeVal := &runtime.TypeMetaValue{TypeInfo: types.BOOLEAN}

	result, err := e.GetLowBound(typeVal)
	if err != nil {
		t.Errorf("GetLowBound with Boolean type should not return error: %v", err)
	}
	boolResult, ok := result.(*runtime.BooleanValue)
	if !ok {
		t.Errorf("GetLowBound should return BooleanValue, got %T", result)
	}
	if boolResult.Value != false {
		t.Errorf("GetLowBound for Boolean expected false, got %v", boolResult.Value)
	}
}

// TestGetHighBound_TypeMetaBoolean tests GetHighBound with Boolean type.
func TestGetHighBound_TypeMetaBoolean(t *testing.T) {
	e := createTestEvaluator()

	typeVal := &runtime.TypeMetaValue{TypeInfo: types.BOOLEAN}

	result, err := e.GetHighBound(typeVal)
	if err != nil {
		t.Errorf("GetHighBound with Boolean type should not return error: %v", err)
	}
	boolResult, ok := result.(*runtime.BooleanValue)
	if !ok {
		t.Errorf("GetHighBound should return BooleanValue, got %T", result)
	}
	if boolResult.Value != true {
		t.Errorf("GetHighBound for Boolean expected true, got %v", boolResult.Value)
	}
}

// TestGetLowBound_UnsupportedType tests GetLowBound with unsupported type.
func TestGetLowBound_UnsupportedType(t *testing.T) {
	e := createTestEvaluator()

	floatVal := &runtime.FloatValue{Value: 3.14}

	_, err := e.GetLowBound(floatVal)
	if err == nil {
		t.Error("GetLowBound with unsupported type should return error")
	}
}

// TestGetHighBound_UnsupportedType tests GetHighBound with unsupported type.
func TestGetHighBound_UnsupportedType(t *testing.T) {
	e := createTestEvaluator()

	floatVal := &runtime.FloatValue{Value: 3.14}

	_, err := e.GetHighBound(floatVal)
	if err == nil {
		t.Error("GetHighBound with unsupported type should return error")
	}
}
