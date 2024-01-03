import { DateTime, Duration, DurationLikeObject } from 'luxon'
import React, { useEffect, useState } from 'react'
import {
  formatTimestamp,
  getDT,
  getDur,
  FormatTimestampArg,
  formatRelative,
} from './timeFormat'

type TimeBaseProps = {
  prefix?: string
  suffix?: string
}

type TimeTimestampProps = TimeBaseProps &
  Omit<FormatTimestampArg, 'time'> & {
    time: string | DateTime | null | undefined | number
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
  const local = formatTimestamp({
    ...props,
    time,
    zone: 'local',
    format: props.format === 'relative' ? 'default' : props.format,
  })

  const title = local + ' local time'
  const zoneStr = ' ' + time.toFormat('ZZZZ')

  return (
    <React.Fragment>
      {prefix}
      <time
        dateTime={time.toISO()}
        title={display !== local ? title : undefined}
        style={{
          textDecorationStyle: 'dotted',
          textUnderlineOffset: '0.25rem',
          textDecorationLine: display !== local ? 'underline' : 'none',
        }}
      >
        {display}
        {display !== local && props.format !== 'relative' && zoneStr}
      </time>
      {suffix}
    </React.Fragment>
  )
}

type TimeDurationProps = TimeBaseProps & {
  precise?: boolean
  duration: DurationLikeObject | string | Duration
  units?: readonly (keyof DurationLikeObject)[]
  min?: DurationLikeObject | string | Duration
}

const TimeDuration: React.FC<TimeDurationProps> = (props) => {
  const dur =
    typeof props.duration === 'string'
      ? Duration.fromISO(props.duration)
      : Duration.fromObject(props.duration)
  const min = props.min ? getDur(props.min) : undefined

  return (
    <React.Fragment>
      {props.prefix}
      <time dateTime={dur.toISO()}>
        {formatRelative({
          dur,
          noQualifier: true,
          units: props.units,
          min,
          precise: props.precise,
        })}
      </time>
      {props.suffix}
    </React.Fragment>
  )
}

export type TimeProps = TimeTimestampProps | TimeDurationProps

function isTime(props: TimeProps): props is TimeTimestampProps {
  return 'time' in props && !!props.time
}

// Time will render a <time> element using Luxon to format the time.
export const Time: React.FC<TimeProps> = (props) => {
  if (isTime(props)) return <TimeTimestamp {...props} />
  return <TimeDuration {...props} />
}
