import React, { FunctionComponent, useEffect, useState } from 'react'
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
    const numWeeks = Math.trunc(timing / 604800) // There are 604800s in a week
    timing = timing - numWeeks * 604800
    if (numWeeks)
      timeString += Math.trunc(numWeeks) + ' Week' + (numWeeks > 1 ? 's ' : ' ')
  }

  if (days) {
    const numDays = Math.trunc(timing / 86400) // There are 86400s in a day
    timing = timing - numDays * 86400
    if (numDays)
      timeString += Math.trunc(numDays) + ' Day' + (numDays > 1 ? 's ' : ' ')
  }

  if (hours) {
    const numHours = Math.trunc(timing / 3600) // There are 3600s in a hour
    timing = timing - numHours * 3600
    if (numHours)
      timeString += Math.trunc(numHours) + ' Hour' + (numHours > 1 ? 's ' : ' ')
  }

  if (minutes) {
    const numMinutes = Math.trunc(timing / 60) // There are 60s in a minute
    timing = timing - numMinutes * 60
    if (numMinutes)
      timeString +=
        Math.trunc(numMinutes) + ' Minute' + (numMinutes > 1 ? 's ' : ' ')
  }

  if (seconds) {
    const numSeconds = Math.trunc(timing / 1) // There are 1s in a second
    if (numSeconds)
      timeString +=
        Math.trunc(numSeconds) + ' Second' + (numSeconds > 1 ? 's ' : ' ')
  }

  return timeString
}

function CountDown(props: CountDownProps): React.JSX.Element | string {
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
    if (!timeString) return expiredMessage || 'Time expired'
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
    return <WrapComponent style={props.style}>{formatTime()}</WrapComponent>
  }
  return formatTime()
}

export default CountDown as FunctionComponent<CountDownProps>
