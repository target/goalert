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
import OtherActions from '../util/OtherActions'

import ServiceLabelSetDialog from './ServiceLabelCreateDialog'
import ServiceLabelEditDialog from './ServiceLabelEditDialog'
import ServiceLabelDeleteDialog from './ServiceLabelDeleteDialog'

const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      labels {
        key
        value
      }
    }
  }
`

const styles = () => ({
  spacing: { marginBottom: 96 },
})

const sortItems = (a, b) => {
  if (a.key.toLowerCase() < b.key.toLowerCase()) return -1
  if (a.key.toLowerCase() > b.key.toLowerCase()) return 1
  if (a.key < b.key) return -1
  if (a.key > b.key) return 1
  return 0
}

@withStyles(styles)
export default class ServiceLabelList extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  state = {
    create: false,
    editKey: null,
    deleteKey: null,
  }

  renderQuery() {
    return (
      <Query
        query={query}
        variables={{ serviceID: this.props.serviceID }}
        render={({ data }) => this.renderList(data.service.labels)}
      />
    )
  }

  renderList(labels) {
    const items = (labels || [])
      .slice()
      .sort(sortItems)
      .map(label => ({
        title: label.key,
        subText: label.value,
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => this.setState({ editKey: label.key }),
              },
              {
                label: 'Delete',
                onClick: () => this.setState({ deleteKey: label.key }),
              },
            ]}
          />
        ),
      }))

    return (
      <FlatList
        data-cy='label-list'
        emptyMessage='No labels exist for this service.'
        items={items}
      />
    )
  }

  render() {
    return (
      <React.Fragment>
        <Grid item xs={12} className={this.props.classes.spacing}>
          <Card>
            <CardContent>{this.renderQuery()}</CardContent>
          </Card>
        </Grid>
        <CreateFAB
          onClick={() => this.setState({ create: true })}
          title='Add Label'
        />
        {this.state.create && (
          <ServiceLabelSetDialog
            serviceID={this.props.serviceID}
            onClose={() => this.setState({ create: false })}
          />
        )}
        {this.state.editKey && (
          <ServiceLabelEditDialog
            serviceID={this.props.serviceID}
            labelKey={this.state.editKey}
            onClose={() => this.setState({ editKey: null })}
          />
        )}
        {this.state.deleteKey && (
          <ServiceLabelDeleteDialog
            serviceID={this.props.serviceID}
            labelKey={this.state.deleteKey}
            onClose={() => this.setState({ deleteKey: null })}
          />
        )}
      </React.Fragment>
    )
  }
}
