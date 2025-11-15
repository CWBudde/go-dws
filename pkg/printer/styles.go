package printer

// This file provides pre-configured printer styles for common use cases.
// These helpers make it easy to create printers without manually constructing Options.

// CompactOptions returns options for compact, single-line output with minimal whitespace.
// Ideal for generating minified code or compact representations.
func CompactOptions() Options {
	return Options{
		Format:           FormatDWScript,
		Style:            StyleCompact,
		IndentWidth:      0,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// DetailedOptions returns options for detailed, well-formatted output with full indentation.
// This is the default style and is ideal for readable source code output.
func DetailedOptions() Options {
	return Options{
		Format:           FormatDWScript,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// MultilineOptions returns options for multiline output where every statement is on a new line.
// Similar to detailed but ensures strict line breaks between statements.
func MultilineOptions() Options {
	return Options{
		Format:           FormatDWScript,
		Style:            StyleMultiline,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// TreeOptions returns options for hierarchical AST tree visualization.
// Useful for debugging and understanding program structure.
func TreeOptions() Options {
	return Options{
		Format:           FormatTree,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// TreeOptionsWithPositions returns options for tree visualization including source positions.
// Includes line:column information for each AST node.
func TreeOptionsWithPositions() Options {
	opts := TreeOptions()
	opts.IncludePositions = true
	return opts
}

// TreeOptionsWithTypes returns options for tree visualization including type information.
// Includes type annotations for expressions and declarations.
func TreeOptionsWithTypes() Options {
	opts := TreeOptions()
	opts.IncludeTypes = true
	return opts
}

// TreeOptionsVerbose returns options for fully detailed tree visualization.
// Includes both positions and type information.
func TreeOptionsVerbose() Options {
	return Options{
		Format:           FormatTree,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: true,
		IncludeTypes:     true,
	}
}

// JSONOptions returns options for JSON representation of the AST.
// Useful for programmatic processing and integration with other tools.
func JSONOptions() Options {
	return Options{
		Format:           FormatJSON,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: false,
		IncludeTypes:     false,
	}
}

// JSONOptionsWithPositions returns options for JSON output including source positions.
func JSONOptionsWithPositions() Options {
	opts := JSONOptions()
	opts.IncludePositions = true
	return opts
}

// JSONOptionsWithTypes returns options for JSON output including type information.
func JSONOptionsWithTypes() Options {
	opts := JSONOptions()
	opts.IncludeTypes = true
	return opts
}

// JSONOptionsVerbose returns options for fully detailed JSON output.
// Includes both positions and type information.
func JSONOptionsVerbose() Options {
	return Options{
		Format:           FormatJSON,
		Style:            StyleDetailed,
		IndentWidth:      2,
		UseSpaces:        true,
		IncludePositions: true,
		IncludeTypes:     true,
	}
}

// CompactPrinter returns a new printer configured for compact output.
// Equivalent to New(CompactOptions()).
func CompactPrinter() *Printer {
	return New(CompactOptions())
}

// DetailedPrinter returns a new printer configured for detailed output.
// Equivalent to New(DetailedOptions()).
func DetailedPrinter() *Printer {
	return New(DetailedOptions())
}

// MultilinePrinter returns a new printer configured for multiline output.
// Equivalent to New(MultilineOptions()).
func MultilinePrinter() *Printer {
	return New(MultilineOptions())
}

// TreePrinter returns a new printer configured for tree visualization.
// Equivalent to New(TreeOptions()).
func TreePrinter() *Printer {
	return New(TreeOptions())
}

// JSONPrinter returns a new printer configured for JSON output.
// Equivalent to New(JSONOptions()).
func JSONPrinter() *Printer {
	return New(JSONOptions())
}
