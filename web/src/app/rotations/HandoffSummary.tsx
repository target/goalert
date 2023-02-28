import React from 'react'
import { RotationType } from '../../schema'
import { Time } from '../util/Time'
import { TimeFormat } from '../util/timeFormat'

export interface HandoffSummaryProps {
  start: string
  shiftLength: number
  type: RotationType
  timeZone: string
}

// handoffSummary returns the summary description for the rotation
export const HandoffSummary: React.FC<HandoffSummaryProps> =
  function HandoffSummary(props: HandoffSummaryProps): JSX.Element {
    if (!props.timeZone) return <span>Loading handoff information...</span>

    let join: string, format: TimeFormat, unit: string
    switch (props.type) {
      case 'hourly':
        join = 'from'
        format = 'locale'
        unit = 'hours'
        break
      case 'daily':
        join = 'at'
        format = 'clock'
        unit = 'days'
        break
      case 'weekly':
        join = 'on'
        format = 'weekday-clock'
        unit = 'weeks'
        break
      default:
        throw new Error('unknown rotation type: ' + props.type)
    }

    return (
      <span>
        Hands off every <Time duration={{ [unit]: props.shiftLength }} /> {join}{' '}
        <Time time={props.start} zone={props.timeZone} format={format} />.
      </span>
    )
  }
