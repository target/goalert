import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyCreateForm from './AdminAPIKeyCreateForm'
import { CreateGQLAPIKeyInput, CreatedGQLAPIKey } from '../../../schema'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

const newGQLAPIKeyQuery = gql`
  mutation CreateGQLAPIKey($input: CreateGQLAPIKeyInput!) {
    createGQLAPIKey(input: $input) {
      id
      token
    }
  }
`

export default function AdminAPIKeysCreateDialog(props: {
  onClose: (param: boolean) => void
  setToken: (token: CreatedGQLAPIKey) => void
  setReloadFlag: (inc: number) => void
  onTokenDialogClose: (prama: boolean) => void
}): JSX.Element {
  const [key, setKey] = useState<CreateGQLAPIKeyInput>({
    name: '',
    description: '',
    allowedFields: [],
    expiresAt: '',
    role: 'unknown',
  })
  const [createAPIKey, createAPIKeyStatus] = useMutation(newGQLAPIKeyQuery, {
    onCompleted: (data) => {
      props.setToken(data.createGQLAPIKey)
      props.onClose(false)
      props.onTokenDialogClose(true)
      props.setReloadFlag(Math.random())
    },
  })
  const [ allowFieldsError, setAllowFieldsError] = useState<string>('')
  const { loading, data, error } = createAPIKeyStatus
  const fieldErrs = fieldErrors(error)
  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const handleOnSubmit = () => {
    if (key.allowedFields.length > 0) {
      createAPIKey({
        variables: {
          input: key,
        },
      }).then((result) => {
        if (!result.errors) {
          return result
        }
      })
    } else {
      setAllowFieldsError('Please choose at least one Allowed Fields.')
    }
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  return (
    <FormDialog
      title='New API Key'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={handleOnSubmit}
      disableBackdropClose
      form={
        <AdminAPIKeyCreateForm
          errors={fieldErrs}
          disabled={loading}
          value={key}
          onChange={setKey}
          allowFieldsError={allowFieldsError}
        />
      }
    />
  )
}
