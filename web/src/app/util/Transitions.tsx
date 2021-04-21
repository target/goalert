import React, { ReactElement } from 'react'
import Fade from '@material-ui/core/Fade'
import Slide from '@material-ui/core/Slide'

export const FadeTransition = React.forwardRef(
  ({ children, ...props }, ref) => (
    <Fade {...props} ref={ref}>
      {children as ReactElement}
    </Fade>
  ),
)
FadeTransition.displayName = 'FadeTransition'

export const SlideTransition = React.forwardRef(
  ({ children, ...props }, ref) => (
    <Slide {...props} direction='left' ref={ref}>
      {children as ReactElement}
    </Slide>
  ),
)
SlideTransition.displayName = 'SlideTransition'
