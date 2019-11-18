import React from 'react'
import p from 'prop-types'
import { Grid, List } from '@material-ui/core'
import CreateAlertListItem from './CreateAlertListItem'
import CreateAlertServiceListItem from './CreateAlertServiceListItem'

export function CreateAlertReview(props) {
  const { createdAlertIDs = [], failedServices = [] } = props

  return (
    <Grid container spacing={2}>
      {createdAlertIDs.length > 0 && (
        <Grid item xs={12}>
          <List aria-label='Successfully created alerts'>
            {createdAlertIDs.map(id => (
              <CreateAlertListItem key={id} id={id} />
            ))}
          </List>
        </Grid>
      )}

      {failedServices.length > 0 && (
        <Grid item xs={12}>
          <List aria-label='Failed alerts'>
            {failedServices.map(svc => {
              return (
                <CreateAlertServiceListItem
                  key={svc.id}
                  id={svc.id}
                  err={svc.message}
                />
              )
            })}
          </List>
        </Grid>
      )}
    </Grid>
  )
}

CreateAlertReview.propTypes = {
  createdAlertIDs: p.arrayOf(p.string),
  failedServices: p.arrayOf(
    p.shape({
      id: p.string,
      message: p.string,
    }),
  ),
}
