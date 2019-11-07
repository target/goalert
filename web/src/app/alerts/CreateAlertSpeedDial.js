import React, { useState } from 'react'
import { KeyChange as ServicesIcon } from 'mdi-material-ui'
import { Notifications as AlertsIcon } from '@material-ui/icons/'
import { makeStyles } from '@material-ui/styles'
import classnames from 'classnames'
import AlertForm from './components/AlertForm'
import CreateAlertDialog from './CreateAlertDialog'
import SpeedDial from '../util/SpeedDial'

const useStyles = makeStyles(theme => ({
  speedDial: {
    position: 'fixed',
    bottom: '2em',
    right: '2em',
    zIndex: 9001,
  },
}))

export default function CreateAlertSpeedDial(props) {
  const classes = useStyles()

  const [showCreateAlertForm, setShowCreateAlertForm] = useState(false)
  const [showCreateAlertDialog, setShowCreateAlertDialog] = useState(false)

  return (
    <React.Fragment>
      <SpeedDial
        data-cy='page-speed-dial'
        color='primary'
        className={classnames(classes.speedDial)}
        label='Create Alert'
        actions={[
          {
            label: 'Alert Multiple Services',
            onClick: () => setShowCreateAlertDialog(true),
            icon: <ServicesIcon />,
          },
          {
            label: 'Create Single Alert',
            onClick: () => setShowCreateAlertForm(true),
            icon: <AlertsIcon />,
          },
        ]}
      />
      <AlertForm
        open={showCreateAlertForm}
        handleRequestClose={() => setShowCreateAlertForm(false)}
      />
      <CreateAlertDialog
        open={showCreateAlertDialog}
        handleRequestClose={() => setShowCreateAlertDialog(false)}
      />
    </React.Fragment>
  )
}
