import React, { useState } from 'react'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import gql from 'graphql-tag'
import { startCase, isEmpty } from 'lodash-es'
import AdminDialog from './AdminDialog'
import PageActions from '../util/PageActions'
import { Form } from '../forms'
import AdminSection from './AdminSection'
import { useQuery } from '@apollo/react-hooks'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import AdminNumberLookup from './AdminNumberLookup'

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
    <div>
      <Grid container spacing={2} className={classes.gridContainer}>
        <Grid item xs={12} className={classes.gridItem}>
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
      </Grid>
    </div>
  )
}
