import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import Button from '@mui/material/Button'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import makeStyles from '@mui/styles/makeStyles'
import _ from 'lodash'
import { POLL_INTERVAL } from '../config'
import { Time } from '../util/Time'
import { AlertLogEntry, NotificationStatus } from '../../schema'
import { AlertColor } from '@mui/material'

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
})

interface AlertDetailLogsProps {
  alertID: number
  showExactTimes?: boolean
}

export default function AlertDetailLogs(
  props: AlertDetailLogsProps,
): React.ReactNode {
  const classes = useStyles()
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
    items: React.ReactNode | React.ReactNode[],
    loadMore?: boolean,
  ): React.ReactNode => {
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

  const renderItem = (event: AlertLogEntry, idx: number): React.ReactNode => {
    const details = _.upperFirst(event?.state?.details ?? '')
    const status = (event?.state?.status ?? '') as NotificationStatus

    return (
      <ListItem key={idx} divider>
        <ListItemText
          primary={event.message}
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
                time={event.timestamp}
                format={props.showExactTimes ? 'default' : 'relative'}
              />
            }
          />
        </div>
      </ListItem>
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
    events.map((event, idx) => renderItem(event, idx)),
    pageInfo.hasNextPage,
  )
}
