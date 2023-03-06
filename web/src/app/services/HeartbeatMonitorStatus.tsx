import React from 'react'
import Avatar from '@mui/material/Avatar'
import Grid from '@mui/material/Grid'
import ListItemAvatar from '@mui/material/ListItemAvatar'
import Typography from '@mui/material/Typography'
import HealthyIcon from '@mui/icons-material/Check'
import UnhealthyIcon from '@mui/icons-material/Clear'
import InactiveIcon from '@mui/icons-material/Remove'
import makeStyles from '@mui/styles/makeStyles'
import useStatusColors from '../theme/useStatusColors'
import { ISOTimestamp } from '../../schema'
import { Time } from '../util/Time'

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

export default function HeartbeatMonitorStatus(props: {
  lastState: 'inactive' | 'healthy' | 'unhealthy'
  lastHeartbeat?: null | ISOTimestamp
}): JSX.Element {
  const classes = useStyles()
  const statusColors = useStatusColors()

  const icon = icons[props.lastState]
  if (!icon) throw new TypeError('invalid state: ' + props.lastState)

  const bgColor = (status?: string): string => {
    switch (status) {
      case 'ok':
      case 'err':
        return statusColors[status]

      default:
        return 'default'
    }
  }

  return (
    <Grid container className={classes.gridContainer}>
      <Grid item xs={12}>
        <ListItemAvatar className={classes.avatarContainer}>
          <Avatar
            aria-label={props.lastState}
            sx={{ bgcolor: bgColor(statusMap[props.lastState]) }}
          >
            {icon}
          </Avatar>
        </ListItemAvatar>
      </Grid>
      <Grid item xs={12} className={classes.durationText}>
        <Typography variant='caption'>
          <Time time={props.lastHeartbeat} format='relative' zero='Inactive' />
        </Typography>
      </Grid>
    </Grid>
  )
}
