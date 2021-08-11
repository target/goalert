import React from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { GenericError } from '../error-pages'

export default function ErrorBoundaryWrapper({ children }) {
  return (
    <ErrorBoundary FallbackComponent={GenericError}>{children}</ErrorBoundary>
  )
}
