package calsub

import (
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/oncall"
)

type renderData struct {
	ApplicationName string
	ScheduleID      uuid.UUID
	ScheduleName    string
	Shifts          []oncall.Shift
	ReminderMinutes []int
	Version         string
	GeneratedAt     time.Time
	FullSchedule    bool
	UserNames       map[string]string
}
