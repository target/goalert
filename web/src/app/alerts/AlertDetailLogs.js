import React, { useState } from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import Divider from '@material-ui/core/Divider'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import { DateTime } from 'luxon'
import _ from 'lodash-es'
import { logTimeFormat } from '../util/timeFormat'
import { POLL_INTERVAL } from '../config'
import { formatTimeSince } from '../util/timeFormat'

const FETCH_LIMIT = 149
const QUERY_LIMIT = 35

const query = gql`
  query getAlert($id: Int!, $input: AlertRecentEventsOptions) {
    alert(id: $id) {
      recentEvents(input: $input) {
        nodes {
          timestamp
          message
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
`
export default function AlertDetailLogs(props) {
  const [poll, setPoll] = useState(POLL_INTERVAL)
  const { data, error, loading, fetchMore } = useQuery(query, {
    pollInterval: poll,
    variables: { id: props.alertID, input: { limit: QUERY_LIMIT } },
  })

  const events = _.get(data, 'alert.recentEvents.nodes', [])
  const pageInfo = _.get(data, 'alert.recentEvents.pageInfo', {})

  const doFetchMore = () => {
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

  const renderList = (items, loadMore) => {
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

  const renderItem = (timestamp, message, idx) => {
    let timestampFmtd = formatTimeSince(timestamp)
    if (props.showExactTimes) {
      timestampFmtd = DateTime.fromISO(timestamp).toLocaleString(
        DateTime.DATETIME_FULL,
      )
    }

    return (
      <ListItem key={idx} divider>
        <ListItemText primary={message} />
        <ListItemSecondaryAction>
          <ListItemText secondary={timestampFmtd} />
        </ListItemSecondaryAction>
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
    events.map((event, idx) => renderItem(event.timestamp, event.message, idx)),
    pageInfo.hasNextPage,
  )
}

AlertDetailLogs.propTypes = {
  alertID: p.number.isRequired,
  showExactTimes: p.bool,
}
