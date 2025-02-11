import React from 'react'
import { nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { gql, useMutation, useQuery } from 'urql'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { GQLAPIKey } from '../../../schema'

// query for deleting API Key which accepts API Key ID
const deleteGQLAPIKeyQuery = gql`
  mutation DeleteGQLAPIKey($id: ID!) {
    deleteGQLAPIKey(id: $id)
  }
`

// query for getting existing API Keys
const query = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
    }
  }
`

export default function AdminAPIKeyDeleteDialog(props: {
  apiKeyID: string
  onClose: (yes: boolean) => void
}): React.JSX.Element {
  const [{ fetching, data, error }] = useQuery({
    query,
  })
  const { apiKeyID, onClose } = props
  const [deleteAPIKeyStatus, deleteAPIKey] = useMutation(deleteGQLAPIKeyQuery)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const apiKeyName = data?.gqlAPIKeys?.find((d: GQLAPIKey) => {
    return d.id === apiKeyID
  })?.name

  function handleOnSubmit(): void {
    deleteAPIKey(
      {
        id: apiKeyID,
      },
      { additionalTypenames: ['GQLAPIKey'] },
    ).then((result) => {
      if (!result.error) onClose(true)
    })
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={'This will delete the API Key ' + apiKeyName + '.'}
      loading={deleteAPIKeyStatus.fetching}
      errors={nonFieldErrors(deleteAPIKeyStatus.error)}
      onClose={onClose}
      onSubmit={handleOnSubmit}
    />
  )
}
