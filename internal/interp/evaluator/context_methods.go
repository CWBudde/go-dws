package evaluator

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/cwbudde/go-dws/internal/builtins"
	"github.com/cwbudde/go-dws/internal/interp/contracts"
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/lexer"
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/platform"
)

// ============================================================================
// Context Interface Implementation (evaluator methods)
// ============================================================================

// NewError creates an error value with location information from the current node.
func (e *builtinContext) NewError(format string, args ...interface{}) Value {
	return e.newError(e.CurrentNode(), format, args...)
}

// builtinContext binds builtin services to a single invocation. The evaluator
// retains no active-context pointer; mutable state belongs to ExecutionContext.
type builtinContext struct {
	*Evaluator
	ctx *ExecutionContext
}

func (e *Evaluator) builtinContext(ctx *ExecutionContext) *builtinContext {
	return &builtinContext{Evaluator: e, ctx: ctx}
}

func (e *builtinContext) CurrentNode() ast.Node {
	return currentNode(e.ctx)
}

func (e *builtinContext) CurrentUnit() string {
	if e.ctx == nil {
		return ""
	}
	return e.ctx.CurrentUnit()
}

// formatSettingsClassName is the script-visible name of the static class whose
// class variables are the only storage for the date/time locale settings.
const formatSettingsClassName = "FormatSettings"

// formatSettingsZoneVar is the FormatSettings class variable that selects the
// default DateTimeZone for values that do not name one explicitly.
const formatSettingsZoneVar = "Zone"

// formatSettingsStringVars binds each string-typed FormatSettings class
// variable to the DateTimeFormatSettings field it feeds. It is a package-level
// table so that reading the settings — which every date/time built-in does —
// costs no allocation.
var formatSettingsStringVars = []struct {
	field func(*builtins.DateTimeFormatSettings) *string
	name  string
}{
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.ShortDateFormat }, "ShortDateFormat"},
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.LongDateFormat }, "LongDateFormat"},
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.ShortTimeFormat }, "ShortTimeFormat"},
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.LongTimeFormat }, "LongTimeFormat"},
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.TimeAMString }, "TimeAMString"},
	{func(s *builtins.DateTimeFormatSettings) *string { return &s.TimePMString }, "TimePMString"},
}

// DateTimeFormatSettings returns the running script's date/time format
// settings. This implements the builtins.Context interface.
//
// The script-visible FormatSettings class variables are the storage, so the
// snapshot is materialised on each call and always reflects the most recent
// assignment the script made. Values that a script has not touched keep their
// defaults, and a class variable holding an unexpected value type is ignored
// rather than treated as an error.
func (e *Evaluator) DateTimeFormatSettings() *builtins.DateTimeFormatSettings {
	settings := builtins.DefaultDateTimeFormatSettings()
	class := e.typeSystem.LookupClass(formatSettingsClassName)
	if class == nil {
		return &settings
	}
	for _, classVar := range formatSettingsStringVars {
		value, owner := class.LookupClassVar(classVar.name)
		if owner == nil {
			continue
		}
		if str, ok := value.(*runtime.StringValue); ok {
			*classVar.field(&settings) = str.Value
		}
	}
	// Zone starts out as the ordinal the bootstrap stored and becomes an enum
	// value once a script assigns DateTimeZone.Local or DateTimeZone.UTC, so
	// both representations have to be understood.
	if value, owner := class.LookupClassVar(formatSettingsZoneVar); owner != nil {
		if ordinal, ok := e.ToInt64(value); ok {
			settings.Zone = builtins.TimeZone(ordinal)
		} else if ordinal, ok := e.GetEnumOrdinal(value); ok {
			settings.Zone = builtins.TimeZone(ordinal)
		}
	}
	return &settings
}

func currentNode(ctx *ExecutionContext) ast.Node {
	if ctx == nil {
		return nil
	}
	return ctx.CurrentNode()
}

// FS returns the filesystem of the platform installed on this engine, falling
// back to the build's default platform so a built-in never has to nil-check it.
func (e *Evaluator) FS() platform.FileSystem {
	if e.engineState.Platform == nil {
		// An evaluator built directly, outside NewWithOptions, has no platform
		// installed. Resolve the default per call rather than caching it here:
		// an engine can be shared across goroutines, and a lazy write to shared
		// state would be a race for the sake of an allocation nobody measured.
		return contracts.DefaultPlatform().FS()
	}
	return e.engineState.Platform.FS()
}

// RandSource returns the random number generator for built-in functions.
func (e *Evaluator) RandSource() *rand.Rand {
	return e.engineState.Random
}

// GetRandSeed returns the current random number generator seed value.
func (e *Evaluator) GetRandSeed() int64 {
	return e.engineState.RandomSeed
}

// SetRandSeed sets the random number generator seed.
func (e *Evaluator) SetRandSeed(seed int64) {
	e.engineState.RandomSeed = seed
	e.engineState.Random.Seed(seed)
}

// Write outputs a string to the configured output writer without a newline.
func (e *Evaluator) Write(s string) {
	if e.output != nil {
		_, _ = io.WriteString(e.output, s)
	}
}

// WriteLine outputs a string to the configured output writer with a newline.
func (e *Evaluator) WriteLine(s string) {
	if e.output != nil {
		_, _ = fmt.Fprintln(e.output, s)
	}
}

// IsAssigned checks if a Variant value has been assigned (is not uninitialized).
func (e *Evaluator) IsAssigned(value Value) bool {
	if value == nil {
		return false
	}

	if _, ok := value.(*runtime.NilValue); ok {
		return false
	}

	// A nil function/method pointer (unbound proc-typed var or field) is unassigned.
	if funcPtr, ok := value.(*runtime.FunctionPointerValue); ok {
		return !funcPtr.IsNil()
	}

	if intfVal, ok := value.(*runtime.InterfaceInstance); ok {
		return intfVal.Object != nil
	}

	if wrapper, ok := value.(runtime.VariantWrapper); ok {
		unwrapped := wrapper.UnwrapVariant()
		return unwrapped != nil
	}

	return true
}

// GetCallStackString returns a formatted string representation of the current call stack.
func (e *builtinContext) GetCallStackString() string {
	if e.ctx == nil {
		return ""
	}
	return e.ctx.GetCallStack().String()
}

// GetCallStackArray returns the current call stack as an array of records.
func (e *builtinContext) GetCallStackArray() Value {
	if e.ctx == nil {
		return &runtime.ArrayValue{
			Elements:  []runtime.Value{},
			ArrayType: types.NewDynamicArrayType(types.VARIANT),
		}
	}

	frames := e.ctx.GetCallStack().Frames()
	elements := make([]runtime.Value, len(frames))

	for idx, frame := range frames {
		fields := make(map[string]runtime.Value)
		fields["FunctionName"] = &runtime.StringValue{Value: frame.FunctionName}

		if frame.Position != nil {
			fields["Line"] = &runtime.IntegerValue{Value: int64(frame.Position.Line)}
			fields["Column"] = &runtime.IntegerValue{Value: int64(frame.Position.Column)}
		} else {
			fields["Line"] = &runtime.IntegerValue{Value: 0}
			fields["Column"] = &runtime.IntegerValue{Value: 0}
		}

		elements[idx] = &runtime.RecordValue{
			Fields:     fields,
			RecordType: nil,
		}
	}

	return &runtime.ArrayValue{
		Elements:  elements,
		ArrayType: types.NewDynamicArrayType(types.VARIANT),
	}
}

// RaiseAssertionFailed raises an EAssertionFailed exception with an optional custom message.
// Self-contained: no longer delegates to ExceptionManager.
func (e *builtinContext) RaiseAssertionFailed(customMessage string) {
	ctx := e.ctx
	if ctx == nil {
		return // No context available, cannot raise exception
	}

	// Build message with position info if available
	var message string
	if currentNode := e.CurrentNode(); currentNode != nil {
		pos := currentNode.Pos()
		message = fmt.Sprintf("Assertion failed [line: %d, column: %d]", pos.Line, pos.Column)
	} else {
		message = "Assertion failed"
	}

	// Append custom message if provided
	if customMessage != "" {
		message = message + " : " + customMessage
	}

	// Look up EAssertionFailed class
	excClass := e.typeSystem.LookupClass("EAssertionFailed")
	if excClass == nil {
		// Fallback to base Exception if class not found
		excClass = e.typeSystem.LookupClass("Exception")
	}

	// Get metadata and create instance
	var metadata *runtime.ClassMetadata
	var instance *runtime.ObjectInstance
	if excClass != nil {
		if classInfo, ok := excClass.(runtime.IClassInfo); ok {
			metadata = classInfo.GetMetadata()
			instance = runtime.NewObjectInstance(classInfo)
			instance.SetField("Message", &runtime.StringValue{Value: message})
		}
	}

	// Create and set exception
	exc := runtime.NewException(metadata, instance, message, nil, ctx.CallStack())
	ctx.SetException(exc)
}

// RaiseException raises a DWScript exception so try/except blocks can handle it.
// Implements the builtins.Context raiser interface (mirrors Interpreter.RaiseException),
// so builtins running on the evaluator context raise catchable exceptions instead of
// falling back to bare error values.
func (e *builtinContext) RaiseException(className, message string, pos any) {
	ctx := e.ctx
	if ctx == nil {
		return // No context available, cannot raise exception
	}

	var lexerPos *lexer.Position
	switch p := pos.(type) {
	case *lexer.Position:
		lexerPos = p
	case lexer.Position:
		lexerPos = &p
	}

	ctx.SetException(e.createException(className, message, lexerPos, ctx))
}

// EvalFunctionPointer executes a function pointer with given arguments.
func (e *builtinContext) EvalFunctionPointer(funcPtr Value, args []Value) Value {
	return e.executeFunctionPointerDirect(funcPtr, args, e.CurrentNode(), e.ctx)
}
