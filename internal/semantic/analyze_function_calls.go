package semantic

import (
	"strings"

	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Expression Analysis
// ============================================================================
func (a *Analyzer) analyzeCallExpression(expr *ast.CallExpression) types.Type {
	// Handle member access expressions (method calls like obj.Method())
	if memberAccess, ok := expr.Function.(*ast.MemberAccessExpression); ok {
		// Analyze the member access to get the method type
		methodType := a.analyzeMemberAccessExpression(memberAccess)
		if methodType == nil {
			// Error already reported by analyzeMemberAccessExpression
			return nil
		}

		// Verify it's a function type
		funcType, ok := methodType.(*types.FunctionType)
		if !ok {
			a.addError("cannot call non-function type %s at %s",
				methodType.String(), expr.Token.Pos.String())
			return nil
		}

		// Validate argument count
		if len(expr.Arguments) != len(funcType.Parameters) {
			a.addError("method call expects %d argument(s), got %d at %s",
				len(funcType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
		}

		// Validate argument types
		for i, arg := range expr.Arguments {
			if i >= len(funcType.Parameters) {
				break // Already reported count mismatch
			}

			// Task 9.2b: Validate var parameter receives an lvalue
			isVar := len(funcType.VarParams) > i && funcType.VarParams[i]
			if isVar && !a.isLValue(arg) {
				a.addError("var parameter %d requires a variable (identifier, array element, or field), got %s at %s",
					i+1, arg.String(), arg.Pos().String())
			}

			paramType := funcType.Parameters[i]
			argType := a.analyzeExpressionWithExpectedType(arg, paramType)
			if argType != nil && !a.canAssign(argType, paramType) {
				a.addError("argument %d has type %s, expected %s at %s",
					i+1, argType.String(), paramType.String(),
					expr.Token.Pos.String())
			}
		}

		return funcType.ReturnType
	}

	// Handle regular function calls (identifier-based)
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		a.addError("function call must use identifier or member access at %s", expr.Token.Pos.String())
		return nil
	}

	// Look up function
	sym, ok := a.symbols.Resolve(funcIdent.Value)
	if !ok {
		// Check if it's a built-in function (using new dispatcher)
		if resultType, isBuiltin := a.analyzeBuiltinFunction(funcIdent.Value, expr.Arguments, expr); isBuiltin {
			return resultType
		}

		// TrimRight built-in function
		if strings.EqualFold(funcIdent.Value, "TrimRight") {
			// TrimRight takes one string argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'TrimRight' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's a string
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'TrimRight' expects string as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// StringReplace built-in function
		if strings.EqualFold(funcIdent.Value, "StringReplace") {
			// StringReplace takes 3-4 arguments: str, old, new, [count]
			if len(expr.Arguments) < 3 || len(expr.Arguments) > 4 {
				a.addError("function 'StringReplace' expects 3 or 4 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// First argument: string to search in
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			if arg1Type != nil && arg1Type != types.STRING {
				a.addError("function 'StringReplace' expects string as first argument, got %s at %s",
					arg1Type.String(), expr.Token.Pos.String())
			}
			// Second argument: old substring
			arg2Type := a.analyzeExpression(expr.Arguments[1])
			if arg2Type != nil && arg2Type != types.STRING {
				a.addError("function 'StringReplace' expects string as second argument, got %s at %s",
					arg2Type.String(), expr.Token.Pos.String())
			}
			// Third argument: new substring
			arg3Type := a.analyzeExpression(expr.Arguments[2])
			if arg3Type != nil && arg3Type != types.STRING {
				a.addError("function 'StringReplace' expects string as third argument, got %s at %s",
					arg3Type.String(), expr.Token.Pos.String())
			}
			// Optional fourth argument: count (integer)
			if len(expr.Arguments) == 4 {
				arg4Type := a.analyzeExpression(expr.Arguments[3])
				if arg4Type != nil && arg4Type != types.INTEGER {
					a.addError("function 'StringReplace' expects integer as fourth argument, got %s at %s",
						arg4Type.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// StringOfChar built-in function
		if strings.EqualFold(funcIdent.Value, "StringOfChar") {
			// StringOfChar takes exactly 2 arguments: character (string) and count (integer)
			if len(expr.Arguments) != 2 {
				a.addError("function 'StringOfChar' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// First argument: character (string)
			arg1Type := a.analyzeExpression(expr.Arguments[0])
			if arg1Type != nil && arg1Type != types.STRING {
				a.addError("function 'StringOfChar' expects string as first argument, got %s at %s",
					arg1Type.String(), expr.Token.Pos.String())
			}
			// Second argument: count (integer)
			arg2Type := a.analyzeExpression(expr.Arguments[1])
			if arg2Type != nil && arg2Type != types.INTEGER {
				a.addError("function 'StringOfChar' expects integer as second argument, got %s at %s",
					arg2Type.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// Format built-in function
		if strings.EqualFold(funcIdent.Value, "Format") {
			// Format takes exactly 2 arguments: format string and array of values
			if len(expr.Arguments) != 2 {
				a.addError("Format() expects exactly 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// First argument: format string (must be String)
			fmtType := a.analyzeExpression(expr.Arguments[0])
			if fmtType != nil && fmtType != types.STRING {
				a.addError("Format() expects string as first argument, got %s at %s",
					fmtType.String(), expr.Token.Pos.String())
			}
			// Second argument: array of values (must be Array type)
			// Task 9.156 & 9.236: Pass ARRAY_OF_CONST (array of Variant) as expected type
			// This allows heterogeneous arrays like ['string', 123, 3.14]
			arrType := a.analyzeExpressionWithExpectedType(expr.Arguments[1], types.ARRAY_OF_CONST)
			if arrType != nil {
				if _, isArray := arrType.(*types.ArrayType); !isArray {
					a.addError("Format() expects array as second argument, got %s at %s",
						arrType.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// Low built-in function
		if strings.EqualFold(funcIdent.Value, "Low") {
			// Low takes one argument (array, enum, or type meta-value) and returns a value of the appropriate type
			if len(expr.Arguments) != 1 {
				a.addError("function 'Low' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array, enum, or basic type (type meta-value)
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); isArray {
					// For arrays, return Integer
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					// For enums, return the same enum type
					return enumType
				}
				// Task 9.134: Handle type meta-values (Integer, Float, Boolean, String)
				switch argType {
				case types.INTEGER:
					return types.INTEGER
				case types.FLOAT:
					return types.FLOAT
				case types.BOOLEAN:
					return types.BOOLEAN
				case types.STRING:
					// String doesn't have a low value, but we allow it for consistency
					return types.INTEGER
				}
				// Neither array, enum, nor type meta-value
				a.addError("function 'Low' expects array, enum, or type name, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// High built-in function
		if strings.EqualFold(funcIdent.Value, "High") {
			// High takes one argument (array, enum, or type meta-value) and returns a value of the appropriate type
			if len(expr.Arguments) != 1 {
				a.addError("function 'High' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument
			argType := a.analyzeExpression(expr.Arguments[0])
			// Verify it's an array, enum, or basic type (type meta-value)
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); isArray {
					// For arrays, return Integer
					return types.INTEGER
				}
				if enumType, isEnum := argType.(*types.EnumType); isEnum {
					// For enums, return the same enum type
					return enumType
				}
				// Task 9.134: Handle type meta-values (Integer, Float, Boolean, String)
				switch argType {
				case types.INTEGER:
					return types.INTEGER
				case types.FLOAT:
					return types.FLOAT
				case types.BOOLEAN:
					return types.BOOLEAN
				case types.STRING:
					// String doesn't have a high value, but we allow it for consistency
					return types.INTEGER
				}
				// Neither array, enum, nor type meta-value
				a.addError("function 'High' expects array, enum, or type name, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// SetLength built-in function
		if strings.EqualFold(funcIdent.Value, "SetLength") {
			// SetLength takes two arguments (array, integer) and returns void
			if len(expr.Arguments) != 2 {
				a.addError("function 'SetLength' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// Analyze the first argument (array)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'SetLength' expects array as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Analyze the second argument (integer)
			lengthType := a.analyzeExpression(expr.Arguments[1])
			if lengthType != nil && lengthType != types.INTEGER {
				a.addError("function 'SetLength' expects integer as second argument, got %s at %s",
					lengthType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Add built-in function
		if strings.EqualFold(funcIdent.Value, "Add") {
			// Add takes two arguments (array, element) and returns void
			if len(expr.Arguments) != 2 {
				a.addError("function 'Add' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// Analyze the first argument (array)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil {
				if _, isArray := argType.(*types.ArrayType); !isArray {
					a.addError("function 'Add' expects array as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Analyze the second argument (element to add)
			a.analyzeExpression(expr.Arguments[1])
			return types.VOID
		}

		// Delete built-in function
		// Delete(array, index) - for arrays (2 args)
		// Delete(string, pos, count) - for strings (3 args)
		if strings.EqualFold(funcIdent.Value, "Delete") {
			if len(expr.Arguments) == 2 {
				// Array delete: Delete(array, index)
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil {
					if _, isArray := argType.(*types.ArrayType); !isArray {
						a.addError("function 'Delete' expects array as first argument for 2-argument form, got %s at %s",
							argType.String(), expr.Token.Pos.String())
					}
				}
				indexType := a.analyzeExpression(expr.Arguments[1])
				if indexType != nil && indexType != types.INTEGER {
					a.addError("function 'Delete' expects integer as second argument, got %s at %s",
						indexType.String(), expr.Token.Pos.String())
				}
				return types.VOID
			} else if len(expr.Arguments) == 3 {
				// String delete: Delete(string, pos, count)
				if _, ok := expr.Arguments[0].(*ast.Identifier); !ok {
					a.addError("function 'Delete' first argument must be a variable at %s",
						expr.Token.Pos.String())
				} else {
					strType := a.analyzeExpression(expr.Arguments[0])
					if strType != nil && strType != types.STRING {
						a.addError("function 'Delete' first argument must be String for 3-argument form, got %s at %s",
							strType.String(), expr.Token.Pos.String())
					}
				}
				posType := a.analyzeExpression(expr.Arguments[1])
				if posType != nil && posType != types.INTEGER {
					a.addError("function 'Delete' second argument must be Integer, got %s at %s",
						posType.String(), expr.Token.Pos.String())
				}
				countType := a.analyzeExpression(expr.Arguments[2])
				if countType != nil && countType != types.INTEGER {
					a.addError("function 'Delete' third argument must be Integer, got %s at %s",
						countType.String(), expr.Token.Pos.String())
				}
				return types.VOID
			} else {
				a.addError("function 'Delete' expects 2 or 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
		}

		// IntToStr built-in function
		if strings.EqualFold(funcIdent.Value, "IntToStr") {
			// IntToStr takes one integer argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'IntToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Integer or a subrange of Integer
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.INTEGER {
				// Check if it's a subrange type with Integer base
				if subrange, ok := argType.(*types.SubrangeType); ok {
					if subrange.BaseType != types.INTEGER {
						a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
							argType.String(), expr.Token.Pos.String())
					}
				} else {
					a.addError("function 'IntToStr' expects Integer as argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// IntToBin built-in function - Task 9.37
		if strings.EqualFold(funcIdent.Value, "IntToBin") {
			// IntToBin takes two integer arguments (value, digits) and returns a string
			if len(expr.Arguments) != 2 {
				a.addError("function 'IntToBin' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze first argument (value) - must be Integer or subrange of Integer
			argType1 := a.analyzeExpression(expr.Arguments[0])
			if argType1 != nil && argType1 != types.INTEGER {
				// Check if it's a subrange type with Integer base
				if subrange, ok := argType1.(*types.SubrangeType); ok {
					if subrange.BaseType != types.INTEGER {
						a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
							argType1.String(), expr.Token.Pos.String())
					}
				} else {
					a.addError("function 'IntToBin' expects Integer as first argument, got %s at %s",
						argType1.String(), expr.Token.Pos.String())
				}
			}
			// Analyze second argument (digits) - must be Integer
			argType2 := a.analyzeExpression(expr.Arguments[1])
			if argType2 != nil && argType2 != types.INTEGER {
				// Check if it's a subrange type with Integer base
				if subrange, ok := argType2.(*types.SubrangeType); ok {
					if subrange.BaseType != types.INTEGER {
						a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
							argType2.String(), expr.Token.Pos.String())
					}
				} else {
					a.addError("function 'IntToBin' expects Integer as second argument, got %s at %s",
						argType2.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// StrToInt built-in function
		if strings.EqualFold(funcIdent.Value, "StrToInt") {
			// StrToInt takes one string argument and returns an integer
			if len(expr.Arguments) != 1 {
				a.addError("function 'StrToInt' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument and verify it's String
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'StrToInt' expects String as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.INTEGER
		}

		// FloatToStr built-in function
		if strings.EqualFold(funcIdent.Value, "FloatToStr") {
			// FloatToStr takes one float argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'FloatToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Float (or coercible to Float, e.g. Integer)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && !a.canAssign(argType, types.FLOAT) {
				a.addError("function 'FloatToStr' expects Float as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// BoolToStr built-in function
		if strings.EqualFold(funcIdent.Value, "BoolToStr") {
			// BoolToStr takes one boolean argument and returns a string
			if len(expr.Arguments) != 1 {
				a.addError("function 'BoolToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument and verify it's Boolean
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.BOOLEAN {
				a.addError("function 'BoolToStr' expects Boolean as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.STRING
		}

		// StrToFloat built-in function
		if strings.EqualFold(funcIdent.Value, "StrToFloat") {
			// StrToFloat takes one string argument and returns a float
			if len(expr.Arguments) != 1 {
				a.addError("function 'StrToFloat' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			// Analyze the argument and verify it's String
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && argType != types.STRING {
				a.addError("function 'StrToFloat' expects String as argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.FLOAT
		}

		// Assert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Assert") {
			// Assert takes 1-2 arguments: Boolean condition and optional String message
			if len(expr.Arguments) < 1 || len(expr.Arguments) > 2 {
				a.addError("function 'Assert' expects 1-2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			// First argument must be Boolean
			condType := a.analyzeExpression(expr.Arguments[0])
			if condType != nil && condType != types.BOOLEAN {
				a.addError("function 'Assert' first argument must be Boolean, got %s at %s",
					condType.String(), expr.Token.Pos.String())
			}
			// If there's a second argument (message), it must be String
			if len(expr.Arguments) == 2 {
				msgType := a.analyzeExpression(expr.Arguments[1])
				if msgType != nil && msgType != types.STRING {
					a.addError("function 'Assert' second argument must be String, got %s at %s",
						msgType.String(), expr.Token.Pos.String())
				}
			}
			return types.VOID
		}

		// Insert built-in procedure
		if strings.EqualFold(funcIdent.Value, "Insert") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'Insert' expects 3 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			sourceType := a.analyzeExpression(expr.Arguments[0])
			if sourceType != nil && sourceType != types.STRING {
				a.addError("function 'Insert' first argument must be String, got %s at %s",
					sourceType.String(), expr.Token.Pos.String())
			}
			if _, ok := expr.Arguments[1].(*ast.Identifier); !ok {
				a.addError("function 'Insert' second argument must be a variable at %s",
					expr.Token.Pos.String())
			} else {
				targetType := a.analyzeExpression(expr.Arguments[1])
				if targetType != nil && targetType != types.STRING {
					a.addError("function 'Insert' second argument must be String, got %s at %s",
						targetType.String(), expr.Token.Pos.String())
				}
			}
			posType := a.analyzeExpression(expr.Arguments[2])
			if posType != nil && posType != types.INTEGER {
				a.addError("function 'Insert' third argument must be Integer, got %s at %s",
					posType.String(), expr.Token.Pos.String())
			}
			return types.VOID
		}

		// Task 9.227: Higher-order functions for working with lambdas
		if strings.EqualFold(funcIdent.Value, "Map") {
			// Map(array, lambda) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Map' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Filter") {
			// Filter(array, predicate) -> array
			if len(expr.Arguments) != 2 {
				a.addError("function 'Filter' expects 2 arguments (array, predicate), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			arrayType := a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			// Verify first argument is an array
			if arrType, ok := arrayType.(*types.ArrayType); ok {
				return arrType // Return same array type
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "Reduce") {
			// Reduce(array, lambda, initial) -> value
			if len(expr.Arguments) != 3 {
				a.addError("function 'Reduce' expects 3 arguments (array, lambda, initial), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])
			initialType := a.analyzeExpression(expr.Arguments[2])

			// Return type is the same as initial value type
			return initialType
		}

		if strings.EqualFold(funcIdent.Value, "ForEach") {
			// ForEach(array, lambda) -> void
			if len(expr.Arguments) != 2 {
				a.addError("function 'ForEach' expects 2 arguments (array, lambda), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VOID
			}
			a.analyzeExpression(expr.Arguments[0])
			a.analyzeExpression(expr.Arguments[1])

			return types.VOID
		}

		// Task 9.95-9.97: Current date/time functions
		if strings.EqualFold(funcIdent.Value, "Now") || strings.EqualFold(funcIdent.Value, "Date") ||
			strings.EqualFold(funcIdent.Value, "Time") || strings.EqualFold(funcIdent.Value, "UTCDateTime") ||
			strings.EqualFold(funcIdent.Value, "UnixTime") || strings.EqualFold(funcIdent.Value, "UnixTimeMSec") {
			if len(expr.Arguments) != 0 {
				a.addError("function '%s' expects 0 arguments, got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			// Now, Date, Time, UTCDateTime return Float (TDateTime)
			// UnixTime, UnixTimeMSec return Integer
			if strings.EqualFold(funcIdent.Value, "UnixTime") || strings.EqualFold(funcIdent.Value, "UnixTimeMSec") {
				return types.INTEGER
			}
			return types.FLOAT
		}

		// Task 9.99-9.101: Date encoding functions
		if strings.EqualFold(funcIdent.Value, "EncodeDate") {
			if len(expr.Arguments) != 3 {
				a.addError("function 'EncodeDate' expects 3 arguments (year, month, day), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && argType != types.INTEGER {
					a.addError("function 'EncodeDate' expects Integer as argument %d, got %s at %s",
						i+1, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		if strings.EqualFold(funcIdent.Value, "EncodeTime") {
			if len(expr.Arguments) != 4 {
				a.addError("function 'EncodeTime' expects 4 arguments (hour, minute, second, msec), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && argType != types.INTEGER {
					a.addError("function 'EncodeTime' expects Integer as argument %d, got %s at %s",
						i+1, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		if strings.EqualFold(funcIdent.Value, "EncodeDateTime") {
			if len(expr.Arguments) != 7 {
				a.addError("function 'EncodeDateTime' expects 7 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && argType != types.INTEGER {
					a.addError("function 'EncodeDateTime' expects Integer as argument %d, got %s at %s",
						i+1, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		// Task 9.103-9.104: Date decoding functions (var parameters)
		if strings.EqualFold(funcIdent.Value, "DecodeDate") {
			if len(expr.Arguments) != 4 {
				a.addError("function 'DecodeDate' expects 4 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// First argument: TDateTime (Float)
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function 'DecodeDate' expects Float/TDateTime as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Other arguments are var parameters (year, month, day) - just analyze them
			for i := 1; i < len(expr.Arguments); i++ {
				a.analyzeExpression(expr.Arguments[i])
			}
			return types.VOID
		}

		if strings.EqualFold(funcIdent.Value, "DecodeTime") {
			if len(expr.Arguments) != 5 {
				a.addError("function 'DecodeTime' expects 5 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// First argument: TDateTime (Float)
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function 'DecodeTime' expects Float/TDateTime as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			// Other arguments are var parameters (hour, minute, second, msec) - just analyze them
			for i := 1; i < len(expr.Arguments); i++ {
				a.analyzeExpression(expr.Arguments[i])
			}
			return types.VOID
		}

		// Task 9.105: Component extraction functions
		if strings.EqualFold(funcIdent.Value, "YearOf") || strings.EqualFold(funcIdent.Value, "MonthOf") ||
			strings.EqualFold(funcIdent.Value, "DayOf") || strings.EqualFold(funcIdent.Value, "HourOf") ||
			strings.EqualFold(funcIdent.Value, "MinuteOf") || strings.EqualFold(funcIdent.Value, "SecondOf") ||
			strings.EqualFold(funcIdent.Value, "DayOfWeek") || strings.EqualFold(funcIdent.Value, "DayOfTheWeek") ||
			strings.EqualFold(funcIdent.Value, "DayOfYear") || strings.EqualFold(funcIdent.Value, "WeekNumber") ||
			strings.EqualFold(funcIdent.Value, "YearOfWeek") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (TDateTime), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.INTEGER
		}

		// Task 9.107-9.109: Formatting functions
		if strings.EqualFold(funcIdent.Value, "FormatDateTime") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'FormatDateTime' expects 2 arguments (format, dt), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.STRING {
					a.addError("function 'FormatDateTime' expects String as first argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			if len(expr.Arguments) > 1 {
				argType := a.analyzeExpression(expr.Arguments[1])
				if argType != nil && argType != types.FLOAT {
					a.addError("function 'FormatDateTime' expects Float/TDateTime as second argument, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		if strings.EqualFold(funcIdent.Value, "DateTimeToStr") || strings.EqualFold(funcIdent.Value, "DateToStr") ||
			strings.EqualFold(funcIdent.Value, "TimeToStr") || strings.EqualFold(funcIdent.Value, "DateToISO8601") ||
			strings.EqualFold(funcIdent.Value, "DateTimeToISO8601") || strings.EqualFold(funcIdent.Value, "DateTimeToRFC822") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (TDateTime), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.STRING
		}

		// Task 9.110-9.111: Parsing functions
		if strings.EqualFold(funcIdent.Value, "StrToDate") || strings.EqualFold(funcIdent.Value, "StrToDateTime") ||
			strings.EqualFold(funcIdent.Value, "StrToTime") || strings.EqualFold(funcIdent.Value, "ISO8601ToDateTime") ||
			strings.EqualFold(funcIdent.Value, "RFC822ToDateTime") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (String), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.STRING {
					a.addError("function '%s' expects String, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		// Task 9.113: Incrementing functions
		if strings.EqualFold(funcIdent.Value, "IncYear") || strings.EqualFold(funcIdent.Value, "IncMonth") ||
			strings.EqualFold(funcIdent.Value, "IncDay") || strings.EqualFold(funcIdent.Value, "IncHour") ||
			strings.EqualFold(funcIdent.Value, "IncMinute") || strings.EqualFold(funcIdent.Value, "IncSecond") {
			if len(expr.Arguments) != 2 {
				a.addError("function '%s' expects 2 arguments (dt, amount), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime as first argument, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			if len(expr.Arguments) > 1 {
				argType := a.analyzeExpression(expr.Arguments[1])
				if argType != nil && argType != types.INTEGER {
					a.addError("function '%s' expects Integer as second argument, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		// Task 9.114: Date difference functions
		if strings.EqualFold(funcIdent.Value, "DaysBetween") || strings.EqualFold(funcIdent.Value, "HoursBetween") ||
			strings.EqualFold(funcIdent.Value, "MinutesBetween") || strings.EqualFold(funcIdent.Value, "SecondsBetween") {
			if len(expr.Arguments) != 2 {
				a.addError("function '%s' expects 2 arguments (dt1, dt2), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			for i, arg := range expr.Arguments {
				argType := a.analyzeExpression(arg)
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime as argument %d, got %s at %s",
						funcIdent.Value, i+1, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.INTEGER
		}

		// Special date functions
		if strings.EqualFold(funcIdent.Value, "IsLeapYear") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'IsLeapYear' expects 1 argument (year), got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.INTEGER {
					a.addError("function 'IsLeapYear' expects Integer, got %s at %s",
						argType.String(), expr.Token.Pos.String())
				}
			}
			return types.BOOLEAN
		}

		if strings.EqualFold(funcIdent.Value, "FirstDayOfYear") || strings.EqualFold(funcIdent.Value, "FirstDayOfNextYear") ||
			strings.EqualFold(funcIdent.Value, "FirstDayOfMonth") || strings.EqualFold(funcIdent.Value, "FirstDayOfNextMonth") ||
			strings.EqualFold(funcIdent.Value, "FirstDayOfWeek") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (TDateTime), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		// Unix time conversion functions
		if strings.EqualFold(funcIdent.Value, "UnixTimeToDateTime") || strings.EqualFold(funcIdent.Value, "UnixTimeMSecToDateTime") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (unixTime), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.INTEGER {
					a.addError("function '%s' expects Integer, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.FLOAT
		}

		if strings.EqualFold(funcIdent.Value, "DateTimeToUnixTime") || strings.EqualFold(funcIdent.Value, "DateTimeToUnixTimeMSec") {
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument (TDateTime), got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
			}
			if len(expr.Arguments) > 0 {
				argType := a.analyzeExpression(expr.Arguments[0])
				if argType != nil && argType != types.FLOAT {
					a.addError("function '%s' expects Float/TDateTime, got %s at %s",
						funcIdent.Value, argType.String(), expr.Token.Pos.String())
				}
			}
			return types.INTEGER
		}

		// Allow calling methods within the current class without explicit Self
		if a.currentClass != nil {
			if methodType, found := a.currentClass.GetMethod(funcIdent.Value); found {
				if len(expr.Arguments) != len(methodType.Parameters) {
					a.addError("method '%s' expects %d arguments, got %d at %s",
						funcIdent.Value, len(methodType.Parameters), len(expr.Arguments), expr.Token.Pos.String())
					return methodType.ReturnType
				}
				for i, arg := range expr.Arguments {
					argType := a.analyzeExpression(arg)
					expectedType := methodType.Parameters[i]
					if argType != nil && !a.canAssign(argType, expectedType) {
						a.addError("argument %d to method '%s' has type %s, expected %s at %s",
							i+1, funcIdent.Value, argType.String(), expectedType.String(), expr.Token.Pos.String())
					}
				}
				return methodType.ReturnType
			}
		}

		// Task 9.232: Variant introspection functions
		if strings.EqualFold(funcIdent.Value, "VarType") {
			// VarType takes one Variant argument and returns an integer type code
			if len(expr.Arguments) != 1 {
				a.addError("function 'VarType' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze the argument (can be Variant or any type)
			a.analyzeExpression(expr.Arguments[0])
			return types.INTEGER
		}

		if strings.EqualFold(funcIdent.Value, "VarIsNull") || strings.EqualFold(funcIdent.Value, "VarIsEmpty") {
			// VarIsNull/VarIsEmpty take one Variant argument and return a boolean
			if len(expr.Arguments) != 1 {
				a.addError("function '%s' expects 1 argument, got %d at %s",
					funcIdent.Value, len(expr.Arguments), expr.Token.Pos.String())
				return types.BOOLEAN
			}
			// Analyze the argument (can be Variant or any type)
			a.analyzeExpression(expr.Arguments[0])
			return types.BOOLEAN
		}

		if strings.EqualFold(funcIdent.Value, "VarIsNumeric") {
			// VarIsNumeric takes one Variant argument and returns a boolean
			if len(expr.Arguments) != 1 {
				a.addError("function 'VarIsNumeric' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.BOOLEAN
			}
			// Analyze the argument (can be Variant or any type)
			a.analyzeExpression(expr.Arguments[0])
			return types.BOOLEAN
		}

		// Task 9.233: Variant conversion functions
		if strings.EqualFold(funcIdent.Value, "VarToStr") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'VarToStr' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			a.analyzeExpression(expr.Arguments[0])
			return types.STRING
		}

		if strings.EqualFold(funcIdent.Value, "VarToInt") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'VarToInt' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			a.analyzeExpression(expr.Arguments[0])
			return types.INTEGER
		}

		if strings.EqualFold(funcIdent.Value, "VarToFloat") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'VarToFloat' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.FLOAT
			}
			a.analyzeExpression(expr.Arguments[0])
			return types.FLOAT
		}

		if strings.EqualFold(funcIdent.Value, "VarAsType") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'VarAsType' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VARIANT
			}
			a.analyzeExpression(expr.Arguments[0])
			argType := a.analyzeExpression(expr.Arguments[1])
			// Second argument should be Integer (type code)
			if argType != nil && !argType.Equals(types.INTEGER) {
				a.addError("VarAsType type code must be Integer, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			return types.VARIANT
		}

		// Task 9.91: ParseJSON built-in function
		if strings.EqualFold(funcIdent.Value, "ParseJSON") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'ParseJSON' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.VARIANT
			}
			// Analyze the argument (should be a string)
			argType := a.analyzeExpression(expr.Arguments[0])
			if argType != nil && !argType.Equals(types.STRING) {
				a.addError("ParseJSON expects String argument, got %s at %s",
					argType.String(), expr.Token.Pos.String())
			}
			// Returns Variant containing a JSONValue
			return types.VARIANT
		}

		// Task 9.94: ToJSON built-in function
		if strings.EqualFold(funcIdent.Value, "ToJSON") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'ToJSON' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze the argument (can be any type)
			a.analyzeExpression(expr.Arguments[0])
			// Returns String
			return types.STRING
		}

		// Task 9.95: ToJSONFormatted built-in function
		if strings.EqualFold(funcIdent.Value, "ToJSONFormatted") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'ToJSONFormatted' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.STRING
			}
			// Analyze first argument (can be any type)
			a.analyzeExpression(expr.Arguments[0])
			// Analyze second argument (should be Integer)
			indentType := a.analyzeExpression(expr.Arguments[1])
			if indentType != nil && !indentType.Equals(types.INTEGER) {
				a.addError("ToJSONFormatted expects Integer as second argument, got %s at %s",
					indentType.String(), expr.Token.Pos.String())
			}
			// Returns String
			return types.STRING
		}

		// Task 9.98: JSONHasField built-in function
		if strings.EqualFold(funcIdent.Value, "JSONHasField") {
			if len(expr.Arguments) != 2 {
				a.addError("function 'JSONHasField' expects 2 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.BOOLEAN
			}
			// Analyze both arguments (object, field name)
			a.analyzeExpression(expr.Arguments[0])
			fieldType := a.analyzeExpression(expr.Arguments[1])
			// Second argument must be String
			if fieldType != nil && !fieldType.Equals(types.STRING) {
				a.addError("JSONHasField expects String as second argument, got %s at %s",
					fieldType.String(), expr.Token.Pos.String())
			}
			// Returns Boolean
			return types.BOOLEAN
		}

		// Task 9.99: JSONKeys built-in function
		if strings.EqualFold(funcIdent.Value, "JSONKeys") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'JSONKeys' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.NewDynamicArrayType(types.STRING)
			}
			// Analyze argument
			a.analyzeExpression(expr.Arguments[0])
			// Returns array of String
			return types.NewDynamicArrayType(types.STRING)
		}

		// Task 9.100: JSONValues built-in function
		if strings.EqualFold(funcIdent.Value, "JSONValues") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'JSONValues' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.NewDynamicArrayType(types.VARIANT)
			}
			// Analyze argument
			a.analyzeExpression(expr.Arguments[0])
			// Returns array of Variant
			return types.NewDynamicArrayType(types.VARIANT)
		}

		// Task 9.102: JSONLength built-in function
		if strings.EqualFold(funcIdent.Value, "JSONLength") {
			if len(expr.Arguments) != 1 {
				a.addError("function 'JSONLength' expects 1 argument, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
				return types.INTEGER
			}
			// Analyze argument
			a.analyzeExpression(expr.Arguments[0])
			// Returns Integer
			return types.INTEGER
		}

		// Task 9.114: GetStackTrace() built-in function
		if strings.EqualFold(funcIdent.Value, "GetStackTrace") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetStackTrace' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns String
			return types.STRING
		}

		// Task 9.116: GetCallStack() built-in function
		if strings.EqualFold(funcIdent.Value, "GetCallStack") {
			if len(expr.Arguments) != 0 {
				a.addError("function 'GetCallStack' expects 0 arguments, got %d at %s",
					len(expr.Arguments), expr.Token.Pos.String())
			}
			// Returns dynamic array of records
			// Each record has: FunctionName: String, FileName: String, Line: Integer, Column: Integer
			// For simplicity in semantic analysis, we return a generic dynamic array type
			return types.NewDynamicArrayType(types.VARIANT)
		}

		a.addError("undefined function '%s' at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Task 9.162: Check if it's a function pointer type first
	if funcPtrType := a.analyzeFunctionPointerCall(expr, sym.Type); funcPtrType != nil {
		return funcPtrType
	}

	// Check that symbol is a function
	funcType, ok := sym.Type.(*types.FunctionType)
	if !ok {
		a.addError("'%s' is not a function at %s", funcIdent.Value, expr.Token.Pos.String())
		return nil
	}

	// Task 9.1: Check argument count with optional parameters support
	// Count required parameters (those without defaults)
	requiredParams := 0
	for _, defaultVal := range funcType.DefaultValues {
		if defaultVal == nil {
			requiredParams++
		}
	}

	// Check argument count is within valid range
	if len(expr.Arguments) < requiredParams {
		// Use more precise error message based on whether function has optional parameters
		if requiredParams == len(funcType.Parameters) {
			// All parameters are required
			a.addError("function '%s' expects %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		} else {
			// Function has optional parameters
			a.addError("function '%s' expects at least %d arguments, got %d at %s",
				funcIdent.Value, requiredParams, len(expr.Arguments),
				expr.Token.Pos.String())
		}
		return nil
	}
	if len(expr.Arguments) > len(funcType.Parameters) {
		a.addError("function '%s' expects at most %d arguments, got %d at %s",
			funcIdent.Value, len(funcType.Parameters), len(expr.Arguments),
			expr.Token.Pos.String())
		return nil
	}

	// Check argument types
	// Task 9.137: Handle lazy parameters - validate expression type without evaluating
	// Task 9.2b: Handle var parameters - validate that argument is an lvalue
	for i, arg := range expr.Arguments {
		expectedType := funcType.Parameters[i]

		// Check if this parameter is lazy
		isLazy := len(funcType.LazyParams) > i && funcType.LazyParams[i]

		// Check if this parameter is var (by-reference)
		isVar := len(funcType.VarParams) > i && funcType.VarParams[i]

		// Task 9.2b: Validate var parameter receives an lvalue
		if isVar && !a.isLValue(arg) {
			a.addError("var parameter %d to function '%s' requires a variable (identifier, array element, or field), got %s at %s",
				i+1, funcIdent.Value, arg.String(), arg.Pos().String())
		}

		if isLazy {
			// For lazy parameters, check expression type but don't evaluate
			// The expression will be passed as-is to the interpreter for deferred evaluation
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("lazy argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		} else {
			// Regular parameter: validate type normally
			argType := a.analyzeExpressionWithExpectedType(arg, expectedType)
			if argType != nil && !a.canAssign(argType, expectedType) {
				a.addError("argument %d to function '%s' has type %s, expected %s at %s",
					i+1, funcIdent.Value, argType.String(), expectedType.String(),
					expr.Token.Pos.String())
			}
		}
	}

	return funcType.ReturnType
}

// analyzeOldExpression analyzes an 'old' expression in a postcondition
// Task 9.143: Return the type of the referenced identifier
func (a *Analyzer) analyzeOldExpression(expr *ast.OldExpression) types.Type {
	if expr.Identifier == nil {
		return nil
	}

	// Look up the identifier in the symbol table
	sym, ok := a.symbols.Resolve(expr.Identifier.Value)
	if !ok {
		// Error already reported in validateOldExpressions
		return nil
	}

	// Return the type of the identifier
	return sym.Type
}
