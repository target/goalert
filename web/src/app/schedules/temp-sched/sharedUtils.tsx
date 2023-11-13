import { DateTime } from 'luxon'
import React, { ReactNode } from 'react'

export type TempSchedValue = {
  start: string
  end: string
  shifts: Shift[]
}

export type Shift = {
  displayStart?: string
  start: string
  end: string
  userID: string
  truncated: boolean

  user?: null | {
    id: string
    name: string
  }
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
}: StepContainerProps): React.ReactNode {
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
