import React from 'react'
import { gql } from '@apollo/client'

import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'

const updateQuery = gql`
  query ($id: ID!) {
    service(id: $id) {
      id
      labels {
        key
        value
      }
    }
  }
`

const mutation = gql`
  mutation ($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`

export default function ServiceLabelDeleteDialog(props) {
  const { labelKey, onClose, serviceID } = props
  function renderDialog(commit, mutStatus) {
    const { loading, error } = mutStatus
    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the label: ${labelKey}`}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={onClose}
        onSubmit={() => {
          const input = {
            key: labelKey,
            value: '',
            target: {
              type: 'service',
              id: serviceID,
            },
          }
          return commit({
            variables: {
              input,
            },
          })
        }}
      />
    )
  }

  function renderMutation() {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={onClose}
        update={(cache) => {
          const { service } = cache.readQuery({
            query: updateQuery,
            variables: { id: serviceID },
          })
          cache.writeQuery({
            query: updateQuery,
            variables: { id: serviceID },
            data: {
              service: {
                ...service,
                labels: (service.labels || []).filter(
                  (l) => l.key !== labelKey,
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

  return renderMutation()
}

ServiceLabelDeleteDialog.propTypes = {
  serviceID: p.string.isRequired,
  labelKey: p.string.isRequired,
  onClose: p.func,
}
