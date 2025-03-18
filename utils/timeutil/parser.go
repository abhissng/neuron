package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// Categorized layouts - reordered for priority and grouped by type
// ISO 8601
var isoLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05Z", // ISO 8601 without explicit offset, assumes UTC
}

// DateTime
var dateTimeLayouts = []string{
	time.DateTime,
	"2006-01-02 15:04:05 -0700 MST",
	"2006-01-02 15:04:05 MST", // No offset, just timezone name
	"2006-01-02 15:04:05 UTC", // Add UTC explicitly
	"2006/01/02 15:04:05",
	"01/02/2006 15:04:05",
	"02/01/2006 15:04:05",
	"2006.01.02 15:04:05",
	"01.02.2006 15:04:05",
	"02.01.2006 15:04:05",
	time.ANSIC,                 // Includes date and time, broad coverage
	"Jan 2 15:04:05 MST 2006",  // More explicit ANSIC-like with single digit day
	"Jan 02 15:04:05 MST 2006", // Explicit ANSIC layout
	time.UnixDate,
	time.RubyDate,
	time.RFC1123, // Includes Timezone info too
	time.RFC850,
	time.RFC822,
}

// DateOnly
var dateOnlyLayouts = []string{
	time.DateOnly,
	"2006-01-02",
	"01/02/2006",
	"02/01/2006",
	"2006.01.02",
	"01.02.2006",
	"02.01.2006",
	"20060102",
	"01022006",
	"02012006",
}

// timeOnly
var timeOnlyLayouts = []string{
	time.TimeOnly,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	"15:04:05",
	"15:04",
	"3:04PM",
	"3:04 PM",
	"3pm", // Added explicitly
	"3PM", // Added for robustness - case variation
	"3 pm",
	"3 pm",
	"03PM",  // Added for robustness - leading zero
	"03 pm", // Added for robustness - leading zero and lowercase pm
}

// ParseTime attempts to parse a string with various common time formats and returns a time.Time.
// It uses categorized lists of layouts to improve parsing efficiency.
func ParseTime(value string) (time.Time, error) {

	fallbackLayouts := []string{ // Less structured, try last
		time.Layout, // Go's default, very general
	}

	// Heuristic: Analyze the input string to choose layout categories to try first
	var prioritizedLayouts []string
	if strings.ContainsAny(value, ":") { // Likely contains time
		if strings.ContainsAny(value, "+-Zz ") { // Likely has timezone info
			prioritizedLayouts = append(prioritizedLayouts, isoLayouts...)
			prioritizedLayouts = append(prioritizedLayouts, dateTimeLayouts...)
		} else {
			prioritizedLayouts = append(prioritizedLayouts, timeOnlyLayouts...)
			prioritizedLayouts = append(prioritizedLayouts, dateTimeLayouts...)
		}
	} else { // Likely date only
		prioritizedLayouts = append(prioritizedLayouts, dateOnlyLayouts...)
		prioritizedLayouts = append(prioritizedLayouts, dateTimeLayouts...) // In case it's date + time without ":"
	}
	prioritizedLayouts = append(prioritizedLayouts, fallbackLayouts...) // Always try fallback layouts

	// Try prioritized layouts
	for _, layout := range prioritizedLayouts {
		parsedTime, err := time.Parse(layout, value)
		if err == nil {
			return parsedTime, nil // Success
		}
	}

	// If still not parsed, return error
	return time.Time{}, fmt.Errorf("failed to parse time string \"%s\" with any of the predefined formats", value)
}
