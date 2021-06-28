import React, { useState, useEffect } from 'react'
import { LinearProgress } from '@material-ui/core'
import { DEFAULT_SPIN_DELAY_MS } from '../config'

export function useLinearProgress(loading?: boolean): JSX.Element | null {
  const [render, setRender] = useState(false)

  useEffect(() => {
    if (!loading) {
      setRender(false)
    } else {
      const t = setTimeout(() => {
        if (loading) setRender(true)
      }, DEFAULT_SPIN_DELAY_MS)

      return () => {
        clearTimeout(t)
      }
    }
  }, [loading])

  return render ? <LinearProgress /> : null
}
