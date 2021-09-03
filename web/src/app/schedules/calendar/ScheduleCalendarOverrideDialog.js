import React, { useContext, useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../../dialogs/FormDialog'
import { DateTime } from 'luxon'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import useOverrideNotices from '../useOverrideNotices'
import { ScheduleCalendarOverrideForm } from './ScheduleCalendarOverrideForm'
import { ScheduleCalendarContext } from '../ScheduleDetails'
import { variantDetails } from '../ScheduleOverrideCreateDialog'

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
    ...props.defaultValue,
  })

  const [activeVariant, setActiveVariant] = useState(
    variantOptions[0] || 'replace',
  )

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
    if (activeVariant === 'temp') {
      onNewTempSched()
      props.onClose()
    } else {
      setStep(step + 1)
    }
  }

  const handleChange = (newValue) => {
    setValue(newValue)
  }

  const handleVariantChange = (newVariant) => {
    setActiveVariant(newVariant)
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title={
        step === 0
          ? variantDetails.choose.title
          : variantDetails[activeVariant].title
      }
      subTitle={
        step === 0
          ? variantDetails.choose.desc
          : variantDetails[activeVariant].desc
      }
      errors={nonFieldErrors(error)}
      notices={notices} // create and edit dialogue
      loading={loading}
      form={
        <ScheduleCalendarOverrideForm
          scheduleID={props.scheduleID}
          activeStep={step}
          value={value}
          onChange={handleChange}
          onVariantChange={handleVariantChange}
          activeVariant={activeVariant}
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
