import React, { useState } from 'react'
import { gql } from '@apollo/client'

import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm from './IntegrationKeyForm'

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
const query = gql`
  query ($serviceID: ID!) {
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
  const { serviceID, onClose } = props

  const renderDialog = (commit, status) => {
    const { loading, error } = status
    return (
      <FormDialog
        maxWidth='sm'
        title='Create New Integration Key'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: { ...value, serviceID: serviceID },
            },
          })
        }}
        form={
          <IntegrationKeyForm
            errors={fieldErrors(error)}
            disabled={loading}
            value={value}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  return (
    <Mutation
      mutation={mutation}
      onCompleted={onClose}
      update={(cache, { data: { createIntegrationKey } }) => {
        const { service } = cache.readQuery({
          query,
          variables: { serviceID: serviceID },
        })
        cache.writeQuery({
          query,
          variables: { serviceID: serviceID },
          data: {
            service: {
              ...service,
              integrationKeys: (service.integrationKeys || []).concat(
                createIntegrationKey,
              ),
            },
          },
        })
      }}
    >
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

IntegrationKeyCreateDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
