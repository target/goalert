import React, { useEffect } from 'react'
import p from 'prop-types'

export default function MountWatcher({
  children,
  onMount = () => {},
  onUnmount = () => {},
}) {
  useEffect(() => {
    onMount()
    return () => {
      onUnmount()
    }
  }, [])
  return children
}

MountWatcher.propTypes = {
  onMount: p.func,
  onUnmount: p.func,
}
