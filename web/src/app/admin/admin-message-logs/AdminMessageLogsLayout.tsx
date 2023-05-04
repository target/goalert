import React, { useState } from 'react'
import { Card, Chip, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AdminMessageLogsControls from './AdminMessageLogsControls'
import AdminMessageLogDrawer from './AdminMessageLogDrawer'
import { DebugMessage } from '../../../schema'

import AdminMessageLogsGraph from './AdminMessageLogsGraph'
import { useURLParams } from '../../actions'
import toTitleCase from '../../util/toTitleCase'
import { useMessageLogs } from './useMessageLogs'
import { DateTime } from 'luxon'
import {
  PaginatedList,
  PaginatedListItemProps,
} from '../../lists/PaginatedList'
import { PageControls } from '../../lists/PageControls'

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

export default function AdminMessageLogsLayout(): JSX.Element {
  const classes = useStyles()
  const [page, setPage] = useState(0)

  // all data is fetched on page load, but the number of logs rendered is limited
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })
  const depKey = `logs-${params.start}-${params.end}`

  const { logs, loading, error } = useMessageLogs(
    {
      search: params.search,
      createdAfter: params.start || DateTime.now().minus({ hours: 8 }).toISO(),
      createdBefore: params.end || DateTime.now().toISO(),
    },
    depKey,
  )

  if (error) {
    return <div>Error: {error.message}</div>
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

        <AdminMessageLogsGraph logs={logs} loadingData={loading} />

        <Grid item xs={12}>
          <Card>
            <PaginatedList
              items={logs.map((log) => mapLogToListItem(log))}
              isLoading={loading}
              itemsPerPage={15}
              page={page}
            />
          </Card>
        </Grid>
        <Grid item xs={12}>
          <PageControls
            pageCount={logs.length / 15}
            page={page}
            setPage={setPage}
            isLoading={loading}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
