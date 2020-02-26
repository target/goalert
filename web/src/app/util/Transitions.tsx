import React from 'react'
import Fade from '@material-ui/core/Fade'
import Slide from '@material-ui/core/Slide'

export const DefaultTransition = React.forwardRef((props, ref) => (
  <Fade {...props} ref={ref} />
))

export const FullscreenTransition = React.forwardRef((props, ref) => (
  <Slide direction='left' {...props} ref={ref} />
))

export const FullscreenExpansion = React.forwardRef((props, ref) => (
  <Slide direction='right' {...props} ref={ref} />
))
