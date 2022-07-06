import React, { useState } from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import { Redirect } from 'wouter'
import _ from 'lodash'
import InputLabel from '@mui/material/InputLabel'
import MenuItem from '@mui/material/MenuItem'
import FormControl from '@mui/material/FormControl'
import Select from '@mui/material/Select'
import { Edit, Delete } from '@mui/icons-material'

import DetailsPage from '../details/DetailsPage'
import ServiceEditDialog from './ServiceEditDialog'
import ServiceDeleteDialog from './ServiceDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import ServiceOnCallList from './ServiceOnCallList'
import AppLink from '../util/AppLink'
import { ServiceAvatar } from '../util/avatars'
import { DateTime } from 'luxon'

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

const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
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

  const [setMaintenanceMode, setMaintenanceModeStatus] = useMutation(mutation)

  if (loading && !_.get(data, 'service.id')) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (!_.get(data, 'service.id')) {
    return showDelete ? <Redirect to='/services' /> : <ObjectNotFound />
  }

  const mm = data.service.maintenanceExpiresAt // maintenance mode

  return (
    <React.Fragment>
      Maintenance Expires At: {data.service.maintenanceExpiresAt}
      <DetailsPage
        avatar={<ServiceAvatar />}
        title={data.service.name}
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
          <FormControl key='maintenance-expires-at' fullWidth>
            <InputLabel id='maintenance-expires-at-label'>
              Maintenance Mode
            </InputLabel>
            <Select
              labelId='maintenance-expires-at-label'
              id='maintenance-expires-at'
              label={mm ? 'Extend Maintenance Mode' : 'Start Maintenance Mode'}
              onChange={(val) => {
                const expireDate = DateTime.now().plus({ hours: val }).toISO()
                setMaintenanceMode({
                  variables: {
                    id: serviceID,
                    maintenanceExpiresAt: expireDate,
                  },
                })
              }}
            >
              <MenuItem value={1}>1 hour from now</MenuItem>
              <MenuItem value={2}>2 hours from now</MenuItem>
              <MenuItem value={4}>4 hours from now</MenuItem>
            </Select>
          </FormControl>,
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
            status: alertStatus(_.get(data, 'alerts.nodes')),
            url: 'alerts',
            subText: 'Manage alerts specific to this service',
          },
          {
            label: 'Heartbeat Monitors',
            url: 'heartbeat-monitors',
            status: hbStatus(_.get(data, 'service.heartbeatMonitors')),
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
