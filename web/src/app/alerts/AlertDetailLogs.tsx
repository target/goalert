import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { useQuery as urqlUseQuery, gql as urqlGql } from 'urql'
import Button from '@mui/material/Button'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import makeStyles from '@mui/styles/makeStyles'
import _ from 'lodash'
import { POLL_INTERVAL } from '../config'
import { Time } from '../util/Time'
import {
  AlertLogEntry,
  MessageStatusHistory,
  NotificationStatus,
} from '../../schema'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  AlertColor,
} from '@mui/material'

const FETCH_LIMIT = 149
const QUERY_LIMIT = 35

const query = gql`
  query getAlert($id: Int!, $input: AlertRecentEventsOptions) {
    alert(id: $id) {
      id
      recentEvents(input: $input) {
        nodes {
          id
          timestamp
          message
          state {
            details
            status
          }
          messageID
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
`

const useStyles = makeStyles({
  logTimeContainer: {
    width: 'max-content',
  },
  sentAccordian: {},
})

const assertNever = (s: never): never => {
  throw new Error('Unknown notification status: ' + s)
}

const getLogStatusClass = (
  status: NotificationStatus,
): AlertColor | undefined => {
  switch (status) {
    case 'OK':
      return 'success'
    case 'WARN':
      return 'warning'
    case 'ERROR':
      return 'error'
    default:
      assertNever(status)
  }
}

const histQuery = urqlGql`
  query getMessageHistory($id: ID!) {
    messageStatusHistory(id: $id) {
      timestamp
      status
      details
    }
  }
`

interface MessageHistoryProps {
  messageID: string
  showExactTimes?: boolean
}
const noSuspense = {
  suspense: false,
}
function MessageHistory(props: MessageHistoryProps): React.ReactNode {
  const classes = useStyles()
  const [hist] = urqlUseQuery({
    query: histQuery,
    variables: { id: props.messageID },
    context: noSuspense,
  })

  const data: MessageStatusHistory[] = hist.data?.messageStatusHistory || []

  return (
    <List>
      {hist.fetching && <ListItem>Loading...</ListItem>}
      {hist.error && <ListItem>Error: {hist.error.message}</ListItem>}
      {data && data.length === 0 && <ListItem>No history found</ListItem>}
      {data &&
        data.map((entry, idx) => (
          <ListItem key={idx} sx={{ paddingRight: 0 }}>
            <ListItemText
              primary={
                'Status: ' + entry.status + (idx === 0 ? ' (current)' : '')
              }
              secondary={entry.details}
            />
            <div>
              <ListItemText
                className={classes.logTimeContainer}
                secondary={
                  <Time
                    time={entry.timestamp}
                    format={props.showExactTimes ? 'default' : 'relative'}
                  />
                }
              />
            </div>
          </ListItem>
        ))}
    </List>
  )
}

interface LogEventProps {
  event: AlertLogEntry
  showExactTimes?: boolean
}
function LogEvent(props: LogEventProps): React.ReactNode {
  const [expanded, setExpanded] = useState(false)
  if (!props.event.messageID)
    return (
      <ListItem divider>
        <LogEventHeader
          event={props.event}
          showExactTimes={props.showExactTimes}
        />
      </ListItem>
    )
  return (
    <ListItem divider>
      <Accordion
        disableGutters
        elevation={1}
        sx={{ boxShadow: 'none', width: '100%' }}
        expanded={expanded}
        onChange={(_, expanded) => setExpanded(expanded)}
      >
        <AccordionSummary
          aria-controls='panel1d-content'
          id='panel1d-header'
          sx={{ padding: 0 }}
        >
          <LogEventHeader
            event={props.event}
            showExactTimes={props.showExactTimes}
          />
        </AccordionSummary>
        <AccordionDetails sx={{ paddingRight: 0 }}>
          {expanded && (
            <MessageHistory
              messageID={props.event.messageID}
              showExactTimes={props.showExactTimes}
            />
          )}
        </AccordionDetails>
      </Accordion>
    </ListItem>
  )
}

function LogEventHeader(props: LogEventProps): React.ReactNode {
  const classes = useStyles()
  const details = _.upperFirst(props.event?.state?.details ?? '')
  const status = (props.event?.state?.status ?? '') as NotificationStatus
  return (
    <React.Fragment>
      <ListItemText
        primary={props.event.message}
        secondary={details}
        secondaryTypographyProps={{
          color: status && getLogStatusClass(status),
        }}
      />
      <div>
        <ListItemText
          className={classes.logTimeContainer}
          secondary={
            <Time
              time={props.event.timestamp}
              format={props.showExactTimes ? 'default' : 'relative'}
            />
          }
        />
      </div>
    </React.Fragment>
  )
}

interface AlertDetailLogsProps {
  alertID: number
  showExactTimes?: boolean
}

export default function AlertDetailLogs(
  props: AlertDetailLogsProps,
): React.JSX.Element {
  const [poll, setPoll] = useState(POLL_INTERVAL)
  const { data, error, loading, fetchMore } = useQuery(query, {
    pollInterval: poll,
    variables: { id: props.alertID, input: { limit: QUERY_LIMIT } },
  })

  const events = _.orderBy(
    data?.alert?.recentEvents?.nodes ?? [],
    ['timestamp'],
    ['desc'],
  )
  const pageInfo = _.get(data, 'alert.recentEvents.pageInfo', {})

  const doFetchMore = (): void => {
    setPoll(0)
    fetchMore({
      variables: {
        id: props.alertID,
        input: {
          after: pageInfo.endCursor,
          limit: FETCH_LIMIT,
        },
      },
      updateQuery: (prev, { fetchMoreResult }) => {
        if (!fetchMoreResult) return prev
        return {
          alert: {
            ...fetchMoreResult.alert,
            recentEvents: {
              ...fetchMoreResult.alert.recentEvents,
              nodes: prev.alert.recentEvents.nodes.concat(
                fetchMoreResult.alert.recentEvents.nodes,
              ),
            },
          },
        }
      },
    })
  }

  const renderList = (
    items: React.JSX.Element | JSX.Element[],
    loadMore?: boolean,
  ): React.JSX.Element => {
    return (
      <List data-cy='alert-logs'>
        {items}
        {loadMore && (
          <Button
            style={{ width: '100%' }}
            onClick={doFetchMore}
            variant='outlined'
            data-cy='load-more-logs'
          >
            Load More
          </Button>
        )}
      </List>
    )
  }

  if (events.length === 0) {
    return renderList(
      <ListItem>
        <ListItemText primary='No events.' />
      </ListItem>,
    )
  }

  if (error) {
    return renderList(
      <ListItem>
        <ListItemText primary={error.message} />
      </ListItem>,
    )
  }

  if (loading && !data) {
    return renderList(
      <ListItem>
        <ListItemText primary='Loading...' />
      </ListItem>,
    )
  }

  return renderList(
    events.map((event, idx) => (
      <LogEvent key={idx} showExactTimes={props.showExactTimes} event={event} />
    )),
    pageInfo.hasNextPage,
  )
}
