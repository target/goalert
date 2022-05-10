import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import p from 'prop-types'
import { useParams } from 'react-router-dom'
import Delete from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import DetailsPage from '../details/DetailsPage'
import StatusUpdateNotification from './UserStatusUpdatePreference'
import { UserAvatar } from '../util/avatars'
import UserContactMethodList from './UserContactMethodList'
import { AddAlarm, SettingsPhone } from '@mui/icons-material'
import SpeedDial from '../util/SpeedDial'
import UserNotificationRuleList from './UserNotificationRuleList'
import { Grid } from '@mui/material'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import _ from 'lodash'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useConfigValue, useSessionInfo } from '../util/RequireConfig'
import UserEditDialog from './UserEditDialog'
import UserDeleteDialog from './UserDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'

const userQuery = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
      role
      name
      email
      contactMethods {
        id
      }
      onCallSteps {
        id
        escalationPolicy {
          id
          assignedTo {
            id
            name
          }
        }
      }
    }
  }
`

const profileQuery = gql`
  query profileInfo($id: ID!) {
    user(id: $id) {
      id
      role
      name
      email
      contactMethods {
        id
      }
      onCallSteps {
        id
        escalationPolicy {
          id
          assignedTo {
            id
            name
          }
        }
      }
      sessions {
        id
      }
    }
  }
`

function serviceCount(onCallSteps = []) {
  const svcs = {}
  ;(onCallSteps || []).forEach((s) =>
    (s.escalationPolicy.assignedTo || []).forEach(
      (svc) => (svcs[svc.id] = true),
    ),
  )

  return Object.keys(svcs).length
}

export default function UserDetails(props) {
  const { userID: _userID } = useParams()
  const userID = props.userID || _userID
  const {
    userID: currentUserID,
    isAdmin,
    ready: isSessionReady,
  } = useSessionInfo()
  const [disclaimer] = useConfigValue('General.NotificationDisclaimer')
  const [createCM, setCreateCM] = useState(false)
  const [createNR, setCreateNR] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)
  const [showUserDeleteDialog, setShowUserDeleteDialog] = useState(false)

  const {
    data,
    loading: isQueryLoading,
    error,
  } = useQuery(isAdmin || userID === currentUserID ? profileQuery : userQuery, {
    variables: { id: userID },
    skip: !userID,
  })

  const loading = !isSessionReady || isQueryLoading

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'user.id')) return loading ? <Spinner /> : <ObjectNotFound />

  const user = _.get(data, 'user')
  const svcCount = serviceCount(user.onCallSteps)
  const sessCount =
    isAdmin || userID === currentUserID ? user.sessions.length : 0

  const disableNR = user.contactMethods.length === 0

  const links = [
    {
      label: 'On-Call Assignments',
      url: 'on-call-assignments',
      subText: svcCount
        ? `On-call for ${svcCount} service${svcCount > 1 ? 's' : ''}`
        : 'Not currently on-call',
    },
  ]

  if (userID === currentUserID) {
    links.push({
      label: 'Schedule Calendar Subscriptions',
      url: 'schedule-calendar-subscriptions',
      subText: 'Manage schedules you have subscribed to',
    })
  }

  if (isAdmin || userID === currentUserID) {
    links.push({
      label: 'Active Sessions',
      url: 'sessions',
      subText: `${sessCount || 'No'} active session${
        sessCount === 1 ? '' : 's'
      }`,
    })
  }

  return (
    <React.Fragment>
      {showEdit && (
        <UserEditDialog
          onClose={() => setShowEdit(false)}
          userID={userID}
          role={user.role}
        />
      )}
      {showUserDeleteDialog && (
        <UserDeleteDialog
          userID={userID}
          onClose={() => setShowUserDeleteDialog(false)}
        />
      )}
      {props.readOnly ? null : (
        <SpeedDial
          label='Add Items'
          actions={[
            {
              label: 'Add Contact Method',
              icon: <SettingsPhone />,
              onClick: () => setCreateCM(true),
            },
            {
              label: 'Add Notification Rule',
              icon: <AddAlarm />,
              disabled: disableNR,
              onClick: () => setCreateNR(true),
            },
          ]}
        />
      )}
      {createCM && (
        <UserContactMethodCreateDialog
          userID={userID}
          disclaimer={disclaimer}
          onClose={(result) => {
            setCreateCM(false)
            setShowVerifyDialogByID(
              result && result.contactMethodID ? result.contactMethodID : null,
            )
          }}
        />
      )}
      {showVerifyDialogByID && (
        <UserContactMethodVerificationDialog
          contactMethodID={showVerifyDialogByID}
          onClose={() => setShowVerifyDialogByID(null)}
        />
      )}
      {createNR && (
        <UserNotificationRuleCreateDialog
          userID={userID}
          onClose={() => setCreateNR(false)}
        />
      )}
      <DetailsPage
        avatar={<UserAvatar userID={userID} />}
        title={user.name + (svcCount ? ' (On-Call)' : '')}
        subheader={user.email}
        pageContent={
          <Grid container spacing={2}>
            <UserContactMethodList userID={userID} readOnly={props.readOnly} />
            <UserNotificationRuleList
              userID={userID}
              readOnly={props.readOnly}
            />
          </Grid>
        }
        primaryActions={
          props.readOnly
            ? []
            : [
                <StatusUpdateNotification
                  key='primary-action-status-updates'
                  userID={userID}
                />,
              ]
        }
        secondaryActions={
          isAdmin
            ? [
                {
                  label: 'Delete',
                  icon: <Delete />,
                  handleOnClick: () => setShowUserDeleteDialog(true),
                },
                {
                  label: 'Edit',
                  icon: <EditIcon />,
                  handleOnClick: () => setShowEdit(true),
                },
                <QuerySetFavoriteButton
                  key='secondary-action-favorite'
                  id={userID}
                  type='user'
                />,
              ]
            : [
                <QuerySetFavoriteButton
                  key='secondary-action-favorite'
                  id={userID}
                  type='user'
                />,
              ]
        }
        links={links}
      />
    </React.Fragment>
  )
}

UserDetails.propTypes = {
  userID: p.string,
  readOnly: p.bool,
}
