import React from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'

import p from 'prop-types'

import { nonFieldErrors } from '../util/errutil'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query($id: ID!) {
    integrationKey(id: $id) {
      id
      name
      serviceID
    }
  }
`

const updateQuery = gql`
  query($id: ID!) {
    service(id: $id) {
      id
      integrationKeys {
        id
        name
      }
    }
  }
`

const mutation = gql`
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function IntegrationKeyDeleteDialog(props) {
  const { loading, error, data } = useQuery(query, {
    pollInterval: 0,
    variables: { id: props.integrationKeyID },
  })

  const [deleteKey, deleteKeyStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
    update: (cache) => {
      const { service } = cache.readQuery({
        query: updateQuery,
        variables: { id: data.integrationKey.serviceID },
      })

      cache.writeQuery({
        query: updateQuery,
        variables: { id: data.integrationKey.serviceID },
        data: {
          service: {
            ...service,
            integrationKeys: (service.integrationKeys || []).filter(
              (key) => key.id !== props.integrationKeyID,
            ),
          },
        },
      })
    },
  })

  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the integration key: ${data.integrationKey.name}`}
      caption='This will prevent the creation of new alerts using this integration key. If you wish to re-enable, a NEW integration key must be created and may require additional reconfiguration of the alert source.'
      loading={deleteKeyStatus.loading}
      errors={nonFieldErrors(deleteKeyStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        const input = [
          {
            type: 'integrationKey',
            id: props.integrationKeyID,
          },
        ]
        return deleteKey({
          variables: {
            input,
          },
        })
      }}
    />
  )
}

IntegrationKeyDeleteDialog.propTypes = {
  integrationKeyID: p.string.isRequired,
  onClose: p.func,
}
