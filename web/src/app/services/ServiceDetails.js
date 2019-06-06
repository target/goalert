import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { Link } from 'react-router-dom'
import PageActions from '../util/PageActions'
import Query from '../util/Query'
import OtherActions from '../util/OtherActions'
import DetailsPage from '../details/DetailsPage'
import ServiceOnCallQuery from './ServiceOnCallQuery'
import ServiceEditDialog from './ServiceEditDialog'
import ServiceDeleteDialog from './ServiceDeleteDialog'
import SetFavoriteButton from './components/SetFavoriteButton'

const query = gql`
  query serviceDetailsQuery($serviceID: ID!) {
    service(id: $serviceID) {
      id
      name
      description
      ep: escalationPolicy {
        id
        name
      }
    }
    alerts(
      input: {
        filterByStatus: [StatusAcknowledged, StatusUnacknowledged]
        filterByServiceID: [$serviceID]
        first: 1
      }
    ) {
      nodes {
        id
        status
      }
    }
  }
`

const titleQuery = gql`
  query titleQuery($serviceID: ID!) {
    service(id: $serviceID) {
      id
      name
      description
    }
  }
`

export default class ServiceDetails extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  state = {
    edit: false,
    delete: false,
  }

  renderData = ({ data }) => {
    let alertStatus
    if (data.alerts) {
      if (!data.alerts.nodes || !data.alerts.nodes.length) {
        alertStatus = 'ok'
      } else if (data.alerts.nodes[0].status === 'StatusUnacknowledged') {
        alertStatus = 'err'
      } else {
        alertStatus = 'warn'
      }
    }

    let titleFooter = null
    if (data.service.ep) {
      titleFooter = (
        <div>
          Escalation Policy:&nbsp;
          <Link to={`/escalation-policies/${data.service.ep.id}`}>
            {data.service.ep.name}
          </Link>
        </div>
      )
    }
    return (
      <React.Fragment>
        <PageActions>
          <SetFavoriteButton serviceID={data.service.id} />
          <OtherActions
            actions={[
              {
                label: 'Edit Service',
                onClick: () => this.setState({ edit: true }),
              },
              {
                label: 'Delete Service',
                onClick: () => this.setState({ delete: true }),
              },
            ]}
          />
        </PageActions>
        <DetailsPage
          title={data.service.name}
          details={data.service.description}
          titleFooter={titleFooter}
          links={[
            {
              label: 'Alerts',
              status: alertStatus,
              url: 'alerts',
            },
            {
              label: 'Integration Keys',
              url: 'integration-keys',
            },
            {
              label: 'Labels',
              url: 'labels',
            },
          ]}
          pageFooter={<ServiceOnCallQuery serviceID={this.props.serviceID} />}
        />
        {this.state.edit && (
          <ServiceEditDialog
            onClose={() => this.setState({ edit: false })}
            serviceID={this.props.serviceID}
          />
        )}
        {this.state.delete && (
          <ServiceDeleteDialog
            onClose={() => this.setState({ delete: false })}
            serviceID={this.props.serviceID}
          />
        )}
      </React.Fragment>
    )
  }

  render() {
    return (
      <Query
        query={query}
        partialQuery={titleQuery}
        variables={{ serviceID: this.props.serviceID }}
        render={this.renderData}
      />
    )
  }
}
