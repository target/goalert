import React from 'react'
import FormControl from '@mui/material/FormControl'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormHelperText from '@mui/material/FormHelperText'
import FormLabel from '@mui/material/FormLabel'
import Grid from '@mui/material/Grid'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import Tooltip from '@mui/material/Tooltip'
import InfoIcon from '@mui/icons-material/Info'
import { TimeZoneSelect, UserSelect } from '../selection'
import { FormField } from '../forms'
import * as _ from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import makeStyles from '@mui/styles/makeStyles'
import { useIsWidthDown } from '../util/useWidth'
import { WizardFormValue } from './WizardForm'
import { RotationType } from '../../schema'

const useStyles = makeStyles(() => ({
  fieldItem: {
    marginLeft: '2.5em',
  },
  sunLabel: {
    display: 'flex',
    alignItems: 'center',
  },
  tooltip: {
    fontSize: 12,
  },
}))

interface WizardScheduleFormProps {
  value: WizardFormValue
  onChange: (value: WizardFormValue) => void
  secondary?: boolean
}

/**
 * Renders the form fields to be used in the wizard that
 * can be used for creating a primary and secondary schedule.
 */
export default function WizardScheduleForm({
  value,
  onChange,
  secondary,
}: WizardScheduleFormProps): React.JSX.Element {
  const fullScreen = useIsWidthDown('md')
  const classes = useStyles()
  const schedType = secondary ? 'secondary' : 'primary'

  const getScheduleField = (field: string): string => {
    return `${secondary ? 'secondarySchedule' : 'primarySchedule'}.${field}`
  }

  function renderFollowTheSun(): React.JSX.Element | undefined {
    const currentSchedule = secondary
      ? value.secondarySchedule
      : value.primarySchedule
    if (currentSchedule.followTheSunRotation.enable === 'yes') {
      return (
        <React.Fragment>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={UserSelect}
              multiple
              name={getScheduleField('followTheSunRotation.users')}
              label={
                <React.Fragment>
                  Who will be on call for your <b>{schedType}</b> follow the sun
                  rotation?
                </React.Fragment>
              }
              formLabel
              fullWidth={fullScreen}
              required
              value={currentSchedule.followTheSunRotation.users}
            />
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={TimeZoneSelect}
              name={getScheduleField('followTheSunRotation.timeZone')}
              label={
                <React.Fragment>
                  What time zone is your <b>{schedType}</b> follow the sun
                  rotation based in?
                </React.Fragment>
              }
              formLabel
              fullWidth={fullScreen}
              required
            />
          </Grid>
        </React.Fragment>
      )
    }
  }

  const handleRotationTypeChange = (
    e: React.ChangeEvent<HTMLInputElement>,
  ): void => {
    const newVal = _.cloneDeep(value)
    if (secondary) {
      newVal.secondarySchedule.rotation.type = e.target.value as RotationType
    } else {
      newVal.primarySchedule.rotation.type = e.target.value as RotationType
    }
    onChange(newVal)
  }

  const handleFollowTheSunToggle = (
    e: React.ChangeEvent<HTMLInputElement>,
  ): void => {
    const newVal = _.cloneDeep(value)
    if (secondary) {
      newVal.secondarySchedule.followTheSunRotation.enable = e.target.value
    } else {
      newVal.primarySchedule.followTheSunRotation.enable = e.target.value
    }
    onChange(newVal)
  }

  function renderRotationFields(): React.JSX.Element {
    const currentSchedule = secondary
      ? value.secondarySchedule
      : value.primarySchedule
    const hideRotationFields =
      currentSchedule.users.length <= 1 ||
      currentSchedule.rotation.type === 'never' ||
      !currentSchedule.rotation.type

    return (
      <React.Fragment>
        <Grid item xs={12} className={classes.fieldItem}>
          <FormControl>
            <FormLabel>
              How often will your <b>{schedType}</b> schedule need to rotate?
            </FormLabel>
            <i>
              <FormHelperText>
                At least two people are required to configure a rotation.
              </FormHelperText>
            </i>
            <RadioGroup
              aria-label='Rotation?'
              row
              value={currentSchedule.rotation.type}
              onChange={handleRotationTypeChange}
            >
              <FormControlLabel
                data-cy={getScheduleField('rotationType.weekly')}
                disabled={currentSchedule.users.length <= 1}
                value='weekly'
                control={<Radio />}
                label='Weekly'
              />
              <FormControlLabel
                data-cy={getScheduleField('rotationType.daily')}
                disabled={currentSchedule.users.length <= 1}
                value='daily'
                control={<Radio />}
                label='Daily'
              />
              <FormControlLabel
                data-cy={getScheduleField('rotationType.never')}
                disabled={currentSchedule.users.length <= 1}
                value='never'
                control={<Radio />}
                label='Never*'
              />
            </RadioGroup>
            <i>
              <FormHelperText>
                *Without rotating, all team members selected will be on call
                simultaneously.
              </FormHelperText>
            </i>
          </FormControl>
        </Grid>
        {!hideRotationFields && (
          <React.Fragment>
            <Grid item className={classes.fieldItem}>
              <FormField
                component={ISODateTimePicker}
                name={getScheduleField('rotation.startDate')}
                required
                label='When should the rotation first hand off to the next team
              member?'
                formLabel
                fullWidth={fullScreen}
              />
            </Grid>
            <Grid item xs={12} className={classes.fieldItem}>
              <FormControl>
                <FormLabel className={classes.sunLabel}>
                  Does your&nbsp;<b>{schedType}</b>&nbsp;schedule need to have
                  follow the sun support?
                  <Tooltip
                    classes={{ tooltip: classes.tooltip }}
                    data-cy='fts-tooltip'
                    disableFocusListener
                    placement='right'
                    title={`
                    “Follow the sun” means that support literally follows the sun—
                    it's a type of global workflow in which alerts can be handled by
                    and passed between offices in different time zones to increase
                    responsiveness and reduce delays.
                  `}
                    PopperProps={{
                      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                      // @ts-ignore
                      'data-cy': 'fts-tooltip-popper',
                    }}
                  >
                    <InfoIcon color='primary' />
                  </Tooltip>
                </FormLabel>
                <RadioGroup
                  aria-label='secondary'
                  name={getScheduleField('.fts')}
                  row
                  value={currentSchedule.followTheSunRotation.enable}
                  onChange={handleFollowTheSunToggle}
                >
                  <FormControlLabel
                    data-cy={getScheduleField('fts.yes')}
                    value='yes'
                    control={<Radio />}
                    label='Yes'
                  />
                  <FormControlLabel
                    data-cy={getScheduleField('fts.no')}
                    value='no'
                    control={<Radio />}
                    label='No'
                  />
                </RadioGroup>
              </FormControl>
            </Grid>
            {renderFollowTheSun()}
          </React.Fragment>
        )}
      </React.Fragment>
    )
  }
  /*
   * Choose users
   *  - if 1 user, add as user assignment
   *  - if 2+ users, add as rotation assignment
   * Create shifts for assignment, if needed
   *
   * Ask if second assignment is needed for evening shifts
   *   - repeat choosing users if yes
   */
  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.fieldItem}>
        <FormField
          component={TimeZoneSelect}
          name={getScheduleField('timeZone')}
          label={
            <React.Fragment>
              What time zone is your <b>{schedType}</b> schedule based in?
            </React.Fragment>
          }
          formLabel
          fullWidth={fullScreen}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.fieldItem}>
        <FormField
          component={UserSelect}
          data-cy={getScheduleField('users')}
          multiple
          name={getScheduleField('users')}
          label={
            <React.Fragment>
              Who will be on call for your <b>{schedType}</b> schedule?
            </React.Fragment>
          }
          formLabel
          fullWidth={fullScreen}
          required
          value={
            value[secondary ? 'secondarySchedule' : 'primarySchedule'].users
          }
        />
      </Grid>
      {renderRotationFields()}
    </React.Fragment>
  )
}
