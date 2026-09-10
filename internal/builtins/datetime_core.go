package builtins

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// =============================================================================
// TDateTime representation
// =============================================================================
//
// A TDateTime is a float64 counting days since 1899-12-30. The integer part is
// the day number and the fractional part is the time of day. Delphi splits a
// negative TDateTime as "day number truncated toward zero, absolute fraction as
// the time of day", so 1899-01-03 12:00 is -361.5 and not -361.5's floor.
//
// The value carries no timezone: it holds wall-clock civil components. Whether
// those components denote local time or UTC is decided by the caller through a
// TimeZone selector, exactly as in the original DWScript implementation
// (Source/dwsDateTime.pas).

const (
	secondsPerDay      = 86400.0
	millisecondsPerDay = 86400000.0

	// unixEpochDateTime is the TDateTime value of 1970-01-01T00:00:00.
	unixEpochDateTime = 25569.0
)

// dateTimeEpoch is the civil origin of TDateTime, used purely as an arithmetic
// device: all conversions go through time.UTC so that no host timezone rule
// can perturb the civil components.
var dateTimeEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)

// TimeZone selects how a TDateTime value is interpreted on input and output.
// It mirrors DWScript's TdwsTimeZone enumeration.
type TimeZone int

const (
	// TimeZoneDefault defers to the current FormatSettings.Zone value.
	TimeZoneDefault TimeZone = 0
	// TimeZoneLocal interprets and renders values as local wall-clock time.
	TimeZoneLocal TimeZone = 1
	// TimeZoneUTC interprets and renders values as UTC.
	TimeZoneUTC TimeZone = 2
)

// dateTimeSplit decomposes a TDateTime into its Delphi day number and the
// millisecond offset within that day. The value is rounded to the nearest
// millisecond first, which is what Delphi's DecodeTime does and what keeps
// EncodeTime(12, 34, 45, 567) from decoding back to 566 milliseconds.
func dateTimeSplit(dt float64) (dayNumber int, msecOfDay int) {
	totalMsec := math.Round(dt * millisecondsPerDay)
	days := math.Trunc(totalMsec / millisecondsPerDay)
	frac := totalMsec - days*millisecondsPerDay
	if frac < 0 {
		frac = -frac
	}
	return int(days), int(frac)
}

// dateTimeJoin composes a TDateTime from a Delphi day number and a millisecond
// offset within that day, following Delphi's sign convention.
func dateTimeJoin(dayNumber, msecOfDay int) float64 {
	frac := float64(msecOfDay) / millisecondsPerDay
	if dayNumber < 0 {
		return float64(dayNumber) - frac
	}
	return float64(dayNumber) + frac
}

// civilDate is the civil calendar/clock decomposition of a TDateTime.
type civilDate struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

// decodeDateTime decomposes a TDateTime into civil calendar and clock fields.
func decodeDateTime(dt float64) civilDate {
	dayNumber, msec := dateTimeSplit(dt)
	t := dateTimeEpoch.AddDate(0, 0, dayNumber)
	return civilDate{
		Year:        t.Year(),
		Month:       int(t.Month()),
		Day:         t.Day(),
		Hour:        msec / 3600000,
		Minute:      (msec / 60000) % 60,
		Second:      (msec / 1000) % 60,
		Millisecond: msec % 1000,
	}
}

// encodeDateOnly converts a civil date into a TDateTime day number.
// It does not validate the date; use isValidDate for that.
func encodeDateOnly(year, month, day int) float64 {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return math.Round(t.Sub(dateTimeEpoch).Hours() / 24)
}

// encodeTimeOnly converts a civil clock reading into the fractional part of a
// TDateTime.
func encodeTimeOnly(hour, minute, second, millisecond int) float64 {
	msec := ((hour*60+minute)*60+second)*1000 + millisecond
	return float64(msec) / millisecondsPerDay
}

// encodeCivil composes a full TDateTime from civil calendar and clock fields.
func encodeCivil(year, month, day, hour, minute, second, millisecond int) float64 {
	dayNumber := int(encodeDateOnly(year, month, day))
	msec := ((hour*60+minute)*60+second)*1000 + millisecond
	return dateTimeJoin(dayNumber, msec)
}

// dateTimeToGoLocal reinterprets a TDateTime's civil fields as a time.Time in
// the host's local timezone.
func dateTimeToGoLocal(dt float64) time.Time {
	c := decodeDateTime(dt)
	return time.Date(c.Year, time.Month(c.Month), c.Day, c.Hour, c.Minute, c.Second,
		c.Millisecond*int(time.Millisecond), time.Local)
}

// goTimeToDateTime converts a time.Time's civil fields into a TDateTime,
// ignoring its location.
func goTimeToDateTime(t time.Time) float64 {
	return encodeCivil(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(),
		t.Nanosecond()/int(time.Millisecond))
}

// localDateTimeToUTCDateTime reinterprets a local wall-clock TDateTime as the
// corresponding UTC wall-clock TDateTime.
func localDateTimeToUTCDateTime(dt float64) float64 {
	return goTimeToDateTime(dateTimeToGoLocal(dt).UTC())
}

// utcDateTimeToLocalDateTime reinterprets a UTC wall-clock TDateTime as the
// corresponding local wall-clock TDateTime.
func utcDateTimeToLocalDateTime(dt float64) float64 {
	c := decodeDateTime(dt)
	t := time.Date(c.Year, time.Month(c.Month), c.Day, c.Hour, c.Minute, c.Second,
		c.Millisecond*int(time.Millisecond), time.UTC)
	return goTimeToDateTime(t.In(time.Local))
}

// nowDateTime returns the current local wall-clock time as a TDateTime.
func nowDateTime() float64 {
	return goTimeToDateTime(time.Now())
}

// utcNowDateTime returns the current UTC wall-clock time as a TDateTime.
func utcNowDateTime() float64 {
	return goTimeToDateTime(time.Now().UTC())
}

// =============================================================================
// Format settings
// =============================================================================

// DateTimeFormatSettings holds the script-visible locale settings that drive
// date/time formatting and parsing. It mirrors the subset of Delphi's
// TFormatSettings that DWScript exposes through the FormatSettings class,
// plus the DWScript-specific Zone selector.
//
// go-dws uses a fixed, locale-independent default set rather than the host
// locale, so that script output is reproducible across machines.
type DateTimeFormatSettings struct {
	ShortDateFormat string
	LongDateFormat  string
	ShortTimeFormat string
	LongTimeFormat  string
	TimeAMString    string
	TimePMString    string
	Zone            TimeZone
}

// DefaultDateTimeFormatSettings returns the settings a fresh execution starts
// with.
func DefaultDateTimeFormatSettings() DateTimeFormatSettings {
	return DateTimeFormatSettings{
		ShortDateFormat: "yyyy-mm-dd",
		LongDateFormat:  "yyyy-mm-dd",
		ShortTimeFormat: "hh:nn",
		LongTimeFormat:  "hh:nn:ss",
		TimeAMString:    "AM",
		TimePMString:    "PM",
		Zone:            TimeZoneLocal,
	}
}

// resolveZone replaces TimeZoneDefault with the settings' own Zone.
func (s *DateTimeFormatSettings) resolveZone(tz TimeZone) TimeZone {
	if tz == TimeZoneDefault {
		if s.Zone == TimeZoneDefault {
			return TimeZoneLocal
		}
		return s.Zone
	}
	return tz
}

var (
	shortDayNames   = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	longDayNames    = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	shortMonthNames = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	longMonthNames  = [12]string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
)

// =============================================================================
// Formatter
// =============================================================================

type dateTimeToken int

const (
	tokLiteral dateTimeToken = iota
	tokD
	tokDD
	tokDDD
	tokDDDD
	tokM
	tokMM
	tokMMM
	tokMMMM
	tokYY
	tokYYYY
	tokH24
	tokHH24
	tokH12
	tokHH12
	tokN
	tokNN
	tokS
	tokSS
	tokZ
	tokZZZ
	tokAMPM
	tokAMSlashPM
	tokASlashP
	tokUUU
)

type dateTimeItem struct {
	Literal string
	Token   dateTimeToken
}

// dateTimeFormatter is a compiled FormatDateTime format string.
type dateTimeFormatter struct {
	items []dateTimeItem
}

func (f *dateTimeFormatter) addLiteral(s string) {
	if n := len(f.items); n > 0 && f.items[n-1].Token == tokLiteral {
		f.items[n-1].Literal += s
		return
	}
	f.items = append(f.items, dateTimeItem{Token: tokLiteral, Literal: s})
}

// addToken appends a token, applying the two context-sensitive rewrites that
// Delphi's format strings require: "m"/"mm" directly after an hour token means
// minutes rather than months, and an am/pm token switches the preceding hour
// token from 24-hour to 12-hour.
func (f *dateTimeFormatter) addToken(tok dateTimeToken) {
	switch tok {
	case tokM, tokMM:
		for i := len(f.items) - 1; i >= 0; i-- {
			t := f.items[i].Token
			if t == tokLiteral {
				continue
			}
			if t >= tokH24 && t <= tokHH12 {
				if tok == tokM {
					tok = tokN
				} else {
					tok = tokNN
				}
			}
			break
		}
	case tokAMPM, tokAMSlashPM, tokASlashP:
		for i := len(f.items) - 1; i >= 0; i-- {
			switch f.items[i].Token {
			case tokH24:
				f.items[i].Token = tokH12
			case tokHH24:
				f.items[i].Token = tokHH12
			case tokH12, tokHH12:
			default:
				continue
			}
			break
		}
	}
	f.items = append(f.items, dateTimeItem{Token: tok})
}

func isFormatLetter(r rune, lower rune) bool {
	return r == lower || r == lower-32
}

// compileDateTimeFormat parses a Delphi/DWScript FormatDateTime format string.
//
// The am/pm markers are recognised the way DWScript does it, by comparing the
// whole remainder of the format string (dwsDateTime.pas uses StrComp), so
// "hh:nn ampm" is a marker while "ampm hh:nn" is literal text.
func compileDateTimeFormat(format string) *dateTimeFormatter {
	f := &dateTimeFormatter{}
	r := []rune(format)
	for p := 0; p < len(r); {
		c := r[p]
		switch {
		case isFormatLetter(c, 'd'):
			n := runLength(r, p, 'd', 4)
			switch n {
			case 1:
				f.addToken(tokD)
			case 2:
				f.addToken(tokDD)
			case 3:
				f.addToken(tokDDD)
			default:
				f.addToken(tokDDDD)
			}
			p += n
		case isFormatLetter(c, 'm'):
			n := runLength(r, p, 'm', 4)
			switch n {
			case 1:
				f.addToken(tokM)
			case 2:
				f.addToken(tokMM)
			case 3:
				f.addToken(tokMMM)
			default:
				f.addToken(tokMMMM)
			}
			p += n
		case isFormatLetter(c, 'y'):
			if p+1 < len(r) && isFormatLetter(r[p+1], 'y') {
				if p+3 < len(r) && isFormatLetter(r[p+2], 'y') && isFormatLetter(r[p+3], 'y') {
					f.addToken(tokYYYY)
					p += 4
				} else {
					f.addToken(tokYY)
					p += 2
				}
			} else {
				f.addLiteral("y")
				p++
			}
		case isFormatLetter(c, 'h'):
			if p+1 < len(r) && isFormatLetter(r[p+1], 'h') {
				f.addToken(tokHH24)
				p += 2
			} else {
				f.addToken(tokH24)
				p++
			}
		case isFormatLetter(c, 'n'):
			if p+1 < len(r) && isFormatLetter(r[p+1], 'n') {
				f.addToken(tokNN)
				p += 2
			} else {
				f.addToken(tokN)
				p++
			}
		case isFormatLetter(c, 's'):
			if p+1 < len(r) && isFormatLetter(r[p+1], 's') {
				f.addToken(tokSS)
				p += 2
			} else {
				f.addToken(tokS)
				p++
			}
		case isFormatLetter(c, 'z'):
			if p+2 < len(r) && isFormatLetter(r[p+1], 'z') && isFormatLetter(r[p+2], 'z') {
				f.addToken(tokZZZ)
				p += 3
			} else {
				f.addToken(tokZ)
				p++
			}
		case c == 'a':
			rest := string(r[p:])
			switch {
			case rest == "ampm":
				f.addToken(tokAMPM)
				p += 4
			case rest == "am/pm":
				f.addToken(tokAMSlashPM)
				p += 5
			case rest == "a/p":
				f.addToken(tokASlashP)
				p += 3
			default:
				f.addLiteral("a")
				p++
			}
		case c == 'u':
			if p+2 < len(r) && r[p+1] == 'u' && r[p+2] == 'u' {
				f.addToken(tokUUU)
				p += 3
			} else {
				f.addLiteral("u")
				p++
			}
		case c == '"' || c == '\'':
			quote := c
			start := p + 1
			end := start
			for end < len(r) && r[end] != quote {
				end++
			}
			if end > start {
				f.addLiteral(string(r[start:end]))
			}
			p = end
			if p < len(r) {
				p++
			}
		default:
			f.addLiteral(string(c))
			p++
		}
	}
	return f
}

// runLength counts how many consecutive case-insensitive occurrences of lower
// start at index p, capped at maximum.
func runLength(r []rune, p int, lower rune, maximum int) int {
	n := 0
	for p+n < len(r) && n < maximum && isFormatLetter(r[p+n], lower) {
		n++
	}
	return n
}

// apply renders a TDateTime with the compiled format and the given settings.
func (f *dateTimeFormatter) apply(dt float64, s *DateTimeFormatSettings) string {
	c := decodeDateTime(dt)
	dayNumber, _ := dateTimeSplit(dt)
	dow := int(dateTimeEpoch.AddDate(0, 0, dayNumber).Weekday()) // 0 = Sunday

	var b strings.Builder
	for _, item := range f.items {
		switch item.Token {
		case tokLiteral:
			b.WriteString(item.Literal)
		case tokD:
			fmt.Fprintf(&b, "%d", c.Day)
		case tokDD:
			fmt.Fprintf(&b, "%02d", c.Day)
		case tokDDD:
			b.WriteString(shortDayNames[dow])
		case tokDDDD:
			b.WriteString(longDayNames[dow])
		case tokM:
			fmt.Fprintf(&b, "%d", c.Month)
		case tokMM:
			fmt.Fprintf(&b, "%02d", c.Month)
		case tokMMM:
			b.WriteString(shortMonthNames[c.Month-1])
		case tokMMMM:
			b.WriteString(longMonthNames[c.Month-1])
		case tokYY:
			fmt.Fprintf(&b, "%02d", c.Year%100)
		case tokYYYY:
			fmt.Fprintf(&b, "%04d", c.Year)
		case tokH24:
			fmt.Fprintf(&b, "%d", c.Hour)
		case tokHH24:
			fmt.Fprintf(&b, "%02d", c.Hour)
		case tokH12:
			fmt.Fprintf(&b, "%d", (c.Hour+11)%12+1)
		case tokHH12:
			fmt.Fprintf(&b, "%02d", (c.Hour+11)%12+1)
		case tokN:
			fmt.Fprintf(&b, "%d", c.Minute)
		case tokNN:
			fmt.Fprintf(&b, "%02d", c.Minute)
		case tokS:
			fmt.Fprintf(&b, "%d", c.Second)
		case tokSS:
			fmt.Fprintf(&b, "%02d", c.Second)
		case tokZ:
			fmt.Fprintf(&b, "%d", c.Millisecond)
		case tokZZZ:
			fmt.Fprintf(&b, "%03d", c.Millisecond)
		case tokAMPM:
			if c.Hour >= 1 && c.Hour <= 12 {
				b.WriteString(s.TimeAMString)
			} else {
				b.WriteString(s.TimePMString)
			}
		case tokAMSlashPM:
			if c.Hour >= 1 && c.Hour <= 12 {
				b.WriteString("am")
			} else {
				b.WriteString("pm")
			}
		case tokASlashP:
			if c.Hour >= 1 && c.Hour <= 12 {
				b.WriteString("a")
			} else {
				b.WriteString("p")
			}
		case tokUUU:
			b.WriteString(utcOffsetString(dt))
		}
	}
	return b.String()
}

// utcOffsetString renders the "uuu" specifier: the local UTC offset in effect
// at the given local TDateTime, as "UTC+2" or "UTC-3:30".
func utcOffsetString(dt float64) string {
	offsetMinutes := int(math.Round((dt - localDateTimeToUTCDateTime(dt)) * 1440))
	var b strings.Builder
	b.WriteString("UTC")
	if offsetMinutes >= 0 {
		b.WriteByte('+')
	} else {
		b.WriteByte('-')
		offsetMinutes = -offsetMinutes
	}
	hours := offsetMinutes / 60
	fmt.Fprintf(&b, "%d", hours)
	if rem := offsetMinutes - hours*60; rem != 0 {
		fmt.Fprintf(&b, ":%02d", rem)
	}
	return b.String()
}

// errInvalidDateTime is raised for TDateTime values outside Delphi's range.
var errInvalidDateTime = fmt.Errorf("Invalid date/time")

// formatDateTimeSettings implements TdwsFormatSettings.FormatDateTime.
func formatDateTimeSettings(format string, dt float64, s *DateTimeFormatSettings, tz TimeZone) (string, error) {
	if format == "" {
		return "", nil
	}
	if dt < -693592 || dt > 2146790052 {
		return "", errInvalidDateTime
	}
	// A time-only value carries no date to shift, so no zone conversion applies.
	if s.resolveZone(tz) == TimeZoneUTC && math.Abs(dt) >= 1 {
		dt = utcDateTimeToLocalDateTime(dt)
	}
	return compileDateTimeFormat(format).apply(dt, s), nil
}

// dateTimeToStrSettings implements TdwsFormatSettings.DateTimeToStr.
func dateTimeToStrSettings(dt float64, s *DateTimeFormatSettings, tz TimeZone) (string, error) {
	return formatDateTimeSettings(s.ShortDateFormat+" "+s.LongTimeFormat, dt, s, tz)
}

// dateToStrSettings implements TdwsFormatSettings.DateToStr.
func dateToStrSettings(dt float64, s *DateTimeFormatSettings, tz TimeZone) (string, error) {
	return formatDateTimeSettings(s.ShortDateFormat, dt, s, tz)
}

// timeToStrSettings implements TdwsFormatSettings.TimeToStr.
func timeToStrSettings(dt float64, s *DateTimeFormatSettings, tz TimeZone) (string, error) {
	return formatDateTimeSettings(s.LongTimeFormat, dt, s, tz)
}

// =============================================================================
// Parser
// =============================================================================

// tryEncodeDate implements TdwsFormatSettings.TryEncodeDate.
func tryEncodeDate(year, month, day int, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
	if !isValidDate(year, month, day) {
		return 0, false
	}
	dt := encodeDateOnly(year, month, day)
	if s.resolveZone(tz) == TimeZoneUTC {
		dt = localDateTimeToUTCDateTime(dt)
	}
	return dt, true
}

// tryEncodeDateTime implements TdwsFormatSettings.TryEncodeDateTime.
func tryEncodeDateTime(year, month, day, hour, minute, second, msec int,
	s *DateTimeFormatSettings, tz TimeZone,
) (float64, bool) {
	if !isValidDate(year, month, day) || !isValidTime(hour, minute, second, msec) {
		return 0, false
	}
	dt := encodeCivil(year, month, day, hour, minute, second, msec)
	if s.resolveZone(tz) == TimeZoneUTC {
		dt = localDateTimeToUTCDateTime(dt)
	}
	return dt, true
}

// dateTimeParser holds the mutable cursor state of tryStrToDateTimeFormat.
type dateTimeParser struct {
	settings *DateTimeFormatSettings
	format   []rune
	str      []rune
	i        int // cursor into format
	p        int // cursor into str
	value    int
	tokStart int
	tokLen   int

	year, month, day           int
	hours, minutes, seconds    int
	msec                       int
	ampm                       byte // 0 none, 'a' or 'p'
	previousWasHour, hourToken bool
}

func (d *dateTimeParser) grabDigits(count int) bool {
	grabbed := false
	for count > 0 && d.p < len(d.str) {
		count--
		c := d.str[d.p]
		if c < '0' || c > '9' {
			break
		}
		d.value = d.value*10 + int(c-'0')
		d.p++
		grabbed = true
	}
	return grabbed
}

// grabName matches one of names (case-insensitively) at the current position,
// after rewinding the cursor by the token length. It returns the 1-based index
// of the match, or 0.
func (d *dateTimeParser) grabName(names []string, tokenLen int) int {
	d.p -= tokenLen
	if d.p < 0 {
		return 0
	}
	rest := string(d.str[d.p:])
	for idx, name := range names {
		if len(rest) >= len(name) && strings.EqualFold(rest[:len(name)], name) {
			d.p += len([]rune(name))
			return idx + 1
		}
	}
	return 0
}

func (d *dateTimeParser) grabAMPM() bool {
	d.i += 3
	d.p--
	if d.p < 0 {
		return false
	}
	rest := string(d.str[d.p:])
	am := d.settings.TimeAMString
	pm := d.settings.TimePMString
	if len(rest) >= len(am) && strings.EqualFold(rest[:len(am)], am) {
		d.p += len([]rune(am))
		d.ampm = 'a'
		return true
	}
	if len(rest) >= len(pm) && strings.EqualFold(rest[:len(pm)], pm) {
		d.p += len([]rune(pm))
		d.ampm = 'p'
		return true
	}
	return false
}

// grabLiteral matches a quoted literal run in the format string against the
// input, mirroring dwsDateTime.pas's GrabLitteral.
func (d *dateTimeParser) grabLiteral(quote rune) bool {
	d.p -= d.tokLen
	if d.tokLen >= 2 {
		// An even run of quote characters is an escaped empty literal.
		d.tokLen &= 1
		if d.tokLen == 0 {
			return true
		}
		d.tokStart = d.i - 1
	}
	for d.i < len(d.format) {
		if d.format[d.i] == quote {
			length := d.i - d.tokStart - 1
			if length > 0 {
				if d.p+length > len(d.str) {
					return false
				}
				if string(d.str[d.p:d.p+length]) != string(d.format[d.tokStart+1:d.tokStart+1+length]) {
					return false
				}
				d.i++
				d.p += length
			}
			return true
		}
		d.i++
	}
	return false
}

// tryStrToDateTimeFormat implements TdwsFormatSettings.TryStrToDateTime with an
// explicit format string. It is a port of the clean-room parser in
// Source/dwsDateTime.pas and reproduces its tolerances exactly, including the
// two-digit-year window and the single trailing character it allows.
func tryStrToDateTimeFormat(format, str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
	d := &dateTimeParser{
		settings: s,
		format:   []rune(format),
		str:      []rune(str),
	}

	for d.i < len(d.format) {
		c := d.format[d.i]
		d.tokStart = d.i
		d.value = 0
		for {
			if d.p < len(d.str) {
				digit := int(d.str[d.p]) - '0'
				if digit >= 0 && digit < 10 && d.value >= 0 {
					d.value = d.value*10 + digit
				} else {
					d.value = -1
				}
			} else {
				d.value = -1
			}
			d.i++
			d.p++
			if d.i >= len(d.format) || d.format[d.i] != c {
				break
			}
		}
		d.tokLen = d.i - d.tokStart

		lower := c
		if lower >= 'A' && lower <= 'Z' {
			lower += 'a' - 'A'
		}

		// Variable length fields accept more digits than the format shows.
		switch d.tokLen {
		case 1:
			switch lower {
			case 'm', 'd', 'h', 'n', 's':
				d.grabDigits(1)
			case 'z':
				d.grabDigits(2)
			}
		case 2:
			if lower == 'y' && !nextFormatCharIsField(d.format, d.i) {
				if d.grabDigits(2) {
					d.tokLen = 4
				}
			}
		}

		d.hourToken = false
		switch lower {
		case 'a':
			if d.tokStart+4 <= len(d.format) && string(d.format[d.tokStart:d.tokStart+4]) == "ampm" {
				if !d.grabAMPM() {
					return 0, false
				}
			}
		case 'd':
			switch d.tokLen {
			case 1, 2:
				d.day = d.value
			case 3:
				if d.grabName(shortDayNames[:], 3) == 0 {
					return 0, false
				}
			case 4:
				if d.grabName(longDayNames[:], 4) == 0 {
					return 0, false
				}
			default:
				return 0, false
			}
		case 'm':
			switch d.tokLen {
			case 1, 2:
				if d.previousWasHour {
					d.minutes = d.value
				} else {
					d.month = d.value
				}
			case 3:
				month := d.grabName(shortMonthNames[:], 3)
				if month == 0 {
					return 0, false
				}
				d.month = month
			case 4:
				month := d.grabName(longMonthNames[:], 4)
				if month == 0 {
					return 0, false
				}
				d.month = month
			default:
				return 0, false
			}
		case 'y':
			switch d.tokLen {
			case 4:
				d.year = d.value
			case 2:
				currentYear := decodeDateTime(nowDateTime()).Year
				d.year = (currentYear/100)*100 + d.value
				if d.year > currentYear+50 {
					d.year -= 100
				}
			default:
				return 0, false
			}
		case 'h':
			if d.tokLen != 1 && d.tokLen != 2 {
				return 0, false
			}
			d.hours = d.value
			d.hourToken = true
		case 'n':
			if d.tokLen != 1 && d.tokLen != 2 {
				return 0, false
			}
			d.minutes = d.value
		case 's':
			if d.tokLen != 1 && d.tokLen != 2 {
				return 0, false
			}
			d.seconds = d.value
		case 'z':
			if d.tokLen != 1 && d.tokLen != 3 {
				return 0, false
			}
			d.msec = d.value
		default:
			if c == '"' || c == '\'' {
				if !d.grabLiteral(c) {
					return 0, false
				}
				d.hourToken = d.previousWasHour
				break
			}
			if d.p < d.tokLen || d.p > len(d.str) {
				return 0, false
			}
			if string(d.str[d.p-d.tokLen:d.p]) != string(d.format[d.tokStart:d.tokStart+d.tokLen]) {
				return 0, false
			}
			d.hourToken = d.previousWasHour
		}
		d.previousWasHour = d.hourToken
	}

	// DWScript tolerates exactly one unconsumed trailing character.
	if d.p+1 < len(d.str) {
		return 0, false
	}
	if d.ampm != 0 && (d.hours < 0 || d.hours > 12) {
		return 0, false
	}
	if d.hours < 0 || d.hours >= 24 || d.minutes < 0 || d.minutes >= 60 ||
		d.seconds < 0 || d.seconds >= 60 || d.msec < 0 || d.msec >= 1000 {
		return 0, false
	}

	var dt float64
	if d.day != 0 || d.month != 0 || d.year != 0 {
		var ok bool
		dt, ok = tryEncodeDateTime(d.year, d.month, d.day, d.hours, d.minutes, d.seconds, d.msec, s, tz)
		if !ok {
			return 0, false
		}
	} else {
		dth := encodeTimeOnly(d.hours, d.minutes, d.seconds, d.msec)
		if s.resolveZone(tz) == TimeZoneUTC {
			today := math.Trunc(nowDateTime())
			dt = localDateTimeToUTCDateTime(today+dth) - math.Trunc(localDateTimeToUTCDateTime(today+dth))
		} else {
			dt = dth
		}
	}
	if d.ampm == 'p' {
		dt += 0.5
	}
	return dt, true
}

// nextFormatCharIsField reports whether the format character at index i starts
// another field, which suppresses the two-to-four digit year extension.
func nextFormatCharIsField(format []rune, i int) bool {
	if i >= len(format) {
		return false
	}
	return strings.ContainsRune("dmhnsz0123456789", format[i])
}

// tryStrToDateTimeSettings implements TdwsFormatSettings.TryStrToDateTime.
func tryStrToDateTimeSettings(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
	for _, format := range []string{
		s.ShortDateFormat + " " + s.ShortTimeFormat,
		s.ShortDateFormat + " " + s.LongTimeFormat,
		s.LongDateFormat + " " + s.LongTimeFormat,
		s.LongDateFormat + " " + s.ShortTimeFormat,
	} {
		if dt, ok := tryStrToDateTimeFormat(format, str, s, tz); ok {
			return dt, true
		}
	}
	return 0, false
}

// tryStrToDateSettings implements TdwsFormatSettings.TryStrToDate.
func tryStrToDateSettings(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
	if dt, ok := tryStrToDateTimeFormat(s.ShortDateFormat, str, s, tz); ok {
		return dt, true
	}
	return tryStrToDateTimeFormat(s.LongDateFormat, str, s, tz)
}

// tryStrToTimeSettings implements TdwsFormatSettings.TryStrToTime.
func tryStrToTimeSettings(str string, s *DateTimeFormatSettings, tz TimeZone) (float64, bool) {
	if dt, ok := tryStrToDateTimeFormat(s.ShortTimeFormat, str, s, tz); ok {
		return dt, true
	}
	return tryStrToDateTimeFormat(s.LongTimeFormat, str, s, tz)
}

// dateTimeConversionError builds the DWScript parsing failure message.
func dateTimeConversionError(str string) error {
	return fmt.Errorf(`Date/time parsing error for "%s"`, str)
}
