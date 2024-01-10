import React from 'react'
import { ConfigProvider } from '../util/RequireConfig'
import { DestTypeProvider } from '../util/useDestinationTypes'
import { Provider as URQLProvider } from 'urql'
import { client as urqlClient } from '../urql'
import { StyledEngineProvider } from '@mui/material'
import { ThemeProvider } from '../theme/themeConfig'
import { ErrorBoundary } from 'react-error-boundary'

import { Settings } from 'luxon'
import { DecoratorFunction } from '@storybook/types'
import { ReactRenderer } from '@storybook/react'
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

type Func = DecoratorFunction<ReactRenderer, object>
type FuncParams = Parameters<Func>

export default function DefaultDecorator(
  Story: FuncParams[0],
  args: FuncParams[1],
): ReturnType<Func> {
  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider
        mode={
          args?.globals?.backgrounds?.value === '#333333' ? 'dark' : 'light'
        }
      >
        <URQLProvider value={urqlClient}>
          <ConfigProvider>
            <DestTypeProvider>
              <ErrorBoundary fallbackRender={fallbackRender}>
                <Story />
              </ErrorBoundary>
            </DestTypeProvider>
          </ConfigProvider>
        </URQLProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  )
}
