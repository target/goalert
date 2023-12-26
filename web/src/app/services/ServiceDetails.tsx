import React, { Suspense, useState } from 'react'
import { gql, useQuery } from 'urql'
import { Redirect } from 'wouter'
import _ from 'lodash'
import { Button, Chip } from '@mui/material'
import { Edit, Delete } from '@mui/icons-material'

import DetailsPage, { LinkStatus } from '../details/DetailsPage'
import ServiceEditDialog from './ServiceEditDialog'
import ServiceDeleteDialog from './ServiceDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import ServiceOnCallList from './ServiceOnCallList'
import AppLink from '../util/AppLink'
import { ServiceAvatar } from '../util/avatars'
import ServiceMaintenanceModeDialog from './ServiceMaintenanceDialog'
import ServiceNotices from './ServiceNotices'
import type { HeartbeatMonitor, Label } from '../../schema'

interface AlertNode {
  id: string
  status: string
}

const query = gql`
  fragment ServiceTitleQuery on Service {
    id
    name
    description
  }

  query serviceDetailsQuery($serviceID: ID!) {
    service(id: $serviceID) {
      ...ServiceTitleQuery
      maintenanceExpiresAt
      labels {
        key
        value
      }
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

const hbStatus = (h: Partial<HeartbeatMonitor>[]): LinkStatus | null => {
  if (!h || !h.length) return null
  if (h.every((m) => m.lastState === 'healthy')) return 'ok'
  if (h.some((m) => m.lastState === 'unhealthy')) return 'err'
  return 'warn'
}

const alertStatus = (a: AlertNode[]): LinkStatus | null => {
  if (!a) return null
  if (!a.length) return 'ok'
  if (a[0].status === 'StatusUnacknowledged') return 'err'
  return 'warn'
}

export default function ServiceDetails(props: {
  serviceID: string
}): JSX.Element {
  const { serviceID } = props
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [showMaintMode, setShowMaintMode] = useState(false)
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { serviceID },
  })

  if (fetching && !_.get(data, 'service.id')) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!_.get(data, 'service.id')) {
    return showDelete ? <Redirect to='/services' /> : <ObjectNotFound />
  }

  const labels: Label[] = data.service.labels || []

  return (
    <React.Fragment>
      <DetailsPage
        avatar={<ServiceAvatar />}
        title={data.service.name}
        notices={<ServiceNotices serviceID={serviceID} />}
        labels={labels}
        subheader={
          <React.Fragment>
            Escalation Policy:{' '}
            {_.get(data, 'service.ep') ? (
              <AppLink to={`/escalation-policies/${data.service.ep.id}`}>
                {data.service.ep.name}
              </AppLink>
            ) : (
              <Spinner text='Looking up policy...' />
            )}
          </React.Fragment>
        }
        details={data.service.description}
        pageContent={<ServiceOnCallList serviceID={serviceID} />}
        primaryActions={[
          <Button
            color='primary'
            variant='contained'
            key='maintence-mode'
            onClick={() => setShowMaintMode(true)}
            aria-label='Maintenance Mode'
          >
            Maintenance Mode
          </Button>,
        ]}
        secondaryActions={[
          {
            label: 'Edit',
            icon: <Edit />,
            handleOnClick: () => setShowEdit(true),
          },
          {
            label: 'Delete',
            icon: <Delete />,
            handleOnClick: () => setShowDelete(true),
          },
          <QuerySetFavoriteButton
            key='secondary-action-favorite'
            id={serviceID}
            type='service'
          />,
        ]}
        links={[
          {
            label: 'Alerts',
            status: alertStatus(data?.alerts?.nodes),
            url: 'alerts',
            subText: 'Manage alerts specific to this service',
          },
          {
            label: 'Heartbeat Monitors',
            url: 'heartbeat-monitors',
            status: hbStatus(data?.service?.heartbeatMonitors),
            subText: 'Manage endpoints monitored for you',
          },
          {
            label: 'Integration Keys',
            url: 'integration-keys',
            subText: 'Manage keys used to create alerts',
          },
          {
            label: 'Labels',
            url: 'labels',
            subText: 'Group together services',
          },
          {
            label: 'Alert Metrics',
            url: 'alert-metrics',
            subText: 'Review alert activity',
          },
        ]}
      />
      <Suspense>
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
        {showMaintMode && (
          <ServiceMaintenanceModeDialog
            onClose={() => setShowMaintMode(false)}
            serviceID={serviceID}
            expiresAt={data.service.maintenanceExpiresAt}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
