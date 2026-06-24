import React from 'react'
import { Grid, List } from '@mui/material'
import CreateAlertListItem from './CreateAlertListItem'
import CreateAlertServiceListItem from './CreateAlertServiceListItem'

interface FailedService {
  id: string
  message: string
}

interface CreateAlertReviewProps {
  createdAlertIDs?: string[]
  failedServices?: FailedService[]
}

export function CreateAlertReview(
  props: CreateAlertReviewProps,
): React.JSX.Element {
  const { createdAlertIDs = [], failedServices = [] } = props

  return (
    <Grid container spacing={2}>
      {createdAlertIDs.length > 0 && (
        <Grid item xs={12}>
          <List aria-label='Successfully created alerts'>
            {createdAlertIDs.map((id: string) => (
              <CreateAlertListItem key={id} id={id} />
            ))}
          </List>
        </Grid>
      )}

      {failedServices.length > 0 && (
        <Grid item xs={12}>
          <List aria-label='Failed alerts'>
            {failedServices.map((svc) => {
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
