import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import DebugMessagesList from './DebugMessagesList'
import DebugMessagesControls from './DebugMessagesControls'
import DebugMessageDetails from './DebugMessageDetails'
import { theme } from '../../mui'
import { DebugMessage } from '../../../schema'

export const MAX_QUERY_ITEMS_COUNT = 1000
const DEFAULT_LOAD_AMOUNT = 50
const LOAD_AMOUNT = 50

const debugMessageLogsQuery = gql`
  query debugMessageLogsQuery($first: Int!) {
    debugMessages(input: { first: $first }) {
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

export default function AdminDebugMessagesLayout(): JSX.Element {
  const classes = useStyles()
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)
  const [showingLimit, setShowingLimit] = useState(DEFAULT_LOAD_AMOUNT)

  const { data, loading, error } = useQuery(debugMessageLogsQuery, {
    variables: { first: MAX_QUERY_ITEMS_COUNT },
  })

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  return (
    <React.Fragment>
      <DebugMessageDetails
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
              Outgoing Message Logs
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <DebugMessagesControls
              showingLimit={showingLimit}
              totalCount={data.debugMessages.length}
            />
          </Grid>
          <Grid item xs={12}>
            <DebugMessagesList
              debugMessages={data.debugMessages.map((d: DebugMessage) => ({
                ...d,
                additionalKeys: {
                  filteredDestination: d.destination.replace('-', ''),
                },
              }))}
              selectedLog={selectedLog}
              onSelect={setSelectedLog}
              showingLimit={showingLimit}
              onResetLoadMore={() => setShowingLimit(DEFAULT_LOAD_AMOUNT)}
              onLoadMore={() => setShowingLimit(showingLimit + LOAD_AMOUNT)}
            />
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
