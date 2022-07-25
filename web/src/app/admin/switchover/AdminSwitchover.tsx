import React, { useEffect, useState } from 'react'
import { useTheme, SvgIconProps, Zoom } from '@mui/material'
import Alert from '@mui/material/Alert'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Grid from '@mui/material/Grid'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import NoResetIcon from 'mdi-material-ui/DatabaseRefreshOutline'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdleIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { gql, useMutation, useQuery } from 'urql'
import { DateTime } from 'luxon'
import { SWOAction, SWONode as SWONodeType, SWOStatus } from '../../../schema'
import Notices, { Notice } from '../../details/Notices'
import SWONode from './SWONode'
import LoadingButton from '@mui/lab/LoadingButton'
import DatabaseOff from 'mdi-material-ui/DatabaseOff'
import DatabaseCheck from 'mdi-material-ui/DatabaseCheck'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Tooltip from '@mui/material/Tooltip'
import RemoveIcon from '@mui/icons-material/PlaylistRemove'
import AddIcon from '@mui/icons-material/PlaylistAdd'
import DownIcon from '@mui/icons-material/ArrowDownward'
import { TransitionGroup } from 'react-transition-group'
import Spinner from '../../loading/components/Spinner'

const query = gql`
  query {
    swoStatus {
      state
      lastError
      lastStatus
      mainDBVersion
      nextDBVersion
      connections {
        name
        count
      }
      nodes {
        id
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
  const [{ fetching, error, data: _data }, refetch] = useQuery({
    query,
  })
  const data = _data?.swoStatus as SWOStatus
  const [lastAction, setLastAction] = useState('')
  const [mutationStatus, commit] = useMutation(mutation)
  const theme = useTheme()

  const curVer = data?.mainDBVersion.split(' on ')
  const nextVer = data?.mainDBVersion.split(' on ')

  useEffect(() => {
    const t = setInterval(() => {
      if (!fetching) refetch()
    }, 1000)
    return () => clearInterval(t)
  }, [])

  if (fetching) {
    return <Spinner />
  }

  if (error && error.message === '[GraphQL] not in SWO mode') {
    return (
      <Grid item container alignItems='center' justifyContent='center'>
        <DatabaseOff color='secondary' style={{ width: '100%', height: 256 }} />
        <Grid item>
          <div style={{ textAlign: 'center' }}>
            <Typography
              color='secondary'
              variant='h6'
              style={{ marginTop: 16 }}
            >
              Unavailable: Application is not in switchover mode.
              <br />
              <br />
              You must start GoAlert with <code>
                GOALERT_DB_URL_NEXT
              </code> or <code>--db-url-next</code> to perform a switchover.
            </Typography>
          </div>
        </Grid>
      </Grid>
    )
  }

  if (data?.state === 'done') {
    return (
      <TransitionGroup appear={false}>
        <Zoom in timeout={500}>
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

  function actionHandler(action: 'reset' | 'execute'): () => void {
    return () => {
      setLastAction(action)
      commit({ action }, { additionalTypenames: ['SWOStatus'] })
    }
  }
  const statusNotices: Notice[] = []
  if (mutationStatus.error) {
    const vars: { action?: SWOAction } = mutationStatus.operation
      ?.variables || {
      action: '',
    }
    statusNotices.push({
      type: 'error',
      message: 'Failed to ' + vars.action,
      details: cptlz(mutationStatus.error.message),
      endNote: DateTime.local().toFormat('fff'),
    })
  }
  if (data?.state === 'unknown' && data?.lastError) {
    statusNotices.push({
      type: 'error',
      message: data.lastError,
    })
  }

  const resetLoad =
    data?.state === 'resetting' ||
    (lastAction === 'reset' && mutationStatus.fetching)
  const executeLoad =
    ['syncing', 'pausing', 'executing'].includes(data?.state) ||
    (lastAction === 'execute' && mutationStatus.fetching)

  function getIcon(): React.ReactNode {
    const i: SvgIconProps = { color: 'primary', sx: { fontSize: '3.5rem' } }

    if (error) {
      return <ErrorIcon {...i} color='error' />
    }
    if (fetching && !data) {
      return (
        <Skeleton variant='circular'>
          <InProgressIcon {...i} />
        </Skeleton>
      )
    }
    if (!['unknown', 'idle', 'done'].includes(data.state)) {
      return <InProgressIcon {...i} />
    }
    if (data.state === 'idle') {
      return <IdleIcon {...i} />
    }
  }

  function getSubheader(): React.ReactNode {
    if (error) return 'Error'
    if (!data) return 'Loading...'
    if (data.state === 'done') return 'Complete'
    if (data.state === 'idle') return 'Ready'
    if (data.state === 'unknown') return 'Needs Reset'
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
    if (data?.state !== 'unknown' && data.lastStatus) {
      return <Typography sx={{ pb: 2 }}>{cptlz(data.lastStatus)}</Typography>
    }
    return <Typography>&nbsp;</Typography> // reserves whitespace
  }

  const headerSize = { titleTypographyProps: { sx: { fontSize: '1.25rem' } } }

  return (
    <Grid container spacing={2}>
      {statusNotices.length > 0 && (
        <Grid item xs={12}>
          <Notices notices={statusNotices.reverse()} />
        </Grid>
      )}
      <Grid item xs={12} sm={12} md={12} lg={4} xl={4}>
        <Card sx={{ height: '100%' }}>
          <CardContent
            sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}
          >
            <CardHeader
              title='Switchover Status'
              avatar={getIcon()}
              subheader={getSubheader()}
              {...headerSize}
              sx={{ p: 0 }}
            />
            {getDetails()}
            <div style={{ flexGrow: 1 }} />
            <ButtonGroup
              orientation={
                theme.breakpoints.up('md') ? 'vertical' : 'horizontal'
              }
              sx={{ width: '100%', pb: '32px' }}
            >
              <LoadingButton
                startIcon={
                  data?.state === 'done' ? <NoResetIcon /> : <ResetIcon />
                }
                disabled={data?.state === 'done' || mutationStatus.fetching}
                variant='outlined'
                size='large'
                loading={resetLoad}
                loadingPosition='start'
                onClick={actionHandler('reset')}
              >
                {resetLoad ? 'Resetting...' : 'Reset'}
              </LoadingButton>
              <LoadingButton
                startIcon={
                  data?.state !== 'idle' ? <NoExecuteIcon /> : <ExecuteIcon />
                }
                disabled={data?.state !== 'idle' || mutationStatus.fetching}
                variant='outlined'
                size='large'
                loading={executeLoad}
                loadingPosition='start'
                onClick={actionHandler('execute')}
              >
                {executeLoad ? 'Executing...' : 'Execute'}
              </LoadingButton>
            </ButtonGroup>
          </CardContent>
        </Card>
      </Grid>

      <Grid item xs={12} sm={12} md={12} lg={8} xl={8}>
        <Card sx={{ height: '100%' }}>
          <CardHeader title='Database Connections' {...headerSize} />
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Application</TableCell>
                <TableCell>Info</TableCell>
                <TableCell align='right'>Count</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data?.connections?.map((row) => (
                <TableRow
                  key={row.name || '(no name)'}
                  selected={
                    row.name.includes('GoAlert') && !row.name.includes('SWO')
                  }
                  sx={{
                    '&:last-child td, &:last-child th': { border: 0 },
                  }}
                >
                  <TableCell component='th' scope='row'>
                    {row?.name?.split('(')[0].replace(/[)(]/g, '') ??
                      '(no name)'}
                  </TableCell>
                  <TableCell component='th' scope='row'>
                    {row?.name?.split('(')[1]?.replace(/[)(]/g, '') ?? '-'}
                  </TableCell>
                  <TableCell align='right'>{row.count}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      </Grid>

      <Grid item xs={12} container spacing={2} justifyContent='space-between'>
        <Grid item xs={12} sm={6} lg={4} xl={3} sx={{ width: '100%' }}>
          <Card sx={{ height: '100%' }}>
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                padding: '0 16px 0 16px',
                marginBottom: '16px',
                height: '100%',
              }}
            >
              <CardHeader title='DB Diff' {...headerSize} />
              <Tooltip title={curVer[1]}>
                <Alert icon={<RemoveIcon />} severity='error'>
                  From {curVer[0]}
                </Alert>
              </Tooltip>
              <DownIcon
                style={{ flexGrow: 1 }}
                sx={{
                  alignSelf: 'center',
                  color: (theme) => theme.palette.primary.main,
                }}
              />
              <Tooltip title={nextVer[1]}>
                <Alert
                  icon={<AddIcon />}
                  severity='success'
                  sx={{ mb: '16px' }}
                >
                  To {nextVer[0]}
                </Alert>
              </Tooltip>
            </div>
          </Card>
        </Grid>

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
              <SWONode key={node.id} node={node} name={friendlyName(node.id)} />
            ))}
      </Grid>
    </Grid>
  )
}
