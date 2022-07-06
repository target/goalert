import React, { useState } from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import { Redirect } from 'wouter'
import _ from 'lodash'
import MenuItem from '@mui/material/MenuItem'
import { Edit, Delete } from '@mui/icons-material'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown'
import ClickAwayListener from '@mui/material/ClickAwayListener'
import Grow from '@mui/material/Grow'
import Paper from '@mui/material/Paper'
import Popper from '@mui/material/Popper'
import MenuList from '@mui/material/MenuList'

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

const options = ['1 hour', '2 hours', '4 hours']

export default function ServiceDetails({ serviceID }) {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)
  const [open, setOpen] = useState(false)
  const anchorRef = React.useRef(null)
  const [selectedIndex, setSelectedIndex] = useState(0)
  const { data, loading, error } = useQuery(query, {
    variables: { serviceID },
    returnPartialData: true,
  })

  const handleClick = () => {
    console.info(`You clicked ${options[selectedIndex]}`)
  }

  const handleMenuItemClick = (event, index) => {
    setSelectedIndex(index)
    setOpen(false)
  }

  const handleToggle = () => {
    setOpen((prevOpen) => !prevOpen)
  }

  const handleClose = (event) => {
    if (anchorRef.current && anchorRef.current.contains(event.target)) {
      return
    }

    setOpen(false)
  }

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
          <div key='maint-mode-key'>
            <ButtonGroup
              variant='contained'
              ref={anchorRef}
              aria-label='split button'
            >
              <Button onClick={handleClick}>Maintenance Mode</Button>
              <Button
                size='small'
                aria-controls={open ? 'split-button-menu' : undefined}
                aria-expanded={open ? 'true' : undefined}
                aria-label='select merge strategy'
                aria-haspopup='menu'
                onClick={handleToggle}
              >
                <ArrowDropDownIcon />
              </Button>
            </ButtonGroup>
            <Popper
              open={open}
              anchorEl={anchorRef.current}
              role={undefined}
              transition
              disablePortal
              placement='bottom'
            >
              {({ TransitionProps }) => (
                <Grow
                  {...TransitionProps}
                  style={{
                    transformOrigin: 'center top',
                  }}
                >
                  <Paper>
                    <ClickAwayListener onClickAway={handleClose}>
                      <MenuList id='split-button-menu' autoFocusItem>
                        {options.map((option, index) => (
                          <MenuItem
                            key={option}
                            selected={index === selectedIndex}
                            onClick={(event) =>
                              handleMenuItemClick(event, index)
                            }
                          >
                            {option}
                          </MenuItem>
                        ))}
                      </MenuList>
                    </ClickAwayListener>
                  </Paper>
                </Grow>
              )}
            </Popper>
          </div>,
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
