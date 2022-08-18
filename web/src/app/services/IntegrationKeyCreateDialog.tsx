import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm from './IntegrationKeyForm'
import { CreateIntegrationKeyInput } from '../../schema'

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
  const [value, setValue] = useState<CreateIntegrationKeyInput>({
    name: '',
    type: 'generic',
    serviceID: props.serviceID,
  })

  const [createIntegrationKeyStatus, createIntegrationKey] =
    useMutation(mutation)

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      loading={createIntegrationKeyStatus.fetching}
      errors={nonFieldErrors(createIntegrationKeyStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        createIntegrationKey(
          { input: { ...value } },
          { additionalTypenames: ['IntegrationKey'] },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }}
      form={
        <IntegrationKeyForm
          errors={fieldErrors(createIntegrationKeyStatus.error)}
          disabled={createIntegrationKeyStatus.fetching}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
