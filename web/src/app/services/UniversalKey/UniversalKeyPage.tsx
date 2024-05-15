import React from 'react'
import UniversalKeyRuleList from './UniversalKeyRuleList'
import { CardContent, Grid, Card } from '@mui/material'

interface UniversalKeyPageProps {
  serviceID: string
  keyName: string
}

export default function UniversalKeyPage({
  serviceID,
  keyName,
}: UniversalKeyPageProps): JSX.Element {
  // TODO: fix not getting keyName
  if (serviceID && keyName) {
    return (
      <React.Fragment>
        <Grid container>
          <Grid item xs={12}>
            <Card>
              <CardContent>{UniversalKeyRuleList()}</CardContent>
            </Card>
          </Grid>
        </Grid>
      </React.Fragment>
    )
  }
  // TODO: change back to <div />
  return (
    <React.Fragment>
      <Grid container>
        <Grid item xs={12}>
          <Card>
            <CardContent>{UniversalKeyRuleList()}</CardContent>
          </Card>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
