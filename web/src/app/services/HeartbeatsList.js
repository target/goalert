import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import withStyles from '@material-ui/core/styles/withStyles'
import gql from 'graphql-tag'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import Query from '../util/Query'
import IconButton from '@material-ui/core/IconButton'
import { Trash } from '../icons'
import HeartbeatCreateDialog from './HeartbeatCreateDialog'
import HeartbeatDeleteDialog from './HeartbeatDeleteDialog'

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

const styles = {
  text: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
}

const sortItems = (a, b) => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

@withStyles(styles)
class HeartbeatDetails extends React.PureComponent {
  static propTypes = {
    // timeoutMinutes: p.int.isRequired,
    // lastState: p.string.isRequired,
    // lastHeartbeatTime: p.string.isRequired,
    // provided by withStyles
    classes: p.object,
  }

  render() {
    return (
      <React.Fragment>
        <div>
          Sends an alert if no beat is received within
          {this.props.timeoutMinutes} minutes from the last timestamp.
        </div>
        <div>Last known state: {this.props.lastState}</div>
        <div>Last report time: {this.props.lastHeartbeatTime}</div>
      </React.Fragment>
    )
  }
}

@withStyles(styles)
export default class HeartbeatsList extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  state = {
    create: false,
    delete: null,
  }

  render() {
    return (
      <React.Fragment>
        <Grid item xs={12} className={this.props.classes.spacing}>
          <Card>
            <CardContent>{this.renderQuery()}</CardContent>
          </Card>
        </Grid>
        <CreateFAB onClick={() => this.setState({ create: true })} />
        {this.state.create && (
          <HeartbeatCreateDialog
            serviceID={this.props.serviceID}
            onClose={() => this.setState({ create: false })}
          />
        )}
        {this.state.delete && (
          <HeartbeatDeleteDialog
            HeartbeatID={this.state.delete}
            onClose={() => this.setState({ delete: null })}
          />
        )}
      </React.Fragment>
    )
  }

  renderQuery() {
    return (
      <Query
        query={query}
        variables={{ serviceID: this.props.serviceID }}
        render={({ data }) => this.renderList(data.service.heartbeatMonitors)}
      />
    )
  }

  renderList(monitors) {
    const items = (monitors || [])
      .slice()
      .sort(sortItems)
      .map(monitor => ({
        title: monitor.name,
        subText: (
          <HeartbeatDetails
            timeoutMinutes={monitor.timeoutMinutes}
            lastState={monitor.lastState}
            lastHeartbeatTime={this.props.lastHeartbeatTime}
            classes={this.props.classes}
          />
        ),
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
}
