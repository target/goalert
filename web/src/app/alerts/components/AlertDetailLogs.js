import React from 'react'
import p from 'prop-types'
import Divider from '@material-ui/core/Divider'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Button from '@material-ui/core/Button'
import moment from 'moment'
import gql from 'graphql-tag'
import { useQuery } from '@apollo/react-hooks'

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
  let logs = []
  const { data, startPolling, stopPolling, fetchMore } = useQuery(query, {
    variables: { id: props.alertID, input: { after: '', limit: 35 } },
  })
  if (data) {
    if (data.alert.data.nodes.length === 0) {
      return (
        <ListItem>
          <ListItemText primary='No events.' />
        </ListItem>
      )
    }
    if (data.alert.data.nodes.length <= 35) {
      startPolling(2500)
    } else stopPolling()

    if (props.showExactTimes) {
      logs = data.alert.data.nodes.map((log, index) => (
        <div key={index}>
          <Divider />
          <ListItem>
            <ListItemText
              primary={moment(log.timestamp)
                .local()
                .format('MMM Do YYYY, h:mm:ss a')}
              secondary={log.message}
            />
          </ListItem>
        </div>
      ))
    } else {
      logs = data.alert.data.nodes.map((log, index) => (
        <div key={index}>
          <Divider />
          <ListItem>
            <ListItemText
              primary={moment(log.timestamp)
                .local()
                .calendar()}
              secondary={log.message}
            />
          </ListItem>
        </div>
      ))
    }
  }
  if (data && data.alert.data.pageInfo.hasNextPage) {
    return (
      <List style={{ padding: 0 }}>
        {logs}
        <Button
          style={{ width: '100%' }}
          onClick={() =>
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
          variant='outlined'
          data-cy='load-more-logs'
        >
          Load More
        </Button>
      </List>
    )
  }
  return (
    <List data-cy='alert-logs' style={{ padding: 0 }}>
      {logs}
    </List>
  )
}
AlertDetailLogs.propTypes = {
  alertID: p.number,
  showExactTimes: p.bool,
}
