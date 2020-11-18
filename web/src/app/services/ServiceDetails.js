import React, { useState } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import { Redirect } from 'react-router-dom'
import _ from 'lodash'

import PageActions from '../util/PageActions'
import OtherActions from '../util/OtherActions'
import DetailsPage from '../details/DetailsPage'
import ServiceEditDialog from './ServiceEditDialog'
import ServiceDeleteDialog from './ServiceDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import ServiceOnCallList from './ServiceOnCallList'
import AppLink from '../util/AppLink'

const query = gql`
  fragment ServiceTitleQuery on Service {
    id
    name
    description
  }

  query serviceDetailsQuery($serviceID: ID!) {
    service(id: $serviceID) {
      ...ServiceTitleQuery
      ep: escalationPolicy {
        id
        name
      }
      heartbeatMonitors {
        id
        lastState
      }
      onCallUsers {
        userID
        userName
        stepNumber
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

const hbStatus = (h) => {
  if (!h || !h.length) return null
  if (h.every((m) => m.lastState === 'healthy')) return 'ok'
  if (h.some((m) => m.lastState === 'unhealthy')) return 'err'
  return 'warn'
}

const alertStatus = (a) => {
  if (!a) return null
  if (!a.length) return 'ok'
  if (a[0].status === 'StatusUnacknowledged') return 'err'
  return 'warn'
}

export default function ServiceDetails({ serviceID }) {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const { data, loading, error } = useQuery(query, {
    variables: { serviceID },
    returnPartialData: true,
  })

  if (loading && !_.get(data, 'service.id')) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!_.get(data, 'service.id')) {
    return showDelete ? <Redirect to='/services' push /> : <ObjectNotFound />
  }

  return (
    <React.Fragment>
      <PageActions>
        <QuerySetFavoriteButton serviceID={serviceID} />
        <OtherActions
          actions={[
            {
              label: 'Edit Service',
              onClick: () => setShowEdit(true),
            },
            {
              label: 'Delete Service',
              onClick: () => setShowDelete(true),
            },
          ]}
        />
      </PageActions>
      <DetailsPage
        title={data.service.name}
        details={data.service.description}
        titleFooter={
          <div>
            Escalation Policy:{' '}
            {_.get(data, 'service.ep') ? (
              <AppLink to={`/escalation-policies/${data.service.ep.id}`}>
                {data.service.ep.name}
              </AppLink>
            ) : (
              <Spinner text='Looking up policy...' />
            )}
          </div>
        }
        links={[
          {
            label: 'Alerts',
            status: alertStatus(_.get(data, 'alerts.nodes')),
            url: 'alerts',
          },
          {
            label: 'Heartbeat Monitors',
            url: 'heartbeat-monitors',
            status: hbStatus(_.get(data, 'service.heartbeatMonitors')),
          },
          { label: 'Integration Keys', url: 'integration-keys' },
          { label: 'Labels', url: 'labels' },
        ]}
        pageFooter={<ServiceOnCallList serviceID={serviceID} />}
      />
      {showEdit && (
        <ServiceEditDialog
          onClose={() => setShowEdit(false)}
          serviceID={serviceID}
        />
      )}
      {showDelete && (
        <ServiceDeleteDialog
          onClose={() => setShowDelete(false)}
          serviceID={serviceID}
        />
      )}
    </React.Fragment>
  )
}

ServiceDetails.propTypes = {
  serviceID: p.string.isRequired,
}
