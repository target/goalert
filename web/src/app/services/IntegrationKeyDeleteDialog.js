import React from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

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
  function renderDialog(data, commit, mutStatus) {
    const { loading, error } = mutStatus

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the integration key: ${data.name}`}
        caption='This will prevent the creation of new alerts using this integration key. If you wish to re-enable, a NEW integration key must be created and may require additional reconfiguration of the alert source.'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => {
          const input = [
            {
              type: 'integrationKey',
              id: props.integrationKeyID,
            },
          ]
          return commit({
            variables: {
              input,
            },
          })
        }}
      />
    )
  }

  function renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={props.onClose}
        update={(cache) => {
          const { service } = cache.readQuery({
            query: updateQuery,
            variables: { id: data.serviceID },
          })
          cache.writeQuery({
            query: updateQuery,
            variables: { id: data.serviceID },
            data: {
              service: {
                ...service,
                integrationKeys: (service.integrationKeys || []).filter(
                  (key) => key.id !== props.integrationKeyID,
                ),
              },
            },
          })
        }}
      >
        {(commit, status) => renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  function renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: props.integrationKeyID }}
        render={({ data }) => renderMutation(data.integrationKey)}
      />
    )
  }

  return renderQuery()
}

IntegrationKeyDeleteDialog.propTypes = {
  integrationKeyID: p.string.isRequired,
  onClose: p.func,
}
