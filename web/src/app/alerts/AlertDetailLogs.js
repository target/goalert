import React, { useState } from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import { makeStyles } from '@material-ui/core'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'
import _ from 'lodash-es'
import { formatTimeSince, formatTimeLocale } from '../util/timeFormat'
import { POLL_INTERVAL } from '../config'
import { textColors } from '../styles/statusStyles'

const FETCH_LIMIT = 149
const QUERY_LIMIT = 35

const query = gql`
  query getAlert($id: Int!, $input: AlertRecentEventsOptions) {
    alert(id: $id) {
      recentEvents(input: $input) {
        nodes {
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
  // colors generated from status colors, but with saturation locked at 75 and value locked at 52.5
  // so that all three passed contrast requirements (WCAG 2 AA)
  ...textColors,
})

export default function AlertDetailLogs(props) {
  const classes = useStyles()
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

  const getLogStatusClass = (status) => {
    switch (status) {
      case 'OK':
        return classes.statusOk
      case 'WARN':
        return classes.statusWarn
      case 'ERROR':
        return classes.statusError
      default:
        return null
    }
  }

  const renderItem = (event, idx) => {
    const details = _.upperFirst(event?.state?.details ?? '')
    const status = event?.state?.status ?? ''
    const detailsProps = {
      classes: {
        root: getLogStatusClass(status),
      },
    }

    let timestamp = formatTimeSince(event.timestamp)
    if (props.showExactTimes) {
      timestamp = formatTimeLocale(event.timestamp, 'full')
    }

    return (
      <ListItem key={idx} divider>
        <ListItemText
          primary={event.message}
          secondary={details}
          secondaryTypographyProps={detailsProps}
        />
        <ListItemSecondaryAction>
          <ListItemText secondary={timestamp} />
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
    events.map((event, idx) => renderItem(event, idx)),
    pageInfo.hasNextPage,
  )
}

AlertDetailLogs.propTypes = {
  alertID: p.number.isRequired,
  showExactTimes: p.bool,
}
