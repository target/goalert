import React from 'react'
import { PropTypes as p } from 'prop-types'
import { FormContainer, FormField } from '../../forms'
import { Grid, TextField } from '@material-ui/core'
import { ScheduleSelect } from '../../selection'
import MaterialSelect from '../../selection/MaterialSelect'
import _ from 'lodash-es'

export const reminderMinutesOptions = [
  {
    label: 'At time of shift',
    value: '0',
  },
  {
    label: '5 minutes before',
    value: '5',
  },
  {
    label: '10 minutes before',
    value: '10',
  },
  {
    label: '30 minutes before',
    value: '30',
  },
  {
    label: '1 hour before',
    value: '60',
  },
  {
    label: '1 day before',
    value: '1440',
  },
]

export default function CalendarSubscribeForm(props) {
  return (
    <FormContainer
      disabled={props.loading}
      errors={props.errors}
      onChange={value => props.onChange(value)}
      optionalLabels
      value={props.value}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            label='Subscription Name'
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
        {renderReminderMinutesFields()}
      </Grid>
    </FormContainer>
  )

  function renderReminderMinutesFields() {
    let arr = _.get(props, 'value.reminderMinutes', []).slice()

    const fields = []
    // push 1 more field than there are values, max of 5 fields
    // every field after the first is optional
    for (let i = 0; i < arr.length + 1 && i < 5; i++) {
      fields.push(
        <Grid key={i} item xs={12}>
          <FormField
            fullWidth
            component={MaterialSelect}
            name={`reminderMinutes[${i}]`}
            label={getAlarmLabel(i)}
            required={i === 0}
            options={reminderMinutesOptions}
            mapValue={() => (arr[i] ? arr[i] : null)}
            overrideOnChange={value => {
              // if the next value is empty and it's not the first index, remove that
              // index field from the array instead of setting to null. this will remove the
              // form field from the dom
              const newReminderMinutes = props.value.reminderMinutes.slice()
              if (!value && i !== 0) {
                // cleanup value array to remove null value
                newReminderMinutes.splice(i, 1)
                props.onChange({
                  ...props.value,
                  reminderMinutes: newReminderMinutes,
                })
              } else if (arr[i]) {
                // update index with new value
                newReminderMinutes[i] = value
                props.onChange({
                  ...props.value,
                  reminderMinutes: newReminderMinutes,
                })
              } else {
                // concat new value to array
                props.onChange({
                  ...props.value,
                  reminderMinutes: newReminderMinutes.concat(value),
                })
              }
            }}
          />
        </Grid>,
      )
    }

    return fields
  }

  function getAlarmLabel(idx) {
    switch (idx) {
      case 0:
        return 'Alarm'
      case 1:
        return 'Second Alarm'
      case 2:
        return 'Third Alarm'
      case 3:
        return 'Fourth Alarm'
      case 4:
        return 'Fifth Alarm'
    }
  }
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
