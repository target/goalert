import { DateTime } from 'luxon'
import React, { useEffect, useState } from 'react'
import { formatTimestamp, TimeFormatOpts } from './timeFormat'

type NoTime = {
  time: '' | null | undefined
  zero: string
}
type Time = {
  time: string
  zero?: string
}

type TimeProps = TimeFormatOpts & {
  prefix?: string
  suffix?: string
  local?: boolean
} & (NoTime | Time)

// Time will render a <time> element using Luxon to format the time.
export const Time: React.FC<TimeProps> = (props) => {
  const [now, setNow] = useState(props.now || DateTime.utc().toISO())
  useEffect(() => {
    if (props.format !== 'relative' && props.format !== 'relative-date') return

    const interval = setInterval(() => {
      setNow(props.now || DateTime.utc().toISO())
    }, 1000)
    return () => clearInterval(interval)
  }, [props.now])
  const display = formatTimestamp({ ...props, now })
  const local = formatTimestamp({ ...props, now, zone: 'local' })

  if (props.local) {
    return (
      <React.Fragment>
        {props.prefix}
        {props.time ? (
          <time dateTime={props.time}>{local} in local time</time>
        ) : (
          props.zero
        )}
        {props.suffix}
      </React.Fragment>
    )
  }

  const title =
    display === local
      ? undefined
      : formatTimestamp({ ...props, zone: 'local', omitSameDate: '' }) +
        ' in local time'

  return (
    <React.Fragment>
      {props.prefix}
      {props.time ? (
        <time dateTime={props.time} title={title}>
          {display}
        </time>
      ) : (
        props.zero
      )}
      {props.suffix}
    </React.Fragment>
  )
}
