import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'

interface ScheduleOverrideCreateDialogProps {
  scheduleID: string
  variant: 'add' | 'remove' | 'replace'
  onClose: () => void
  removeUserReadOnly?: boolean
}

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

const mutation = gql`
  mutation ($input: CreateUserOverrideInput!) {
    createUserOverride(input: $input) {
      id
    }
  }
`

export default function ScheduleOverrideCreateDialog({
  scheduleID,
  variant,
  onClose,
  removeUserReadOnly,
}: ScheduleOverrideCreateDialogProps): JSX.Element {
  const [value, setValue] = useState({
    addUserID: '',
    removeUserID: '',
    start: DateTime.local().startOf('hour').toISO(),
    end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
  })

  const notices = useOverrideNotices(scheduleID, value)

  const [{ fetching, error }, commit] = useMutation(mutation)

  return (
    <FormDialog
      onClose={onClose}
      title={variantDetails[variant].title}
      subTitle={variantDetails[variant].desc}
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialog
      onSubmit={() =>
        commit(
          {
            input: {
              ...value,
              scheduleID,
            },
          },
          { additionalTypenames: ['UserOverrideConnection', 'Schedule'] },
        ).then((result) => {
          if (!result.error) onClose()
        })
      }
      form={
        <ScheduleOverrideForm
          add={variant !== 'remove'}
          remove={variant !== 'add'}
          scheduleID={scheduleID}
          disabled={fetching}
          errors={fieldErrors(error)}
          value={value}
          onChange={(newValue) => setValue(newValue)}
          removeUserReadOnly={removeUserReadOnly}
        />
      }
    />
  )
}
