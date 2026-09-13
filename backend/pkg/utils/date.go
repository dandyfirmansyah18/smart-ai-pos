package utils

import "time"

// ParseFlexibleDate attempts to parse a date string using common ISO, RFC, and standard date format fallbacks.
// If the input string is empty or unparseable, it returns time.Now().
func ParseFlexibleDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
	}

	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}
