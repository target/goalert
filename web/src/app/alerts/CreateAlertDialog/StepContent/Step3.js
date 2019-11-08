import React, { useEffect } from 'react'
import { Divider, Grid, List, Typography } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import _ from 'lodash-es'
import AlertListItem from '../AlertListItem'
import ServiceListItem from '../ServiceListItem'

const useStyles = makeStyles(theme => ({
  noPaddingBottom: {
    paddingBottom: '0 !important',
  },
  noPaddingTop: {
    paddingTop: '0 !important',
  },
  spaceBetween: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
}))

const pluralize = num => `${num !== 1 ? 's' : ''}`

export default function Step3(props) {
  const { formFields, mutationStatus, setActiveStep } = props

  const classes = useStyles()

  const alertsCreated = mutationStatus.alertsCreated || {}
  const graphQLErrors = _.get(mutationStatus, 'alertsFailed.graphQLErrors', [])

  const numFailed = graphQLErrors.length
  const numCreated = Object.keys(alertsCreated).length - numFailed

  useEffect(() => {
    if (numFailed === 0) return

    const step0FieldNames = ['Details', 'Summary']

    if (
      graphQLErrors.some(err =>
        step0FieldNames.includes(err.extensions.fieldName),
      )
    ) {
      setActiveStep(0)
    } else if (
      graphQLErrors[0].extensions.isMultiFieldError &&
      graphQLErrors[0].extensions.fieldErrors.some(err =>
        step0FieldNames.includes(err.fieldName),
      )
    ) {
      setActiveStep(0)
    }
  }, [graphQLErrors])

  const HeaderItem = props => (
    <Grid item xs={12} className={classes.noPaddingBottom}>
      <Typography variant='subtitle1' component='h3'>
        {props.text}
      </Typography>
      <Divider />
    </Grid>
  )

  return (
    <Grid container spacing={2}>
      {numCreated > 0 && (
        <HeaderItem
          text={`Successfully created ${numCreated} alert${pluralize(
            numCreated,
          )}`}
        />
      )}

      {numCreated > 0 && (
        <Grid item xs={12} className={classes.noPaddingTop}>
          <List aria-label='Successfully created alerts'>
            {Object.keys(alertsCreated).map(alias => {
              const alert = alertsCreated[alias]
              if (!alert) return null
              return <AlertListItem key={alias} id={alertsCreated[alias].id} />
            })}
          </List>
        </Grid>
      )}

      {numFailed > 0 && (
        <HeaderItem
          text={`Failed to create ${numFailed} alert${pluralize(numFailed)}`}
        />
      )}

      {numFailed > 0 && (
        <Grid item xs={12} className={classes.noPaddingTop}>
          <List aria-label='Failed alerts'>
            {graphQLErrors.map(err => {
              const index = err.path[0].split(/(\d+)$/)[1]
              const serviceId = formFields.selectedServices[index]
              return (
                <ServiceListItem
                  key={serviceId}
                  id={serviceId}
                  err={err.message}
                />
              )
            })}
          </List>
        </Grid>
      )}
    </Grid>
  )
}
