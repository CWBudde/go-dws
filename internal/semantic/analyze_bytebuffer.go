package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// byteBufferMemberType returns the result type of a ByteBuffer member, whether
// it is reached as a property (b.Length), as a parameterless call (b.ToJSON) or
// as a method call (b.GetByte(0)).
//
// ByteBuffer members are intrinsics rather than declared class members, so they
// are resolved from this table instead of from the type registry. Unknown member
// names resolve to VOID, keeping diagnostics for the runtime, which knows the
// exact arity of every accessor.
func byteBufferMemberType(memberName string) types.Type {
	switch ident.Normalize(memberName) {
	case "length", "position",
		"getbyte", "getint8", "getword", "getint16",
		"getdword", "getint32", "getint64":
		return types.INTEGER

	case "getsingle", "getdouble", "getextended":
		return types.FLOAT

	case "getdata", "tojson", "todatastring", "tobase64", "tohexstring":
		return types.STRING

	case "getintegers":
		return types.NewDynamicArrayType(types.INTEGER)

	case "copy":
		return types.BYTE_BUFFER

	default:
		// Setters, Assign* and anything unrecognised: no value.
		return types.VOID
	}
}

// analyzeByteBufferMethodResult analyzes the arguments of a method call on a
// ByteBuffer receiver and returns the member's result type.
func (a *Analyzer) analyzeByteBufferMethodResult(memberName string, args []ast.Expression) types.Type {
	for _, arg := range args {
		a.analyzeExpression(arg)
	}
	return byteBufferMemberType(memberName)
}

// isByteBufferTypeName reports whether name denotes the built-in ByteBuffer type
// and is not shadowed by a user symbol or a user-declared type.
func (a *Analyzer) isByteBufferTypeName(name string) bool {
	if !ident.Equal(name, "ByteBuffer") {
		return false
	}
	if _, shadowed := a.symbols.Resolve(name); shadowed {
		return false
	}
	return !a.hasType(name)
}
