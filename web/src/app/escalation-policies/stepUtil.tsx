import React, { ComponentType, ReactElement, ReactNode } from 'react'
import { sortBy } from 'lodash'
import Chip from '@mui/material/Chip'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import {
  RotationChip,
  ScheduleChip,
  UserChip,
  SlackChip,
  WebhookChip,
} from '../util/Chips'
import { Target } from '../../schema'

interface Step {
  id: string
  delayMinutes: number
  targets: Target[]
}

export function getStepNumber(stepID: string, steps: Step[]): number {
  const sids = steps.map((s) => s.id)
  return sids.indexOf(stepID) + 1
}

/*
 * Renders the mui chips for each target on the step
 */
export function renderChips({ targets: _t }: Step): ReactElement {
  // copy and sort by type then name
  const targets = sortBy(_t.slice(), ['type', 'name'])

  if (!targets || targets.length === 0) {
    return <Chip label='No targets' />
  }

  const items = targets.map((tgt) => {
    const tgtChip = (
      Chip: ComponentType<{ id: string; label: string }>,
    ): React.ReactNode => <Chip id={tgt.id} label={tgt.name} />

    let chip = null
    switch (tgt.type) {
      case 'user':
        chip = tgtChip(UserChip)
        break
      case 'schedule':
        chip = tgtChip(ScheduleChip)
        break
      case 'rotation':
        chip = tgtChip(RotationChip)
        break
      case 'slackChannel':
      case 'notificationChannel':
        chip = tgtChip(SlackChip)
        break
      case 'chanWebhook':
        chip = tgtChip(WebhookChip)
        break
    }

    if (chip) {
      return (
        <Grid item key={tgt.id + ':' + tgt.type}>
          {chip}
        </Grid>
      )
    }
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
  steps: Step[],
  step: Step,
  repeat: number,
): ReactNode {
  const len = steps.length
  const isLastStep = getStepNumber(step.id, steps) === len

  // if it's the last step and should not repeat, do not render end text
  if (isLastStep && repeat === 0) {
    return null
  }

  const pluralizer = (x: number): string => (x === 1 ? '' : 's')

  let repeatText = `Move on to step #${
    getStepNumber(step.id, steps) + 1
  } after ${step.delayMinutes} minute${pluralizer(step.delayMinutes)}`

  if (isLastStep && getStepNumber(step.id, steps) === 1) {
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
