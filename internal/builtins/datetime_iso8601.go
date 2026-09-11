package builtins

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// =============================================================================
// ISO 8601
// =============================================================================

// ISO8601Precision selects how much of the time component DateTimeToISO8601
// emits.
type ISO8601Precision int

const (
	// ISO8601PrecAuto omits seconds when they are zero.
	ISO8601PrecAuto ISO8601Precision = iota
	// ISO8601PrecSeconds always emits seconds.
	ISO8601PrecSeconds
	// ISO8601PrecMilliseconds always emits seconds and milliseconds.
	ISO8601PrecMilliseconds
)

// parseISO8601Precision maps the DateTimeToISO8601 precision argument onto an
// ISO8601Precision. Unknown values fall back to the automatic precision, as in
// the original implementation.
func parseISO8601Precision(prec string) ISO8601Precision {
	switch prec {
	case "sec":
		return ISO8601PrecSeconds
	case "msec":
		return ISO8601PrecMilliseconds
	default:
		return ISO8601PrecAuto
	}
}

// formatISO8601 renders a TDateTime in extended ISO 8601 form with a trailing
// "Z", matching DWScript's DateTimeToISO8601(dt, True, prec).
func formatISO8601(dt float64, prec ISO8601Precision) string {
	c := decodeDateTime(dt)
	var b strings.Builder
	fmt.Fprintf(&b, "%04d-%02d-%02dT%02d:%02d", c.Year, c.Month, c.Day, c.Hour, c.Minute)
	if c.Second != 0 || prec == ISO8601PrecSeconds || prec == ISO8601PrecMilliseconds {
		fmt.Fprintf(&b, ":%02d", c.Second)
	}
	if prec == ISO8601PrecMilliseconds {
		fmt.Fprintf(&b, ".%03d", c.Millisecond)
	}
	b.WriteByte('Z')
	return b.String()
}

// formatDateISO8601 renders only the date part of a TDateTime.
func formatDateISO8601(dt float64) string {
	c := decodeDateTime(dt)
	return fmt.Sprintf("%04d-%02d-%02d", c.Year, c.Month, c.Day)
}

// iso8601Error is the error type raised by the ISO 8601 parser. Its message is
// reported verbatim to scripts, so the wording must match DWScript's.
type iso8601Error struct {
	msg string
}

func (e *iso8601Error) Error() string { return e.msg }

type iso8601Parser struct {
	s []rune
	p int
}

// at returns the rune at the cursor, or 0 at end of input, mirroring the
// null-terminated scanning of the original.
func (r *iso8601Parser) at() rune {
	if r.p >= len(r.s) {
		return 0
	}
	return r.s[r.p]
}

func (r *iso8601Parser) readDigit() (int, error) {
	c := r.at()
	if c < '0' || c > '9' {
		return 0, &iso8601Error{fmt.Sprintf("Unexpected character (%d) instead of digit", c)}
	}
	r.p++
	return int(c - '0'), nil
}

func (r *iso8601Parser) read2Digits() (int, error) {
	hi, err := r.readDigit()
	if err != nil {
		return 0, err
	}
	lo, err := r.readDigit()
	if err != nil {
		return 0, err
	}
	return hi*10 + lo, nil
}

func (r *iso8601Parser) read4Digits() (int, error) {
	hi, err := r.read2Digits()
	if err != nil {
		return 0, err
	}
	lo, err := r.read2Digits()
	if err != nil {
		return 0, err
	}
	return hi*100 + lo, nil
}

// readUpTo6Digits reads a fractional-seconds run and scales it to milliseconds.
func (r *iso8601Parser) readUpTo6Digits() (int, error) {
	nbDigits := 1
	value, err := r.readDigit()
	if err != nil {
		return 0, err
	}
	for nbDigits < 6 {
		c := r.at()
		if c < '0' || c > '9' {
			break
		}
		digit, err := r.readDigit()
		if err != nil {
			return 0, err
		}
		nbDigits++
		value = value*10 + digit
	}
	switch nbDigits {
	case 1:
		value *= 100
	case 2:
		value *= 10
	case 4:
		value /= 10
	case 5:
		value /= 100
	case 6:
		value /= 1000
	}
	return value, nil
}

// parseDateTimeISO8601 returns the parsed TDateTime and the UTC offset as a
// day fraction. An offset of exactly 1 means "no offset was present", which is
// the sentinel the original uses.
//
//nolint:gocyclo // a position-by-position scanner; each branch owns one diagnostic
func parseDateTimeISO8601(v string) (dt, utcOffset float64, err error) {
	utcOffset = 1
	if v == "" {
		return 0, utcOffset, nil
	}
	r := &iso8601Parser{s: []rune(v)}

	year, err := r.read4Digits()
	if err != nil {
		return 0, 0, err
	}
	separator := r.at() == '-'
	if separator {
		r.p++
	}
	month, err := r.read2Digits()
	if err != nil {
		return 0, 0, err
	}
	if separator {
		if r.at() != '-' {
			return 0, 0, &iso8601Error{`"-" expected after month`}
		}
		r.p++
	}
	day, err := r.read2Digits()
	if err != nil {
		return 0, 0, err
	}
	if !isValidDate(year, month, day) {
		return 0, 0, &iso8601Error{"Invalid argument to date encode"}
	}
	dt = encodeDateOnly(year, month, day)

	switch r.at() {
	case 0:
		return dt, utcOffset, nil
	case 'T', 't', ' ':
		r.p++
	default:
		return 0, 0, &iso8601Error{`"T" expected after date`}
	}

	hour, err := r.read2Digits()
	if err != nil {
		return 0, 0, err
	}
	separator = r.at() == ':'
	if separator {
		r.p++
	}
	minute, err := r.read2Digits()
	if err != nil {
		return 0, 0, err
	}
	second, msec := 0, 0
	if c := r.at(); c == ':' || (c >= '0' && c <= '9') {
		if separator {
			if c != ':' {
				return 0, 0, &iso8601Error{`":" expected after minutes`}
			}
			r.p++
		} else if c == ':' {
			return 0, 0, &iso8601Error{`Unexpected ":" after minutes`}
		}
		second, err = r.read2Digits()
		if err != nil {
			return 0, 0, err
		}
		if r.at() == '.' {
			r.p++
			msec, err = r.readUpTo6Digits()
			if err != nil {
				return 0, 0, err
			}
		}
	}
	if !isValidTime(hour, minute, second, msec) {
		return 0, 0, &iso8601Error{"Invalid argument to time encode"}
	}
	dt += encodeTimeOnly(hour, minute, second, msec)

	switch r.at() {
	case 0:
		return dt, utcOffset, nil
	case 'Z', 'z':
		r.p++
	case '+', '-':
		sign := 1.0
		if r.at() == '-' {
			sign = -1
		}
		r.p++
		offsetHours, err := r.read2Digits()
		if err != nil {
			return 0, 0, err
		}
		offsetMinutes := 0
		if c := r.at(); c == ':' || (c >= '0' && c <= '9') {
			if separator {
				if c != ':' {
					return 0, 0, &iso8601Error{`":" expected after offset hours`}
				}
				r.p++
			}
			offsetMinutes, err = r.read2Digits()
			if err != nil {
				return 0, 0, err
			}
		}
		utcOffset = sign * encodeTimeOnly(offsetHours, offsetMinutes, 0, 0)
	}

	if r.at() != 0 {
		return 0, 0, &iso8601Error{"Unsupported or invalid ISO8601 format"}
	}
	return dt, utcOffset, nil
}

// parseISO8601 parses an ISO 8601 date/time into a TDateTime, applying any
// UTC offset present in the string.
func parseISO8601(v string) (float64, error) {
	dt, utcOffset, err := parseDateTimeISO8601(v)
	if err != nil {
		return 0, err
	}
	if utcOffset != 1 {
		dt -= utcOffset
	}
	return dt, nil
}

// =============================================================================
// RFC 822
// =============================================================================

// formatRFC822 renders a TDateTime as an RFC 822 date, labelled GMT. The value
// is used as-is; no timezone conversion is applied.
func formatRFC822(dt float64) string {
	c := decodeDateTime(dt)
	dayNumber, _ := dateTimeSplit(dt)
	dow := int(dateTimeEpoch.AddDate(0, 0, dayNumber).Weekday())
	return fmt.Sprintf("%s, %02d %s %04d %02d:%02d:%02d GMT",
		shortDayNames[dow], c.Day, shortMonthNames[c.Month-1], c.Year, c.Hour, c.Minute, c.Second)
}

// parseRFC822 parses an RFC 822 date. It accepts both the classic
// "Thu, 08 Oct 2009 00:00:00 GMT" form and the JavaScript
// "Fri Feb 11 2022 10:25:55 GMT+0100 (...)" form. Unparsable input yields 0
// rather than an error, matching WebUtils.RFC822ToDateTime.
func parseRFC822(str string) float64 {
	if str == "" {
		return 0
	}
	start := 0
	if idx := strings.IndexByte(str, ','); idx >= 0 {
		start = idx + 1
	}
	const maxItems = 6
	fields := strings.Fields(str[start:])
	if len(fields) > maxItems {
		fields = fields[:maxItems]
	}
	if len(fields) < 5 {
		return 0
	}

	var year, month, day, hour, minute, second, deltaHours int
	if len(fields) > 5 && strings.Contains(fields[4], ":") && strings.HasPrefix(fields[5], "GMT+") {
		month = rfc822Month(fields[1])
		day = rfc822TwoDigits(fields[2])
		year = rfc822Year(fields[3])
		hour, minute, second = rfc822HMS(fields[4])
		deltaHours = rfc822SignedInt(fields[5][3:])
	} else {
		day = rfc822TwoDigits(fields[0])
		month = rfc822Month(fields[1])
		year = rfc822Year(fields[2])
		hour, minute, second = rfc822HMS(fields[3])
		deltaHours = rfc822SignedInt(fields[4])
	}

	deltaDays := 0
	for hour >= 24 && hour != rfc822InvalidHour {
		hour -= 24
		deltaDays++
	}
	if hour == rfc822InvalidHour || !isValidDate(year, month, day) || !isValidTime(hour, minute, second, 0) {
		return 0
	}
	dt := encodeDateOnly(year, month, day) + encodeTimeOnly(hour, minute, second, 0)
	// deltaHours is the signed HHMM offset, so dividing by 100 hours undoes it.
	return dt - float64(deltaHours)/2400.0 + float64(deltaDays)
}

// rfc822InvalidHour marks an unparsable time-of-day field.
const rfc822InvalidHour = 65535

// rfc822SignedInt parses a numeric zone offset such as "0100" or "-0500".
// Anything unparsable counts as no offset, in line with the rest of the RFC 822
// scanner, which degrades to 0 rather than failing.
func rfc822SignedInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func rfc822TwoDigits(s string) int {
	if len(s) != 2 {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func rfc822Year(s string) int {
	switch len(s) {
	case 2:
		v, err := strconv.Atoi(s)
		if err != nil {
			return rfc822InvalidHour
		}
		return v + 2000
	case 4:
		v, err := strconv.Atoi(s)
		if err != nil {
			return rfc822InvalidHour
		}
		return v
	default:
		return rfc822InvalidHour
	}
}

func rfc822Month(s string) int {
	for i, name := range shortMonthNames {
		if strings.EqualFold(s, name) {
			return i + 1
		}
	}
	return 13
}

func rfc822HMS(s string) (hour, minute, second int) {
	switch len(s) {
	case 5: // hh:nn
		if s[2] != ':' {
			return rfc822InvalidHour, 0, 0
		}
		return rfc822TwoDigits(s[0:2]), rfc822TwoDigits(s[3:5]), 0
	case 8: // hh:nn:ss
		if s[2] != ':' || s[5] != ':' {
			return rfc822InvalidHour, 0, 0
		}
		return rfc822TwoDigits(s[0:2]), rfc822TwoDigits(s[3:5]), rfc822TwoDigits(s[6:8])
	default:
		return rfc822InvalidHour, 0, 0
	}
}

// =============================================================================
// Unix time
// =============================================================================

// unixTimeMSecToDateTime converts a Unix timestamp in milliseconds to a UTC
// TDateTime.
func unixTimeMSecToDateTime(unixTimeMS int64) float64 {
	return float64(unixTimeMS)/millisecondsPerDay + unixEpochDateTime
}

// dateTimeToUnixTime converts a UTC TDateTime to a Unix timestamp in seconds.
func dateTimeToUnixTime(dt float64) int64 {
	return int64(math.Round(dt*secondsPerDay)) - int64(unixEpochDateTime)*86400
}

// dateTimeToUnixTimeMSec converts a UTC TDateTime to a Unix timestamp in
// milliseconds.
func dateTimeToUnixTimeMSec(dt float64) int64 {
	return int64(math.Round(dt*millisecondsPerDay)) - int64(unixEpochDateTime)*86400000
}
