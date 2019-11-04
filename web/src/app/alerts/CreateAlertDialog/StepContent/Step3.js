import React from 'react'
import { List, ListItem, Paper, Chip, Typography } from '@material-ui/core'
import { makeStyles, emphasize } from '@material-ui/core/styles'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import { ServiceChip } from '../../../util/Chips'
import _ from 'lodash-es'
import AlertListItem from '../AlertListItem'

const useStyles = makeStyles(theme => ({
  openAll: {
    backgroundColor: theme.palette.grey[100],
    height: theme.spacing(3),
    color: theme.palette.grey[800],
    fontWeight: theme.typography.fontWeightRegular,
    '&:hover, &:focus': {
      backgroundColor: theme.palette.grey[300],
      textDecoration: 'none',
    },
    '&:active': {
      boxShadow: theme.shadows[1],
      backgroundColor: emphasize(theme.palette.grey[300], 0.12),
      textDecoration: 'none',
    },
  },
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

  const OpenAll = () => (
    <Chip
      component='button'
      label='Open All'
      icon={<OpenInNewIcon fontSize='small' />}
      onClick={() => {
        formFields.selectedServices.forEach(id => {
          window.open(`/alerts/${id}`)
        })
      }}
      className={classes.openAll}
    />
  )

  return (
    <Paper elevation={0}>
      {numCreated > 0 && (
        <div>
          <span className={classes.spaceBetween}>
            <Typography variant='subtitle1' component='h3'>
              {`Successfully created ${numCreated} alerts`}
            </Typography>
            <OpenAll />
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
