package cmd

import (
	"fmt"
	"os"

	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/spf13/cobra"
)

var (
	showPos    bool
	showType   bool
	onlyErrors bool
)

var lexCmd = &cobra.Command{
	Use:   "lex [file]",
	Short: "Tokenize a DWScript file or expression",
	Long: `Tokenize (lex) a DWScript program and print the resulting tokens.

This command is useful for debugging the lexer and understanding how
DWScript source code is tokenized.

Examples:
  # Tokenize a script file
  dwscript lex script.dws

  # Tokenize an inline expression
  dwscript lex -e "var x: Integer := 42;"

  # Show token types and positions
  dwscript lex --show-type --show-pos script.dws

  # Show only errors (illegal tokens)
  dwscript lex --only-errors script.dws`,
	Args: cobra.MaximumNArgs(1),
	RunE: lexScript,
}

func init() {
	rootCmd.AddCommand(lexCmd)

	lexCmd.Flags().StringVarP(&evalExpr, "eval", "e", "", "tokenize inline code instead of reading from file")
	lexCmd.Flags().BoolVar(&showPos, "show-pos", false, "show token positions (line:column)")
	lexCmd.Flags().BoolVar(&showType, "show-type", false, "show token type names")
	lexCmd.Flags().BoolVar(&onlyErrors, "only-errors", false, "show only illegal/error tokens")
}

func lexScript(cmd *cobra.Command, args []string) error {
	var input string
	var filename string

	// Determine input source
	if evalExpr != "" {
		// Inline expression provided
		input = evalExpr
		filename = "<eval>"
	} else if len(args) == 1 {
		// File path provided
		filename = args[0]
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		input = string(content)
	} else {
		return fmt.Errorf("either provide a file path or use -e flag for inline code")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Tokenizing: %s\n", filename)
		fmt.Printf("Input length: %d bytes\n", len(input))
		fmt.Println("---")
	}

	// Create lexer
	l := lexer.New(input)

	// Tokenize and print
	tokenCount := 0
	errorCount := 0

	for {
		tok := l.NextToken()

		// Skip if only showing errors and this isn't an error
		if onlyErrors && tok.Type != lexer.ILLEGAL {
			if tok.Type == lexer.EOF {
				break
			}
			continue
		}

		// Count tokens
		tokenCount++
		if tok.Type == lexer.ILLEGAL {
			errorCount++
		}

		// Print token
		printToken(tok)

		// Stop at EOF
		if tok.Type == lexer.EOF {
			break
		}
	}

	if verbose {
		fmt.Println("---")
		fmt.Printf("Total tokens: %d\n", tokenCount)
		if errorCount > 0 {
			fmt.Printf("Errors: %d\n", errorCount)
		}
	}

	// Exit with error if there were illegal tokens
	if onlyErrors && errorCount > 0 {
		return fmt.Errorf("found %d illegal token(s)", errorCount)
	}

	return nil
}

func printToken(tok lexer.Token) {
	// Format: [TYPE] "literal" @line:col
	var output string

	if showType {
		output = fmt.Sprintf("[%-12s]", tok.Type)
	}

	// Add literal value
	if tok.Type == lexer.EOF {
		output += " EOF"
	} else if tok.Type == lexer.ILLEGAL {
		output += fmt.Sprintf(" ⚠️  ILLEGAL: %q", tok.Literal)
	} else if tok.Literal == "" {
		output += fmt.Sprintf(" %s", tok.Type)
	} else {
		output += fmt.Sprintf(" %q", tok.Literal)
	}

	// Add position if requested
	if showPos {
		output += fmt.Sprintf(" @%d:%d", tok.Pos.Line, tok.Pos.Column)
	}

	fmt.Println(output)
}
