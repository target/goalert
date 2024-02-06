import React from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../../forms'
import { Checkbox, FormControlLabel, Grid, TextField } from '@mui/material'
import { ScheduleSelect } from '../../selection'

export default function CalendarSubscribeForm(props) {
  return (
    <FormContainer
      disabled={props.loading}
      errors={props.errors}
      onChange={(value) => props.onChange(value)}
      optionalLabels
      removeFalseyIdxs
      value={props.value}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            placeholder='My iCloud Calendar'
            hint='This name is only used for GoAlert, and will not appear in your calendar app.'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={ScheduleSelect}
            disabled={props.scheduleReadOnly}
            fullWidth
            required
            label='Schedule'
            name='scheduleID'
          />
        </Grid>
        <Grid item xs={12}>
          <FormControlLabel
            control={
              <FormField component={Checkbox} checkbox name='fullSchedule' />
            }
            label='Include on-call shifts of other users'
            labelPlacement='end'
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

CalendarSubscribeForm.propTypes = {
  errors: p.array,
  loading: p.bool,
  onChange: p.func.isRequired,
  scheduleReadOnly: p.bool,
  value: p.shape({
    scheduleID: p.string,
    name: p.string,
    reminderMinutes: p.array,
  }).isRequired,
}
