import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreateGQLAPIKeyInput, CreatedGQLAPIKey } from '../../../schema'
import Spinner from '../../loading/components/Spinner'

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
  const [createAPIKeyStatus, createAPIKey] = useMutation(newGQLAPIKeyQuery)
  const [allowFieldsError, setAllowFieldsError] = useState(true)
  const { loading, data, error } = createAPIKeyStatus
  let fieldErrs = fieldErrors(error)
  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const handleOnSubmit = () => {
    createAPIKey({
      input: key,
    }).then((result: any) => {
      if (!result.error) {
        props.setToken(result.data.createGQLAPIKey)
        props.onClose(false)
        props.onTokenDialogClose(true)
        props.setReloadFlag(Math.random())
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
      title='New API Key'
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
          value={key}
          onChange={setKey}
          allowFieldsError={allowFieldsError}
          setAllowFieldsError={setAllowFieldsError}
          create
        />
      }
    />
  )
}
