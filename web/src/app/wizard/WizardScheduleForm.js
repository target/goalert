import React from 'react'
import p from 'prop-types'
import FormControl from '@material-ui/core/FormControl'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormHelperText from '@material-ui/core/FormHelperText'
import FormLabel from '@material-ui/core/FormLabel'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import InputAdornment from '@material-ui/core/InputAdornment'
import Radio from '@material-ui/core/Radio'
import RadioGroup from '@material-ui/core/RadioGroup'
import DateRange from '@material-ui/icons/DateRange'
import Tooltip from '@material-ui/core/Tooltip'
import withStyles from '@material-ui/core/styles/withStyles'
import InfoIcon from '@material-ui/icons/Info'
import { DateTimePicker } from 'material-ui-pickers'
import { TimeZoneSelect, UserSelect } from '../selection'
import { FormField } from '../forms'
import { value as valuePropType } from './propTypes'
import { set } from 'lodash-es'

const styles = {
  div: {
    paddingLeft: '3em',
  },
  formField: {
    width: '20em',
  },
  inputAdornment: {
    marginBottom: '3px',
  },
  shiftLength: {
    marginTop: '4px',
  },
  sunLabel: {
    display: 'flex',
    alignItems: 'center',
  },
  tooltip: {
    fontSize: 12,
  },
}

/**
 * Renders the form fields to be used in the wizard that
 * can be used for creating a primary and secondary schedule.
 */
@withStyles(styles)
export default class WizardScheduleForm extends React.Component {
  static propTypes = {
    onChange: p.func.isRequired,
    value: valuePropType,
    secondary: p.bool,
  }

  getKey = () => {
    const { secondary } = this.props
    return secondary ? 'secondarySchedule' : 'primarySchedule'
  }

  changeRotationType = e => {
    const { onChange, value } = this.props
    onChange(set(value, [this.getKey(), 'rotation', 'type'], e.target.value))
    this.forceUpdate()
  }

  enableFollowTheSun = e => {
    const { onChange, value } = this.props
    onChange(
      set(
        value,
        [this.getKey(), 'followTheSunRotation', 'enable'],
        e.target.value,
      ),
    )
    this.forceUpdate()
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
  render() {
    const { classes, secondary, value } = this.props
    const key = secondary ? 'secondarySchedule' : 'primarySchedule'
    const schedType = secondary ? 'secondary' : 'primary'

    return (
      <React.Fragment>
        <Grid item xs={12}>
          <div className={classes.div}>
            <FormField
              component={TimeZoneSelect}
              name={`${key}.timeZone`}
              label={
                <React.Fragment>
                  What time zone is your <b>{schedType}</b> schedule based in?
                </React.Fragment>
              }
              formLabel
              required
              className={classes.formField}
            />
          </div>
        </Grid>
        <Grid item xs={12}>
          <div className={classes.div}>
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
              required
              value={value[key].users}
              className={classes.formField}
            />
          </div>
        </Grid>
        {this.renderRotationFields(key, schedType)}
      </React.Fragment>
    )
  }

  renderRotationFields(key, schedType) {
    const { classes, value } = this.props
    const hideRotationFields =
      value[key].users.length <= 1 ||
      value[key].rotation.type === 'never' ||
      !value[key].rotation.type

    return (
      <React.Fragment>
        <Grid item xs={12}>
          <div className={classes.div}>
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
                onChange={this.changeRotationType}
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
          </div>
        </Grid>
        {!hideRotationFields && (
          <React.Fragment>
            <Grid item>
              <div className={classes.div}>
                <FormField
                  component={DateTimePicker}
                  name={`${key}.rotation.startDate`}
                  required
                  mapOnChangeValue={value => value.toISO()}
                  label='When should the rotation first hand off to the next team
              member?'
                  formLabel
                  className={classes.formField}
                  showTodayButton
                  InputProps={{
                    endAdornment: (
                      <InputAdornment position='end'>
                        <IconButton>
                          <DateRange />
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
              </div>
            </Grid>
            <Grid item xs={12}>
              <div className={classes.div}>
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
                    onChange={this.enableFollowTheSun}
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
              </div>
            </Grid>
            {this.renderFollowTheSun(key, schedType)}
          </React.Fragment>
        )}
      </React.Fragment>
    )
  }

  renderFollowTheSun(key, schedType) {
    const { classes, value } = this.props

    if (value[key].followTheSunRotation.enable === 'yes') {
      return (
        <React.Fragment>
          <Grid item xs={12}>
            <div className={classes.div}>
              <FormField
                component={UserSelect}
                multiple
                name={`${key}.followTheSunRotation.users`}
                label={
                  <React.Fragment>
                    Who will be on call for your <b>{schedType}</b> follow the
                    sun rotation?
                  </React.Fragment>
                }
                formLabel
                required
                value={value[key].followTheSunRotation.users}
                className={classes.formField}
              />
            </div>
          </Grid>
          <Grid item xs={12}>
            <div className={classes.div}>
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
                required
                className={classes.formField}
              />
            </div>
          </Grid>
        </React.Fragment>
      )
    }
  }
}
