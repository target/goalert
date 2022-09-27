import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Chip, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import DebugMessagesControls from './DebugMessagesControls'
import DebugMessageDetails from './DebugMessageDetails'
import { DebugMessage } from '../../../schema'
import { useURLParams } from '../../actions'
import SimpleListPage from '../../lists/SimpleListPage'
import { DateTime } from 'luxon'
import toTitleCase from '../../util/toTitleCase'

const query = gql`
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
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
  card: {
    margin: theme.spacing(1),
    cursor: 'pointer',
  },
  textField: {
    backgroundColor: 'white',
    borderRadius: '4px',
    minWidth: 250,
  },
}))

export default function AdminDebugMessagesLayout(): JSX.Element {
  const classes = useStyles()

  // all data is fetched on page load, but the number of logs rendered is limited
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const [params, setParams] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  const { data, loading, error } = useQuery(query, {
    variables: {
      createdAfter: params.start,
      createdBefore: params.end,
    },
  })

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  return (
    <React.Fragment>
      <DebugMessageDetails
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
          <DebugMessagesControls
            value={params}
            onChange={(newParams) => {
              setParams(newParams)
            }}
          />
        </Grid>
        <Grid item xs={12} container spacing={2}>
          <SimpleListPage
            query={query}
            noSearch
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
