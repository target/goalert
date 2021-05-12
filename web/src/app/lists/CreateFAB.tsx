import React from 'react'
import AddIcon from '@material-ui/icons/Add'
import { Fab, Tooltip } from '@material-ui/core'

interface CreateFabProps {
  onClick: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void
  title: string
}

export default function CreateFAB(props: CreateFabProps): JSX.Element {
  const { onClick, title } = props

  return (
    <Tooltip title={title} aria-label={title} placement='left'>
      <Fab
        aria-label='Create New'
        data-cy='page-fab'
        color='primary'
        style={{
          zIndex: 9001,
          position: 'fixed',
          bottom: '1em',
          right: '1em',
        }}
        onClick={onClick}
      >
        <AddIcon />
      </Fab>
    </Tooltip>
  )
}
