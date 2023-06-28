import React from 'react'
import { gql, useQuery } from 'urql'
import { FormContainer, FormField } from '../forms'
import { Grid, Typography } from '@mui/material'

import { UserSelect } from '../selection'
import { mapOverrideUserError } from './util'
import DialogContentError from '../dialogs/components/DialogContentError'
import _ from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import { useScheduleTZ } from './useScheduleTZ'
import { fmtLocal } from '../util/timeFormat'
import { FieldError } from '../util/errutil'
import { ensureInterval } from './timeUtil'

const query = gql`
  query ($id: ID!) {
    userOverride(id: $id) {
      id
      addUser {
        id
        name
      }
      removeUser {
        id
        name
      }
      start
      end
    }
  }
`

interface ScheduleOverrideFormValue {
  addUserID: string
  removeUserID: string
  start: string
  end: string
}

interface ScheduleOverrideFormProps {
  scheduleID: string
  value: ScheduleOverrideFormValue
  add?: boolean
  remove?: boolean
  disabled: boolean
  errors: Array<FieldError>
  onChange: (value: ScheduleOverrideFormValue) => void
  removeUserReadOnly?: boolean
}

export default function ScheduleOverrideForm(
  props: ScheduleOverrideFormProps,
): JSX.Element {
  const {
    add,
    remove,
    errors = [],
    scheduleID,
    value,
    removeUserReadOnly,
    ...formProps
  } = props

  const { zone, isLocalZone } = useScheduleTZ(scheduleID)

  const conflictingUserFieldError = props.errors.find(
    (e) => e && e.field === 'userID',
  )

  // used to grab conflicting errors from pre-existing overrides
  const [{ data }] = useQuery({
    query,
    variables: {
      id: _.get(conflictingUserFieldError, 'details.CONFLICTING_ID', ''),
    },
    requestPolicy: 'cache-first',
    pause: !conflictingUserFieldError,
  })

  const userConflictErrors = errors
    .filter((e) => e.field !== 'userID')
    .concat(
      conflictingUserFieldError
        ? mapOverrideUserError(_.get(data, 'userOverride'), value, zone)
        : [],
    )

  return (
    <FormContainer
      optionalLabels
      errors={errors.concat(userConflictErrors)}
      value={value}
      {...formProps}
      onChange={(newValue: ScheduleOverrideFormValue) => {
        formProps.onChange(ensureInterval(value, newValue))
      }}
    >
      <Grid container spacing={2}>
        {remove && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={UserSelect}
              name='removeUserID'
              label={
                add && remove ? 'User Currently Scheduled' : 'User to Remove'
              }
              required
              disabled={removeUserReadOnly}
            />
          </Grid>
        )}
        {add && (
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={UserSelect}
              required
              name='addUserID'
              label='User to Add'
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <Typography color='textSecondary' sx={{ fontStyle: 'italic' }}>
            Times shown in schedule timezone ({zone || '...'})
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={zone}
            required
            name='start'
            disabled={!zone}
            hint={isLocalZone ? '' : fmtLocal(value.start)}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={zone}
            name='end'
            required
            disabled={!zone}
            hint={isLocalZone ? '' : fmtLocal(value.end)}
          />
        </Grid>
        {conflictingUserFieldError && (
          <DialogContentError error={conflictingUserFieldError.message} />
        )}
      </Grid>
    </FormContainer>
  )
}
