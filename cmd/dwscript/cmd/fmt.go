package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/pkg/printer"
	"github.com/spf13/cobra"
)

var (
	// Format command flags
	fmtWrite      bool   // -w: write result to (source) file instead of stdout
	fmtList       bool   // -l: list files whose formatting differs from gofmt's
	fmtDiff       bool   // -d: display diffs instead of rewriting files
	fmtStyle      string // --style: formatting style (detailed, compact, multiline)
	fmtIndent     int    // --indent: number of spaces per indentation level
	fmtUseTabs    bool   // --tabs: use tabs instead of spaces for indentation
	fmtSimplify   bool   // -s: simplify code (future enhancement)
	fmtRecursive  bool   // -r: process directories recursively
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [files or directories...]",
	Short: "Format DWScript source files",
	Long: `Format DWScript source files using the AST-driven formatter.

The formatter reads DWScript source code, parses it into an AST,
and then pretty-prints it back to source code with consistent formatting.

Usage:
  dwscript fmt file.dws              # Format to stdout
  dwscript fmt -w file.dws           # Overwrite file with formatted version
  dwscript fmt -l *.dws              # List files that need formatting
  dwscript fmt -d file.dws           # Show diff of changes
  dwscript fmt -r src/               # Format all .dws files in directory

By default, fmt formats the files named on the command line and writes
the result to standard output. If no path is provided, it reads from
standard input.

Flags:
  -w         write result to (source) file instead of stdout
  -l         list files whose formatting differs
  -d         display diffs instead of rewriting files
  -r         process directories recursively
  --style    formatting style: detailed (default), compact, or multiline
  --indent   number of spaces per indentation level (default: 2)
  --tabs     use tabs instead of spaces for indentation

Examples:
  # Format a single file to stdout
  dwscript fmt hello.dws

  # Format and overwrite files
  dwscript fmt -w file1.dws file2.dws

  # Format from stdin
  cat script.dws | dwscript fmt

  # List all files that need formatting
  dwscript fmt -l -r src/

  # Show what would change
  dwscript fmt -d script.dws

  # Use compact style
  dwscript fmt --style compact script.dws

  # Use tabs for indentation
  dwscript fmt --tabs -w script.dws`,
	RunE: runFmt,
}

func init() {
	rootCmd.AddCommand(fmtCmd)

	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "write result to (source) file instead of stdout")
	fmtCmd.Flags().BoolVarP(&fmtList, "list", "l", false, "list files whose formatting differs")
	fmtCmd.Flags().BoolVarP(&fmtDiff, "diff", "d", false, "display diffs instead of rewriting files")
	fmtCmd.Flags().BoolVarP(&fmtRecursive, "recursive", "r", false, "process directories recursively")
	fmtCmd.Flags().StringVar(&fmtStyle, "style", "detailed", "formatting style: detailed, compact, or multiline")
	fmtCmd.Flags().IntVar(&fmtIndent, "indent", 2, "number of spaces per indentation level")
	fmtCmd.Flags().BoolVar(&fmtUseTabs, "tabs", false, "use tabs instead of spaces for indentation")
	fmtCmd.Flags().BoolVarP(&fmtSimplify, "simplify", "s", false, "simplify code (future enhancement)")
}

func runFmt(cmd *cobra.Command, args []string) error {
	// Validate flags
	if fmtWrite && fmtList {
		return fmt.Errorf("cannot use -w and -l together")
	}
	if fmtWrite && fmtDiff {
		return fmt.Errorf("cannot use -w and -d together")
	}

	// Parse style option
	var style printer.Style
	switch strings.ToLower(fmtStyle) {
	case "detailed":
		style = printer.StyleDetailed
	case "compact":
		style = printer.StyleCompact
	case "multiline":
		style = printer.StyleMultiline
	default:
		return fmt.Errorf("unknown style: %s (use detailed, compact, or multiline)", fmtStyle)
	}

	// Build printer options
	opts := printer.Options{
		Format:      printer.FormatDWScript,
		Style:       style,
		IndentWidth: fmtIndent,
		UseSpaces:   !fmtUseTabs,
	}

	// If no files specified, read from stdin
	if len(args) == 0 {
		return formatStdin(opts)
	}

	// Process each file/directory
	hasErrors := false
	for _, path := range args {
		if err := processPath(path, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", path, err)
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("formatting failed for one or more files")
	}

	return nil
}

// processPath processes a file or directory
func processPath(path string, opts printer.Options) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if fmtRecursive {
			return processDirectory(path, opts)
		}
		return fmt.Errorf("%s is a directory (use -r to process recursively)", path)
	}

	return formatFile(path, opts)
}

// processDirectory recursively processes all .dws files in a directory
func processDirectory(dir string, opts printer.Options) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .dws files
		if !strings.HasSuffix(path, ".dws") {
			return nil
		}

		if err := formatFile(path, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", path, err)
			// Continue processing other files
		}

		return nil
	})
}

// formatStdin reads from stdin, formats it, and writes to stdout
func formatStdin(opts printer.Options) error {
	// Read all input
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading stdin: %w", err)
	}

	// Format the source
	formatted, err := formatSource(string(src), opts)
	if err != nil {
		return err
	}

	// Write to stdout
	fmt.Print(formatted)
	return nil
}

// formatFile formats a single file
func formatFile(filename string, opts printer.Options) error {
	// Read the file
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	original := string(src)

	// Format the source
	formatted, err := formatSource(original, opts)
	if err != nil {
		return err
	}

	// Check if formatting changed the file
	changed := original != formatted

	// Handle different output modes
	switch {
	case fmtList:
		// List mode: only print filename if formatting would change it
		if changed {
			fmt.Println(filename)
		}

	case fmtDiff:
		// Diff mode: show differences
		if changed {
			fmt.Printf("--- %s (original)\n", filename)
			fmt.Printf("+++ %s (formatted)\n", filename)
			showDiff(original, formatted)
		}

	case fmtWrite:
		// Write mode: overwrite the file if changed
		if changed {
			if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("error writing file: %w", err)
			}
			if verbose {
				fmt.Printf("Formatted %s\n", filename)
			}
		}

	default:
		// Default mode: write to stdout
		fmt.Print(formatted)
	}

	return nil
}

// formatSource parses and formats source code
func formatSource(source string, opts printer.Options) (string, error) {
	// Tokenize
	l := lexer.New(source)

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parse errors
	if len(p.Errors()) > 0 {
		var errMsg strings.Builder
		errMsg.WriteString("Parse errors:\n")
		for _, err := range p.Errors() {
			errMsg.WriteString(fmt.Sprintf("  %s\n", err))
		}
		return "", fmt.Errorf("%s", errMsg.String())
	}

	// Format using the printer
	pr := printer.New(opts)
	formatted := pr.Print(program)

	return formatted, nil
}

// showDiff shows a simple line-by-line diff
// TODO: Use a proper diff algorithm for better output
func showDiff(original, formatted string) {
	origLines := strings.Split(original, "\n")
	fmtLines := strings.Split(formatted, "\n")

	maxLines := len(origLines)
	if len(fmtLines) > maxLines {
		maxLines = len(fmtLines)
	}

	for i := 0; i < maxLines; i++ {
		var origLine, fmtLine string
		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(fmtLines) {
			fmtLine = fmtLines[i]
		}

		if origLine != fmtLine {
			if origLine != "" {
				fmt.Printf("- %s\n", origLine)
			}
			if fmtLine != "" {
				fmt.Printf("+ %s\n", fmtLine)
			}
		}
	}
}

// isFormattedCorrectly checks if a file is already correctly formatted
func isFormattedCorrectly(source string, opts printer.Options) (bool, error) {
	formatted, err := formatSource(source, opts)
	if err != nil {
		return false, err
	}
	return source == formatted, nil
}

// FormatBytes formats source code provided as bytes
// This is useful for integration with other tools
func FormatBytes(src []byte, opts printer.Options) ([]byte, error) {
	formatted, err := formatSource(string(src), opts)
	if err != nil {
		return nil, err
	}
	return []byte(formatted), nil
}

// FormatFile is a convenience function to format a file in-place
// Returns true if the file was modified
func FormatFile(filename string, opts printer.Options) (bool, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	formatted, err := FormatBytes(src, opts)
	if err != nil {
		return false, err
	}

	changed := !bytes.Equal(src, formatted)
	if changed {
		if err := os.WriteFile(filename, formatted, 0644); err != nil {
			return false, err
		}
	}

	return changed, nil
}
