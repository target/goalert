import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreatedGQLAPIKey, GQLAPIKey } from '../../../schema'
import Spinner from '../../loading/components/Spinner'

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
// query for updating api key which accepts UpdateGQLAPIKeyInput
const updateGQLAPIKeyQuery = gql`
  mutation UpdateGQLAPIKey($input: UpdateGQLAPIKeyInput!) {
    updateGQLAPIKey(input: $input)
  }
`

export default function AdminAPIKeysActionDialog(props: {
  onClose: (param: boolean) => void
  setToken: (token: CreatedGQLAPIKey) => void
  setReloadFlag: (inc: number) => void
  onTokenDialogClose: (prama: boolean) => void
  create: boolean
  apiKey: GQLAPIKey
  setAPIKey: (param: GQLAPIKey) => void
  setSelectedAPIKey: (param: GQLAPIKey) => void
}): JSX.Element {
  let query = updateGQLAPIKeyQuery
  const { create, apiKey, setAPIKey, setSelectedAPIKey } = props

  if (props.create) {
    query = newGQLAPIKeyQuery
  }

  const [apiKeyActionStatus, apiKeyAction] = useMutation(query)
  const [allowFieldsError, setAllowFieldsError] = useState(true)
  const { fetching, data, error } = apiKeyActionStatus
  let fieldErrs = fieldErrors(error)
  // handles form on submit event, based on the action type (edit, create) it will send the necessary type of parameter
  // token is also being set here when create action is used
  const handleOnSubmit = (): void => {
    const updateKey = {
      name: apiKey.name,
      description: apiKey.description,
      id: apiKey.id,
    }

    const createKey = {
      name: apiKey.name,
      description: apiKey.description,
      allowedFields: apiKey.allowedFields,
      expiresAt: apiKey.expiresAt,
      role: apiKey.role,
    }

    apiKeyAction({
      input: create ? createKey : updateKey,
    }).then((result) => {
      if (!result.error) {
        props.setReloadFlag(Math.random())
        props.onClose(false)

        if (props.create) {
          props.setToken(result.data.createGQLAPIKey)
          props.onTokenDialogClose(true)
        } else {
          setSelectedAPIKey(apiKey)
        }
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
    <FormDialog
      title={props.create ? 'New API Key' : 'Update API Key'}
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
          disabled={fetching}
          value={apiKey}
          onChange={setAPIKey}
          allowFieldsError={allowFieldsError}
          setAllowFieldsError={setAllowFieldsError}
          create={props.create}
        />
      }
    />
  )
}
