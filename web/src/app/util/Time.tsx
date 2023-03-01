import { DateTime, Duration, DurationLikeObject } from 'luxon'
import React, { useEffect, useState } from 'react'
import {
  formatTimestamp,
  getDT,
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
  const [, setTS] = useState('') // force re-render
  useEffect(() => {
    if (!['relative', 'relative-date'].includes(props.format || '')) return
    const interval = setInterval(() => setTS(DateTime.utc().toISO()), 1000)
    return () => clearInterval(interval)
  }, [props.format])

  const { prefix, zero, suffix } = props
  if (!props.time && !zero) return null
  if (!props.time)
    return (
      <React.Fragment>
        {prefix}
        {zero}
        {suffix}
      </React.Fragment>
    )

  const time = getDT(props.time, props.zone)
  const display = formatTimestamp({ ...props, time })
  const local = formatTimestamp({ ...props, time, zone: 'local' })

  const title =
    formatTimestamp({ ...props, time, zone: 'local' }) + ' in local time'
  const zoneStr = ' ' + time.toFormat('ZZZZ')

  return (
    <React.Fragment>
      {prefix}
      <time
        dateTime={props.time}
        title={display !== local ? title : undefined}
        style={{
          textDecorationStyle: 'dotted',
          textDecorationLine: display !== local ? 'underline' : 'none',
        }}
      >
        {display}
        {display !== local && zoneStr}
      </time>
      {suffix}
    </React.Fragment>
  )
}

type TimeDurationProps = TimeBaseProps & {
  precise?: boolean
  duration: DurationLikeObject | string | Duration
  units?: readonly (keyof DurationLikeObject)[]
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
        {props.precise
          ? toRelativePrecise(dur, true, props.units)
          : dur.toHuman()}
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
