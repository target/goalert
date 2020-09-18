import React, { ReactNode } from 'react'
import SwipeableViews from 'react-swipeable-views'
import { virtualize, bindKeyboard } from 'react-swipeable-views-utils'
import { FormContainer } from '../../forms'
import { FieldError } from '../../util/errutil'
import SchedulesTimesStep from './ScheduleTimesStep'
import AddShiftsStep from './AddShiftsStep'
import ReviewStep from './ReviewStep'
import SuccessStep from './SuccessStep'

// allows changing the index programatically
const VirtualizeAnimatedViews = bindKeyboard(virtualize(SwipeableViews))

interface FixedScheduleFormProps {
  activeStep: number
  setStep: (step: number) => void
  value: any
  onChange: (val: any) => any
  disabled: boolean
  errors: FieldError[]
}

export default function FixedScheduleForm({
  activeStep,
  setStep,
  ...rest
}: FixedScheduleFormProps) {
  const bodyStyle = {
    display: 'flex',
    justifyContent: 'center', // horizontal align
    height: '100%',
    width: '100%',
  }

  const containerStyle = {
    width: '35%', // ensures form fields don't shrink down too small
    marginTop: '5%', // slightly lower below dialog title toolbar
  }

  interface SlideRenderer {
    index: number
    key: number
  }
  function renderSlide({ index, key }: SlideRenderer): ReactNode {
    switch (index) {
      case 0:
        return <SchedulesTimesStep key={key} />
      case 1:
        return (
          <AddShiftsStep
            key={key}
            value={rest.value}
            onChange={rest.onChange}
          />
        )
      case 2:
        return <ReviewStep key={key} />
      case 3:
        return <SuccessStep key={key} />
      default:
        return null
    }
  }

  return (
    <div style={bodyStyle}>
      <div style={containerStyle}>
        <FormContainer optionalLabels {...rest}>
          <VirtualizeAnimatedViews
            index={activeStep}
            onChangeIndex={(i: number) => setStep(i)}
            slideRenderer={renderSlide}
          />
        </FormContainer>
      </div>
    </div>
  )
}
