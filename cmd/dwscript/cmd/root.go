package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information (set by build flags)
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var (
	// Global flags
	unitSearchPaths []string
	verbose         bool
)

var rootCmd = &cobra.Command{
	Use:   "dwscript",
	Short: "DWScript interpreter and compiler",
	Long: `go-dws is a Go implementation of the DWScript scripting language.

DWScript is a full-featured Object Pascal-based scripting language with:
  - Strong static typing with type inference
  - Object-oriented programming (classes, interfaces, inheritance)
  - Functions and procedures with nested scopes
  - Comprehensive built-in functions

This is a faithful port from the original Delphi implementation,
preserving 100% of DWScript's syntax and semantics.`,
	Version: Version,
}

// ErrSilent signals a failure whose diagnostics were already written by the command;
// main prints nothing further and exits non-zero.
var ErrSilent = errors.New("silent failure")

// Execute runs the root command. Errors are returned to main, which prints them
// exactly once (cobra's own error echo is silenced).
func Execute() error {
	rootCmd.SilenceErrors = true
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s " .}}{{end}}{{printf "version %%s" .Version}}
Commit: %s
Built:  %s
`, GitCommit, BuildDate))

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringSliceVarP(&unitSearchPaths, "include", "I", []string{}, "unit search paths (can be specified multiple times)")
}
