import React, { useState } from 'react'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Grid from '@mui/material/Grid'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import { SvgIconProps, TypographyProps } from '@mui/material'
import PingIcon from 'mdi-material-ui/DatabaseMarker'
import NoResetIcon from 'mdi-material-ui/DatabaseRefreshOutline'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdleIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { gql, useMutation, useQuery } from '@apollo/client'
import Notices, { Notice } from '../../details/Notices'
import { DateTime } from 'luxon'
import { SWONode } from '../../../schema'

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
    const i: SvgIconProps = { color: 'primary', sx: { fontSize: '3.5rem' } }
    const t: TypographyProps = {
      variant: 'caption',
      sx: { display: 'flex' },
      flexDirection: 'column',
    }

    if (error) {
      return (
        <Typography {...t}>
          <ErrorIcon {...i} color='error' />
          Error
        </Typography>
      )
    }

    if (loading) {
      return (
        <Typography {...t}>
          <Skeleton variant='circular'>
            <InProgressIcon {...i} />
          </Skeleton>
          Loading...
        </Typography>
      )
    }

    if (!data.isIdle && !data.isDone) {
      return (
        <Typography {...t}>
          <InProgressIcon {...i} />
          In Progress
        </Typography>
      )
    }

    if (data.isIdle) {
      return (
        <Typography {...t}>
          <IdleIcon {...i} />
          Idle
        </Typography>
      )
    }
  }

  function getDetails(): React.ReactNode {
    if (error) {
      return <Typography color='error'>{cptlz(error.message)}</Typography>
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

      <Grid item xs={12} container>
        {data?.nodes.length > 0 &&
          data.nodes.map((node: SWONode, idx: number) => (
            <Grid item key={idx}>
              <Card>
                <CardHeader details={node.status} />
                <CardContent>
                  <Typography>
                    {node.canExec ? 'Executable' : 'Not Executable'}
                  </Typography>
                  <Typography>
                    {node.oldValid ? 'Old is valid' : 'Old is invalid'}
                  </Typography>
                  <Typography>
                    {node.newValid ? 'New is valid' : 'New is invalid'}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
      </Grid>
    </Grid>
  )
}
