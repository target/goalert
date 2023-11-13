import React from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { nonFieldErrors } from '../util/errutil'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query ($id: ID!) {
    integrationKey(id: $id) {
      id
      name
      serviceID
    }
  }
`

const mutation = gql`
  mutation ($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function IntegrationKeyDeleteDialog(props: {
  integrationKeyID: string
  onClose: () => void
}): React.ReactNode {
  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { id: props.integrationKeyID },
  })

  const [deleteKeyStatus, deleteKey] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!fetching && !deleteKeyStatus.fetching && data?.integrationKey === null) {
    return (
      <FormDialog
        alert
        title='No longer exists'
        onClose={() => props.onClose()}
        subTitle='That integration key does not exist or is already deleted.'
      />
    )
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the integration key: ${data?.integrationKey?.name}`}
      caption='This will prevent the creation of new alerts using this integration key. If you wish to re-enable, a NEW integration key must be created and may require additional reconfiguration of the alert source.'
      loading={deleteKeyStatus.fetching}
      errors={nonFieldErrors(deleteKeyStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        const input = [
          {
            type: 'integrationKey',
            id: props.integrationKeyID,
          },
        ]
        return deleteKey(
          { input },
          { additionalTypenames: ['IntegrationKey'] },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }}
    />
  )
}
