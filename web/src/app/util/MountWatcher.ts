import { useEffect } from 'react'

type MountWatcherProps = {
  children: any
  onMount: () => void
  onUnmount: () => void
}

export default function MountWatcher(props: MountWatcherProps) {
  const { onMount, onUnmount } = props
  useEffect(() => {
    onMount()
    return () => onUnmount()
  }, [])

  return props.children
}
