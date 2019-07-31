import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Chip from '@material-ui/core/Chip'
import Grid from '@material-ui/core/Grid'
import ListItem from '@material-ui/core/ListItem'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'
import { sortBy } from 'lodash-es'
import { withStyles } from '@material-ui/core/styles'
import { RotationChip, ScheduleChip, UserChip, SlackChip } from '../util/Chips'
import PolicyStepEditDialog from './PolicyStepEditDialog'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import OtherActions from '../util/OtherActions'
import { setURLParam } from '../actions/main'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { resetURLParams } from '../actions'

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

const styles = {
  centerFlex: {
    display: 'flex',
    alignItems: 'center',
    height: 'fit-content',
  },
}

@connect(
  state => ({
    editStep: urlParamSelector(state)('editStep'),
  }),
  dispatch => ({
    setEditStep: value => dispatch(setURLParam('editStep', value)),
    resetEditStep: () => dispatch(resetURLParams('editStep')),
  }),
)
@withStyles(styles)
export default class PolicyStep extends Component {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    repeat: p.number.isRequired, // # of times EP repeats escalation process
    step: shapeStep.isRequired,
    steps: p.arrayOf(shapeStep).isRequired,
  }

  state = {
    delete: false,
  }

  getStepNumber = sid => {
    const sids = this.props.steps.map(s => s.id)
    return sids.indexOf(sid) + 1
  }

  /*
   * Renders the mui chips for each target on the step
   */
  renderChips = () => {
    const { targets: _t } = this.props.step

    // copy and sort by type then name
    const targets = sortBy(_t.slice(), ['type', 'name'])

    if (!targets || targets.length === 0) {
      return <Chip label={'No targets'} />
    }

    const items = targets.map(tgt => {
      const tgtChip = Chip => <Chip id={tgt.id} name={tgt.name} />

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
  renderDelayMessage = () => {
    const { repeat, step, steps } = this.props
    const len = steps.length
    const isLastStep = this.getStepNumber(step.id) === len

    // if it's the last step and should not repeat, do not render end text
    if (isLastStep && repeat === 0) {
      return null
    }

    const pluralizer = x => (x === 1 ? '' : 's')

    let repeatText = `Move on to step #${this.getStepNumber(step.id) +
      1} after ${step.delayMinutes} minute${pluralizer(step.delayMinutes)}`

    if (isLastStep && this.getStepNumber(step.id) === 1) {
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

  render() {
    const {
      classes,
      editStep,
      index,
      resetEditStep,
      setEditStep,
      step,
    } = this.props

    return (
      <React.Fragment key={step.id}>
        <ListItem id={index}>
          <Grid container spacing={2}>
            <Grid item className={classes.centerFlex}>
              <Typography component='h4' variant='subtitle1'>
                <b>Step #{this.getStepNumber(step.id)}:</b>
              </Typography>
            </Grid>
            <Grid item xs={10}>
              {this.renderChips()}
            </Grid>
            <Grid item xs={12}>
              {this.renderDelayMessage()}
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
                  onClick: () => this.setState({ delete: true }),
                },
              ]}
              positionRelative
            />
          </ListItemSecondaryAction>
        </ListItem>
        {editStep === step.id && (
          <PolicyStepEditDialog
            escalationPolicyID={this.props.escalationPolicyID}
            onClose={resetEditStep}
            step={this.props.step}
          />
        )}
        {this.state.delete && (
          <PolicyStepDeleteDialog
            escalationPolicyID={this.props.escalationPolicyID}
            onClose={() => this.setState({ delete: false })}
            stepID={this.props.step.id}
          />
        )}
      </React.Fragment>
    )
  }
}
