import React from 'react'
import p from 'prop-types'
import { FormContainer } from '../forms'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import ChooseOverrideForm from './ChooseOverrideForm'

export function ScheduleCalendarOverrideForm(props) {
  const { activeStep, ...otherProps } = props

  return (
    <FormContainer optionalLabels {...otherProps}>
      {activeStep === 0 && (
        <ChooseOverrideForm
          scheduleID={props.scheduleID} // todo
          disabled={props.disabled}
          value={props.value}
          errors={props.errors}
          onChange={props.onChange}
          removeUserReadOnly={props.removeUserReadOnly}
          variantOptions={props.variantOptions}
        />
      )}
      {activeStep === 1 && (
        <ScheduleOverrideForm
          add={props.value.variant !== 'remove'}
          remove={props.value.variant !== 'add'}
          scheduleID={props.scheduleID} // todo
          disabled={props.disabled}
          errors={props.errors}
          value={props.value}
          onChange={props.onChange}
          removeUserReadOnly={props.removeUserReadOnly}
        />
      )}
    </FormContainer>
  )
}

ScheduleCalendarOverrideForm.propTypes = {
  activeStep: p.number.isRequired,
  value: p.shape({
    summary: p.string,
    details: p.string,
    serviceIDs: p.arrayOf(p.string),
  }),
  errors: p.arrayOf(
    p.shape({
      field: p.oneOf(['summary', 'details', 'serviceID']).isRequired,
      message: p.string.isRequired,
    }),
  ),
  scheduleID: p.string.isRequired,
  variantOptions: p.arrayOf(p.string),
}
