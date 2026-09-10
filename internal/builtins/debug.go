package builtins

import (
	"github.com/cwbudde/go-dws/internal/interp/runtime"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Debug Built-ins
// ============================================================================
//
// DWScript exposes three introspection points over the running call stack:
//
//	CurrentSourceCodeLocation : TSourceCodeLocation
//	CallerSourceCodeLocation  : TSourceCodeLocation
//	CurrentStackTrace         : String
//
// Their values depend on where they appear and on the frames below them, so the
// evaluator computes them; this file only owns the shared record type and the
// names, which the semantic analyzer needs as well.

// MainModuleName is the module name DWScript reports for the main script.
const MainModuleName = "*MainModule*"

// SourceCodeLocationTypeName is the script-visible name of the location record.
const SourceCodeLocationTypeName = "TSourceCodeLocation"

// Field names of TSourceCodeLocation, in declaration order.
const (
	SourceCodeLocationFileField = "File"
	SourceCodeLocationLineField = "Line"
	SourceCodeLocationNameField = "Name"
)

// Names of the debug built-ins the evaluator resolves against the call stack.
const (
	CurrentSourceCodeLocationName = "CurrentSourceCodeLocation"
	CallerSourceCodeLocationName  = "CallerSourceCodeLocation"
	CurrentStackTraceName         = "CurrentStackTrace"
)

// sourceCodeLocationType is the single shared TSourceCodeLocation record type,
// so that the analyzer and the runtime agree on one type identity.
var sourceCodeLocationType = types.NewRecordType(SourceCodeLocationTypeName, map[string]types.Type{
	SourceCodeLocationFileField: types.STRING,
	SourceCodeLocationLineField: types.INTEGER,
	SourceCodeLocationNameField: types.STRING,
})

// SourceCodeLocationType returns the TSourceCodeLocation record type.
func SourceCodeLocationType() *types.RecordType {
	return sourceCodeLocationType
}

// NewSourceCodeLocation builds a TSourceCodeLocation record value.
func NewSourceCodeLocation(file string, line int64, name string) *runtime.RecordValue {
	location := runtime.NewRecordValue(sourceCodeLocationType, nil)
	location.SetRecordField(SourceCodeLocationFileField, &runtime.StringValue{Value: file})
	location.SetRecordField(SourceCodeLocationLineField, &runtime.IntegerValue{Value: line})
	location.SetRecordField(SourceCodeLocationNameField, &runtime.StringValue{Value: name})
	return location
}
