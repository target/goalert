import React from 'react'
import p from 'prop-types'
import AddIcon from '@material-ui/icons/Add'
import Fab from '@material-ui/core/Fab'
import Tooltip from '@material-ui/core/Tooltip'

export default function CreateFAB(props) {
  const { onClick, title } = props

  return (
    <Tooltip title={title} aria-label={title} placement='left'>
      <Fab
        aria-label='Create New'
        data-cy='page-fab'
        color='primary'
        style={{ position: 'fixed', bottom: '2em', right: '2em' }}
        onClick={onClick}
      >
        <AddIcon />
      </Fab>
    </Tooltip>
  )
}

CreateFAB.propTypes = {
  onClick: p.func,
  title: p.string,
}
