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

type TimeDurationProps = TimeBaseProps & {
  precise?: boolean
} & { duration: DurationLikeObject | string }

type TimeTimestampProps = TimeBaseProps &
  Omit<TimeFormatOpts, 'time'> & {
    time: string | null | undefined
    zero?: string
  }

type TimeProps = TimeDurationProps | TimeTimestampProps

const TimeTimestamp: React.FC<TimeTimestampProps> = (props) => {
  const [now, setNow] = useState(props.now || DateTime.utc().toISO())
  useEffect(() => {
    if (props.format !== 'relative' && props.format !== 'relative-date') return

    const interval = setInterval(() => {
      setNow(props.now || DateTime.utc().toISO())
    }, 1000)
    return () => clearInterval(interval)
  }, [props.now])
  const time = props.time || ''
  const display = formatTimestamp({ ...props, time, now })
  const local = formatTimestamp({ ...props, time, now, zone: 'local' })

  const title =
    display === local
      ? undefined
      : formatTimestamp({ ...props, time, zone: 'local', omitSameDate: '' }) +
        ' in local time'

  const hasValue = Boolean(props.time || props.zero)

  return (
    <React.Fragment>
      {hasValue && props.prefix}
      {props.time ? (
        <time dateTime={props.time} title={title}>
          {display}
        </time>
      ) : (
        props.zero
      )}
      {hasValue && props.suffix}
    </React.Fragment>
  )
}

const TimeDuration: React.FC<TimeDurationProps> = (props) => {
  let dur: Duration
  if ('duration' in props) {
    dur =
      typeof props.duration === 'string'
        ? Duration.fromISO(props.duration)
        : Duration.fromObject(props.duration)
  } else {
    dur = Duration.fromObject(props)
  }

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

// Time will render a <time> element using Luxon to format the time.
export const Time: React.FC<TimeProps> = (props) => {
  if ('time' in props) return <TimeTimestamp {...props} />
  return <TimeDuration {...props} />
}
