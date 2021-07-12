import React from 'react'
import classnames from 'classnames'
import AddIcon from '@material-ui/icons/Add'
import { Fab, Tooltip, makeStyles, FabProps } from '@material-ui/core'

const useStyles = makeStyles((theme) => ({
  fab: {
    position: 'fixed',
    bottom: '16px',
    right: '16px',
    zIndex: 9001,
  },
  transitionUp: {
    transform: 'translate3d(0, -62px, 0)',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.enteringScreen,
      easing: theme.transitions.easing.easeOut,
    }),
  },
  transitionDown: {
    transform: 'translate3d(0, 0, 0)',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.leavingScreen,
      easing: theme.transitions.easing.sharp,
    }),
  },
}))

interface CreateFabProps extends Omit<FabProps, 'children'> {
  title: string
  transition?: boolean
}

export default function CreateFAB(props: CreateFabProps): JSX.Element {
  const { title, transition, ...fabProps } = props
  const classes = useStyles()

  const transitionClass = transition
    ? classes.transitionUp
    : classes.transitionDown

  return (
    <Tooltip title={title} aria-label={title} placement='left'>
      <Fab
        aria-label={title}
        data-cy='page-fab'
        color='primary'
        className={classnames(classes.fab, transitionClass)}
        {...fabProps}
      >
        <AddIcon />
      </Fab>
    </Tooltip>
  )
}
