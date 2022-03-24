import React, { useState } from 'react'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Divider from '@mui/material/Divider'
import Grid from '@mui/material/Grid'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import Skeleton from '@mui/material/Skeleton'
import Typography from '@mui/material/Typography'
import { useTheme, SvgIconProps } from '@mui/material'
import PingIcon from 'mdi-material-ui/DatabaseMarker'
import NoResetIcon from 'mdi-material-ui/DatabaseRefreshOutline'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdleIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { gql, useMutation, useQuery } from '@apollo/client'
import TrueIcon from 'mdi-material-ui/CheckboxMarkedCircleOutline'
import FalseIcon from 'mdi-material-ui/CloseCircleOutline'
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
  const theme = useTheme()
  const { loading, error, data: _data } = useQuery(query)
  const data = _data?.swoStatus

  const [statusNotices, setStatusNotices] = useState<Notice[]>([])
  const [commit] = useMutation(mutation)

  function getIcon(): React.ReactNode {
    const i: SvgIconProps = { color: 'primary', sx: { fontSize: '3.5rem' } }

    if (error) {
      return <ErrorIcon {...i} color='error' />
    }
    if (loading) {
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
    if (loading) return 'Loading...'
    if (!data.isIdle && !data.isDone) return 'In progress'
    if (data.isIdle) return 'Idle'
    return null
  }

  function getDetails(): React.ReactNode {
    if (error) {
      return <Typography color='error'>{cptlz(error.message)}</Typography>
    }
    if (data?.details) {
      return cptlz(data.details)
    }
    return 'Testing some details yeehaw'
  }

  return (
    <Grid container spacing={4}>
      {statusNotices.length > 0 && (
        <Grid item xs={12}>
          <Notices notices={statusNotices.reverse()} />
        </Grid>
      )}

      <Grid item xs={4}>
        <Card sx={{ width: '100%' }}>
          <CardHeader
            title='Switchover Status'
            titleTypographyProps={{ sx: { fontSize: '1.25rem' } }}
            avatar={getIcon()}
            subheader={getSubheader()}
            sx={{ pb: 0 }}
          />
          <CardContent>
            <Typography sx={{ pb: 2 }}>{getDetails()}</Typography>
            <ButtonGroup orientation='vertical' sx={{ width: '100%' }}>
              <Button
                startIcon={<PingIcon />}
                variant='outlined'
                size='large'
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
              >
                Ping
              </Button>
              <Button
                startIcon={data?.isDone ? <NoResetIcon /> : <ResetIcon />}
                disabled={data?.isDone}
                variant='outlined'
                size='large'
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
              >
                Reset
              </Button>
              <Button
                startIcon={!data?.isIdle ? <NoExecuteIcon /> : <ExecuteIcon />}
                disabled={!data?.isIdle}
                variant='outlined'
                size='large'
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
              >
                Execute
              </Button>
            </ButtonGroup>
          </CardContent>
        </Card>
      </Grid>

      <Grid item xs={8} container>
        {data?.nodes.length > 0 &&
          data.nodes
            .sort((a: SWONode, b: SWONode) => {
              if (a.id < b.id) return 1
              if (a.id > b.id) return -1
              return 0
            })
            .map((node: SWONode, idx: number) => (
              <Grid item key={idx} sx={{ minWidth: 300 }}>
                <Card>
                  <Typography color={theme.palette.primary.main} sx={{ p: 2 }}>
                    Node {idx + 1}
                  </Typography>
                  <List
                    subheader={
                      <React.Fragment>
                        <Divider />
                        <Typography
                          color={theme.palette.secondary.main}
                          sx={{ p: 2 }}
                        >
                          Status: {node.status}
                        </Typography>
                        <Divider />
                      </React.Fragment>
                    }
                  >
                    <ListItem>
                      <ListItemText primary='Executable?' />
                      <ListItemSecondaryAction>
                        {node.canExec ? (
                          <TrueIcon color='success' />
                        ) : (
                          <FalseIcon color='error' />
                        )}
                      </ListItemSecondaryAction>
                    </ListItem>
                    <ListItem>
                      <ListItemText primary='Is the old node valid?' />
                      <ListItemSecondaryAction>
                        {node.oldValid ? (
                          <TrueIcon color='success' />
                        ) : (
                          <FalseIcon color='error' />
                        )}
                      </ListItemSecondaryAction>
                    </ListItem>
                    <ListItem>
                      <ListItemText primary='Is the new node valid?' />
                      <ListItemSecondaryAction>
                        {node.newValid ? (
                          <TrueIcon color='success' />
                        ) : (
                          <FalseIcon color='error' />
                        )}
                      </ListItemSecondaryAction>
                    </ListItem>
                  </List>
                </Card>
              </Grid>
            ))}
      </Grid>
    </Grid>
  )
}
