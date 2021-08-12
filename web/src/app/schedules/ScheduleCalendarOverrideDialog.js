import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import { DateTime } from 'luxon'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import useOverrideNotices from './useOverrideNotices'
import { ScheduleCalendarOverrideForm } from './ScheduleCalendarOverrideForm'
const copyText = {
  add: {
    title: 'Temporarily Add a User',
    desc: 'This will add a new shift for the selected user, while the override is active. Existing shifts will remain unaffected.',
  },
  remove: {
    title: 'Temporarily Remove a User',
    desc: 'This will remove (or split/shorten) shifts belonging to the selected user, while the override is active.',
  },
  replace: {
    title: 'Temporarily Replace a User',
    desc: 'This will replace the selected user with another during any existing shifts, while the override is active. No new shifts will be created, only who is on-call will be changed.',
  },
  choose: {
    title: 'Choose an override action',
    desc: 'This will...',
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
  const [step, setStep] = useState(0)
  const [value, setValue] = useState({
    addUserID: '',
    removeUserID: '',
    start: DateTime.local().startOf('hour').toISO(),
    end: DateTime.local().startOf('hour').plus({ hours: 8 }).toISO(),
    variant: 'choose',
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

  const onNext = () => {
    setStep(step + 1)
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={step === 0 ? copyText.choose.title : copyText[value.variant].title}
      subTitle={
        step === 0 ? copyText.choose.desc : copyText[value.variant].desc
      }
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialogue
      loading={loading}
      form={
        <ScheduleCalendarOverrideForm
          scheduleID={props.scheduleID}
          activeStep={step}
          value={value}
          onChange={(newValue) => setValue(newValue)}
          disabled={loading}
          errors={fieldErrors(error)}
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
  onChooseOverrideType: p.func,
}
