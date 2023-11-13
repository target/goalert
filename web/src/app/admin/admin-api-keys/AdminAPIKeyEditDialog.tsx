import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreateGQLAPIKeyInput, GQLAPIKey } from '../../../schema'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

// query for updating api key which accepts UpdateGQLAPIKeyInput
const updateGQLAPIKeyQuery = gql`
  mutation UpdateGQLAPIKey($input: UpdateGQLAPIKeyInput!) {
    updateGQLAPIKey(input: $input)
  }
`
// query for getting existing API Key information
const query = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
      description
      expiresAt
      query
      role
    }
  }
`
export default function AdminAPIKeyEditDialog(props: {
  onClose: (param: boolean) => void
  apiKeyID: string
}): React.ReactNode {
  const { apiKeyID, onClose } = props
  const [{ fetching, data, error }] = useQuery({
    query,
  })
  const key: GQLAPIKey | null =
    data?.gqlAPIKeys?.find((d: GQLAPIKey) => d.id === apiKeyID) || null
  const [apiKeyActionStatus, apiKeyAction] = useMutation(updateGQLAPIKeyQuery)
  const [apiKeyInput, setAPIKeyInput] = useState<CreateGQLAPIKeyInput | null>(
    null,
  )

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />
  // handles form on submit event, based on the action type (edit, create) it will send the necessary type of parameter
  // token is also being set here when create action is used
  const handleOnSubmit = (): void => {
    apiKeyAction(
      {
        input: {
          name: apiKeyInput?.name,
          description: apiKeyInput?.description,
          id: apiKeyID,
        },
      },
      { additionalTypenames: ['GQLAPIKey'] },
    ).then((result) => {
      if (result.error) return

      onClose(false)
    })
  }

  if (fetching || key === null) {
    return <Spinner />
  }

  return (
    <FormDialog
      title='Update API Key'
      loading={apiKeyActionStatus.fetching}
      errors={nonFieldErrors(apiKeyActionStatus.error)}
      onClose={onClose}
      onSubmit={handleOnSubmit}
      form={
        <AdminAPIKeyForm
          errors={fieldErrors(apiKeyActionStatus.error)}
          onChange={setAPIKeyInput}
          value={apiKeyInput || key}
        />
      }
    />
  )
}
