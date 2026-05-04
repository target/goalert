import React, { useState, ChangeEvent } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { FormContainer, FormField } from '../../forms'
import {
  FormControlLabel,
  Switch,
  Alert,
  Button,
  Typography,
} from '@mui/material'
import IMAPOAuthDialog from './IMAPOAuthDialog'

const createMutation = gql`
  mutation ($input: CreateServiceIMAPConfigInput!) {
    createServiceIMAPConfig(input: $input) {
      serviceID
    }
  }
`

const updateMutation = gql`
  mutation ($input: UpdateServiceIMAPConfigInput!) {
    updateServiceIMAPConfig(input: $input)
  }
`

const testMutation = gql`
  mutation ($input: CreateServiceIMAPConfigInput!) {
    testIMAPConnection(input: $input)
  }
`

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id
      imapConfig {
        enabled
        host
        port
        username
        useTLS
        mailbox
        pollIntervalMinutes
        markAsRead
        deleteAfter
        includeHeaders
        includeFrom
        includeTo
        includeSubject
        includeBody
      }
    }
  }
`

interface Value {
  enabled: boolean
  host: string
  port: number
  username: string
  useTLS: boolean
  mailbox: string
  pollIntervalMinutes: number
  markAsRead: boolean
  deleteAfter: boolean
  includeHeaders: boolean
  includeFrom: boolean
  includeTo: boolean
  includeSubject: boolean
  includeBody: boolean
  oauthClientID?: string
  oauthClientSecret?: string
  oauthRefreshToken?: string
}

export default function ServiceIMAPConfigDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)
  const [testSuccess, setTestSuccess] = useState<boolean | null>(null)
  const [showOAuthDialog, setShowOAuthDialog] = useState<boolean>(false)

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })

  const isUpdate = data?.service?.imapConfig != null

  const [createStatus, create] = useMutation(createMutation)
  const [updateStatus, update] = useMutation(updateMutation)
  const [testStatus, testConnection] = useMutation(testMutation)

  const mutationStatus = isUpdate ? updateStatus : createStatus
  const mutate = isUpdate ? update : create

  const handleTestConnection = (): void => {
    setTestSuccess(null)
    testConnection({
      input: {
        serviceID: props.serviceID,
        // eslint-disable-next-line @typescript-eslint/no-use-before-define
        ...currentValue,
      },
    }).then((result) => {
      if (!result.error) {
        setTestSuccess(true)
      } else {
        setTestSuccess(false)
      }
    })
  }

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const defaultValue: Value = {
    enabled: data?.service?.imapConfig?.enabled ?? true,
    host: data?.service?.imapConfig?.host ?? 'imap.gmail.com',
    port: data?.service?.imapConfig?.port ?? 993,
    username: data?.service?.imapConfig?.username ?? '',
    useTLS: data?.service?.imapConfig?.useTLS ?? true,
    mailbox: data?.service?.imapConfig?.mailbox ?? 'INBOX',
    pollIntervalMinutes: data?.service?.imapConfig?.pollIntervalMinutes ?? 5,
    markAsRead: data?.service?.imapConfig?.markAsRead ?? false,
    deleteAfter: data?.service?.imapConfig?.deleteAfter ?? false,
    includeHeaders: data?.service?.imapConfig?.includeHeaders ?? false,
    includeFrom: data?.service?.imapConfig?.includeFrom ?? true,
    includeTo: data?.service?.imapConfig?.includeTo ?? true,
    includeSubject: data?.service?.imapConfig?.includeSubject ?? true,
    includeBody: data?.service?.imapConfig?.includeBody ?? true,
    oauthClientID: '',
    oauthClientSecret: '',
    oauthRefreshToken: '',
  }

  const currentValue = value || defaultValue

  const handleTokenReceived = (refreshToken: string): void => {
    setValue({
      ...currentValue,
      oauthRefreshToken: refreshToken,
    })
    setShowOAuthDialog(false)
  }

  return (
    <React.Fragment>
      {showOAuthDialog && (
        <IMAPOAuthDialog
          clientID={currentValue.oauthClientID || ''}
          clientSecret={currentValue.oauthClientSecret || ''}
          onClose={() => setShowOAuthDialog(false)}
          onTokenReceived={handleTokenReceived}
        />
      )}
      <FormDialog
        maxWidth='md'
        title={isUpdate ? 'Edit IMAP Configuration' : 'Configure IMAP'}
        loading={mutationStatus.fetching}
        errors={nonFieldErrors(mutationStatus.error)}
        onClose={props.onClose}
        onSubmit={() =>
          mutate(
            {
              input: {
                serviceID: props.serviceID,
                ...currentValue,
              },
            },
            { additionalTypenames: ['Service'] },
          ).then((result) => {
            if (!result.error) {
              props.onClose()
            }
          })
        }
        form={
          <FormContainer
            value={currentValue}
            errors={fieldErrors(mutationStatus.error)}
            disabled={mutationStatus.fetching}
            onChange={(newValue: Value) => setValue(newValue)}
            optionalLabels
          >
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <Alert severity='info'>
                  Configure IMAP settings for this service to monitor incoming
                  emails. OAuth credentials are optional and can be set per
                  service to monitor multiple Gmail accounts.
                </Alert>
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.enabled}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({ ...currentValue, enabled: e.target.checked })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Enable IMAP Email Monitoring'
                />
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Gmail Address'
                  name='username'
                  required
                  hint='The Gmail email address to monitor'
                />
              </Grid>

              <Grid item xs={12} sm={8}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='IMAP Host'
                  name='host'
                  required
                />
              </Grid>

              <Grid item xs={12} sm={4}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Port'
                  name='port'
                  type='number'
                  required
                />
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Mailbox'
                  name='mailbox'
                  required
                  hint='The mailbox folder to monitor (usually INBOX)'
                />
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='Poll Interval (minutes)'
                  name='pollIntervalMinutes'
                  type='number'
                  required
                  hint='How often to check for new emails (1-1440 minutes)'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.useTLS}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({ ...currentValue, useTLS: e.target.checked })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Use TLS'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.markAsRead}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          markAsRead: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Mark processed emails as read'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.deleteAfter}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          deleteAfter: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Delete emails after processing'
                />
              </Grid>

              <Grid item xs={12}>
                <Alert severity='warning'>
                  OAuth 2.0 Credentials (Optional - for Gmail OAuth
                  authentication)
                </Alert>
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='OAuth Client ID'
                  name='oauthClientID'
                  hint='From Google Cloud Console - required to obtain refresh token'
                />
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='OAuth Client Secret'
                  name='oauthClientSecret'
                  type='password'
                  hint='From Google Cloud Console - required to obtain refresh token'
                />
              </Grid>

              <Grid item xs={12}>
                <Button
                  variant='outlined'
                  onClick={() => setShowOAuthDialog(true)}
                  disabled={
                    !currentValue.oauthClientID ||
                    !currentValue.oauthClientSecret
                  }
                  fullWidth
                >
                  Get OAuth Refresh Token
                </Button>
                {(!currentValue.oauthClientID ||
                  !currentValue.oauthClientSecret) && (
                  <Typography
                    variant='caption'
                    color='text.secondary'
                    sx={{ mt: 1, display: 'block' }}
                  >
                    Fill in Client ID and Secret above to enable this button
                  </Typography>
                )}
              </Grid>

              <Grid item xs={12}>
                <FormField
                  fullWidth
                  component={TextField}
                  label='OAuth Refresh Token'
                  name='oauthRefreshToken'
                  type='password'
                  hint='OAuth 2.0 refresh token for long-term access'
                />
              </Grid>

              <Grid item xs={12}>
                <Button
                  variant='outlined'
                  onClick={handleTestConnection}
                  disabled={
                    testStatus.fetching ||
                    !currentValue.username ||
                    !currentValue.host
                  }
                >
                  {testStatus.fetching ? 'Testing...' : 'Test Connection'}
                </Button>
                {testSuccess === true && (
                  <Alert severity='success' sx={{ mt: 1 }}>
                    Connection test successful!
                  </Alert>
                )}
                {testSuccess === false && (
                  <Alert severity='error' sx={{ mt: 1 }}>
                    Connection test failed. {testStatus.error?.message}
                  </Alert>
                )}
              </Grid>

              <Grid item xs={12}>
                <Alert severity='info'>
                  Alert Content Configuration - Choose which email fields to
                  include in created alerts
                </Alert>
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.includeHeaders}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          includeHeaders: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Include email headers (Date, Message-ID, etc.)'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.includeFrom}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          includeFrom: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Include From field'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.includeTo}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          includeTo: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Include To field'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.includeSubject}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          includeSubject: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Include Subject field'
                />
              </Grid>

              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={currentValue.includeBody}
                      onChange={(e: ChangeEvent<HTMLInputElement>) =>
                        setValue({
                          ...currentValue,
                          includeBody: e.target.checked,
                        })
                      }
                      disabled={mutationStatus.fetching}
                    />
                  }
                  label='Include email body'
                />
              </Grid>
            </Grid>
          </FormContainer>
        }
      />
    </React.Fragment>
  )
}
