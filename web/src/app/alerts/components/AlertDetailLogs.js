import React, { useState } from 'react'
import p from 'prop-types'
import Divider from '@material-ui/core/Divider'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Button from '@material-ui/core/Button'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import { DateTime, Interval } from 'luxon'
import { POLL_INTERVAL } from '../../config'

const LIMIT = 149

const query = gql`
  query getAlert($id: Int!, $input: AlertRecentEventsOptions) {
    alert(id: $id) {
      data: recentEvents(input: $input) {
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
    variables: { id: props.alertID, input: { after: '', limit: 35 } },
  })

  const doFetchMore = () => {
    setPoll(0)
    fetchMore({
      variables: {
        id: props.alertID,
        input: {
          after: data.alert.data.pageInfo.endCursor,
          limit: LIMIT,
        },
      },
      updateQuery: (prev, { fetchMoreResult }) => {
        if (!fetchMoreResult) return prev
        return {
          alert: {
            ...fetchMoreResult.alert,
            data: {
              ...fetchMoreResult.alert.data,
              nodes: prev.alert.data.nodes.concat(
                fetchMoreResult.alert.data.nodes,
              ),
            },
          },
        }
      },
    })
  }
  const renderList = (items, loadMore) => {
    return (
      <List data-cy='alert-logs' style={{ padding: 0 }}>
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
      <ListItemText primary={relativeTime(timestamp)} secondary={message} />
    )
  }
  const relativeTime = timestamp => {
    const to = DateTime.fromISO(timestamp)
    const from = DateTime.local()
      .setZone(to.zone)
      .startOf('day')
    if (Interval.after(from, { days: 1 }).contains(to))
      return 'Today at ' + to.toFormat('h:mm a')
    if (Interval.before(from, { days: 1 }).contains(to))
      return 'Yesterday at ' + to.toFormat('h:mm a')
    if (Interval.before(from, { weeks: 1 }).contains(to))
      return 'Last ' + to.weekdayLong + ' at ' + to.toFormat('h:mm a')
    return to.toFormat('MM/dd/yyyy')
  }

  if (data && data.alert.data.nodes.length === 0) {
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
    data &&
      data.alert.data.nodes.map((log, index) => (
        <div key={index}>
          <Divider />
          <ListItem>{renderItems(log.timestamp, log.message)}</ListItem>
        </div>
      )),
    data && data.alert.data.pageInfo.hasNextPage,
  )
}
AlertDetailLogs.propTypes = {
  alertID: p.number.isRequired,
  showExactTimes: p.bool,
}
