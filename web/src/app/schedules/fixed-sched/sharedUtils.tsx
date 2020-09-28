import { DateTime } from 'luxon'
import React, { ReactNode } from 'react'

export interface Value {
  start: string
  end: string
  shifts: Shift[]
}

export interface Shift {
  start: string
  end: string
  user: User | null
}

export interface User {
  label: string
  value: string
}

// removes bottom margin from content text so form fields
// don't have a bunch of whitespace above them
export const contentText = {
  marginBottom: 0,
}

export const fmt = (t: string) =>
  DateTime.fromISO(t).toLocaleString(DateTime.DATETIME_MED)

interface StepContainerProps {
  children: ReactNode
  width?: string
}
export function StepContainer({ children, width = '75%' }: StepContainerProps) {
  const bodyStyle = {
    display: 'flex',
    justifyContent: 'center', // horizontal align
    width: '100%',
    marginTop: '2%', // slightly lower below dialog title toolbar
  }

  // adjusts width of centered child components
  const containerStyle = {
    width,
    height: 'fit-content',
  }

  return (
    <div style={bodyStyle}>
      <div style={containerStyle}>{children}</div>
    </div>
  )
}
