import React, { useState } from 'react'
import { useMutation, gql } from 'urql'
import IntegrationKeyForm, { Value } from './IntegrationKeyForm'
import { Redirect } from 'wouter'
import FormDialog from '../dialogs/FormDialog'
import { useExpFlag } from '../util/useExpFlag'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

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
  const [createKeyStatus, createKey] = useMutation(mutation)
  const hasUnivKeysFlag = useExpFlag('univ-keys')
  let caption
  if (hasUnivKeysFlag && value?.type === 'universal') {
    caption = 'Submit to configure universal key rules'
  }

  console.log(createKeyStatus)

  if (createKeyStatus?.data?.createIntegrationKey?.type === 'universal') {
    return (
      <Redirect
        to={`/services/${serviceID}/integration-keys/${createKeyStatus.data.createIntegrationKey.id}`}
      />
    )
  }

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      subTitle={caption}
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
