package app

import "time"

// SetTimeOffset sets the current time offset for the application.
func (a *App) SetTimeOffset(dur time.Duration) {
	a.timeMx.Lock()
	defer a.timeMx.Unlock()

	a.timeOffset = dur
}

// Now returns the current time for the application.
func (a *App) Now() time.Time {
	t := time.Now()

	a.timeMx.Lock()
	defer a.timeMx.Unlock()

	return t.Add(a.timeOffset).UTC()
}
