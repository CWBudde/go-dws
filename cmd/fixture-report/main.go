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
//	go build -o bin/dwscript ./cmd/dwscript
//	go run ./cmd/fixture-report [--category NAME] [--list-fails] [--timeout SECS] [--cli PATH]
//
// Note: the CLI does not emit DWScript's "Errors >>>>" diagnostic envelope, so the *Fail
// error-detection categories score ~0% here. For those, trust TestDWScriptFixtures, which
// compares against the compiler's structured diagnostics.
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
)

const fixturesBase = "testdata/fixtures"

// normalize mirrors scripts' normalization: CRLF→LF, right-trim each line, strip the whole.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// runOne executes the CLI on a single fixture, returning combined stdout+stderr (or a
// sentinel on timeout).
func runOne(cli, pasFile string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cli, "run", pasFile)
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
	flag.Parse()

	if _, err := os.Stat(*cli); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s not found. Build it: go build -o bin/dwscript ./cmd/dwscript\n", *cli)
		return 2
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
	expBytes, err := os.ReadFile(it.txtFile)
	if err != nil {
		return result{category: it.category, name: name, noExp: true}
	}
	expected := normalize(string(expBytes))
	got := normalize(runOne(cli, it.pasFile, timeout))
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
