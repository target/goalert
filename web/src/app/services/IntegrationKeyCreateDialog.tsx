import React, { useState } from 'react'
import { useMutation, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm, { Value } from './IntegrationKeyForm'
import { Redirect } from 'wouter'

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
  const [goEdit, setGoEdit] = useState(false)
  const [value, setValue] = useState<Value | null>(null)
  const { serviceID, onClose } = props

  const [createKeyStatus, createKey] = useMutation(mutation)

  if (goEdit) {
    return <Redirect to='/rule-editor/11111111-1111-1111-111111111111' />
  }

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      loading={createKeyStatus.fetching}
      errors={nonFieldErrors(createKeyStatus.error)}
      onClose={onClose}
      onSubmit={(): void => {
        if (!value) return
        if (value.type === 'uik') {
          setGoEdit(true)
          return
        }
        createKey(
          { input: { serviceID, ...value } },
          { additionalTypenames: ['IntegrationKey', 'Service'] },
        ).then(onClose)
      }}
      form={
        <IntegrationKeyForm
          errors={fieldErrors(createKeyStatus.error)}
          disabled={createKeyStatus.fetching}
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
