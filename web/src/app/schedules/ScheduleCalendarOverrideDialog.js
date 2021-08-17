import React, { useContext, useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'
import { ScheduleCalendarOverrideForm } from './ScheduleCalendarOverrideForm'
import { ScheduleCalendarContext } from './ScheduleDetails'

export const variantDetails = {
  add: {
    title: 'Temporarily Add a User',
    desc: 'This will add a new shift for the selected user, while the override is active. Existing shifts will remain unaffected.',
    name: 'Add',
    helperText: 'This will add a user to the schedule',
  },
  remove: {
    title: 'Temporarily Remove a User',
    desc: 'This will remove (or split/shorten) shifts belonging to the selected user, while the override is active.',
    name: 'Remove',
    helperText: 'This will remove a user from the schedule',
  },
  replace: {
    title: 'Temporarily Replace a User',
    desc: 'This will replace the selected user with another during any existing shifts, while the override is active. No new shifts will be created, only who is on-call will be changed.',
    name: 'Replace',
    helperText: 'This will replace a user from the schedule',
  },
  temp: {
    title: 'Create a temporary schedule',
    desc: 'Replace the entire schedule for a given period of time',
    name: 'Temporary Schedule',
    helperText: 'Replace the entire schedule for a given period of time',
  },
  choose: {
    title: 'Choose an override action',
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

export default function ScheduleCalendarOverrideDialog(props) {
  const { variantOptions = ['replace', 'remove', 'add', 'temp'] } = props
  const [step, setStep] = useState(0)
  const [value, setValue] = useState({
    addUserID: '',
    removeUserID: '',
    start: DateTime.local().startOf('hour').toISO(),
    end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
    variant: variantOptions[0] || 'remove', // Default to first variant on the list. Otherwise 'remove'.
    ...props.defaultValue,
  })

  const { onNewTempSched } = useContext(ScheduleCalendarContext)

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

  const onNext = () => {
    if (value.variant === 'temp') {
      onNewTempSched()
      props.onClose()
    } else {
      setStep(step + 1)
    }
  }

  const handleVariantChange = (newValue) => {
    setValue(newValue)
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={
        step === 0
          ? variantDetails.choose.title
          : variantDetails[value.variant].title
      }
      subTitle={
        step === 0
          ? variantDetails.choose.desc
          : variantDetails[value.variant].desc
      }
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialogue
      loading={loading}
      form={
        <ScheduleCalendarOverrideForm
          scheduleID={props.scheduleID}
          activeStep={step}
          value={value}
          onChange={handleVariantChange}
          disabled={loading}
          errors={fieldErrors(error)}
          variantOptions={variantOptions}
        />
      }
      onSubmit={() => mutate()}
      onNext={step < 1 ? onNext : null}
      onBack={step > 0 ? () => setStep(step - 1) : null}
    />
  )
}

ScheduleCalendarOverrideDialog.defaultProps = {
  defaultValue: {},
}

ScheduleCalendarOverrideDialog.propTypes = {
  scheduleID: p.string.isRequired,
  onClose: p.func,
  removeUserReadOnly: p.bool,
  defaultValue: p.shape({
    addUserID: p.string,
    removeUserID: p.string,
    start: p.string,
    end: p.string,
  }),
  variantOptions: p.arrayOf(p.string),
}
