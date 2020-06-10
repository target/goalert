import React, { useState } from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm from './IntegrationKeyForm'

const mutation = gql`
  mutation($input: CreateIntegrationKeyInput!) {
    createIntegrationKey(input: $input) {
      id
      name
      type
      href
    }
  }
`
const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id
      integrationKeys {
        id
        name
        type
        href
      }
    }
  }
`

export default function IntegrationKeyCreateDialog(props) {
  const [value, setValue] = useState({ name: '', type: 'generic' })

  const [createKey, createKeyStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
    update: (cache, { data: { createIntegrationKey } }) => {
      const { service } = cache.readQuery({
        query,
        variables: { serviceID: props.serviceID },
      })
      cache.writeData({
        query,
        variables: { serviceID: props.serviceID },
        data: {
          service: {
            ...service,
            integrationKeys: (service.integrationKeys || []).concat(
              createIntegrationKey,
            ),
          },
        },
      })
    },
  })

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Integration Key'
      loading={createKeyStatus.loading}
      errors={nonFieldErrors(createKeyStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        return createKey({
          variables: {
            input: { ...value, serviceID: props.serviceID },
          },
        })
      }}
      form={
        <IntegrationKeyForm
          errors={fieldErrors(createKeyStatus.error)}
          disabled={createKeyStatus.loading}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

IntegrationKeyCreateDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
