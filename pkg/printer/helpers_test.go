package printer_test

import (
	"testing"

	"github.com/cwbudde/go-dws/pkg/printer"
)

// TestFormatString tests Format.String()
func TestFormatString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		format   printer.Format
	}{
		{"FormatDWScript", printer.FormatDWScript, "dwscript"},
		{"FormatTree", printer.FormatTree, "tree"},
		{"FormatJSON", printer.FormatJSON, "json"},
		{"unknown format", printer.Format(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.expected {
				t.Errorf("Format.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestStyleString tests Style.String()
func TestStyleString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		style    printer.Style
	}{
		{"StyleDetailed", printer.StyleDetailed, "detailed"},
		{"StyleCompact", printer.StyleCompact, "compact"},
		{"StyleMultiline", printer.StyleMultiline, "multiline"},
		{"unknown style", printer.Style(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.String()
			if got != tt.expected {
				t.Errorf("Style.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestMultilineOptions tests MultilineOptions()
func TestMultilineOptions(t *testing.T) {
	opts := printer.MultilineOptions()

	if opts.Format != printer.FormatDWScript {
		t.Errorf("MultilineOptions().Format = %v, want FormatDWScript", opts.Format)
	}
	if opts.Style != printer.StyleMultiline {
		t.Errorf("MultilineOptions().Style = %v, want StyleMultiline", opts.Style)
	}
	if opts.IndentWidth != 2 {
		t.Errorf("MultilineOptions().IndentWidth = %d, want 2", opts.IndentWidth)
	}
	if !opts.UseSpaces {
		t.Error("MultilineOptions().UseSpaces = false, want true")
	}
	if opts.IncludePositions {
		t.Error("MultilineOptions().IncludePositions = true, want false")
	}
	if opts.IncludeTypes {
		t.Error("MultilineOptions().IncludeTypes = true, want false")
	}
}

// TestTreeOptionsWithPositions tests TreeOptionsWithPositions()
func TestTreeOptionsWithPositions(t *testing.T) {
	opts := printer.TreeOptionsWithPositions()

	if opts.Format != printer.FormatTree {
		t.Errorf("TreeOptionsWithPositions().Format = %v, want FormatTree", opts.Format)
	}
	if !opts.IncludePositions {
		t.Error("TreeOptionsWithPositions().IncludePositions = false, want true")
	}
	if opts.IncludeTypes {
		t.Error("TreeOptionsWithPositions().IncludeTypes = true, want false")
	}
}

// TestTreeOptionsWithTypes tests TreeOptionsWithTypes()
func TestTreeOptionsWithTypes(t *testing.T) {
	opts := printer.TreeOptionsWithTypes()

	if opts.Format != printer.FormatTree {
		t.Errorf("TreeOptionsWithTypes().Format = %v, want FormatTree", opts.Format)
	}
	if opts.IncludePositions {
		t.Error("TreeOptionsWithTypes().IncludePositions = true, want false")
	}
	if !opts.IncludeTypes {
		t.Error("TreeOptionsWithTypes().IncludeTypes = false, want true")
	}
}

// TestTreeOptionsVerbose tests TreeOptionsVerbose()
func TestTreeOptionsVerbose(t *testing.T) {
	opts := printer.TreeOptionsVerbose()

	if opts.Format != printer.FormatTree {
		t.Errorf("TreeOptionsVerbose().Format = %v, want FormatTree", opts.Format)
	}
	if !opts.IncludePositions {
		t.Error("TreeOptionsVerbose().IncludePositions = false, want true")
	}
	if !opts.IncludeTypes {
		t.Error("TreeOptionsVerbose().IncludeTypes = false, want true")
	}
}

// TestJSONOptionsWithPositions tests JSONOptionsWithPositions()
func TestJSONOptionsWithPositions(t *testing.T) {
	opts := printer.JSONOptionsWithPositions()

	if opts.Format != printer.FormatJSON {
		t.Errorf("JSONOptionsWithPositions().Format = %v, want FormatJSON", opts.Format)
	}
	if !opts.IncludePositions {
		t.Error("JSONOptionsWithPositions().IncludePositions = false, want true")
	}
	if opts.IncludeTypes {
		t.Error("JSONOptionsWithPositions().IncludeTypes = true, want false")
	}
}

// TestJSONOptionsWithTypes tests JSONOptionsWithTypes()
func TestJSONOptionsWithTypes(t *testing.T) {
	opts := printer.JSONOptionsWithTypes()

	if opts.Format != printer.FormatJSON {
		t.Errorf("JSONOptionsWithTypes().Format = %v, want FormatJSON", opts.Format)
	}
	if opts.IncludePositions {
		t.Error("JSONOptionsWithTypes().IncludePositions = true, want false")
	}
	if !opts.IncludeTypes {
		t.Error("JSONOptionsWithTypes().IncludeTypes = false, want true")
	}
}

// TestJSONOptionsVerbose tests JSONOptionsVerbose()
func TestJSONOptionsVerbose(t *testing.T) {
	opts := printer.JSONOptionsVerbose()

	if opts.Format != printer.FormatJSON {
		t.Errorf("JSONOptionsVerbose().Format = %v, want FormatJSON", opts.Format)
	}
	if !opts.IncludePositions {
		t.Error("JSONOptionsVerbose().IncludePositions = false, want true")
	}
	if !opts.IncludeTypes {
		t.Error("JSONOptionsVerbose().IncludeTypes = false, want true")
	}
}

// TestMultilinePrinter tests MultilinePrinter()
func TestMultilinePrinter(t *testing.T) {
	p := printer.MultilinePrinter()
	if p == nil {
		t.Fatal("MultilinePrinter() returned nil")
	}
	// Just verify it creates a printer without panic
}

// TestPrinterHelpers tests various printer helper functions by using them
func TestPrinterHelpers(t *testing.T) {
	// Create printers to ensure they work
	printers := []struct {
		p    *printer.Printer
		name string
	}{
		{"CompactPrinter", printer.CompactPrinter()},
		{"DetailedPrinter", printer.DetailedPrinter()},
		{"MultilinePrinter", printer.MultilinePrinter()},
		{"TreePrinter", printer.TreePrinter()},
		{"JSONPrinter", printer.JSONPrinter()},
	}

	for _, test := range printers {
		t.Run(test.name, func(t *testing.T) {
			if test.p == nil {
				t.Fatalf("%s returned nil", test.name)
			}
		})
	}
}

// TestOptionsHelpers tests various option helper functions
func TestOptionsHelpers(t *testing.T) {
	options := []struct {
		name string
		opts printer.Options
	}{
		{"CompactOptions", printer.CompactOptions()},
		{"DetailedOptions", printer.DetailedOptions()},
		{"MultilineOptions", printer.MultilineOptions()},
		{"TreeOptions", printer.TreeOptions()},
		{"TreeOptionsWithPositions", printer.TreeOptionsWithPositions()},
		{"TreeOptionsWithTypes", printer.TreeOptionsWithTypes()},
		{"TreeOptionsVerbose", printer.TreeOptionsVerbose()},
		{"JSONOptions", printer.JSONOptions()},
		{"JSONOptionsWithPositions", printer.JSONOptionsWithPositions()},
		{"JSONOptionsWithTypes", printer.JSONOptionsWithTypes()},
		{"JSONOptionsVerbose", printer.JSONOptionsVerbose()},
	}

	for _, test := range options {
		t.Run(test.name, func(t *testing.T) {
			// Verify that IndentWidth is set (basic sanity check)
			if test.opts.IndentWidth < 0 {
				t.Errorf("%s has negative IndentWidth: %d", test.name, test.opts.IndentWidth)
			}
		})
	}
}
