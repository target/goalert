import React, { ReactElement, ReactNode } from 'react'
import { sortBy } from 'lodash'
import Chip from '@mui/material/Chip'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import { Destination } from '../../schema'
import DestinationChip from '../util/DestinationChip'

export function renderChipsDest(_a: Destination[]): ReactElement {
  const actions = sortBy(_a.slice(), ['type', 'displayInfo.text'])
  if (!actions || actions.length === 0) {
    return <Chip label='No actions' />
  }

  const items = actions.map((a, idx) => {
    return (
      <Grid item key={idx}>
        <DestinationChip {...a.displayInfo} />
      </Grid>
    )
  })

  return (
    <Grid container spacing={1}>
      {items}
    </Grid>
  )
}

/*
 * Renders the delay message, dependent on if the escalation policy
 * repeats, and if the message is rendering on the last step
 */
export function renderDelayMessage(
  step: { delayMinutes: number },
  idx: number,
  repeat: number,
  isLastStep: boolean,
): ReactNode {
  // if it's the last step and should not repeat, do not render end text
  if (isLastStep && repeat === 0) {
    return null
  }

  const pluralizer = (x: number): string => (x === 1 ? '' : 's')

  let repeatText = `Move on to step #${
    idx + 2
  } after ${step.delayMinutes} minute${pluralizer(step.delayMinutes)}`

  if (isLastStep && idx === 0) {
    repeatText = `Repeat after ${step.delayMinutes} minutes`
  }

  // repeats
  if (isLastStep) {
    repeatText = `Go back to step #1 after ${
      step.delayMinutes
    } minute${pluralizer(step.delayMinutes)}`
  }

  return (
    <Typography variant='caption' component='p' sx={{ pt: 2 }}>
      {repeatText}
    </Typography>
  )
}
