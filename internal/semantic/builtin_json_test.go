package semantic

import (
	"testing"
)

// ============================================================================
// Built-in JSON Functions Tests
// ============================================================================
// These tests cover the built-in JSON manipulation functions to improve
// coverage of analyze_builtin_json.go (currently at 0% coverage)

// ParseJSON function tests
func TestBuiltinParseJSON_Object(t *testing.T) {
	input := `
		var json := ParseJSON('{"name": "John", "age": 30}');
	`
	expectNoErrors(t, input)
}

func TestBuiltinParseJSON_Array(t *testing.T) {
	input := `
		var json := ParseJSON('[1, 2, 3, 4, 5]');
	`
	expectNoErrors(t, input)
}

func TestBuiltinParseJSON_Nested(t *testing.T) {
	input := `
		var json := ParseJSON('{"user": {"name": "John", "roles": ["admin", "user"]}}');
	`
	expectNoErrors(t, input)
}

func TestBuiltinParseJSON_Empty(t *testing.T) {
	input := `
		var json := ParseJSON('{}');
	`
	expectNoErrors(t, input)
}

func TestBuiltinParseJSON_InvalidJSON(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var json := ParseJSON('not valid json');
	`
	expectNoErrors(t, input)
}

func TestBuiltinParseJSON_InvalidType(t *testing.T) {
	input := `
		var json := ParseJSON(42);
	`
	expectError(t, input, "string")
}

// ToJSON function tests
func TestBuiltinToJSON_Variant(t *testing.T) {
	input := `
		var v: Variant := 42;
		var json := ToJSON(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinToJSON_String(t *testing.T) {
	input := `
		var s := 'hello';
		var json := ToJSON(s);
	`
	expectNoErrors(t, input)
}

func TestBuiltinToJSON_Integer(t *testing.T) {
	input := `
		var n := 42;
		var json := ToJSON(n);
	`
	expectNoErrors(t, input)
}

// ToJSONFormatted function tests
func TestBuiltinToJSONFormatted_Basic(t *testing.T) {
	input := `
		var v: Variant;
		var json := ToJSONFormatted(v);
	`
	expectNoErrors(t, input)
}

func TestBuiltinToJSONFormatted_WithIndent(t *testing.T) {
	input := `
		var v: Variant;
		var json := ToJSONFormatted(v, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinToJSONFormatted_NoIndent(t *testing.T) {
	input := `
		var v: Variant;
		var json := ToJSONFormatted(v, 0);
	`
	expectNoErrors(t, input)
}

// JSONHasField function tests
func TestBuiltinJSONHasField_Exists(t *testing.T) {
	input := `
		var json := ParseJSON('{"name": "John"}');
		var hasName := JSONHasField(json, 'name');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONHasField_NotExists(t *testing.T) {
	input := `
		var json := ParseJSON('{"name": "John"}');
		var hasAge := JSONHasField(json, 'age');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONHasField_InvalidArgCount(t *testing.T) {
	input := `
		var json := ParseJSON('{}');
		var hasField := JSONHasField(json);
	`
	expectError(t, input, "argument")
}

// JSONKeys function tests
func TestBuiltinJSONKeys_Basic(t *testing.T) {
	input := `
		var json := ParseJSON('{"name": "John", "age": 30}');
		var keys := JSONKeys(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONKeys_Empty(t *testing.T) {
	input := `
		var json := ParseJSON('{}');
		var keys := JSONKeys(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONKeys_Array(t *testing.T) {
	// Should analyze without error (might return empty or error at runtime)
	input := `
		var json := ParseJSON('[1, 2, 3]');
		var keys := JSONKeys(json);
	`
	expectNoErrors(t, input)
}

// JSONValues function tests
func TestBuiltinJSONValues_Basic(t *testing.T) {
	input := `
		var json := ParseJSON('{"name": "John", "age": 30}');
		var values := JSONValues(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONValues_Empty(t *testing.T) {
	input := `
		var json := ParseJSON('{}');
		var values := JSONValues(json);
	`
	expectNoErrors(t, input)
}

// JSONLength function tests
func TestBuiltinJSONLength_Object(t *testing.T) {
	input := `
		var json := ParseJSON('{"a": 1, "b": 2, "c": 3}');
		var len := JSONLength(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONLength_Array(t *testing.T) {
	input := `
		var json := ParseJSON('[1, 2, 3, 4, 5]');
		var len := JSONLength(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSONLength_Empty(t *testing.T) {
	input := `
		var json := ParseJSON('{}');
		var len := JSONLength(json);
	`
	expectNoErrors(t, input)
}

// Combined JSON operations tests
func TestBuiltinJSON_ParseAndQuery(t *testing.T) {
	input := `
		var jsonStr := '{"name": "John", "age": 30, "email": "john@example.com"}';
		var json := ParseJSON(jsonStr);
		var hasName := JSONHasField(json, 'name');
		var keys := JSONKeys(json);
		var len := JSONLength(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_CreateAndStringify(t *testing.T) {
	input := `
		var v: Variant;
		var json := ToJSON(v);
		var formatted := ToJSONFormatted(v, 2);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_RoundTrip(t *testing.T) {
	input := `
		var original := '{"test": 123}';
		var parsed := ParseJSON(original);
		var serialized := ToJSON(parsed);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_ArrayOperations(t *testing.T) {
	input := `
		var jsonStr := '[1, 2, 3, 4, 5]';
		var json := ParseJSON(jsonStr);
		var len := JSONLength(json);
		var keys := JSONKeys(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_NestedAccess(t *testing.T) {
	input := `
		var jsonStr := '{"user": {"name": "John", "address": {"city": "NYC"}}}';
		var json := ParseJSON(jsonStr);
		var hasUser := JSONHasField(json, 'user');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_InFunction(t *testing.T) {
	input := `
		function ParseUser(jsonStr: String): Variant;
		begin
			Result := ParseJSON(jsonStr);
		end;

		var user := ParseUser('{"name": "John"}');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_InExpression(t *testing.T) {
	input := `
		var isValid := JSONHasField(ParseJSON('{"name": "John"}'), 'name');
	`
	expectNoErrors(t, input)
}

// Edge cases
func TestBuiltinJSON_EmptyString(t *testing.T) {
	// Should analyze without error (runtime check)
	input := `
		var json := ParseJSON('');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_LargeObject(t *testing.T) {
	input := `
		var jsonStr := '{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8}';
		var json := ParseJSON(jsonStr);
		var len := JSONLength(json);
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_SpecialCharacters(t *testing.T) {
	input := `
		var json := ParseJSON('{"text": "Hello \"World\""}');
	`
	expectNoErrors(t, input)
}

func TestBuiltinJSON_UnicodeCharacters(t *testing.T) {
	input := `
		var json := ParseJSON('{"emoji": "😀", "chinese": "中文"}');
	`
	expectNoErrors(t, input)
}
