import React from 'react'
import { Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles/makeStyles'
import { theme } from '../mui'

const useStyles = makeStyles<typeof theme>((theme) => ({
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

export default function AdminMetrics() {
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
            Alert Metrics
          </Typography>
        </Grid>
      </Grid>
    </Grid>
  )
}
