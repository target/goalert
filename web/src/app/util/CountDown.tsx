import React, { useEffect, useState } from 'react'
import { DateTime } from 'luxon'

interface CountDownProps {
  end: string
  expiredMessage?: string
  prefix?: string
  expiredTimeout?: number
  seconds?: boolean
  minutes?: boolean
  hours?: boolean
  days?: boolean
  weeks?: boolean
  WrapComponent?: typeof React.Component
  style?: React.CSSProperties
}

export function formatTimeRemaining(
  timeRemaining: number,
  weeks: boolean | undefined,
  days: boolean | undefined,
  hours: boolean | undefined,
  minutes: boolean | undefined,
  seconds: boolean | undefined,
): string {
  let timing = timeRemaining
  let timeString = ''

  if (weeks) {
    const numWeeks = parseInt(`${timing / 604800}`, 10) // There are 604800s in a week
    timing = timing - numWeeks * 604800
    if (numWeeks) timeString += numWeeks + ' Week' + (numWeeks > 1 ? 's ' : ' ')
  }

  if (days) {
    const numDays = parseInt(`${timing / 86400}`, 10) // There are 86400s in a day
    timing = timing - numDays * 86400
    if (numDays) timeString += numDays + ' Day' + (numDays > 1 ? 's ' : ' ')
  }

  if (hours) {
    const numHours = parseInt(`${timing / 3600}`, 10) // There are 3600s in a hour
    timing = timing - numHours * 3600
    if (numHours) timeString += numHours + ' Hour' + (numHours > 1 ? 's ' : ' ')
  }

  if (minutes) {
    const numMinutes = parseInt(`${timing / 60}`, 10) // There are 60s in a minute
    timing = timing - numMinutes * 60
    if (numMinutes)
      timeString += numMinutes + ' Minute' + (numMinutes > 1 ? 's ' : ' ')
  }

  if (seconds) {
    const numSeconds = parseInt(`${timing / 1}`, 10) // There are 1s in a second
    if (numSeconds)
      timeString += numSeconds + ' Second' + (numSeconds > 1 ? 's ' : ' ')
  }

  return timeString
}

function CountDown(props: CountDownProps): JSX.Element | string {
  const [timeRemaining, setTimeRemaining] = useState(
    DateTime.fromISO(props.end).toSeconds() - DateTime.local().toSeconds(),
  )

  function formatTime(): string {
    const {
      expiredTimeout,
      prefix,
      expiredMessage,
      weeks,
      days,
      hours,
      minutes,
      seconds,
    } = props

    const timeout = expiredTimeout || 1

    // display if there is no other time
    if (timeRemaining < timeout) return expiredMessage || 'Time expired'

    let timeString = formatTimeRemaining(
      timeRemaining,
      weeks,
      days,
      hours,
      minutes,
      seconds,
    )
    if (prefix) timeString = prefix + timeString

    return timeString
  }

  useEffect(() => {
    const _counter = setInterval(() => {
      setTimeRemaining(
        DateTime.fromISO(props.end).toSeconds() - DateTime.local().toSeconds(),
      )
    }, 1000)

    return () => {
      clearInterval(_counter)
    }
  }, [])

  const WrapComponent = props.WrapComponent

  if (WrapComponent) {
    return (
      <WrapComponent {...props} style={props.style}>
        {formatTime()}
      </WrapComponent>
    )
  }
  return formatTime()
}

export default CountDown
