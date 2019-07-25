import React, { useState } from 'react'
import p from 'prop-types'
import HeartbeatDeleteDialog from './HeartbeatDeleteDialog'
import OtherActions from '../util/OtherActions'
import Avatar from '@material-ui/core/Avatar'
import Grid from '@material-ui/core/Grid'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import Typography from '@material-ui/core/Typography'
import HealthyIcon from '@material-ui/icons/Check'
import UnhealthyIcon from '@material-ui/icons/Clear'
import InactiveIcon from '@material-ui/icons/Remove'
import { makeStyles } from '@material-ui/core/styles'
import { green, red } from '@material-ui/core/colors'
import HeartbeatMonitorEditDialog from './HeartbeatMonitorEditDialog'
import { formatTimeSince } from '../util/timeFormat'
import CopyText from '../util/CopyText'

export default function HeartbeatMonitorListItem(props) {
  return (
    <React.Fragment>
      {`Timeout: ${props.timeoutMinutes} minute${
        props.timeoutMinutes > 1 ? 's' : ''
      }`}
      <br />
      <CopyText title='Copy API key' value={props.href} />
    </React.Fragment>
  )
}

HeartbeatMonitorListItem.propTypes = {
  href: p.string.isRequired,
  timeoutMinutes: p.number.isRequired,
}

export function HeartbeatMonitorListItemActions(props) {
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  return (
    <React.Fragment>
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => setShowEditDialog(true),
          },
          {
            label: 'Delete',
            onClick: () => setShowDeleteDialog(true),
          },
        ]}
      />
      {showDeleteDialog && (
        <HeartbeatDeleteDialog
          heartbeatID={props.monitorID}
          onClose={() => setShowDeleteDialog(false)}
          refetchQueries={props.refetchQueries}
        />
      )}
      {showEditDialog && (
        <HeartbeatMonitorEditDialog
          monitorID={props.monitorID}
          onClose={() => setShowEditDialog(false)}
          refetchQueries={props.refetchQueries}
        />
      )}
    </React.Fragment>
  )
}

HeartbeatMonitorListItemActions.propTypes = {
  monitorID: p.string.isRequired,
  refetchQueries: p.arrayOf(p.string),
}

const useStyles = makeStyles({
  unhealthy: {
    color: '#fff',
    backgroundColor: red[500],
  },
  healthy: {
    color: '#fff',
    backgroundColor: green[500],
  },
})

export function HeartbeatMonitorListItemAvatar(props) {
  const classes = useStyles()

  function renderLastHeartbeat() {
    return (
      <Typography variant='caption'>
        {formatTimeSince(props.lastHeartbeat)}
      </Typography>
    )
  }

  switch (props.lastState) {
    case 'healthy':
      return (
        <Grid container>
          <Grid item xs={12}>
            <ListItemAvatar>
              <Avatar aria-label='Healthy' className={classes.healthy}>
                <HealthyIcon />
              </Avatar>
            </ListItemAvatar>
          </Grid>
          <Grid item xs={12}>
            {renderLastHeartbeat()}
          </Grid>
        </Grid>
      )
    case 'unhealthy':
      return (
        <Grid container>
          <Grid item xs={12}>
            <ListItemAvatar>
              <Avatar aria-label='Unhealthy' className={classes.unhealthy}>
                <UnhealthyIcon />
              </Avatar>
            </ListItemAvatar>
          </Grid>
          <Grid item xs={12}>
            {renderLastHeartbeat()}
          </Grid>
        </Grid>
      )
    case 'inactive':
      return (
        <Grid container>
          <Grid item xs={12}>
            <ListItemAvatar>
              <Avatar aria-label='Inactive'>
                <InactiveIcon />
              </Avatar>
            </ListItemAvatar>
          </Grid>
          <Grid item xs={12}>
            {renderLastHeartbeat()}
          </Grid>
        </Grid>
      )
    default:
      return null
  }
}

HeartbeatMonitorListItemAvatar.propTypes = {
  lastState: p.oneOf(['inactive', 'healthy', 'unhealthy']).isRequired,
  lastHeartbeat: p.string,
}
