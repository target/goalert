import React from 'react'
import { Target } from '../../schema'
import AppLink from '../util/AppLink'

export const ScheduleLink = (schedule: Target): React.ReactNode => {
  return <AppLink to={`/schedules/${schedule.id}`}>{schedule.name}</AppLink>
}
