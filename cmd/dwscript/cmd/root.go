package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set by build flags)
	Version   = "0.1.0-dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
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

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s " .}}{{end}}{{printf "version %%s" .Version}}
Commit: %s
Built:  %s
`, GitCommit, BuildDate))

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func exitWithError(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	os.Exit(1)
}
