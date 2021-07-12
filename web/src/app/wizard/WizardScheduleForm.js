import React from 'react'
import p from 'prop-types'
import { makeStyles } from '@material-ui/core'
import FormControl from '@material-ui/core/FormControl'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormHelperText from '@material-ui/core/FormHelperText'
import FormLabel from '@material-ui/core/FormLabel'
import Grid from '@material-ui/core/Grid'
import Radio from '@material-ui/core/Radio'
import RadioGroup from '@material-ui/core/RadioGroup'
import Tooltip from '@material-ui/core/Tooltip'
import { isWidthDown } from '@material-ui/core/withWidth'
import InfoIcon from '@material-ui/icons/Info'
import { TimeZoneSelect, UserSelect } from '../selection'
import { FormField } from '../forms'
import { value as valuePropType } from './propTypes'
import { set } from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import useWidth from '../util/useWidth'

const styles = makeStyles(() => ({
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

/**
 * Renders the form fields to be used in the wizard that
 * can be used for creating a primary and secondary schedule.
 */

function renderFollowTheSun(key, schedType, classes, value, width) {
  if (value[key].followTheSunRotation.enable === 'yes') {
    return (
      <React.Fragment>
        <Grid item xs={12} className={classes.fieldItem}>
          <FormField
            component={UserSelect}
            multiple
            name={`${key}.followTheSunRotation.users`}
            label={
              <React.Fragment>
                Who will be on call for your <b>{schedType}</b> follow the sun
                rotation?
              </React.Fragment>
            }
            formLabel
            fullWidth={isWidthDown('md', width)}
            required
            value={value[key].followTheSunRotation.users}
          />
        </Grid>
        <Grid item xs={12} className={classes.fieldItem}>
          <FormField
            component={TimeZoneSelect}
            name={`${key}.followTheSunRotation.timeZone`}
            label={
              <React.Fragment>
                What time zone is your <b>{schedType}</b> follow the sun
                rotation based in?
              </React.Fragment>
            }
            formLabel
            fullWidth={isWidthDown('md', width)}
            required
          />
        </Grid>
      </React.Fragment>
    )
  }
}

function renderRotationFields(
  key,
  schedType,
  classes,
  value,
  width,
  handleFollowTheSunToggle,
  handleRotationTypeChange,
) {
  const hideRotationFields =
    value[key].users.length <= 1 ||
    value[key].rotation.type === 'never' ||
    !value[key].rotation.type

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
            value={value[key].rotation.type}
            onChange={handleRotationTypeChange}
          >
            <FormControlLabel
              data-cy={`${key}.rotationType.weekly`}
              disabled={value[key].users.length <= 1}
              value='weekly'
              control={<Radio />}
              label='Weekly'
            />
            <FormControlLabel
              data-cy={`${key}.rotationType.daily`}
              disabled={value[key].users.length <= 1}
              value='daily'
              control={<Radio />}
              label='Daily'
            />
            <FormControlLabel
              data-cy={`${key}.rotationType.never`}
              disabled={value[key].users.length <= 1}
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
              name={`${key}.rotation.startDate`}
              required
              label='When should the rotation first hand off to the next team
              member?'
              formLabel
              fullWidth={isWidthDown('md', width)}
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
                    'data-cy': 'fts-tooltip-popper',
                  }}
                >
                  <InfoIcon color='primary' />
                </Tooltip>
              </FormLabel>
              <RadioGroup
                aria-label='secondary'
                name={`${key}.fts`}
                row
                value={value[key].followTheSunRotation.enable}
                onChange={handleFollowTheSunToggle}
              >
                <FormControlLabel
                  data-cy={`${key}.fts.yes`}
                  value='yes'
                  control={<Radio />}
                  label='Yes'
                />
                <FormControlLabel
                  data-cy={`${key}.fts.no`}
                  value='no'
                  control={<Radio />}
                  label='No'
                />
              </RadioGroup>
            </FormControl>
          </Grid>
          {renderFollowTheSun(key, schedType, classes, value, width)}
        </React.Fragment>
      )}
    </React.Fragment>
  )
}

export default function WizardScheduleForm({
  secondary,
  forceUpdate,
  onChange,
  value,
}) {
  const width = useWidth()
  const classes = styles()

  function getKey() {
    return secondary ? 'secondarySchedule' : 'primarySchedule'
  }

  function handleRotationTypeChange(e) {
    onChange(set(value, [getKey(), 'rotation', 'type'], e.target.value))
    forceUpdate()
  }

  function handleFollowTheSunToggle(e) {
    onChange(
      set(value, [getKey(), 'followTheSunRotation', 'enable'], e.target.value),
    )
    forceUpdate()
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
  const key = secondary ? 'secondarySchedule' : 'primarySchedule'
  const schedType = secondary ? 'secondary' : 'primary'

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.fieldItem}>
        <FormField
          component={TimeZoneSelect}
          name={`${key}.timeZone`}
          label={
            <React.Fragment>
              What time zone is your <b>{schedType}</b> schedule based in?
            </React.Fragment>
          }
          formLabel
          fullWidth={isWidthDown('md', width)}
          required
        />
      </Grid>
      <Grid item xs={12} className={classes.fieldItem}>
        <FormField
          component={UserSelect}
          data-cy={`${key}.users`}
          multiple
          name={`${key}.users`}
          label={
            <React.Fragment>
              Who will be on call for your <b>{schedType}</b> schedule?
            </React.Fragment>
          }
          formLabel
          fullWidth={isWidthDown('md', width)}
          required
          value={value[key].users}
        />
      </Grid>
      {renderRotationFields(
        key,
        schedType,
        classes,
        value,
        width,
        handleFollowTheSunToggle,
        handleRotationTypeChange,
      )}
    </React.Fragment>
  )
}

WizardScheduleForm.propTypes = {
  onChange: p.func.isRequired,
  value: valuePropType,
  secondary: p.bool,
}
