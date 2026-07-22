package view

import "time"

// FormatDateTimeHuman formats a date-time string in human-readable form.
func FormatDateTimeHuman(dt, format string) string {
	t, err := time.Parse(format, dt)
	if err != nil {
		return dt
	}
	return t.Format("Mon, 02 Jan 06")
}
