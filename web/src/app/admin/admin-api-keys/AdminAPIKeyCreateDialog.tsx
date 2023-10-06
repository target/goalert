import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreatedGQLAPIKey, CreateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeysTokenDialog from './AdminAPIKeyTokenDialog'
import Spinner from '../../loading/components/Spinner'
import { DateTime } from 'luxon'
// query for creating new api key which accepts CreateGQLAPIKeyInput param
// return token created upon successfull transaction
const newGQLAPIKeyQuery = gql`
  mutation CreateGQLAPIKey($input: CreateGQLAPIKeyInput!) {
    createGQLAPIKey(input: $input) {
      id
      token
    }
  }
`

export default function AdminAPIKeyCreateDialog(props: {
  onClose: (param: boolean) => void
}): JSX.Element {
  const { onClose } = props
  const [apiKey, setAPIKey] = useState<CreateGQLAPIKeyInput>({
    name: '',
    description: '',
    expiresAt: DateTime.utc().plus({ days: 7 }).toISO(),
    allowedFields: [],
    role: 'user',
  })
  const [apiKeyActionStatus, apiKeyAction] = useMutation(newGQLAPIKeyQuery)
  const { fetching, data, error } = apiKeyActionStatus
  const [tokenDialogClose, onTokenDialogClose] = useState<boolean>(true)
  const [token, setToken] = useState<CreatedGQLAPIKey>({} as CreatedGQLAPIKey)
  let fieldErrs = fieldErrors(error)
  // handles form on submit event, based on the action type (edit, create) it will send the necessary type of parameter
  // token is also being set here when create action is used
  const handleOnSubmit = (): void => {
    apiKeyAction(
      {
        input: {
          name: apiKey.name,
          description: apiKey.description,
          allowedFields: apiKey.allowedFields,
          expiresAt: apiKey.expiresAt,
          role: apiKey.role,
        },
      },
      { additionalTypenames: ['GQLAPIKey'] },
    ).then((result) => {
      if (!result.error) {
        setToken(result.data.createGQLAPIKey)
        onTokenDialogClose(false)
      }
    })
  }

  if (fetching && !data) {
    return <Spinner />
  }

  if (error) {
    fieldErrs = fieldErrs.map((err) => {
      return err
    })
  }

  return (
    <React.Fragment>
      {tokenDialogClose ? (
        <FormDialog
          title='New API Key'
          loading={fetching}
          errors={nonFieldErrors(error)}
          onClose={() => {
            props.onClose(false)
          }}
          onSubmit={handleOnSubmit}
          disableBackdropClose
          form={
            <AdminAPIKeyForm
              errors={fieldErrs}
              value={apiKey}
              onChange={setAPIKey}
              create
            />
          }
        />
      ) : (
        <AdminAPIKeysTokenDialog
          value={token}
          onClose={() => {
            onTokenDialogClose(true)
            onClose(false)
          }}
        />
      )}
    </React.Fragment>
  )
}
