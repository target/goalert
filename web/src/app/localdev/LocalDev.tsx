import React, { useState } from 'react'
import {
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
  Grid,
  Typography,
} from '@mui/material'
import { OpenInNew } from 'mdi-material-ui'
import { ConfigID } from '../../schema'
import AdminDialog from '../admin/AdminDialog'

type DevToolProps = {
  name: string
  desc: string

  config?: Partial<Record<ConfigID, string>>
  configDesc?: string

  url?: string
}

export default function LocalDev(): React.JSX.Element {
  const [updateConfig, setUpdateConfig] = useState<Partial<
    Record<ConfigID, string>
  > | null>(null)

  function DevTool(props: DevToolProps): React.JSX.Element {
    return (
      <Grid item xs={6}>
        <Card>
          <CardHeader title={props.name} />
          <CardContent>
            <Typography>{props.desc}</Typography>
          </CardContent>
          <CardActions>
            {props.config && (
              <Button
                title={props.configDesc}
                size='small'
                onClick={() => {
                  setUpdateConfig(props.config || null)
                }}
              >
                Update Config
              </Button>
            )}
            {props.url && (
              <Button target='_blank' size='small' href={props.url}>
                Open <OpenInNew fontSize='small' />
              </Button>
            )}
          </CardActions>
        </Card>
      </Grid>
    )
  }

  return (
    <React.Fragment>
      <Grid container spacing={2}>
        <DevTool
          name='SMTP'
          desc='Enable email contact methods to the local mail server.'
          config={{
            'SMTP.Enable': 'true',
            'SMTP.From': 'goalert@localhost',
            'SMTP.Address': 'localhost:1025',
            'SMTP.Username': '',
            'SMTP.Password': '',
            'SMTP.DisableTLS': 'true',
          }}
          url='http://localhost:8025'
        />

        <DevTool
          name='Prometheus'
          desc='Prometheus UI for viewing app metrics.'
          url='http://localhost:9090/graph?g0.expr=go_memstats_alloc_bytes&g0.tab=0&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h'
        />

        <DevTool
          name='gRPC'
          desc='UI for interacting with the gRPC server interface (sysapi).'
          url='http://localhost:8234'
        />
        <DevTool
          name='GraqphiQL'
          desc='UI for interacting with the GraphQL API.'
          url='/api/graphql/explore'
        />
        <DevTool
          name='OIDC'
          desc='Configure OIDC using the local test server.'
          config={{
            'OIDC.Enable': 'true',
            'OIDC.ClientID': 'test-client',
            'OIDC.ClientSecret': 'test-secret',
            'OIDC.IssuerURL': 'http://127.0.0.1:9998/oidc',
            'OIDC.UserInfoNamePath': 'preferred_username',
          }}
        />
        <DevTool
          name='pprof'
          desc='Debug and profile the running server.'
          url='http://localhost:6060/debug/pprof/'
        />
      </Grid>

      {updateConfig && (
        <AdminDialog
          value={updateConfig}
          onClose={() => setUpdateConfig(null)}
        />
      )}
    </React.Fragment>
  )
}
