package interp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/frontend"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// Paths (relative to this package directory) used by the fixture harness.
const (
	fixturesRoot        = "../../testdata/fixtures"
	fixtureBaselinePath = fixturesRoot + "/baselines.json"
	fixtureStatusPath   = fixturesRoot + "/TEST_STATUS.md"
	// Set FIXTURE_UPDATE_BASELINE=1 to regenerate baselines.json and TEST_STATUS.md
	// from the current pass counts instead of gating against them.
	fixtureUpdateEnv = "FIXTURE_UPDATE_BASELINE"
	// fixtureWorkerEnv marks a re-executed test binary as a fixture worker (see
	// TestFixtureWorkerMain). It is an implementation detail of the sandboxing scheme.
	fixtureWorkerEnv = "DWS_FIXTURE_WORKER"
	// fixtureRespSentinel prefixes worker responses on stdout so they can be told apart
	// from Go's own test-framework chatter.
	fixtureRespSentinel = "@@DWSFIXTURE@@ "
	// Per-fixture execution timeout: enough for compute-heavy fixtures in slower
	// environments while still catching infinite loops. This matches the CLI
	// fixture-report default; race-detector builds multiply it (see
	// fixtureTimeoutScale).
	fixtureTimeout = 60 * time.Second * fixtureTimeoutScale
	// How long the parent waits for a worker response before assuming the fixture hung
	// and killing (then restarting) the worker.
	fixtureWorkerTimeout = fixtureTimeout + 3*time.Second
)

// TestDWScriptFixtures runs the comprehensive DWScript test suite (~2,100 tests) and
// enforces a per-category pass-count baseline so that a green CI run means "the language
// works", not "the parts we test work" (see PLAN.md §P0).
//
// Categories are auto-discovered from testdata/fixtures/, so no category can be silently
// omitted. Every category is exercised on every run; individual fixture failures are
// recorded but do NOT fail the build. The build fails only when a category's pass count
// drops below the baseline recorded in testdata/fixtures/baselines.json — i.e. a
// regression. Improvements are logged with a nudge to raise the baseline.
//
// Each fixture is executed inside a re-executed worker subprocess (see TestFixtureWorkerMain)
// so that a runaway loop or a pathological allocation is killed and isolated instead of
// OOM-ing or hanging the whole test binary.
//
// To refresh the baselines and TEST_STATUS.md after intentionally changing behavior:
//
//	FIXTURE_UPDATE_BASELINE=1 go test ./internal/interp -run TestDWScriptFixtures
//	# or: just fixture-update
//
// Individual categories can still be run with:
//
//	go test -v ./internal/interp -run TestDWScriptFixtures/CategoryName
func TestDWScriptFixtures(t *testing.T) {
	updateMode := os.Getenv(fixtureUpdateEnv) == "1"

	categories, err := discoverFixtureCategories(fixturesRoot)
	if err != nil {
		t.Fatalf("Failed to discover fixture categories: %v", err)
	}

	baselines := loadFixtureBaselines(t, updateMode)

	// Run every fixture across all categories through the worker pool, then tally per category.
	results := runFixturesInWorkers(t, buildFixtureWorkList(categories))

	outcomes := make([]categoryOutcome, 0, len(categories))
	totalPassed, totalFailed, totalSkipped, totalTests := 0, 0, 0, 0

	for _, category := range categories {
		category := category
		outcome := tallyCategory(category, results)

		totalPassed += outcome.passed
		totalFailed += outcome.failed
		totalSkipped += outcome.skipped
		totalTests += outcome.total
		outcomes = append(outcomes, outcome)

		t.Run(category.name, func(t *testing.T) {
			gateCategory(t, outcome, baselines[category.name], updateMode)
		})
	}

	t.Logf("Overall: %d passed, %d failed, %d skipped (out of %d total)",
		totalPassed, totalFailed, totalSkipped, totalTests)

	if updateMode {
		if err := writeFixtureBaselines(outcomes); err != nil {
			t.Fatalf("Failed to write %s: %v", fixtureBaselinePath, err)
		}
		if err := writeFixtureStatus(outcomes, totalPassed, totalFailed, totalSkipped, totalTests); err != nil {
			t.Fatalf("Failed to write %s: %v", fixtureStatusPath, err)
		}
		t.Logf("Updated %s and %s from current pass counts.", fixtureBaselinePath, fixtureStatusPath)
	}
}

// buildFixtureWorkList flattens every category's .pas files into a single ordered work list.
func buildFixtureWorkList(categories []fixtureCategory) []fixtureRequest {
	var items []fixtureRequest
	for _, category := range categories {
		hintsLevel := category.hintsLevel
		if hintsLevel == 0 {
			hintsLevel = semantic.HintsLevelPedantic
		}
		for _, pf := range category.pasFiles {
			items = append(items, fixtureRequest{
				Pas:          pf,
				ExpectErrors: category.expectErrors,
				Hints:        int(hintsLevel),
			})
		}
	}
	return items
}

// tallyCategory aggregates worker results for one category into a categoryOutcome.
func tallyCategory(category fixtureCategory, results map[string]fixtureResponse) categoryOutcome {
	outcome := categoryOutcome{category: category}
	for _, pf := range category.pasFiles {
		outcome.total++
		switch testResult(results[pf].Result) {
		case testResultPassed:
			outcome.passed++
		case testResultSkipped:
			outcome.skipped++
		default:
			outcome.failed++
			outcome.failNames = append(outcome.failNames, strings.TrimSuffix(filepath.Base(pf), ".pas"))
		}
	}
	return outcome
}

// gateCategory logs the category result and, outside update mode, fails the build if the
// pass count dropped below the recorded baseline.
func gateCategory(t *testing.T, outcome categoryOutcome, baseline int, updateMode bool) {
	t.Helper()
	t.Logf("Category %s: %d passed, %d failed, %d skipped (%s)",
		outcome.category.name, outcome.passed, outcome.failed, outcome.skipped, outcome.category.description)

	if updateMode {
		return
	}

	switch {
	case outcome.passed < baseline:
		t.Errorf("REGRESSION in %s: %d fixtures pass but baseline is %d (%d newly broken). "+
			"Sample failures: %s\nIf this drop is intentional run `just fixture-update` to reset the baseline.",
			outcome.category.name, outcome.passed, baseline, baseline-outcome.passed,
			sampleFailures(outcome.failNames))
	case outcome.passed > baseline:
		t.Logf("IMPROVEMENT in %s: %d fixtures pass (baseline %d). "+
			"Run `just fixture-update` to raise the baseline and lock in the gain.",
			outcome.category.name, outcome.passed, baseline)
	}
}

// fixtureCategory describes one directory of fixtures under testdata/fixtures.
type fixtureCategory struct {
	name         string
	path         string
	description  string
	pasFiles     []string
	hintsLevel   semantic.HintsLevel
	expectErrors bool
}

// categoryOutcome captures the measured result of running a category.
type categoryOutcome struct {
	failNames []string
	category  fixtureCategory
	total     int
	passed    int
	failed    int
	skipped   int
}

// hintsLevelOverrides lists the few categories DWScript does not run under its pedantic
// hint harness. Everything else defaults to pedantic (matching the reference test runner).
var hintsLevelOverrides = map[string]semantic.HintsLevel{
	"Algorithms":      semantic.HintsLevelNormal,
	"FunctionsString": semantic.HintsLevelNormal,
}

// categoryDescriptions provides human-readable descriptions for TEST_STATUS.md. Categories
// without an entry fall back to a generic label; the harness never depends on this map for
// behavior.
var categoryDescriptions = map[string]string{
	"SimpleScripts":    "Basic language features and scripts",
	"Algorithms":       "Algorithm implementations",
	"ArrayPass":        "Array operations and features",
	"AssociativePass":  "Associative arrays/maps",
	"SetOfPass":        "Set operations",
	"OverloadsPass":    "Function/method overloading",
	"GenericsPass":     "Generic types and methods",
	"HelpersPass":      "Type helpers",
	"LambdaPass":       "Lambda expressions",
	"InterfacesPass":   "Interface declarations and usage",
	"InnerClassesPass": "Nested class declarations",
	"FailureScripts":   "Compilation and runtime error detection",
}

// discoverFixtureCategories enumerates every directory under root that contains .pas files.
// expectErrors is inferred from the category name so error-detection suites are handled
// correctly without a hand-maintained allow-list.
func discoverFixtureCategories(root string) ([]fixtureCategory, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var categories []fixtureCategory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		dir := filepath.Join(root, name)

		pasFiles, err := filepath.Glob(filepath.Join(dir, "*.pas"))
		if err != nil {
			return nil, err
		}
		if len(pasFiles) == 0 {
			continue
		}
		sort.Strings(pasFiles)

		description := categoryDescriptions[name]
		if description == "" {
			description = name + " fixtures"
		}

		categories = append(categories, fixtureCategory{
			name:         name,
			path:         dir,
			description:  description,
			pasFiles:     pasFiles,
			expectErrors: isErrorCategory(name),
			hintsLevel:   hintsLevelOverrides[name],
		})
	}

	sort.Slice(categories, func(i, j int) bool { return categories[i].name < categories[j].name })
	return categories, nil
}

// isErrorCategory reports whether a category holds error-detection ("*Fail") fixtures whose
// expected output is a compiler/runtime diagnostic rather than program output.
func isErrorCategory(name string) bool {
	switch name {
	case "FailureScripts", "COMConnectorFailure":
		return true
	}
	return ident.HasSuffix(name, "Fail")
}

type testResult int

const (
	testResultPassed testResult = iota
	testResultFailed
	testResultSkipped
)

// fixtureRequest is one unit of work sent from the parent to a worker subprocess.
type fixtureRequest struct {
	Pas          string `json:"pas"`
	Hints        int    `json:"hints"`
	ExpectErrors bool   `json:"expect_errors"`
}

// fixtureResponse is a worker's verdict for a single fixture.
type fixtureResponse struct {
	Detail string `json:"detail"`
	Result int    `json:"result"`
}

// TestFixtureWorkerMain is the entry point for a re-executed worker subprocess. It is a
// no-op unless DWS_FIXTURE_WORKER=1 (set only by runFixturesInWorkers). The worker reads
// fixtureRequest JSON lines from stdin and writes fixtureResponse JSON lines (sentinel-
// prefixed) to stdout, running each fixture in-process. Because it is a separate process,
// the parent can kill it if a fixture hangs or blows up memory.
func TestFixtureWorkerMain(t *testing.T) {
	if os.Getenv(fixtureWorkerEnv) != "1" {
		t.Skip("not a fixture worker process")
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			var req fixtureRequest
			var resp fixtureResponse
			if jsonErr := json.Unmarshal(line, &req); jsonErr == nil {
				result, detail := runFixtureTest(req.Pas, req.ExpectErrors, semantic.HintsLevel(req.Hints))
				resp = fixtureResponse{Result: int(result), Detail: detail}
			} else {
				// Emit an explicit failure instead of staying silent, so the parent fails
				// fast with a useful reason rather than waiting for the hang timeout.
				resp = fixtureResponse{Result: int(testResultFailed), Detail: "worker: malformed request JSON: " + jsonErr.Error()}
			}
			payload, _ := json.Marshal(resp)
			_, _ = writer.WriteString(fixtureRespSentinel)
			_, _ = writer.Write(payload)
			_ = writer.WriteByte('\n')
			_ = writer.Flush()
		}
		if err != nil {
			break
		}
	}
}

// runFixturesInWorkers runs every request through a pool of persistent worker subprocesses,
// each isolating fixtures so hangs/OOMs cannot take down the test binary. Results are keyed
// by fixture .pas path.
func runFixturesInWorkers(t *testing.T, items []fixtureRequest) map[string]fixtureResponse {
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8
	}
	if numWorkers < 1 {
		numWorkers = 1
	}
	if len(items) < numWorkers {
		numWorkers = len(items)
	}

	work := make(chan fixtureRequest)
	results := make(map[string]fixtureResponse, len(items))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &fixtureWorker{}
			defer w.stop()
			for req := range work {
				resp := w.run(req)
				mu.Lock()
				results[req.Pas] = resp
				mu.Unlock()
			}
		}()
	}

	for _, it := range items {
		work <- it
	}
	close(work)
	wg.Wait()

	// Any request without a recorded result (should not happen) counts as failed.
	for _, it := range items {
		if _, ok := results[it.Pas]; !ok {
			results[it.Pas] = fixtureResponse{Result: int(testResultFailed), Detail: "no worker result"}
		}
	}
	return results
}

// fixtureWorker owns a single re-executed worker subprocess and restarts it whenever a
// fixture hangs (timeout) or the process dies.
type fixtureWorker struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdinC interface{ Close() error }
	respCh chan []byte
	deadCh chan struct{}
}

// fixtureVerdict is the (result, detail) pair produced when scoring one fixture.
type fixtureVerdict struct {
	detail string
	result testResult
}

// start (re-)launches the worker subprocess.
func (w *fixtureWorker) start() error {
	cmd := exec.Command(os.Args[0], "-test.run=^TestFixtureWorkerMain$", "-test.timeout=0")
	cmd.Env = append(os.Environ(), fixtureWorkerEnv+"=1")
	// Surface worker stderr (panics, runtime diagnostics) to the parent so fixture
	// regressions stay debuggable. Only stdout carries the sentinel-prefixed protocol.
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	w.cmd = cmd
	w.stdin = bufio.NewWriter(stdin)
	w.stdinC = stdin
	w.respCh = make(chan []byte, 1)
	w.deadCh = make(chan struct{})

	respCh, deadCh := w.respCh, w.deadCh
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, fixtureRespSentinel) {
				payload := []byte(strings.TrimPrefix(line, fixtureRespSentinel))
				select {
				case respCh <- payload:
				default:
				}
			}
		}
		close(deadCh)
	}()
	return nil
}

// stop terminates the worker subprocess.
func (w *fixtureWorker) stop() {
	if w.cmd == nil {
		return
	}
	if w.stdinC != nil {
		_ = w.stdinC.Close()
	}
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
	_ = w.cmd.Wait()
	w.cmd = nil
}

// run executes one fixture on the worker, transparently starting/restarting the subprocess.
// A hung fixture is reported as a timeout failure and the worker is recycled.
func (w *fixtureWorker) run(req fixtureRequest) fixtureResponse {
	if w.cmd == nil {
		if err := w.start(); err != nil {
			return fixtureResponse{Result: int(testResultFailed), Detail: "worker start failed: " + err.Error()}
		}
	}

	payload, _ := json.Marshal(req)
	payload = append(payload, '\n')
	if _, err := w.stdin.Write(payload); err != nil {
		w.stop()
		return fixtureResponse{Result: int(testResultFailed), Detail: "worker write failed: " + err.Error()}
	}
	if err := w.stdin.Flush(); err != nil {
		w.stop()
		return fixtureResponse{Result: int(testResultFailed), Detail: "worker flush failed: " + err.Error()}
	}

	select {
	case line := <-w.respCh:
		var resp fixtureResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			w.stop()
			return fixtureResponse{Result: int(testResultFailed), Detail: "bad worker response: " + err.Error()}
		}
		return resp
	case <-w.deadCh:
		w.stop()
		return fixtureResponse{Result: int(testResultFailed), Detail: "worker crashed (likely out of memory)"}
	case <-time.After(fixtureWorkerTimeout):
		w.stop() // recycle: the fixture hung the whole process
		return fixtureResponse{Result: int(testResultFailed), Detail: "timed out after 5s (likely infinite loop)"}
	}
}

func fixtureHintsAndWarnings(compileResult *frontend.Result) []string {
	var hintsAndWarnings []string
	for _, diag := range compileResult.Diagnostics {
		text := diag.String()
		if strings.HasPrefix(text, "Hint:") || strings.HasPrefix(text, "Warning:") {
			hintsAndWarnings = append(hintsAndWarnings, text)
		}
	}
	return hintsAndWarnings
}

// runFixtureTest runs a single fixture in-process and returns its result plus, on failure, a
// short human-readable detail string. It always runs inside a worker subprocess (see
// TestFixtureWorkerMain), so a panic is recovered here but a hang/OOM is handled by the
// parent killing the worker.
func runFixtureTest(pasFile string, expectErrors bool, hintsLevel semantic.HintsLevel) (result testResult, detail string) {
	// Convert a panic in the evaluator into a recorded failure instead of aborting the run.
	defer func() {
		if r := recover(); r != nil {
			result = testResultFailed
			detail = fmt.Sprintf("PANIC: %v\n%s", r, string(debug.Stack()))
		}
	}()

	// Read the .pas source file with encoding detection.
	source, err := detectAndDecodeFile(pasFile)
	if err != nil {
		return testResultFailed, fmt.Sprintf("failed to read source: %v", err)
	}

	// A fixture is only scored when it ships an expected .txt; a missing file means the
	// fixture is intentionally not scored (skipped). Any other read/decode error is a real
	// problem and must fail rather than masquerade as a skip.
	txtFile := strings.TrimSuffix(pasFile, ".pas") + ".txt"
	expectedContent, err := detectAndDecodeFile(txtFile)
	if err != nil {
		// errors.Is sees through the %w wrapping added by detectAndDecodeFile;
		// os.IsNotExist would not and would misreport a missing file as failed.
		if errors.Is(err, os.ErrNotExist) {
			return testResultSkipped, ""
		}
		return testResultFailed, fmt.Sprintf("failed to read expected output: %v", err)
	}

	compileResult := frontend.Compile(source, pasFile, hintsLevel)

	var v fixtureVerdict
	if expectErrors {
		v = scoreErrorFixture(compileResult, expectedContent)
	} else {
		v = scoreSuccessFixture(compileResult, expectedContent)
	}
	return v.result, v.detail
}

// scoreErrorFixture scores an error-detection ("*Fail") fixture against its expected
// diagnostic output.
func scoreErrorFixture(compileResult *frontend.Result, expectedContent string) fixtureVerdict {
	// Compile-time diagnostics: compare them to the expected error listing.
	if compileErrors := compileResult.DiagnosticStrings(); len(compileErrors) > 0 {
		actualErrors := strings.Join(compileErrors, "\n")
		if normalizeOutput(actualErrors) == normalizeOutput(expectedContent) {
			return fixtureVerdict{result: testResultPassed}
		}
		return fixtureVerdict{result: testResultFailed, detail: diffDetail(expectedContent, actualErrors)}
	}

	// Compiled cleanly: the error is expected at runtime.
	buf, value := evalFixture(compileResult)
	if value == nil || value.Type() != "ERROR" {
		return fixtureVerdict{result: testResultFailed, detail: "expected errors but program ran cleanly"}
	}
	actualOutput := runtimeErrorOutput(compileResult, value, buf, expectedContent)
	if normalizeOutput(actualOutput) == normalizeOutput(expectedContent) {
		return fixtureVerdict{result: testResultPassed}
	}
	return fixtureVerdict{result: testResultFailed, detail: diffDetail(expectedContent, actualOutput)}
}

// scoreSuccessFixture scores a success fixture against its expected program output.
func scoreSuccessFixture(compileResult *frontend.Result, expectedContent string) fixtureVerdict {
	if compileResult.HasFatalDiagnostics() {
		return fixtureVerdict{result: testResultFailed,
			detail: "unexpected compile diagnostics:\n" + strings.Join(compileResult.DiagnosticStrings(), "\n")}
	}
	if compileResult.Analyzer != nil && !compileResult.SemanticSuccessful {
		return fixtureVerdict{result: testResultFailed,
			detail: "semantic analysis failed:\n" + strings.Join(compileResult.DiagnosticStrings(), "\n")}
	}

	buf, value := evalFixture(compileResult)
	if value != nil && value.Type() == "ERROR" {
		actualOutput := runtimeErrorOutput(compileResult, value, buf, expectedContent)
		if normalizeOutput(actualOutput) == normalizeOutput(expectedContent) {
			return fixtureVerdict{result: testResultPassed}
		}
		return fixtureVerdict{result: testResultFailed, detail: diffDetail(expectedContent, actualOutput)}
	}

	actualOutput := wrapHintsEnvelope(compileResult, buf.String())
	if normalizeOutput(actualOutput) != normalizeOutput(expectedContent) {
		return fixtureVerdict{result: testResultFailed, detail: diffDetail(expectedContent, actualOutput)}
	}
	return fixtureVerdict{result: testResultPassed}
}

// evalFixture compiles-then-evaluates a program in-process, returning the captured output
// buffer and the resulting value. The worker-subprocess timeout guards against hangs.
func evalFixture(compileResult *frontend.Result) (*bytes.Buffer, Value) {
	var buf bytes.Buffer
	interp := New(&buf)
	if compileResult.SemanticInfo != nil {
		interp.SetSemanticInfo(compileResult.SemanticInfo)
	}
	return &buf, interp.Eval(compileResult.Program)
}

// runtimeErrorOutput formats a runtime error, wrapping it in DWScript's Errors/Result
// envelope when the expected output uses one.
func runtimeErrorOutput(compileResult *frontend.Result, result Value, buf *bytes.Buffer, expectedContent string) string {
	formattedError := formatRuntimeErrorValue(result)
	if !strings.Contains(expectedContent, "Errors >>>>") {
		return formattedError
	}
	var b strings.Builder
	b.WriteString("Errors >>>>\n")
	for _, hint := range fixtureHintsAndWarnings(compileResult) {
		b.WriteString(hint)
		b.WriteString("\n")
	}
	b.WriteString(formattedError)
	b.WriteString("\nResult >>>>\n")
	b.WriteString(buf.String())
	return b.String()
}

// wrapHintsEnvelope wraps program output in DWScript's Errors/Result envelope when the
// compilation produced hints or warnings.
func wrapHintsEnvelope(compileResult *frontend.Result, output string) string {
	hintsAndWarnings := fixtureHintsAndWarnings(compileResult)
	if len(hintsAndWarnings) == 0 {
		return output
	}
	var b strings.Builder
	b.WriteString("Errors >>>>\n")
	for _, hint := range hintsAndWarnings {
		b.WriteString(hint)
		b.WriteString("\n")
	}
	b.WriteString("Result >>>>\n")
	b.WriteString(output)
	return b.String()
}

// diffDetail renders a compact expected/actual mismatch message for logging.
func diffDetail(expected, actual string) string {
	return fmt.Sprintf("output mismatch\n--- expected ---\n%s\n--- actual ---\n%s", expected, actual)
}

// sampleFailures returns up to 20 failing fixture names for a regression message.
func sampleFailures(names []string) string {
	const max = 20
	if len(names) <= max {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:max], ", ") + fmt.Sprintf(", … (+%d more)", len(names)-max)
}

// loadFixtureBaselines reads the per-category baseline pass counts. A missing file is only
// tolerated in update mode (the very first run that creates it).
func loadFixtureBaselines(t *testing.T, updateMode bool) map[string]int {
	data, err := os.ReadFile(fixtureBaselinePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Outside update mode a missing baseline would gate every category against
			// zero, silently turning the P0 regression gate into a no-op. Fail closed so
			// a packaging/rename mistake can't disable the gate; only the initial
			// update-mode run is allowed to proceed without it.
			if !updateMode {
				t.Fatalf("baseline file %s not found; the regression gate cannot run. "+
					"Run `just fixture-update` (FIXTURE_UPDATE_BASELINE=1) to create it.", fixtureBaselinePath)
			}
			return map[string]int{}
		}
		t.Fatalf("Failed to read %s: %v", fixtureBaselinePath, err)
	}

	baselines := map[string]int{}
	if err := json.Unmarshal(data, &baselines); err != nil {
		t.Fatalf("Failed to parse %s: %v", fixtureBaselinePath, err)
	}
	return baselines
}

// writeFixtureBaselines persists the measured per-category pass counts as the new baseline.
func writeFixtureBaselines(outcomes []categoryOutcome) error {
	baselines := make(map[string]int, len(outcomes))
	for _, o := range outcomes {
		baselines[o.category.name] = o.passed
	}
	data, err := json.MarshalIndent(baselines, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(fixtureBaselinePath, data, 0o644)
}

// writeFixtureStatus regenerates testdata/fixtures/TEST_STATUS.md from the current run.
func writeFixtureStatus(outcomes []categoryOutcome, totalPassed, totalFailed, totalSkipped, totalTests int) error {
	var b strings.Builder
	b.WriteString("# Test Status Tracking\n\n")
	b.WriteString("> **Generated file — do not edit by hand.**\n")
	b.WriteString("> Regenerate with `just fixture-update` " +
		"(`FIXTURE_UPDATE_BASELINE=1 go test ./internal/interp -run TestDWScriptFixtures`).\n\n")
	fmt.Fprintf(&b, "**Generated**: %s\n\n", time.Now().UTC().Format("2006-01-02"))

	scored := totalPassed + totalFailed
	pct := 0.0
	if scored > 0 {
		pct = 100 * float64(totalPassed) / float64(scored)
	}
	b.WriteString("## Overall\n\n")
	b.WriteString("| Metric | Value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Categories | %d |\n", len(outcomes))
	fmt.Fprintf(&b, "| Fixtures (total) | %d |\n", totalTests)
	fmt.Fprintf(&b, "| Passed | %d |\n", totalPassed)
	fmt.Fprintf(&b, "| Failed | %d |\n", totalFailed)
	fmt.Fprintf(&b, "| Skipped (no expected .txt) | %d |\n", totalSkipped)
	fmt.Fprintf(&b, "| **Scored pass rate** | **%.0f%%** (%d/%d) |\n\n", pct, totalPassed, scored)

	b.WriteString("## Per-category\n\n")
	b.WriteString("Pass% is over *scored* fixtures (those with an expected `.txt`).\n\n")
	b.WriteString("| Category | Total | Pass | Fail | Skip | Pass% |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|\n")
	for _, o := range outcomes {
		catScored := o.passed + o.failed
		catPct := 0.0
		if catScored > 0 {
			catPct = 100 * float64(o.passed) / float64(catScored)
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %.0f%% |\n",
			o.category.name, o.total, o.passed, o.failed, o.skipped, catPct)
	}

	return os.WriteFile(fixtureStatusPath, []byte(b.String()), 0o644)
}
