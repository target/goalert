import React, { useMemo, useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Grid } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import AdminMessageLogsList from './AdminMessageLogsList'
import AdminMessageLogsControls from './AdminMessageLogsControls'
import AdminMessageLogDrawer from './AdminMessageLogDrawer'
import { DebugMessage } from '../../../schema'

import AdminMessageLogsGraph from './AdminMessageLogsGraph'
import { useWorker } from '../../worker'
import { Options } from './useMessageLogs'
import { useURLParam, useURLParams } from '../../actions'

export const MAX_QUERY_ITEMS_COUNT = 1000
const RENDER_AMOUNT = 50

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
}))

export default function AdminMessageLogsLayout(): JSX.Element {
  const classes = useStyles()

  // all data is fetched on page load, but the number of logs rendered is limited
  const [numRendered, setNumRendered] = useState(RENDER_AMOUNT)
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  // graph duration set with ISO duration values, e.g. P1D for a daily duration
  const [duration] = useURLParam<string>('interval', 'P1D')

  const { data, loading, error } = useQuery(debugMessageLogsQuery, {
    variables: { first: MAX_QUERY_ITEMS_COUNT },
  })

  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  // useMemo to use same object reference
  const opts: Options = useMemo(
    () => ({
      data: (data?.debugMessages as DebugMessage[]) ?? [],
      start: params.start,
      end: params.end,
      search: params.search,
      duration,
    }),
    [data?.debugMessages, params.start, params.end, params.search, duration],
  )

  const messageLogData = useWorker('useMessageLogs', opts, {
    graphData: [],
    filteredData: [],
  })

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  const [{ graphData, filteredData }] = messageLogData

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
          <AdminMessageLogsControls
            resetCount={
              () => setNumRendered(RENDER_AMOUNT) // reset to # of first page results
            }
          />
        </Grid>
        {graphData.length > 0 && (
          <Grid item xs={12}>
            <AdminMessageLogsGraph
              data={graphData}
              totalCount={filteredData.length}
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <AdminMessageLogsList
            debugMessages={filteredData}
            selectedLog={selectedLog}
            onSelect={setSelectedLog}
            hasMore={numRendered < filteredData.length}
            onLoadMore={() => setNumRendered(numRendered + RENDER_AMOUNT)}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
