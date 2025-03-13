import React from 'react'
import { FormContainer, FormField } from '../../forms'
import { Checkbox, FormControlLabel, Grid, TextField } from '@mui/material'
import { ScheduleSelect } from '../../selection'

interface CalendarSubscribeFormProps {
  loading?: boolean
  errors?: Array<{ message: string; field: string; helpLink: string }>
  value: { scheduleId: string; name: string; reminderMinutes: Array<string> }
  onChange: (value: object) => void
  scheduleReadOnly?: boolean
}

export default function CalendarSubscribeForm(
  props: CalendarSubscribeFormProps,
): React.ReactNode {
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
