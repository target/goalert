import React, { useEffect, useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import CopyText from '../../util/CopyText'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import AdminAPIKeyForm from './AdminAPIKeyForm'
import { CreateGQLAPIKeyInput, GQLAPIKey } from '../../../schema'
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

const fromExistingQuery = gql`
  query {
    gqlAPIKeys {
      id
      name
      description
      role
      query
      createdAt
      expiresAt
    }
  }
`

function AdminAPIKeyToken(props: { token: string }): React.ReactNode {
  return (
    <Grid item xs={12}>
      <Typography sx={{ mb: 3 }}>
        <CopyText title={props.token} value={props.token} placement='bottom' />
      </Typography>
      <FormHelperText>
        Please copy and save the token as this is the only time you'll be able
        to view it.
      </FormHelperText>
    </Grid>
  )
}

// nextName will increment the number (if any) at the end of the name.
function nextName(name: string): string {
  const match = name.match(/^(.*?)\s*(\d+)?$/)
  if (!match) return name
  const [, base, num] = match
  if (!num) return `${base} 2`
  return `${base} ${parseInt(num) + 1}`
}

export default function AdminAPIKeyCreateDialog(props: {
  onClose: () => void
  fromID?: string
}): React.ReactNode {
  const [value, setValue] = useState<CreateGQLAPIKeyInput>({
    name: '',
    description: '',
    expiresAt: DateTime.utc().plus({ days: 7 }).toISO(),
    query: '',
    role: 'user',
  })
  const [status, createKey] = useMutation(newGQLAPIKeyQuery)
  const token = status.data?.createGQLAPIKey?.token || null
  const [{ data }] = useQuery({
    query: fromExistingQuery,
    pause: !props.fromID,
  })

  useEffect(() => {
    if (!data?.gqlAPIKeys?.length) return
    const from = data.gqlAPIKeys.find((k: GQLAPIKey) => k.id === props.fromID)
    if (!from) return

    const created = DateTime.fromISO(from.createdAt)
    const expires = DateTime.fromISO(from.expiresAt)

    const keyLifespan = expires.diff(created, 'days').days

    setValue({
      name: nextName(from.name),
      description: from.description,
      query: from.query,
      expiresAt: DateTime.utc().plus({ days: keyLifespan }).toISO(),
      role: from.role,
    })
  }, [data?.gqlAPIKeys])

  // handles form on submit event, based on the action type (edit, create) it will send the necessary type of parameter
  // token is also being set here when create action is used
  const handleOnSubmit = (): void => {
    createKey(
      {
        input: {
          name: value.name,
          description: value.description,
          query: value.query,
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
              marginTop: '8px',
            }}
          >
            <SuccessIcon sx={{ mr: (theme) => theme.spacing(1) }} />
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
