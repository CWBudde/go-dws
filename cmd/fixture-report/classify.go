package main

// Classifying the failures is a separate question from counting them.
// baselines.json holds per-category pass-count floors, so it cannot tell a family
// shrinking from a fixture swapping one wrong line for another, and the pass/fail
// table cannot tell a fixture that is one line from passing from one that produces
// nothing at all. This diffs every failing fixture against its expectation, buckets
// it by distance and by what kind of line differs, and reduces the differing
// diagnostics to message *shapes* so the recurring ones can be counted (PLAN.md T8).

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// lcsCellLimit caps the diff's DP table. Fixtures are small — the largest
// expectation in the tree is a few hundred lines — so this only guards against a
// runaway output (an infinite loop printing until the timeout) costing gigabytes.
const lcsCellLimit = 4_000_000

// Failure kinds, in the order they are reported.
const (
	kindDiagnostics = "diagnostics" // only diagnostic lines differ
	kindOutput      = "output"      // only program output differs
	kindMixed       = "mixed"       // both
	kindEmpty       = "empty"       // the fixture produced nothing at all
	kindCrash       = "crash"       // the compiler or runtime panicked
	kindTimeout     = "timeout"     // the fixture did not terminate
)

var kindOrder = []string{kindDiagnostics, kindOutput, kindMixed, kindEmpty, kindCrash, kindTimeout}

var (
	// positionRe matches DWScript's position suffix. Everything from it rightward is
	// dropped when shaping a line: the anchor is a separate question from the sentence,
	// and mixing them would split one shape into dozens.
	positionRe = regexp.MustCompile(`\[line: \d+, column: \d+\]`)
	// severityRe recognises a diagnostic that carries no position (a runtime error
	// inside the envelope, for instance).
	severityRe = regexp.MustCompile(`^(Syntax Error|Runtime Error|Compile Error|Hint|Warning|Error|Note):`)
	doubleQtRe = regexp.MustCompile(`"[^"]*"`)
	singleQtRe = regexp.MustCompile(`'[^']*'`)
	digitsRe   = regexp.MustCompile(`\d+`)
	spacesRe   = regexp.MustCompile(`\s+`)
)

// isDiagnostic reports whether a line is a compiler or runtime message rather than
// program output. A position suffix is proof; so is a severity prefix, which is how
// the unpositioned runtime errors inside the "Errors >>>>" envelope are caught.
func isDiagnostic(line string) bool {
	return positionRe.MatchString(line) || severityRe.MatchString(line)
}

// shapeOf reduces a diagnostic to its message shape: position dropped, quoted names
// and numbers replaced by placeholders. `Incompatible types: "Integer" and "String"
// [line: 3, column: 12]` and the same sentence about two other types are one shape,
// which is what makes them countable.
func shapeOf(line string) string {
	if loc := positionRe.FindStringIndex(line); loc != nil {
		line = line[:loc[0]]
	}
	line = doubleQtRe.ReplaceAllString(line, `"X"`)
	line = singleQtRe.ReplaceAllString(line, `'X'`)
	line = digitsRe.ReplaceAllString(line, "N")
	return strings.TrimRight(spacesRe.ReplaceAllString(strings.TrimSpace(line), " "), " ,")
}

// lineDiff returns the expected lines absent from got, the got lines that were not
// expected, and the number of lines that would have to change for the fixture to
// pass. All three come from one longest-common-subsequence alignment, so a line that
// merely moved is reported once on each side rather than as a wholesale rewrite.
//
// distance is line-level edit distance, counted per hunk: a run of lines replaced by
// the same number of other lines is that many edits, not twice that many. This is the
// number the near-miss queue is ordered by, so it has to mean "how many output lines
// are wrong", and a fixture that emits the right diagnostic in the wrong words has to
// come out as 1 rather than 2.
func lineDiff(expected, got []string) (missing, spurious []string, distance int) {
	n, m := len(expected), len(got)
	if n*m > lcsCellLimit {
		missing, spurious = multisetDiff(expected, got)
		distance = len(missing)
		if len(spurious) > distance {
			distance = len(spurious)
		}
		return missing, spurious, distance
	}

	// dp[i][j] = length of the LCS of expected[i:] and got[j:], flattened.
	dp := make([]int, (n+1)*(m+1))
	at := func(i, j int) int { return i*(m+1) + j }
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case expected[i] == got[j]:
				dp[at(i, j)] = dp[at(i+1, j+1)] + 1
			case dp[at(i+1, j)] >= dp[at(i, j+1)]:
				dp[at(i, j)] = dp[at(i+1, j)]
			default:
				dp[at(i, j)] = dp[at(i, j+1)]
			}
		}
	}

	// Walk the alignment, accumulating the current hunk. A hunk ends at the first
	// line both sides agree on.
	var hunkMissing, hunkSpurious int
	closeHunk := func() {
		if hunkMissing > hunkSpurious {
			distance += hunkMissing
		} else {
			distance += hunkSpurious
		}
		hunkMissing, hunkSpurious = 0, 0
	}

	i, j := 0, 0
	for i < n && j < m {
		switch {
		case expected[i] == got[j]:
			closeHunk()
			i++
			j++
		case dp[at(i+1, j)] >= dp[at(i, j+1)]:
			missing = append(missing, expected[i])
			hunkMissing++
			i++
		default:
			spurious = append(spurious, got[j])
			hunkSpurious++
			j++
		}
	}
	hunkMissing += n - i
	hunkSpurious += m - j
	closeHunk()

	return append(missing, expected[i:]...), append(spurious, got[j:]...), distance
}

// multisetDiff is the fallback for a runaway output: it ignores order and reports
// each side's surplus occurrences.
func multisetDiff(expected, got []string) (missing, spurious []string) {
	counts := map[string]int{}
	for _, line := range got {
		counts[line]++
	}
	for _, line := range expected {
		if counts[line] > 0 {
			counts[line]--
			continue
		}
		missing = append(missing, line)
	}
	remaining := map[string]int{}
	for _, line := range got {
		remaining[line]++
	}
	for _, line := range expected {
		if remaining[line] > 0 {
			remaining[line]--
		}
	}
	for _, line := range got {
		if remaining[line] > 0 {
			remaining[line]--
			spurious = append(spurious, line)
		}
	}
	return missing, spurious
}

// classification is one failing fixture's diff, reduced.
type classification struct {
	kind     string
	missing  []string
	spurious []string
	distance int
}

// classify diffs one failure and buckets it. A fixture that never terminated or that
// crashed the compiler has no meaningful diff, so its distance is its whole
// expectation: nothing about it is one edit from passing.
func classify(expected, got string) classification {
	switch {
	case got == timeoutSentinel:
		return classification{kind: kindTimeout, distance: len(splitLines(expected))}
	case strings.Contains(got, "panic: ") && strings.Contains(got, "goroutine "):
		return classification{kind: kindCrash, distance: len(splitLines(expected))}
	}

	expLines, gotLines := splitLines(expected), splitLines(got)
	missing, spurious, distance := lineDiff(expLines, gotLines)

	kind := kindMixed
	switch {
	case len(gotLines) == 0:
		kind = kindEmpty
	case !anyDiagnostic(missing) && !anyDiagnostic(spurious):
		kind = kindOutput
	case allDiagnostic(missing) && allDiagnostic(spurious):
		kind = kindDiagnostics
	}

	return classification{kind: kind, missing: missing, spurious: spurious, distance: distance}
}

// splitLines keeps blank lines. normalize has already right-trimmed every line and
// stripped the leading and trailing blanks, so a blank line that survives is one the
// program actually printed and the expectation actually wants: `Print("")` between two
// other writes is the whole difference in SimpleScripts/print_multi_args. Dropping
// blanks here would score that fixture as distance 0 — failing, with nothing wrong.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func anyDiagnostic(lines []string) bool {
	for _, line := range lines {
		if isDiagnostic(line) {
			return true
		}
	}
	return false
}

func allDiagnostic(lines []string) bool {
	for _, line := range lines {
		if !isDiagnostic(line) {
			return false
		}
	}
	return true
}

// shapeTally counts one message shape across the failing set.
type shapeTally struct {
	shape string
	lines int
	// fixtures is how many failing fixtures mention the shape at all.
	fixtures int
	// sole is how many of those it is the *only* shape on its side for, with nothing
	// but diagnostics differing anywhere in the fixture — the fixtures that fixing
	// this one shape plausibly closes outright. A diagnostic emitted in the wrong
	// words counts: the wrong wording sits on the other side of the diff and goes
	// away with the fix. fixtures ranks the work; sole predicts the yield, and the
	// gap between them is how much else has to be right first.
	sole int
}

// tallyShapes counts diagnostic shapes on one side of a set of diffs. side picks that
// side; the whole classification is still needed to decide whether a shape is sole.
func tallyShapes(classes []classification, side func(classification) []string) []shapeTally {
	lines := map[string]int{}
	fixtures := map[string]int{}
	sole := map[string]int{}
	for _, c := range classes {
		shapes := map[string]bool{}
		for _, line := range side(c) {
			if !isDiagnostic(line) {
				continue
			}
			shape := shapeOf(line)
			lines[shape]++
			shapes[shape] = true
		}
		for shape := range shapes {
			fixtures[shape]++
		}
		if len(shapes) != 1 || c.kind != kindDiagnostics {
			continue
		}
		for shape := range shapes {
			sole[shape]++
		}
	}
	tallies := make([]shapeTally, 0, len(lines))
	for shape, count := range lines {
		tallies = append(tallies, shapeTally{
			shape: shape, lines: count, fixtures: fixtures[shape], sole: sole[shape],
		})
	}
	sort.Slice(tallies, func(i, j int) bool {
		if tallies[i].fixtures != tallies[j].fixtures {
			return tallies[i].fixtures > tallies[j].fixtures
		}
		if tallies[i].lines != tallies[j].lines {
			return tallies[i].lines > tallies[j].lines
		}
		return tallies[i].shape < tallies[j].shape
	})
	return tallies
}

// shapeTopN is how many message shapes the report ranks. The tail is a long list of
// one-offs; the head is where the work is.
const shapeTopN = 20

// printClassification renders the classification section: where each category's
// failures sit, and which message shapes recur across them. perFixture adds the
// per-fixture distances, which is the list the near-miss queue is built from.
func printClassification(failed []result, perFixture bool) {
	if len(failed) == 0 {
		fmt.Println("\nClassification: nothing failing.")
		return
	}

	type bucket struct {
		kinds               map[string]int
		fails, near1, near2 int
	}
	buckets := map[string]*bucket{}
	var order []string
	var classes []classification

	for _, r := range failed {
		if r.class == nil {
			continue
		}
		b := buckets[r.category]
		if b == nil {
			b = &bucket{kinds: map[string]int{}}
			buckets[r.category] = b
			order = append(order, r.category)
		}
		b.fails++
		b.kinds[r.class.kind]++
		if r.class.distance == 1 {
			b.near1++
		}
		if r.class.distance <= 2 {
			b.near2++
		}
		classes = append(classes, *r.class)
	}
	sort.Strings(order)

	fmt.Printf("\nClassification of %d failing fixtures\n", len(classes))
	fmt.Println("Distance is line-level edit distance: \"=1\" means one output line is wrong or absent.")
	fmt.Printf("\n%-26s%6s%6s%6s%7s%8s%7s%7s%7s%9s\n",
		"Category", "Fail", "=1", "<=2", "diag", "output", "mixed", "empty", "crash", "timeout")
	totals := &bucket{kinds: map[string]int{}}
	for _, cat := range order {
		b := buckets[cat]
		fmt.Printf("%-26s%6d%6d%6d%7d%8d%7d%7d%7d%9d\n", cat, b.fails, b.near1, b.near2,
			b.kinds[kindDiagnostics], b.kinds[kindOutput], b.kinds[kindMixed],
			b.kinds[kindEmpty], b.kinds[kindCrash], b.kinds[kindTimeout])
		totals.fails += b.fails
		totals.near1 += b.near1
		totals.near2 += b.near2
		for _, k := range kindOrder {
			totals.kinds[k] += b.kinds[k]
		}
	}
	fmt.Println(strings.Repeat("-", 89))
	fmt.Printf("%-26s%6d%6d%6d%7d%8d%7d%7d%7d%9d\n", "TOTAL", totals.fails, totals.near1, totals.near2,
		totals.kinds[kindDiagnostics], totals.kinds[kindOutput], totals.kinds[kindMixed],
		totals.kinds[kindEmpty], totals.kinds[kindCrash], totals.kinds[kindTimeout])

	printShapes("Missing diagnostics (expected, never produced)",
		tallyShapes(classes, func(c classification) []string { return c.missing }))
	printShapes("Spurious diagnostics (produced, not expected)",
		tallyShapes(classes, func(c classification) []string { return c.spurious }))

	if perFixture {
		fmt.Println("\nPer-fixture distance:")
		sorted := append([]result(nil), failed...)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].class == nil || sorted[j].class == nil {
				return sorted[j].class == nil
			}
			if sorted[i].class.distance != sorted[j].class.distance {
				return sorted[i].class.distance < sorted[j].class.distance
			}
			return sorted[i].category+"/"+sorted[i].name < sorted[j].category+"/"+sorted[j].name
		})
		for _, r := range sorted {
			if r.class == nil {
				continue
			}
			fmt.Printf("  %4d  %-12s %s/%s\n", r.class.distance, r.class.kind, r.category, r.name)
		}
	}
}

// printShapes writes one ranked shape table, ranked by fixtures blocked rather than
// by occurrences: a single fixture emitting one shape twenty times is still one
// fixture, and it is fixtures that the pass rate counts.
func printShapes(title string, tallies []shapeTally) {
	fmt.Printf("\n%s — top %d of %d shapes\n", title, min(shapeTopN, len(tallies)), len(tallies))
	fmt.Printf("%8s%6s%8s  %s\n", "Fixtures", "Sole", "Lines", "Shape")
	for i, t := range tallies {
		if i >= shapeTopN {
			break
		}
		shape := t.shape
		if len(shape) > 96 {
			shape = shape[:93] + "..."
		}
		fmt.Printf("%8d%6d%8d  %s\n", t.fixtures, t.sole, t.lines, shape)
	}
}
