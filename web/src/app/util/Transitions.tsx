import React, { ReactElement } from 'react'
import Fade from '@material-ui/core/Fade'
import Slide from '@material-ui/core/Slide'

export const DefaultTransition = React.forwardRef(
  ({ children, ...props }, ref) => (
    <Fade {...props} ref={ref}>
      {children as ReactElement}
    </Fade>
  ),
)

export const FullscreenTransition = React.forwardRef(
  ({ children, ...props }, ref) => (
    <Slide {...props} direction='left' ref={ref}>
      {children as ReactElement}
    </Slide>
  ),
)

export const FullscreenExpansion = React.forwardRef(
  ({ children, ...props }, ref) => (
    <Slide {...props} direction='right' ref={ref}>
      {children as ReactElement}
    </Slide>
  ),
)
