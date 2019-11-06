import React from 'react'
import { List, ListItem, Paper, Typography } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { ServiceChip } from '../../../util/Chips'
import _ from 'lodash-es'
import AlertListItem from '../AlertListItem'

const useStyles = makeStyles(theme => ({
  spaceBetween: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
}))

export default props => {
  const { formFields, mutationStatus } = props

  const classes = useStyles()

  const alertsCreated = mutationStatus.alertsCreated || {}
  const graphQLErrors = _.get(mutationStatus, 'alertsFailed.graphQLErrors', [])

  const numCreated = Object.keys(alertsCreated).length

  return (
    <Paper elevation={0}>
      {numCreated > 0 && (
        <div>
          <span className={classes.spaceBetween}>
            <Typography variant='subtitle1' component='h3'>
              {`Successfully created ${numCreated} alerts`}
            </Typography>
          </span>
          <List aria-label='Successfully created alerts'>
            {Object.keys(alertsCreated).map((alias, i) => (
              <AlertListItem key={i} id={alertsCreated[alias].id} />
            ))}
          </List>
        </div>
      )}

      {graphQLErrors.length > 0 && (
        <div>
          <Typography variant='h6' component='h3'>
            Failed to create alerts on these services:
          </Typography>

          <List aria-label='Failed alerts'>
            {graphQLErrors.map((err, i) => {
              const index = err.path[0].split(/(\d+)$/)[1]
              const serviceId = formFields.selectedServices[index]
              return (
                <ListItem key={i}>
                  <ServiceChip id={serviceId} />
                </ListItem>
              )
            })}
          </List>
        </div>
      )}
    </Paper>
  )
}
