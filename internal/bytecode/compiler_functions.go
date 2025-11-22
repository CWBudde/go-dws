package bytecode

import (
	"fmt"

	"github.com/cwbudde/go-dws/pkg/ast"
	pkgident "github.com/cwbudde/go-dws/pkg/ident"
)

func (c *Compiler) compileLambdaExpression(expr *ast.LambdaExpression) error {
	if expr == nil {
		return c.errorf(nil, "nil lambda expression")
	}

	name := fmt.Sprintf("lambda@%d", lineOf(expr))
	child := c.newChildCompiler(name)
	child.beginScope()

	for _, param := range expr.Parameters {
		if param == nil || param.Name == nil {
			return c.errorf(expr, "lambda parameter missing identifier")
		}
		paramType := typeFromAnnotation(param.Type)
		if _, err := child.declareLocal(param.Name, paramType); err != nil {
			return err
		}
	}

	if expr.Body != nil {
		if err := child.compileBlock(expr.Body); err != nil {
			return err
		}
	} else {
		child.chunk.WriteSimple(OpLoadNil, lineOf(expr))
		child.chunk.Write(OpReturn, 1, 0, lineOf(expr))
	}

	child.endScope()
	child.chunk.LocalCount = int(child.maxSlot)
	child.ensureFunctionReturn(lineOf(expr))
	child.chunk.Optimize()

	fn := NewFunctionObject(name, child.chunk, len(expr.Parameters))
	fn.UpvalueDefs = child.buildUpvalueDefs()

	fnIndex := c.chunk.AddConstant(FunctionValue(fn))
	if fnIndex > 0xFFFF {
		return c.errorf(expr, "constant pool overflow")
	}

	upvalueCount := len(fn.UpvalueDefs)
	if upvalueCount > 0xFF {
		return c.errorf(expr, "too many upvalues in lambda (max 255)")
	}

	c.chunk.Write(OpClosure, byte(upvalueCount), uint16(fnIndex), lineOf(expr))
	return nil
}

func (c *Compiler) compileCallExpression(expr *ast.CallExpression) error {
	argCount := len(expr.Arguments)
	if argCount > 0xFF {
		return c.errorf(expr, "too many arguments in function call: %d", argCount)
	}

	if ident, ok := expr.Function.(*ast.Identifier); ok {
		if info, ok := c.directCallInfo(ident); ok {
			for _, arg := range expr.Arguments {
				if err := c.compileExpression(arg); err != nil {
					return err
				}
			}
			c.chunk.Write(OpCall, byte(argCount), info.constIndex, lineOf(expr))
			return nil
		}
		// Check if it's a builtin function (only if not shadowed by local/upvalue/global)
		if c.isBuiltinFunction(ident.Value) {
			// Check if shadowed by local, enclosing variable, or global - if so, skip builtin path
			if _, ok := c.resolveLocal(ident.Value); !ok {
				if !c.hasEnclosingLocal(ident.Value) {
					if _, ok := c.resolveGlobal(ident.Value); !ok {
						// Not shadowed - compile as builtin call
						for _, arg := range expr.Arguments {
							if err := c.compileExpression(arg); err != nil {
								return err
							}
						}
						builtinValue := BuiltinValue(pkgident.Normalize(ident.Value))
						constIdx := c.chunk.AddConstant(builtinValue)
						if constIdx > 0xFFFF {
							return c.errorf(expr, "too many constants")
						}
						c.chunk.Write(OpCall, byte(argCount), uint16(constIdx), lineOf(expr))
						return nil
					}
				}
			}
		}
	}

	if err := c.compileExpression(expr.Function); err != nil {
		return err
	}

	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	c.chunk.Write(OpCallIndirect, byte(argCount), 0, lineOf(expr))
	return nil
}

func (c *Compiler) directCallInfo(ident *ast.Identifier) (functionInfo, bool) {
	if ident == nil || c.functions == nil {
		return functionInfo{}, false
	}

	if _, ok := c.resolveLocal(ident.Value); ok {
		return functionInfo{}, false
	}
	if c.hasEnclosingLocal(ident.Value) {
		return functionInfo{}, false
	}

	info, ok := c.functions[pkgident.Normalize(ident.Value)]
	return info, ok
}

// isBuiltinFunction checks if a name refers to a built-in function.
// This should match the list in internal/semantic/analyze_builtins.go
func (c *Compiler) isBuiltinFunction(name string) bool {
	lowerName := pkgident.Normalize(name)
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
