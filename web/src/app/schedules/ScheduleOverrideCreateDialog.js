import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'

export const variantDetails = {
  add: {
    title: 'Temporarily Add User',
    desc: 'Create a new shift for the selected user. Existing shifts will remain unaffected.',
    name: 'Additional Coverage',
    helperText: 'Add an additional on-call user for a given time frame.',
  },
  remove: {
    title: 'Temporarily Remove User',
    desc: 'Remove (or split/shorten) shifts belonging to the selected user during the specified time frame.',
    name: 'Remove Coverage',
    helperText: "Remove one user's shifts during a given time frame.",
  },
  replace: {
    title: 'Temporarily Replace User',
    desc: 'Select one user to take over the shifts of another user during the specified time frame.',
    name: "Cover Someone's Shifts",
    helperText:
      'Have one user take over the shifts of another user during a given time frame.',
  },
  temp: {
    title: 'Create Temporary Schedule',
    desc: 'Replace the entire schedule for a given period of time.',
    name: 'Temporary Schedule',
    helperText:
      'Define a fixed shift-by-shift schedule to use for a given time frame.',
  },
  choose: {
    title: 'Choose Override Action',
    desc: 'Select the type of override you would like to apply to this schedule.',
    name: 'Choose',
    helperText: 'Determine which override action you want to take.',
  },
}
  add: {
    title: 'Temporarily Add User',
    desc: 'This will add a new shift for the selected user while the override is active. Existing shifts will remain unaffected.',
    name: 'Additional Coverage',
    helperText: 'Add an additional on-call user for a specified time.',
  },
  remove: {
    title: 'Temporarily Remove User',
    desc: 'This will remove (or split/shorten) shifts belonging to the selected user while the override is active.',
    name: 'Remove Coverage',
    helperText: "Remove one user's shifts for a specified time.",
  },
  replace: {
    title: 'Temporarily Replace User',
    desc: 'This will replace the selected user with another during any existing shifts while the override is active. No new shifts will be created. Only who is on-call will be changed.',
    name: "Cover Someone's Shifts",
    helperText: "Have a user take over another's shifts for a specified time.",
  },
  temp: {
    title: 'Create a temporary schedule',
    desc: 'Replace the entire schedule for a given period of time',
    name: 'Temporary Schedule',
    helperText:
      'Define a fixed shift-by-shift schedule to use for a specified time.',
  },
  choose: {
    title: 'Choose Override Action',
    desc: 'This will create a temporary override to the existing schedule.',
    name: 'Choose',
    helperText: 'This will determine which override action you want to take',
  },
}

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
