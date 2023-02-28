import { DateTime } from 'luxon'
import React, { useEffect, useState } from 'react'
import { formatTimestamp, TimeFormatOpts } from './timeFormat'

type TimeProps = Omit<TimeFormatOpts, 'time'> & {
  prefix?: string
  suffix?: string
  local?: boolean
  time: string | null | undefined
  zero?: string
}

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
  const time = props.time || ''
  const display = formatTimestamp({ ...props, time, now })
  const local = formatTimestamp({ ...props, time, now, zone: 'local' })

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
      : formatTimestamp({ ...props, time, zone: 'local', omitSameDate: '' }) +
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
