package cmd

import (
	"fmt"
	"io"
	"strings"
)

// diffContextLines is the number of unchanged lines printed around each hunk,
// matching the default of `diff -u` and `gofmt -d`.
const diffContextLines = 3

// diffOpKind classifies a single line-level edit operation.
type diffOpKind int

const (
	diffEqual diffOpKind = iota
	diffDelete
	diffInsert
)

// diffOp is one line-level edit operation. Idx is the index of the line in the
// source text it comes from: the old text for diffEqual and diffDelete, the new
// text for diffInsert.
type diffOp struct {
	Kind diffOpKind
	Idx  int
}

// diffText holds a text split into lines plus whether it ended with a newline.
type diffText struct {
	lines        []string
	finalNewline bool
}

// splitDiffLines splits s into lines, remembering whether the text ended with a
// newline. An empty string yields no lines at all, so that comparing an empty
// file against a non-empty one reports a pure insertion.
func splitDiffLines(s string) diffText {
	if s == "" {
		return diffText{lines: nil, finalNewline: true}
	}
	if strings.HasSuffix(s, "\n") {
		return diffText{lines: strings.Split(strings.TrimSuffix(s, "\n"), "\n"), finalNewline: true}
	}
	return diffText{lines: strings.Split(s, "\n"), finalNewline: false}
}

// noEOLSentinel is appended to the comparison key of a final line that is not
// newline-terminated, so that an otherwise identical last line still shows up
// as a change when only its termination differs (what `diff -u` does).
const noEOLSentinel = "\x00\\ No newline at end of file"

// compareKeys returns the values the diff algorithm matches on.
func (t diffText) compareKeys() []string {
	if t.finalNewline || len(t.lines) == 0 {
		return t.lines
	}
	keys := make([]string, len(t.lines))
	copy(keys, t.lines)
	keys[len(keys)-1] += noEOLSentinel
	return keys
}

// maxDiffEditDistance bounds the edit distance the forward pass will explore.
//
// Backtracking needs one snapshot of the furthest-reaching endpoints per edit
// distance d, and only the diagonals k in [-d, d] are ever read back, so a
// snapshot holds 2d+1 ints and a complete trace costs 8*(D+1)^2 bytes on a
// 64-bit build. That is quadratic in the edit distance, which for two texts
// with no line in common approaches the sum of their lengths: two 5,000-line
// inputs would allocate gigabytes and can take the process down with it.
//
// Past this bound the exact edit script is abandoned in favour of a whole-file
// replacement (see replaceAllOps). Real formatter diffs are far below it - the
// distance is roughly twice the number of reformatted lines, so this tolerates
// about 1,400 changed lines - and its cost is capped at roughly 67 MiB.
const maxDiffEditDistance = 2896

// myersDiff computes a minimal line-level edit script transforming a into b
// using Myers' O(ND) difference algorithm. The returned operations are in
// source order. When the edit distance exceeds maxDiffEditDistance the exact
// script is given up and the whole of a is reported as replaced by b.
func myersDiff(a, b []string) []diffOp {
	trace, ok := myersTrace(a, b)
	if !ok {
		return replaceAllOps(len(a), len(b))
	}
	return myersBacktrack(trace, len(a), len(b))
}

// replaceAllOps is the bounded fallback for inputs whose edit distance exceeds
// maxDiffEditDistance: every old line is deleted and every new line inserted,
// which buildHunks renders as a single whole-file hunk.
func replaceAllOps(n, m int) []diffOp {
	ops := make([]diffOp, 0, n+m)
	for i := 0; i < n; i++ {
		ops = append(ops, diffOp{Kind: diffDelete, Idx: i})
	}
	for i := 0; i < m; i++ {
		ops = append(ops, diffOp{Kind: diffInsert, Idx: i})
	}
	return ops
}

// myersTrace runs the forward pass of Myers' algorithm, returning one snapshot
// of the furthest-reaching path endpoints per edit distance. Snapshot d covers
// the diagonals k in [-d, d] - the only ones myersBacktrack reads at that
// distance - stored at index d+k, so the trace grows quadratically in the edit
// distance rather than in the product of the input lengths. ok is false when
// the edit distance exceeds maxDiffEditDistance, in which case no trace is
// returned.
func myersTrace(a, b []string) (trace [][]int, ok bool) {
	n, m := len(a), len(b)
	offset := n + m
	maxD := min(n+m, maxDiffEditDistance)
	// Two slots of slack: nextX reads the neighbouring diagonals k±1, which
	// reach one past the widest diagonal the search itself visits.
	v := make([]int, 2*offset+3)
	trace = make([][]int, 0, maxD+1)

	for d := 0; d <= maxD; d++ {
		snapshot := make([]int, 2*d+1)
		copy(snapshot, v[offset-d:offset+d+1])
		trace = append(trace, snapshot)

		for k := -d; k <= d; k += 2 {
			x := nextX(v, offset, k, d)
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[offset+k] = x
			if x >= n && y >= m {
				return trace, true
			}
		}
	}
	return nil, false
}

// nextX picks the furthest-reaching endpoint on diagonal k for edit distance d:
// either by moving down from k+1 (an insertion) or right from k-1 (a deletion).
func nextX(v []int, offset, k, d int) int {
	if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
		return v[offset+k+1] // move down: insert b[y]
	}
	return v[offset+k-1] + 1 // move right: delete a[x]
}

// myersBacktrack walks the forward-pass trace backwards and reconstructs the
// edit script in source order. Snapshot d is indexed by d+k, covering the
// diagonals k in [-d, d]; at distance 0 the predecessor is the origin.
func myersBacktrack(trace [][]int, n, m int) []diffOp {
	var reversed []diffOp
	x, y := n, m
	for d := len(trace) - 1; d >= 0 && (x > 0 || y > 0); d-- {
		prevX, prevY := 0, 0
		if d > 0 {
			prev := trace[d]
			k := x - y

			prevK := k - 1
			if k == -d || (k != d && prev[d+k-1] < prev[d+k+1]) {
				prevK = k + 1
			}
			prevX = prev[d+prevK]
			prevY = prevX - prevK
		}

		for x > prevX && y > prevY {
			x--
			y--
			reversed = append(reversed, diffOp{Kind: diffEqual, Idx: x})
		}
		if d == 0 {
			break
		}
		if x > prevX {
			reversed = append(reversed, diffOp{Kind: diffDelete, Idx: prevX})
		} else {
			reversed = append(reversed, diffOp{Kind: diffInsert, Idx: prevY})
		}
		x, y = prevX, prevY
	}

	ops := make([]diffOp, len(reversed))
	for i, op := range reversed {
		ops[len(reversed)-1-i] = op
	}
	return ops
}

// diffHunk is a contiguous group of operations surrounded by context lines.
type diffHunk struct {
	ops      []diffOp
	oldStart int // 0-based index of the first old line in the hunk
	newStart int // 0-based index of the first new line in the hunk
	oldCount int
	newCount int
}

// buildHunks groups the edit script into unified-diff hunks with up to
// diffContextLines lines of context on each side.
func buildHunks(ops []diffOp) []diffHunk {
	var hunks []diffHunk

	i := 0
	oldLine, newLine := 0, 0

	for i < len(ops) {
		if ops[i].Kind == diffEqual {
			oldLine++
			newLine++
			i++
			continue
		}

		// Start of a change: back up over the leading context.
		start, lead := i, 0
		for start > 0 && ops[start-1].Kind == diffEqual && lead < diffContextLines {
			start--
			lead++
		}

		end := hunkEnd(ops, i)
		oldCount, newCount := countLines(ops[start:end])

		hunks = append(hunks, diffHunk{
			ops:      ops[start:end],
			oldStart: oldLine - lead,
			newStart: newLine - lead,
			oldCount: oldCount,
			newCount: newCount,
		})

		// Advance the line counters past every op this hunk consumed beyond
		// the leading context, which was counted already.
		consumedOld, consumedNew := countLines(ops[i:end])
		oldLine += consumedOld
		newLine += consumedNew
		i = end
	}

	return hunks
}

// hunkEnd returns the exclusive end index of the hunk that starts changing at
// index i: it absorbs runs of equal lines shorter than twice the context (so
// nearby changes share one hunk) and then up to diffContextLines trailing
// context lines.
func hunkEnd(ops []diffOp, i int) int {
	end := i
	for end < len(ops) {
		if ops[end].Kind != diffEqual {
			end++
			continue
		}
		run := 0
		for end+run < len(ops) && ops[end+run].Kind == diffEqual {
			run++
		}
		if end+run < len(ops) && run <= 2*diffContextLines {
			end += run
			continue
		}
		break
	}
	for trail := 0; end < len(ops) && ops[end].Kind == diffEqual && trail < diffContextLines; trail++ {
		end++
	}
	return end
}

// countLines reports how many old-side and new-side lines the operations cover.
func countLines(ops []diffOp) (oldCount, newCount int) {
	for _, op := range ops {
		switch op.Kind {
		case diffEqual:
			oldCount++
			newCount++
		case diffDelete:
			oldCount++
		case diffInsert:
			newCount++
		}
	}
	return oldCount, newCount
}

// formatRange renders one side of a hunk header the way GNU diff does: the
// count is omitted when exactly one line is covered, and a zero-length range
// points at the line before the change.
func formatRange(start, count int) string {
	if count == 0 {
		return fmt.Sprintf("%d,0", start)
	}
	if count == 1 {
		return fmt.Sprintf("%d", start+1)
	}
	return fmt.Sprintf("%d,%d", start+1, count)
}

// WriteUnifiedDiff writes a unified diff of original vs. formatted to w,
// including the `---`/`+++` file header. Nothing is written when the two texts
// are identical. It reports whether any difference was found.
func WriteUnifiedDiff(w io.Writer, filename, original, formatted string) (bool, error) {
	if original == formatted {
		return false, nil
	}

	oldText := splitDiffLines(original)
	newText := splitDiffLines(formatted)

	ops := myersDiff(oldText.compareKeys(), newText.compareKeys())
	hunks := buildHunks(ops)
	if len(hunks) == 0 {
		return true, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s (original)\n", filename)
	fmt.Fprintf(&sb, "+++ %s (formatted)\n", filename)

	for _, h := range hunks {
		fmt.Fprintf(&sb, "@@ -%s +%s @@\n", formatRange(h.oldStart, h.oldCount), formatRange(h.newStart, h.newCount))
		for _, op := range h.ops {
			switch op.Kind {
			case diffEqual:
				writeDiffLine(&sb, " ", oldText, op.Idx)
			case diffDelete:
				writeDiffLine(&sb, "-", oldText, op.Idx)
			case diffInsert:
				writeDiffLine(&sb, "+", newText, op.Idx)
			}
		}
	}

	_, err := io.WriteString(w, sb.String())
	return true, err
}

// writeDiffLine writes a single diff body line, appending the standard
// "\ No newline at end of file" marker when the text's last line is unterminated.
func writeDiffLine(sb *strings.Builder, prefix string, text diffText, idx int) {
	sb.WriteString(prefix)
	sb.WriteString(text.lines[idx])
	sb.WriteString("\n")
	if !text.finalNewline && idx == len(text.lines)-1 {
		sb.WriteString("\\ No newline at end of file\n")
	}
}
