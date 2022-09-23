import React, { useState } from 'react'
import { useMutation, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm, { Value } from './IntegrationKeyForm'

const mutation = gql`
  mutation ($input: CreateIntegrationKeyInput!) {
    createIntegrationKey(input: $input) {
      id
      name
      type
      href
    }
  }
`

export default function IntegrationKeyCreateDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)
  const { serviceID, onClose } = props

  const [createIntegrationKeyStatus, createIntegrationKey] =
    useMutation(mutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      loading={createIntegrationKeyStatus.fetching}
      errors={nonFieldErrors(createIntegrationKeyStatus.error)}
      onClose={onClose}
      onSubmit={(): void => {
        createIntegrationKey(
          { input: { serviceID, ...value } },
          { additionalTypenames: ['IntegrationKey'] },
        ).then(onClose)
      }}
      form={
        <IntegrationKeyForm
          errors={fieldErrors(createIntegrationKeyStatus.error)}
          disabled={createIntegrationKeyStatus.fetching}
          value={
            value || {
              name: '',
              type: 'generic',
            }
          }
          onChange={(value): void => setValue(value)}
        />
      }
    />
  )
}
