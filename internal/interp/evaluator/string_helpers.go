package evaluator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// evalStringHelper evaluates built-in string helper methods and properties directly.
// Returns nil only for unknown specs.
func (e *Evaluator) evalStringHelper(spec string, selfValue Value, args []Value, node ast.Node) Value {
	switch spec {
	case "__string_toupper":
		return e.evalStringToUpper(selfValue, args, node)
	case "__string_tolower":
		return e.evalStringToLower(selfValue, args, node)
	case "__string_length":
		return e.evalStringLength(selfValue, node)
	case "__string_tostring":
		return e.evalStringToString(selfValue, args, node)
	case "__string_tointeger":
		return e.evalStringToInteger(selfValue, args, node)
	case "__string_tofloat":
		return e.evalStringToFloat(selfValue, args, node)
	case "__string_startswith":
		return e.evalStringStartsWith(selfValue, args, node)
	case "__string_endswith":
		return e.evalStringEndsWith(selfValue, args, node)
	case "__string_contains":
		return e.evalStringContains(selfValue, args, node)
	case "__string_indexof":
		return e.evalStringIndexOf(selfValue, args, node)
	case "__string_matches":
		return e.evalStringMatches(selfValue, args, node)
	case "__string_isascii":
		return e.evalStringIsASCII(selfValue, args, node)
	case "__string_copy":
		return e.evalStringCopy(selfValue, args, node)
	case "__string_before":
		return e.evalStringBefore(selfValue, args, node)
	case "__string_after":
		return e.evalStringAfter(selfValue, args, node)
	case "__string_trim":
		return e.evalStringTrim(selfValue, args, node)
	case "__string_trimleft":
		return e.evalStringTrimLeft(selfValue, args, node)
	case "__string_trimright":
		return e.evalStringTrimRight(selfValue, args, node)
	case "__string_split":
		return e.evalStringSplit(selfValue, args, node)
	case "__string_tojson":
		return e.evalStringToJSON(selfValue, args, node)
	case "__string_tohtml":
		return e.evalStringToHTML(selfValue, args, node)
	case "__string_tohtmlattribute":
		return e.evalStringToHTMLAttribute(selfValue, args, node)
	case "__string_tocsstext":
		return e.evalStringToCSSText(selfValue, args, node)
	case "__string_toxml":
		return e.evalStringToXML(selfValue, args, node)
	default:
		return nil
	}
}

func (e *Evaluator) evalStringToUpper(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToUpper", 0)
	if errVal != nil {
		return errVal
	}
	return &runtime.StringValue{Value: strings.ToUpper(strVal.Value)}
}

func (e *Evaluator) evalStringToLower(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToLower", 0)
	if errVal != nil {
		return errVal
	}
	return &runtime.StringValue{Value: strings.ToLower(strVal.Value)}
}

func (e *Evaluator) evalStringLength(selfValue Value, node ast.Node) Value {
	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.Length property requires string receiver")
	}
	return &runtime.IntegerValue{Value: int64(utf8.RuneCountInString(strVal.Value))}
}

func (e *Evaluator) evalStringToString(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToString", 0)
	if errVal != nil {
		return errVal
	}
	return strVal
}

func (e *Evaluator) evalStringToInteger(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToInteger", 0)
	if errVal != nil {
		return errVal
	}

	intValue, err := strconv.ParseInt(strings.TrimSpace(strVal.Value), 10, 64)
	if err != nil {
		return e.newError(node, "%q is not a valid integer value", strVal.Value)
	}
	return &runtime.IntegerValue{Value: intValue}
}

func (e *Evaluator) evalStringToFloat(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToFloat", 0)
	if errVal != nil {
		return errVal
	}

	floatValue, err := strconv.ParseFloat(strings.TrimSpace(strVal.Value), 64)
	if err != nil {
		return e.newError(node, "%q is not a valid floating-point value", strVal.Value)
	}
	return &runtime.FloatValue{Value: floatValue}
}

func (e *Evaluator) evalStringStartsWith(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.StartsWith")
	if errVal != nil {
		return errVal
	}
	if argVal.Value == "" {
		return &runtime.BooleanValue{Value: false}
	}
	return &runtime.BooleanValue{Value: strings.HasPrefix(strVal.Value, argVal.Value)}
}

func (e *Evaluator) evalStringEndsWith(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.EndsWith")
	if errVal != nil {
		return errVal
	}
	if argVal.Value == "" {
		return &runtime.BooleanValue{Value: false}
	}
	return &runtime.BooleanValue{Value: strings.HasSuffix(strVal.Value, argVal.Value)}
}

func (e *Evaluator) evalStringContains(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.Contains")
	if errVal != nil {
		return errVal
	}
	return &runtime.BooleanValue{Value: strings.Contains(strVal.Value, argVal.Value)}
}

func (e *Evaluator) evalStringIndexOf(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.IndexOf", -1)
	if errVal != nil {
		return errVal
	}
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "String.IndexOf expects 1 or 2 arguments, got %d", len(args))
	}

	needleVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return e.newError(node, "String.IndexOf expects String as first argument, got %s", args[0].Type())
	}

	offset := int64(1)
	if len(args) == 2 {
		offsetVal, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "String.IndexOf startIndex must be Integer, got %s", args[1].Type())
		}
		offset = offsetVal.Value
	}

	return &runtime.IntegerValue{Value: evalPosEx(needleVal.Value, strVal.Value, offset)}
}

func (e *Evaluator) evalStringMatches(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.Matches")
	if errVal != nil {
		return errVal
	}
	return &runtime.BooleanValue{Value: wildcardMatch(strVal.Value, argVal.Value)}
}

func (e *Evaluator) evalStringIsASCII(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.IsASCII", 0)
	if errVal != nil {
		return errVal
	}
	for _, r := range strVal.Value {
		if r > 127 {
			return &runtime.BooleanValue{Value: false}
		}
	}
	return &runtime.BooleanValue{Value: true}
}

func (e *Evaluator) evalStringCopy(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.Copy", -1)
	if errVal != nil {
		return errVal
	}
	if len(args) < 1 || len(args) > 2 {
		return e.newError(node, "String.Copy expects 1 or 2 arguments, got %d", len(args))
	}

	startVal, ok := args[0].(*runtime.IntegerValue)
	if !ok {
		return e.newError(node, "String.Copy start must be Integer, got %s", args[0].Type())
	}

	length := int64(2147483647)
	if len(args) == 2 {
		lengthVal, ok := args[1].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "String.Copy length must be Integer, got %s", args[1].Type())
		}
		length = lengthVal.Value
	}

	runes := []rune(strVal.Value)
	start := int(startVal.Value) - 1
	if start < 0 {
		start = 0
	}
	if start >= len(runes) || length <= 0 {
		return &runtime.StringValue{Value: ""}
	}

	end := start + int(length)
	if end > len(runes) {
		end = len(runes)
	}

	return &runtime.StringValue{Value: string(runes[start:end])}
}

func (e *Evaluator) evalStringBefore(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.Before")
	if errVal != nil {
		return errVal
	}
	idx := strings.Index(strVal.Value, argVal.Value)
	if idx < 0 {
		return &runtime.StringValue{Value: strVal.Value}
	}
	return &runtime.StringValue{Value: strVal.Value[:idx]}
}

func (e *Evaluator) evalStringAfter(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.After")
	if errVal != nil {
		return errVal
	}
	idx := strings.Index(strVal.Value, argVal.Value)
	if idx < 0 {
		return &runtime.StringValue{Value: ""}
	}
	return &runtime.StringValue{Value: strVal.Value[idx+len(argVal.Value):]}
}

func (e *Evaluator) evalStringTrim(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.Trim", -1)
	if errVal != nil {
		return errVal
	}

	switch len(args) {
	case 0:
		return &runtime.StringValue{Value: strings.Trim(strVal.Value, " \t\n\r")}
	case 2:
		leftVal, lok := args[0].(*runtime.IntegerValue)
		rightVal, rok := args[1].(*runtime.IntegerValue)
		if !lok || !rok {
			return e.newError(node, "String.Trim expects integer counts")
		}
		left := clampNonNegative(int(leftVal.Value))
		right := clampNonNegative(int(rightVal.Value))
		runes := []rune(strVal.Value)
		if left+right >= len(runes) {
			return &runtime.StringValue{Value: ""}
		}
		return &runtime.StringValue{Value: string(runes[left : len(runes)-right])}
	default:
		return e.newError(node, "String.Trim expects 0 or 2 arguments")
	}
}

func (e *Evaluator) evalStringTrimLeft(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.TrimLeft", -1)
	if errVal != nil {
		return errVal
	}

	switch len(args) {
	case 0:
		return &runtime.StringValue{Value: strings.TrimLeft(strVal.Value, " \t\n\r")}
	case 1:
		countVal, ok := args[0].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "String.TrimLeft expects Integer argument, got %s", args[0].Type())
		}
		return &runtime.StringValue{Value: trimLeftCount(strVal.Value, int(countVal.Value))}
	default:
		return e.newError(node, "String.TrimLeft expects 0 or 1 argument")
	}
}

func (e *Evaluator) evalStringTrimRight(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.TrimRight", -1)
	if errVal != nil {
		return errVal
	}

	switch len(args) {
	case 0:
		return &runtime.StringValue{Value: strings.TrimRight(strVal.Value, " \t\n\r")}
	case 1:
		countVal, ok := args[0].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "String.TrimRight expects Integer argument, got %s", args[0].Type())
		}
		return &runtime.StringValue{Value: trimRightCount(strVal.Value, int(countVal.Value))}
	default:
		return e.newError(node, "String.TrimRight expects 0 or 1 argument")
	}
}

func (e *Evaluator) evalStringSplit(selfValue Value, args []Value, node ast.Node) Value {
	strVal, argVal, errVal := e.requireStringPairHelper(selfValue, args, node, "String.Split")
	if errVal != nil {
		return errVal
	}

	var parts []string
	if argVal.Value == "" {
		if strVal.Value == "" {
			parts = []string{}
		} else {
			for _, r := range []rune(strVal.Value) {
				parts = append(parts, string(r))
			}
		}
	} else {
		parts = strings.Split(strVal.Value, argVal.Value)
	}

	elements := make([]Value, len(parts))
	for idx, part := range parts {
		elements[idx] = &runtime.StringValue{Value: part}
	}

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.STRING),
	}
}

func (e *Evaluator) evalStringToJSON(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToJSON", 0)
	if errVal != nil {
		return errVal
	}
	encoded, err := json.Marshal(strVal.Value)
	if err != nil {
		return e.newError(node, "String.ToJSON failed: %v", err)
	}
	return &runtime.StringValue{Value: string(encoded)}
}

func (e *Evaluator) evalStringToHTML(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToHTML", 0)
	if errVal != nil {
		return errVal
	}
	return &runtime.StringValue{Value: htmlEncode(strVal.Value)}
}

func (e *Evaluator) evalStringToHTMLAttribute(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToHtmlAttribute", 0)
	if errVal != nil {
		return errVal
	}
	return &runtime.StringValue{Value: htmlAttributeEncode(strVal.Value)}
}

func (e *Evaluator) evalStringToCSSText(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, args, node, "String.ToCSSText", 0)
	if errVal != nil {
		return errVal
	}
	return &runtime.StringValue{Value: cssEncode(strVal.Value)}
}

func (e *Evaluator) evalStringToXML(selfValue Value, args []Value, node ast.Node) Value {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, "String.ToXML", -1)
	if errVal != nil {
		return errVal
	}
	if len(args) > 1 {
		return e.newError(node, "String.ToXML expects 0 or 1 argument")
	}

	mode := 0
	if len(args) == 1 {
		modeVal, ok := args[0].(*runtime.IntegerValue)
		if !ok {
			return e.newError(node, "String.ToXML expects Integer as second argument (mode), got %s", args[0].Type())
		}
		mode = int(modeVal.Value)
	}

	encoded, err := xmlEncode(strVal.Value, mode)
	if err != nil {
		return e.newError(node, "%v", err)
	}
	return &runtime.StringValue{Value: encoded}
}

func (e *Evaluator) requireStringHelperReceiver(selfValue Value, args []Value, node ast.Node, name string, expectedArgs int) (*runtime.StringValue, Value) {
	if expectedArgs >= 0 && len(args) != expectedArgs {
		if expectedArgs == 0 {
			return nil, e.newError(node, "%s does not take arguments", name)
		}
		return nil, e.newError(node, "%s expects exactly %d arguments", name, expectedArgs)
	}

	strVal, ok := selfValue.(*runtime.StringValue)
	if !ok {
		return nil, e.newError(node, "%s requires string receiver", name)
	}
	return strVal, nil
}

func (e *Evaluator) requireStringPairHelper(selfValue Value, args []Value, node ast.Node, name string) (*runtime.StringValue, *runtime.StringValue, Value) {
	strVal, errVal := e.requireStringHelperReceiver(selfValue, nil, node, name, -1)
	if errVal != nil {
		return nil, nil, errVal
	}
	if len(args) != 1 {
		return nil, nil, e.newError(node, "%s expects exactly 1 argument", name)
	}
	argVal, ok := args[0].(*runtime.StringValue)
	if !ok {
		return nil, nil, e.newError(node, "%s expects String argument, got %s", name, args[0].Type())
	}
	return strVal, argVal, nil
}

func evalPosEx(needle, haystack string, offset int64) int64 {
	if offset < 1 || needle == "" {
		return 0
	}

	haystackRunes := []rune(haystack)
	needleRunes := []rune(needle)
	startIdx := int(offset) - 1
	if startIdx >= len(haystackRunes) {
		return 0
	}

	for idx := startIdx; idx <= len(haystackRunes)-len(needleRunes); idx++ {
		match := true
		for needleIdx := 0; needleIdx < len(needleRunes); needleIdx++ {
			if haystackRunes[idx+needleIdx] != needleRunes[needleIdx] {
				match = false
				break
			}
		}
		if match {
			return int64(idx + 1)
		}
	}

	return 0
}

func trimLeftCount(s string, count int) string {
	runes := []rune(s)
	count = clampNonNegative(count)
	if count >= len(runes) {
		return ""
	}
	return string(runes[count:])
}

func trimRightCount(s string, count int) string {
	runes := []rune(s)
	count = clampNonNegative(count)
	if count >= len(runes) {
		return ""
	}
	return string(runes[:len(runes)-count])
}

func htmlEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func htmlAttributeEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r > 255 {
			b.WriteRune(r)
			continue
		}

		code := int(r)
		if code >= 10 && code <= 99 {
			fmt.Fprintf(&b, "&#%d;", code)
		} else {
			fmt.Fprintf(&b, "&#x%X;", code)
		}
	}

	return b.String()
}

func cssEncode(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r > 255 {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('\\')
		b.WriteRune(r)
	}
	return b.String()
}

func xmlEncode(s string, mode int) (string, error) {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		if (r >= 1 && r <= 8) || (r >= 11 && r <= 12) || (r >= 14 && r <= 31) {
			switch mode {
			case 0:
				continue
			case 1:
				fmt.Fprintf(&b, "&#%d;", r)
				continue
			default:
				return "", fmt.Errorf("Unsupported character #%d", r)
			}
		}

		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}

	return b.String(), nil
}

func wildcardMatch(str, pattern string) bool {
	return wildcardMatchRunes([]rune(str), []rune(pattern), 0, 0)
}

func wildcardMatchRunes(str, pattern []rune, si, pi int) bool {
	if si == len(str) && pi == len(pattern) {
		return true
	}
	if pi == len(pattern) {
		return false
	}

	if pattern[pi] == '*' {
		for pi < len(pattern) && pattern[pi] == '*' {
			pi++
		}
		if pi == len(pattern) {
			return true
		}
		for si <= len(str) {
			if wildcardMatchRunes(str, pattern, si, pi) {
				return true
			}
			si++
		}
		return false
	}

	if si == len(str) {
		return false
	}
	if pattern[pi] == '?' || pattern[pi] == str[si] {
		return wildcardMatchRunes(str, pattern, si+1, pi+1)
	}
	return false
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
