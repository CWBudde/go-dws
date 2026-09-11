package cmd

import (
	"strings"
	"testing"
)

// TestWriteUnifiedDiff checks the unified-diff renderer against output captured
// from GNU `diff -u --label "x (original)" --label "x (formatted)"` on the same
// inputs.
func TestWriteUnifiedDiff(t *testing.T) {
	const header = "--- x (original)\n+++ x (formatted)\n"

	tests := []struct {
		name       string
		original   string
		formatted  string
		want       string
		wantDiffer bool
	}{
		{
			name:       "identical",
			original:   "a\nb\nc\n",
			formatted:  "a\nb\nc\n",
			want:       "",
			wantDiffer: false,
		},
		{
			name:      "pure insertion",
			original:  "a\nb\nc\n",
			formatted: "a\nb\nX\nc\n",
			want: header + "@@ -1,3 +1,4 @@\n" +
				" a\n b\n+X\n c\n",
			wantDiffer: true,
		},
		{
			name:      "pure deletion",
			original:  "a\nb\nc\nd\n",
			formatted: "a\nb\nd\n",
			want: header + "@@ -1,4 +1,3 @@\n" +
				" a\n b\n-c\n d\n",
			wantDiffer: true,
		},
		{
			name:      "replacement",
			original:  "a\nb\nc\n",
			formatted: "a\nX\nc\n",
			want: header + "@@ -1,3 +1,3 @@\n" +
				" a\n-b\n+X\n c\n",
			wantDiffer: true,
		},
		{
			name:      "change at the very start",
			original:  "a\nb\nc\nd\ne\n",
			formatted: "X\nb\nc\nd\ne\n",
			want: header + "@@ -1,4 +1,4 @@\n" +
				"-a\n+X\n b\n c\n d\n",
			wantDiffer: true,
		},
		{
			name:      "change at the very end",
			original:  "a\nb\nc\nd\ne\n",
			formatted: "a\nb\nc\nd\nX\n",
			want: header + "@@ -2,4 +2,4 @@\n" +
				" b\n c\n d\n-e\n+X\n",
			wantDiffer: true,
		},
		{
			name:      "multiple separated hunks",
			original:  "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20\n",
			formatted: "1\nX\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\nY\n20\n",
			want: header +
				"@@ -1,5 +1,5 @@\n 1\n-2\n+X\n 3\n 4\n 5\n" +
				"@@ -16,5 +16,5 @@\n 16\n 17\n 18\n-19\n+Y\n 20\n",
			wantDiffer: true,
		},
		{
			name:      "missing trailing newline in the formatted text",
			original:  "a\nb\nc\n",
			formatted: "a\nb\nc",
			want: header + "@@ -1,3 +1,3 @@\n" +
				" a\n b\n-c\n+c\n\\ No newline at end of file\n",
			wantDiffer: true,
		},
		{
			name:      "missing trailing newline in the original text",
			original:  "a\nb\nc",
			formatted: "a\nb\nc\n",
			want: header + "@@ -1,3 +1,3 @@\n" +
				" a\n b\n-c\n\\ No newline at end of file\n+c\n",
			wantDiffer: true,
		},
		{
			name:      "blank lines are content",
			original:  "a\n\n\nb\n",
			formatted: "a\n\nb\n",
			want: header + "@@ -1,4 +1,3 @@\n" +
				" a\n \n-\n b\n",
			wantDiffer: true,
		},
		{
			name:      "empty original",
			original:  "",
			formatted: "a\nb\n",
			want: header + "@@ -0,0 +1,2 @@\n" +
				"+a\n+b\n",
			wantDiffer: true,
		},
		{
			name:      "everything deleted",
			original:  "a\nb\n",
			formatted: "",
			want: header + "@@ -1,2 +0,0 @@\n" +
				"-a\n-b\n",
			wantDiffer: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			differs, err := WriteUnifiedDiff(&sb, "x", tt.original, tt.formatted)
			if err != nil {
				t.Fatalf("WriteUnifiedDiff() error = %v", err)
			}
			if differs != tt.wantDiffer {
				t.Errorf("WriteUnifiedDiff() differs = %v, want %v", differs, tt.wantDiffer)
			}
			if got := sb.String(); got != tt.want {
				t.Errorf("WriteUnifiedDiff() output mismatch\n--- got ---\n%s--- want ---\n%s", got, tt.want)
			}
		})
	}
}

// TestMyersDiffDoesNotCascade verifies that a single inserted line does not make
// every following line report as changed, which was the defect of the previous
// positional comparison.
func TestMyersDiffDoesNotCascade(t *testing.T) {
	original := "a\nb\nc\nd\ne\nf\ng\n"
	formatted := "a\nINSERTED\nb\nc\nd\ne\nf\ng\n"

	var sb strings.Builder
	if _, err := WriteUnifiedDiff(&sb, "x", original, formatted); err != nil {
		t.Fatalf("WriteUnifiedDiff() error = %v", err)
	}

	changed := 0
	for _, line := range strings.Split(sb.String(), "\n") {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			changed++
		}
	}
	if changed != 1 {
		t.Errorf("expected exactly one changed line, got %d:\n%s", changed, sb.String())
	}
}
