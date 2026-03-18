package frontend

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	dwserrors "github.com/cwbudde/go-dws/internal/errors"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/parser"
	"github.com/cwbudde/go-dws/internal/semantic"
	"github.com/cwbudde/go-dws/pkg/ast"
)

// Phase identifies the compilation phase that produced a diagnostic.
type Phase string

const (
	PhaseParsing  Phase = "parsing"
	PhaseSemantic Phase = "semantic"
)

// Severity is the normalized severity level for front-end diagnostics.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
	SeverityHint
)

// Diagnostic is a normalized compile-front-end diagnostic emitted by parsing or semantic analysis.
type Diagnostic struct {
	Message  string
	Rendered string
	Code     string
	Phase    Phase
	Line     int
	Column   int
	Length   int
	Severity Severity
	Fatal    bool
	// BlocksSemantic marks parser diagnostics that should stop semantic analysis
	// because the recovered AST/result is not trustworthy enough to continue.
	BlocksSemantic bool
}

// Render returns the centralized rendered form of the diagnostic.
func (d Diagnostic) Render() string {
	if d.Rendered != "" {
		return d.Rendered
	}
	if d.Phase == PhaseParsing && d.Line > 0 && d.Column > 0 {
		return dwserrors.FormatDWScriptError(d.Message, d.Line, d.Column)
	}
	if d.Line > 0 && d.Column > 0 {
		return fmt.Sprintf("%s at %d:%d", d.Message, d.Line, d.Column)
	}
	return d.Message
}

// String returns the rendered form of the diagnostic.
func (d Diagnostic) String() string {
	return d.Render()
}

// Result is the shared front-end compile result for parser and semantic diagnostics.
type Result struct {
	Program            *ast.Program
	Analyzer           *semantic.Analyzer
	SemanticInfo       *ast.SemanticInfo
	Diagnostics        []Diagnostic
	SemanticAttempted  bool
	SemanticSuccessful bool
}

// HasFatalDiagnostics reports whether compilation produced fatal front-end diagnostics.
func (r *Result) HasFatalDiagnostics() bool {
	for _, diag := range r.Diagnostics {
		if diag.Fatal {
			return true
		}
	}
	return false
}

// HasFatalDiagnosticsInPhase reports whether compilation produced fatal diagnostics in a specific phase.
func (r *Result) HasFatalDiagnosticsInPhase(phase Phase) bool {
	for _, diag := range r.Diagnostics {
		if diag.Phase == phase && diag.Fatal {
			return true
		}
	}
	return false
}

// HasSemanticBlockingDiagnostics reports whether compilation produced diagnostics
// that should prevent entering semantic analysis.
func (r *Result) HasSemanticBlockingDiagnostics() bool {
	for _, diag := range r.Diagnostics {
		if diag.BlocksSemantic {
			return true
		}
	}
	return false
}

// HasSemanticBlockingDiagnosticsInPhase reports whether a phase produced diagnostics
// that should prevent entering semantic analysis.
func (r *Result) HasSemanticBlockingDiagnosticsInPhase(phase Phase) bool {
	for _, diag := range r.Diagnostics {
		if diag.Phase == phase && diag.BlocksSemantic {
			return true
		}
	}
	return false
}

// DiagnosticStrings returns the rendered diagnostics in emission order.
func (r *Result) DiagnosticStrings() []string {
	out := make([]string, 0, len(r.Diagnostics))
	for _, diag := range r.Diagnostics {
		out = append(out, diag.Render())
	}
	return out
}

// Parse parses source and collects parser diagnostics without running semantic analysis.
func Parse(source string) *Result {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	return &Result{
		Program:     program,
		Diagnostics: filterDiagnostics(parserDiagnostics(p.Errors())),
	}
}

// Compile parses source and, if parsing succeeds, runs semantic analysis.
// This is the shared compile-front-end boundary for diagnostics collection.
func Compile(source, filename string, hintsLevel semantic.HintsLevel) *Result {
	result := Parse(source)
	return compileParsedResult(result, source, filename, hintsLevel)
}

func compileParsedResult(result *Result, source, filename string, hintsLevel semantic.HintsLevel) *Result {
	if result.Program == nil || result.HasSemanticBlockingDiagnosticsInPhase(PhaseParsing) {
		return result
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.SetHintsLevel(hintsLevel)
	analyzer.SetSource(source, filename)
	result.Analyzer = analyzer
	result.SemanticAttempted = true

	err := safeAnalyze(analyzer, result)
	result.SemanticInfo = analyzer.GetSemanticInfo()
	result.Diagnostics = append(result.Diagnostics, semanticDiagnostics(analyzer)...)
	sortDiagnostics(result.Diagnostics)
	result.Diagnostics = filterDiagnostics(result.Diagnostics)
	sortDiagnostics(result.Diagnostics)
	result.SemanticSuccessful = err == nil

	return result
}

func safeAnalyze(analyzer *semantic.Analyzer, result *Result) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("semantic analysis panic: %v", recovered)
			if result != nil && !result.HasFatalDiagnostics() {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Message:  "internal semantic analysis panic",
					Rendered: fmt.Sprintf("internal semantic analysis panic: %v\n%s", recovered, strings.TrimSpace(string(debug.Stack()))),
					Code:     "E_SEMANTIC_PANIC",
					Phase:    PhaseSemantic,
					Severity: SeverityError,
					Fatal:    true,
				})
			}
		}
	}()

	return analyzer.Analyze(result.Program)
}

func sortDiagnostics(diags []Diagnostic) {
	sort.SliceStable(diags, func(i, j int) bool {
		left := diags[i]
		right := diags[j]

		if left.Severity != SeverityError && right.Severity != SeverityError {
			return false
		}
		if left.Severity != SeverityError && right.Severity == SeverityError {
			if left.Line == right.Line {
				return true
			}
			return false
		}
		if left.Severity == SeverityError && right.Severity != SeverityError {
			if left.Line == right.Line {
				return false
			}
			return false
		}

		leftBucket := diagnosticDeferredBucket(left)
		rightBucket := diagnosticDeferredBucket(right)
		if leftBucket != rightBucket {
			return leftBucket < rightBucket
		}

		if left.Line == 0 && right.Line != 0 {
			return false
		}
		if left.Line != 0 && right.Line == 0 {
			return true
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Phase == right.Phase && left.Severity != right.Severity {
			leftNonError := left.Severity != SeverityError
			rightNonError := right.Severity != SeverityError
			if leftNonError != rightNonError {
				return leftNonError
			}
		}
		leftCount := diagnosticArgumentCountPriority(left)
		rightCount := diagnosticArgumentCountPriority(right)
		if leftCount != rightCount {
			return leftCount < rightCount
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		if strings.Contains(left.Message, `Unknown name "`) && strings.Contains(right.Message, `";" expected`) {
			return true
		}
		if strings.Contains(right.Message, `Unknown name "`) && strings.Contains(left.Message, `";" expected`) {
			return false
		}
		leftPhase := diagnosticPhasePriority(left)
		rightPhase := diagnosticPhasePriority(right)
		if leftPhase != rightPhase {
			return leftPhase < rightPhase
		}
		leftSpecificity := diagnosticSpecificityPriority(left)
		rightSpecificity := diagnosticSpecificityPriority(right)
		if leftSpecificity != rightSpecificity {
			return leftSpecificity < rightSpecificity
		}
		return false
	})
}

func diagnosticArgumentCountPriority(diag Diagnostic) int {
	message := strings.TrimPrefix(diag.Message, "Syntax Error: ")
	switch message {
	case "More arguments expected",
		"Too many arguments",
		"No arguments expected":
		return 0
	default:
		return 1
	}
}

func diagnosticDeferredBucket(diag Diagnostic) int {
	if diag.Phase == PhaseSemantic &&
		((strings.HasPrefix(diag.Message, `Method "`) &&
			strings.Contains(diag.Message, `" not implemented`)) ||
			(strings.HasPrefix(diag.Message, `Class "`) &&
				strings.Contains(diag.Message, ` isn't defined completely`))) {
		return 1
	}
	return 0
}

func diagnosticPhasePriority(diag Diagnostic) int {
	if diag.Phase == PhaseParsing {
		return 0
	}
	return 1
}

func diagnosticSpecificityPriority(diag Diagnostic) int {
	switch {
	case diag.Message == "Syntax Error: Object reference needed to read/write an object field":
		return 0
	case diagnosticVisibleMember.MatchString(diag.Message):
		return 0
	case diag.Message == "Syntax Error: Class method or constructor expected":
		return 1
	case diagnosticAccessibleMember.MatchString(diag.Message):
		return 1
	default:
		return 0
	}
}

func parserDiagnostics(errors []*parser.ParserError) []Diagnostic {
	diags := make([]Diagnostic, 0, len(errors))
	for _, err := range errors {
		if err == nil {
			continue
		}
		message := normalizeParserDiagnosticMessage(err.Message)
		diags = append(diags, Diagnostic{
			Message:        message,
			Code:           err.Code,
			Phase:          PhaseParsing,
			Line:           err.Pos.Line,
			Column:         err.Pos.Column,
			Length:         err.Length,
			Severity:       SeverityError,
			Fatal:          true,
			BlocksSemantic: parserDiagnosticBlocksSemantic(err),
		})
	}
	return diags
}

func parserDiagnosticBlocksSemantic(err *parser.ParserError) bool {
	if err == nil {
		return false
	}

	switch err.Code {
	case parser.ErrUnexpectedToken,
		parser.ErrMissingSemicolon,
		parser.ErrMissingLParen,
		parser.ErrMissingRParen,
		parser.ErrMissingRBracket,
		parser.ErrMissingRBrace,
		parser.ErrInvalidExpression,
		parser.ErrNoPrefixParse,
		parser.ErrExpectedIdent,
		parser.ErrExpectedType,
		parser.ErrExpectedOperator,
		parser.ErrInvalidSyntax,
		parser.ErrMissingThen,
		parser.ErrMissingDo,
		parser.ErrMissingOf,
		parser.ErrMissingTo,
		parser.ErrMissingIn,
		parser.ErrMissingColon,
		parser.ErrMissingAssign,
		parser.ErrInvalidType:
		return false
	case parser.ErrMissingEnd:
		if err != nil && strings.Contains(err.Message, "class declaration") {
			return false
		}
		return true
	default:
		// Unknown parser error codes are treated conservatively until their
		// recovery semantics are classified explicitly.
		return true
	}
}

var parserBlockContextSuffix = regexp.MustCompile(` \(in .* block starting at line \d+\)$`)
var semanticVisibilityError = regexp.MustCompile(`^cannot (?:access|call) (?:private|protected) (?:field|method|property|class variable) '([^']+)' of class '([^']+)'$`)
var semanticImplicitVisibilityError = regexp.MustCompile(`^cannot access (?:private|protected) field '([^']+)'$`)
var diagnosticVisibleMember = regexp.MustCompile(`^Member symbol "([^"]+)" is not visible from this scope$`)
var diagnosticAccessibleMember = regexp.MustCompile(`^There is no accessible member with name "([^"]+)" for type `)

func normalizeParserDiagnosticMessage(message string) string {
	message = parserBlockContextSuffix.ReplaceAllString(message, "")

	if strings.HasPrefix(message, "function call must use identifier or member access") {
		return "Not a method"
	}

	switch message {
	case "expected 'do' after while condition":
		return "DO expected"
	case "expected 'end' to close block":
		return "End of block expected"
	case "expected 'do' after exception type":
		return "DO expected"
	case "expected ':' after exception variable":
		return `Colon ":" expected`
	case "expected identifier after 'on'":
		return "Name expected"
	case "expected ']' to close array index":
		return `"]" expected`
	case "expected ';' after function signature":
		return `";" expected`
	case "expected identifier in var declaration":
		return "Name expected"
	case "expected identifier after 'type'":
		return "Name expected"
	case "expected '=' after type name":
		return `"=" expected`
	case "expected ';' after type declaration":
		return `";" expected`
	case "expected ';' after variable declaration":
		return `";" expected`
	}

	return message
}

func semanticDiagnostics(analyzer *semantic.Analyzer) []Diagnostic {
	if analyzer == nil {
		return nil
	}

	structuredErrors := analyzer.StructuredErrors()
	legacyErrors := analyzer.Errors()
	diags := make([]Diagnostic, 0, len(structuredErrors)+len(legacyErrors))
	seen := make(map[string]struct{})
	structuredByMessage := make(map[string][]*semantic.SemanticError, len(structuredErrors))
	for _, err := range structuredErrors {
		if err == nil {
			continue
		}
		key := err.Error()
		structuredByMessage[key] = append(structuredByMessage[key], err)
	}

	for _, errStr := range legacyErrors {
		if candidates := structuredByMessage[errStr]; len(candidates) > 0 {
			err := candidates[0]
			structuredByMessage[errStr] = candidates[1:]
			message, line, column, rendered := normalizeSemanticDiagnostic(err.Error(), err.Message, err.Pos.Line, err.Pos.Column, severityFromSemantic(err.Severity))
			diag := Diagnostic{
				Message:  message,
				Rendered: rendered,
				Code:     string(err.Type),
				Phase:    PhaseSemantic,
				Line:     line,
				Column:   column,
				Severity: severityFromSemantic(err.Severity),
				Fatal:    err.Severity == semantic.SeverityError,
			}
			if _, ok := seen[diag.Render()]; ok {
				continue
			}
			seen[diag.Render()] = struct{}{}
			diags = append(diags, diag)
			continue
		}

		severity, fatal := inferStringSeverity(errStr)
		line, column, message := extractPosition(errStr)
		code := semanticCodeForSeverity(severity)
		message, line, column, rendered := normalizeSemanticDiagnostic(errStr, message, line, column, severity)
		diag := Diagnostic{
			Message:  message,
			Rendered: rendered,
			Code:     code,
			Phase:    PhaseSemantic,
			Line:     line,
			Column:   column,
			Severity: severity,
			Fatal:    fatal,
		}
		if _, ok := seen[diag.Render()]; ok {
			continue
		}
		seen[diag.Render()] = struct{}{}
		diags = append(diags, diag)
	}

	return diags
}

func filterDiagnostics(diags []Diagnostic) []Diagnostic {
	if len(diags) == 0 {
		return diags
	}

	filtered := make([]Diagnostic, 0, len(diags))
	hasEarlierFatal := false
	nameExpectedByLine := make(map[int]bool)
	colonExpectedByLine := make(map[int]bool)
	dotExpectedByLine := make(map[int]bool)
	seenNameExpectedColumnByLine := make(map[int]int)
	hasUnknownName := false

	for _, diag := range diags {
		if strings.Contains(diag.Message, "Name expected") {
			if prevCol, ok := seenNameExpectedColumnByLine[diag.Line]; ok && diag.Column > 0 && prevCol > 0 && diag.Column-prevCol <= 8 {
				continue
			}
			seenNameExpectedColumnByLine[diag.Line] = diag.Column
			nameExpectedByLine[diag.Line] = true
		}
		if strings.Contains(diag.Message, `Colon ":" expected`) {
			colonExpectedByLine[diag.Line] = true
		}
		if strings.Contains(diag.Message, `Dot "." expected`) {
			dotExpectedByLine[diag.Line] = true
		}
		if strings.Contains(diag.Message, "Unknown name \"") {
			hasUnknownName = true
		}

		if typeName := unknownTypeName(diag.Message); typeName != "" {
			for _, existing := range filtered {
				if existing.Line == diag.Line && strings.Contains(existing.Message, `";" expected`) {
					diag.Message = `Unknown name "` + typeName + `"`
					diag.Rendered = dwserrors.FormatDWScriptError(diag.Message, existing.Line, existing.Column)
					diag.Line = existing.Line
					diag.Column = existing.Column
					break
				}
			}
		}

		if diag.Phase == PhaseParsing && diag.Message == "expected 'end' to close class declaration" && hasUnknownName {
			continue
		}
		if diag.Phase == PhaseParsing && diag.Message == "Expression expected" && len(filtered) > 0 {
			prev := filtered[len(filtered)-1]
			if strings.Contains(prev.Message, "Record fields must be declared before record methods") && diag.Line == prev.Line+1 {
				continue
			}
		}

		drop, replaceIdx := classifyDiagnosticForFilter(diag, filtered, hasEarlierFatal, nameExpectedByLine, colonExpectedByLine, dotExpectedByLine)
		if drop {
			continue
		}
		if replaceIdx >= 0 {
			filtered = append(filtered[:replaceIdx], filtered[replaceIdx+1:]...)
		}
		if diag.Fatal {
			hasEarlierFatal = true
		}
		filtered = append(filtered, diag)
	}

	return filtered
}

func classifyDiagnosticForFilter(diag Diagnostic, filtered []Diagnostic, hasEarlierFatal bool, nameExpectedByLine map[int]bool, colonExpectedByLine map[int]bool, dotExpectedByLine map[int]bool) (drop bool, replaceIdx int) {
	replaceIdx = -1

	if nameExpectedByLine[diag.Line] && strings.Contains(diag.Message, "Expression expected before COLON") {
		return true, -1
	}
	if colonExpectedByLine[diag.Line] &&
		(strings.Contains(diag.Message, "variable declaration requires a type or initializer") ||
			strings.Contains(diag.Message, "must have either a type annotation or an initializer")) {
		return true, -1
	}
	if dotExpectedByLine[diag.Line] && strings.Contains(diag.Message, "already declared") {
		return true, -1
	}

	if strings.Contains(diag.Message, "instance property '") && strings.Contains(diag.Message, "cannot be a class method") {
		return true, -1
	}

	for i, existing := range filtered {
		if existing.Line != diag.Line || existing.Phase != diag.Phase {
			continue
		}

		if existing.Message == "Syntax Error: Object reference needed to read/write an object field" &&
			diag.Message == "Syntax Error: Class method or constructor expected" {
			return true, -1
		}
		if existing.Message == "Syntax Error: Class method or constructor expected" &&
			diag.Message == "Syntax Error: Object reference needed to read/write an object field" {
			return false, i
		}

		existingVisibleMember, existingIsVisible := visibleMemberName(existing.Message)
		diagVisibleMember, diagIsVisible := visibleMemberName(diag.Message)
		existingAccessibleMember, existingIsAccessible := accessibleMemberName(existing.Message)
		diagAccessibleMember, diagIsAccessible := accessibleMemberName(diag.Message)

		if existingIsVisible && diagIsAccessible && existingVisibleMember == diagAccessibleMember {
			return true, -1
		}
		if existingIsAccessible && diagIsVisible && existingAccessibleMember == diagVisibleMember {
			return false, i
		}
	}

	if !hasEarlierFatal {
		return false, replaceIdx
	}

	if diag.Phase == PhaseParsing && diag.Message == "expected 'end' to close unit declaration" {
		return true, -1
	}

	return false, replaceIdx
}

func unknownTypeName(message string) string {
	message = strings.TrimPrefix(message, "Syntax Error: ")
	const prefix = "unknown type '"
	if !strings.HasPrefix(message, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(message, prefix)
	if idx := strings.Index(rest, "'"); idx >= 0 {
		return rest[:idx]
	}
	return ""
}

func visibleMemberName(message string) (string, bool) {
	matches := diagnosticVisibleMember.FindStringSubmatch(message)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func accessibleMemberName(message string) (string, bool) {
	matches := diagnosticAccessibleMember.FindStringSubmatch(message)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

func normalizeSemanticDiagnostic(original, message string, line, column int, severity Severity) (string, int, int, string) {
	if matches := semanticVisibilityError.FindStringSubmatch(message); len(matches) == 3 {
		memberName := matches[1]
		if column > 0 {
			column++
		}
		normalized := fmt.Sprintf(`Member symbol "%s" is not visible from this scope`, memberName)
		return normalized, line, column, dwserrors.FormatDWScriptError(normalized, line, column)
	}
	if matches := semanticImplicitVisibilityError.FindStringSubmatch(message); len(matches) == 2 {
		memberName := matches[1]
		normalized := fmt.Sprintf(`Member symbol "%s" is not visible from this scope`, memberName)
		return normalized, line, column, dwserrors.FormatDWScriptError(normalized, line, column)
	}

	if strings.HasPrefix(message, "Syntax Error:") && line > 0 && column > 0 {
		renderMessage := strings.TrimSpace(strings.TrimPrefix(message, "Syntax Error:"))
		return message, line, column, dwserrors.FormatDWScriptError(renderMessage, line, column)
	}

	if strings.HasPrefix(message, "Error:") && line > 0 && column > 0 {
		return message, line, column, fmt.Sprintf("%s [line: %d, column: %d]", message, line, column)
	}

	if severity == SeverityError && line > 0 && column > 0 && !strings.HasPrefix(message, "Syntax Error:") {
		return message, line, column, dwserrors.FormatDWScriptError(message, line, column)
	}

	return message, line, column, original
}

func severityFromSemantic(sev semantic.ErrorSeverity) Severity {
	switch sev {
	case semantic.SeverityWarning:
		return SeverityWarning
	case semantic.SeverityInfo:
		return SeverityInfo
	case semantic.SeverityHint:
		return SeverityHint
	default:
		return SeverityError
	}
}

func inferStringSeverity(err string) (Severity, bool) {
	switch {
	case strings.HasPrefix(err, "Hint:"):
		return SeverityHint, false
	case strings.HasPrefix(err, "Warning:"):
		return SeverityWarning, false
	case strings.HasPrefix(err, "Info:"):
		return SeverityInfo, false
	default:
		return SeverityError, true
	}
}

func semanticCodeForSeverity(sev Severity) string {
	switch sev {
	case SeverityWarning:
		return "W_SEMANTIC"
	case SeverityInfo:
		return "I_SEMANTIC"
	case SeverityHint:
		return "H_SEMANTIC"
	default:
		return "E_SEMANTIC"
	}
}

func extractPosition(errStr string) (int, int, string) {
	lineIdx := strings.Index(errStr, "[line: ")
	if lineIdx != -1 {
		closeBracket := strings.Index(errStr[lineIdx:], "]")
		if closeBracket != -1 {
			posPart := errStr[lineIdx+7 : lineIdx+closeBracket]
			parts := strings.Split(posPart, ", column: ")
			if len(parts) == 2 {
				line, lineErr := strconv.Atoi(strings.TrimSpace(parts[0]))
				column, colErr := strconv.Atoi(strings.TrimSpace(parts[1]))
				if lineErr == nil && colErr == nil {
					message := strings.TrimSpace(errStr[:lineIdx])
					return line, column, message
				}
			}
		}
	}

	if idx := strings.LastIndex(errStr, " at "); idx != -1 {
		pos := errStr[idx+4:]
		parts := strings.Split(pos, ":")
		if len(parts) == 2 {
			line, lineErr := strconv.Atoi(parts[0])
			column, colErr := strconv.Atoi(parts[1])
			if lineErr == nil && colErr == nil {
				return line, column, strings.TrimSpace(errStr[:idx])
			}
		}
	}

	return 0, 0, errStr
}
