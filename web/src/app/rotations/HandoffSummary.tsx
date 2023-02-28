import { Duration } from 'luxon'
import React from 'react'
import { RotationType } from '../../schema'
import { Time } from '../util/Time'

export interface HandoffSummaryProps {
  start: string
  shiftLength: number
  type: RotationType
  timeZone: string
}

export const HourlyHandoffSummary: React.FC<HandoffSummaryProps> = (props) => {
  return (
    <span>
      Hands off every{' '}
      {Duration.fromObject({ hours: props.shiftLength }).toHuman()} from{' '}
      <Time time={props.start} zone={props.timeZone} />.
    </span>
  )
}

export const DailyHandoffSummary: React.FC<HandoffSummaryProps> = (props) => {
  const prefix =
    'Hands off every ' +
    Duration.fromObject({ days: props.shiftLength }).toHuman() +
    ' at '

  return (
    <Time
      prefix={prefix}
      time={props.start}
      zone={props.timeZone}
      format='clock'
      suffix='.'
    />
  )
}

export const WeeklyHandoffSummary: React.FC<HandoffSummaryProps> = (props) => {
  const prefix =
    'Hands off every ' +
    Duration.fromObject({ weeks: props.shiftLength }).toHuman() +
    ' on '

  return (
    <Time
      prefix={prefix}
      time={props.start}
      zone={props.timeZone}
      format='weekday-clock'
      suffix='.'
    />
  )
}

// handoffSummary returns the summary description for the rotation
export const HandoffSummary: React.FC<HandoffSummaryProps> =
  function HandoffSummary(props: HandoffSummaryProps): JSX.Element {
    if (!props.timeZone) return <span>Loading handoff information...</span>

    switch (props.type) {
      case 'hourly':
        return <HourlyHandoffSummary {...props} />
      case 'daily':
        return <DailyHandoffSummary {...props} />
      case 'weekly':
        return <WeeklyHandoffSummary {...props} />
    }

    throw new Error('unknown rotation type: ' + props.type)
  }
