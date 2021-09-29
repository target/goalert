import React, { useEffect } from 'react'
import { PropTypes as p } from 'prop-types'
import { event, set, pageview, initialize } from 'react-ga'
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
  if (isInitialized) event(eventProps)
}

function GoogleAnalytics(props) {
  const { pathname = '', search = '' } = props.location

  useEffect(() => {
    const page = pathname + search
    set({
      page,
      location: `${window.location.origin}${page}`,
      ...props.options,
    })
    pageview(page)
  }, [pathname, search])

  return null
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
  initialize(trackingID, {
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
