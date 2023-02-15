import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import p from 'prop-types'
import Button from '@mui/material/Button'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import _ from 'lodash'
import { formatTimeSince } from '../util/timeFormat'
import { POLL_INTERVAL } from '../config'

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

export default function AlertDetailLogs(props) {
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

  const getLogStatusClass = (status) => {
    switch (status) {
      case 'OK':
        return 'success'
      case 'WARN':
        return 'warning'
      case 'ERROR':
        return 'error'
      default:
        return null
    }
  }

  const renderItem = (event, idx) => {
    const details = _.upperFirst(event?.state?.details ?? '')
    const status = event?.state?.status ?? ''

    let timestamp = formatTimeSince(event.timestamp)
    if (props.showExactTimes) {
      timestamp = DateTime.fromISO(event.timestamp).toLocaleString(
        DateTime.DATETIME_FULL,
      )
    }

    return (
      <ListItem key={idx} divider>
        <ListItemText
          primary={event.message}
          secondary={details}
          secondaryTypographyProps={{ color: getLogStatusClass(status) }}
        />
        <div>
          <ListItemText
            className={classes.logTimeContainer}
            secondary={timestamp}
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

AlertDetailLogs.propTypes = {
  alertID: p.number.isRequired,
  showExactTimes: p.bool,
}
