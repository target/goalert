import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import p from 'prop-types'
import DetailsPage from '../details/DetailsPage'
import StatusUpdateNotification from './UserStatusUpdatePreference'
import { UserAvatar } from '../util/avatars'
import UserContactMethodList from './UserContactMethodList'
import { AddAlarm, SettingsPhone } from '@material-ui/icons'
import SpeedDial from '../util/SpeedDial'
import UserNotificationRuleList from './UserNotificationRuleList'
import { Grid } from '@material-ui/core'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import _ from 'lodash'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useConfigValue, useSessionInfo } from '../util/RequireConfig'
import AppLink from '../util/AppLink'

const userQuery = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
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

const useStyles = makeStyles({
  gravatarText: {
    textAlign: 'center',
    paddingTop: '0.5em',
    display: 'block',
  },
  profileImage: {
    width: 128,
    height: 128,
    margin: 'auto',
  },
})

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
  const classes = useStyles()

  const {
    userID: currentUserID,
    isAdmin,
    ready: isSessionReady,
  } = useSessionInfo()
  const [disclaimer] = useConfigValue('General.NotificationDisclaimer')
  const [createCM, setCreateCM] = useState(false)
  const [createNR, setCreateNR] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState(null)

  const { data, loading: isQueryLoading, error } = useQuery(
    isAdmin || props.userID === currentUserID ? profileQuery : userQuery,
    {
      variables: { id: props.userID },
      skip: !isSessionReady,
    },
  )

  const loading = !isSessionReady || isQueryLoading

  if (error) return <GenericError error={error.message} />
  if (!_.get(data, 'user.id')) return loading ? <Spinner /> : <ObjectNotFound />

  const user = _.get(data, 'user')
  const svcCount = serviceCount(user.onCallSteps)
  const sessCount =
    isAdmin || props.userID === currentUserID ? user.sessions.length : 0

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

  if (props.userID === currentUserID) {
    links.push({
      label: 'Schedule Calendar Subscriptions',
      url: 'schedule-calendar-subscriptions',
    })
  }

  if (isAdmin || props.userID === currentUserID) {
    links.push({
      label: 'Active Sessions',
      url: 'sessions',
      subText: `${sessCount} active session${sessCount === 1 ? '' : 's'}`,
    })
  }

  return (
    <React.Fragment>
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
          userID={props.userID}
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
          userID={props.userID}
          onClose={() => setCreateNR(false)}
        />
      )}
      <DetailsPage
        title={user.name + (svcCount ? ' (On-Call)' : '')}
        details={user.email}
        icon={
          <React.Fragment>
            <UserAvatar
              userID={props.userID}
              className={classes.profileImage}
            />
            <Typography variant='caption' className={classes.gravatarText}>
              Provided by{' '}
              <AppLink to='https://gravatar.com' newTab>
                Gravatar
              </AppLink>
            </Typography>
          </React.Fragment>
        }
        links={links}
        titleFooter={
          props.readOnly ? null : (
            <StatusUpdateNotification userID={props.userID} />
          )
        }
        pageFooter={
          <Grid container spacing={2}>
            <UserContactMethodList
              userID={props.userID}
              readOnly={props.readOnly}
            />
            <UserNotificationRuleList
              userID={props.userID}
              readOnly={props.readOnly}
            />
          </Grid>
        }
      />
    </React.Fragment>
  )
}

UserDetails.propTypes = {
  userID: p.string.isRequired,
  readOnly: p.bool,
}
