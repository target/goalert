import React from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { nonFieldErrors } from '../util/errutil'
import { GenericError } from '../error-pages'
import FormDialog from '../dialogs/FormDialog'
import {
  Checkbox,
  FormControl,
  FormControlLabel,
  FormHelperText,
} from '@mui/material'

const query = gql`
  query ($id: ID!) {
    integrationKey(id: $id) {
      id
      name
      serviceID
      externalSystemName
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
}): React.JSX.Element {
  const [{ error, data }] = useQuery({
    query,
    variables: { id: props.integrationKeyID },
  })
  const extSystemName = data?.integrationKey?.externalSystemName || ''
  const [confirmed, setConfirmed] = React.useState(!extSystemName) // only require confirmation if external system name is present
  const [confirmError, setConfirmError] = React.useState(false)

  const [deleteKeyStatus, deleteKey] = useMutation(mutation)

  if (error) return <GenericError error={error.message} />

  if (!data?.integrationKey) {
    return (
      <FormDialog
        alert
        title='No longer exists'
        onClose={() => props.onClose()}
        subTitle='That integration key does not exist or is already deleted.'
      />
    )
  }

  let form = null
  if (data?.integrationKey?.externalSystemName) {
    form = (
      <FormControl style={{ width: '100%' }} error={confirmError}>
        <FormControlLabel
          control={
            <Checkbox
              checked={confirmed}
              onChange={(e) => {
                setConfirmed(e.target.checked)
                setConfirmError(false)
              }}
            />
          }
          label='I understand the consequences of deleting this key'
        />
        <FormHelperText>
          {confirmError ? (
            'Please confirm'
          ) : (
            <React.Fragment>
              Deleting this key may break integrations with&nbsp;
              {extSystemName}
            </React.Fragment>
          )}
        </FormHelperText>
      </FormControl>
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
      form={form}
      onSubmit={() => {
        if (!confirmed) {
          setConfirmError(true)
          return
        }
        setConfirmError(false)

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
