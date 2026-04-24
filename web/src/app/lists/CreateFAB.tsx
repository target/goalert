import React from 'react'
import AddIcon from '@mui/icons-material/Add'
import { Fab, Tooltip, FabProps } from '@mui/material'

interface CreateFabProps extends Omit<FabProps, 'children'> {
  title: string
  transition?: boolean
}

export default function CreateFAB(props: CreateFabProps): React.JSX.Element {
  const { title, transition, ...fabProps } = props

  return (
    <Tooltip title={title} aria-label={title} placement='left'>
      <Fab
        aria-label={title}
        data-cy='page-fab'
        color='secondary'
        sx={(theme) => ({
          position: 'fixed',
          bottom: '16px',
          right: '16px',
          zIndex: 9001,
          ...(transition
            ? {
                transform: 'translate3d(0, -62px, 0)',
                transition: theme.transitions.create('transform', {
                  duration: theme.transitions.duration.enteringScreen,
                  easing: theme.transitions.easing.easeOut,
                }),
              }
            : {
                transform: 'translate3d(0, 0, 0)',
                transition: theme.transitions.create('transform', {
                  duration: theme.transitions.duration.leavingScreen,
                  easing: theme.transitions.easing.sharp,
                }),
              }),
        })}
        {...fabProps}
      >
        <AddIcon />
      </Fab>
    </Tooltip>
  )
}
