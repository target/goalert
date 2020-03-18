import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import ReactGA from 'react-ga'
import { Route } from 'react-router-dom'

let isInitialized = false

/*
example usage:
sendGAEvent({
  category: 'Service',
  action: action + '  Completed',
})
*/
export function sendGAEvent(eventProps) {
  if (isInitialized) ReactGA.event(eventProps)
}

class GoogleAnalytics extends Component {
  componentDidMount() {
    this.logPageChange(this.props.location.pathname, this.props.location.search)
  }

  componentDidUpdate({ location: prevLocation }) {
    const {
      location: { pathname, search },
    } = this.props
    const isDifferentPathname = pathname !== prevLocation.pathname
    const isDifferentSearch = search !== prevLocation.search

    if (isDifferentPathname || isDifferentSearch) {
      this.logPageChange(pathname, search)
    }
  }

  logPageChange(pathname, search = '') {
    const page = pathname + search
    const { location } = window
    ReactGA.set({
      page,
      location: `${location.origin}${page}`,
      ...this.props.options,
    })
    ReactGA.pageview(page)
  }

  render() {
    return null
  }
}

GoogleAnalytics.propTypes = {
  location: p.shape({
    pathname: p.string,
    search: p.string,
  }).isRequired,
  options: p.object,
}

const RouteTracker = () => <Route component={GoogleAnalytics} />

const init = (trackingID, options = {}) => {
  ReactGA.initialize(trackingID, {
    ...options,
  })

  isInitialized = true
  return isInitialized
}

export default {
  GoogleAnalytics,
  RouteTracker,
  init,
}
