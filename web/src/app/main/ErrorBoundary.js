import React from 'react'
import { GenericError } from '../error-pages'

export default class ErrorBoundary extends React.PureComponent {
  state = { hasError: false }

  componentDidCatch(error, info) {
    // Display fallback UI
    this.setState({ hasError: true })
    console.error(error, info)
    // TODO: log and/or call some API
  }

  render() {
    if (this.state.hasError) {
      return <GenericError />
    }
    return this.props.children
  }
}
