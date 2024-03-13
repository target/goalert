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
  const [value, setValue] = useState<Value | null>({
    name: '',
    type: 'generic',
    summary: '',
    details: '',
    filters: [],
    deduplication: [],
  })
  const { serviceID, onClose } = props

  const [createKeyStatus, createKey] = useMutation(mutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      loading={createKeyStatus.fetching}
      errors={nonFieldErrors(createKeyStatus.error)}
      onClose={onClose}
      onSubmit={(): void => {
        createKey(
          { input: { serviceID, ...value } },
          { additionalTypenames: ['IntegrationKey', 'Service'] },
        ).then(onClose)
      }}
      form={
        <IntegrationKeyForm
          errors={fieldErrors(createKeyStatus.error)}
          disabled={createKeyStatus.fetching}
          value={value}
          onChange={(value): void => setValue(value)}
        />
      }
    />
  )
}
