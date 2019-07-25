import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import gql from 'graphql-tag'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import IconButton from '@material-ui/core/IconButton'
import { Trash } from '../icons'
import HeartbeatCreateDialog from './HeartbeatCreateDialog'
import HeartbeatDeleteDialog from './HeartbeatDeleteDialog'
import { makeStyles } from '@material-ui/core'

const query = gql`
  query($serviceID: ID!) {
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

export default function HeartbeatsList(props) {
  const classes = useStyles()
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState(null)

  function renderList(monitors) {
    const items = (monitors || [])
      .slice()
      .sort(sortItems)
      .map(monitor => ({
        title: monitor.name,
        /* subText: (
          <HeartbeatDetails
            timeoutMinutes={monitor.timeoutMinutes}
            lastState={monitor.lastState}
            lastHeartbeatTime={
              this.props.lastHeartbeatTime
                ? this.props.lastHeartbeatTime
                : 'not yet reported'
            }
            classes={this.props.classes}
          />
        ), */
        action: (
          <IconButton onClick={() => this.setState({ delete: monitor.id })}>
            <Trash />
          </IconButton>
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
      {Boolean(showDeleteDialogByID) && (
        <HeartbeatDeleteDialog
          HeartbeatID={showDeleteDialogByID}
          onClose={() => setShowDeleteDialogByID(null)}
        />
      )}
    </React.Fragment>
  )
}
HeartbeatsList.propTypes = {
  serviceID: p.string.isRequired,
}
