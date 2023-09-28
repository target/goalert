import { useEffect } from 'react'
import p from 'prop-types'

export default function MountWatcher(props) {
  useEffect(() => {
    props.onMount()
    return () => props.onUnmount()
  }, [])

  return props.children
}

MountWatcher.propTypes = {
  onMount: p.func,
  onUnmount: p.func,
}

MountWatcher.defaultProps = {
  onMount: () => {},
  onUnmount: () => {},
}
