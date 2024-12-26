import React, { useState, useEffect } from 'react'
import CircularProgress from '@mui/material/CircularProgress'
import Typography from '@mui/material/Typography'

import { DEFAULT_SPIN_DELAY_MS, DEFAULT_SPIN_WAIT_MS } from '../../config'

interface SpinnerProps {
  // Wait `delayMs` milliseconds before rendering a spinner.
  delayMs?: number

  // Wait `waitMs` before calling onReady.
  waitMs?: number

  // onSpin is called when the spinner starts spinning.
  onSpin?: () => void

  // onReady is called once the spinner has spun for `waitMs`.
  onReady?: () => void

  // text indicates being used as a text placeholder
  text?: string
}

/*
 * Show a loading spinner in the center of the container.
 */
export default function Spinner(props: SpinnerProps): React.JSX.Element | null {
  const [spin, setSpin] = useState(false)
  const { delayMs = DEFAULT_SPIN_DELAY_MS, waitMs = DEFAULT_SPIN_WAIT_MS } =
    props

  useEffect(() => {
    let _spin = setTimeout(() => {
      setSpin(true)
      if (props.onSpin) props.onSpin()

      if (waitMs && props.onReady) {
        _spin = setTimeout(props.onReady, waitMs)
      }
    }, delayMs)

    return () => {
      clearTimeout(_spin)
    }
  }, [])

  if (props.delayMs && !spin) return null

  const style: React.CSSProperties = props.text
    ? {
        height: '1.5em',
        color: 'gray',
        display: 'flex',
        alignItems: 'center',
      }
    : {
        position: 'absolute',
        top: 'calc(50% - 20px)',
        left: 'calc(50% - 20px)',
        zIndex: 99999,
      }

  return (
    <div style={style}>
      <CircularProgress
        data-cy='loading-spinner'
        size={props.text ? '1em' : '40px'}
      />
      &nbsp;<Typography variant='body2'>{props.text}</Typography>
    </div>
  )
}
