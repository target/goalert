import React, { useState, useEffect } from 'react'
import { useMutation } from 'react-apollo'
import gql from 'graphql-tag'
import { fieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import FixedScheduleForm from './FixedScheduleForm'

const mutation = gql`
  mutation($input: SetScheduleShiftsInput!) {
    setScheduleShifts(input: $input)
  }
`

interface CreateFixedScheduleDialogProps {
  open: boolean
  onClose: () => void
  scheduleID: string
}

export default function CreateFixedScheduleDialog({
  open,
  onClose,
  scheduleID,
}: CreateFixedScheduleDialogProps) {
  const [step, setStep] = useState(0)
  const [value, setValue] = useState({
    // todo: fill in defaults
    scheduleID,
    start: '',
    end: '',
    shifts: [], // { userID, start, end } fields
  })

  const [submit, { loading, error, data }] = useMutation(mutation)

  const fieldErrs = fieldErrors(error)
  const stepOneErrs = fieldErrs.some((e) =>
    ['summary', 'details'].includes(e.field),
  )
  const stepTwoErrs = fieldErrs.some((e) =>
    ['summary', 'details'].includes(e.field),
  )

  useEffect(() => {
    if (stepOneErrs) setStep(0)
    else if (stepTwoErrs) setStep(1)
  }, [stepOneErrs, stepTwoErrs])

  const isComplete = data && !loading && !error
  return (
    open && (
      <FormDialog
        fullScreen
        disableGutters
        title='Define a Fixed Schedule Adjustment'
        primaryActionLabel={isComplete ? 'Done' : null}
        onClose={onClose}
        loading={loading}
        form={
          <FixedScheduleForm
            activeStep={step}
            value={value}
            onChange={(newValue: any) => setValue(newValue)}
            disabled={loading}
            errors={fieldErrors(error)}
          />
        }
        onSubmit={() => (isComplete ? onClose() : submit())}
        onNext={step === 3 ? null : () => setStep(step + 1)}
        onBack={step === 0 ? null : () => setStep(step - 1)}
      />
    )
  )
}
