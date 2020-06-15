import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import CircularProgress from '@material-ui/core/CircularProgress'
import Typography from '@material-ui/core/Typography'

import { DEFAULT_SPIN_DELAY_MS, DEFAULT_SPIN_WAIT_MS } from '../../config'

/*
 * Show a loading spinner in the center of the container.
 */
export default function Spinner(props) {
  const [spin, setSpin] = useState(false)

  useEffect(() => {
    let _spin = setTimeout(() => {
      _spin = null
      setSpin(true)
      if (props.onSpin) props.onSpin()

      if (props.waitMs && props.onReady) {
        _spin = setTimeout(props.onReady, props.waitMs)
      }
    }, props.delayMs)

    return () => {
      clearTimeout(_spin)
    }
  })

  if (props.delayMs && !spin) return null

  const style = props.text
    ? {
        height: '1.5em',
        color: 'gray',
        display: 'flex',
        alignItems: 'center',
      }
    : { position: 'absolute', top: '50%', left: '50%' }

  return (
    <div style={style}>
      <CircularProgress size={props.text ? '1em' : '40px'} />
      &nbsp;<Typography variant='body2'>{props.text}</Typography>
    </div>
  )
}

Spinner.propTypes = {
  // Wait `delayMs` milliseconds before rendering a spinner.
  delayMs: p.number,

  // Wait `waitMs` before calling onReady.
  waitMs: p.number,

  // onSpin is called when the spinner starts spinning.
  onSpin: p.func,

  // onReady is called once the spinner has spun for `waitMs`.
  onReady: p.func,

  // text indicates being used as a text placeholder
  text: p.string,
}

Spinner.defaultProps = {
  delayMs: DEFAULT_SPIN_DELAY_MS,
  waitMs: DEFAULT_SPIN_WAIT_MS,
}
