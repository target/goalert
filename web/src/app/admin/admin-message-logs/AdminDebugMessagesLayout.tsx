import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import DebugMessagesList from './DebugMessagesList'
import DebugMessagesControls from './DebugMessagesControls'
import DebugMessageDetails from './DebugMessageDetails'
import { DebugMessage } from '../../../schema'
import { useURLParams } from '../../actions'
import { DateTime } from 'luxon'
import { useFuse } from './useFuse'

export const MAX_QUERY_ITEMS_COUNT = 1000
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
  const [numRendered, setNumRendered] = useState(LOAD_AMOUNT)
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const { data, loading, error } = useQuery(debugMessageLogsQuery, {
    variables: { first: MAX_QUERY_ITEMS_COUNT },
  })

  const [params, setParams] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  const results = useFuse<DebugMessage>({
    data:
      data?.debugMessages.map((d: DebugMessage) => ({
        ...d,
        additionalKeys: {
          filteredDestination: d.destination.replace('-', ''),
        },
      })) || [],
    keys: [
      'destination',
      'userName',
      'serviceName',
      'status',
      'additionalKeys.filteredDestination',
    ],
    search: params.search,
    options: {
      shouldSort: false,
      showResultsWhenNoSearchTerm: true,
      ignoreLocation: true,
      useExtendedSearch: true,
    },
  })

  const startDT = params.start ? DateTime.fromISO(params.start) : null
  const endDT = params.end ? DateTime.fromISO(params.end) : null

  const filteredResults = results.filter((result) => {
    const createdAtDT = DateTime.fromISO(result.item.createdAt)
    if (startDT && startDT > createdAtDT) return false
    if (endDT && endDT < createdAtDT) return false
    return true
  })

  const displayedResults = filteredResults.slice(0, numRendered)

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
              value={params}
              onChange={(newParams) => {
                setParams(newParams)
                setNumRendered(LOAD_AMOUNT)
              }}
              displayedCount={displayedResults.length}
              resultsCount={filteredResults.length}
            />
          </Grid>
          <Grid item xs={12}>
            <DebugMessagesList
              debugMessages={displayedResults}
              selectedLog={selectedLog}
              onSelect={setSelectedLog}
              hasMore={numRendered < results.length}
              onLoadMore={() => setNumRendered(numRendered + LOAD_AMOUNT)}
            />
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
