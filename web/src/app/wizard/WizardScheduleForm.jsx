import React from 'react'
import p from 'prop-types'
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
import { value as valuePropType } from './propTypes'
import * as _ from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import makeStyles from '@mui/styles/makeStyles'
import { useIsWidthDown } from '../util/useWidth'

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

/**
 * Renders the form fields to be used in the wizard that
 * can be used for creating a primary and secondary schedule.
 */
export default function WizardScheduleForm({ value, onChange, secondary }) {
  const fullScreen = useIsWidthDown('md')
  const classes = useStyles()

  function renderFollowTheSun(key, schedType) {
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
              fullWidth={fullScreen}
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
              fullWidth={fullScreen}
              required
            />
          </Grid>
        </React.Fragment>
      )
    }
  }

  const getKey = () => {
    return secondary ? 'secondarySchedule' : 'primarySchedule'
  }

  const handleRotationTypeChange = (e) => {
    const newVal = _.cloneDeep(value)
    newVal[getKey()].rotation.type = e.target.value
    onChange(newVal)
  }

  const handleFollowTheSunToggle = (e) => {
    const newVal = _.cloneDeep(value)
    newVal[getKey()].followTheSunRotation.enable = e.target.value
    onChange(newVal)
  }

  function renderRotationFields(key, schedType) {
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
            {renderFollowTheSun(key, schedType)}
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
          fullWidth={fullScreen}
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
          fullWidth={fullScreen}
          required
          value={value[key].users}
        />
      </Grid>
      {renderRotationFields(key, schedType)}
    </React.Fragment>
  )
}

WizardScheduleForm.propTypes = {
  onChange: p.func.isRequired,
  value: valuePropType,
  secondary: p.bool,
}
