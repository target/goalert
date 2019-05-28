import React from 'react'
import { Link } from 'react-router-dom'

export const ScheduleLink = schedule => {
  return <Link to={`/schedules/${schedule.id}`}>{schedule.name}</Link>
}
