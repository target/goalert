import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../forms'
import Badge from '@material-ui/core/Badge'
import Grid from '@material-ui/core/Grid'
import Stepper from '@material-ui/core/Stepper'
import Step from '@material-ui/core/Step'
import StepButton from '@material-ui/core/StepButton'
import StepContent from '@material-ui/core/StepContent'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import {
  RotationSelect,
  ScheduleSelect,
  SlackChannelSelect,
  UserSelect,
} from '../selection'

import {
  RotateRight as RotationsIcon,
  Today as SchedulesIcon,
  Group as UsersIcon,
} from '@material-ui/icons'
import { SlackBW as SlackIcon } from '../icons/components/Icons'
import { Config } from '../util/RequireConfig'
import NumberField from '../util/NumberField'

const useStyles = makeStyles(() => ({
  badge: {
    top: -1,
    right: -1,
    backgroundColor: '#cd1831',
  },
  optional: {
    float: 'left',
    textAlign: 'left',
  },
  label: {
    paddingRight: '0.4em',
  },
  stepperRoot: {
    padding: 0,
  },
}))

function PolicyStepForm(props) {
  const [step, setStep] = useState(0)
  const { disabled, value } = props
  const classes = useStyles()

  function handleStepChange(stepChange) {
    if (stepChange === step) {
      setStep(null) // close
    } else {
      setStep(stepChange) // open
    }
  }

  // takes a list of { id, type } targets and return the ids for a specific type
  const getTargetsByType = (type) => (tgts) =>
    tgts
      .filter((t) => t.type === type) // only the list of the current type
      .map((t) => t.id) // array of ID strings

  // takes a list of ids and return a list of { id, type } concatted with the new set of specific types
  const makeSetTargetType = (curTgts) => (type) => (newTgts) =>
    curTgts
      .filter((t) => t.type !== type) // current targets without any of the current type
      .concat(newTgts.map((id) => ({ id, type }))) // add the list of current type to the end

  // then form fields would all point to `targets` but can map values
  const setTargetType = makeSetTargetType(value.targets)

  const badgeMeUpScotty = (len, txt) => (
    <Badge
      badgeContent={len}
      color='primary'
      invisible={!len}
      classes={{
        badge: classes.badge,
      }}
      tabIndex='0'
      aria-label={`Toggle ${txt}`}
    >
      <Typography className={classes.label}>{txt}</Typography>
    </Badge>
  )

  const optionalText = (
    <Typography
      className={classes.optional}
      color='textSecondary'
      variant='caption'
    >
      Optional
    </Typography>
  )

  return (
    <FormContainer {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Config>
            {(cfg) => (
              <Stepper
                activeStep={step}
                nonLinear
                orientation='vertical'
                classes={{
                  root: classes.stepperRoot,
                }}
              >
                <Step>
                  <StepButton
                    aria-expanded={(step === 0).toString()}
                    data-cy='rotations-step'
                    icon={<RotationsIcon />}
                    optional={optionalText}
                    onClick={() => handleStepChange(0)}
                    tabIndex='-1'
                  >
                    {badgeMeUpScotty(
                      getTargetsByType('rotation')(value.targets).length,
                      'Add Rotations',
                    )}
                  </StepButton>
                  <StepContent>
                    <FormField
                      component={RotationSelect}
                      disabled={disabled}
                      fieldName='targets'
                      fullWidth
                      label='Select Rotation(s)'
                      multiple
                      name='rotations'
                      mapValue={getTargetsByType('rotation')}
                      mapOnChangeValue={setTargetType('rotation')}
                    />
                  </StepContent>
                </Step>
                <Step>
                  <StepButton
                    aria-expanded={(step === 1).toString()}
                    data-cy='schedules-step'
                    icon={<SchedulesIcon />}
                    optional={optionalText}
                    onClick={() => handleStepChange(1)}
                    tabIndex='-1'
                  >
                    {badgeMeUpScotty(
                      getTargetsByType('schedule')(value.targets).length,
                      'Add Schedules',
                    )}
                  </StepButton>
                  <StepContent>
                    <FormField
                      component={ScheduleSelect}
                      disabled={disabled}
                      fieldName='targets'
                      fullWidth
                      label='Select Schedule(s)'
                      multiple
                      name='schedules'
                      mapValue={getTargetsByType('schedule')}
                      mapOnChangeValue={setTargetType('schedule')}
                    />
                  </StepContent>
                </Step>
                {cfg['Slack.Enable'] && (
                  <Step>
                    <StepButton
                      aria-expanded={(step === 2).toString()}
                      data-cy='slack-channels-step'
                      icon={<SlackIcon />}
                      optional={optionalText}
                      onClick={() => handleStepChange(2)}
                      tabIndex='-1'
                    >
                      {badgeMeUpScotty(
                        getTargetsByType('slackChannel')(value.targets).length,
                        'Add Slack Channels',
                      )}
                    </StepButton>
                    <StepContent>
                      <FormField
                        component={SlackChannelSelect}
                        disabled={disabled}
                        fieldName='targets'
                        fullWidth
                        label='Select Channel(s)'
                        multiple
                        name='slackChannels'
                        mapValue={getTargetsByType('slackChannel')}
                        mapOnChangeValue={setTargetType('slackChannel')}
                      />
                    </StepContent>
                  </Step>
                )}
                <Step>
                  <StepButton
                    aria-expanded={(step === 3).toString()}
                    data-cy='users-step'
                    icon={<UsersIcon />}
                    optional={optionalText}
                    onClick={() =>
                      handleStepChange(cfg['Slack.Enable'] ? 3 : 2)
                    }
                    tabIndex='-1'
                  >
                    {badgeMeUpScotty(
                      getTargetsByType('user')(value.targets).length,
                      'Add Users',
                    )}
                  </StepButton>
                  <StepContent>
                    <FormField
                      component={UserSelect}
                      disabled={disabled}
                      fieldName='targets'
                      fullWidth
                      label='Select User(s)'
                      multiple
                      name='users'
                      mapValue={getTargetsByType('user')}
                      mapOnChangeValue={setTargetType('user')}
                    />
                  </StepContent>
                </Step>
              </Stepper>
            )}
          </Config>
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={NumberField}
            disabled={disabled}
            fullWidth
            label='Delay (minutes)'
            name='delayMinutes'
            required
            min={1}
            max={9000}
            hint={
              value.delayMinutes === '0'
                ? 'This will cause the step to immediately escalate'
                : `This will cause the step to escalate after ${value.delayMinutes}m`
            }
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

PolicyStepForm.propTypes = {
  value: p.shape({
    targets: p.arrayOf(
      p.shape({ id: p.string.isRequired, type: p.string.isRequired }),
    ),
    delayMinutes: p.string.isRequired,
  }).isRequired,

  errors: p.arrayOf(
    p.shape({
      field: p.oneOf(['targets', 'delayMinutes']).isRequired,
      message: p.string.isRequired,
    }),
  ),

  disabled: p.bool,
  onChange: p.func,
}

export default PolicyStepForm
