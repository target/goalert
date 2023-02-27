import React from 'react'
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
  const display = formatTimestamp(props)
  const local = formatTimestamp({ ...props, zone: 'local' })

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
