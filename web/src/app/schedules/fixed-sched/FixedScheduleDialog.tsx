import React, { useState, useEffect, ReactNode } from 'react'
import { useMutation } from 'react-apollo'
import gql from 'graphql-tag'
import { fieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { Shift, Value } from './sharedUtils'
import _ from 'lodash-es'
import { FormContainer } from '../../forms'
import { bindKeyboard, virtualize } from 'react-swipeable-views-utils'
import SwipeableViews from 'react-swipeable-views'
import AddShiftsStep from './AddShiftsStep'
import ReviewStep from './ReviewStep'
import ScheduleTimesStep from './ScheduleTimesStep'
// allows changing the index programatically
const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

const mutation = gql`
  mutation($input: SetScheduleShiftsInput!) {
    setScheduleShifts(input: $input)
  }
`

type FixedScheduleDialogProps = {
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

  // MOCK DATA
  // const mockStart = '2020-09-15T00:00:00.000Z'
  // const mockEnd = '2020-09-30T00:00:00.000Z'
  // const mockShift: Shift = {
  //   end: '2020-09-24T21:02:00.000Z',
  //   start: '2020-09-23T21:02:00.000Z',
  //   userID: '307e25a3-2377-4b19-9fce-68c5569d2d12',
  // }
  // const mockShifts: Shift[] = _.fill(Array(1), mockShift)
  // const [step, setStep] = useState(1) // edit starting on step 2
  // const [value, setValue] = useState({
  //   start: mockStart,
  //   end: mockEnd,
  //   shifts: mockShifts,
  // })

  // NOT MOCK DATA
  const [step, setStep] = useState(edit ? 1 : 0) // edit starting on step 2
  const [value, setValue] = useState({
    start: _value?.start ?? '',
    end: _value?.end ?? '',
    shifts: _value?.shifts ?? [],
  })

  const [submit, { loading, error, data }] = useMutation(mutation, {
    variables: {
      input: {
        ...value,
        scheduleID,
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

  type SlideRenderer = {
    index: number
    key: number
  }
  function renderSlide({ index, key }: SlideRenderer): ReactNode {
    switch (index) {
      case 0:
        return (
          <ScheduleTimesStep
            key={key}
            stepText='STEP 1 of 3'
            scheduleID={scheduleID}
          />
        )
      case 1:
        return (
          <AddShiftsStep
            key={key}
            value={value.shifts}
            onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
            stepText={edit ? 'STEP 1 of 2' : 'STEP 2 of 3'}
            scheduleID={scheduleID}
            start={value.start}
            end={value.end}
          />
        )
      case 2:
        if (step !== 2) return null
        return (
          <ReviewStep
            key={key}
            value={value}
            stepText={edit ? 'STEP 2 of 2' : 'STEP 3 of 3'}
          />
        )
      default:
        return null
    }
  }

  return (
    <FormDialog
      fullScreen
      disableGutters
      title='Define a Fixed Schedule'
      primaryActionLabel={isComplete ? 'Done' : null}
      onClose={onClose}
      loading={loading}
      form={
        <FormContainer
          optionalLabels
          disabled={loading}
          value={value}
          onChange={(newValue: Value) => setValue(newValue)}
          errors={fieldErrors(error)}
        >
          <VirtualizeAnimatedViews
            index={step}
            onChangeIndex={(i: number) => setStep(i)}
            slideRenderer={renderSlide}
            disabled // disables slides from changing outside of action buttons
            containerStyle={{ height: '100%' }}
            style={{ height: '100%' }}
          />
        </FormContainer>
      }
      onSubmit={() => (isComplete ? onClose() : submit())}
      onNext={step === 2 ? null : () => setStep(step + 1)}
      onBack={(edit ? step === 1 : step === 0) ? null : () => setStep(step - 1)}
    />
  )
}
