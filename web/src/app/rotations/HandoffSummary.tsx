import React from 'react'
import { RotationType } from '../../schema'
import { Time } from '../util/Time'
import { DateTime } from 'luxon'

export interface HandoffSummaryProps {
  start: string
  shiftLength: number
  type: RotationType
  timeZone: string
}

function dur(p: HandoffSummaryProps): JSX.Element | string {
  if (p.type === 'hourly') return <Time duration={{ hours: p.shiftLength }} />
  if (p.type === 'daily') return <Time duration={{ days: p.shiftLength }} />
  if (p.type === 'weekly') return <Time duration={{ weeks: p.shiftLength }} />
  if (p.type === 'monthly') return <Time duration={{ months: p.shiftLength }} />
  throw new Error('unknown rotation type: ' + p.type)
}

function ts(p: HandoffSummaryProps): JSX.Element | string {
  if (p.type === 'hourly')
    return <Time prefix='from ' time={p.start} zone={p.timeZone} />
  if (p.type === 'daily')
    return <Time prefix='at ' time={p.start} zone={p.timeZone} format='clock' />
  if (p.type === 'weekly')
    return (
      <Time
        prefix='on '
        time={p.start}
        zone={p.timeZone}
        format='weekday-clock'
      />
    )
  if (p.type === 'monthly')
    return (
      <Time
        prefix='from '
        time={DateTime.fromISO(p.start).plus({ months: 1 }).startOf('month')}
        zone={p.timeZone}
      />
    )
  throw new Error('unknown rotation type: ' + p.type)
}

// handoffSummary returns the summary description for the rotation
export const HandoffSummary: React.FC<HandoffSummaryProps> =
  function HandoffSummary(props: HandoffSummaryProps): JSX.Element {
    if (!props.timeZone) return <span>Loading handoff information...</span>
    return (
      <span>
        Time Zone: {props.timeZone}
        <br />
        {props.type === 'monthly' ? (
          <React.Fragment>
            Hands off at the beginning of every month starting {ts(props)}.
          </React.Fragment>
        ) : (
          <React.Fragment>
            Hands off every {dur(props)} {ts(props)}.
          </React.Fragment>
        )}
      </span>
    )
  }
