import React, { useState } from 'react'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Grid from '@mui/material/Grid'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import { Fade, SvgIconProps, Tooltip, Zoom } from '@mui/material'
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
import { SWONode as SWONodeType, SWOStatus } from '../../../schema'
import Notices, { Notice } from '../../details/Notices'
import SWONode from './SWONode'
import LoadingButton from '@mui/lab/LoadingButton'
import DatabaseOff from 'mdi-material-ui/DatabaseOff'
import DatabaseCheck from 'mdi-material-ui/DatabaseCheck'
import { Info } from '@mui/icons-material'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Paper from '@mui/material/Paper'
import { TransitionGroup } from 'react-transition-group'
import Spinner from '../../loading/components/Spinner'

const query = gql`
  query {
    swoStatus {
      isDone
      isIdle
      isResetting
      isExecuting
      details
      errors
      connections {
        name
        count
      }
      nodes {
        id
        status
        canExec
        oldValid
        newValid
        isLeader
      }
    }
  }
`

let n = 1
const names: { [key: string]: string } = {}

// friendlyName will assign a persistant "friendly" name to the node.
//
// This ensures a specific ID will always refer to the same node. This
// is so that it is clear if a node dissapears or a new one appears.
//
// Note: `Node 1` on one browser tab may not be the same node as `Node 1`
// on another browser tab.
function friendlyName(id: string): string {
  if (!names[id]) {
    names[id] = `Node ${n++}`
  }
  return names[id]
}

const mutation = gql`
  mutation ($action: SWOAction!) {
    swoAction(action: $action)
  }
`

function cptlz(s: string): string {
  return s.charAt(0).toUpperCase() + s.substring(1)
}

export default function AdminSwitchover(): JSX.Element {
  const { loading, error, data: _data } = useQuery(query, { pollInterval: 250 })
  const data = _data?.swoStatus as SWOStatus
  const [lastAction, setLastAction] = useState('')
  const [_statusNotices, setStatusNotices] = useState<Notice[]>([])
  const [commit, mutationStatus] = useMutation(mutation)

  if (loading) {
    return <Spinner />
  }

  if (error && error.message == 'not in SWO mode') {
    return (
      <Grid item container alignItems='center' justifyContent='center'>
        <DatabaseOff color='secondary' style={{ width: '100%', height: 256 }} />
        <Grid item>
          <Typography color='secondary' variant='h6' style={{ marginTop: 16 }}>
            Unavailable: Application is not in switchover mode.{' '}
            <Tooltip
              title='--db-url-next or GOALERT_DB_URL must be set. See SWO documentation.'
              placement='top'
            >
              <Info />
            </Tooltip>
          </Typography>
        </Grid>
      </Grid>
    )
  }

  if (data?.isDone) {
    return (
      <TransitionGroup appear={false}>
        <Zoom in={true} timeout={500}>
          <Grid item container alignItems='center' justifyContent='center'>
            <DatabaseCheck
              color='primary'
              style={{ width: '100%', height: 256 }}
            />
            <Grid item>
              <Typography
                color='primary'
                variant='h6'
                style={{ marginTop: 16 }}
              >
                DB switchover is complete.
              </Typography>
            </Grid>
          </Grid>
        </Zoom>
      </TransitionGroup>
    )
  }

  function actionHandler(action: 'ping' | 'reset' | 'execute'): () => void {
    return () => {
      setLastAction(action)
      commit({
        variables: {
          action,
        },
        onError: (error) => {
          setStatusNotices([
            ..._statusNotices,
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

  const statusNotices = _statusNotices.concat(
    (data?.errors ?? []).map((message: string) => ({
      type: 'error',
      message,
    })),
  )

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
    <TransitionGroup appear={false}>
      <Fade out>
        <Grid container spacing={4}>
          {statusNotices.length > 0 && (
            <Grid item xs={12}>
              <Notices notices={statusNotices.reverse()} />
            </Grid>
          )}
          <Grid item container>
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
                      startIcon={
                        !data?.isIdle ? <NoExecuteIcon /> : <ExecuteIcon />
                      }
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
            <Grid item paddingLeft={1}>
              <Card>
                <CardHeader title='Database Connections' />
                <TableContainer component={Paper}>
                  <Table size='small'>
                    <TableHead>
                      <TableRow>
                        <TableCell>Application Name</TableCell>
                        <TableCell align='right'>Count</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {data?.connections?.map((row) => (
                        <TableRow
                          key={row.name || '(no name)'}
                          sx={{
                            '&:last-child td, &:last-child th': { border: 0 },
                          }}
                        >
                          <TableCell component='th' scope='row'>
                            {row.name || '(no name)'}
                          </TableCell>
                          <TableCell align='right'>{row.count}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Card>
            </Grid>
          </Grid>
          <Grid item container>
            {data?.nodes.length > 0 &&
              data.nodes
                .slice()
                .sort((a: SWONodeType, b: SWONodeType) => {
                  const aName = friendlyName(a.id)
                  const bName = friendlyName(b.id)
                  if (aName < bName) return -1
                  if (aName > bName) return 1
                  return 0
                })
                .map((node: SWONodeType) => (
                  <SWONode
                    key={node.id}
                    node={node}
                    name={friendlyName(node.id)}
                  />
                ))}
          </Grid>
        </Grid>
      </Fade>
    </TransitionGroup>
  )
}
