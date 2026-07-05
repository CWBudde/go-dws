// Package lexer provides lexical analysis for DWScript source code.
// This file implements source inclusion directives:
//
//	{$INCLUDE 'file'}       / {$I 'file'}   — splice the file's contents inline
//	{$INCLUDE_ONCE 'file'}                  — splice only the first time seen
//
// Inclusion is handled entirely at the lexer level: when an include directive is
// encountered, the current scan state is pushed onto an include stack and the
// referenced file becomes the active input. When the included input is exhausted,
// the parent file resumes exactly where it left off. This mirrors DWScript, where
// includes are a pure textual substitution performed before parsing.
package lexer

import (
	"path/filepath"
	"strings"

	"github.com/cwbudde/go-dws/internal/encoding"
)

// maxIncludeDepth bounds the include stack to guard against runaway recursion via
// plain {$INCLUDE} (which, unlike {$INCLUDE_ONCE}, has no self-reference guard).
const maxIncludeDepth = 1024

// IncludeResolver loads the source referenced by an include directive.
//
// name is the (unquoted) file name exactly as written in the directive. fromPath
// is the canonical path of the file that contains the directive (empty for the
// top-level source), so relative includes can be resolved against the including
// file's own directory — matching DWScript, where nested includes resolve relative
// to the file they appear in rather than the entry point.
//
// The resolver returns the decoded file contents together with a canonical path
// used both to de-duplicate {$INCLUDE_ONCE} references and to resolve any further
// relative includes nested inside the returned content. A non-nil error reports
// that the file could not be resolved or read.
type IncludeResolver func(name, fromPath string) (content string, canonical string, err error)

// includeFrame captures the lexer's scan state at the point an included file was
// entered, so the parent file can be resumed once the include is exhausted.
type includeFrame struct {
	input        string
	path         string
	ch           rune
	position     int
	readPosition int
	line         int
	column       int
}

// NewFileIncludeResolver returns an IncludeResolver that reads include files from
// the filesystem. Relative names are resolved against the directory of the file
// containing the directive (fromPath); at the top level, where no including file
// exists yet, they resolve against baseDir. Files are decoded with the same
// BOM/UTF-16 handling as top-level sources so included content matches DWScript's
// file reading.
func NewFileIncludeResolver(baseDir string) IncludeResolver {
	return func(name, fromPath string) (string, string, error) {
		path := name
		if !filepath.IsAbs(path) {
			dir := baseDir
			if fromPath != "" {
				dir = filepath.Dir(fromPath)
			}
			path = filepath.Join(dir, name)
		}
		content, err := encoding.DecodeFile(path)
		if err != nil {
			return "", "", err
		}
		canonical, absErr := filepath.Abs(path)
		if absErr != nil {
			canonical = path
		}
		return content, canonical, nil
	}
}

// handleInclude processes an {$INCLUDE}/{$I}/{$INCLUDE_ONCE} directive. content is
// the full directive body (e.g. `include 'foo.inc'`); name is the lower-cased
// directive keyword.
func (l *Lexer) handleInclude(name, content string, parentActive bool, startPos Position) {
	if !parentActive {
		return
	}

	filename := extractIncludeArg(content)
	if filename == "" {
		l.addIncludeError("file name expected after $"+name, startPos)
		return
	}

	// {$I %FILE%}, {$I %LINE%}, ... are compile-time value substitutions, not file
	// inclusions. They are not supported yet; leave them untouched rather than
	// treating the token as a file path.
	if strings.HasPrefix(filename, "%") {
		return
	}

	if l.includeResolver == nil {
		return
	}

	includeContent, canonical, err := l.includeResolver(filename, l.currentIncludePath)
	if err != nil {
		l.addIncludeError("cannot open include file '"+filename+"': "+err.Error(), startPos)
		return
	}

	if name == "include_once" {
		if l.includedOnce == nil {
			l.includedOnce = make(map[string]struct{})
		}
		if _, seen := l.includedOnce[canonical]; seen {
			return
		}
		l.includedOnce[canonical] = struct{}{}
	}

	if len(l.includeStack) >= maxIncludeDepth {
		l.addIncludeError("include nesting too deep (possible cyclic {$INCLUDE})", startPos)
		return
	}

	l.enterInclude(includeContent, canonical)
}

// enterInclude pushes the current scan state and switches the lexer to lex the
// included content. The saved state points at the first character following the
// directive's closing '}', so the parent resumes there once the include ends.
// canonical is the resolved path of the included file, tracked so that relative
// includes nested inside it resolve against its own directory.
func (l *Lexer) enterInclude(content, canonical string) {
	l.includeStack = append(l.includeStack, includeFrame{
		input:        l.input,
		path:         l.currentIncludePath,
		position:     l.position,
		readPosition: l.readPosition,
		line:         l.line,
		column:       l.column,
		ch:           l.ch,
	})
	l.includeCount++

	l.input = content
	l.currentIncludePath = canonical
	l.position = 0
	l.readPosition = 0
	l.line = 1
	l.column = 0
	l.readChar()
}

// popInclude restores the scan state saved when the current included file was
// entered, resuming the parent file.
func (l *Lexer) popInclude() {
	frame := l.includeStack[len(l.includeStack)-1]
	l.includeStack = l.includeStack[:len(l.includeStack)-1]
	l.input = frame.input
	l.currentIncludePath = frame.path
	l.position = frame.position
	l.readPosition = frame.readPosition
	l.line = frame.line
	l.column = frame.column
	l.ch = frame.ch
}

// extractIncludeArg returns the file-name argument of an include directive body,
// stripped of surrounding quotes. content is the directive text with the leading
// keyword still present (e.g. `include 'foo.inc'` or `i "bar.inc"`).
func extractIncludeArg(content string) string {
	trimmed := strings.TrimSpace(content)
	// Drop the leading directive keyword.
	if idx := strings.IndexFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	}); idx >= 0 {
		trimmed = strings.TrimSpace(trimmed[idx+1:])
	} else {
		return ""
	}
	return unquoteIncludeArg(trimmed)
}

// unquoteIncludeArg removes a matching pair of surrounding single or double quotes.
func unquoteIncludeArg(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
