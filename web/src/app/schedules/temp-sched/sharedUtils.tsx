import { DateTime, Interval } from 'luxon'
import React, { ReactNode } from 'react'

export type Value = {
  start: string
  end: string
  shifts: Shift[]
}

export type Shift = {
  start: string
  end: string
  userID: string

  user?: {
    id: string
    name: string
  }
}

const parseInterval = (start: string, end: string): Interval =>
  Interval.fromDateTimes(DateTime.fromISO(start), DateTime.fromISO(end))

export function validateShift(
  schedStart: string,
  schedEnd: string,
  shift: Shift,
): Error | null {
  const schedSpan = parseInterval(schedStart, schedEnd)
  const shiftSpan = parseInterval(shift.start, shift.end)

  // these two just for completeness but should never happen
  if (!shiftSpan.isValid) return new Error('invalid shift times')
  if (!schedSpan.isValid) return new Error('invalid schedule times')

  if (!schedSpan.engulfs(shiftSpan))
    return new Error('shift extends beyond temporary schedule')

  return null
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
