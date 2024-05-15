import React from 'react'
import UniversalKeyRuleList from './UniversalKeyRuleList'
import { CardContent, Grid, Card, CardHeader, Typography } from '@mui/material'
import { gql, useQuery } from 'urql'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { IntegrationKey, Service } from '../../../schema'
import Markdown from '../../util/Markdown'
import { Redirect } from 'wouter'

interface UniversalKeyPageProps {
  serviceID: string
  keyID: string
}

const query = gql`
  query UniversalKeyPage($keyID: ID!, $serviceID: ID!) {
    integrationKey(id: $keyID) {
      id
      name
      serviceID
      # tokenInfo {
      #   primaryHint
      #   secondaryHint
      # }
    }
    service(id: $serviceID) {
      id
      name
    }
  }
`

export default function UniversalKeyPage(
  props: UniversalKeyPageProps,
): React.ReactNode {
  const [q] = useQuery<{
    integrationKey: IntegrationKey
    service: Service
  }>({
    query,
    variables: {
      keyID: props.keyID,
      serviceID: props.serviceID,
    },
  })

  // Redirect to the correct service if the key is not in the service
  if (
    q.data &&
    q.data.integrationKey &&
    q.data.integrationKey.serviceID !== props.serviceID
  ) {
    return (
      <Redirect
        to={`/services/${q.data.integrationKey.serviceID}/integration-keys/${props.keyID}`}
      />
    )
  }

  if (q.error) {
    return <GenericError error={q.error.message} />
  }
  if (!q.data) return <ObjectNotFound type='integration key' />

  return (
    <React.Fragment>
      <Grid container>
        <Grid item xs={12}>
          <Card>
            <Grid item xs container direction='column'>
              <Grid item>
                <CardHeader
                  title={q.data.integrationKey.name}
                  subheader={`Service: ${q.data.service.name}`}
                  titleTypographyProps={{
                    'data-cy': 'title',
                    variant: 'h5',
                    component: 'h1',
                  }}
                  subheaderTypographyProps={{
                    'data-cy': 'subheader',
                    variant: 'body1',
                  }}
                />
              </Grid>
              <Grid item sx={{ pl: '16px', pr: '16px' }}>
                <Typography
                  component='div'
                  variant='subtitle1'
                  color='textSecondary'
                  data-cy='details'
                >
                  <Markdown value='' />
                </Typography>
              </Grid>
            </Grid>
          </Card>
          <br />
          <Card>
            <CardContent>{UniversalKeyRuleList()}</CardContent>
          </Card>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
