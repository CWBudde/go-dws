package builtins

// VarTypeConstant names a script-visible Variant type code constant.
type VarTypeConstant struct {
	Name  string
	Value int64
}

// VarTypeConstants lists the Variant type-code constants DWScript exposes to
// scripts, so that `VarType(x) = varString` and `VarAsType(x, varInteger)`
// compile and agree with the codes VarType returns.
//
// The values follow Delphi's System.Variants codes. DWScript's native integer
// is 64-bit, so VarType reports varInt64 for integer values.
func VarTypeConstants() []VarTypeConstant {
	return []VarTypeConstant{
		{"varEmpty", varEmpty},
		{"varNull", varNull},
		{"varSmallint", varSmallint},
		{"varInteger", varInteger},
		{"varSingle", varSingle},
		{"varDouble", varDouble},
		{"varCurrency", varCurrency},
		{"varDate", varDate},
		{"varOleStr", varOleStr},
		{"varDispatch", varDispatch},
		{"varError", varError},
		{"varBoolean", varBoolean},
		{"varVariant", varVariant},
		{"varUnknown", varUnknown},
		{"varDecimal", varDecimal},
		{"varByte", varByte},
		{"varWord", varWord},
		{"varLongWord", varLongWord},
		{"varInt64", varInt64},
		{"varUInt64", varUInt64},
		{"varString", varString},
		{"varUString", varUString},
		{"varArray", varArray},
	}
}
