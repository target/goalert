import React, { ComponentType, ReactElement, ReactNode, useState } from 'react'
import Chip from '@mui/material/Chip'
import Grid from '@mui/material/Grid'
import ListItem from '@mui/material/ListItem'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import Typography from '@mui/material/Typography'
import { sortBy } from 'lodash'
import makeStyles from '@mui/styles/makeStyles'
import {
  RotationChip,
  ScheduleChip,
  UserChip,
  SlackChip,
  WebhookChip,
} from '../util/Chips'
import PolicyStepEditDialog from './PolicyStepEditDialog'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import OtherActions from '../util/OtherActions'
import { useResetURLParams, useURLParam } from '../actions'
import { Target } from '../../schema'

interface PolicyStepProps {
  escalationPolicyID: string
  repeat: number // # of times EP repeats escalation process
  step: Step
  steps: Step[]
  index: number
  selected: boolean
}

interface Step {
  id: string
  delayMinutes: number
  targets: Target[]
}

const useStyles = makeStyles(() => ({
  centerFlex: {
    display: 'flex',
    alignItems: 'center',
    height: 'fit-content',
  },
}))

function PolicyStep(props: PolicyStepProps): JSX.Element {
  const classes = useStyles()

  const [editStep, setEditStep] = useURLParam<string>('editStep', '')
  const resetEditStep = useResetURLParams('editStep')
  const [deleteStep, setDeleteStep] = useState(false)

  function getStepNumber(sid: string): number {
    const sids = props.steps.map((s) => s.id)
    return sids.indexOf(sid) + 1
  }

  /*
   * Renders the mui chips for each target on the step
   */
  function renderChips(): ReactElement {
    const { targets: _t } = props.step

    // copy and sort by type then name
    const targets = sortBy(_t.slice(), ['type', 'name'])

    if (!targets || targets.length === 0) {
      return <Chip label='No targets' />
    }

    const items = targets.map((tgt) => {
      const tgtChip = (
        Chip: ComponentType<{ id: string; label: string }>,
      ): JSX.Element => <Chip id={tgt.id} label={tgt.name} />

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
  function renderDelayMessage(): ReactNode {
    const { repeat, step, steps } = props
    const len = steps.length
    const isLastStep = getStepNumber(step.id) === len

    // if it's the last step and should not repeat, do not render end text
    if (isLastStep && repeat === 0) {
      return null
    }

    const pluralizer = (x: number): string => (x === 1 ? '' : 's')

    let repeatText = `Move on to step #${getStepNumber(step.id) + 1} after ${
      step.delayMinutes
    } minute${pluralizer(step.delayMinutes)}`

    if (isLastStep && getStepNumber(step.id) === 1) {
      repeatText = `Repeat after ${step.delayMinutes} minutes`
    }

    // repeats
    if (isLastStep) {
      repeatText = `Go back to step #1 after ${
        step.delayMinutes
      } minute${pluralizer(step.delayMinutes)}`
    }

    return (
      <Typography variant='caption' component='p'>
        {repeatText}
      </Typography>
    )
  }

  const { index, step } = props

  return (
    <React.Fragment key={step.id}>
      <ListItem key={index} selected={props.selected}>
        <Grid container spacing={2}>
          <Grid item className={classes.centerFlex}>
            <Typography component='h4' variant='subtitle1'>
              <b>Step #{getStepNumber(step.id)}:</b>
            </Typography>
          </Grid>
          <Grid item xs={10}>
            {renderChips()}
          </Grid>
          <Grid item xs={12}>
            {renderDelayMessage()}
          </Grid>
        </Grid>
        <ListItemSecondaryAction>
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setEditStep(step.id),
              },
              {
                label: 'Delete',
                onClick: () => setDeleteStep(true),
              },
            ]}
          />
        </ListItemSecondaryAction>
      </ListItem>
      {editStep === step.id && (
        <PolicyStepEditDialog
          escalationPolicyID={props.escalationPolicyID}
          onClose={resetEditStep}
          step={props.step}
        />
      )}
      {deleteStep && (
        <PolicyStepDeleteDialog
          escalationPolicyID={props.escalationPolicyID}
          onClose={() => setDeleteStep(false)}
          stepID={props.step.id}
        />
      )}
    </React.Fragment>
  )
}

export default PolicyStep
