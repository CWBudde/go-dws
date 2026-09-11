package builtins

import "time"

// goTimeToDelphiDateTime converts a UTC time.Time into a TDateTime by taking
// its civil fields. It exists so the pre-existing date/time tests keep reading
// naturally after the core was rewritten around civil arithmetic.
func goTimeToDelphiDateTime(t time.Time) float64 {
	return goTimeToDateTime(t.UTC())
}

// delphiDateTimeToGoTime converts a TDateTime into the equivalent UTC
// time.Time, the inverse of goTimeToDelphiDateTime.
func delphiDateTimeToGoTime(dt float64) time.Time {
	c := decodeDateTime(dt)
	return time.Date(c.Year, time.Month(c.Month), c.Day, c.Hour, c.Minute, c.Second,
		c.Millisecond*int(time.Millisecond), time.UTC)
}
