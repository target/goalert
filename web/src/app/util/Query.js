import React from 'react'
import p from 'prop-types'
import { Query as ApolloQuery } from 'react-apollo'
import Spinner from '../loading/components/Spinner'
import { isEmpty } from 'lodash-es'
import { GenericError, ObjectNotFound } from '../error-pages/Errors'

import { POLL_ERROR_INTERVAL, POLL_INTERVAL } from '../config'

const hasNull = data =>
  isEmpty(data) || Object.keys(data).some(key => data[key] === null)

export function withQuery(
  query,
  mapQueryToProps,
  mapPropsToQueryProps = () => ({}),
) {
  return Component =>
    function WithQuery(componentProps) {
      return (
        <Query
          {...mapPropsToQueryProps(componentProps)}
          query={query}
          render={renderProps => (
            <Component {...componentProps} {...mapQueryToProps(renderProps)} />
          )}
        />
      )
    }
}

export default class Query extends React.PureComponent {
  static propTypes = {
    render: p.func.isRequired,

    // disable polling (for non-error states)
    noPoll: p.bool,

    // do not render an error or not-found message
    noError: p.bool,

    // do not render a spinner when loading
    noSpin: p.bool,

    // client will override the default (graphql2) client.
    client: p.object,

    // partialQuery will return the result from the cache instead
    // of a spinner, if possible.
    partialQuery: p.object,

    // override fetchPolicy, set to `cache-and-network` otherwise
    fetchPolicy: p.oneOf([
      'cache-first',
      'cache-and-network',
      'network-only',
      'cache-only',
      'no-cache',
    ]),
  }

  state = {
    spin: false,
  }

  renderSpinner() {
    return (
      <Spinner
        delayMs={200}
        waitMs={1500}
        onSpin={() => this.setState({ spin: true })}
        onReady={() => this.setState({ spin: false })}
      />
    )
  }

  renderResult = args => {
    if (this.state.spin) {
      if (this.props.noSpin)
        return this.props.render({ ...args, loading: true })
      return this.renderSpinner()
    }
    const { error, data, loading, startPolling: _startPolling } = args

    let startPolling = _startPolling
    if (new URLSearchParams(location.search).get('poll') === '0') {
      // global polling disable for debugging
      startPolling = () => {}
    }

    if (!hasNull(data) || this.props.skip) {
      if (!this.props.noPoll) startPolling(POLL_INTERVAL)
      return this.props.render(args)
    }

    if (!data && loading) {
      if (this.props.partialQuery) {
        try {
          const data = this.props.client.readQuery({
            query: this.props.partialQuery,
            variables: this.props.variables,
          })

          if (!hasNull(data))
            return this.props.render({ ...args, partial: true, data })
        } catch (e) {
          // wrap readQuery in try/catch...
          // https://github.com/apollographql/react-apollo/issues/1776#issuecomment-372237940
        }
      }

      if (this.props.noSpin)
        return this.props.render({ ...args, loading: true })
      return this.renderSpinner()
    }

    if (error) {
      const pol = this.props.fetchPolicy
      if (pol !== 'cache-only' && pol !== 'cache-first')
        startPolling(POLL_ERROR_INTERVAL)
      if (this.props.noError)
        return this.props.render({ ...args, error: error.message })
      return <GenericError error={error.message} />
    }

    if (this.props.noError)
      return this.props.render({ ...args, error: 'not found' })
    return <ObjectNotFound />
  }

  render() {
    const {
      // pull out our custom props
      render,
      noPoll,
      partialQuery,
      // and default-override ones
      client,
      fetchPolicy,
      ...rest
    } = this.props

    return (
      <ApolloQuery
        client={client}
        fetchPolicy={fetchPolicy || 'cache-and-network'}
        {...rest}
      >
        {this.renderResult}
      </ApolloQuery>
    )
  }
}
