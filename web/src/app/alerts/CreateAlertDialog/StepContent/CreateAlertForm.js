import React from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../../../forms'
import { CreateAlertInfo } from './CreateAlertInfo'
import { CreateAlertServiceSelect } from './CreateAlertServiceSelect'
import { CreateAlertConfirm } from './CreateAlertConfirm'

export function CreateAlertForm(props) {
  const { activeStep, ...otherProps } = props

  return (
    <FormContainer optionalLabels {...otherProps}>
      {activeStep === 0 && <CreateAlertInfo />}
      {activeStep === 1 && (
        <FormField
          required
          render={(props) => <CreateAlertServiceSelect {...props} />}
          name='serviceIDs'
        />
      )}
      {activeStep === 2 && <CreateAlertConfirm />}
    </FormContainer>
  )
}

CreateAlertForm.propTypes = {
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
}
