import { DateTime } from 'luxon'
import React, { ReactNode } from 'react'

interface UserObject {
  userID: string
}
export interface UserInfoObject {
  // adding in id and name to match graphql format
  user: { id: string; name: string }
}

export function useUserInfo<T extends UserObject>(
  items: T[],
): (T & UserInfoObject)[] {
  return items.map((item: T) => ({
    ...item,
    user: { id: item.userID, name: 'Bob' },
  }))
}

export interface Value {
  start: string
  end: string
  shifts: Shift[]
}

export type Shift = {
  start: string
  end: string
  userID: string
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
  }

  return (
    <div style={bodyStyle}>
      <div style={containerStyle}>{children}</div>
    </div>
  )
}
