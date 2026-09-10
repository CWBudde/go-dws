# Date and Time

DWScript has no dedicated date type. `TDateTime` is an alias of `Float`, and a value is the
number of days since 1899-12-30, with the fractional part carrying the time of day. Because the
alias is a `Float`, an `Integer` is accepted anywhere a `TDateTime` is expected: `YearOf(0)` and
`FormatDateTime('yyyy', 0)` both compile.

Values are wall-clock numbers. Nothing in a `TDateTime` records which zone it belongs to, so
every function that turns a number into a moment — or a moment into a number — takes an optional
`DateTimeZone` telling it how to read the value.

```pascal
var d := EncodeDate(2019, 1, 31) + EncodeTime(1, 30, 0, 0);
PrintLn(DateTimeToStr(d));                      // 2019-01-31 01:30:00
PrintLn(FormatDateTime('dddd, d mmmm yyyy', d)); // Thursday, 31 January 2019
```

## DateTimeZone

`DateTimeZone` is a scoped enumeration with three members:

| Member | Meaning |
| --- | --- |
| `DateTimeZone.Default` | defer to `FormatSettings.Zone` |
| `DateTimeZone.Local` | the host's local wall-clock time |
| `DateTimeZone.UTC` | Coordinated Universal Time |

Omitting the argument is the same as passing `DateTimeZone.Default`, which resolves through
`FormatSettings.Zone` and ultimately to `DateTimeZone.Local`.

```pascal
var utc := EncodeDate(2019, 6, 1, DateTimeZone.UTC);
PrintLn(DateTimeToStr(utc, DateTimeZone.UTC));   // rendered as UTC
PrintLn(DateTimeToStr(utc, DateTimeZone.Local)); // the same instant, local clock
```

`Now`, `Date` and `Time` return local values; `UTCDateTime` returns a UTC one.
`LocalDateTimeToUTCDateTime` and `UTCDateTimeToLocalDateTime` convert between the two, and
`LocalDateTimeToUnixTime` / `UnixTimeToLocalDateTime` bridge to Unix epoch seconds.

## FormatSettings

`FormatSettings` is an engine-provided static class holding the locale settings that the
conversion built-ins fall back on. It is **mutable**: a script assigns to its class variables
and the very next call observes the change, because the class variables are the only storage —
there is no separate snapshot to keep in sync.

| Class variable | Type | Default | Used by |
| --- | --- | --- | --- |
| `ShortDateFormat` | `String` | `yyyy-mm-dd` | `DateToStr`, `DateTimeToStr`, `StrToDate`, `StrToDateTime` |
| `LongDateFormat` | `String` | `yyyy-mm-dd` | long-form date rendering |
| `ShortTimeFormat` | `String` | `hh:nn` | `TimeToStr`, `StrToTime` |
| `LongTimeFormat` | `String` | `hh:nn:ss` | `DateTimeToStr`, `StrToDateTime` |
| `TimeAMString` | `String` | `AM` | the `ampm` specifier |
| `TimePMString` | `String` | `PM` | the `ampm` specifier |
| `Zone` | `DateTimeZone` | `DateTimeZone.Local` | every built-in given no explicit zone |

```pascal
FormatSettings.ShortDateFormat := 'dd/mm/yyyy';
PrintLn(DateToStr(EncodeDate(2019, 1, 31)));  // 31/01/2019

FormatSettings.Zone := DateTimeZone.UTC;
PrintLn(DateTimeToStr(Now));                  // now rendered as UTC
```

Settings are per-engine, not global: each run starts from the defaults above.

## Format specifiers

`FormatDateTime` and `ParseDateTime` share one specifier set. Letters are case-insensitive;
anything else is copied through, and text can be quoted with `'…'` or `"…"` to protect letters
that would otherwise be read as specifiers.

| Specifier | Meaning |
| --- | --- |
| `d`, `dd` | day of month, unpadded / zero-padded |
| `ddd`, `dddd` | abbreviated / full weekday name |
| `m`, `mm` | month, unpadded / zero-padded |
| `mmm`, `mmmm` | abbreviated / full month name |
| `yy`, `yyyy` | two-digit / four-digit year |
| `h`, `hh` | hour, unpadded / zero-padded |
| `n`, `nn` | minute, unpadded / zero-padded |
| `s`, `ss` | second, unpadded / zero-padded |
| `z`, `zzz` | milliseconds, unpadded / zero-padded |
| `am/pm`, `a/p` | literal `am`/`pm` and `a`/`p` markers |
| `ampm` | marker taken from `FormatSettings.TimeAMString` / `TimePMString` |
| `uuu` | UTC offset in effect for the value, as `UTC+2` or `UTC-3:30` |

`m` directly after an hour specifier means *minutes*, not month, matching Delphi. Any of the
three am/pm markers also switches `h`/`hh` to a 12-hour clock, so `hh:nn am/pm` renders 15:07 as
`03:07 pm`. A marker is recognised only when it is the entire remainder of the format string,
which is how DWScript itself scans it: `'hh:nn ampm'` is a marker, `'ampm hh:nn'` is literal
text.

## Function reference

**Creation** — `EncodeDate(y, m, d [, zone])`, `EncodeTime(h, n, s, ms)`,
`EncodeDateTime(y, m, d, h, n, s, ms [, zone])`, `Now`, `Date`, `Time`, `UTCDateTime`.

**Decomposition** — `DecodeDate`, `DecodeTime`, `YearOf`, `MonthOf`, `MonthOfYear`, `DayOf`,
`DayOfMonth`, `HourOf`, `MinuteOf`, `SecondOf`, `DayOfWeek`, `DayOfTheWeek`, `DayOfYear`,
`WeekNumber`, `DateToWeekNumber`, `YearOfWeek`, `DateToYearOfWeek`, `IsLeapYear`.

**Arithmetic** — `IncYear`, `IncMonth`, `IncWeek`, `IncDay`, `IncHour`, `IncMinute`,
`IncSecond`, `IncMilliSecond`, `DaysBetween`, `HoursBetween`, `MinutesBetween`,
`SecondsBetween`, `FirstDayOfYear`, `FirstDayOfNextYear`, `FirstDayOfMonth`,
`FirstDayOfNextMonth`, `FirstDayOfWeek`.

`IncMonth` clamps to the end of the shorter month, so 31 January plus one month is 28 or
29 February.

**Formatting** — `FormatDateTime(fmt, value [, zone])`, `DateTimeToStr`, `DateToStr`,
`TimeToStr`, `DateToISO8601`, `DateTimeToISO8601(value [, precision])`, `DateTimeToRFC822`.

**Parsing** — `StrToDate`, `StrToTime`, `StrToDateTime`, their `…Def` variants that return a
caller-supplied fallback instead of failing, `ParseDateTime(fmt, text [, zone])`,
`ISO8601ToDateTime`, `RFC822ToDateTime`.

A parse failure without a `…Def` fallback raises `Date/time parsing error for "<text>"`. The
ISO 8601 scanner reports per-position detail instead, e.g.
`Unexpected character (121) instead of digit in TryParse`,
`"-" expected after month in TryParse`, `Invalid argument to date encode in TryParse` or
`Unsupported or invalid ISO8601 format in TryParse`.

**Unix time** — `UnixTime`, `UnixTimeMSec`, `UnixTimeToDateTime`, `DateTimeToUnixTime`,
`UnixTimeMSecToDateTime`, `DateTimeToUnixTimeMSec`, `LocalDateTimeToUnixTime`,
`UnixTimeToLocalDateTime`.

**Other** — `Sleep(msec)` pauses the script; a non-positive argument returns immediately.

## Testing note

Several fixtures in `testdata/fixtures/FunctionsTime/` come from DWScript's own suite and were
written on a Central European machine: some hard-code the CET/CEST offsets, and others refuse
to run at all under UTC (`Cannot perform test for GMT+0`). The fixture harness and
`cmd/fixture-report` therefore run every fixture with `TZ=Europe/Berlin` so the pass count does
not depend on where the suite runs.
