import { DateTime, Duration, DurationLikeObject } from 'luxon'
import React, { useEffect, useState } from 'react'
import {
  formatTimestamp,
  TimeFormatOpts,
  toRelativePrecise,
} from './timeFormat'

type TimeBaseProps = {
  prefix?: string
  suffix?: string
}

type TimeTimestampProps = TimeBaseProps &
  Omit<TimeFormatOpts, 'time'> & {
    time: string | null | undefined
    zero?: string
  }

const TimeTimestamp: React.FC<TimeTimestampProps> = (props) => {
  const nowProp = 'now' in props ? (props.now as string) : ''
  const [now, setNow] = useState(nowProp || DateTime.utc().toISO())
  useEffect(() => {
    if (props.format !== 'relative' && props.format !== 'relative-date') return

    const interval = setInterval(() => {
      setNow(nowProp || DateTime.utc().toISO())
    }, 1000)
    return () => clearInterval(interval)
  }, [nowProp])
  const time = props.time || ''
  const display = formatTimestamp({ ...props, time, now })
  const local = formatTimestamp({ ...props, time, now, zone: 'local' })

  const title =
    formatTimestamp({ ...props, time, zone: 'local' }) + ' in local time'
  const zoneStr =
    ' ' + DateTime.fromISO(time, { zone: props.zone }).toFormat('ZZZZ')

  const tag = props.time ? (
    <time
      dateTime={props.time}
      title={display !== local ? title : undefined}
      style={{
        textDecorationStyle: 'dotted',
        textDecorationLine: title ? 'underline' : 'none',
      }}
    >
      {display}
      {display !== local && zoneStr}
    </time>
  ) : (
    props.zero
  )

  return (
    <React.Fragment>
      {tag && props.prefix}
      {tag}
      {tag && props.suffix}
    </React.Fragment>
  )
}

type TimeDurationProps = TimeBaseProps & {
  precise?: boolean
  duration: DurationLikeObject | string | Duration
}

const TimeDuration: React.FC<TimeDurationProps> = (props) => {
  const dur =
    typeof props.duration === 'string'
      ? Duration.fromISO(props.duration)
      : Duration.fromObject(props.duration)

  return (
    <React.Fragment>
      {props.prefix}
      <time dateTime={dur.toISO()}>
        {props.precise ? toRelativePrecise(dur) : dur.toHuman()}
      </time>
      {props.suffix}
    </React.Fragment>
  )
}

export type TimeProps = TimeTimestampProps | TimeDurationProps

// Time will render a <time> element using Luxon to format the time.
export const Time: React.FC<TimeProps> = (props) => {
  if ('time' in props) return <TimeTimestamp {...props} />
  return <TimeDuration {...props} />
}
