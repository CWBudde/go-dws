package evaluator

import (
	"time"
)

// ============================================================================
// DateTime Helper Functions for DecodeDate/DecodeTime
// ============================================================================
//
// These functions are needed by the DecodeDate/DecodeTime built-ins.
// They convert Delphi TDateTime format (float64) to Go time.Time and extract components.

// DelphiEpoch is the reference date for TDateTime calculations.
// Delphi's TDateTime uses December 30, 1899 as day 0.
var delphiEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)

// Constants for time calculations
const secondsPerDay = 86400.0

// delphiDateTimeToGoTime converts a Delphi TDateTime float64 to Go time.Time.
// TDateTime is a float64 where:
//   - Integer part = number of days since December 30, 1899
//   - Fractional part = time of day (0.5 = noon, 0.25 = 6am)
func delphiDateTimeToGoTime(dt float64) time.Time {
	seconds := dt * secondsPerDay
	duration := time.Duration(seconds * float64(time.Second))
	return delphiEpoch.Add(duration)
}

// extractDateComponents extracts year, month, day from a TDateTime value.
func extractDateComponents(dt float64) (year, month, day int) {
	t := delphiDateTimeToGoTime(dt)
	return t.Year(), int(t.Month()), t.Day()
}

// extractTimeComponents extracts hour, minute, second, millisecond from a TDateTime value.
func extractTimeComponents(dt float64) (hour, minute, second, millisecond int) {
	t := delphiDateTimeToGoTime(dt)
	return t.Hour(), t.Minute(), t.Second(), t.Nanosecond() / 1000000
}
