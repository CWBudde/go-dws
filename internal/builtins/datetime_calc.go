package builtins

// =============================================================================
// Date Encoding Functions
// =============================================================================

// EncodeDate implements the EncodeDate() built-in function.
// EncodeDate(year, month, day: Integer; tz: DateTimeZone = Default): TDateTime
func EncodeDate(ctx Context, args []Value) Value {
	if len(args) < 3 || len(args) > 4 {
		return ctx.NewError("EncodeDate() expects 3 or 4 arguments (year, month, day, [zone]), got %d", len(args))
	}
	parts, errVal := integerArgs(ctx, args, 3, "EncodeDate")
	if errVal != nil {
		return errVal
	}
	dt, ok := tryEncodeDate(parts[0], parts[1], parts[2], settings(ctx), zoneArg(ctx, args, 3))
	if !ok {
		return ctx.NewError("EncodeDate() invalid date: %d-%02d-%02d", parts[0], parts[1], parts[2])
	}
	return floatResult(dt)
}

// EncodeTime implements the EncodeTime() built-in function.
// EncodeTime(hour, minute, second, msec: Integer): TDateTime
func EncodeTime(ctx Context, args []Value) Value {
	if len(args) != 4 {
		return ctx.NewError("EncodeTime() expects 4 arguments (hour, minute, second, msec), got %d", len(args))
	}
	parts, errVal := integerArgs(ctx, args, 4, "EncodeTime")
	if errVal != nil {
		return errVal
	}
	if !isValidTime(parts[0], parts[1], parts[2], parts[3]) {
		return ctx.NewError("EncodeTime() invalid time: %02d:%02d:%02d.%03d",
			parts[0], parts[1], parts[2], parts[3])
	}
	return floatResult(encodeTimeOnly(parts[0], parts[1], parts[2], parts[3]))
}

// EncodeDateTime implements the EncodeDateTime() built-in function.
// EncodeDateTime(year, month, day, hour, minute, second, msec: Integer;
//
//	tz: DateTimeZone = Default): TDateTime
func EncodeDateTime(ctx Context, args []Value) Value {
	if len(args) < 7 || len(args) > 8 {
		return ctx.NewError("EncodeDateTime() expects 7 or 8 arguments, got %d", len(args))
	}
	parts, errVal := integerArgs(ctx, args, 7, "EncodeDateTime")
	if errVal != nil {
		return errVal
	}
	dt, ok := tryEncodeDateTime(parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6],
		settings(ctx), zoneArg(ctx, args, 7))
	if !ok {
		return ctx.NewError("EncodeDateTime() invalid date/time")
	}
	return floatResult(dt)
}

// integerArgs reads the first count arguments as Integers.
func integerArgs(ctx Context, args []Value, count int, fn string) ([]int, Value) {
	parts := make([]int, count)
	for i := 0; i < count; i++ {
		value, errVal := integerArg(ctx, args, i, fn)
		if errVal != nil {
			return nil, errVal
		}
		parts[i] = int(value)
	}
	return parts, nil
}

// =============================================================================
// Date Decoding Functions (Var Parameters)
// =============================================================================

// NOTE: DecodeDate and DecodeTime are implemented in internal/interp/builtins_datetime_calc.go
// as var-param functions (taking []ast.Expression). They cannot be migrated to the builtins
// package because they need direct AST access to modify variables in-place.
//
// DecodeDate(dt: TDateTime; var year, month, day: Integer)
// DecodeTime(dt: TDateTime; var hour, minute, second, msec: Integer)
//
// These functions are registered in callBuiltinWithVarParam() in functions_builtins.go.

// =============================================================================
// Incrementing Functions
// =============================================================================

// incrementBy reads the (dt, count) argument pair shared by the Inc* family.
// The count defaults to 1 when omitted.
func incrementBy(ctx Context, args []Value, fn string) (dt float64, count int64, errVal Value) {
	if len(args) < 1 || len(args) > 2 {
		return 0, 0, ctx.NewError("%s() expects 1 or 2 arguments (dt, [count]), got %d", fn, len(args))
	}
	dt, errVal = dateTimeArg(ctx, args, 0, fn)
	if errVal != nil {
		return 0, 0, errVal
	}
	count = 1
	if len(args) == 2 {
		count, errVal = integerArg(ctx, args, 1, fn)
		if errVal != nil {
			return 0, 0, errVal
		}
	}
	return dt, count, nil
}

// IncYear implements the IncYear() built-in function.
func IncYear(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncYear")
	if errVal != nil {
		return errVal
	}
	return floatResult(incMonths(dt, int(count)*12))
}

// IncMonth implements the IncMonth() built-in function.
func IncMonth(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncMonth")
	if errVal != nil {
		return errVal
	}
	return floatResult(incMonths(dt, int(count)))
}

// IncWeek implements the IncWeek() built-in function.
func IncWeek(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncWeek")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + 7*float64(count))
}

// IncDay implements the IncDay() built-in function.
func IncDay(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncDay")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + float64(count))
}

// IncHour implements the IncHour() built-in function.
func IncHour(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncHour")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + float64(count)/24)
}

// IncMinute implements the IncMinute() built-in function.
func IncMinute(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncMinute")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + float64(count)/1440)
}

// IncSecond implements the IncSecond() built-in function.
func IncSecond(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncSecond")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + float64(count)/secondsPerDay)
}

// IncMilliSecond implements the IncMilliSecond() built-in function.
func IncMilliSecond(ctx Context, args []Value) Value {
	dt, count, errVal := incrementBy(ctx, args, "IncMilliSecond")
	if errVal != nil {
		return errVal
	}
	return floatResult(dt + float64(count)/millisecondsPerDay)
}

// =============================================================================
// Date Difference Functions
// =============================================================================

// dateTimePair reads the two TDateTime arguments of the *Between family.
func dateTimePair(ctx Context, args []Value, fn string) (dt1, dt2 float64, errVal Value) {
	if len(args) != 2 {
		return 0, 0, ctx.NewError("%s() expects 2 arguments, got %d", fn, len(args))
	}
	dt1, errVal = dateTimeArg(ctx, args, 0, fn)
	if errVal != nil {
		return 0, 0, errVal
	}
	dt2, errVal = dateTimeArg(ctx, args, 1, fn)
	if errVal != nil {
		return 0, 0, errVal
	}
	return dt1, dt2, nil
}

// DaysBetween implements the DaysBetween() built-in function.
func DaysBetween(ctx Context, args []Value) Value {
	dt1, dt2, errVal := dateTimePair(ctx, args, "DaysBetween")
	if errVal != nil {
		return errVal
	}
	return intResult(daysBetween(dt1, dt2))
}

// HoursBetween implements the HoursBetween() built-in function.
func HoursBetween(ctx Context, args []Value) Value {
	dt1, dt2, errVal := dateTimePair(ctx, args, "HoursBetween")
	if errVal != nil {
		return errVal
	}
	return intResult(hoursBetween(dt1, dt2))
}

// MinutesBetween implements the MinutesBetween() built-in function.
func MinutesBetween(ctx Context, args []Value) Value {
	dt1, dt2, errVal := dateTimePair(ctx, args, "MinutesBetween")
	if errVal != nil {
		return errVal
	}
	return intResult(minutesBetween(dt1, dt2))
}

// SecondsBetween implements the SecondsBetween() built-in function.
func SecondsBetween(ctx Context, args []Value) Value {
	dt1, dt2, errVal := dateTimePair(ctx, args, "SecondsBetween")
	if errVal != nil {
		return errVal
	}
	return intResult(secondsBetween(dt1, dt2))
}
