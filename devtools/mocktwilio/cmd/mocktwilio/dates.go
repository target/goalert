package main

import (
	"fmt"
	"time"
)

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func since(t time.Time) string {
	switch {
	case time.Since(t) < time.Minute:
		return "Just Now"
	case time.Since(t) < time.Hour:
		return fmt.Sprintf("%d min", int(time.Since(t).Minutes()))
	case sameDate(t, time.Now()):
		return t.Format("3:04 PM")
	case t.After(time.Now().AddDate(0, 0, -7)):
		return t.Format("Mon")
	case t.Year() == time.Now().Year():
		return t.Format("Jan 2")
	default:
		return t.Format("1/2/06")
	}
}

func timeHeader(eventTime time.Time) string {
	if sameDate(eventTime, time.Now()) {
		return eventTime.Format("3:04 PM")
	}

	// if yesterday
	if sameDate(eventTime, time.Now().Add(-24*time.Hour)) {
		return "Yesterday - " + eventTime.Format("3:04 PM")
	}

	return eventTime.Format("Monday - 3:04 PM")
}
