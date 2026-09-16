// Package fixtureconfig shares upstream fixture scoring rules between the Go
// harness and the CLI compatibility report.
package fixtureconfig

import (
	"errors"
	"os"

	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// pedanticCategories are the fixture directories collected by upstream's
// UScriptTests (testdata/fixtures/UScriptTests.pas, lines 73-107), the only
// output-comparison runner that raises Config.HintsLevel to hlPedantic
// (line 111). Every other runner leaves it at the hlStrict default
// (dwsCompiler.pas TdwsConfiguration.Create), so pedantic-only diagnostics such
// as case-mismatch hints must stay silent there: FunctionsMath/lcm.pas spells
// the same builtin "Lcm" and "lcm" and expects no hint for either.
var pedanticCategories = map[string]bool{
	"SimpleScripts":           true,
	"ArrayPass":               true,
	"FailureScripts":          true,
	"AttributesFail":          true,
	"LambdaPass":              true,
	"LambdaFail":              true,
	"InterfacesPass":          true,
	"InterfacesFail":          true,
	"OperatorOverloadPass":    true,
	"OperatorOverloadFail":    true,
	"OverloadsPass":           true,
	"OverloadsFail":           true,
	"HelpersPass":             true,
	"HelpersFail":             true,
	"PropertyExpressionsPass": true,
	"PropertyExpressionsFail": true,
	"SetOfPass":               true,
	"SetOfFail":               true,
	"AssociativePass":         true,
	"AssociativeFail":         true,
	"GenericsPass":            true,
	"GenericsFail":            true,
	"InnerClassesPass":        true,
	"InnerClassesFail":        true,
}

// HintsLevel returns the compiler hint level used by a category's upstream runner.
func HintsLevel(category string) semantic.HintsLevel {
	if pedanticCategories[category] {
		return semantic.HintsLevelPedantic
	}
	return semantic.HintsLevelStrict
}

// ReadExpected reads a fixture's expectation with encoding detection. Output
// comparison suites treat a missing file as an empty expectation. Other runners,
// unsupported host setup, and failure suites without exact expectations remain
// unscored. An existing file is always used, and read/decode errors other than
// absence are returned.
func ReadExpected(category, path string) (content string, scored bool, err error) {
	content, err = encoding.DecodeFile(path)
	if !errors.Is(err, os.ErrNotExist) {
		return content, true, err
	}
	switch category {
	case "BuildScripts", "AutoFormat", "External", "DelegateLib", "FailureScripts":
		return "", false, nil
	default:
		return "", true, nil
	}
}
