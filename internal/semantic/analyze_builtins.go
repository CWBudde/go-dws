package semantic

import (
	"github.com/cwbudde/go-dws/internal/types"
	"github.com/cwbudde/go-dws/pkg/ast"
	"github.com/cwbudde/go-dws/pkg/ident"
)

// ============================================================================
// Expression Analysis
// ============================================================================

// isLValue checks if an expression is an lvalue (can be assigned to).
//
// An lvalue is:
//   - An identifier (variable)
//   - An index expression (array[i], string[i])
//   - A member access expression (record.field, object.field)
func (a *Analyzer) isLValue(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.Identifier:
		return true
	case *ast.IndexExpression:
		return true
	case *ast.MemberAccessExpression:
		return true
	default:
		return false
	}
}

// isBuiltinFunction checks if a name refers to a built-in function.
func (a *Analyzer) isBuiltinFunction(name string) bool {
	// Normalize to lowercase for case-insensitive matching
	lowerName := ident.Normalize(name)

	// List of all built-in functions that can be called without parentheses
	// This should match the list in the interpreter's isBuiltinFunction
	switch lowerName {
	case "println", "print", "ord", "integer", "length", "copy", "concat",
		"indexof", "contains", "reverse", "sort", "pos", "uppercase",
		"lowercase", "trim", "trimleft", "trimright", "stringreplace", "stringofchar",
		"substr", "substring", "leftstr", "rightstr", "midstr",
		"strbeginswith", "strendswith", "strcontains", "posex", "revpos", "strfind",
		"strsplit", "strjoin", "strarraypack",
		"strbefore", "strbeforelast", "strafter", "strafterlast", "strbetween",
		"isdelimiter", "lastdelimiter", "finddelimiter",
		"padleft", "padright", "strdeleteleft", "deleteleft", "strdeleteright", "deleteright",
		"reversestring", "quotedstr", "stringofstring", "dupestring",
		"normalizestring", "normalize", "stripaccents",
		"sametext", "comparetext", "comparestr", "ansicomparetext", "ansicomparestr",
		"comparelocalestr", "strmatches", "strisascii",
		"format", "abs", "min", "max", "sqr", "power", "sqrt", "sin",
		"cos", "tan", "random", "randomize", "randomint", "setrandseed", "randseed", "randg", "exp", "ln", "log2", "round",
		"trunc", "frac", "chr", "setlength", "high", "low", "assigned",
		"degtorad", "radtodeg", "arcsin", "arccos", "arctan", "arctan2",
		"cotan", "hypot", "sinh", "cosh", "tanh", "arcsinh", "arccosh", "arctanh",
		"typeof", "typeofclass", "sizeof", "typename", "delete", "strtoint", "strtofloat",
		"inttostr", "inttobin", "floattostr", "floattostrf", "booltostr", "strtobool",
		"vartostr", "varisnull", "varisempty", "varisclear", "varisarray", "varisstr", "varisnumeric", "vartype", "varclear",
		"include", "exclude", "map", "filter", "reduce", "foreach",
		"maxint", "minint",
		"now", "date", "time", "utcdatetime", "encodedate", "encodetime",
		"encodedatetime", "yearof", "monthof", "dayof", "hourof", "minuteof",
		"secondof", "millisecondof", "dayofweek", "dayofyear", "weekofyear",
		"datetimetostr", "datetostr", "timetostr", "formatdatetime",
		"incyear", "incmonth", "incweek", "incday", "inchour", "incminute",
		"incsecond", "incmillisecond", "daysbetween", "hoursbetween",
		"minutesbetween", "secondsbetween", "millisecondsbetween",
		"isleapyear", "daysinmonth", "daysinyear", "startofday", "endofday",
		"startofmonth", "endofmonth", "startofyear", "endofyear", "istoday",
		"isyesterday", "istomorrow", "issameday", "comparedate", "comparetime",
		"comparedatetime", "parsejson", "tojson", "tojsonformatted",
		"jsonhasfield", "jsonkeys", "jsonvalues", "jsonlength",
		"getstacktrace", "getcallstack":
		return true
	default:
		return false
	}
}

// builtinDeclarationName returns the canonical casing for a built-in function name.
// Used for pedantic hinting when the source uses a different case.
func (a *Analyzer) builtinDeclarationName(name string) string {
	switch ident.Normalize(name) {
	case "println":
		return "PrintLn"
	case "print":
		return "Print"
	default:
		return name
	}
}

// getBuiltinFunctionPointerType returns the function pointer type for a built-in function
// when it's used as a function reference (not called). Returns nil if the builtin
// doesn't have a known function pointer type or shouldn't be used as a reference.
//
// This is needed for higher-order functions like Map, Filter, etc. that take
// function references as arguments.
func (a *Analyzer) getBuiltinFunctionPointerType(name string) *types.FunctionPointerType {
	lowerName := ident.Normalize(name)

	switch lowerName {
	// Type conversion functions commonly used with Map
	case "inttostr":
		// IntToStr(value: Integer): String
		return types.NewFunctionPointerType([]types.Type{types.INTEGER}, types.STRING)
	case "floattostr":
		// FloatToStr(value: Float): String
		return types.NewFunctionPointerType([]types.Type{types.FLOAT}, types.STRING)
	case "booltostr":
		// BoolToStr(value: Boolean): String
		return types.NewFunctionPointerType([]types.Type{types.BOOLEAN}, types.STRING)
	case "strtoint":
		// StrToInt(s: String): Integer
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.INTEGER)
	case "strtofloat":
		// StrToFloat(s: String): Float
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.FLOAT)
	case "uppercase":
		// UpperCase(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	case "lowercase":
		// LowerCase(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	case "trim":
		// Trim(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	case "trimleft":
		// TrimLeft(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	case "trimright":
		// TrimRight(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	case "length":
		// Length(s: String): Integer (also works with arrays)
		return types.NewFunctionPointerType([]types.Type{types.VARIANT}, types.INTEGER)
	case "chr":
		// Chr(code: Integer): String
		return types.NewFunctionPointerType([]types.Type{types.INTEGER}, types.STRING)
	case "ord":
		// Ord(s: String): Integer (also works with chars, enums)
		return types.NewFunctionPointerType([]types.Type{types.VARIANT}, types.INTEGER)
	case "abs":
		// Abs(value: Float): Float
		return types.NewFunctionPointerType([]types.Type{types.FLOAT}, types.FLOAT)
	case "sqr":
		// Sqr(value: Float): Float
		return types.NewFunctionPointerType([]types.Type{types.FLOAT}, types.FLOAT)
	case "sqrt":
		// Sqrt(value: Float): Float
		return types.NewFunctionPointerType([]types.Type{types.FLOAT}, types.FLOAT)
	case "reversestring":
		// ReverseString(s: String): String
		return types.NewFunctionPointerType([]types.Type{types.STRING}, types.STRING)
	default:
		// No known function pointer type for this builtin
		return nil
	}
}
