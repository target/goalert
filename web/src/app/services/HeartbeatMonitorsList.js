import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import gql from 'graphql-tag'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import HeartbeatCreateDialog from './HeartbeatCreateDialog'
import { makeStyles } from '@material-ui/core'
import HeartbeatMonitorListItem, {
  HeartbeatMonitorListItemActions,
} from './HeartbeatMonitorListItem'

const query = gql`
  query monitorQuery($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      heartbeatMonitors {
        id
        name
        timeoutMinutes
        lastState
        lastHeartbeat
      }
    }
  }
`

const useStyles = makeStyles(theme => ({
  text: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
}))

const sortItems = (a, b) => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

export default function HeartbeatMonitorsList(props) {
  const classes = useStyles()
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  function renderList(monitors) {
    const items = (monitors || [])
      .slice()
      .sort(sortItems)
      .map(monitor => ({
        title: monitor.name,
        subText: (
          <HeartbeatMonitorListItem
            timeoutMinutes={monitor.timeoutMinutes}
            lastState={monitor.lastState}
            lastHeartbeatTime={
              monitor.lastHeartbeat ? monitor.lastHeartbeat : 'not yet reported'
            }
          />
        ),
        secondaryAction: (
          <HeartbeatMonitorListItemActions
            monitorID={monitor.id}
            refetchQueries={['monitorQuery']}
          />
        ),
      }))

    return (
      <FlatList
        data-cy='monitors'
        emptyMessage='No monitors exist for this service.'
        items={items}
      />
    )
  }

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            <Query
              query={query}
              variables={{ serviceID: props.serviceID }}
              render={({ data }) => renderList(data.service.heartbeatMonitors)}
            />
          </CardContent>
        </Card>
      </Grid>
      <CreateFAB onClick={() => setShowCreateDialog(true)} />
      {showCreateDialog && (
        <HeartbeatCreateDialog
          serviceID={props.serviceID}
          onClose={() => setShowCreateDialog(false)}
        />
      )}
    </React.Fragment>
  )
}
HeartbeatMonitorsList.propTypes = {
  serviceID: p.string.isRequired,
}
