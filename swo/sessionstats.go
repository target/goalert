package swo

import "time"

type StatsManager struct {
	sessCh  chan sessionRecord
	statsCh chan Stats
}

type Stats struct {
	Last1Min  TimeframeStats
	Last5Min  TimeframeStats
	Last15Min TimeframeStats
}

type TimeframeStats struct {
	Count   int
	AvgTime time.Duration
	MaxTime time.Duration
}

type sessionRecord struct {
	Dur time.Duration
	End time.Time
}

func NewStatsManager() *StatsManager {
	sm := &StatsManager{
		sessCh:  make(chan sessionRecord, 100),
		statsCh: make(chan Stats),
	}
	go sm.loop()

	return sm
}

func (sm *StatsManager) Start() (stop func()) {
	start := time.Now()
	return func() {
		end := time.Now()
		sm.sessCh <- sessionRecord{end.Sub(start), end}
	}
}

func (sm *StatsManager) Stats() Stats { return <-sm.statsCh }

func (sm *StatsManager) loop() {
	var sessRecs []sessionRecord
	var stats Stats
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			sessRecs, stats = updateStats(sessRecs)
		case sess := <-sm.sessCh:
			sessRecs = append(sessRecs, sess)
		case sm.statsCh <- stats:
		}
	}
}

func statsForTime(sessRecs []sessionRecord, t time.Time) (s TimeframeStats) {
	for _, sess := range sessRecs {
		if !sess.End.After(t) {
			continue
		}
		s.Count++
		s.AvgTime += sess.Dur
		if sess.Dur > s.MaxTime {
			s.MaxTime = sess.Dur
		}
	}
	if s.Count == 0 {
		return s
	}

	s.AvgTime = s.AvgTime / time.Duration(s.Count)

	return s
}

func updateStats(sessRecs []sessionRecord) ([]sessionRecord, Stats) {
	n := time.Now()

	var s Stats

	s.Last15Min = statsForTime(sessRecs, n.Add(-15*time.Minute))
	sessRecs = sessRecs[len(sessRecs)-s.Last15Min.Count:]

	s.Last5Min = statsForTime(sessRecs, n.Add(-5*time.Minute))
	s.Last1Min = statsForTime(sessRecs[len(sessRecs)-s.Last5Min.Count:], n.Add(-time.Minute))

	return sessRecs, s
}
