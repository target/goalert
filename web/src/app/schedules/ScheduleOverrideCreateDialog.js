import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'

import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'

const copyText = {
  add: {
    title: 'Temporarily Add a User',
    desc:
      'This will add a new shift for the selected user, while the override is active. Existing shifts will remain unaffected.',
  },
  remove: {
    title: 'Temporarily Remove a User',
    desc:
      'This will remove (or split/shorten) shifts belonging to the selected user, while the override is active.',
  },
  replace: {
    title: 'Temporarily Replace a User',
    desc:
      'This will replace the selected user with another during any existing shifts, while the override is active. No new shifts will be created, only who is on-call will be changed.',
  },
}

const mutation = gql`
  mutation($input: CreateUserOverrideInput!) {
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
      title={copyText[props.variant].title}
      subTitle={copyText[props.variant].desc}
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
