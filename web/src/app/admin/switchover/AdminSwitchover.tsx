import React from 'react'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import PingIcon from 'mdi-material-ui/SourceCommitStartNextLocal'
import RestartIcon from '@mui/icons-material/Refresh'
import ExecuteIcon from '@mui/icons-material/Start'
import { gql, useMutation, useQuery } from '@apollo/client'
import ErrorIcon from '@mui/icons-material/Report'

const query = gql`
  query {
    swoStatus {
      isDone
      isIdle
      details
      nodes {
        ID
        status
        canExec
        oldValid
        newValid
      }
    }
  }
`

const mutation = gql`
  mutation ($action: SWOAction!) {
    swoAction(action: $action)
  }
`

export default function AdminSwitchover(): JSX.Element {
  const { loading, error, data } = useQuery(query)
  const [commit, mutationRes] = useMutation(mutation)

  // todo: loading skeletons
  // todo: error message on query error

  return (
    <Grid container spacing={4} justifyContent='center'>
      {error && (
        <Grid item>
          <Card>
            <CardContent>
              <Grid
                container
                spacing={2}
                direction='column'
                alignItems='center'
              >
                <Grid item>
                  <ErrorIcon color='error' sx={{ fontSize: '3.5rem' }} />
                </Grid>
                <Grid item>
                  <Typography>{error.message}</Typography>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      )}

      <Grid item>
        <Card sx={{ minWidth: 300 }}>
          <CardContent>
            <Grid container spacing={2} direction='column'>
              <Grid item>
                <Typography sx={{ fontSize: '1.25rem' }}>
                  Switchover Status
                </Typography>
              </Grid>
              <Grid item>
                <Typography>{data?.status ?? 'Some Status'}</Typography>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      </Grid>

      <Grid item>
        <Button
          onClick={() => commit({ variables: { action: 'ping' } })}
          size='large'
          variant='outlined'
          startIcon={<PingIcon />}
        >
          Ping
        </Button>
      </Grid>
      <Grid item>
        <Button
          onClick={() => commit({ variables: { action: 'reset' } })}
          disabled={data?.isDone}
          size='large'
          variant='outlined'
          startIcon={<RestartIcon />}
        >
          Reset
        </Button>
      </Grid>
      <Grid item>
        <Button
          onClick={() => commit({ variables: { action: 'execute' } })}
          disabled={!data?.isIdle || (!data?.isIdle && data?.isDone)}
          size='large'
          variant='outlined'
          startIcon={<ExecuteIcon />}
        >
          Execute
        </Button>
      </Grid>
    </Grid>
  )
}
