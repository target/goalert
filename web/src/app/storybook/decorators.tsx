import React from 'react'
import { ConfigProvider } from '../util/RequireConfig'
import { Provider as URQLProvider } from 'urql'
import { client as urqlClient } from '../urql'
import { StyledEngineProvider } from '@mui/material'
import { ThemeProvider } from '../theme/themeConfig'
import { ErrorBoundary } from 'react-error-boundary'

import { Settings } from 'luxon'
Settings.throwOnInvalid = true

interface Error {
  message: string
}

type FallbackProps = {
  error: Error
  resetErrorBoundary: () => void
}

function fallbackRender({
  error,
  resetErrorBoundary,
}: FallbackProps): React.ReactNode {
  return (
    <div role='alert'>
      <p>Thrown error:</p>
      <pre style={{ color: 'red' }}>{error.message}</pre>
      <button onClick={resetErrorBoundary}>Retry</button>
    </div>
  )
}

type StoryArgs = {
  globals: {
    backgrounds: {
      value: string
    }
  }
}

export default function DefaultDecorator(
  Story: React.JSXElementConstructor<unknown>,
  args: StoryArgs,
): React.ReactNode {
  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider
        mode={
          args?.globals?.backgrounds?.value === '#333333' ? 'dark' : 'light'
        }
      >
        <URQLProvider value={urqlClient}>
          <ConfigProvider>
            <ErrorBoundary fallbackRender={fallbackRender}>
              <Story />
            </ErrorBoundary>
          </ConfigProvider>
        </URQLProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  )
}
