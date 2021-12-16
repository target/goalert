import React, { ReactElement } from 'react'
import Fade from '@mui/material/Fade'
import Slide from '@mui/material/Slide'

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
