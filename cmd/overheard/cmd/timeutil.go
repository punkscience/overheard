package cmd

import (
	"fmt"
	"strings"
	"time"
)

// parseBestEffortTime parses a string like "Wed 6:00pm" into the next occurrence of that time.
func parseBestEffortTime(tStr string) (time.Time, error) {
	parts := strings.Split(tStr, " ")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: expected 'Day HH:MMam/pm'")
	}

	dayStr := parts[0]
	timeStr := parts[1]

	// Parse the time part
	parsedTime, err := time.Parse("3:04pm", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time part: %w", err)
	}

	// Find the next occurrence of the day of the week
	now := time.Now()
	targetWeekday, err := parseWeekday(dayStr)
	if err != nil {
		return time.Time{}, err
	}

	daysToAdd := (targetWeekday - now.Weekday() + 7) % 7
	// If the day is today, check if the time has already passed.
	if daysToAdd == 0 && (now.Hour() > parsedTime.Hour() || (now.Hour() == parsedTime.Hour() && now.Minute() > parsedTime.Minute())) {
		daysToAdd = 7
	}

	year, month, day := now.Date()
	targetTime := time.Date(year, month, day+int(daysToAdd), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())

	return targetTime, nil
}

func parseWeekday(dayStr string) (time.Weekday, error) {
	switch strings.ToLower(dayStr) {
	case "sun", "sunday":
		return time.Sunday, nil
	case "mon", "monday":
		return time.Monday, nil
	case "tue", "tuesday":
		return time.Tuesday, nil
	case "wed", "wednesday":
		return time.Wednesday, nil
	case "thu", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("invalid day of the week: %s", dayStr)
	}
}
