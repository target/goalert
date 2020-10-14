import React, { useState, ReactNode } from 'react'
import { useMutation } from 'react-apollo'
import gql from 'graphql-tag'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { Shift } from './sharedUtils'
import _ from 'lodash-es'
import { FormContainer } from '../../forms'
import { bindKeyboard, virtualize } from 'react-swipeable-views-utils'
import SwipeableViews from 'react-swipeable-views'
import TempSchedAddShiftsStep from './TempSchedAddShiftsStep'
import TempSchedTimesStep from './TempSchedTimesStep'
import { SetTemporaryScheduleInput } from '../../../schema'
// allows changing the index programatically
const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

const mutation = gql`
  mutation($input: SetTemporaryScheduleInput!) {
    setTemporarySchedule(input: $input)
  }
`

type TempScheduleDialogProps = {
  onClose: () => void
  value?: SetTemporaryScheduleInput
}

export default function TempSchedDialog({
  onClose,
  value: _value,
}: TempScheduleDialogProps): JSX.Element {
  const edit = Boolean(_value)

  const [step, setStep] = useState(edit ? 1 : 0) // edit starting on step 2
  const [value, setValue] = useState({
    scheduleID: _value?.scheduleID ?? '',
    start: _value?.start ?? '',
    end: _value?.end ?? '',
    shifts: (_value?.shifts ?? []).map((s) =>
      _.pick(s, 'start', 'end', 'userID'),
    ),
  })

  const [submit, { loading, error, data }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: value,
    },
  })

  const isComplete = data && !loading && !error

  type SlideRenderer = {
    index: number
    key: number
  }
  function renderSlide({ index, key }: SlideRenderer): ReactNode {
    switch (index) {
      case 0:
        return (
          <TempSchedTimesStep
            key={key}
            stepText='STEP 1 OF 2'
            scheduleID={value.scheduleID}
          />
        )
      case 1:
        return (
          <TempSchedAddShiftsStep
            key={key}
            value={value.shifts}
            onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
            stepText={edit ? '' : 'STEP 2 OF 2'}
            scheduleID={value.scheduleID}
            start={value.start}
            end={value.end}
          />
        )
      default:
        return null
    }
  }

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const errs = nonFieldErrs.concat(fieldErrs)

  return (
    <FormDialog
      fullScreen
      disableGutters
      title='Define a Temporary Schedule'
      primaryActionLabel={isComplete ? 'Done' : null}
      onClose={onClose}
      loading={loading}
      errors={errs}
      form={
        <FormContainer
          optionalLabels
          disabled={loading}
          value={value}
          onChange={(newValue: SetTemporaryScheduleInput) => setValue(newValue)}
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
      onNext={step === 1 ? null : () => setStep(step + 1)}
      onBack={(edit ? step === 1 : step === 0) ? null : () => setStep(step - 1)}
    />
  )
}
