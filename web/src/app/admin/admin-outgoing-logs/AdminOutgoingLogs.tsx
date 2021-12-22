import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Box, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import OutgoingLogsList from './OutgoingLogsList'
import OutgoingLogsFilter from './OutgoingLogsFilter'
import OutgoingLogDetails from './OutgoingLogDetails'
import { theme } from '../../mui'
import { DebugMessage } from '../../../schema'
import Search from '../../util/Search'

const debugMessageLogsQuery = gql`
  query debugMessageLogsQuery {
    debugMessages(input: { first: 1000 }) {
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
  }
`

const useStyles = makeStyles<typeof theme>((theme) => ({
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

export default function AdminOutgoingLogs(): JSX.Element {
  const classes = useStyles()
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const { data, loading, error } = useQuery(debugMessageLogsQuery)

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  return (
    <React.Fragment>
      <OutgoingLogDetails
        open={Boolean(selectedLog)}
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
        <Grid container item xs={12}>
          <Grid item xs={12}>
            <Typography
              component='h2'
              variant='subtitle1'
              color='textSecondary'
              classes={{ subtitle1: classes.groupTitle }}
            >
              Outgoing Messages
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Box
              display='flex'
              flexDirection='row'
              alignItems='flex-end'
              justifyContent='space-between'
            >
              <div>
                <OutgoingLogsFilter />
              </div>
              <div style={{ paddingBottom: '.25rem' }}>
                <Search />
              </div>
            </Box>
          </Grid>
          <Grid item xs={12}>
            <OutgoingLogsList
              debugMessages={data.debugMessages}
              onSelect={setSelectedLog}
            />
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
