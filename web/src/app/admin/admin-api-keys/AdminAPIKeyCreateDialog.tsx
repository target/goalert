import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import CopyText from '../../util/CopyText'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreateGQLAPIKeyInput } from '../../../schema'
import { CheckCircleOutline as SuccessIcon } from '@mui/icons-material'
import { DateTime } from 'luxon'
import { Grid, Typography, FormHelperText } from '@mui/material'

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

function AdminAPIKeyToken(props: { token: string }): React.ReactNode {
  return (
    <Grid item xs={12}>
      <Typography>
        <CopyText title={props.token} value={props.token} placement='bottom' />
      </Typography>
      <FormHelperText>
        Please copy and save the token as this is the only time you'll be able
        to view it.
      </FormHelperText>
    </Grid>
  )
}

export default function AdminAPIKeyCreateDialog(props: {
  onClose: () => void
}): React.ReactNode {
  const [value, setValue] = useState<CreateGQLAPIKeyInput>({
    name: '',
    description: '',
    expiresAt: DateTime.utc().plus({ days: 7 }).toISO(),
    allowedFields: [],
    role: 'user',
  })
  const [status, createKey] = useMutation(newGQLAPIKeyQuery)
  const token = status.data?.createGQLAPIKey?.token || null

  // handles form on submit event, based on the action type (edit, create) it will send the necessary type of parameter
  // token is also being set here when create action is used
  const handleOnSubmit = (): void => {
    createKey(
      {
        input: {
          name: value.name,
          description: value.description,
          allowedFields: value.allowedFields,
          expiresAt: value.expiresAt,
          role: value.role,
        },
      },
      { additionalTypenames: ['GQLAPIKey'] },
    )
  }

  return (
    <FormDialog
      title={
        token ? (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <SuccessIcon sx={{ marginRight: (theme) => theme.spacing(1) }} />
            <Typography>Success!</Typography>
          </div>
        ) : (
          'Create New API Key'
        )
      }
      subTitle={token ? 'Your API key has been created!' : ''}
      loading={status.fetching}
      errors={nonFieldErrors(status.error)}
      onClose={() => {
        props.onClose()
      }}
      onSubmit={token ? props.onClose : handleOnSubmit}
      alert={!!token}
      disableBackdropClose={!!token}
      form={
        token ? (
          <AdminAPIKeyToken token={token} />
        ) : (
          <AdminAPIKeyForm
            errors={fieldErrors(status.error)}
            value={value}
            onChange={setValue}
            create
          />
        )
      }
    />
  )
}
