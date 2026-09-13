// Package fixtureconfig shares upstream fixture scoring rules between the Go
// harness and the CLI compatibility report.
package fixtureconfig

import (
	"errors"
	"os"

	"github.com/cwbudde/go-dws/internal/encoding"
	"github.com/cwbudde/go-dws/internal/semantic"
)

// HintsLevel returns the compiler hint level used by a category's upstream runner.
func HintsLevel(category string) semantic.HintsLevel {
	switch category {
	case "Algorithms", "FunctionsString", "Memory":
		return semantic.HintsLevelNormal
	default:
		return semantic.HintsLevelPedantic
	}
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
