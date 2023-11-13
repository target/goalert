import React, { useState } from 'react'
import { Chip, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { DateTime } from 'luxon'
import { gql } from 'urql'
import AdminMessageLogsControls from './AdminMessageLogsControls'
import AdminMessageLogDrawer from './AdminMessageLogDrawer'
import { DebugMessage } from '../../../schema'
import AdminMessageLogsGraph from './AdminMessageLogsGraph'
import toTitleCase from '../../util/toTitleCase'
import QueryList from '../../lists/QueryList'
import { PaginatedListItemProps } from '../../lists/PaginatedList'
import { useMessageLogsParams } from './util'

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

export default function AdminMessageLogsLayout(): React.ReactNode {
  const classes = useStyles()
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const [{ search, start, end }] = useMessageLogsParams()

  const logsInput = {
    search,
    createdAfter: start,
    createdBefore: end,
  }

  function mapLogToListItem(log: DebugMessage): PaginatedListItemProps {
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
      onClick: () => setSelectedLog(log as DebugMessage),
      selected: (log as DebugMessage).id === selectedLog?.id,
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
      action: (
        <Typography variant='body2' color='textSecondary'>
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
          <QueryList
            query={query}
            variables={{ input: { ...logsInput } }}
            noSearch
            mapDataNode={(n) => mapLogToListItem(n as DebugMessage)}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
