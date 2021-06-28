import React from 'react'
import { Fade, LinearProgress } from '@material-ui/core'

function FadeLinearProgress(): JSX.Element {
  return (
    <Fade
      in
      style={{
        transitionDelay: '800ms',
      }}
    >
      <LinearProgress />
    </Fade>
  )
}

export default FadeLinearProgress
