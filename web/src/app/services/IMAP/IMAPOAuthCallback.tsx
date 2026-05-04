import React, { useEffect } from 'react'
import { Box, Typography, CircularProgress } from '@mui/material'

export default function IMAPOAuthCallback(): JSX.Element {
  useEffect(() => {
    // Extract code from URL
    const params = new URLSearchParams(window.location.search)
    const code = params.get('code')
    const error = params.get('error')

    if (error) {
      window.opener?.postMessage(
        {
          type: 'imap-oauth-error',
          error,
        },
        window.location.origin,
      )
      setTimeout(() => window.close(), 2000)
      return
    }

    if (code) {
      // Send code back to parent window
      window.opener?.postMessage(
        {
          type: 'imap-oauth-code',
          code,
        },
        window.location.origin,
      )

      // Close window after short delay
      setTimeout(() => window.close(), 1000)
    }
  }, [])

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        gap: 2,
      }}
    >
      <CircularProgress />
      <Typography variant='h6'>Authorization successful!</Typography>
      <Typography variant='body2' color='text.secondary'>
        This window will close automatically...
      </Typography>
    </Box>
  )
}
