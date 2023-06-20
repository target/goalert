import React, { ReactElement } from 'react'
import { Theme } from '@mui/material'
import Fade from '@mui/material/Fade'
import Slide from '@mui/material/Slide'
import makeStyles from '@mui/styles/makeStyles'
import { TransitionProps } from '@mui/material/transitions'

export const transitionStyles = makeStyles((theme: Theme) => {
  return {
    transition: {
      [theme.breakpoints.down('md')]: {
        flex: 1,
      },
      [theme.breakpoints.up('md')]: {
        '& input:focus': {
          minWidth: 275,
        },
        '& input:not(:placeholder-shown)': {
          minWidth: 275,
        },
        '& input': {
          minWidth: 180,
          transitionProperty: 'min-width',
          transitionDuration: theme.transitions.duration.standard,
          transitionTimingFunction: theme.transitions.easing.easeInOut,
        },
      },
    },
  }
})

export const FadeTransition = React.forwardRef(
  ({ children, ...props }: TransitionProps, ref) => (
    <Fade {...props} ref={ref}>
      {children as ReactElement}
    </Fade>
  ),
)
FadeTransition.displayName = 'FadeTransition'

export const SlideTransition = React.forwardRef(
  ({ children, ...props }: TransitionProps, ref) => (
    <Slide {...props} direction='left' ref={ref}>
      {children as ReactElement}
    </Slide>
  ),
)
SlideTransition.displayName = 'SlideTransition'
