import React from 'react'
import { MuiLink } from '../util/AppLink'

export const ScheduleLink = (schedule) => {
  return <MuiLink to={`/schedules/${schedule.id}`}>{schedule.name}</MuiLink>
}
