import React, { ReactNode } from 'react'
import SwipeableViews from 'react-swipeable-views'
import { virtualize, bindKeyboard } from 'react-swipeable-views-utils'
import { FormContainer } from '../../forms'
import { FieldError } from '../../util/errutil'
import SchedulesTimesStep from './ScheduleTimesStep'
import AddShiftsStep from './AddShiftsStep'
import ReviewStep from './ReviewStep'

// allows changing the index programatically
const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

interface FixedScheduleFormProps {
  scheduleID: string
  activeStep: number
  setStep: (step: number) => void
  edit?: boolean
  value: any
  onChange: (val: any) => any
  disabled: boolean
  errors: FieldError[]
}

export default function FixedScheduleForm({
  scheduleID,
  activeStep,
  setStep,
  edit,
  ...rest
}: FixedScheduleFormProps) {
  interface SlideRenderer {
    index: number
    key: number
  }
  function renderSlide({ index, key }: SlideRenderer): ReactNode {
    switch (index) {
      case 0:
        return <SchedulesTimesStep key={key} value={rest.value} />
      case 1:
        return (
          <AddShiftsStep
            key={key}
            value={rest.value}
            onChange={rest.onChange}
            edit={edit}
          />
        )
      case 2:
        return (
          <ReviewStep
            key={key}
            activeStep={activeStep}
            scheduleID={scheduleID}
            value={rest.value}
            edit={edit}
          />
        )
      default:
        return null
    }
  }

  return (
    <FormContainer optionalLabels {...rest}>
      <VirtualizeAnimatedViews
        index={activeStep}
        onChangeIndex={(i: number) => setStep(i)}
        slideRenderer={renderSlide}
        disabled // disables slides from changing outside of action buttons
        slideStyle={{ overflow: 'hidden' }}
      />
    </FormContainer>
  )
}
