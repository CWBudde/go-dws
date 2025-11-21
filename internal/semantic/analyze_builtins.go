package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
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
	lowerName := strings.ToLower(name)

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
