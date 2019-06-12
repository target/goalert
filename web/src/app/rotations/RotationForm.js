import React from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { TimePicker } from 'material-ui-pickers'
import { TimeZoneSelect } from '../selection'
import { TextField, Grid, MenuItem } from '@material-ui/core'
import { startCase } from 'lodash-es'
import { DateTime, Info } from 'luxon'

const rotationTypes = ['hourly', 'daily', 'weekly']

export default class RotationForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string.isRequired,
      description: p.string.isRequired,
      timeZone: p.string.isRequired,
      type: p.oneOf(rotationTypes).isRequired,
      shiftLength: p.number.isRequired,
      start: p.string.isRequired,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf([
          'name',
          'description',
          'timeZone',
          'type',
          'start',
          'shiftLength',
        ]).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func.isRequired,
  }

  onChange = values => {
    if (!this.props.onChange) return
    this.props.onChange({
      ...values,
    })
  }

  render() {
    return (
      <FormContainer optionalLabels {...this.props}>
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              name='name'
              label='Name'
              required
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              multiline
              name='description'
              label='Description'
              required
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TimeZoneSelect}
              multiline
              name='timeZone'
              fieldName='timeZone'
              label='Time Zone'
              required
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              select
              required
              label='Rotation Type'
              name='type'
            >
              {rotationTypes.map(type => (
                <MenuItem value={type} key={type}>
                  {startCase(type)}
                </MenuItem>
              ))}
            </FormField>
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              required
              type='number'
              name='shiftLength'
              label='Shift Length'
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TimePicker}
              label='Handoff Time'
              name='start'
              required
              mapOnChangeValue={value => value.toISO()}
            />
          </Grid>
          {this.props.value.type === 'weekly' && this.renderDayOfWeekField()}
        </Grid>
      </FormContainer>
    )
  }

  renderDayOfWeekField() {
    return (
      <Grid item xs={12}>
        <TextField
          fullWidth
          select
          required
          label='Day of Week'
          name='dayOfWeek'
          value={this.dayOfWeek()}
          onChange={e => this.setDayOfWeek(e.target.value)}
        >
          {Info.weekdaysFormat('long').map((day, idx) => {
            return (
              <MenuItem key={day} value={idx + 1}>
                {day}
              </MenuItem>
            )
          })}
        </TextField>
      </Grid>
    )
  }

  dayOfWeek() {
    const { start, timeZone } = this.props.value
    return DateTime.fromISO(start, { zone: timeZone }).weekday
  }

  setDayOfWeek(weekday) {
    const { start, timeZone, ...other } = this.props.value
    this.props.onChange({
      ...other,
      timeZone,
      start: DateTime.fromISO(start, { zone: timeZone })
        .set({ weekday })
        .toISO(),
    })
  }
}
