package semantic

import (
	"github.com/cwbudde/go-dws/internal/ast"
	"github.com/cwbudde/go-dws/internal/types"
)

// ============================================================================
// Date/Time Built-in Function Analysis
// ============================================================================

// analyzeNow analyzes the Now built-in function.
// Now takes no arguments and returns a Float (TDateTime).
func (a *Analyzer) analyzeNow(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Now' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeDate analyzes the Date built-in function.
// Date takes no arguments and returns a Float (TDateTime).
func (a *Analyzer) analyzeDate(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Date' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeTime analyzes the Time built-in function.
// Time takes no arguments and returns a Float (TDateTime).
func (a *Analyzer) analyzeTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'Time' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeUTCDateTime analyzes the UTCDateTime built-in function.
// UTCDateTime takes no arguments and returns a Float (TDateTime).
func (a *Analyzer) analyzeUTCDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'UTCDateTime' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.FLOAT
}

// analyzeUnixTime analyzes the UnixTime built-in function.
// UnixTime takes no arguments and returns an Integer.
func (a *Analyzer) analyzeUnixTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'UnixTime' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeUnixTimeMSec analyzes the UnixTimeMSec built-in function.
// UnixTimeMSec takes no arguments and returns an Integer.
func (a *Analyzer) analyzeUnixTimeMSec(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 0 {
		a.addError("function 'UnixTimeMSec' expects 0 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	return types.INTEGER
}

// analyzeEncodeDate analyzes the EncodeDate built-in function.
// EncodeDate takes 3 arguments (year, month, day) and returns a Float (TDateTime).
func (a *Analyzer) analyzeEncodeDate(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 3 {
		a.addError("function 'EncodeDate' expects 3 arguments (year, month, day), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'EncodeDate' expects Integer as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeEncodeTime analyzes the EncodeTime built-in function.
// EncodeTime takes 4 arguments (hour, minute, second, msec) and returns a Float (TDateTime).
func (a *Analyzer) analyzeEncodeTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'EncodeTime' expects 4 arguments (hour, minute, second, msec), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'EncodeTime' expects Integer as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeEncodeDateTime analyzes the EncodeDateTime built-in function.
// EncodeDateTime takes 7 arguments and returns a Float (TDateTime).
func (a *Analyzer) analyzeEncodeDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 7 {
		a.addError("function 'EncodeDateTime' expects 7 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'EncodeDateTime' expects Integer as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeDecodeDate analyzes the DecodeDate built-in procedure.
// DecodeDate takes 4 arguments (dt, var year, var month, var day) and returns void.
func (a *Analyzer) analyzeDecodeDate(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 4 {
		a.addError("function 'DecodeDate' expects 4 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// First argument: TDateTime (Float)
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DecodeDate' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Other arguments are var parameters (year, month, day) - just analyze them
	for i := 1; i < len(args); i++ {
		a.analyzeExpression(args[i])
	}
	return types.VOID
}

// analyzeDecodeTime analyzes the DecodeTime built-in procedure.
// DecodeTime takes 5 arguments (dt, var hour, var minute, var second, var msec) and returns void.
func (a *Analyzer) analyzeDecodeTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 5 {
		a.addError("function 'DecodeTime' expects 5 arguments, got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	// First argument: TDateTime (Float)
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DecodeTime' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	// Other arguments are var parameters (hour, minute, second, msec) - just analyze them
	for i := 1; i < len(args); i++ {
		a.analyzeExpression(args[i])
	}
	return types.VOID
}

// analyzeYearOf analyzes the YearOf built-in function.
// YearOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeYearOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'YearOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'YearOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeMonthOf analyzes the MonthOf built-in function.
// MonthOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeMonthOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'MonthOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'MonthOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDayOf analyzes the DayOf built-in function.
// DayOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDayOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DayOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DayOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeHourOf analyzes the HourOf built-in function.
// HourOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeHourOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'HourOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'HourOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeMinuteOf analyzes the MinuteOf built-in function.
// MinuteOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeMinuteOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'MinuteOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'MinuteOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeSecondOf analyzes the SecondOf built-in function.
// SecondOf takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeSecondOf(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'SecondOf' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'SecondOf' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDayOfWeek analyzes the DayOfWeek built-in function.
// DayOfWeek takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDayOfWeek(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DayOfWeek' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DayOfWeek' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDayOfTheWeek analyzes the DayOfTheWeek built-in function.
// DayOfTheWeek takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDayOfTheWeek(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DayOfTheWeek' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DayOfTheWeek' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDayOfYear analyzes the DayOfYear built-in function.
// DayOfYear takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDayOfYear(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DayOfYear' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DayOfYear' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeWeekNumber analyzes the WeekNumber built-in function.
// WeekNumber takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeWeekNumber(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'WeekNumber' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'WeekNumber' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeYearOfWeek analyzes the YearOfWeek built-in function.
// YearOfWeek takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeYearOfWeek(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'YearOfWeek' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'YearOfWeek' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeFormatDateTime analyzes the FormatDateTime built-in function.
// FormatDateTime takes 2 arguments (format, dt) and returns a String.
func (a *Analyzer) analyzeFormatDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'FormatDateTime' expects 2 arguments (format, dt), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'FormatDateTime' expects String as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FormatDateTime' expects Float/TDateTime as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDateTimeToStr analyzes the DateTimeToStr built-in function.
// DateTimeToStr takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeDateTimeToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateTimeToStr' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateTimeToStr' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDateToStr analyzes the DateToStr built-in function.
// DateToStr takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeDateToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateToStr' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateToStr' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeTimeToStr analyzes the TimeToStr built-in function.
// TimeToStr takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeTimeToStr(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'TimeToStr' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'TimeToStr' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDateToISO8601 analyzes the DateToISO8601 built-in function.
// DateToISO8601 takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeDateToISO8601(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateToISO8601' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateToISO8601' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDateTimeToISO8601 analyzes the DateTimeToISO8601 built-in function.
// DateTimeToISO8601 takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeDateTimeToISO8601(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateTimeToISO8601' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateTimeToISO8601' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeDateTimeToRFC822 analyzes the DateTimeToRFC822 built-in function.
// DateTimeToRFC822 takes 1 argument (TDateTime) and returns a String.
func (a *Analyzer) analyzeDateTimeToRFC822(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateTimeToRFC822' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateTimeToRFC822' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.STRING
}

// analyzeStrToDate analyzes the StrToDate built-in function.
// StrToDate takes 1 argument (String) and returns a Float (TDateTime).
func (a *Analyzer) analyzeStrToDate(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToDate' expects 1 argument (String), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'StrToDate' expects String, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeStrToDateTime analyzes the StrToDateTime built-in function.
// StrToDateTime takes 1 argument (String) and returns a Float (TDateTime).
func (a *Analyzer) analyzeStrToDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToDateTime' expects 1 argument (String), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'StrToDateTime' expects String, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeStrToTime analyzes the StrToTime built-in function.
// StrToTime takes 1 argument (String) and returns a Float (TDateTime).
func (a *Analyzer) analyzeStrToTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'StrToTime' expects 1 argument (String), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'StrToTime' expects String, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeISO8601ToDateTime analyzes the ISO8601ToDateTime built-in function.
// ISO8601ToDateTime takes 1 argument (String) and returns a Float (TDateTime).
func (a *Analyzer) analyzeISO8601ToDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'ISO8601ToDateTime' expects 1 argument (String), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'ISO8601ToDateTime' expects String, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeRFC822ToDateTime analyzes the RFC822ToDateTime built-in function.
// RFC822ToDateTime takes 1 argument (String) and returns a Float (TDateTime).
func (a *Analyzer) analyzeRFC822ToDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'RFC822ToDateTime' expects 1 argument (String), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.STRING {
			a.addError("function 'RFC822ToDateTime' expects String, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncYear analyzes the IncYear built-in function.
// IncYear takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncYear(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncYear' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncYear' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncYear' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncMonth analyzes the IncMonth built-in function.
// IncMonth takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncMonth(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncMonth' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncMonth' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncMonth' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncDay analyzes the IncDay built-in function.
// IncDay takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncDay(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncDay' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncDay' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncDay' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncHour analyzes the IncHour built-in function.
// IncHour takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncHour(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncHour' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncHour' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncHour' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncMinute analyzes the IncMinute built-in function.
// IncMinute takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncMinute(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncMinute' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncMinute' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncMinute' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeIncSecond analyzes the IncSecond built-in function.
// IncSecond takes 2 arguments (dt, amount) and returns a Float (TDateTime).
func (a *Analyzer) analyzeIncSecond(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'IncSecond' expects 2 arguments (dt, amount), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'IncSecond' expects Float/TDateTime as first argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	if len(args) > 1 {
		argType := a.analyzeExpression(args[1])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IncSecond' expects Integer as second argument, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeDaysBetween analyzes the DaysBetween built-in function.
// DaysBetween takes 2 arguments (dt1, dt2) and returns an Integer.
func (a *Analyzer) analyzeDaysBetween(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'DaysBetween' expects 2 arguments (dt1, dt2), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DaysBetween' expects Float/TDateTime as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeHoursBetween analyzes the HoursBetween built-in function.
// HoursBetween takes 2 arguments (dt1, dt2) and returns an Integer.
func (a *Analyzer) analyzeHoursBetween(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'HoursBetween' expects 2 arguments (dt1, dt2), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'HoursBetween' expects Float/TDateTime as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeMinutesBetween analyzes the MinutesBetween built-in function.
// MinutesBetween takes 2 arguments (dt1, dt2) and returns an Integer.
func (a *Analyzer) analyzeMinutesBetween(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'MinutesBetween' expects 2 arguments (dt1, dt2), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'MinutesBetween' expects Float/TDateTime as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeSecondsBetween analyzes the SecondsBetween built-in function.
// SecondsBetween takes 2 arguments (dt1, dt2) and returns an Integer.
func (a *Analyzer) analyzeSecondsBetween(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 2 {
		a.addError("function 'SecondsBetween' expects 2 arguments (dt1, dt2), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	for i, arg := range args {
		argType := a.analyzeExpression(arg)
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'SecondsBetween' expects Float/TDateTime as argument %d, got %s at %s",
				i+1, argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeIsLeapYear analyzes the IsLeapYear built-in function.
// IsLeapYear takes 1 argument (year) and returns a Boolean.
func (a *Analyzer) analyzeIsLeapYear(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'IsLeapYear' expects 1 argument (year), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'IsLeapYear' expects Integer, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.BOOLEAN
}

// analyzeFirstDayOfYear analyzes the FirstDayOfYear built-in function.
// FirstDayOfYear takes 1 argument (TDateTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeFirstDayOfYear(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FirstDayOfYear' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FirstDayOfYear' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeFirstDayOfNextYear analyzes the FirstDayOfNextYear built-in function.
// FirstDayOfNextYear takes 1 argument (TDateTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeFirstDayOfNextYear(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FirstDayOfNextYear' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FirstDayOfNextYear' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeFirstDayOfMonth analyzes the FirstDayOfMonth built-in function.
// FirstDayOfMonth takes 1 argument (TDateTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeFirstDayOfMonth(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FirstDayOfMonth' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FirstDayOfMonth' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeFirstDayOfNextMonth analyzes the FirstDayOfNextMonth built-in function.
// FirstDayOfNextMonth takes 1 argument (TDateTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeFirstDayOfNextMonth(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FirstDayOfNextMonth' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FirstDayOfNextMonth' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeFirstDayOfWeek analyzes the FirstDayOfWeek built-in function.
// FirstDayOfWeek takes 1 argument (TDateTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeFirstDayOfWeek(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'FirstDayOfWeek' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'FirstDayOfWeek' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeUnixTimeToDateTime analyzes the UnixTimeToDateTime built-in function.
// UnixTimeToDateTime takes 1 argument (unixTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeUnixTimeToDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'UnixTimeToDateTime' expects 1 argument (unixTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'UnixTimeToDateTime' expects Integer, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeUnixTimeMSecToDateTime analyzes the UnixTimeMSecToDateTime built-in function.
// UnixTimeMSecToDateTime takes 1 argument (unixTime) and returns a Float (TDateTime).
func (a *Analyzer) analyzeUnixTimeMSecToDateTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'UnixTimeMSecToDateTime' expects 1 argument (unixTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.INTEGER {
			a.addError("function 'UnixTimeMSecToDateTime' expects Integer, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.FLOAT
}

// analyzeDateTimeToUnixTime analyzes the DateTimeToUnixTime built-in function.
// DateTimeToUnixTime takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDateTimeToUnixTime(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateTimeToUnixTime' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateTimeToUnixTime' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}

// analyzeDateTimeToUnixTimeMSec analyzes the DateTimeToUnixTimeMSec built-in function.
// DateTimeToUnixTimeMSec takes 1 argument (TDateTime) and returns an Integer.
func (a *Analyzer) analyzeDateTimeToUnixTimeMSec(args []ast.Expression, callExpr *ast.CallExpression) types.Type {
	if len(args) != 1 {
		a.addError("function 'DateTimeToUnixTimeMSec' expects 1 argument (TDateTime), got %d at %s",
			len(args), callExpr.Token.Pos.String())
	}
	if len(args) > 0 {
		argType := a.analyzeExpression(args[0])
		if argType != nil && argType != types.FLOAT {
			a.addError("function 'DateTimeToUnixTimeMSec' expects Float/TDateTime, got %s at %s",
				argType.String(), callExpr.Token.Pos.String())
		}
	}
	return types.INTEGER
}
