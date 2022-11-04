import React, { useState } from 'react'
import { gql } from '@apollo/client'
import { Chip, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import AdminMessageLogsControls from './AdminMessageLogsControls'
import AdminMessageLogDrawer from './AdminMessageLogDrawer'
import { DebugMessage } from '../../../schema'

import AdminMessageLogsGraph from './AdminMessageLogsGraph'
import { useURLParams } from '../../actions'
import { DateTime } from 'luxon'
import SimpleListPage from '../../lists/SimpleListPage'
import toTitleCase from '../../util/toTitleCase'

export const query = gql`
  query messageLogsQuery($input: MessageLogSearchOptions) {
    data: messageLogs(input: $input) {
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

export default function AdminMessageLogsLayout(): JSX.Element {
  const classes = useStyles()

  // all data is fetched on page load, but the number of logs rendered is limited
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

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
          <SimpleListPage
            query={query}
            noSearch
            mapVariables={(vars) => {
              if (params.search) vars.input.search = params.search
              if (params.start) vars.input.createdAfter = params.start
              if (params.end) vars.input.createdBefore = params.end
              return vars
            }}
            mapDataNode={(n) => {
              const status = toTitleCase(n.status)
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
                onClick: () => setSelectedLog(n as DebugMessage),
                selected: (n as DebugMessage).id === selectedLog?.id,
                title: `${n.type} Notification`,
                subText: (
                  <Grid container spacing={2} direction='column'>
                    <Grid item>Destination: {n.destination}</Grid>
                    {n.serviceName && (
                      <Grid item>Service: {n.serviceName}</Grid>
                    )}
                    {n.userName && <Grid item>User: {n.userName}</Grid>}
                    <Grid item>
                      <Chip label={status} style={statusStyles} />
                    </Grid>
                  </Grid>
                ),
                action: (
                  <Typography variant='body2' color='textSecondary'>
                    {DateTime.fromISO(n.createdAt).toFormat('fff')}
                  </Typography>
                ),
              }
            }}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
