import React from 'react'
import { FormContainer, FormField } from '../../../forms'
import { CreateAlertInfo } from './CreateAlertInfo'
import {
  CreateAlertServiceSelect,
  CreateAlertServiceSelectProps,
} from './CreateAlertServiceSelect'
import { CreateAlertConfirm } from './CreateAlertConfirm'
import { Value } from '../CreateAlertDialog'

// TODO: extend FormContainer once that file has been converted to typescript
interface CreateAlertFormProps {
  activeStep: number
  value: Value

  errors: Error[]

  onChange?: (newValue: Value) => void
  disabled?: boolean

  mapValue?: () => void
  mapOnChangeValue?: () => void

  // Enables functionality to remove an incoming value at it's index from
  // an array field if the new value is falsey.
  removeFalseyIdxs?: boolean
}

// TODO: remove this interface once FormContainer.js has been converted to TS
interface Error {
  message: string
  field: string
  helpLink?: string
}

export function CreateAlertForm({
  activeStep,
  ...otherProps
}: CreateAlertFormProps): JSX.Element {
  return (
    <FormContainer optionalLabels {...otherProps}>
      {activeStep === 0 && <CreateAlertInfo />}
      {activeStep === 1 && (
        <FormField
          required
          render={(props: CreateAlertServiceSelectProps) => (
            <CreateAlertServiceSelect {...props} />
          )}
          name='serviceIDs'
        />
      )}
      {activeStep === 2 && <CreateAlertConfirm />}
    </FormContainer>
  )
}
