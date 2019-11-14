import React, { useState } from 'react'
import { makeStyles } from '@material-ui/styles'
import CreateAlertDialog from './CreateAlertDialog/CreateAlertDialog'
import Fab from '@material-ui/core/Fab'
import AddIcon from '@material-ui/icons/Add'
import Tooltip from '@material-ui/core/Tooltip'
import p from 'prop-types'

const useStyles = makeStyles(theme => ({
  fab: {
    position: 'fixed',
    bottom: '2em',
    right: '2em',
    zIndex: 9001,
  },
}))

export default function CreateAlertFab(props) {
  const classes = useStyles()
  const [open, setOpen] = useState(false)

  return (
    <React.Fragment>
      <Tooltip title='Create Alert' aria-label='Create Alert' placement='left'>
        <Fab
          data-cy='page-fab'
          className={classes.fab}
          color='primary'
          onClick={() => setOpen(true)}
        >
          <AddIcon />
        </Fab>
      </Tooltip>
      {open && (
        <CreateAlertDialog
          onClose={() => setOpen(false)}
          serviceID={props.serviceID}
        />
      )}
    </React.Fragment>
  )
}

CreateAlertFab.propTypes = {
  serviceID: p.string,
}
