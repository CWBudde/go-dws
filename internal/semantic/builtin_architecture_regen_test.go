package semantic

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

// encodeJSONString renders s as a JSON string without Go's default HTML
// escaping, so that "<nil>" stays readable in the golden file.
func encodeJSONString(t *testing.T, s string) string {
	t.Helper()
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(s); err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(buf.String(), "\n")
}

type regenCase struct {
	Arguments string   `json:"arguments"`
	Result    []string `json:"result"`
}

// TestRegenerateBuiltinCompatibility rewrites the recorded golden file in
// place, preserving its one-entry-per-line layout. Run it with
// REGEN_BUILTIN_COMPAT=1 after deliberately changing a builtin signature, then
// review the diff.
func TestRegenerateBuiltinCompatibility(t *testing.T) {
	if os.Getenv("REGEN_BUILTIN_COMPAT") != "1" {
		t.Skip("set REGEN_BUILTIN_COMPAT=1 to regenerate")
	}
	path := "testdata/builtin_analysis_compatibility.json"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string][]regenCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("{\n")
	for i, name := range names {
		patterns := make([]string, 0)
		for _, entry := range cases[name] {
			patterns = append(patterns, strings.Split(entry.Arguments, ",")...)
		}
		groups := map[string][]string{}
		for _, pattern := range patterns {
			parts := make([]string, 0)
			for _, item := range builtinCompatibilityResult(name, pattern) {
				parts = append(parts, encodeJSONString(t, item))
			}
			key := "[" + strings.Join(parts, ", ") + "]"
			groups[key] = append(groups[key], pattern)
		}
		lines := make([]string, 0, len(groups))
		for result, group := range groups {
			sort.Strings(group)
			args := encodeJSONString(t, strings.Join(group, ","))
			lines = append(lines, "    {\"arguments\": "+args+", \"result\": "+result+"}")
		}
		sort.Strings(lines)
		b.WriteString("  " + encodeJSONString(t, name) + ": [\n")
		b.WriteString(strings.Join(lines, ",\n"))
		b.WriteString("\n  ]")
		if i != len(names)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}
