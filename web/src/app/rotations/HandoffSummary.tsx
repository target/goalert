import React from 'react'
import { RotationType } from '../../schema'
import { Time } from '../util/Time'

export interface HandoffSummaryProps {
  start: string
  shiftLength: number
  type: RotationType
  timeZone: string
}

// handoffSummary returns the summary description for the rotation
export const HandoffSummary: React.FC<HandoffSummaryProps> =
  function HandoffSummary(rotation: HandoffSummaryProps): JSX.Element {
    const tz = rotation.timeZone

    if (!tz) return <span>Loading handoff information...</span>

    if (rotation.type === 'hourly') {
      return (
        <span>
          Hands off every{' '}
          {rotation.shiftLength === 1
            ? 'hour'
            : rotation.shiftLength + ' hours'}
          .
        </span>
      )
    }

    const unit = rotation.type === 'daily' ? 'day' : 'week'
    const lengthDesc =
      rotation.shiftLength === 1
        ? unit
        : rotation.shiftLength + ' ' + unit + 's'

    const prefix = `Hands off every ${lengthDesc} ${
      unit === 'day' ? 'at' : 'on'
    } `

    return (
      <Time
        prefix={prefix}
        time={rotation.start}
        zone={tz}
        format={unit === 'day' ? 'clock' : 'weekday-clock'}
        suffix='.'
      />
    )
  }
