import React from 'react'
import Loader from 'react-loadable'
import Spinner from '../loading/components/Spinner'
import { DEFAULT_SPIN_DELAY_MS, DEFAULT_SPIN_WAIT_MS } from '../config'

function Loading({ error }) {
  if (error) console.error(error)
  return null
}

function LoadingSpinner({ error }) {
  if (error) console.error(error)

  // no delay since it's already handled below
  return <Spinner delayMs={0} />
}

// onDemand is a convenience wrapper for Loadable (from `react-loadable`)
//
// The first, and only required argument, is a loading function that returns a promise.
//
// The second (options) is passed directly to Loader, with the exception of the `wait` option.
// If `wait` is specified, it will impose a minimum delay.
//
// For example, if `delay` is 200 ms, and wait is 1000 ms:
// - A module that loads in 100ms will never show a loading spinner
// - A module that loads in 300ms will show a loading spinner after 200ms, and for 1000ms after
// - A module that loads in 1200ms will show a loading spinner after 200ms, and for 1000ms after
//
// This allows the prevention of flicker in varying network conditions.
export default function onDemand(_load, options = {}) {
  if (!('delay' in options)) {
    options.delay = DEFAULT_SPIN_DELAY_MS
  }
  if (!('wait' in options)) {
    options.wait = DEFAULT_SPIN_WAIT_MS
  }
  if (!('loading' in options)) {
    options.loading = options.spin ? LoadingSpinner : Loading
  }

  const { wait, spin, ...rest } = options

  const loader = () => {
    let waitPromise
    setTimeout(() => {
      waitPromise = new Promise((resolve) => {
        setTimeout(resolve, wait)
      })
    }, options.delay)

    return _load().then((res) => {
      if (!waitPromise) {
        return res
      }

      return waitPromise.then(() => res)
    })
  }

  const l = Loader({
    ...rest,
    loader,
  })
  setTimeout(() => l.preload(), 3000)
  return l
}
