import React, { useState } from 'react'
import { Chip, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { DateTime } from 'luxon'
import { gql, useQuery } from 'urql'
import AdminMessageLogsControls from './AdminMessageLogsControls'
import AdminMessageLogDrawer from './AdminMessageLogDrawer'
import { DebugMessage, MessageLogConnection } from '../../../schema'
import AdminMessageLogsGraph from './AdminMessageLogsGraph'
import toTitleCase from '../../util/toTitleCase'
import { useMessageLogsParams } from './util'
import ListPageControls from '../../lists/ListPageControls'
import FlatList, { FlatListListItem } from '../../lists/FlatList'

const query = gql`
  query messageLogsQuery($input: MessageLogSearchOptions) {
    messageLogs(input: $input) {
      nodes {
        id
        createdAt
        updatedAt
        type
        status
        userID
        userName
        source
        destination
        serviceID
        serviceName
        alertID
        providerID
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const useStyles = makeStyles((theme: Theme) => ({
  containerDefault: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '100%',
      transition: `max-width ${theme.transitions.duration.leavingScreen}ms ease`,
    },
  },
  containerSelected: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '70%',
      transition: `max-width ${theme.transitions.duration.enteringScreen}ms ease`,
    },
  },
}))

const context = { suspense: false }

export default function AdminMessageLogsLayout(): JSX.Element {
  const classes = useStyles()
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const [{ search, start, end }] = useMessageLogsParams()
  const [cursor, setCursor] = useState('')

  const logsInput = {
    search,
    createdAfter: start,
    createdBefore: end,
    after: cursor,
  }

  const [q] = useQuery<{ messageLogs: MessageLogConnection }>({
    query,
    variables: { input: logsInput },
    context,
  })
  const nextCursor = q.data?.messageLogs.pageInfo.hasNextPage
    ? q.data?.messageLogs.pageInfo.endCursor
    : ''

  function mapLogToListItem(log: DebugMessage): FlatListListItem {
    const status = toTitleCase(log.status)
    const statusDict = {
      success: {
        backgroundColor: '#EDF6ED',
        color: '#1D4620',
      },
      error: {
        backgroundColor: '#FDEBE9',
        color: '#611A15',
      },
      warning: {
        backgroundColor: '#FFF4E5',
        color: '#663C00',
      },
      info: {
        backgroundColor: '#E8F4FD',
        color: '#0E3C61',
      },
    }

    let statusStyles
    const s = status.toLowerCase()
    if (s.includes('deliver')) statusStyles = statusDict.success
    if (s.includes('sent')) statusStyles = statusDict.success
    if (s.includes('fail')) statusStyles = statusDict.error
    if (s.includes('temp')) statusStyles = statusDict.warning
    if (s.includes('pend')) statusStyles = statusDict.info

    return {
      onClick: () => setSelectedLog(log),
      selected: log.id === selectedLog?.id,
      title: `${log.type} Notification`,
      subText: (
        <Grid container spacing={2} direction='column'>
          <Grid item>Destination: {log.destination}</Grid>
          {log.serviceName && <Grid item>Service: {log.serviceName}</Grid>}
          {log.userName && <Grid item>User: {log.userName}</Grid>}
          <Grid item>
            <Chip label={status} style={statusStyles} />
          </Grid>
        </Grid>
      ),
      secondaryAction: (
        <Typography variant='body2' component='div' color='textSecondary'>
          {DateTime.fromISO(log.createdAt).toFormat('fff')}
        </Typography>
      ),
    }
  }

  return (
    <React.Fragment>
      <AdminMessageLogDrawer
        onClose={() => setSelectedLog(null)}
        log={selectedLog}
      />
      <Grid
        container
        spacing={2}
        className={
          selectedLog ? classes.containerSelected : classes.containerDefault
        }
      >
        <Grid item xs={12}>
          <AdminMessageLogsControls />
        </Grid>

        <AdminMessageLogsGraph />

        <Grid item xs={12}>
          <ListPageControls
            nextCursor={nextCursor}
            onCursorChange={setCursor}
            loading={q.fetching}
            slots={{
              list: (
                <FlatList
                  emptyMessage='No results'
                  items={q.data?.messageLogs.nodes.map(mapLogToListItem) || []}
                />
              ),
            }}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
