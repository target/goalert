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
    SwitchOverState {
      isDone
      isIdle
      nodes {
        id
        status
        canExecute
        oldValid
        newValid
      }
    }
  }
`

const mutation = gql`
  mutation ($action: SwitchoverAction!) {
    switchoverAction(action: $action)
  }
`

function getActions(isDone: boolean, isIdle: boolean): Array<string> {
  if (isDone && isIdle) {
    return ['ping']
  }

  if (isDone && !isIdle) {
    return ['ping']
  }

  if (!isDone && isIdle) {
    return ['ping', 'reset', 'execute']
  }

  return []
}

export default function AdminSwitchover(): JSX.Element {
  const queryRes = useQuery(query)
  const [commit, mutationRes] = useMutation(mutation)
  const s = queryRes.data

  return (
    <Grid container spacing={2} justifyContent='space-between'>
      <Grid item>
        <Card>
          <CardContent>
            <Typography>{s?.status ?? 'Some Status'}</Typography>
          </CardContent>
        </Card>
      </Grid>
      <Grid item>
        <Button
          onClick={() => commit({ variables: { action: 'refresh' } })}
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
