import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import Chip from '@material-ui/core/Chip'
import Grid from '@material-ui/core/Grid'
import ListItem from '@material-ui/core/ListItem'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'
import { sortBy } from 'lodash'
import { makeStyles } from '@material-ui/core/styles'
import { RotationChip, ScheduleChip, UserChip, SlackChip } from '../util/Chips'
import PolicyStepEditDialog from './PolicyStepEditDialog'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import OtherActions from '../util/OtherActions'
import { useResetURLParams, useURLParam } from '../actions'

const shapeStep = p.shape({
  id: p.string.isRequired,
  delayMinutes: p.number.isRequired,
  targets: p.arrayOf(
    p.shape({
      id: p.string.isRequired,
      name: p.string.isRequired,
      type: p.string.isRequired,
    }),
  ).isRequired,
})

const useStyles = makeStyles(() => ({
  centerFlex: {
    display: 'flex',
    alignItems: 'center',
    height: 'fit-content',
  },
}))

function PolicyStep(props) {
  const classes = useStyles()

  const [editStep, setEditStep] = useURLParam('editStep', null)
  const resetEditStep = useResetURLParams('editStep')
  const [deleteStep, setDeleteStep] = useState(false)

  function getStepNumber(sid) {
    const sids = props.steps.map((s) => s.id)
    return sids.indexOf(sid) + 1
  }

  /*
   * Renders the mui chips for each target on the step
   */
  function renderChips() {
    const { targets: _t } = props.step

    // copy and sort by type then name
    const targets = sortBy(_t.slice(), ['type', 'name'])

    if (!targets || targets.length === 0) {
      return <Chip label='No targets' />
    }

    const items = targets.map((tgt) => {
      const tgtChip = (Chip) => <Chip id={tgt.id} name={tgt.name} />

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
  function renderDelayMessage() {
    const { repeat, step, steps } = props
    const len = steps.length
    const isLastStep = getStepNumber(step.id) === len

    // if it's the last step and should not repeat, do not render end text
    if (isLastStep && repeat === 0) {
      return null
    }

    const pluralizer = (x) => (x === 1 ? '' : 's')

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
      <ListItem id={index}>
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
            positionRelative
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

PolicyStep.propTypes = {
  escalationPolicyID: p.string.isRequired,
  repeat: p.number.isRequired, // # of times EP repeats escalation process
  step: shapeStep.isRequired,
  steps: p.arrayOf(shapeStep).isRequired,
  index: p.number,
}

export default PolicyStep
