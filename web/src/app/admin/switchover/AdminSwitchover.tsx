import React from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'

const useStyles = makeStyles((theme: Theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
}))

export default function AdminSwitchover(): JSX.Element {
  const classes = useStyles()

  return (
    <Grid container spacing={2} className={classes.gridContainer}>
      <Grid container item xs={12}>
        <Grid item xs={12}>
          <Typography
            component='h2'
            variant='subtitle1'
            color='textSecondary'
            classes={{ subtitle1: classes.groupTitle }}
          >
            Twilio Number Lookup
          </Typography>
        </Grid>
      </Grid>
      <Grid container item xs={12}>
        <Grid item xs={12}>
          <Typography
            component='h2'
            variant='subtitle1'
            color='textSecondary'
            classes={{ subtitle1: classes.groupTitle }}
          >
            Send SMS
          </Typography>
        </Grid>
      </Grid>
    </Grid>
  )
}
