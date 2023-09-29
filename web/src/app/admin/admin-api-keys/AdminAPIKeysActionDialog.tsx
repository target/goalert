import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreatedGQLAPIKey, GQLAPIKey } from '../../../schema'
import Spinner from '../../loading/components/Spinner'

const newGQLAPIKeyQuery = gql`
  mutation CreateGQLAPIKey($input: CreateGQLAPIKeyInput!) {
    createGQLAPIKey(input: $input) {
      id
      token
    }
  }
`

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
  const { loading, data, error } = apiKeyActionStatus
  let fieldErrs = fieldErrors(error)
  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const handleOnSubmit = () => {
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
    }).then((result: any) => {
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

  if (loading && !data) {
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
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={() => {
        props.onClose(false)
      }}
      onSubmit={handleOnSubmit}
      disableBackdropClose
      form={
        <AdminAPIKeyForm
          errors={fieldErrs}
          disabled={loading}
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
