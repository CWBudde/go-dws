// Package printer provides pretty-printing and formatting functionality for DWScript AST nodes.
//
// This package separates presentation logic from the AST structure, allowing multiple
// output formats and styles without modifying the AST node definitions.
//
// # Overview
//
// The printer package converts AST nodes into human-readable text representations.
// It supports multiple output formats and styles to suit different use cases:
//
//   - DWScript syntax: Valid DWScript source code (default)
//   - Tree format: Hierarchical AST structure visualization
//   - JSON format: Machine-readable JSON representation
//   - Compact format: Minimal whitespace, single-line output
//
// # Basic Usage
//
// The simplest way to print an AST node is using the Print function:
//
//	program := parser.ParseProgram()
//	output := printer.Print(program)
//	fmt.Println(output)
//
// # Custom Formatting
//
// For more control, create a Printer with custom options:
//
//	p := printer.New(printer.Options{
//	    Format:      printer.FormatDWScript,
//	    Style:       printer.StyleDetailed,
//	    IndentWidth: 2,
//	})
//	output := p.Print(program)
//
// # Output Formats
//
// FormatDWScript produces valid DWScript source code:
//
//	type TMyClass = class
//	  FValue: Integer;
//	  procedure DoSomething;
//	end;
//
// FormatTree produces a hierarchical structure view:
//
//	Program
//	  ClassDecl (TMyClass)
//	    FieldDecl (FValue: Integer)
//	    FunctionDecl (DoSomething)
//
// FormatJSON produces machine-readable JSON:
//
//	{
//	  "type": "Program",
//	  "statements": [
//	    {
//	      "type": "ClassDecl",
//	      "name": "TMyClass",
//	      ...
//	    }
//	  ]
//	}
//
// # Styles
//
// StyleCompact minimizes whitespace:
//
//	type TMyClass=class FValue:Integer;procedure DoSomething;end;
//
// StyleDetailed adds full indentation and spacing:
//
//	type TMyClass = class
//	  FValue: Integer;
//	  procedure DoSomething;
//	end;
//
// StyleMultiline ensures every statement is on a new line with proper indentation.
//
// # Design Rationale
//
// Prior to this package, AST nodes contained extensive String() methods (some 50+ lines)
// that mixed structural concerns with presentation logic. This made the AST harder to
// maintain and limited output to a single hardcoded format.
//
// By extracting formatting logic into a dedicated printer package:
//
//   - AST nodes stay focused on structure (simpler, smaller code)
//   - Multiple output formats are possible without touching AST
//   - Formatting changes don't require AST modifications
//   - Better separation of concerns (structure vs. presentation)
//   - Easier to add new formats (e.g., GraphViz, XML)
//
// AST nodes retain minimal String() methods for debugging (e.g., "ClassDecl(TMyClass)"),
// while the printer handles all production-quality formatting.
package printer
