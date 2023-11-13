import React from 'react'
import classnames from 'classnames'
import AddIcon from '@mui/icons-material/Add'
import { Fab, Tooltip, FabProps } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'

const useStyles = makeStyles((theme: Theme) => ({
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

export default function CreateFAB(props: CreateFabProps): React.ReactNode {
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
        color='secondary'
        className={classnames(classes.fab, transitionClass)}
        {...fabProps}
      >
        <AddIcon />
      </Fab>
    </Tooltip>
  )
}
