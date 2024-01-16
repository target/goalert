import { DateTime, Duration, Interval } from 'luxon'
import React, { ReactNode } from 'react'

export type TempSchedValue = {
  start: string
  end: string
  shifts: Shift[]
  shiftDur: Duration
}

export type Shift = {
  displayStart?: string
  start: string
  end: string
  userID: string
  truncated: boolean
  custom?: boolean

  user?: null | {
    id: string
    name: string
  }
}

export function inferDuration(shifts: Shift[]): Duration | null {
  if (shifts.length === 0) {
    return null
  }
  const durations: Duration[] = []
  for (let i = 0; i < shifts.length; i++) {
    const startDateTime = DateTime.fromISO(shifts[i].start).toObject()
    const endDateTime = DateTime.fromISO(shifts[i].end).toObject()
    if (startDateTime && endDateTime) {
      const interval = Interval.fromDateTimes(
        DateTime.fromISO(shifts[i].start),
        DateTime.fromISO(shifts[i].end),
      )
      durations.push(interval.toDuration())
    }
  }
  if (durations.length === 0) {
    return null
  }

  const totalDurations = durations.reduce((acc, duration) => acc.plus(duration))

  const hours = totalDurations.as('hours') / durations.length
  const days = totalDurations.as('days') / durations.length
  const weeks = totalDurations.as('weeks') / durations.length

  const maxDuration = Math.max(hours, days, weeks)
  if (maxDuration === hours) return Duration.fromObject({ hours })
  if (maxDuration === days) return Duration.fromObject({ days })
  if (maxDuration === weeks) return Duration.fromObject({ weeks })

  return null
}

// defaultTempScheduleValue returns a timespan, with no shifts,
// of the following week.
export function defaultTempSchedValue(zone: string): TempSchedValue {
  // We want the start to be the _next_ start-of-week for the current locale.
  // For example, if today is Sunday, we want the start to be next Sunday.
  // If today is Saturday, we want the start to be tomorrow.
  const startDT = DateTime.local()
    .setZone(zone)
    .startOf('week')
    .plus({ weeks: 1 })

  return {
    start: startDT.toISO(),
    end: startDT.plus({ days: 7 }).toISO(),
    shifts: [],
    shiftDur: Duration.fromObject({ days: 1 }),
  }
}

// removes bottom margin from content text so form fields
// don't have a bunch of whitespace above them
export const contentText = {
  marginBottom: 0,
}

export const fmt = (t: string, zone = 'local'): string =>
  DateTime.fromISO(t, { zone }).toLocaleString(DateTime.DATETIME_MED)

type StepContainerProps = {
  children: ReactNode
  width?: string
}
export function StepContainer({
  children,
  width = '75%',
  ...rest
}: StepContainerProps): JSX.Element {
  const bodyStyle = {
    display: 'flex',
    justifyContent: 'center', // horizontal align
    width: '100%',
    height: '100%',
  }

  // adjusts width of centered child components
  const containerStyle = {
    width,
  }

  return (
    <div style={bodyStyle}>
      <div style={containerStyle} {...rest}>
        {children}
      </div>
    </div>
  )
}

// dtToDuration takes two date times and returns the duration between the two
export function dtToDuration(a: DateTime, b: DateTime): number {
  if (!a.isValid || !b.isValid) return -1
  return b.diff(a, 'hours').hours
}
