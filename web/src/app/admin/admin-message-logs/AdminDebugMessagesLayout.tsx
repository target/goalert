import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Grid } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import DebugMessagesList from './DebugMessagesList'
import DebugMessagesControls from './DebugMessagesControls'
import DebugMessageDetails from './DebugMessageDetails'
import { DebugMessage } from '../../../schema'
import { useURLParams } from '../../actions'
import { DateTime, Interval } from 'luxon'
import DebugMessageGraph from './DebugMessageGraph'

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
}))

export default function AdminDebugMessagesLayout(): JSX.Element {
  const classes = useStyles()

  // all data is fetched on page load, but the number of logs rendered is limited
  const [numRendered, setNumRendered] = useState(LOAD_AMOUNT)
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)

  const { data, loading, error } = useQuery(debugMessageLogsQuery, {
    variables: { first: MAX_QUERY_ITEMS_COUNT },
  })

  const [params] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  const startDT = params.start ? DateTime.fromISO(params.start) : null
  const endDT = params.end ? DateTime.fromISO(params.end) : null

  const filteredData: DebugMessage[] = data?.debugMessages
    .filter((msg: DebugMessage) => {
      const createdAtDT = DateTime.fromISO(msg.createdAt)
      if (params.search) {
        if (
          params.search === msg.alertID?.toString() ||
          params.search === msg.createdAt ||
          params.search === msg.destination ||
          params.search === msg.serviceID ||
          params.search === msg.serviceName ||
          params.search === msg.userID ||
          params.search === msg.userName
        ) {
          return true
        }
        return false
      }
      if (startDT && startDT > createdAtDT) return false
      if (endDT && endDT < createdAtDT) return false
      return true
    })
    .sort((_a: DebugMessage, _b: DebugMessage) => {
      const a = DateTime.fromISO(_a.createdAt)
      const b = DateTime.fromISO(_b.createdAt)
      if (a < b) return 1
      if (a > b) return -1
      return 0
    })

  const paginatedData = filteredData.slice(0, numRendered)
  const hasData = filteredData?.length > 0
  let ivl: Interval | null = null

  if (startDT && endDT && hasData) {
    ivl = Interval.fromDateTimes(startDT, endDT)
  } else if ((!startDT || !endDT) && hasData) {
    ivl = Interval.fromDateTimes(
      DateTime.fromISO(filteredData[filteredData.length - 1].createdAt).startOf(
        'day',
      ),
      DateTime.fromISO(filteredData[0].createdAt).endOf('day'),
    )
  }

  const intervalType = 'daily'
  const graphData = ivl
    ? ivl.splitBy({ days: 1 }).map((i) => {
        const date = i.start.toLocaleString({ month: 'short', day: 'numeric' })
        const label = i.start.toLocaleString({
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        })

        const dayCount = filteredData.filter((msg: DebugMessage) =>
          i.contains(DateTime.fromISO(msg.createdAt)),
        )

        return {
          date,
          label,
          count: dayCount.length,
        }
      })
    : []

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
            resetCount={
              () => setNumRendered(LOAD_AMOUNT) // reset to # of first page results
            }
          />
        </Grid>
        {paginatedData.length > 0 && (
          <Grid item xs={12}>
            <DebugMessageGraph
              data={graphData}
              intervalType={intervalType}
              totalCount={filteredData.length}
            />
          </Grid>
        )}
        <Grid item xs={12}>
          <DebugMessagesList
            debugMessages={paginatedData}
            selectedLog={selectedLog}
            onSelect={setSelectedLog}
            hasMore={numRendered < filteredData.length}
            onLoadMore={() => setNumRendered(numRendered + LOAD_AMOUNT)}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
