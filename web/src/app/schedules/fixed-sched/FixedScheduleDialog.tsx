import React, { useState, useEffect } from 'react'
import { useMutation } from 'react-apollo'
import gql from 'graphql-tag'
import { fieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import FixedScheduleForm from './FixedScheduleForm'
import { Shift, Value } from './sharedUtils'
import _ from 'lodash-es'

const mutation = gql`
  mutation($input: SetScheduleShiftsInput!) {
    setScheduleShifts(input: $input)
  }
`

interface FixedScheduleDialogProps {
  onClose: () => void
  scheduleID: string
  value?: Value
}

export default function FixedScheduleDialog({
  onClose,
  scheduleID,
  value: _value,
}: FixedScheduleDialogProps) {
  const edit = Boolean(_value)

  const mockStart = '2020-09-15T00:00:00.000Z'
  const mockEnd = '2020-09-30T00:00:00.000Z'
  const mockShift: Shift = {
    end: '2020-09-24T21:02:00.000Z',
    start: '2020-09-23T21:02:00.000Z',
    user: {
      label: 'Cierra Mayer',
      value: '307e25a3-2377-4b19-9fce-68c5569d2d12',
    },
  }
  const mockShifts: Shift[] = _.fill(Array(30), mockShift)

  // const [step, setStep] = useState(edit ? 1 : 0) // edit starting on step 2
  const [step, setStep] = useState(1) // edit starting on step 2
  const [value, setValue] = useState({
    start: mockStart || (_value?.start ?? ''),
    end: mockEnd || (_value?.end ?? ''),
    shifts: mockShifts || (_value?.shifts ?? []),
  })

  const [submit, { loading, error, data }] = useMutation(mutation, {
    variables: {
      input: {
        scheduleID,
        start: value.start,
        end: value.end,
        shifts: value.shifts.map((shift: Shift) => ({
          start: shift.start,
          end: shift.end,
          userID: shift.user?.value,
        })),
      },
    },
  })

  const fieldErrs = fieldErrors(error)
  const stepOneErrs = fieldErrs.some((e) => ['start', 'end'].includes(e.field))

  // array.fill fn?
  const stepTwoErrs = fieldErrs.some((e) =>
    ['summary', 'details'].includes(e.field),
  )

  useEffect(() => {
    if (stepOneErrs) setStep(0)
    else if (stepTwoErrs) setStep(1)
  }, [stepOneErrs, stepTwoErrs])

  const isComplete = data && !loading && !error
  return (
    <FormDialog
      fullScreen
      disableGutters
      title='Define a Fixed Schedule'
      primaryActionLabel={isComplete ? 'Done' : null}
      onClose={onClose}
      loading={loading}
      form={
        <FixedScheduleForm
          scheduleID={scheduleID}
          activeStep={step}
          setStep={setStep}
          edit={edit}
          value={value}
          onChange={(newValue: any) => setValue(newValue)}
          disabled={loading}
          errors={fieldErrors(error)}
        />
      }
      onSubmit={() => (isComplete ? onClose() : submit())}
      onNext={step === 2 ? null : () => setStep(step + 1)}
      onBack={(edit ? step === 1 : step === 0) ? null : () => setStep(step - 1)}
    />
  )
}
