import { gql } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import HeartbeatMonitorCreateDialog from './HeartbeatMonitorCreateDialog'
import { makeStyles } from '@material-ui/core'
import HeartbeatMonitorEditDialog from './HeartbeatMonitorEditDialog'
import HeartbeatMonitorDeleteDialog from './HeartbeatMonitorDeleteDialog'
import OtherActions from '../util/OtherActions'
import HeartbeatMonitorStatus from './HeartbeatMonitorStatus'
import CopyText from '../util/CopyText'

// generates a single alert if a POST is not received before the timeout
const HEARTBEAT_MONITOR_DESCRIPTION =
  'Heartbeat monitors create an alert if no heartbeat is received (a POST request) before the configured timeout.'

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
        href
      }
    }
  }
`

const useStyles = makeStyles(() => ({
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

export default function HeartbeatMonitorList(props) {
  const classes = useStyles()
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialogByID, setShowEditDialogByID] = useState(null)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)

  function renderList(monitors) {
    const items = (monitors || [])
      .slice()
      .sort(sortItems)
      .map((monitor) => ({
        icon: (
          <HeartbeatMonitorStatus
            lastState={monitor.lastState}
            lastHeartbeat={monitor.lastHeartbeat}
          />
        ),
        title: monitor.name,
        subText: (
          <React.Fragment>
            {`Timeout: ${monitor.timeoutMinutes} minute${
              monitor.timeoutMinutes > 1 ? 's' : ''
            }`}
            <br />
            <CopyText title='Copy URL' value={monitor.href} />
          </React.Fragment>
        ),
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setShowEditDialogByID(monitor.id),
              },
              {
                label: 'Delete',
                onClick: () => setShowDeleteDialogByID(monitor.id),
              },
            ]}
          />
        ),
      }))

    return (
      <FlatList
        data-cy='monitors'
        emptyMessage='No heartbeat monitors exist for this service.'
        headerNote={HEARTBEAT_MONITOR_DESCRIPTION}
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
      <CreateFAB
        onClick={() => setShowCreateDialog(true)}
        title='Create Heartbeat Monitor'
      />
      {showCreateDialog && (
        <HeartbeatMonitorCreateDialog
          serviceID={props.serviceID}
          onClose={() => setShowCreateDialog(false)}
        />
      )}
      {showEditDialogByID && (
        <HeartbeatMonitorEditDialog
          monitorID={showEditDialogByID}
          onClose={() => setShowEditDialogByID(null)}
        />
      )}
      {showDeleteDialogByID && (
        <HeartbeatMonitorDeleteDialog
          monitorID={showDeleteDialogByID}
          onClose={() => setShowDeleteDialogByID(null)}
        />
      )}
    </React.Fragment>
  )
}
HeartbeatMonitorList.propTypes = {
  serviceID: p.string.isRequired,
}
