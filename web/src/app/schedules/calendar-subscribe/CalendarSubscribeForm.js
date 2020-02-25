import React from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../../forms'
import { Grid, TextField } from '@material-ui/core'
import { ScheduleSelect } from '../../selection'

export default function CalendarSubscribeForm(props) {
  return (
    <FormContainer
      disabled={props.loading}
      errors={props.errors}
      onChange={value => props.onChange(value)}
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
            label='Subscription Source Name'
            placeholder='My iCloud Calendar'
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
