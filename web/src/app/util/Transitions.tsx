import React, { ReactElement } from 'react'
import Fade from '@material-ui/core/Fade'
import Slide from '@material-ui/core/Slide'

export const DefaultTransition = React.forwardRef((props, ref) => (
  <Fade ref={ref}>{props.children as ReactElement}</Fade>
))

export const FullscreenTransition = React.forwardRef((props, ref) => (
  <Slide direction='left' ref={ref}>
    {props.children as ReactElement}
  </Slide>
))

export const FullscreenExpansion = React.forwardRef((props, ref) => (
  <Slide direction='right' ref={ref}>
    {props.children as ReactElement}
  </Slide>
))
