/* eslint-disable @typescript-eslint/no-use-before-define */
import React, { useState, useEffect, useRef } from 'react'
import { gql, useMutation } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { Alert, Box, Typography, Link } from '@mui/material'

const generateOAuthURLMutation = gql`
  mutation GenerateIMAPOAuthURL($input: GenerateIMAPOAuthURLInput!) {
    generateIMAPOAuthURL(input: $input) {
      authURL
      state
    }
  }
`

const exchangeOAuthCodeMutation = gql`
  mutation ExchangeIMAPOAuthCode($input: ExchangeIMAPOAuthCodeInput!) {
    exchangeIMAPOAuthCode(input: $input) {
      refreshToken
      accessToken
      expiresAt
    }
  }
`

interface IMAPOAuthDialogProps {
  clientID: string
  clientSecret: string
  onClose: () => void
  onTokenReceived: (refreshToken: string) => void
}

export default function IMAPOAuthDialog(
  props: IMAPOAuthDialogProps,
): JSX.Element {
  const oauthStateRef = useRef<string>('')
  const [oauthWindow, setOAuthWindow] = useState<Window | null>(null)
  const [step, setStep] = useState<'authorizing' | 'success' | 'error'>(
    'authorizing',
  )
  const [errorMessage, setErrorMessage] = useState<string>('')
  const [refreshToken, setRefreshToken] = useState<string>('')

  const [generateURLStatus, generateURL] = useMutation(generateOAuthURLMutation)
  const [exchangeCodeStatus, exchangeCode] = useMutation(
    exchangeOAuthCodeMutation,
  )

  useEffect(() => {
    // Start OAuth flow immediately when dialog opens
    handleStartOAuth()
  }, [])

  const handleStartOAuth = (): void => {
    const redirectURL = `${window.location.origin}/imap-oauth-callback`

    generateURL({
      input: {
        clientID: props.clientID,
        clientSecret: props.clientSecret,
        redirectURL,
      },
    }).then((result) => {
      if (!result.error && result.data?.generateIMAPOAuthURL) {
        const { authURL, state } = result.data.generateIMAPOAuthURL
        oauthStateRef.current = state

        // Open OAuth consent in popup
        const width = 600
        const height = 700
        const left = window.screen.width / 2 - width / 2
        const top = window.screen.height / 2 - height / 2
        const popup = window.open(
          authURL,
          'oauth',
          `width=${width},height=${height},left=${left},top=${top}`,
        )
        setOAuthWindow(popup)

        // Listen for OAuth callback
        window.addEventListener('message', handleOAuthCallback)
      } else {
        setStep('error')
        setErrorMessage(result.error?.message || 'Failed to generate OAuth URL')
      }
    })
  }

  const handleOAuthCallback = (event: MessageEvent): void => {
    if (event.origin !== window.location.origin) {
      return
    }

    if (event.data.type === 'imap-oauth-error') {
      setStep('error')
      setErrorMessage(event.data.error || 'OAuth authorization failed')
      if (oauthWindow) {
        oauthWindow.close()
      }
      window.removeEventListener('message', handleOAuthCallback)
      return
    }

    if (event.data.type === 'imap-oauth-code') {
      const { code } = event.data

      // Close popup
      if (oauthWindow) {
        oauthWindow.close()
      }

      // Exchange code for token
      const redirectURL = `${window.location.origin}/imap-oauth-callback`
      exchangeCode({
        input: {
          code,
          state: oauthStateRef.current,
          redirectURL,
        },
      }).then((result) => {
        if (!result.error && result.data?.exchangeIMAPOAuthCode) {
          const { refreshToken: token } = result.data.exchangeIMAPOAuthCode
          setRefreshToken(token)
          setStep('success')
          props.onTokenReceived(token)
        } else {
          setStep('error')
          setErrorMessage(
            result.error?.message || 'Failed to exchange authorization code',
          )
        }

        // Clean up listener
        window.removeEventListener('message', handleOAuthCallback)
      })
    }
  }

  const getDialogContent = (): JSX.Element => {
    if (step === 'authorizing') {
      return (
        <Box sx={{ textAlign: 'center', py: 4 }}>
          <Typography variant='h6' gutterBottom>
            Authorizing with Google...
          </Typography>
          <Typography variant='body2' color='text.secondary' sx={{ mb: 2 }}>
            A popup window has been opened. Please authorize GoAlert to access
            your Gmail account.
          </Typography>
          <Alert severity='info'>
            <Typography variant='body2'>
              If the popup was blocked, click{' '}
              <Link
                href='#'
                onClick={(e) => {
                  e.preventDefault()
                  handleStartOAuth()
                }}
              >
                here
              </Link>{' '}
              to try again.
            </Typography>
          </Alert>
          {(exchangeCodeStatus.error || generateURLStatus.error) && (
            <Alert severity='error' sx={{ mt: 2 }}>
              {(exchangeCodeStatus.error || generateURLStatus.error)?.message}
            </Alert>
          )}
        </Box>
      )
    }

    if (step === 'error') {
      return (
        <Box sx={{ py: 2 }}>
          <Alert severity='error'>
            <Typography variant='body1' gutterBottom>
              <strong>Authorization Failed</strong>
            </Typography>
            <Typography variant='body2'>{errorMessage}</Typography>
          </Alert>
          <Typography variant='body2' color='text.secondary' sx={{ mt: 2 }}>
            Please check your OAuth credentials and try again.
          </Typography>
        </Box>
      )
    }

    return (
      <Box sx={{ py: 2 }}>
        <Alert severity='success' sx={{ mb: 2 }}>
          <Typography variant='body1' gutterBottom>
            <strong>✓ Authorization Successful!</strong>
          </Typography>
          <Typography variant='body2'>
            Your OAuth refresh token has been obtained and automatically
            populated in the configuration form.
          </Typography>
        </Alert>
        <Alert severity='info'>
          <Typography
            variant='body2'
            sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}
          >
            {refreshToken.substring(0, 50)}...
          </Typography>
        </Alert>
      </Box>
    )
  }

  return (
    <FormDialog
      title='Obtain OAuth Refresh Token'
      onClose={props.onClose}
      loading={generateURLStatus.fetching || exchangeCodeStatus.fetching}
      onSubmit={() => props.onClose()}
      form={getDialogContent()}
      primaryActionLabel={
        step === 'success' ? 'Done' : step === 'error' ? 'Close' : undefined
      }
      disableSubmit={step === 'authorizing'}
    />
  )
}
