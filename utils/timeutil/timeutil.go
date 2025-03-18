package timeutil

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

/*
# SUPPORTED CONVERSION SPECIFICATIONS

| pattern | description |
|:--------|:------------|
| %A      | national representation of the full weekday name |
| %a      | national representation of the abbreviated weekday |
| %B      | national representation of the full month name |
| %b      | national representation of the abbreviated month name |
| %C      | (year / 100) as decimal number; single digits are preceded by a zero |
| %c      | national representation of time and date |
| %D      | equivalent to %m/%d/%y |
| %d      | day of the month as a decimal number (01-31) |
| %e      | the day of the month as a decimal number (1-31); single digits are preceded by a blank |
| %F      | equivalent to %Y-%m-%d |
| %H      | the hour (24-hour clock) as a decimal number (00-23) |
| %h      | same as %b |
| %I      | the hour (12-hour clock) as a decimal number (01-12) |
| %j      | the day of the year as a decimal number (001-366) |
| %k      | the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank |
| %l      | the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank |
| %M      | the minute as a decimal number (00-59) |
| %m      | the month as a decimal number (01-12) |
| %n      | a newline |
| %p      | national representation of either "ante meridiem" (a.m.)  or "post meridiem" (p.m.)  as appropriate. |
| %R      | equivalent to %H:%M |
| %r      | equivalent to %I:%M:%S %p |
| %S      | the second as a decimal number (00-60) |
| %T      | equivalent to %H:%M:%S |
| %t      | a tab |
| %U      | the week number of the year (Sunday as the first day of the week) as a decimal number (00-53) |
| %u      | the weekday (Monday as the first day of the week) as a decimal number (1-7) |
| %V      | the week number of the year (Monday as the first day of the week) as a decimal number (01-53) |
| %v      | equivalent to %e-%b-%Y |
| %W      | the week number of the year (Monday as the first day of the week) as a decimal number (00-53) |
| %w      | the weekday (Sunday as the first day of the week) as a decimal number (0-6) |
| %X      | national representation of the time |
| %x      | national representation of the date |
| %Y      | the year with century as a decimal number |
| %y      | the year without century as a decimal number (00-99) |
| %Z      | the time zone name |
| %z      | the time zone offset from UTC |
| %%      | a '%' |
*/

var formatRegex *regexp.Regexp // Compiled regular expression

// Map of strftime-style directives to Go's layout format
var replacementsMap = map[string]string{
	"%Y": "2006",
	"%y": "06",
	"%m": "01",
	"%d": "02",
	"%H": "15",
	"%I": "03",
	"%M": "04",
	"%S": "05",
	"%f": "999999",
	"%z": "-0700",
	"%Z": "MST",
	"%A": "Monday",
	"%a": "Mon",
	"%B": "January",
	"%b": "Jan",
	"%p": "PM",
	"%x": "01/02/06",   // national date representation
	"%X": time.Kitchen, // national time representation
	"%%": "%",
	"%D": "01/02/06",   // %m/%d/%y
	"%F": "2006-01-02", // %Y-%m-%d
	"%h": "Jan",        // same as %b
	"%n": "\n",
	"%R": "15:04",       // %H:%M
	"%r": "03:04:05 PM", // %I:%M:%S %p
	"%T": "15:04:05",    // %H:%M:%S
	"%t": "\t",
	"%v": "_2-Jan-2006", // %e-%b-%Y (using _2 for %e approx)
	"%e": "_2",          // the day of the month as a decimal number (1-31); single digits are preceded by a blank
	"%c": time.ANSIC,    // national representation of time and date (approx)
	"%j": "002",         // day of the year (approximation, needs special handling for real %j)
	"%k": "15",          // the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank (using %H - no leading space in Go format)
	"%l": "3",           // the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank (using %I - no leading space in Go format)
	"%U": "01",          // week number of the year (Sunday as the first day of the week) as a decimal number (00-53) (approximation)
	"%u": "1",           // the weekday (Monday as the first day of the week) as a decimal number (1-7) (approximation)
	"%V": "01",          // the week number of the year (Monday as the first day of the week) as a decimal number (01-53) (approximation)
	"%W": "01",          // the week number of the year (Monday as the first day of the week) as a decimal number (00-53) (approximation)
	"%w": "0",           // the weekday (Sunday as the first day of the week) as a decimal number (0-6) (approximation)
	"%C": "06",          // century (year/100) as decimal number (approximation using 06 for century)
}

func init() {
	// Build the regular expression dynamically from the replacements map.
	var replacements []string
	for k := range replacementsMap {
		replacements = append(replacements, regexp.QuoteMeta(k)) // Escape special regex chars
	}
	regexStr := strings.Join(replacements, "|") // Create "OR" regex
	formatRegex = regexp.MustCompile(regexStr)

}

// TimeWrapper provides a high-level interface for time operations
type TimeWrapper struct {
	time.Time
}

// LoadLocation wraps time.LoadLocation
func LoadLocation(location string) (*time.Location, error) {
	var loc *time.Location
	var err error

	if strings.TrimSpace(location) == "" {
		loc = time.UTC
	} else {
		loc, err = time.LoadLocation(location)
		if err != nil {
			return nil, fmt.Errorf("failed to load location: %w", err)
		}
	}
	return loc, nil
}

// Now returns current time
func Now() TimeWrapper {
	return TimeWrapper{time.Now()}
}

// NewTime creates a new TimeWrapper instance
func NewTime(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) TimeWrapper {
	return TimeWrapper{
		time.Date(year, month, day, hour, min, sec, nsec, loc),
	}
}

// Format converts strftime-style directives to Go's layout format
func (t TimeWrapper) Format(format string) string {
	// goLayout := convertLayout(format)
	// formattedTime := t.Time.Format(goLayout)

	// Post-format processing for patterns that Go's time.Format doesn't directly support or requires special handling.
	formattedTime := formatRegex.ReplaceAllStringFunc(format, func(match string) string {
		switch match {
		case "%j":
			return fmt.Sprintf("%03d", t.YearDay()) // Day of year (001-366)
		case "%k":
			hour := t.Hour()
			return fmt.Sprintf("%2d", hour) // Hour (24-hour clock, space padded)
		case "%l":
			hour := t.Hour() % 12
			if hour == 0 {
				hour = 12
			}
			return fmt.Sprintf("%2d", hour) // Hour (12-hour clock, space padded)
		case "%U":
			_, week := t.ISOWeek() // Actually ISO week, not Sunday-first week, approximation
			return fmt.Sprintf("%02d", week)
		case "%u":
			weekday := t.Weekday()
			if weekday == time.Sunday {
				return "7" // Monday as 1, Sunday as 7
			}
			return fmt.Sprintf("%d", weekday)
		case "%V":
			_, week := t.ISOWeek()
			if week < 1 {
				week = 52 // Or 53, depending on year. For simplicity using 52
				// year = year - 1 // Previous year
			}
			return fmt.Sprintf("%02d", week) // ISO week number (01-53)
		case "%W":
			_, week := t.ISOWeek() // Actually ISO week, not Monday-first week, approximation
			return fmt.Sprintf("%02d", week)
		case "%w":
			weekday := t.Weekday()
			return fmt.Sprintf("%d", weekday) // Weekday (Sunday as 0, Saturday as 6)
		case "%C":
			return fmt.Sprintf("%02d", t.Year()/100) // Century (year/100)
		default:
			// For other patterns, return the standard formatted part (already done before ReplaceAllStringFunc)
			return replacementsMap[match] // This line is technically not needed as formattedTime already has it, but for safety.
		}
	})
	return formattedTime
}

// ConvertLayout translates common strftime directives to Go time format
func ConvertLayout(format string) string {
	return formatRegex.ReplaceAllStringFunc(format, func(match string) string {
		if goFormat, ok := replacementsMap[match]; ok {
			return goFormat
		}
		return match // If no replacement found, return the match as is (for literal chars)
	})
}

// In returns time in different location
func (t TimeWrapper) In(loc *time.Location) TimeWrapper {
	return TimeWrapper{t.Time.In(loc)}
}

// AddDate adds years/months/days
func (t TimeWrapper) AddDate(years, months, days int) TimeWrapper {
	return TimeWrapper{t.Time.AddDate(years, months, days)}
}
