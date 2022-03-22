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

const query = gql`
  query {
    data: swoStatus {
      isDone
      isIdle
      details
      nodes {
        id
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
  const queryRes = useQuery(query)
  const [commit, mutationRes] = useMutation(mutation)
  const s = queryRes.data

  return (
    <Grid container spacing={4} justifyContent='center'>
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
                <Typography>{s?.status ?? 'Some Status'}</Typography>
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
          disabled={queryRes?.data?.isDone}
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
          disabled={
            !queryRes?.data?.isIdle ||
            (!queryRes?.data?.isIdle && queryRes?.data?.isDone)
          }
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
