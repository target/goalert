import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'
import { variantDetails } from './ScheduleCalendarOverrideDialog'

const mutation = gql`
  mutation ($input: CreateUserOverrideInput!) {
    createUserOverride(input: $input) {
      id
    }
  }
`
export default function ScheduleOverrideCreateDialog(props) {
  const [value, setValue] = useState({
    addUserID: '',
    removeUserID: '',
    start: DateTime.local().startOf('hour').toISO(),
    end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
    ...props.defaultValue,
  })

  const notices = useOverrideNotices(props.scheduleID, value)

  const [mutate, { loading, error }] = useMutation(mutation, {
    variables: {
      input: {
        ...value,
        scheduleID: props.scheduleID,
      },
    },
    onCompleted: props.onClose,
  })

  return (
    <FormDialog
      onClose={props.onClose}
      title={variantDetails[props.variant].title}
      subTitle={variantDetails[props.variant].desc}
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialogue
      onSubmit={() => mutate()}
      form={
        <ScheduleOverrideForm
          add={props.variant !== 'remove'}
          remove={props.variant !== 'add'}
          scheduleID={props.scheduleID}
          disabled={loading}
          errors={fieldErrors(error)}
          value={value}
          onChange={(newValue) => setValue(newValue)}
          removeUserReadOnly={props.removeUserReadOnly}
        />
      }
    />
  )
}

ScheduleOverrideCreateDialog.defaultProps = {
  defaultValue: {},
}

ScheduleOverrideCreateDialog.propTypes = {
  scheduleID: p.string.isRequired,
  variant: p.oneOf(['add', 'remove', 'replace']).isRequired,
  onClose: p.func,
  removeUserReadOnly: p.bool,
  defaultValue: p.shape({
    addUserID: p.string,
    removeUserID: p.string,
    start: p.string,
    end: p.string,
  }),
}
