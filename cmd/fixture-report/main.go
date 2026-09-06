// Command fixture-report prints an honest DWScript fixture compatibility report.
//
// It runs every testdata/fixtures/*/*.pas through the built dwscript CLI and compares the
// normalized output to the sibling .txt. It prints a per-category pass/fail table and a
// total. This is the *ground-truth* compatibility metric for the port: unlike the in-repo
// Go test harness (internal/interp.TestDWScriptFixtures), it does not skip categories and it
// exercises the real CLI end to end.
//
// Usage:
//
//	go run ./cmd/fixture-report [--category NAME] [--list-fails] [--timeout SECS] [--cli PATH]
//	                            [--build=false] [--allow-stale]
//
// The CLI is run in harness mode with the same per-category hint level as
// TestDWScriptFixtures: `run --diagnostics=plain --test-envelope --hints LEVEL` for
// execution suites (runtime errors and hints inside the "Errors >>>>" envelope) and
// `run --diagnostics=plain --compile-only --hints LEVEL` for the *Fail error-detection
// suites (the compiler's message list, nothing executed), so both runners score on the
// same terms as DWScript's own test runner.
//
// By default the CLI binary is rebuilt before the run. With --build=false the binary must
// be newer than every tracked Go source file, or the report refuses to run (a stale binary
// silently produces wrong numbers); --allow-stale overrides that check.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/pkg/ident"
)

const fixturesBase = "testdata/fixtures"

// hintsLevelOverrides mirrors internal/interp/fixture_test.go (hintsLevelOverrides): the
// reference harness runs everything at pedantic except these categories. Keep the two in
// sync; the harness lives in a _test file and cannot be imported from here.
var hintsLevelOverrides = map[string]string{
	"Algorithms":      "normal",
	"FunctionsString": "normal",
}

// isErrorCategory mirrors internal/interp/fixture_test.go: error-detection suites are
// compiled only and compared against the compiler's message list, like DWScript's
// CompilationFailure runner.
func isErrorCategory(category string) bool {
	return category == "FailureScripts" || category == "COMConnectorFailure" ||
		ident.HasSuffix(category, "Fail")
}

// hintsLevelFor returns the --hints level the CLI must run a category at.
func hintsLevelFor(category string) string {
	if level, ok := hintsLevelOverrides[category]; ok {
		return level
	}
	return "pedantic"
}

// buildCLI rebuilds the dwscript binary at path from the current sources.
func buildCLI(path string) error {
	cmd := exec.Command("go", "build", "-o", path, "./cmd/dwscript")
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	return cmd.Run()
}

// trackedGoSources lists the Go sources the dwscript binary depends on: git-tracked
// non-test *.go outside this tool's own directory, plus go.mod and go.sum. Without git it
// falls back to walking cmd/, internal/ and pkg/.
func trackedGoSources() []string {
	var files []string
	keep := func(f string) bool {
		return strings.HasSuffix(f, ".go") && !strings.HasSuffix(f, "_test.go") &&
			!strings.HasPrefix(filepath.ToSlash(f), "cmd/fixture-report/")
	}
	out, err := exec.Command("git", "ls-files", "-z", "--", "*.go").Output()
	if err == nil {
		for _, f := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
			if keep(f) {
				files = append(files, f)
			}
		}
		return append(files, "go.mod", "go.sum")
	}
	for _, root := range []string{"cmd", "internal", "pkg"} {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && keep(path) {
				files = append(files, path)
			}
			return nil
		})
	}
	return append(files, "go.mod", "go.sum")
}

// binaryIsStale reports whether any of sources is newer than the binary at bin, returning
// the newest such source. Sources that do not exist are ignored.
func binaryIsStale(bin string, sources []string) (stale bool, newest string, err error) {
	info, err := os.Stat(bin)
	if err != nil {
		return false, "", err
	}
	binTime := info.ModTime()
	var newestTime time.Time
	for _, src := range sources {
		si, err := os.Stat(src)
		if err != nil {
			continue
		}
		if si.ModTime().After(binTime) && si.ModTime().After(newestTime) {
			newestTime = si.ModTime()
			newest = src
		}
	}
	return newest != "", newest, nil
}

// normalize mirrors scripts' normalization: CRLF→LF, right-trim each line, strip the whole.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// runOne executes the CLI on a single fixture in harness mode, returning combined
// stdout+stderr (or a sentinel on timeout).
func runOne(cli, category, pasFile string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mode := "--test-envelope"
	if isErrorCategory(category) {
		mode = "--compile-only"
	}
	cmd := exec.CommandContext(ctx, cli, "run",
		"--diagnostics=plain", mode, "--hints", hintsLevelFor(category), pasFile)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "__TIMEOUT__"
	}
	// A non-zero exit (script emitted errors) is expected and its output is what we compare.
	// Only a failure to launch the CLI (missing binary, etc.) is surfaced distinctly.
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return "__CLI_ERROR__: " + err.Error()
	}
	return string(out)
}

// categoryStat accumulates per-category counts.
type categoryStat struct {
	total int
	pass  int
	fail  int
	noExp int
}

// workItem is one fixture to evaluate.
type workItem struct {
	category string
	pasFile  string
	txtFile  string
}

// result is the verdict for one work item.
type result struct {
	category string
	name     string
	pass     bool
	fail     bool
	noExp    bool
}

func main() {
	os.Exit(run())
}

func run() int {
	category := flag.String("category", "", "only run this category")
	listFails := flag.Bool("list-fails", false, "print failing fixture names")
	timeoutSecs := flag.Int("timeout", 20, "per-fixture timeout in seconds")
	cli := flag.String("cli", "./bin/dwscript", "path to the dwscript CLI binary")
	build := flag.Bool("build", true, "rebuild the CLI binary from the current sources before running")
	allowStale := flag.Bool("allow-stale", false, "with --build=false, run even if the binary is older than the Go sources")
	flag.Parse()

	if *build {
		if err := buildCLI(*cli); err != nil {
			fmt.Fprintf(os.Stderr, "error: building %s failed: %v\n", *cli, err)
			return 2
		}
		fmt.Fprintf(os.Stderr, "built %s\n", *cli)
	} else {
		stale, newest, err := binaryIsStale(*cli, trackedGoSources())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s not found. Build it: just build (or drop --build=false)\n", *cli)
			return 2
		}
		if stale && !*allowStale {
			fmt.Fprintf(os.Stderr, "error: %s is older than %s; rebuild it (just build) or pass --allow-stale\n", *cli, newest)
			return 2
		}
	}

	items, err := collectItems(*category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	results := evaluate(*cli, items, time.Duration(*timeoutSecs)*time.Second)

	// Aggregate per category, preserving a stable (sorted) category order.
	stats := map[string]*categoryStat{}
	var order []string
	var fails []string
	for _, r := range results {
		st := stats[r.category]
		if st == nil {
			st = &categoryStat{}
			stats[r.category] = st
			order = append(order, r.category)
		}
		st.total++
		switch {
		case r.noExp:
			st.noExp++
		case r.pass:
			st.pass++
		default:
			st.fail++
			fails = append(fails, r.category+"/"+r.name)
		}
	}
	sort.Strings(order)
	sort.Strings(fails)

	printReport(order, stats)

	if *listFails {
		fmt.Println("\nFailing fixtures:")
		for _, name := range fails {
			fmt.Printf("  %s\n", name)
		}
	}
	return 0
}

// collectItems builds the work list of fixtures (those with an expected .txt are scored;
// those without are counted as NoExp).
func collectItems(only string) ([]workItem, error) {
	entries, err := os.ReadDir(fixturesBase)
	if err != nil {
		return nil, err
	}

	var items []workItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cat := entry.Name()
		if only != "" && cat != only {
			continue
		}
		pasFiles, err := filepath.Glob(filepath.Join(fixturesBase, cat, "*.pas"))
		if err != nil {
			return nil, err
		}
		sort.Strings(pasFiles)
		for _, pf := range pasFiles {
			items = append(items, workItem{
				category: cat,
				pasFile:  pf,
				txtFile:  strings.TrimSuffix(pf, ".pas") + ".txt",
			})
		}
	}
	return items, nil
}

// evaluate runs the work list through a bounded worker pool and returns one result each,
// in input order (results[i] corresponds to items[i]).
func evaluate(cli string, items []workItem, timeout time.Duration) []result {
	workers := runtime.NumCPU()
	if workers > 8 {
		workers = 8
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(items) {
		workers = len(items)
	}

	type job struct {
		item workItem
		idx  int
	}
	jobs := make(chan job)
	results := make([]result, len(items))
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.idx] = evaluateOne(cli, j.item, timeout)
			}
		}()
	}
	for i, it := range items {
		jobs <- job{idx: i, item: it}
	}
	close(jobs)
	wg.Wait()
	return results
}

// evaluateOne scores a single fixture.
func evaluateOne(cli string, it workItem, timeout time.Duration) result {
	name := strings.TrimSuffix(filepath.Base(it.pasFile), ".pas")
	expContent, err := encoding.DecodeFile(it.txtFile)
	if err != nil {
		// Only a missing .txt means "not scored". Any other read/decode error
		// (e.g. malformed UTF-16) is a real problem and must count as a
		// failure instead of silently shrinking the scored set.
		if errors.Is(err, os.ErrNotExist) {
			return result{category: it.category, name: name, noExp: true}
		}
		fmt.Fprintf(os.Stderr, "warning: %s: cannot decode expected output: %v\n", it.txtFile, err)
		return result{category: it.category, name: name, fail: true}
	}
	expected := normalize(expContent)
	got := normalize(runOne(cli, it.category, it.pasFile, timeout))
	if got == expected {
		return result{category: it.category, name: name, pass: true}
	}
	return result{category: it.category, name: name, fail: true}
}

// printReport writes the per-category table and the total row.
func printReport(order []string, stats map[string]*categoryStat) {
	fmt.Printf("%-26s%5s%6s%6s%6s%7s\n", "Category", "Tot", "Pass", "Fail", "NoExp", "Pass%")
	var tPass, tFail, tNoExp, tTot int
	for _, cat := range order {
		st := stats[cat]
		scored := st.pass + st.fail
		pct := 0.0
		if scored > 0 {
			pct = 100 * float64(st.pass) / float64(scored)
		}
		fmt.Printf("%-26s%5d%6d%6d%6d%6.0f%%\n", cat, st.total, st.pass, st.fail, st.noExp, pct)
		tPass += st.pass
		tFail += st.fail
		tNoExp += st.noExp
		tTot += st.total
	}
	scored := tPass + tFail
	total := 0.0
	if scored > 0 {
		total = 100 * float64(tPass) / float64(scored)
	}
	fmt.Println(strings.Repeat("-", 56))
	fmt.Printf("%-26s%5d%6d%6d%6d%6.0f%%\n", "TOTAL", tTot, tPass, tFail, tNoExp, total)
}
