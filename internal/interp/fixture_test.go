package interp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestDWScriptFixtures runs all DWScript test fixtures from the reference repository
// using go-snaps for snapshot testing. This provides comprehensive coverage of
// DWScript language features based on the original test suite.
func TestDWScriptFixtures(t *testing.T) {
	// Define test categories and their expected behavior
	// Includes all 64 test categories from the original DWScript test suite
	testCategories := []struct {
		name            string
		path            string
		description     string
		expectErrors    bool
		requiresLibs    bool
		requiresCodegen bool
		skip            bool
		hintsLevel      semantic.HintsLevel
	}{
		// Core Language Tests - Pass Cases
		{
			name:         "SimpleScripts",
			path:         "../../testdata/fixtures/SimpleScripts",
			expectErrors: false,
			description:  "Basic language features and scripts",
			skip:         false, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "Algorithms",
			path:         "../../testdata/fixtures/Algorithms",
			expectErrors: false,
			description:  "Algorithm implementations",
			skip:         false,                     // TODO: Re-enable after implementing missing features
			hintsLevel:   semantic.HintsLevelNormal, // Algorithms aren't part of DWScript's pedantic harness
		},
		{
			name:         "ArrayPass",
			path:         "../../testdata/fixtures/ArrayPass",
			expectErrors: false,
			description:  "Array operations and features",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "AssociativePass",
			path:         "../../testdata/fixtures/AssociativePass",
			expectErrors: false,
			description:  "Associative arrays/maps",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "SetOfPass",
			path:         "../../testdata/fixtures/SetOfPass",
			expectErrors: false,
			description:  "Set operations",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "OverloadsPass",
			path:         "../../testdata/fixtures/OverloadsPass",
			expectErrors: false,
			description:  "Function/method overloading (39 tests)",
			skip:         false, // Enabled for Phase 9 Stage 6 testing
		},
		{
			name:         "OperatorOverloadPass",
			path:         "../../testdata/fixtures/OperatorOverloadPass",
			expectErrors: false,
			description:  "Operator overloading",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "GenericsPass",
			path:         "../../testdata/fixtures/GenericsPass",
			expectErrors: false,
			description:  "Generic types and methods",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "HelpersPass",
			path:         "../../testdata/fixtures/HelpersPass",
			expectErrors: false,
			description:  "Type helpers",
			skip:         false,
		},
		{
			name:         "LambdaPass",
			path:         "../../testdata/fixtures/LambdaPass",
			expectErrors: false,
			description:  "Lambda expressions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "PropertyExpressionsPass",
			path:         "../../testdata/fixtures/PropertyExpressionsPass",
			expectErrors: false,
			description:  "Property expressions",
			skip:         true, // Most tests require unimplemented features (const expr, helpers, etc)
		},
		{
			name:         "InterfacesPass",
			path:         "../../testdata/fixtures/InterfacesPass",
			expectErrors: false,
			description:  "Interface declarations and usage",
			skip:         false,
		},
		{
			name:         "InnerClassesPass",
			path:         "../../testdata/fixtures/InnerClassesPass",
			expectErrors: false,
			description:  "Nested class declarations",
			skip:         false,
		},

		// Core Language Tests - Failure Cases
		{
			name:         "FailureScripts",
			path:         "../../testdata/fixtures/FailureScripts",
			expectErrors: true,
			description:  "Compilation and runtime errors",
			skip:         false, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "AssociativeFail",
			path:         "../../testdata/fixtures/AssociativeFail",
			expectErrors: true,
			description:  "Associative array error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "SetOfFail",
			path:         "../../testdata/fixtures/SetOfFail",
			expectErrors: true,
			description:  "Set operation error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "OverloadsFail",
			path:         "../../testdata/fixtures/OverloadsFail",
			expectErrors: true,
			description:  "Overloading error cases",
			skip:         false, //  Enabled to validate error messages
		},
		{
			name:         "OperatorOverloadFail",
			path:         "../../testdata/fixtures/OperatorOverloadFail",
			expectErrors: true,
			description:  "Operator overload error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "GenericsFail",
			path:         "../../testdata/fixtures/GenericsFail",
			expectErrors: true,
			description:  "Generic type error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "HelpersFail",
			path:         "../../testdata/fixtures/HelpersFail",
			expectErrors: true,
			description:  "Type helper error cases",
			skip:         false,
		},
		{
			name:         "LambdaFail",
			path:         "../../testdata/fixtures/LambdaFail",
			expectErrors: true,
			description:  "Lambda expression error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "PropertyExpressionsFail",
			path:         "../../testdata/fixtures/PropertyExpressionsFail",
			expectErrors: true,
			description:  "Property expression error cases",
			skip:         true, // Most tests require unimplemented features (readonly, write expr, etc)
		},
		{
			name:         "InterfacesFail",
			path:         "../../testdata/fixtures/InterfacesFail",
			expectErrors: true,
			description:  "Interface error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "InnerClassesFail",
			path:         "../../testdata/fixtures/InnerClassesFail",
			expectErrors: true,
			description:  "Nested class error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "AttributesFail",
			path:         "../../testdata/fixtures/AttributesFail",
			expectErrors: true,
			description:  "Attribute error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},

		// Built-in Functions - Pass Cases
		{
			name:         "FunctionsMath",
			path:         "../../testdata/fixtures/FunctionsMath",
			expectErrors: false,
			description:  "Mathematical functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsMath3D",
			path:         "../../testdata/fixtures/FunctionsMath3D",
			expectErrors: false,
			description:  "3D math functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsMathComplex",
			path:         "../../testdata/fixtures/FunctionsMathComplex",
			expectErrors: false,
			description:  "Complex number functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsString",
			path:         "../../testdata/fixtures/FunctionsString",
			expectErrors: false,
			description:  "String manipulation functions",
			skip:         false,                     // Enabled for task 9.17.3
			hintsLevel:   semantic.HintsLevelNormal, // FunctionsString tests aren't part of DWScript's pedantic harness
		},
		{
			name:         "FunctionsTime",
			path:         "../../testdata/fixtures/FunctionsTime",
			expectErrors: false,
			description:  "Date/time functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsByteBuffer",
			path:         "../../testdata/fixtures/FunctionsByteBuffer",
			expectErrors: false,
			description:  "Byte buffer operations",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsFile",
			path:         "../../testdata/fixtures/FunctionsFile",
			expectErrors: false,
			description:  "File I/O functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsGlobalVars",
			path:         "../../testdata/fixtures/FunctionsGlobalVars",
			expectErrors: false,
			description:  "Global variable functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsVariant",
			path:         "../../testdata/fixtures/FunctionsVariant",
			expectErrors: false,
			description:  "Variant type functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsRTTI",
			path:         "../../testdata/fixtures/FunctionsRTTI",
			expectErrors: false,
			description:  "Runtime type information functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "FunctionsDebug",
			path:         "../../testdata/fixtures/FunctionsDebug",
			expectErrors: false,
			description:  "Debug/diagnostic functions",
			skip:         true, // TODO: Re-enable after implementing missing features
		},

		// Library Tests - Require External Dependencies
		{
			name:         "ClassesLib",
			path:         "../../testdata/fixtures/ClassesLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "Classes library tests",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "JSONConnectorPass",
			path:         "../../testdata/fixtures/JSONConnectorPass",
			expectErrors: false,
			description:  "JSON parsing and generation",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "JSONConnectorFail",
			path:         "../../testdata/fixtures/JSONConnectorFail",
			expectErrors: true,
			description:  "JSON error cases",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "LinqJSON",
			path:         "../../testdata/fixtures/LinqJSON",
			expectErrors: false,
			description:  "LINQ-style JSON queries",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "Linq",
			path:         "../../testdata/fixtures/Linq",
			expectErrors: false,
			description:  "LINQ-style queries",
			skip:         true, // TODO: Re-enable after implementing missing features
		},
		{
			name:         "DOMParser",
			path:         "../../testdata/fixtures/DOMParser",
			expectErrors: false,
			description:  "XML/DOM parsing",
			skip:         true,
		},
		{
			name:         "DelegateLib",
			path:         "../../testdata/fixtures/DelegateLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "Delegate library tests",
			skip:         true,
		},
		{
			name:         "DataBaseLib",
			path:         "../../testdata/fixtures/DataBaseLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "Database operations - requires sqlite3.dll",
			skip:         true,
		},
		{
			name:         "COMConnector",
			path:         "../../testdata/fixtures/COMConnector",
			expectErrors: false,
			requiresLibs: true,
			description:  "COM interop tests - Windows only",
			skip:         true,
		},
		{
			name:         "COMConnectorFailure",
			path:         "../../testdata/fixtures/COMConnectorFailure",
			expectErrors: true,
			requiresLibs: true,
			description:  "COM error cases - Windows only",
			skip:         true,
		},
		{
			name:         "EncodingLib",
			path:         "../../testdata/fixtures/EncodingLib",
			expectErrors: false,
			description:  "Encoding/decoding functions",
			skip:         true,
		},
		{
			name:         "CryptoLib",
			path:         "../../testdata/fixtures/CryptoLib",
			expectErrors: false,
			description:  "Cryptographic functions",
			skip:         true,
		},
		{
			name:         "TabularLib",
			path:         "../../testdata/fixtures/TabularLib",
			expectErrors: false,
			description:  "Tabular data operations",
			skip:         true,
		},
		{
			name:         "TimeSeriesLib",
			path:         "../../testdata/fixtures/TimeSeriesLib",
			expectErrors: false,
			description:  "Time series data",
			skip:         true,
		},
		{
			name:         "SystemInfoLib",
			path:         "../../testdata/fixtures/SystemInfoLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "System information",
			skip:         true,
		},
		{
			name:         "IniFileLib",
			path:         "../../testdata/fixtures/IniFileLib",
			expectErrors: false,
			description:  "INI file operations",
			skip:         true,
		},
		{
			name:         "WebLib",
			path:         "../../testdata/fixtures/WebLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "Web/HTTP operations",
			skip:         true,
		},
		{
			name:         "GraphicsLib",
			path:         "../../testdata/fixtures/GraphicsLib",
			expectErrors: false,
			requiresLibs: true,
			description:  "Graphics operations",
			skip:         true,
		},

		// Advanced Features
		{
			name:         "BigInteger",
			path:         "../../testdata/fixtures/BigInteger",
			expectErrors: false,
			description:  "Arbitrary precision integers",
			skip:         true,
		},
		{
			name:         "Memory",
			path:         "../../testdata/fixtures/Memory",
			expectErrors: false,
			description:  "Memory management tests",
			skip:         true,
		},
		{
			name:         "AutoFormat",
			path:         "../../testdata/fixtures/AutoFormat",
			expectErrors: false,
			description:  "Code auto-formatting",
			skip:         true,
		},

		/*
			// Codegen Tests - Require Stage 12 Implementation
			{
				name:            "BuildScripts",
				path:            "../../testdata/fixtures/BuildScripts",
				expectErrors:    false,
				requiresCodegen: true,
				description:     "Build and compilation tests - includes JS transpilation",
				skip:            true,
			},
			{
				name:            "JSFilterScripts",
				path:            "../../testdata/fixtures/JSFilterScripts",
				expectErrors:    false,
				requiresCodegen: true,
				description:     "JavaScript filter scripts",
			},
			{
				name:            "JSFilterScriptsFail",
				path:            "../../testdata/fixtures/JSFilterScriptsFail",
				expectErrors:    true,
				requiresCodegen: true,
				description:     "JavaScript filter error cases (5 .dws files)",
			},
			{
				name:            "HTMLFilterScripts",
				path:            "../../testdata/fixtures/HTMLFilterScripts",
				expectErrors:    false,
				requiresCodegen: true,
				description:     "HTML filter scripts",
			},
		*/
	}

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	for _, category := range testCategories {
		t.Run(category.name, func(t *testing.T) {
			categoryPassed := 0
			categoryFailed := 0
			categorySkipped := 0

			// Skip categories marked for temporary skip
			if category.skip {
				t.Skipf("Test category %s temporarily skipped", category.name)
				return
			}

			// Skip categories that require features not yet implemented
			if category.requiresCodegen {
				t.Skipf("Test category %s requires JavaScript codegen (Stage 12) - skipping", category.name)
				return
			}

			// Check if the test directory exists
			if _, err := os.Stat(category.path); os.IsNotExist(err) {
				t.Skipf("Test category %s not found at %s", category.name, category.path)
				return
			}

			// Find all .pas files in the category
			pasFiles, err := filepath.Glob(filepath.Join(category.path, "*.pas"))
			if err != nil {
				t.Fatalf("Failed to find test files in %s: %v", category.path, err)
			}

			if len(pasFiles) == 0 {
				t.Skipf("No .pas files found in %s", category.path)
				return
			}

			// Run ALL test files (no whitelist filtering - we want to see what fails)
			for _, pasFile := range pasFiles {
				testName := strings.TrimSuffix(filepath.Base(pasFile), ".pas")
				totalTests++
				hintsLevel := category.hintsLevel
				if hintsLevel == 0 {
					hintsLevel = semantic.HintsLevelPedantic
				}

				t.Run(testName, func(t *testing.T) {
					result := runFixtureTest(t, pasFile, category.expectErrors, hintsLevel)
					switch result {
					case testResultPassed:
						categoryPassed++
						passedTests++
					case testResultFailed:
						categoryFailed++
						failedTests++
					case testResultSkipped:
						categorySkipped++
						skippedTests++
					}
				})
			}

			t.Logf("Category %s: %d passed, %d failed, %d skipped (%s)",
				category.name, categoryPassed, categoryFailed, categorySkipped, category.description)
		})
	}

	t.Logf("Overall: %d passed, %d failed, %d skipped (out of %d total)",
		passedTests, failedTests, skippedTests, totalTests)
}

type testResult int

const (
	testResultPassed testResult = iota
	testResultFailed
	testResultSkipped
)

// runFixtureTest runs a single fixture test
func runFixtureTest(t *testing.T, pasFile string, expectErrors bool, hintsLevel semantic.HintsLevel) testResult {
	// Add panic recovery to identify which test is crashing
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PANIC in %s: %v\n\nStack trace:\n%s",
				filepath.Base(pasFile), r, string(debug.Stack()))
		}
	}()

	// Log which test is being executed (helpful for debugging)
	t.Logf("Executing: %s", filepath.Base(pasFile))

	// Read the .pas source file with encoding detection
	source, err := detectAndDecodeFile(pasFile)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", pasFile, err)
	}

	// Check if there's an expected output/error file
	txtFile := strings.TrimSuffix(pasFile, ".pas") + ".txt"
	var expectedContent string
	var hasExpectedFile bool

	if content, err := detectAndDecodeFile(txtFile); err == nil {
		expectedContent = content
		hasExpectedFile = true
	}

	// Parse the source
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	// Collect parse errors
	var parseErrors []string
	if len(p.Errors()) > 0 {
		parseErrors = make([]string, len(p.Errors()))
		for i, err := range p.Errors() {
			parseErrors[i] = err.Error()
		}
	}

	// If we expect errors and got parse errors, this might be a successful failure test
	if expectErrors && len(parseErrors) > 0 {
		if hasExpectedFile {
			// Compare parse errors with expected content
			actualErrors := strings.Join(parseErrors, "\n")
			if normalizeOutput(actualErrors) == normalizeOutput(expectedContent) {
				return testResultPassed
			} else {
				t.Errorf("Parse error mismatch for %s:\nExpected:\n%s\nActual:\n%s",
					filepath.Base(pasFile), expectedContent, actualErrors)
				return testResultFailed
			}
		}
		// No expected file but got errors - consider it passed for failure tests
		return testResultPassed
	}

	// If we expect errors but got no parse errors, run semantic analysis and try execution
	if expectErrors && len(parseErrors) == 0 {
		// Run semantic analysis to catch semantic errors (overload violations, type errors, etc.)
		// Task 6.1.2: Enable experimental passes for multi-pass analysis architecture (on demand!)
		// analyzer := semantic.NewAnalyzerWithExperimentalPasses()
		analyzer := semantic.NewAnalyzer()
		analyzer.SetHintsLevel(hintsLevel)
		analyzer.SetSource(source, pasFile)

		var semanticErrors []string
		if err := analyzer.Analyze(program); err != nil {
			semanticErrors = analyzer.Errors()
		}

		// If we got semantic errors, check them against expected
		if len(semanticErrors) > 0 {
			if hasExpectedFile {
				actualErrors := strings.Join(semanticErrors, "\n")
				if normalizeOutput(actualErrors) == normalizeOutput(expectedContent) {
					return testResultPassed
				} else {
					t.Errorf("Runtime error mismatch for %s:\nExpected:\n%s\nActual:\n%s",
						filepath.Base(pasFile), expectedContent, actualErrors)
					return testResultFailed
				}
			}
			return testResultPassed
		}

		// No semantic errors, try execution for runtime errors
		var buf bytes.Buffer
		interp := New(&buf)
		// Task 9.5.4: Pass semantic info to interpreter for class variable access
		if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
			interp.SetSemanticInfo(semanticInfo)
		}

		// Execute with timeout to prevent infinite loops from hanging tests
		// Timeout is set to 5 seconds - enough for normal tests, catches infinite loops
		type evalResult struct {
			value Value
		}
		resultChan := make(chan evalResult, 1)

		go func() {
			resultChan <- evalResult{value: interp.Eval(program)}
		}()

		var result Value
		select {
		case res := <-resultChan:
			result = res.value
		case <-time.After(5 * time.Second):
			// Timeout - likely an infinite loop
			t.Errorf("Test %s timed out after 5 seconds (likely infinite loop)", filepath.Base(pasFile))
			return testResultFailed
		}

		if result != nil && result.Type() == "ERROR" {
			// Got runtime error as expected
			if hasExpectedFile {
				formattedError := formatRuntimeErrorValue(result)
				actualOutput := formattedError
				if strings.Contains(expectedContent, "Errors >>>>") {
					actualOutput = "Errors >>>>\n" + formattedError + "\nResult >>>>\n" + buf.String()
				}

				if normalizeOutput(actualOutput) == normalizeOutput(expectedContent) {
					return testResultPassed
				} else {
					t.Errorf("Runtime error mismatch for %s:\nExpected:\n%s\nActual:\n%s",
						filepath.Base(pasFile), expectedContent, actualOutput)
					return testResultFailed
				}
			}
			return testResultPassed
		}

		// Expected errors but got none - this is a failure
		t.Errorf("Expected errors for %s but got none", filepath.Base(pasFile))
		return testResultFailed
	}

	// For success tests, we expect no parse errors
	if len(parseErrors) > 0 {
		t.Errorf("Unexpected parse errors for %s: %v", filepath.Base(pasFile), parseErrors)
		return testResultFailed
	}

	// Run semantic analysis before execution
	// This enables proper type checking and disambiguation of array vs set literals
	// Task 6.1.2: Enable experimental passes for multi-pass analysis architecture
	// analyzer := semantic.NewAnalyzerWithExperimentalPasses()
	analyzer := semantic.NewAnalyzer()
	analyzer.SetHintsLevel(hintsLevel)
	analyzer.SetSource(source, pasFile)

	if err := analyzer.Analyze(program); err != nil {
		// For success tests, semantic errors are failures
		// Format errors nicely for debugging
		var errorMsg strings.Builder
		errorMsg.WriteString("Semantic analysis failed:\n")
		for _, semErr := range analyzer.Errors() {
			errorMsg.WriteString("  - ")
			errorMsg.WriteString(semErr)
			errorMsg.WriteString("\n")
		}
		t.Errorf("%s: %s", filepath.Base(pasFile), errorMsg.String())
		return testResultFailed
	}

	// Execute the program with timeout
	var buf bytes.Buffer
	interp := New(&buf)
	// Task 9.5.4: Pass semantic info to interpreter for class variable access
	if semanticInfo := analyzer.GetSemanticInfo(); semanticInfo != nil {
		interp.SetSemanticInfo(semanticInfo)
	}

	// Execute with timeout to prevent infinite loops from hanging tests
	type evalResult struct {
		value Value
	}
	resultChan := make(chan evalResult, 1)

	go func() {
		resultChan <- evalResult{value: interp.Eval(program)}
	}()

	var result Value
	select {
	case res := <-resultChan:
		result = res.value
	case <-time.After(5 * time.Second):
		// Timeout - likely an infinite loop
		t.Errorf("Test %s timed out after 5 seconds (likely infinite loop)", filepath.Base(pasFile))
		return testResultFailed
	}

	// Check for runtime errors
	if result != nil && result.Type() == "ERROR" {
		if hasExpectedFile {
			formattedError := formatRuntimeErrorValue(result)

			// DWScript fixtures wrap runtime errors in an Errors/Result block
			actualOutput := formattedError
			if strings.Contains(expectedContent, "Errors >>>>") {
				actualOutput = "Errors >>>>\n" + formattedError + "\nResult >>>>\n" + buf.String()
			}

			if normalizeOutput(actualOutput) == normalizeOutput(expectedContent) {
				return testResultPassed
			}

			t.Errorf("Runtime error mismatch for %s:\nExpected:\n%s\nActual:\n%s",
				filepath.Base(pasFile), expectedContent, actualOutput)
			return testResultFailed
		}

		t.Errorf("Runtime error in %s: %v", filepath.Base(pasFile), result.String())
		return testResultFailed
	}

	// Capture output and prepend any hints from semantic analysis
	actualOutput := buf.String()

	// Check if there are hints or warnings (but not actual errors) from semantic analysis
	analyzerErrors := analyzer.Errors()
	var hintsAndWarnings []string
	for _, err := range analyzerErrors {
		if strings.HasPrefix(err, "Hint:") || strings.HasPrefix(err, "Warning:") {
			hintsAndWarnings = append(hintsAndWarnings, err)
		}
	}

	// If there are hints or warnings, format output as "Errors >>>>\n<hints/warnings>\nResult >>>>\n<output>"
	if len(hintsAndWarnings) > 0 {
		var formattedOutput strings.Builder
		formattedOutput.WriteString("Errors >>>>\n")
		for _, hint := range hintsAndWarnings {
			formattedOutput.WriteString(hint)
			formattedOutput.WriteString("\n")
		}
		formattedOutput.WriteString("Result >>>>\n")
		formattedOutput.WriteString(actualOutput)
		actualOutput = formattedOutput.String()
	}

	// Use go-snaps for snapshot testing
	testName := filepath.Base(pasFile)
	if hasExpectedFile {
		// If we have an expected file, compare with it
		if normalizeOutput(actualOutput) != normalizeOutput(expectedContent) {
			t.Errorf("Output mismatch for %s:\nExpected:\n%s\nActual:\n%s",
				testName, expectedContent, actualOutput)
			return testResultFailed
		}
	} else {
		// Use go-snaps snapshot
		snaps.MatchSnapshot(t, fmt.Sprintf("%s_output", testName), actualOutput)
	}

	return testResultPassed
}
