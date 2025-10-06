package schedulemanager

import "github.com/google/uuid"

type UpdateSchedArgs struct {
	ScheduleID uuid.UUID
}

func (UpdateSchedArgs) Kind() string { return "schedule-manager-update" }
