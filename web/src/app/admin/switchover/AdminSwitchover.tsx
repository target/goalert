import React, { useState } from 'react'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Grid from '@mui/material/Grid'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import { SvgIconProps } from '@mui/material'
import PingIcon from 'mdi-material-ui/DatabaseMarker'
import NoResetIcon from 'mdi-material-ui/DatabaseRefreshOutline'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdleIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { gql, useMutation, useQuery } from '@apollo/client'
import { DateTime } from 'luxon'
import { SWONode as SWONodeType } from '../../../schema'
import Notices, { Notice } from '../../details/Notices'
import SWONode from './SWONode'
import LoadingButton from '@mui/lab/LoadingButton'

const query = gql`
  query {
    swoStatus {
      isDone
      isIdle
      isResetting
      isExecuting
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
  const [lastAction, setLastAction] = useState('')
  const [statusNotices, setStatusNotices] = useState<Notice[]>([])
  const [commit, mutationStatus] = useMutation(mutation)

  function actionHandler(action: 'ping' | 'reset' | 'execute'): () => void {
    return () => {
      setLastAction(action)
      commit({
        variables: {
          action,
        },
        onError: (error) => {
          setStatusNotices([
            ...statusNotices,
            {
              type: 'error',
              message: 'Failed to ' + action,
              details: cptlz(error.message),
              endNote: DateTime.local().toFormat('fff'),
            },
          ])
        },
      })
    }
  }

  const pingLoad = lastAction === 'ping' && mutationStatus.loading
  const resetLoad =
    data?.isResetting || (lastAction === 'reset' && mutationStatus.loading)
  const executeLoad =
    data?.isExecuting || (lastAction === 'execute' && mutationStatus.loading)

  function getIcon(): React.ReactNode {
    const i: SvgIconProps = { color: 'primary', sx: { fontSize: '3.5rem' } }

    if (error) {
      return <ErrorIcon {...i} color='error' />
    }
    if (loading && !data) {
      return (
        <Skeleton variant='circular'>
          <InProgressIcon {...i} />
        </Skeleton>
      )
    }
    if (!data.isIdle && !data.isDone) {
      return <InProgressIcon {...i} />
    }
    if (data.isIdle) {
      return <IdleIcon {...i} />
    }
  }

  function getSubheader(): React.ReactNode {
    if (error) return 'Error'
    if (!data) return 'Loading...'
    if (data.isDone) return 'Complete'
    if (data.isIdle) return 'Ready'
    if (!data.isExecuting && !data.isResetting) return 'Needs Reset'
    return 'Busy'
  }

  function getDetails(): React.ReactNode {
    if (error) {
      return (
        <Typography color='error' sx={{ pb: 2 }}>
          {cptlz(error.message)}
        </Typography>
      )
    }
    if (data?.details) {
      return <Typography sx={{ pb: 2 }}>{cptlz(data.details)}</Typography>
    }
    return null
  }

  return (
    <Grid container spacing={4}>
      {statusNotices.length > 0 && (
        <Grid item xs={12}>
          <Notices notices={statusNotices.reverse()} />
        </Grid>
      )}

      <Grid item>
        <Card sx={{ width: '350px' }}>
          <CardHeader
            title='Switchover Status'
            titleTypographyProps={{ sx: { fontSize: '1.25rem' } }}
            avatar={getIcon()}
            subheader={getSubheader()}
            sx={{ pb: 0 }}
          />
          <CardContent>
            {getDetails()}
            <ButtonGroup orientation='vertical' sx={{ width: '100%' }}>
              <LoadingButton
                startIcon={<PingIcon />}
                variant='outlined'
                size='large'
                disabled={mutationStatus.loading}
                loading={pingLoad}
                loadingPosition='start'
                onClick={actionHandler('ping')}
              >
                {pingLoad ? 'Sending ping...' : 'Ping'}
              </LoadingButton>
              <LoadingButton
                startIcon={data?.isDone ? <NoResetIcon /> : <ResetIcon />}
                disabled={data?.isDone || mutationStatus.loading}
                variant='outlined'
                size='large'
                loading={
                  data?.isResetting ||
                  (lastAction === 'reset' && mutationStatus.loading)
                }
                loadingPosition='start'
                onClick={actionHandler('reset')}
              >
                {resetLoad ? 'Resetting...' : 'Reset'}
              </LoadingButton>
              <LoadingButton
                startIcon={!data?.isIdle ? <NoExecuteIcon /> : <ExecuteIcon />}
                disabled={!data?.isIdle || mutationStatus.loading}
                variant='outlined'
                size='large'
                loading={
                  data?.isExecuting ||
                  (lastAction === 'execute' && mutationStatus.loading)
                }
                loadingPosition='start'
                onClick={actionHandler('execute')}
              >
                {executeLoad ? 'Executing...' : 'Execute'}
              </LoadingButton>
            </ButtonGroup>
          </CardContent>
        </Card>
      </Grid>

      <Grid item container>
        {data?.nodes.length > 0 &&
          data.nodes
            .slice()
            .sort((a: SWONodeType, b: SWONodeType) => {
              if (a.id < b.id) return 1
              if (a.id > b.id) return -1
              return 0
            })
            .map((node: SWONodeType, idx: number) => (
              <SWONode key={idx} node={node} index={idx} />
            ))}
      </Grid>
    </Grid>
  )
}
