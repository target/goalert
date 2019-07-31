import React from 'react'
import p from 'prop-types'
import Avatar from '@material-ui/core/Avatar'
import Grid from '@material-ui/core/Grid'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import Typography from '@material-ui/core/Typography'
import HealthyIcon from '@material-ui/icons/Check'
import UnhealthyIcon from '@material-ui/icons/Clear'
import InactiveIcon from '@material-ui/icons/Remove'
import { makeStyles } from '@material-ui/core/styles'
import { formatTimeSince } from '../util/timeFormat'
import { colors } from '../util/statusStyles'

const useStyles = makeStyles({
  gridContainer: {
    width: 'min-content',
    marginRight: '1em',
  },
  unhealthy: {
    color: '#fff',
    backgroundColor: colors.statusError,
  },
  healthy: {
    color: '#fff',
    backgroundColor: colors.statusOK,
  },
  avatarContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
  durationText: {
    textAlign: 'center',
  },
})

const icons = {
  healthy: <HealthyIcon />,
  unhealthy: <UnhealthyIcon />,
  inactive: <InactiveIcon />,
}

export default function HeartbeatMonitorStatus(props) {
  const classes = useStyles()

  const icon = icons[props.lastState]
  if (!icon) throw new TypeError('invalid state: ' + props.lastState)

  return (
    <Grid container className={classes.gridContainer}>
      <Grid item xs={12}>
        <ListItemAvatar className={classes.avatarContainer}>
          <Avatar
            aria-label={props.lastState}
            className={classes[props.lastState]}
          >
            {icon}
          </Avatar>
        </ListItemAvatar>
      </Grid>
      <Grid item xs={12} className={classes.durationText}>
        <Typography variant='caption'>
          {formatTimeSince(props.lastHeartbeat) || 'Inactive'}
        </Typography>
      </Grid>
    </Grid>
  )
}

HeartbeatMonitorStatus.propTypes = {
  lastState: p.oneOf(['inactive', 'healthy', 'unhealthy']).isRequired,
  lastHeartbeat: p.string,
}
