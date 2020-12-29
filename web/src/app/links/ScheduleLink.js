import React from 'react'
import AppLink from '../util/AppLink'

export const ScheduleLink = (schedule) => {
  return <AppLink to={`/schedules/${schedule.id}`}>{schedule.name}</AppLink>
}
