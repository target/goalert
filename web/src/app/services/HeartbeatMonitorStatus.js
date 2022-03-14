import React from 'react'
import p from 'prop-types'
import Avatar from '@mui/material/Avatar'
import Grid from '@mui/material/Grid'
import ListItemAvatar from '@mui/material/ListItemAvatar'
import Typography from '@mui/material/Typography'
import HealthyIcon from '@mui/icons-material/Check'
import UnhealthyIcon from '@mui/icons-material/Clear'
import InactiveIcon from '@mui/icons-material/Remove'
import makeStyles from '@mui/styles/makeStyles'
import { formatTimeSince } from '../util/timeFormat'
import useStatusColors from '../theme/useStatusColors'

const useStyles = makeStyles({
  gridContainer: {
    width: 'min-content',
    marginRight: '1em',
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

const statusMap = {
  healthy: 'ok',
  unhealthy: 'err',
  inactive: '',
}

export default function HeartbeatMonitorStatus(props) {
  const classes = useStyles()
  const statusColors = useStatusColors()

  const icon = icons[props.lastState]
  if (!icon) throw new TypeError('invalid state: ' + props.lastState)

  return (
    <Grid container className={classes.gridContainer}>
      <Grid item xs={12}>
        <ListItemAvatar className={classes.avatarContainer}>
          <Avatar
            aria-label={props.lastState}
            sx={{ bgcolor: statusColors[statusMap[props.lastState]] }}
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
