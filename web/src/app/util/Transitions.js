import React from 'react'
import Fade from '@material-ui/core/Fade'
import Slide from '@material-ui/core/Slide'

export function DefaultTransition(props) {
  return <Fade {...props} />
}

export function FullscreenTransition(props) {
  return <Slide direction='left' {...props} />
}

export function FullscreenExpansion(props) {
  return <Slide direction='right' {...props} />
}
