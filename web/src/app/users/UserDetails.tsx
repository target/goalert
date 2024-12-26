import React, { Suspense, useState } from 'react'
import { useQuery, gql } from 'urql'
import Delete from '@mui/icons-material/Delete'
import LockOpenIcon from '@mui/icons-material/LockOpen'
import DetailsPage from '../details/DetailsPage'
import { UserAvatar } from '../util/avatars'
import UserContactMethodList from './UserContactMethodList'
import { AddAlarm, SettingsPhone } from '@mui/icons-material'
import SpeedDial from '../util/SpeedDial'
import UserNotificationRuleList from './UserNotificationRuleList'
import { Grid } from '@mui/material'
import UserNotificationRuleCreateDialog from './UserNotificationRuleCreateDialog'
import UserContactMethodVerificationDialog from './UserContactMethodVerificationDialog'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useSessionInfo } from '../util/RequireConfig'
import UserEditDialog from './UserEditDialog'
import UserDeleteDialog from './UserDeleteDialog'
import { QuerySetFavoriteButton } from '../util/QuerySetFavoriteButton'
import { User } from '../../schema'
import { useIsWidthDown } from '../util/useWidth'
import UserShiftsCalendar from './UserShiftsCalendar'
import UserContactMethodCreateDialog from './UserContactMethodCreateDialog'

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
      onCallOverview {
        serviceCount
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
      onCallOverview {
        serviceCount
      }
      sessions {
        id
      }
    }
  }
`

export default function UserDetails(props: {
  userID: string
  readOnly: boolean
}): React.JSX.Element {
  const userID = props.userID
  const { userID: currentUserID, isAdmin } = useSessionInfo()

  const [createCM, setCreateCM] = useState(false)
  const [createNR, setCreateNR] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [showVerifyDialogByID, setShowVerifyDialogByID] = useState<
    string | null | undefined
  >(null)
  const [showUserDeleteDialog, setShowUserDeleteDialog] = useState(false)
  const mobile = useIsWidthDown('md')

  const [{ data, error }] = useQuery<{ user: User }>({
    query: isAdmin || userID === currentUserID ? profileQuery : userQuery,
    variables: { id: userID },
  })

  const user = data?.user
  if (error) return <GenericError error={error.message} />
  if (!user) return <ObjectNotFound />

  const svcCount = user.onCallOverview.serviceCount
  const sessCount = user?.sessions?.length ?? 0

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

  const options: (
    | React.JSX.Element
    | {
        label: string
        icon: React.JSX.Element
        handleOnClick: () => void
      }
  )[] = [
    <QuerySetFavoriteButton
      key='secondary-action-favorite'
      id={userID}
      type='user'
    />,
  ]

  if (isAdmin || userID === currentUserID) {
    options.unshift({
      label: 'Edit Access',
      icon: <LockOpenIcon />,
      handleOnClick: () => setShowEdit(true),
    })
  }
  if (isAdmin) {
    options.unshift({
      label: 'Delete',
      icon: <Delete />,
      handleOnClick: () => setShowUserDeleteDialog(true),
    })
  }

  return (
    <React.Fragment>
      {/* dialogs only shown on mobile via FAB button */}
      {mobile && !props.readOnly ? (
        <SpeedDial
          label='Add Items'
          actions={[
            {
              label: 'Create Contact Method',
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
      ) : null}
      <Suspense>
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
        {createCM && (
          <UserContactMethodCreateDialog
            userID={userID}
            onClose={(contactMethodID) => {
              setCreateCM(false)
              setShowVerifyDialogByID(contactMethodID)
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
      </Suspense>

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
            {!mobile && (
              <Suspense>
                <Grid item xs={12}>
                  <UserShiftsCalendar userID={userID} />
                </Grid>
              </Suspense>
            )}
          </Grid>
        }
        secondaryActions={options}
        links={links}
      />
    </React.Fragment>
  )
}
