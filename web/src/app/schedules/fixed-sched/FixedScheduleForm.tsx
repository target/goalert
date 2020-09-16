import React from 'react'
import { FormContainer } from '../../forms'
import { FieldError } from '../../util/errutil'
import SchedulesTimesStep from './ScheduleTimesStep'
import AddShiftsStep from './AddShiftsStep'
import ReviewStep from './ReviewStep'
import SuccessStep from './SuccessStep'

interface FixedScheduleFormProps {
  activeStep: number
  value: any
  onChange: (val: any) => any
  disabled: boolean
  errors: FieldError[]
}

export default function FixedScheduleForm(props: FixedScheduleFormProps) {
  const { activeStep, ...otherProps } = props

  const bodyStyle = {
    display: 'flex',
    alignItems: 'center', // vertical align
    justifyContent: 'center', // horizontal align
    height: '100%',
    width: '100%',
  }

  const containerStyle = {
    width: '35%', // ensures form fields don't shrink down too small
    marginBottom: '10%', // slightly raise higher than center of screen
  }

  return (
    <div style={bodyStyle}>
      <div style={containerStyle}>
        <FormContainer optionalLabels {...otherProps}>
          {activeStep === 0 && <SchedulesTimesStep />}
          {activeStep === 1 && <AddShiftsStep />}
          {activeStep === 2 && <ReviewStep />}
          {activeStep === 3 && <SuccessStep />}
        </FormContainer>
      </div>
    </div>
  )
}
