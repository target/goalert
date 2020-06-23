import React from 'react'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import AdminNumberLookup from './AdminNumberLookup'
import AdminSMSSend from './AdminSMSSend'

const useStyles = makeStyles((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  gridItem: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
}))

export default function AdminToolbox(): JSX.Element {
  const classes = useStyles()

  return (
    <Grid container spacing={2} className={classes.gridContainer}>
      <Grid container item xs={12} className={classes.gridItem}>
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
        <Grid item xs={12}>
          <AdminNumberLookup />
        </Grid>
      </Grid>
      <Grid container item xs={12} className={classes.gridItem}>
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
        <Grid item xs={12}>
          <AdminSMSSend />
        </Grid>
      </Grid>
    </Grid>
  )
}
