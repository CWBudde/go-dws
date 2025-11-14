package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/bytecode"
	"github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/spf13/cobra"
)

var (
	outputFile      string
	skipTypeCheck   bool
	disassemble     bool
	compileVerbose  bool
)

var compileCmd = &cobra.Command{
	Use:   "compile [file]",
	Short: "Compile a DWScript file to bytecode",
	Long: `Compile a DWScript program to bytecode and save it as a .dwc file.

The compiled bytecode can be loaded and executed much faster than parsing
the source code each time. This is useful for production deployments or
frequently run scripts.

Examples:
  # Compile a script to bytecode
  dwscript compile script.dws

  # Compile with custom output file
  dwscript compile script.dws -o output.dwc

  # Compile and show disassembled bytecode
  dwscript compile script.dws --disassemble

  # Compile without type checking (faster but less safe)
  dwscript compile script.dws --skip-type-check`,
	Args: cobra.ExactArgs(1),
	RunE: compileScript,
}

func init() {
	rootCmd.AddCommand(compileCmd)

	compileCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (default: <input>.dwc)")
	compileCmd.Flags().BoolVar(&skipTypeCheck, "skip-type-check", false, "skip semantic type checking (faster but less safe)")
	compileCmd.Flags().BoolVar(&disassemble, "disassemble", false, "show disassembled bytecode after compilation")
	compileCmd.Flags().BoolVarP(&compileVerbose, "verbose", "v", false, "verbose output")
}

func compileScript(_ *cobra.Command, args []string) error {
	filename := args[0]

	// Read the source file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	input := string(content)

	if compileVerbose {
		fmt.Fprintf(os.Stderr, "Compiling %s...\n", filename)
	}

	// Lexer: tokenize the input
	l := lexer.New(input)

	// Parser: build the AST
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		compilerErrors := make([]*errors.CompilerError, 0, len(p.Errors()))
		for _, perr := range p.Errors() {
			compilerErrors = append(compilerErrors, errors.NewCompilerError(
				perr.Pos,
				perr.Message,
				input,
				filename,
			))
		}

		fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("parsing failed with %d error(s)", len(p.Errors()))
	}

	// Extract used units to determine if we need to skip type checking
	usedUnits := extractUsedUnits(program)
	hasUnits := len(usedUnits) > 0

	// Handle unit loading for bytecode compilation
	var compiledProgram *ast.Program
	if hasUnits {
		var searchPaths []string
		if len(unitSearchPaths) > 0 {
			searchPaths = unitSearchPaths
		} else {
			searchPaths = []string{filepath.Dir(filename)}
		}

		// Build bytecode program with units
		var unitRegistry interface{} // Not actually used for compilation
		compiledProgram, unitRegistry, err = buildBytecodeProgram(program, usedUnits, searchPaths)
		_ = unitRegistry // Avoid unused variable warning
		if err != nil {
			return fmt.Errorf("failed to prepare bytecode program: %w", err)
		}
	} else {
		compiledProgram = program
	}

	// Run semantic analysis if type checking is enabled and no units are used
	if !skipTypeCheck && !hasUnits {
		analyzer := semantic.NewAnalyzer()
		analyzer.SetSource(input, filename)

		if err := analyzer.Analyze(program); err != nil {
			p.SetSemanticErrors(analyzer.Errors())

			var compilerErrors []*errors.CompilerError
			if len(analyzer.StructuredErrors()) > 0 {
				for _, semErr := range analyzer.StructuredErrors() {
					compilerErrors = append(compilerErrors, semErr.ToCompilerError(input, filename))
				}
			} else {
				compilerErrors = errors.FromStringErrors(analyzer.Errors(), input, filename)
			}

			fmt.Fprint(os.Stderr, errors.FormatErrors(compilerErrors, true))
			fmt.Fprintln(os.Stderr)
			return fmt.Errorf("semantic analysis failed with %d error(s)", len(analyzer.Errors()))
		}
	} else if compileVerbose && hasUnits {
		fmt.Fprintf(os.Stderr, "Type checking disabled (program uses units)\n")
	}

	// Compile to bytecode
	compiler := bytecode.NewCompiler(filename)
	chunk, err := compiler.Compile(compiledProgram)
	if err != nil {
		return fmt.Errorf("bytecode compilation failed: %w", err)
	}

	if compileVerbose {
		fmt.Fprintf(os.Stderr, "Bytecode compilation successful\n")
		fmt.Fprintf(os.Stderr, "  Instructions: %d\n", len(chunk.Code))
		fmt.Fprintf(os.Stderr, "  Constants: %d\n", len(chunk.Constants))
		fmt.Fprintf(os.Stderr, "  Locals: %d\n", chunk.LocalCount)
	}

	// Disassemble if requested
	if disassemble {
		fmt.Fprintf(os.Stderr, "\n== Disassembled Bytecode (%s) ==\n", chunk.Name)
		bytecode.NewDisassembler(chunk, os.Stderr).Disassemble()
		fmt.Fprintln(os.Stderr)
	}

	// Serialize the bytecode
	serializer := bytecode.NewSerializer()
	data, err := serializer.SerializeChunk(chunk)
	if err != nil {
		return fmt.Errorf("failed to serialize bytecode: %w", err)
	}

	// Determine output filename
	outFile := outputFile
	if outFile == "" {
		// Default: replace extension with .dwc
		ext := filepath.Ext(filename)
		if ext != "" {
			outFile = strings.TrimSuffix(filename, ext) + ".dwc"
		} else {
			outFile = filename + ".dwc"
		}
	}

	// Write the bytecode file
	if err := os.WriteFile(outFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outFile, err)
	}

	if compileVerbose {
		fmt.Fprintf(os.Stderr, "Bytecode written to %s (%d bytes)\n", outFile, len(data))
	} else {
		fmt.Printf("Compiled %s -> %s\n", filename, outFile)
	}

	return nil
}
