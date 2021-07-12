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

export default function IntegrationKeyCreateDialog() {
  const [state, setState] = useState({
    value: { name: '', type: 'generic' },
    errors: [],
  })

  const renderDialog = (commit, status) => {
    const { loading, error } = status
    return (
      <FormDialog
        maxWidth='sm'
        title='Create New Integration Key'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: { ...state.value, serviceID: props.serviceID },
            },
          })
        }}
        form={
          <IntegrationKeyForm
            errors={fieldErrors(error)}
            disabled={loading}
            value={state.value}
            onChange={(value) => setState({ value })}
          />
        }
      />
    )
  }

  return (
    <Mutation
      mutation={mutation}
      onCompleted={props.onClose}
      update={(cache, { data: { createIntegrationKey } }) => {
        const { service } = cache.readQuery({
          query,
          variables: { serviceID: props.serviceID },
        })
        cache.writeQuery({
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
