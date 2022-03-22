import React, { useState } from 'react'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Grid from '@mui/material/Grid'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import PingIcon from 'mdi-material-ui/DatabaseMarker'
import NoResetIcon from 'mdi-material-ui/DatabaseRefreshOutline'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdlingIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { gql, useMutation, useQuery } from '@apollo/client'
import Notices, { Notice } from '../../details/Notices'
import { DateTime } from 'luxon'

const query = gql`
  query {
    swoStatus {
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

function cptlz(s: string): string {
  return s.charAt(0).toUpperCase() + s.substring(1)
}

export default function AdminSwitchover(): JSX.Element {
  const { loading, error, data: _data } = useQuery(query)
  const data = _data?.swoStatus

  const [statusNotices, setStatusNotices] = useState<Notice[]>([])
  const [commit] = useMutation(mutation)

  function getIcon(): React.ReactNode {
    if (error) {
      return <ErrorIcon color='error' sx={{ fontSize: '3.5rem' }} />
    }

    if (loading) {
      return (
        <Skeleton variant='circular'>
          <InProgressIcon color='primary' sx={{ fontSize: '3.5rem' }} />
        </Skeleton>
      )
    }

    // todo: in progress state icon
    if (!data.isIdle && !data.isDone) {
      return <InProgressIcon color='primary' sx={{ fontSize: '3.5rem' }} />
    }

    if (data.isIdle) {
      return <IdlingIcon color='primary' sx={{ fontSize: '3.5rem' }} />
    }
  }

  function getDetails(): React.ReactNode {
    if (error) {
      return <Typography color='error'>{cptlz(error.message)}</Typography>
    }

    if (loading) {
      return <Typography>Loading...</Typography>
    }

    if (data?.details) {
      return <Typography>{cptlz(data.details)}</Typography>
    }

    return null
  }

  const minHeight = 90
  const buttonSx = { display: 'grid', minHeight, minWidth: minHeight }
  const iconSx = { justifySelf: 'center', height: '1.25em', width: '1.25em' }
  return (
    <Grid container spacing={4}>
      {statusNotices.length > 0 && (
        <Grid item xs={12}>
          <Notices notices={statusNotices.reverse()} />
        </Grid>
      )}

      <Grid item>
        <Card sx={{ minWidth: 300, minHeight }}>
          <CardHeader
            title='Switchover Status'
            titleTypographyProps={{ sx: { fontSize: '1.25rem' } }}
            avatar={getIcon()}
            subheader={getDetails()}
          />
        </Card>
      </Grid>

      <Grid item>
        <Button
          onClick={() =>
            commit({
              variables: { action: 'ping' },
              onCompleted: () => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'success',
                    message: 'Successfully pinged',
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
              onError: (error) => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'error',
                    message: 'Failed to ping',
                    details: cptlz(error.message),
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
            })
          }
          size='large'
          variant='outlined'
          sx={buttonSx}
        >
          <PingIcon sx={iconSx} />
          Ping
        </Button>
      </Grid>
      <Grid item>
        <Button
          onClick={() =>
            commit({
              variables: { action: 'reset' },
              onCompleted: () => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'success',
                    message: 'Successfully reset',
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
              onError: (error) => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'error',
                    message: 'Failed to reset',
                    details: cptlz(error.message),
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
            })
          }
          disabled={data?.isDone}
          size='large'
          variant='outlined'
          sx={buttonSx}
        >
          {data?.isDone ? (
            <NoResetIcon sx={iconSx} />
          ) : (
            <ResetIcon sx={iconSx} />
          )}
          Reset
        </Button>
      </Grid>
      <Grid item>
        <Button
          onClick={() =>
            commit({
              variables: { action: 'execute' },
              onCompleted: () => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'success',
                    message: 'Successfully executed',
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
              onError: (error) => {
                setStatusNotices([
                  ...statusNotices,
                  {
                    type: 'error',
                    message: 'Failed to execute',
                    details: cptlz(error.message),
                    endNote: DateTime.local().toFormat('fff'),
                  },
                ])
              },
            })
          }
          disabled={!data?.isIdle}
          size='large'
          variant='outlined'
          sx={buttonSx}
        >
          {!data?.isIdle ? (
            <NoExecuteIcon sx={iconSx} />
          ) : (
            <ExecuteIcon sx={iconSx} />
          )}
          Execute
        </Button>
      </Grid>
    </Grid>
  )
}
