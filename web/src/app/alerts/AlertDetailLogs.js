import React, { useState } from 'react'
import p from 'prop-types'
import Divider from '@material-ui/core/Divider'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Button from '@material-ui/core/Button'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import { DateTime } from 'luxon'
import { logTimeFormat } from '../util/timeFormat'
import { POLL_INTERVAL } from '../config'
import _ from 'lodash-es'

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
  const renderItems = (timestamp, message) => {
    if (props.showExactTimes)
      return (
        <ListItemText
          primary={DateTime.fromISO(timestamp).toFormat(
            'MMM dd yyyy, h:mm:ss a',
          )}
          secondary={message}
        />
      )
    return (
      <ListItemText
        primary={logTimeFormat(timestamp, DateTime.local().startOf('day'))}
        secondary={message}
      />
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
    events.map((log, index) => (
      <div key={index}>
        <Divider />
        <ListItem>{renderItems(log.timestamp, log.message)}</ListItem>
      </div>
    )),
    pageInfo.hasNextPage,
  )
}
AlertDetailLogs.propTypes = {
  alertID: p.number.isRequired,
  showExactTimes: p.bool,
}
